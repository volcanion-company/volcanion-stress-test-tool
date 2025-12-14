package engine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

var (
	ErrTestAlreadyRunning = errors.New("test is already running")
	ErrTestNotFound       = errors.New("test not found")
)

// LoadGenerator manages multiple concurrent test runs
type LoadGenerator struct {
	activeTests map[string]*TestExecution
	mu          sync.RWMutex
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
func NewLoadGenerator() *LoadGenerator {
	return &LoadGenerator{
		activeTests: make(map[string]*TestExecution),
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

	// Create scheduler
	scheduler := NewScheduler(plan, metrics)

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
}

// GetActiveTestCount returns the number of currently running tests
func (lg *LoadGenerator) GetActiveTestCount() int {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	return len(lg.activeTests)
}
