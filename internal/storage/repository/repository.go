package repository

import (
	"errors"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

var (
	ErrTestPlanNotFound = errors.New("test plan not found")
	ErrTestRunNotFound  = errors.New("test run not found")
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
