package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // imported for side-effects: registers the postgres driver
	"go.uber.org/zap"
)

// DBConfig holds database connection configuration
type DBConfig struct {
	DSN           string
	MaxConns      int
	MaxIdleConns  int
	MaxRetries    int           // Maximum number of connection retries (default: 5)
	RetryInterval time.Duration // Initial retry interval (default: 1s)
}

// NewDB creates a new database connection with exponential backoff retry
func NewDB(cfg DBConfig, logger *zap.Logger) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	// Set defaults for retry configuration
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}
	retryInterval := cfg.RetryInterval
	if retryInterval <= 0 {
		retryInterval = time.Second
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Ping with exponential backoff retry
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err == nil {
			logger.Info("database connection established",
				zap.Int("max_conns", cfg.MaxConns),
				zap.Int("max_idle_conns", cfg.MaxIdleConns),
				zap.Int("attempts", attempt),
			)
			return db, nil
		}

		lastErr = err
		if attempt < maxRetries {
			// Cap exponent to avoid large shifts and potential overflow
			exp := attempt - 1
			if exp < 0 {
				exp = 0
			}
			if exp > 10 {
				exp = 10
			}
			// Build multiplier by repeated doubling to avoid integer->uint conversions
			mult := time.Duration(1)
			for i := 0; i < exp; i++ {
				mult *= 2
			}
			waitTime := retryInterval * mult // Exponential backoff (capped)
			if waitTime > 30*time.Second {
				waitTime = 30 * time.Second // Cap at 30 seconds
			}
			logger.Warn("database connection failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Duration("retry_in", waitTime),
				zap.Error(err),
			)
			time.Sleep(waitTime)
		}
	}

	db.Close()
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, lastErr)
}
