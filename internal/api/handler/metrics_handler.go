package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/postgres"
	"go.uber.org/zap"
)

// MetricsHandler handles HTTP requests for historical metrics
type MetricsHandler struct {
	snapshotRepo *postgres.MetricsSnapshotRepository
	logger       *zap.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(snapshotRepo *postgres.MetricsSnapshotRepository, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		snapshotRepo: snapshotRepo,
		logger:       logger,
	}
}

// GetHistoricalMetrics returns time-series metrics for a test run
// GET /api/metrics/history/:runId?start=<timestamp>&end=<timestamp>&limit=<int>
func (h *MetricsHandler) GetHistoricalMetrics(c *gin.Context) {
	runID := c.Param("runId")
	if runID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "run_id is required"})
		return
	}

	// Check if limit is provided (for latest N snapshots)
	limitStr := c.Query("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}

		snapshots, err := h.snapshotRepo.GetLatestSnapshots(runID, limit)
		if err != nil {
			h.logger.Error("failed to get latest snapshots", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve metrics"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"run_id":    runID,
			"count":     len(snapshots),
			"snapshots": snapshots,
		})
		return
	}

	// Parse time range for historical query
	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start timestamp format (use RFC3339)"})
			return
		}
	} else {
		// Default to 1 hour ago
		start = time.Now().Add(-1 * time.Hour)
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end timestamp format (use RFC3339)"})
			return
		}
	} else {
		// Default to now
		end = time.Now()
	}

	snapshots, err := h.snapshotRepo.GetSnapshots(runID, start, end)
	if err != nil {
		h.logger.Error("failed to get metrics snapshots", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"run_id":    runID,
		"start":     start,
		"end":       end,
		"count":     len(snapshots),
		"snapshots": snapshots,
	})
}
