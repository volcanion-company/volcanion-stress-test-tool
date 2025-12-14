package service

import (
	"testing"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

func init() {
	// Initialize logger for tests
	logger.Init("error") // Use error level to reduce noise
}

func TestNewTestService(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	if service == nil {
		t.Fatal("Expected TestService to be created")
	}
}

func TestCreateTestPlan(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "Test Plan 1",
		TargetURL:   "http://localhost:8080/api/test",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
		RampUpSec:   5,
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	if plan.ID == "" {
		t.Error("Expected plan ID to be generated")
	}
	if plan.Name != req.Name {
		t.Errorf("Name mismatch: expected %s, got %s", req.Name, plan.Name)
	}
	if plan.TargetURL != req.TargetURL {
		t.Errorf("TargetURL mismatch: expected %s, got %s", req.TargetURL, plan.TargetURL)
	}
	if plan.Users != req.Users {
		t.Errorf("Users mismatch: expected %d, got %d", req.Users, plan.Users)
	}
	if plan.TimeoutMs != cfg.DefaultTimeout {
		t.Errorf("Expected default timeout %d, got %d", cfg.DefaultTimeout, plan.TimeoutMs)
	}
}

func TestCreateTestPlanExceedsMaxWorkers(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     50,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "Excessive Plan",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       100, // Exceeds max workers
		DurationSec: 60,
	}

	_, err := service.CreateTestPlan(req)
	if err == nil {
		t.Error("Expected error when users exceed max workers")
	}
}

func TestCreateTestPlanWithCustomTimeout(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "Custom Timeout Plan",
		TargetURL:   "http://localhost:8080",
		Method:      "POST",
		Users:       10,
		DurationSec: 60,
		TimeoutMs:   5000,
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	if plan.TimeoutMs != 5000 {
		t.Errorf("Expected custom timeout 5000, got %d", plan.TimeoutMs)
	}
}

func TestGetTestPlan(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "Get Test Plan",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
	}

	created, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	retrieved, err := service.GetTestPlan(created.ID)
	if err != nil {
		t.Fatalf("Failed to get test plan: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID mismatch: expected %s, got %s", created.ID, retrieved.ID)
	}
	if retrieved.Name != created.Name {
		t.Errorf("Name mismatch: expected %s, got %s", created.Name, retrieved.Name)
	}
}

func TestGetTestPlanNotFound(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	_, err := service.GetTestPlan("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent plan")
	}
}

func TestGetAllTestPlans(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	// Create multiple plans
	for i := 0; i < 5; i++ {
		req := &model.CreateTestPlanRequest{
			Name:        "Plan " + string(rune('A'+i)),
			TargetURL:   "http://localhost:8080",
			Method:      "GET",
			Users:       10,
			DurationSec: 60,
		}
		_, err := service.CreateTestPlan(req)
		if err != nil {
			t.Fatalf("Failed to create test plan: %v", err)
		}
	}

	plans, err := service.GetAllTestPlans()
	if err != nil {
		t.Fatalf("Failed to get all test plans: %v", err)
	}

	if len(plans) != 5 {
		t.Errorf("Expected 5 plans, got %d", len(plans))
	}
}

func TestCreateTestPlanWithHeaders(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:      "Plan with Headers",
		TargetURL: "http://localhost:8080",
		Method:    "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Body:        `{"key": "value"}`,
		Users:       10,
		DurationSec: 60,
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	if len(plan.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(plan.Headers))
	}
	if plan.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header mismatch")
	}
	if plan.Body != `{"key": "value"}` {
		t.Errorf("Body mismatch")
	}
}

func TestCreateTestPlanTimestamp(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	beforeCreate := time.Now()

	req := &model.CreateTestPlanRequest{
		Name:        "Timestamp Test",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	afterCreate := time.Now()

	if plan.CreatedAt.Before(beforeCreate) || plan.CreatedAt.After(afterCreate) {
		t.Errorf("CreatedAt timestamp %v not within expected range", plan.CreatedAt)
	}
}

func TestCreateTestPlanDefaultRatePattern(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "Default Rate Pattern",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	if plan.RatePattern != model.RatePatternFixed {
		t.Errorf("Expected default rate pattern %s, got %s", model.RatePatternFixed, plan.RatePattern)
	}
}

func TestCreateTestPlanWithSLA(t *testing.T) {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}

	service := NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)

	req := &model.CreateTestPlanRequest{
		Name:        "SLA Test",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
		SLA: &model.SLAConfig{
			MaxP95Latency: 500,
			MaxErrorRate:  0.01,
			MinRPS:        100,
		},
	}

	plan, err := service.CreateTestPlan(req)
	if err != nil {
		t.Fatalf("Failed to create test plan: %v", err)
	}

	if plan.SLA == nil {
		t.Fatal("Expected SLA to be set")
	}
	if plan.SLA.MaxP95Latency != 500 {
		t.Errorf("MaxP95Latency mismatch: expected 500, got %f", plan.SLA.MaxP95Latency)
	}
	if plan.SLA.MaxErrorRate != 0.01 {
		t.Errorf("MaxErrorRate mismatch: expected 0.01, got %f", plan.SLA.MaxErrorRate)
	}
}
