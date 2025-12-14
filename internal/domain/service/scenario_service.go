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

// ScenarioService handles business logic for scenario operations
type ScenarioService struct {
	scenarioRepo          repository.ScenarioRepository
	scenarioExecutionRepo repository.ScenarioExecutionRepository
	executor              *engine.ScenarioExecutor
}

// NewScenarioService creates a new scenario service
func NewScenarioService(
	scenarioRepo repository.ScenarioRepository,
	scenarioExecutionRepo repository.ScenarioExecutionRepository,
	executor *engine.ScenarioExecutor,
) *ScenarioService {
	return &ScenarioService{
		scenarioRepo:          scenarioRepo,
		scenarioExecutionRepo: scenarioExecutionRepo,
		executor:              executor,
	}
}

// CreateScenario creates a new scenario
func (s *ScenarioService) CreateScenario(req *model.CreateScenarioRequest) (*model.Scenario, error) {
	scenario := &model.Scenario{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		Variables:   req.Variables,
		CreatedAt:   time.Now(),
	}

	if err := s.scenarioRepo.Create(scenario); err != nil {
		logger.Log.Error("Failed to create scenario", zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Scenario created",
		zap.String("scenario_id", scenario.ID),
		zap.String("name", scenario.Name),
		zap.Int("steps", len(scenario.Steps)))

	return scenario, nil
}

// GetScenario retrieves a scenario by ID
func (s *ScenarioService) GetScenario(id string) (*model.Scenario, error) {
	return s.scenarioRepo.GetByID(id)
}

// GetAllScenarios retrieves all scenarios
func (s *ScenarioService) GetAllScenarios() ([]*model.Scenario, error) {
	return s.scenarioRepo.GetAll()
}

// DeleteScenario deletes a scenario
func (s *ScenarioService) DeleteScenario(id string) error {
	return s.scenarioRepo.Delete(id)
}

// ExecuteScenario executes a scenario
func (s *ScenarioService) ExecuteScenario(req *model.ExecuteScenarioRequest) (*model.ScenarioExecution, error) {
	// Get scenario
	scenario, err := s.scenarioRepo.GetByID(req.ScenarioID)
	if err != nil {
		return nil, err
	}

	// Execute scenario
	execution, err := s.executor.Execute(scenario, req.Variables)
	if err != nil {
		// Even if execution failed, we want to store the result
		logger.Log.Warn("Scenario execution failed",
			zap.String("scenario_id", scenario.ID),
			zap.String("execution_id", execution.ID),
			zap.Error(err))
	}

	// Store execution result
	if storeErr := s.scenarioExecutionRepo.Create(execution); storeErr != nil {
		logger.Log.Error("Failed to store scenario execution",
			zap.String("execution_id", execution.ID),
			zap.Error(storeErr))
	}

	return execution, err
}

// GetScenarioExecution retrieves a scenario execution by ID
func (s *ScenarioService) GetScenarioExecution(id string) (*model.ScenarioExecution, error) {
	return s.scenarioExecutionRepo.GetByID(id)
}

// GetScenarioExecutions retrieves all executions for a scenario
func (s *ScenarioService) GetScenarioExecutions(scenarioID string) ([]*model.ScenarioExecution, error) {
	return s.scenarioExecutionRepo.GetByScenarioID(scenarioID)
}
