package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
)

// ScenarioHandler handles HTTP requests for scenario operations
type ScenarioHandler struct {
	service *service.ScenarioService
}

// NewScenarioHandler creates a new scenario handler
func NewScenarioHandler(service *service.ScenarioService) *ScenarioHandler {
	return &ScenarioHandler{service: service}
}

// CreateScenario handles POST /api/scenarios
func (h *ScenarioHandler) CreateScenario(c *gin.Context) {
	var req model.CreateScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenario, err := h.service.CreateScenario(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, scenario)
}

// GetScenario handles GET /api/scenarios/:id
func (h *ScenarioHandler) GetScenario(c *gin.Context) {
	id := c.Param("id")

	scenario, err := h.service.GetScenario(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	c.JSON(http.StatusOK, scenario)
}

// GetAllScenarios handles GET /api/scenarios
func (h *ScenarioHandler) GetAllScenarios(c *gin.Context) {
	scenarios, err := h.service.GetAllScenarios()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scenarios": scenarios,
		"count":     len(scenarios),
	})
}

// DeleteScenario handles DELETE /api/scenarios/:id
func (h *ScenarioHandler) DeleteScenario(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteScenario(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scenario not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "scenario deleted successfully"})
}

// ExecuteScenario handles POST /api/scenarios/execute
func (h *ScenarioHandler) ExecuteScenario(c *gin.Context) {
	var req model.ExecuteScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	execution, err := h.service.ExecuteScenario(&req)
	if err != nil {
		// Return execution result even on error for debugging
		if execution != nil {
			c.JSON(http.StatusExpectationFailed, execution)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetScenarioExecution handles GET /api/scenarios/executions/:id
func (h *ScenarioHandler) GetScenarioExecution(c *gin.Context) {
	id := c.Param("id")

	execution, err := h.service.GetScenarioExecution(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// GetScenarioExecutions handles GET /api/scenarios/:id/executions
func (h *ScenarioHandler) GetScenarioExecutions(c *gin.Context) {
	scenarioID := c.Param("id")

	executions, err := h.service.GetScenarioExecutions(scenarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scenario_id": scenarioID,
		"executions":  executions,
		"count":       len(executions),
	})
}
