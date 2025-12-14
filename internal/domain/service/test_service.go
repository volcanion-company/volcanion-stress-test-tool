package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/engine"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
	"go.uber.org/zap"
)

// TestService handles business logic for test operations
type TestService struct {
	planRepo    repository.TestPlanRepository
	runRepo     repository.TestRunRepository
	metricsRepo repository.MetricsRepository
	generator   *engine.LoadGenerator
}

// NewTestService creates a new test service
func NewTestService(
	planRepo repository.TestPlanRepository,
	runRepo repository.TestRunRepository,
	metricsRepo repository.MetricsRepository,
	generator *engine.LoadGenerator,
) *TestService {
	return &TestService{
		planRepo:    planRepo,
		runRepo:     runRepo,
		metricsRepo: metricsRepo,
		generator:   generator,
	}
}

// CreateTestPlan creates a new test plan
func (s *TestService) CreateTestPlan(req *model.CreateTestPlanRequest) (*model.TestPlan, error) {
	plan := &model.TestPlan{
		ID:          uuid.New().String(),
		Name:        req.Name,
		TargetURL:   req.TargetURL,
		Method:      req.Method,
		Headers:     req.Headers,
		Body:        req.Body,
		Users:       req.Users,
		RampUpSec:   req.RampUpSec,
		DurationSec: req.DurationSec,
		TimeoutMs:   req.TimeoutMs,
		CreatedAt:   time.Now(),
	}

	// Set default timeout if not specified
	if plan.TimeoutMs == 0 {
		plan.TimeoutMs = 30000 // 30 seconds default
	}

	if err := s.planRepo.Create(plan); err != nil {
		return nil, err
	}

	logger.Log.Info("Test plan created",
		zap.String("plan_id", plan.ID),
		zap.String("name", plan.Name))

	return plan, nil
}

// GetTestPlan retrieves a test plan by ID
func (s *TestService) GetTestPlan(id string) (*model.TestPlan, error) {
	return s.planRepo.GetByID(id)
}

// GetAllTestPlans retrieves all test plans
func (s *TestService) GetAllTestPlans() ([]*model.TestPlan, error) {
	return s.planRepo.GetAll()
}

// StartTest starts a new test run
func (s *TestService) StartTest(planID string) (*model.TestRun, error) {
	// Get test plan
	plan, err := s.planRepo.GetByID(planID)
	if err != nil {
		return nil, err
	}

	// Create test run
	run := &model.TestRun{
		ID:        uuid.New().String(),
		PlanID:    planID,
		Status:    model.StatusRunning,
		StartAt:   time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.runRepo.Create(run); err != nil {
		return nil, err
	}

	// Start the load generator
	metrics, err := s.generator.StartTest(run.ID, plan)
	if err != nil {
		// Update run status to failed
		run.Status = model.StatusFailed
		now := time.Now()
		run.EndAt = &now
		_ = s.runRepo.Update(run)
		return nil, err
	}

	// Save initial metrics
	if err := s.metricsRepo.Save(metrics); err != nil {
		logger.Log.Error("Failed to save metrics",
			zap.String("run_id", run.ID),
			zap.Error(err))
	}

	// Start background goroutine to monitor test completion
	go s.monitorTestRun(run)

	logger.Log.Info("Test run started",
		zap.String("run_id", run.ID),
		zap.String("plan_id", planID))

	return run, nil
}

// monitorTestRun monitors a test run and updates its status when complete
func (s *TestService) monitorTestRun(run *model.TestRun) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// Check if test is still running
		if !s.generator.IsRunning(run.ID) {
			// Test completed, update status
			run.Status = model.StatusCompleted
			now := time.Now()
			run.EndAt = &now

			if err := s.runRepo.Update(run); err != nil {
				logger.Log.Error("Failed to update test run",
					zap.String("run_id", run.ID),
					zap.Error(err))
			}

			// Get final metrics and save
			if metrics, err := s.generator.GetMetrics(run.ID); err == nil {
				_ = s.metricsRepo.Save(metrics)
			}

			logger.Log.Info("Test run completed",
				zap.String("run_id", run.ID),
				zap.String("status", string(run.Status)))

			return
		}

		// Update metrics
		if metrics, err := s.generator.GetMetrics(run.ID); err == nil {
			_ = s.metricsRepo.Save(metrics)
		}
	}
}

// StopTest stops a running test
func (s *TestService) StopTest(runID string) error {
	// Get test run
	run, err := s.runRepo.GetByID(runID)
	if err != nil {
		return err
	}

	// Stop the generator
	if err := s.generator.StopTest(runID); err != nil {
		return err
	}

	// Update run status
	run.Status = model.StatusCancelled
	now := time.Now()
	run.EndAt = &now

	if err := s.runRepo.Update(run); err != nil {
		return err
	}

	logger.Log.Info("Test run stopped",
		zap.String("run_id", runID))

	return nil
}

// GetTestRun retrieves a test run by ID
func (s *TestService) GetTestRun(id string) (*model.TestRun, error) {
	return s.runRepo.GetByID(id)
}

// GetAllTestRuns retrieves all test runs
func (s *TestService) GetAllTestRuns() ([]*model.TestRun, error) {
	return s.runRepo.GetAll()
}

// GetTestMetrics retrieves metrics for a test run
func (s *TestService) GetTestMetrics(runID string) (*model.Metrics, error) {
	// Try to get from active tests first
	if s.generator.IsRunning(runID) {
		return s.generator.GetMetrics(runID)
	}

	// Get from repository
	return s.metricsRepo.GetByRunID(runID)
}

// GetLiveMetrics retrieves real-time metrics for a running test
func (s *TestService) GetLiveMetrics(runID string) (*model.Metrics, error) {
	if !s.generator.IsRunning(runID) {
		// Test not running, get stored metrics
		return s.metricsRepo.GetByRunID(runID)
	}

	return s.generator.GetMetrics(runID)
}
