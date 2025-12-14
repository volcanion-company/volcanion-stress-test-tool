package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

// TestPlanHandler handles test plan related endpoints
type TestPlanHandler struct {
	service   *service.TestService
	validator *domain.Validator
}

// NewTestPlanHandler creates a new test plan handler
func NewTestPlanHandler(service *service.TestService) *TestPlanHandler {
	return &TestPlanHandler{
		service:   service,
		validator: domain.NewValidator(),
	}
}

// CreateTestPlan handles POST /api/test-plans
func (h *TestPlanHandler) CreateTestPlan(c *gin.Context) {
	var req model.CreateTestPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if err := h.validator.ValidateTestPlan(&req); err != nil {
		logger.Log.Warn("Validation failed", zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	plan, err := h.service.CreateTestPlan(&req)
	if err != nil {
		logger.Log.Error("Failed to create test plan", zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// GetTestPlans handles GET /api/test-plans
func (h *TestPlanHandler) GetTestPlans(c *gin.Context) {
	plans, err := h.service.GetAllTestPlans()
	if err != nil {
		logger.Log.Error("Failed to get test plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plans)
}

// GetTestPlan handles GET /api/test-plans/:id
func (h *TestPlanHandler) GetTestPlan(c *gin.Context) {
	id := c.Param("id")

	plan, err := h.service.GetTestPlan(id)
	if err != nil {
		logger.Log.Warn("Test plan not found", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "test plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}
