package model

import "time"

// RatePattern defines the rate control pattern
type RatePattern string

const (
	RatePatternFixed RatePattern = "fixed" // Fixed RPS throughout test
	RatePatternStep  RatePattern = "step"  // Step up RPS in stages
	RatePatternRamp  RatePattern = "ramp"  // Linear ramp up
	RatePatternSpike RatePattern = "spike" // Sudden spike then back to base
)

// RateStep defines a step in step/spike rate patterns
type RateStep struct {
	RPS         int `json:"rps" binding:"min=0"`
	DurationSec int `json:"duration_sec" binding:"min=1"`
}

// SLAConfig defines SLA thresholds for test validation
type SLAConfig struct {
	MaxP95Latency float64 `json:"max_p95_latency,omitempty"` // Max P95 latency in ms
	MaxP99Latency float64 `json:"max_p99_latency,omitempty"` // Max P99 latency in ms
	MaxErrorRate  float64 `json:"max_error_rate,omitempty"`  // Max error rate percentage (0-100)
	MinRPS        float64 `json:"min_rps,omitempty"`         // Minimum RPS to maintain
}

// TestPlan defines the configuration for a stress test
type TestPlan struct {
	ID          string            `json:"id"`
	Name        string            `json:"name" binding:"required"`
	TargetURL   string            `json:"target_url" binding:"required,url"`
	Method      string            `json:"method" binding:"required"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Users       int               `json:"users" binding:"required,min=1"`
	RampUpSec   int               `json:"ramp_up_sec" binding:"min=0"`
	DurationSec int               `json:"duration_sec" binding:"required,min=1"`
	TimeoutMs   int               `json:"timeout_ms" binding:"min=0"`
	TargetRPS   int               `json:"target_rps" binding:"min=0"` // 0 means unlimited
	RatePattern RatePattern       `json:"rate_pattern,omitempty"`     // Default: fixed
	RateSteps   []RateStep        `json:"rate_steps,omitempty"`       // For step/spike patterns
	SLA         *SLAConfig        `json:"sla,omitempty"`              // SLA thresholds
	CreatedAt   time.Time         `json:"created_at"`
}

// TestRunStatus represents the status of a test run
type TestRunStatus string

const (
	StatusRunning   TestRunStatus = "running"
	StatusCompleted TestRunStatus = "completed"
	StatusCancelled TestRunStatus = "cancelled"
	StatusFailed    TestRunStatus = "failed"
)

// StopReason indicates how a test run ended
type StopReason string

const (
	ReasonCompleted StopReason = "completed" // Normal completion
	ReasonCancelled StopReason = "cancelled" // User cancelled
	ReasonFailed    StopReason = "failed"    // Error or SLA violation
)

// TestRun represents an execution instance of a TestPlan
type TestRun struct {
	ID         string        `json:"id"`
	PlanID     string        `json:"plan_id"`
	Status     TestRunStatus `json:"status"`
	StopReason *StopReason   `json:"stop_reason,omitempty"` // How the run ended
	StartAt    time.Time     `json:"start_at"`
	EndAt      *time.Time    `json:"end_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

// CreateTestPlanRequest represents the request to create a test plan
type CreateTestPlanRequest struct {
	Name        string            `json:"name" binding:"required"`
	TargetURL   string            `json:"target_url" binding:"required,url"`
	Method      string            `json:"method" binding:"required"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Users       int               `json:"users" binding:"required,min=1"`
	RampUpSec   int               `json:"ramp_up_sec" binding:"min=0"`
	DurationSec int               `json:"duration_sec" binding:"required,min=1"`
	TimeoutMs   int               `json:"timeout_ms" binding:"min=0"`
	TargetRPS   int               `json:"target_rps" binding:"min=0"`
	RatePattern RatePattern       `json:"rate_pattern,omitempty"`
	RateSteps   []RateStep        `json:"rate_steps,omitempty"`
	SLA         *SLAConfig        `json:"sla,omitempty"`
}

// StartTestRequest represents the request to start a test
type StartTestRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}
