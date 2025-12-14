package repository

import (
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

var (
	ErrTestPlanNotFound = domain.NewNotFoundError("test plan", "")
	ErrTestRunNotFound  = domain.NewNotFoundError("test run", "")
)

// TestPlanRepository defines interface for test plan storage
type TestPlanRepository interface {
	Create(plan *model.TestPlan) error
	GetByID(id string) (*model.TestPlan, error)
	GetAll() ([]*model.TestPlan, error)
	Delete(id string) error
}

// TestRunRepository defines interface for test run storage
type TestRunRepository interface {
	Create(run *model.TestRun) error
	GetByID(id string) (*model.TestRun, error)
	GetAll() ([]*model.TestRun, error)
	Update(run *model.TestRun) error
	Delete(id string) error
}

// MetricsRepository defines interface for metrics storage
type MetricsRepository interface {
	Save(metrics *model.Metrics) error
	GetByRunID(runID string) (*model.Metrics, error)
	Delete(runID string) error
}

// ScenarioRepository defines interface for scenario storage
type ScenarioRepository interface {
	Create(scenario *model.Scenario) error
	GetByID(id string) (*model.Scenario, error)
	GetAll() ([]*model.Scenario, error)
	Delete(id string) error
}

// ScenarioExecutionRepository defines interface for scenario execution storage
type ScenarioExecutionRepository interface {
	Create(execution *model.ScenarioExecution) error
	GetByID(id string) (*model.ScenarioExecution, error)
	GetByScenarioID(scenarioID string) ([]*model.ScenarioExecution, error)
	Update(execution *model.ScenarioExecution) error
}
