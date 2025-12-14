package service

import (
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

// TestPlanService handles test plan operations
type TestPlanService struct {
	planRepo repository.TestPlanRepository
}

// NewTestPlanService creates a new test plan service
func NewTestPlanService(planRepo repository.TestPlanRepository) *TestPlanService {
	return &TestPlanService{
		planRepo: planRepo,
	}
}

// GetTestPlan retrieves a test plan by ID
func (s *TestPlanService) GetTestPlan(id string) (*model.TestPlan, error) {
	return s.planRepo.GetByID(id)
}

// GetAllTestPlans retrieves all test plans
func (s *TestPlanService) GetAllTestPlans() ([]*model.TestPlan, error) {
	return s.planRepo.GetAll()
}
