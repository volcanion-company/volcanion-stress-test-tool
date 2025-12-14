package postgres

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrMaxRetriesReached = errors.New("max connection retries reached")
	ErrHealthCheckFailed = errors.New("database health check failed")
)

// ResilientDBConfig extends DBConfig with resilience settings
type ResilientDBConfig struct {
	DBConfig

	// Retry settings
	MaxRetries     int           // Maximum number of connection retries
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	BackoffFactor  float64       // Backoff multiplier

	// Health check settings
	HealthCheckInterval time.Duration // How often to check connection health
	HealthCheckTimeout  time.Duration // Timeout for health check query

	// Reconnect settings
	EnableAutoReconnect bool // Enable automatic reconnection on failure
}

// DefaultResilientConfig returns default resilience settings
func DefaultResilientConfig() ResilientDBConfig {
	return ResilientDBConfig{
		MaxRetries:          5,
		InitialBackoff:      100 * time.Millisecond,
		MaxBackoff:          30 * time.Second,
		BackoffFactor:       2.0,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		EnableAutoReconnect: true,
	}
}

// ResilientDB wraps sql.DB with connection resilience
type ResilientDB struct {
	db     *sql.DB
	config ResilientDBConfig
	logger *zap.Logger

	mu           sync.RWMutex
	isConnected  bool
	lastError    error
	reconnecting bool

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewResilientDB creates a database connection with resilience features
func NewResilientDB(cfg ResilientDBConfig, logger *zap.Logger) (*ResilientDB, error) {
	rdb := &ResilientDB{
		config: cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Try to establish initial connection with retries
	if err := rdb.connectWithRetry(); err != nil {
		return nil, err
	}

	// Start health check goroutine
	if cfg.HealthCheckInterval > 0 {
		rdb.wg.Add(1)
		go rdb.healthCheckLoop()
	}

	return rdb, nil
}

// connectWithRetry attempts to connect with exponential backoff
func (rdb *ResilientDB) connectWithRetry() error {
	var lastErr error
	backoff := rdb.config.InitialBackoff

	for attempt := 0; attempt <= rdb.config.MaxRetries; attempt++ {
		if attempt > 0 {
			rdb.logger.Info("Retrying database connection",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)

			// Exponential backoff
			backoff = time.Duration(float64(backoff) * rdb.config.BackoffFactor)
			if backoff > rdb.config.MaxBackoff {
				backoff = rdb.config.MaxBackoff
			}
		}

		db, err := rdb.connect()
		if err == nil {
			rdb.mu.Lock()
			rdb.db = db
			rdb.isConnected = true
			rdb.lastError = nil
			rdb.mu.Unlock()

			rdb.logger.Info("Database connection established",
				zap.Int("attempts", attempt+1),
			)
			return nil
		}

		lastErr = err
		rdb.logger.Warn("Database connection attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	return errors.Join(ErrMaxRetriesReached, lastErr)
}

// connect establishes a single connection attempt
func (rdb *ResilientDB) connect() (*sql.DB, error) {
	if rdb.config.DSN == "" {
		return nil, errors.New("database DSN is required")
	}

	db, err := sql.Open("postgres", rdb.config.DSN)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(rdb.config.MaxConns)
	db.SetMaxIdleConns(rdb.config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// healthCheckLoop periodically checks database connection health
func (rdb *ResilientDB) healthCheckLoop() {
	defer rdb.wg.Done()

	ticker := time.NewTicker(rdb.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := rdb.healthCheck(); err != nil {
				rdb.logger.Warn("Database health check failed", zap.Error(err))

				if rdb.config.EnableAutoReconnect {
					go rdb.attemptReconnect()
				}
			}
		case <-rdb.stopCh:
			return
		}
	}
}

// healthCheck performs a health check on the database connection
func (rdb *ResilientDB) healthCheck() error {
	rdb.mu.RLock()
	db := rdb.db
	rdb.mu.RUnlock()

	if db == nil {
		return ErrConnectionFailed
	}

	ctx, cancel := context.WithTimeout(context.Background(), rdb.config.HealthCheckTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		rdb.mu.Lock()
		rdb.isConnected = false
		rdb.lastError = err
		rdb.mu.Unlock()
		return errors.Join(ErrHealthCheckFailed, err)
	}

	return nil
}

// attemptReconnect tries to re-establish connection
func (rdb *ResilientDB) attemptReconnect() {
	rdb.mu.Lock()
	if rdb.reconnecting {
		rdb.mu.Unlock()
		return
	}
	rdb.reconnecting = true
	rdb.mu.Unlock()

	defer func() {
		rdb.mu.Lock()
		rdb.reconnecting = false
		rdb.mu.Unlock()
	}()

	rdb.logger.Info("Attempting database reconnection")

	// Close existing connection
	rdb.mu.RLock()
	oldDB := rdb.db
	rdb.mu.RUnlock()

	if oldDB != nil {
		oldDB.Close()
	}

	// Try to reconnect
	if err := rdb.connectWithRetry(); err != nil {
		rdb.logger.Error("Database reconnection failed", zap.Error(err))
	} else {
		rdb.logger.Info("Database reconnection successful")
	}
}

// DB returns the underlying sql.DB, checking health first
func (rdb *ResilientDB) DB() *sql.DB {
	rdb.mu.RLock()
	defer rdb.mu.RUnlock()
	return rdb.db
}

// IsConnected returns current connection status
func (rdb *ResilientDB) IsConnected() bool {
	rdb.mu.RLock()
	defer rdb.mu.RUnlock()
	return rdb.isConnected
}

// LastError returns the last connection error
func (rdb *ResilientDB) LastError() error {
	rdb.mu.RLock()
	defer rdb.mu.RUnlock()
	return rdb.lastError
}

// Stats returns database connection pool statistics
func (rdb *ResilientDB) Stats() sql.DBStats {
	rdb.mu.RLock()
	db := rdb.db
	rdb.mu.RUnlock()

	if db == nil {
		return sql.DBStats{}
	}
	return db.Stats()
}

// Close shuts down the resilient database connection
func (rdb *ResilientDB) Close() error {
	close(rdb.stopCh)
	rdb.wg.Wait()

	rdb.mu.Lock()
	defer rdb.mu.Unlock()

	if rdb.db != nil {
		return rdb.db.Close()
	}
	return nil
}

// ExecContext executes a query with automatic retry on connection errors
func (rdb *ResilientDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := rdb.withRetry(func() (interface{}, error) {
		return rdb.db.ExecContext(ctx, query, args...)
	})
	if err != nil {
		return nil, err
	}
	return result.(sql.Result), nil
}

// QueryContext queries with automatic retry on connection errors
func (rdb *ResilientDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	result, err := rdb.withRetry(func() (interface{}, error) {
		return rdb.db.QueryContext(ctx, query, args...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*sql.Rows), nil
}

// QueryRowContext queries a single row with automatic retry
func (rdb *ResilientDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	rdb.mu.RLock()
	db := rdb.db
	rdb.mu.RUnlock()
	return db.QueryRowContext(ctx, query, args...)
}

// withRetry wraps a database operation with retry logic
func (rdb *ResilientDB) withRetry(op func() (interface{}, error)) (interface{}, error) {
	maxRetries := 3
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		rdb.mu.RLock()
		if !rdb.isConnected {
			rdb.mu.RUnlock()
			return nil, ErrConnectionFailed
		}
		rdb.mu.RUnlock()

		result, err := op()
		if err == nil {
			return result, nil
		}

		// Check if error is retryable (connection-related)
		if !isRetryableError(err) {
			return nil, err
		}

		rdb.logger.Debug("Retrying database operation",
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)

		time.Sleep(backoff)
		backoff = time.Duration(math.Min(float64(backoff*2), float64(5*time.Second)))
	}

	return nil, ErrMaxRetriesReached
}

// isRetryableError checks if an error is connection-related and retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no connection",
		"EOF",
		"timeout",
		"deadline exceeded",
	}
	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
