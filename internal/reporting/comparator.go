package reporting

import (
	"fmt"
	"math"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

// ComparisonResult represents the comparison between two test runs
type ComparisonResult struct {
	BaselineRun     *model.TestRun     `json:"baseline_run"`
	ComparisonRun   *model.TestRun     `json:"comparison_run"`
	BaselineMetrics *model.Metrics     `json:"baseline_metrics"`
	CompareMetrics  *model.Metrics     `json:"compare_metrics"`
	Differences     *MetricDifferences `json:"differences"`
	Summary         string             `json:"summary"`
	ComparedAt      time.Time          `json:"compared_at"`
}

// MetricDifferences represents the differences between two test runs
type MetricDifferences struct {
	TotalRequests      DiffValue `json:"total_requests"`
	SuccessfulRequests DiffValue `json:"successful_requests"`
	FailedRequests     DiffValue `json:"failed_requests"`
	SuccessRate        DiffValue `json:"success_rate"`
	AvgResponseTime    DiffValue `json:"avg_response_time"`
	MinResponseTime    DiffValue `json:"min_response_time"`
	MaxResponseTime    DiffValue `json:"max_response_time"`
	P50                DiffValue `json:"p50"`
	P75                DiffValue `json:"p75"`
	P95                DiffValue `json:"p95"`
	P99                DiffValue `json:"p99"`
	RequestsPerSecond  DiffValue `json:"requests_per_second"`
}

// DiffValue represents a difference between two metric values
type DiffValue struct {
	Baseline       float64 `json:"baseline"`
	Comparison     float64 `json:"comparison"`
	AbsoluteDiff   float64 `json:"absolute_diff"`
	PercentageDiff float64 `json:"percentage_diff"`
	Improved       bool    `json:"improved"`
	Degraded       bool    `json:"degraded"`
}

// Comparator handles comparison between test runs
type Comparator struct{}

// NewComparator creates a new Comparator
func NewComparator() *Comparator {
	return &Comparator{}
}

// Compare compares two test runs and their metrics
func (c *Comparator) Compare(baselineRun *model.TestRun, baselineMetrics *model.Metrics,
	comparisonRun *model.TestRun, compareMetrics *model.Metrics) (*ComparisonResult, error) {

	if baselineMetrics == nil || compareMetrics == nil {
		return nil, fmt.Errorf("both test runs must have metrics")
	}

	differences := c.calculateDifferences(baselineMetrics, compareMetrics)
	summary := c.generateSummary(differences)

	return &ComparisonResult{
		BaselineRun:     baselineRun,
		ComparisonRun:   comparisonRun,
		BaselineMetrics: baselineMetrics,
		CompareMetrics:  compareMetrics,
		Differences:     differences,
		Summary:         summary,
		ComparedAt:      time.Now(),
	}, nil
}

// calculateDifferences calculates all metric differences
func (c *Comparator) calculateDifferences(baseline, comparison *model.Metrics) *MetricDifferences {
	// Calculate success rates
	baselineSuccessRate := float64(0)
	if baseline.TotalRequests > 0 {
		baselineSuccessRate = float64(baseline.SuccessRequests) / float64(baseline.TotalRequests) * 100
	}
	compareSuccessRate := float64(0)
	if comparison.TotalRequests > 0 {
		compareSuccessRate = float64(comparison.SuccessRequests) / float64(comparison.TotalRequests) * 100
	}

	return &MetricDifferences{
		TotalRequests:      c.diff(float64(baseline.TotalRequests), float64(comparison.TotalRequests), true),
		SuccessfulRequests: c.diff(float64(baseline.SuccessRequests), float64(comparison.SuccessRequests), true),
		FailedRequests:     c.diff(float64(baseline.FailedRequests), float64(comparison.FailedRequests), false),
		SuccessRate:        c.diff(baselineSuccessRate, compareSuccessRate, true),
		AvgResponseTime:    c.diff(baseline.AvgLatencyMs, comparison.AvgLatencyMs, false),
		MinResponseTime:    c.diff(baseline.MinLatencyMs, comparison.MinLatencyMs, false),
		MaxResponseTime:    c.diff(baseline.MaxLatencyMs, comparison.MaxLatencyMs, false),
		P50:                c.diff(baseline.P50LatencyMs, comparison.P50LatencyMs, false),
		P75:                c.diff(baseline.P75LatencyMs, comparison.P75LatencyMs, false),
		P95:                c.diff(baseline.P95LatencyMs, comparison.P95LatencyMs, false),
		P99:                c.diff(baseline.P99LatencyMs, comparison.P99LatencyMs, false),
		RequestsPerSecond:  c.diff(baseline.RequestsPerSec, comparison.RequestsPerSec, true),
	}
}

// diff calculates difference between two values
// higherIsBetter indicates whether an increase is an improvement
func (c *Comparator) diff(baseline, comparison float64, higherIsBetter bool) DiffValue {
	absoluteDiff := comparison - baseline
	percentageDiff := float64(0)
	if baseline != 0 {
		percentageDiff = (absoluteDiff / baseline) * 100
	}

	// Determine improvement/degradation
	improved := false
	degraded := false
	if higherIsBetter {
		improved = absoluteDiff > 0
		degraded = absoluteDiff < 0
	} else {
		improved = absoluteDiff < 0
		degraded = absoluteDiff > 0
	}

	return DiffValue{
		Baseline:       baseline,
		Comparison:     comparison,
		AbsoluteDiff:   absoluteDiff,
		PercentageDiff: percentageDiff,
		Improved:       improved,
		Degraded:       degraded,
	}
}

// generateSummary generates a human-readable summary of the comparison
func (c *Comparator) generateSummary(diff *MetricDifferences) string {
	improvements := 0
	degradations := 0

	metrics := []DiffValue{
		diff.SuccessRate,
		diff.AvgResponseTime,
		diff.P95,
		diff.P99,
		diff.RequestsPerSecond,
	}

	for _, metric := range metrics {
		if metric.Improved {
			improvements++
		}
		if metric.Degraded {
			degradations++
		}
	}

	if improvements > degradations {
		return "Overall Performance Improved"
	} else if degradations > improvements {
		return "Overall Performance Degraded"
	}
	return "Overall Performance Similar"
}

// CompareMultiple compares multiple test runs against a baseline
func (c *Comparator) CompareMultiple(baseline *model.TestRun, baselineMetrics *model.Metrics,
	runs []*model.TestRun, metrics []*model.Metrics) ([]*ComparisonResult, error) {

	if len(runs) != len(metrics) {
		return nil, fmt.Errorf("runs and metrics length mismatch")
	}

	results := make([]*ComparisonResult, 0, len(runs))
	for i, run := range runs {
		result, err := c.Compare(baseline, baselineMetrics, run, metrics[i])
		if err != nil {
			return nil, fmt.Errorf("failed to compare run %s: %w", run.ID, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// PerformanceScore calculates an overall performance score (0-100)
func (c *Comparator) PerformanceScore(metrics *model.Metrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0
	}

	// Success rate weight: 40%
	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests)
	successScore := successRate * 40

	// Response time score (lower is better): 40%
	// Normalize response time (assume 1000ms is poor, 100ms is excellent)
	avgRTScore := 0.0
	switch {
	case metrics.AvgLatencyMs <= 100:
		avgRTScore = 40
	case metrics.AvgLatencyMs >= 1000:
		avgRTScore = 0
	default:
		avgRTScore = 40 * (1 - (metrics.AvgLatencyMs-100)/900)
	}

	// P95 score (lower is better): 20%
	p95Score := 0.0
	switch {
	case metrics.P95LatencyMs <= 200:
		p95Score = 20
	case metrics.P95LatencyMs >= 2000:
		p95Score = 0
	default:
		p95Score = 20 * (1 - (metrics.P95LatencyMs-200)/1800)
	}

	totalScore := successScore + avgRTScore + p95Score
	return math.Round(totalScore*100) / 100
}

// IsSLAMet checks if test run meets specified SLA criteria
func (c *Comparator) IsSLAMet(metrics *model.Metrics, sla *SLACriteria) *SLAResult {
	violations := []string{}

	// Check success rate
	successRate := float64(0)
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	}
	if successRate < sla.MinSuccessRate {
		violations = append(violations, fmt.Sprintf("Success rate %.2f%% below minimum %.2f%%", successRate, sla.MinSuccessRate))
	}

	// Check average response time
	if metrics.AvgLatencyMs > sla.MaxAvgResponseTime {
		violations = append(violations, fmt.Sprintf("Avg response time %.2fms exceeds maximum %.2fms", metrics.AvgLatencyMs, sla.MaxAvgResponseTime))
	}

	// Check P95
	if metrics.P95LatencyMs > sla.MaxP95 {
		violations = append(violations, fmt.Sprintf("P95 %.2fms exceeds maximum %.2fms", metrics.P95LatencyMs, sla.MaxP95))
	}

	// Check P99
	if metrics.P99LatencyMs > sla.MaxP99 {
		violations = append(violations, fmt.Sprintf("P99 %.2fms exceeds maximum %.2fms", metrics.P99LatencyMs, sla.MaxP99))
	}

	// Check minimum throughput
	if metrics.RequestsPerSec < sla.MinRequestsPerSecond {
		violations = append(violations, fmt.Sprintf("Throughput %.2f req/s below minimum %.2f req/s", metrics.RequestsPerSec, sla.MinRequestsPerSecond))
	}

	return &SLAResult{
		Met:        len(violations) == 0,
		Violations: violations,
		CheckedAt:  time.Now(),
	}
}

// SLACriteria defines SLA thresholds
type SLACriteria struct {
	MinSuccessRate       float64 `json:"min_success_rate"`        // Minimum success rate percentage
	MaxAvgResponseTime   float64 `json:"max_avg_response_time"`   // Maximum average response time in ms
	MaxP95               float64 `json:"max_p95"`                 // Maximum P95 response time in ms
	MaxP99               float64 `json:"max_p99"`                 // Maximum P99 response time in ms
	MinRequestsPerSecond float64 `json:"min_requests_per_second"` // Minimum throughput
}

// SLAResult represents the result of SLA checking
type SLAResult struct {
	Met        bool      `json:"met"`
	Violations []string  `json:"violations"`
	CheckedAt  time.Time `json:"checked_at"`
}
