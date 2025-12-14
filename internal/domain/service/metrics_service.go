package service

import (
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

// MetricsService handles metrics operations
type MetricsService struct {
	metricsRepo repository.MetricsRepository
}

// NewMetricsService creates a new metrics service
func NewMetricsService(metricsRepo repository.MetricsRepository) *MetricsService {
	return &MetricsService{
		metricsRepo: metricsRepo,
	}
}

// GetMetrics retrieves metrics for a test run
func (s *MetricsService) GetMetrics(runID string) (*model.Metrics, error) {
	return s.metricsRepo.GetByRunID(runID)
}

// SaveMetrics saves metrics for a test run
func (s *MetricsService) SaveMetrics(metrics *model.Metrics) error {
	return s.metricsRepo.Save(metrics)
}
