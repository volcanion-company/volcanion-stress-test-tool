package repository

import (
	"sync"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

var (
	ErrScenarioNotFound          = domain.NewNotFoundError("scenario", "")
	ErrScenarioExecutionNotFound = domain.NewNotFoundError("scenario execution", "")
)

// MemoryScenarioRepository implements ScenarioRepository using in-memory storage
type MemoryScenarioRepository struct {
	scenarios map[string]*model.Scenario
	mu        sync.RWMutex
}

// NewMemoryScenarioRepository creates a new in-memory scenario repository
func NewMemoryScenarioRepository() *MemoryScenarioRepository {
	return &MemoryScenarioRepository{
		scenarios: make(map[string]*model.Scenario),
	}
}

func (r *MemoryScenarioRepository) Create(scenario *model.Scenario) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scenarios[scenario.ID] = scenario
	return nil
}

func (r *MemoryScenarioRepository) GetByID(id string) (*model.Scenario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scenario, exists := r.scenarios[id]
	if !exists {
		return nil, ErrScenarioNotFound
	}
	return scenario, nil
}

func (r *MemoryScenarioRepository) GetAll() ([]*model.Scenario, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scenarios := make([]*model.Scenario, 0, len(r.scenarios))
	for _, scenario := range r.scenarios {
		scenarios = append(scenarios, scenario)
	}
	return scenarios, nil
}

func (r *MemoryScenarioRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.scenarios[id]; !exists {
		return ErrScenarioNotFound
	}
	delete(r.scenarios, id)
	return nil
}

// MemoryScenarioExecutionRepository implements ScenarioExecutionRepository using in-memory storage
type MemoryScenarioExecutionRepository struct {
	executions map[string]*model.ScenarioExecution
	mu         sync.RWMutex
}

// NewMemoryScenarioExecutionRepository creates a new in-memory scenario execution repository
func NewMemoryScenarioExecutionRepository() *MemoryScenarioExecutionRepository {
	return &MemoryScenarioExecutionRepository{
		executions: make(map[string]*model.ScenarioExecution),
	}
}

func (r *MemoryScenarioExecutionRepository) Create(execution *model.ScenarioExecution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executions[execution.ID] = execution
	return nil
}

func (r *MemoryScenarioExecutionRepository) GetByID(id string) (*model.ScenarioExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	execution, exists := r.executions[id]
	if !exists {
		return nil, ErrScenarioExecutionNotFound
	}
	return execution, nil
}

func (r *MemoryScenarioExecutionRepository) GetByScenarioID(scenarioID string) ([]*model.ScenarioExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executions := make([]*model.ScenarioExecution, 0)
	for _, execution := range r.executions {
		if execution.ScenarioID == scenarioID {
			executions = append(executions, execution)
		}
	}
	return executions, nil
}

func (r *MemoryScenarioExecutionRepository) Update(execution *model.ScenarioExecution) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.executions[execution.ID]; !exists {
		return ErrScenarioExecutionNotFound
	}
	r.executions[execution.ID] = execution
	return nil
}
