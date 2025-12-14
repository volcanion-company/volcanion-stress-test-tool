package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/audit"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditLogger *audit.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditLogger *audit.Logger) *AuditHandler {
	return &AuditHandler{
		auditLogger: auditLogger,
	}
}

// GetAuditLogs retrieves audit logs with optional filters
// GET /api/audit/logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	filter := audit.AuditFilter{
		Limit: 100, // Default limit
	}

	// Parse query parameters
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	if eventType := c.Query("event_type"); eventType != "" {
		filter.EventType = audit.EventType(eventType)
	}

	if ip := c.Query("ip_address"); ip != "" {
		filter.IPAddress = ip
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		filter.ResourceID = resourceID
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	events := h.auditLogger.Query(filter)
	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
		"total":  h.auditLogger.Count(),
	})
}

// ExportAuditLogs exports audit logs as JSON
// GET /api/audit/export
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	filter := audit.AuditFilter{}

	// Parse filters (same as GetAuditLogs)
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	// Export as JSON
	data, err := h.auditLogger.ExportFiltered(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export audit logs"})
		return
	}

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename=audit-logs.json")
	c.Data(http.StatusOK, "application/json", data)
}
