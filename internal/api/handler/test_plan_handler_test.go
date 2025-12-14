package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/config"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init("error")
}
func setupTestService() *service.TestService {
	planRepo := repository.NewMemoryTestPlanRepository()
	runRepo := repository.NewMemoryTestRunRepository()
	metricsRepo := repository.NewMemoryMetricsRepository()
	cfg := &config.Config{
		MaxWorkers:     100,
		DefaultTimeout: 30000,
	}
	return service.NewTestService(planRepo, runRepo, metricsRepo, nil, cfg)
}

func TestCreateTestPlanHandler(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	router := gin.New()
	router.POST("/api/test-plans", handler.CreateTestPlan)

	reqBody := model.CreateTestPlanRequest{
		Name:        "Test Plan",
		TargetURL:   "http://localhost:8080/api/test",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
		RampUpSec:   5,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/test-plans", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var plan model.TestPlan
	if err := json.Unmarshal(w.Body.Bytes(), &plan); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if plan.ID == "" {
		t.Error("Expected plan ID to be set")
	}
	if plan.Name != reqBody.Name {
		t.Errorf("Name mismatch: expected %s, got %s", reqBody.Name, plan.Name)
	}
}

func TestCreateTestPlanHandlerInvalidJSON(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	router := gin.New()
	router.POST("/api/test-plans", handler.CreateTestPlan)

	req := httptest.NewRequest(http.MethodPost, "/api/test-plans", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetTestPlansHandler(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	// Create some test plans first
	for i := 0; i < 3; i++ {
		svc.CreateTestPlan(&model.CreateTestPlanRequest{
			Name:        "Plan " + string(rune('A'+i)),
			TargetURL:   "http://localhost:8080",
			Method:      "GET",
			Users:       10,
			DurationSec: 60,
		})
	}

	router := gin.New()
	router.GET("/api/test-plans", handler.GetTestPlans)

	req := httptest.NewRequest(http.MethodGet, "/api/test-plans", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var plans []*model.TestPlan
	if err := json.Unmarshal(w.Body.Bytes(), &plans); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(plans) != 3 {
		t.Errorf("Expected 3 plans, got %d", len(plans))
	}
}

func TestGetTestPlanHandler(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	// Create a test plan
	plan, _ := svc.CreateTestPlan(&model.CreateTestPlanRequest{
		Name:        "Test Plan",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
	})

	router := gin.New()
	router.GET("/api/test-plans/:id", handler.GetTestPlan)

	req := httptest.NewRequest(http.MethodGet, "/api/test-plans/"+plan.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var retrieved model.TestPlan
	if err := json.Unmarshal(w.Body.Bytes(), &retrieved); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if retrieved.ID != plan.ID {
		t.Errorf("ID mismatch: expected %s, got %s", plan.ID, retrieved.ID)
	}
}

func TestGetTestPlanHandlerNotFound(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	router := gin.New()
	router.GET("/api/test-plans/:id", handler.GetTestPlan)

	req := httptest.NewRequest(http.MethodGet, "/api/test-plans/non-existent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCreateTestPlanHandlerWithHeaders(t *testing.T) {
	svc := setupTestService()
	handler := NewTestPlanHandler(svc)

	router := gin.New()
	router.POST("/api/test-plans", handler.CreateTestPlan)

	reqBody := model.CreateTestPlanRequest{
		Name:      "Plan with Headers",
		TargetURL: "http://localhost:8080/api/test",
		Method:    "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Body:        `{"key": "value"}`,
		Users:       10,
		DurationSec: 60,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/test-plans", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var plan model.TestPlan
	if err := json.Unmarshal(w.Body.Bytes(), &plan); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(plan.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(plan.Headers))
	}
}
