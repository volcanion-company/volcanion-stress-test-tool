package audit

import (
	"time"
)

// EventType represents the type of audit event
type EventType string

const (
	EventTestPlanCreated    EventType = "test_plan.created"
	EventTestPlanDeleted    EventType = "test_plan.deleted"
	EventTestRunStarted     EventType = "test_run.started"
	EventTestRunStopped     EventType = "test_run.stopped"
	EventScenarioCreated    EventType = "scenario.created"
	EventScenarioExecuted   EventType = "scenario.executed"
	EventScenarioDeleted    EventType = "scenario.deleted"
	EventAPIKeyCreated      EventType = "api_key.created"
	EventAPIKeyRevoked      EventType = "api_key.revoked"
	EventUserLogin          EventType = "user.login"
	EventUnauthorizedAccess EventType = "security.unauthorized_access"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	EventType  EventType              `json:"event_type"`
	UserID     string                 `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Role       string                 `json:"role,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	StatusCode int                    `json:"status_code"`
	Duration   int64                  `json:"duration_ms"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// AuditFilter represents filter criteria for querying audit logs
type AuditFilter struct {
	StartTime  *time.Time
	EndTime    *time.Time
	UserID     string
	EventType  EventType
	IPAddress  string
	ResourceID string
	Limit      int
	Offset     int
}
