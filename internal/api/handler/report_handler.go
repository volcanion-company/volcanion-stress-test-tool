package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/service"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/reporting"
	"go.uber.org/zap"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	testRunService  *service.TestRunService
	testPlanService *service.TestPlanService
	metricsService  *service.MetricsService
	exporter        *reporting.Exporter
	comparator      *reporting.Comparator
	reportStore     *reporting.ReportStore
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	testRunService *service.TestRunService,
	testPlanService *service.TestPlanService,
	metricsService *service.MetricsService,
	reportStore *reporting.ReportStore,
) *ReportHandler {
	return &ReportHandler{
		testRunService:  testRunService,
		testPlanService: testPlanService,
		metricsService:  metricsService,
		exporter:        reporting.NewExporter(),
		comparator:      reporting.NewComparator(),
		reportStore:     reportStore,
	}
}

// ExportTestRunRequest represents export request
type ExportTestRunRequest struct {
	Format string `form:"format" binding:"required,oneof=json csv html"`
}

// ExportTestRun exports a test run in the specified format
// @Summary Export test run
// @Description Export test run results in JSON, CSV, or HTML format
// @Tags reports
// @Param id path string true "Test Run ID"
// @Param format query string true "Export format (json, csv, html)"
// @Success 200 {object} object
// @Router /api/v1/reports/test-runs/{id}/export [get]
func (h *ReportHandler) ExportTestRun(c *gin.Context) {
	runID := c.Param("id")

	var req ExportTestRunRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format parameter"})
		return
	}

	// Get test run
	testRun, err := h.testRunService.GetTestRun(runID)
	if err != nil {
		zap.L().Error("Failed to get test run", zap.Error(err), zap.String("run_id", runID))
		c.JSON(http.StatusNotFound, gin.H{"error": "test run not found"})
		return
	}

	// Get test plan
	testPlan, err := h.testPlanService.GetTestPlan(testRun.PlanID)
	if err != nil {
		zap.L().Error("Failed to get test plan", zap.Error(err), zap.String("plan_id", testRun.PlanID))
		c.JSON(http.StatusNotFound, gin.H{"error": "test plan not found"})
		return
	}

	// Get final metrics if test is completed
	var metrics *model.Metrics
	if testRun.Status == model.StatusCompleted || testRun.Status == model.StatusCancelled {
		metrics, _ = h.metricsService.GetMetrics(runID)
	}

	// Export to buffer
	var buf bytes.Buffer
	format := reporting.ExportFormat(req.Format)
	if err := h.exporter.ExportTestRun(&buf, format, testRun, testPlan, metrics); err != nil {
		zap.L().Error("Failed to export test run", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export test run"})
		return
	}

	// Set appropriate content type and filename
	contentType := h.getContentType(format)
	filename := fmt.Sprintf("test-run-%s-%s.%s", testPlan.Name, testRun.ID[:8], req.Format)

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

// CompareTestRunsRequest represents comparison request
type CompareTestRunsRequest struct {
	BaselineRunID   string `json:"baseline_run_id" binding:"required"`
	ComparisonRunID string `json:"comparison_run_id" binding:"required"`
}

// CompareTestRuns compares two test runs
// @Summary Compare test runs
// @Description Compare metrics between two test runs
// @Tags reports
// @Accept json
// @Produce json
// @Param request body CompareTestRunsRequest true "Comparison request"
// @Success 200 {object} reporting.ComparisonResult
// @Router /api/v1/reports/compare [post]
func (h *ReportHandler) CompareTestRuns(c *gin.Context) {
	var req CompareTestRunsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get baseline run
	baselineRun, err := h.testRunService.GetTestRun(req.BaselineRunID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "baseline run not found"})
		return
	}

	baselineMetrics, err := h.metricsService.GetMetrics(req.BaselineRunID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "baseline run has no metrics"})
		return
	}

	// Get comparison run
	comparisonRun, err := h.testRunService.GetTestRun(req.ComparisonRunID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comparison run not found"})
		return
	}

	compareMetrics, err := h.metricsService.GetMetrics(req.ComparisonRunID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comparison run has no metrics"})
		return
	}

	// Compare
	result, err := h.comparator.Compare(baselineRun, baselineMetrics, comparisonRun, compareMetrics)
	if err != nil {
		zap.L().Error("Failed to compare test runs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compare test runs"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckSLARequest represents SLA check request
type CheckSLARequest struct {
	MinSuccessRate       float64 `json:"min_success_rate" binding:"required,min=0,max=100"`
	MaxAvgResponseTime   float64 `json:"max_avg_response_time" binding:"required,min=0"`
	MaxP95               float64 `json:"max_p95" binding:"required,min=0"`
	MaxP99               float64 `json:"max_p99" binding:"required,min=0"`
	MinRequestsPerSecond float64 `json:"min_requests_per_second" binding:"required,min=0"`
}

// CheckSLA checks if a test run meets SLA criteria
// @Summary Check SLA
// @Description Check if test run meets specified SLA criteria
// @Tags reports
// @Accept json
// @Produce json
// @Param id path string true "Test Run ID"
// @Param criteria body CheckSLARequest true "SLA criteria"
// @Success 200 {object} reporting.SLAResult
// @Router /api/v1/reports/test-runs/{id}/sla [post]
func (h *ReportHandler) CheckSLA(c *gin.Context) {
	runID := c.Param("id")

	var req CheckSLARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get metrics
	metrics, err := h.metricsService.GetMetrics(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metrics not found"})
		return
	}

	// Check SLA
	sla := &reporting.SLACriteria{
		MinSuccessRate:       req.MinSuccessRate,
		MaxAvgResponseTime:   req.MaxAvgResponseTime,
		MaxP95:               req.MaxP95,
		MaxP99:               req.MaxP99,
		MinRequestsPerSecond: req.MinRequestsPerSecond,
	}

	result := h.comparator.IsSLAMet(metrics, sla)
	c.JSON(http.StatusOK, result)
}

// CreateShareableReportRequest represents shareable report creation request
type CreateShareableReportRequest struct {
	TestRunID string `json:"test_run_id" binding:"required"`
	Format    string `json:"format" binding:"required,oneof=json csv html"`
	TTLHours  int    `json:"ttl_hours" binding:"required,min=1,max=168"` // 1 hour to 7 days
	Title     string `json:"title"`
}

// CreateShareableReport creates a shareable report with a unique URL
// @Summary Create shareable report
// @Description Generate a shareable report with a unique URL
// @Tags reports
// @Accept json
// @Produce json
// @Param request body CreateShareableReportRequest true "Report request"
// @Success 201 {object} object
// @Router /api/v1/reports/share [post]
func (h *ReportHandler) CreateShareableReport(c *gin.Context) {
	var req CreateShareableReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get test run
	testRun, err := h.testRunService.GetTestRun(req.TestRunID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test run not found"})
		return
	}

	// Get test plan
	testPlan, err := h.testPlanService.GetTestPlan(testRun.PlanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "test plan not found"})
		return
	}

	// Get final metrics
	var metrics *model.Metrics
	if testRun.Status == model.StatusCompleted || testRun.Status == model.StatusCancelled {
		metrics, _ = h.metricsService.GetMetrics(req.TestRunID)
	}

	// Generate report
	var buf bytes.Buffer
	format := reporting.ExportFormat(req.Format)
	if err := h.exporter.ExportTestRun(&buf, format, testRun, testPlan, metrics); err != nil {
		zap.L().Error("Failed to generate report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	// Store report
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("%s - %s", testPlan.Name, testRun.ID[:8])
	}

	metadata := map[string]string{
		"test_run_id":  testRun.ID,
		"test_plan_id": testPlan.ID,
		"test_name":    testPlan.Name,
	}

	ttl := time.Duration(req.TTLHours) * time.Hour
	report, err := h.reportStore.Store(title, format, buf.Bytes(), ttl, metadata)
	if err != nil {
		zap.L().Error("Failed to store report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store report"})
		return
	}

	// Generate shareable URL
	shareURL := fmt.Sprintf("/api/v1/reports/shared/%s", report.ID)

	c.JSON(http.StatusCreated, gin.H{
		"id":         report.ID,
		"title":      report.Title,
		"format":     report.Format,
		"share_url":  shareURL,
		"expires_at": report.ExpiresAt,
		"created_at": report.CreatedAt,
	})
}

// GetSharedReport retrieves a shared report
// @Summary Get shared report
// @Description Retrieve a shared report by ID
// @Tags reports
// @Param id path string true "Report ID"
// @Success 200 {object} object
// @Router /api/v1/reports/shared/{id} [get]
func (h *ReportHandler) GetSharedReport(c *gin.Context) {
	reportID := c.Param("id")

	report, err := h.reportStore.Get(reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found or expired"})
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", report.ContentType)

	// For HTML, display inline; for others, download
	if report.Format == reporting.FormatHTML {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s.html", report.ID))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", report.ID, report.Format))
	}

	c.Data(http.StatusOK, report.ContentType, report.Content)
}

// ListSharedReports lists all active shared reports
// @Summary List shared reports
// @Description List all active (non-expired) shared reports
// @Tags reports
// @Produce json
// @Success 200 {array} reporting.StoredReport
// @Router /api/v1/reports/shared [get]
func (h *ReportHandler) ListSharedReports(c *gin.Context) {
	reports := h.reportStore.List()
	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"count":   len(reports),
	})
}

// DeleteSharedReport deletes a shared report
// @Summary Delete shared report
// @Description Delete a shared report by ID
// @Tags reports
// @Param id path string true "Report ID"
// @Success 204
// @Router /api/v1/reports/shared/{id} [delete]
func (h *ReportHandler) DeleteSharedReport(c *gin.Context) {
	reportID := c.Param("id")

	if err := h.reportStore.Delete(reportID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetReportStats returns statistics about stored reports
// @Summary Get report statistics
// @Description Get statistics about stored reports
// @Tags reports
// @Produce json
// @Success 200 {object} object
// @Router /api/v1/reports/stats [get]
func (h *ReportHandler) GetReportStats(c *gin.Context) {
	stats := h.reportStore.GetStats()
	c.JSON(http.StatusOK, stats)
}

// getContentType returns the appropriate content type for a format
func (h *ReportHandler) getContentType(format reporting.ExportFormat) string {
	switch format {
	case reporting.FormatJSON:
		return "application/json"
	case reporting.FormatCSV:
		return "text/csv; charset=utf-8"
	case reporting.FormatHTML:
		return "text/html; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
