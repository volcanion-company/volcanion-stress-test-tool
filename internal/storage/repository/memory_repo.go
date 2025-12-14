package repository

import (
	"sync"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

// MemoryTestPlanRepository implements TestPlanRepository with in-memory storage
type MemoryTestPlanRepository struct {
	plans map[string]*model.TestPlan
	mu    sync.RWMutex
}

// NewMemoryTestPlanRepository creates a new in-memory test plan repository
func NewMemoryTestPlanRepository() *MemoryTestPlanRepository {
	return &MemoryTestPlanRepository{
		plans: make(map[string]*model.TestPlan),
	}
}

func (r *MemoryTestPlanRepository) Create(plan *model.TestPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plans[plan.ID] = plan
	return nil
}

func (r *MemoryTestPlanRepository) GetByID(id string) (*model.TestPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.plans[id]
	if !exists {
		return nil, ErrTestPlanNotFound
	}
	return plan, nil
}

func (r *MemoryTestPlanRepository) GetAll() ([]*model.TestPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plans := make([]*model.TestPlan, 0, len(r.plans))
	for _, plan := range r.plans {
		plans = append(plans, plan)
	}
	return plans, nil
}

func (r *MemoryTestPlanRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plans, id)
	return nil
}

// MemoryTestRunRepository implements TestRunRepository with in-memory storage
type MemoryTestRunRepository struct {
	runs map[string]*model.TestRun
	mu   sync.RWMutex
}

// NewMemoryTestRunRepository creates a new in-memory test run repository
func NewMemoryTestRunRepository() *MemoryTestRunRepository {
	return &MemoryTestRunRepository{
		runs: make(map[string]*model.TestRun),
	}
}

func (r *MemoryTestRunRepository) Create(run *model.TestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = run
	return nil
}

func (r *MemoryTestRunRepository) GetByID(id string) (*model.TestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	run, exists := r.runs[id]
	if !exists {
		return nil, ErrTestRunNotFound
	}
	return run, nil
}

func (r *MemoryTestRunRepository) GetAll() ([]*model.TestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	runs := make([]*model.TestRun, 0, len(r.runs))
	for _, run := range r.runs {
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *MemoryTestRunRepository) Update(run *model.TestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.runs[run.ID]; !exists {
		return ErrTestRunNotFound
	}
	r.runs[run.ID] = run
	return nil
}

func (r *MemoryTestRunRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, id)
	return nil
}

// MemoryMetricsRepository implements MetricsRepository with in-memory storage
type MemoryMetricsRepository struct {
	metrics map[string]*model.Metrics
	mu      sync.RWMutex
}

// NewMemoryMetricsRepository creates a new in-memory metrics repository
func NewMemoryMetricsRepository() *MemoryMetricsRepository {
	return &MemoryMetricsRepository{
		metrics: make(map[string]*model.Metrics),
	}
}

func (r *MemoryMetricsRepository) Save(metrics *model.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[metrics.RunID] = metrics
	return nil
}

func (r *MemoryMetricsRepository) GetByRunID(runID string) (*model.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics, exists := r.metrics[runID]
	if !exists {
		return nil, ErrTestRunNotFound
	}
	return metrics, nil
}

func (r *MemoryMetricsRepository) Delete(runID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, runID)
	return nil
}
