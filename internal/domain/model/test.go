package model

import "time"

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

// TestRun represents an execution instance of a TestPlan
type TestRun struct {
	ID        string        `json:"id"`
	PlanID    string        `json:"plan_id"`
	Status    TestRunStatus `json:"status"`
	StartAt   time.Time     `json:"start_at"`
	EndAt     *time.Time    `json:"end_at,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
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
}

// StartTestRequest represents the request to start a test
type StartTestRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}
