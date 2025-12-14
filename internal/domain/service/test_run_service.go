package service

import (
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

// TestRunService handles test run operations
type TestRunService struct {
	runRepo repository.TestRunRepository
}

// NewTestRunService creates a new test run service
func NewTestRunService(runRepo repository.TestRunRepository) *TestRunService {
	return &TestRunService{
		runRepo: runRepo,
	}
}

// GetTestRun retrieves a test run by ID
func (s *TestRunService) GetTestRun(id string) (*model.TestRun, error) {
	return s.runRepo.GetByID(id)
}

// GetAllTestRuns retrieves all test runs
func (s *TestRunService) GetAllTestRuns() ([]*model.TestRun, error) {
	return s.runRepo.GetAll()
}
