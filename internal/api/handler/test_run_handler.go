package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

// TestRunHandler handles test run related endpoints
type TestRunHandler struct {
	service *service.TestService
}

// NewTestRunHandler creates a new test run handler
func NewTestRunHandler(service *service.TestService) *TestRunHandler {
	return &TestRunHandler{
		service: service,
	}
}

// StartTest handles POST /api/test-runs/start
func (h *TestRunHandler) StartTest(c *gin.Context) {
	var req model.StartTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	run, err := h.service.StartTest(req.PlanID)
	if err != nil {
		logger.Log.Error("Failed to start test", zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusCreated, run)
}

// StopTest handles POST /api/test-runs/:id/stop
func (h *TestRunHandler) StopTest(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.StopTest(id); err != nil {
		logger.Log.Error("Failed to stop test", zap.String("id", id), zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test stopped successfully"})
}

// GetTestRuns handles GET /api/test-runs
func (h *TestRunHandler) GetTestRuns(c *gin.Context) {
	runs, err := h.service.GetAllTestRuns()
	if err != nil {
		logger.Log.Error("Failed to get test runs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, runs)
}

// GetTestRun handles GET /api/test-runs/:id
func (h *TestRunHandler) GetTestRun(c *gin.Context) {
	id := c.Param("id")

	run, err := h.service.GetTestRun(id)
	if err != nil {
		logger.Log.Warn("Test run not found", zap.String("id", id), zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, run)
}

// GetTestMetrics handles GET /api/test-runs/:id/metrics
func (h *TestRunHandler) GetTestMetrics(c *gin.Context) {
	id := c.Param("id")

	metrics, err := h.service.GetTestMetrics(id)
	if err != nil {
		logger.Log.Warn("Metrics not found", zap.String("id", id), zap.Error(err))
		MapErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetLiveMetrics handles GET /api/test-runs/:id/live
func (h *TestRunHandler) GetLiveMetrics(c *gin.Context) {
	id := c.Param("id")

	metrics, err := h.service.GetLiveMetrics(id)
	if err != nil {
		logger.Log.Warn("Live metrics not found", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "metrics not found"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
