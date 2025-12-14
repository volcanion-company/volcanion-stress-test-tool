package engine

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"go.uber.org/zap"
)

var (
	ErrTestAlreadyRunning = domain.ErrAlreadyRunning
	ErrTestNotFound       = domain.NewNotFoundError("test", "")
)

// LoadGenerator manages multiple concurrent test runs
type LoadGenerator struct {
	activeTests  map[string]*TestExecution
	mu           sync.RWMutex
	sharedClient *http.Client
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
	collector    *metrics.Collector
}

// TestExecution holds the runtime state of a test
type TestExecution struct {
	RunID     string
	Plan      *model.TestPlan
	Metrics   *model.Metrics
	Scheduler *Scheduler
	StartTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLoadGenerator creates a new load generator instance
func NewLoadGenerator(collector *metrics.Collector) *LoadGenerator {
	// Create shared HTTP client with optimized transport
	sharedTransport := &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     200,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	sharedClient := &http.Client{
		Transport: sharedTransport,
		// Timeout is set per-request in worker
	}

	shutdownCtx, shutdownFunc := context.WithCancel(context.Background())

	return &LoadGenerator{
		activeTests:  make(map[string]*TestExecution),
		sharedClient: sharedClient,
		shutdownCtx:  shutdownCtx,
		shutdownFunc: shutdownFunc,
		collector:    collector,
	}
}

// StartTest initiates a new test run
func (lg *LoadGenerator) StartTest(runID string, plan *model.TestPlan) (*model.Metrics, error) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	// Check if test is already running
	if _, exists := lg.activeTests[runID]; exists {
		return nil, ErrTestAlreadyRunning
	}

	// Create metrics
	metrics := model.NewMetrics(runID)

	// Create scheduler with shared client
	scheduler := NewScheduler(plan, metrics, lg.sharedClient, lg.collector)

	// Create test execution context
	ctx, cancel := context.WithCancel(context.Background())

	execution := &TestExecution{
		RunID:     runID,
		Plan:      plan,
		Metrics:   metrics,
		Scheduler: scheduler,
		StartTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	lg.activeTests[runID] = execution

	// Update active tests gauge
	lg.collector.SetActiveTests(len(lg.activeTests))

	// Start the test in background
	go lg.runTest(execution)

	logger.Log.Info("Test started",
		zap.String("run_id", runID),
		zap.String("plan_id", plan.ID))

	return metrics, nil
}

// runTest executes the test and handles cleanup
func (lg *LoadGenerator) runTest(execution *TestExecution) {
	// Start scheduler
	if err := execution.Scheduler.Start(); err != nil {
		logger.Log.Error("Failed to start scheduler",
			zap.String("run_id", execution.RunID),
			zap.Error(err))
		lg.cleanupTest(execution.RunID)
		return
	}

	// Wait for test to complete
	execution.Scheduler.Wait()

	// Calculate total duration
	duration := time.Since(execution.StartTime)
	execution.Metrics.Mu.Lock()
	execution.Metrics.TotalDurationMs = duration.Milliseconds()
	execution.Metrics.Mu.Unlock()

	logger.Log.Info("Test completed",
		zap.String("run_id", execution.RunID),
		zap.Duration("duration", duration),
		zap.Int64("total_requests", execution.Metrics.TotalRequests))

	// Cleanup after test
	lg.cleanupTest(execution.RunID)
}

// StopTest stops a running test
func (lg *LoadGenerator) StopTest(runID string) error {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	execution, exists := lg.activeTests[runID]
	if !exists {
		return ErrTestNotFound
	}

	// Stop the scheduler
	execution.Scheduler.Stop()
	execution.cancel()

	logger.Log.Info("Test stopped",
		zap.String("run_id", runID))

	return nil
}

// GetMetrics retrieves current metrics for a running test
func (lg *LoadGenerator) GetMetrics(runID string) (*model.Metrics, error) {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	execution, exists := lg.activeTests[runID]
	if !exists {
		return nil, ErrTestNotFound
	}

	return execution.Metrics.GetSnapshot(), nil
}

// IsRunning checks if a test is currently running
func (lg *LoadGenerator) IsRunning(runID string) bool {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	_, exists := lg.activeTests[runID]
	return exists
}

// cleanupTest removes test from active tests
func (lg *LoadGenerator) cleanupTest(runID string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()
	delete(lg.activeTests, runID)
	lg.collector.SetActiveTests(len(lg.activeTests))
}

// GetActiveTestCount returns the number of currently running tests
func (lg *LoadGenerator) GetActiveTestCount() int {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	return len(lg.activeTests)
}

// Shutdown stops all active tests and waits for completion
func (lg *LoadGenerator) Shutdown(timeout time.Duration) error {
	logger.Log.Info("Shutting down load generator",
		zap.Int("active_tests", lg.GetActiveTestCount()))

	// Signal shutdown to all tests
	lg.shutdownFunc()

	// Get all active test IDs
	lg.mu.RLock()
	testIDs := make([]string, 0, len(lg.activeTests))
	for id := range lg.activeTests {
		testIDs = append(testIDs, id)
	}
	lg.mu.RUnlock()

	// Stop each test
	for _, runID := range testIDs {
		_ = lg.StopTest(runID)
	}

	// Wait for all tests to complete with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if lg.GetActiveTestCount() == 0 {
			logger.Log.Info("All tests stopped")
			return nil
		}

		select {
		case <-ctx.Done():
			logger.Log.Warn("Shutdown timeout, forcing cleanup",
				zap.Int("remaining_tests", lg.GetActiveTestCount()))
			return ctx.Err()
		case <-ticker.C:
			// Continue waiting
		}
	}
}
