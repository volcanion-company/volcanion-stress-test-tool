package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
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
	config      *config.Config
}

// NewTestService creates a new test service
func NewTestService(
	planRepo repository.TestPlanRepository,
	runRepo repository.TestRunRepository,
	metricsRepo repository.MetricsRepository,
	generator *engine.LoadGenerator,
	cfg *config.Config,
) *TestService {
	return &TestService{
		planRepo:    planRepo,
		runRepo:     runRepo,
		metricsRepo: metricsRepo,
		generator:   generator,
		config:      cfg,
	}
}

// CreateTestPlan creates a new test plan
func (s *TestService) CreateTestPlan(req *model.CreateTestPlanRequest) (*model.TestPlan, error) {
	// Validate against max workers
	if req.Users > s.config.MaxWorkers {
		return nil, fmt.Errorf("users (%d) exceeds maximum allowed workers (%d)", req.Users, s.config.MaxWorkers)
	}

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
		TargetRPS:   req.TargetRPS,
		RatePattern: req.RatePattern,
		RateSteps:   req.RateSteps,
		SLA:         req.SLA,
		CreatedAt:   time.Now(),
	}

	// Set defaults
	if plan.TimeoutMs == 0 {
		plan.TimeoutMs = s.config.DefaultTimeout
	}
	if plan.RatePattern == "" {
		plan.RatePattern = model.RatePatternFixed
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
		return nil, domain.NewNotFoundError("test plan", planID)
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

		// Get current run status to check if cancelled
		currentRun, err := s.runRepo.GetByID(run.ID)
		if err != nil {
			logger.Log.Error("Failed to get test run",
				zap.String("run_id", run.ID),
				zap.Error(err))
			return
		}

		// If already cancelled or failed, don't overwrite
		if currentRun.Status == model.StatusCancelled || currentRun.Status == model.StatusFailed {
			logger.Log.Info("Test run already terminated",
				zap.String("run_id", run.ID),
				zap.String("status", string(currentRun.Status)))
			return
		}

		// Check if test is still running
		if !s.generator.IsRunning(run.ID) {
			// Test completed naturally
			reason := model.ReasonCompleted
			currentRun.Status = model.StatusCompleted
			currentRun.StopReason = &reason
			now := time.Now()
			currentRun.EndAt = &now

			if err := s.runRepo.Update(currentRun); err != nil {
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
				zap.String("status", string(currentRun.Status)))

			return
		}

		// Update metrics and check SLA
		if metrics, err := s.generator.GetMetrics(run.ID); err == nil {
			_ = s.metricsRepo.Save(metrics)

			// Check SLA violations
			plan, _ := s.planRepo.GetByID(currentRun.PlanID)
			if plan != nil && plan.SLA != nil {
				if s.checkSLAViolation(metrics, plan.SLA) {
					// Mark as failed due to SLA violation
					reason := model.ReasonFailed
					currentRun.Status = model.StatusFailed
					currentRun.StopReason = &reason
					now := time.Now()
					currentRun.EndAt = &now

					// Stop the test
					_ = s.generator.StopTest(run.ID)

					if err := s.runRepo.Update(currentRun); err != nil {
						logger.Log.Error("Failed to update test run after SLA violation",
							zap.String("run_id", run.ID),
							zap.Error(err))
					}

					logger.Log.Warn("Test run failed due to SLA violation",
						zap.String("run_id", run.ID))

					return
				}
			}
		}
	}
}

// checkSLAViolation checks if current metrics violate SLA thresholds
func (s *TestService) checkSLAViolation(metrics *model.Metrics, sla *model.SLAConfig) bool {
	if sla == nil {
		return false
	}

	// Check P95 latency
	if sla.MaxP95Latency > 0 && metrics.P95LatencyMs > sla.MaxP95Latency {
		logger.Log.Warn("SLA violation: P95 latency exceeded",
			zap.Float64("current", metrics.P95LatencyMs),
			zap.Float64("max", sla.MaxP95Latency))
		return true
	}

	// Check P99 latency
	if sla.MaxP99Latency > 0 && metrics.P99LatencyMs > sla.MaxP99Latency {
		logger.Log.Warn("SLA violation: P99 latency exceeded",
			zap.Float64("current", metrics.P99LatencyMs),
			zap.Float64("max", sla.MaxP99Latency))
		return true
	}

	// Check error rate
	if sla.MaxErrorRate > 0 && metrics.TotalRequests > 0 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
		if errorRate > sla.MaxErrorRate {
			logger.Log.Warn("SLA violation: Error rate exceeded",
				zap.Float64("current", errorRate),
				zap.Float64("max", sla.MaxErrorRate))
			return true
		}
	}

	// Check minimum RPS
	if sla.MinRPS > 0 && metrics.CurrentRPS > 0 && metrics.CurrentRPS < sla.MinRPS {
		logger.Log.Warn("SLA violation: RPS below minimum",
			zap.Float64("current", metrics.CurrentRPS),
			zap.Float64("min", sla.MinRPS))
		return true
	}

	return false
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
	reason := model.ReasonCancelled
	run.Status = model.StatusCancelled
	run.StopReason = &reason
	now := time.Now()
	run.EndAt = &now

	if err := s.runRepo.Update(run); err != nil {
		return err
	}

	// Save final metrics
	if metrics, err := s.generator.GetMetrics(runID); err == nil {
		_ = s.metricsRepo.Save(metrics)
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
