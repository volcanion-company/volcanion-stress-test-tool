package model

import "time"

// Scenario represents a multi-step test workflow
type Scenario struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description,omitempty"`
	Steps       []Step    `json:"steps" binding:"required,min=1"`
	Variables   Variables `json:"variables,omitempty"` // Global variables
	CreatedAt   time.Time `json:"created_at"`
}

// Step represents a single step in a scenario
type Step struct {
	Name        string               `json:"name" binding:"required"`
	Method      string               `json:"method" binding:"required"`
	URL         string               `json:"url" binding:"required"`
	Headers     map[string]string    `json:"headers,omitempty"`
	Body        string               `json:"body,omitempty"`
	TimeoutMs   int                  `json:"timeout_ms,omitempty"`
	Extractions []VariableExtraction `json:"extractions,omitempty"` // Extract variables from response
	Assertions  []Assertion          `json:"assertions,omitempty"`  // Validate response
	SkipIf      *Condition           `json:"skip_if,omitempty"`     // Conditional execution
	ThinkTimeMs int                  `json:"think_time_ms,omitempty"`
}

// VariableExtraction defines how to extract a value from response
type VariableExtraction struct {
	Name   string         `json:"name" binding:"required"`   // Variable name to store
	Source string         `json:"source" binding:"required"` // "body", "header", "status"
	Type   ExtractionType `json:"type" binding:"required"`   // "jsonpath", "regex", "header"
	Path   string         `json:"path" binding:"required"`   // JSONPath, regex pattern, or header name
}

// ExtractionType defines how to extract values
type ExtractionType string

const (
	ExtractionJSONPath ExtractionType = "jsonpath"
	ExtractionRegex    ExtractionType = "regex"
	ExtractionHeader   ExtractionType = "header"
	ExtractionStatus   ExtractionType = "status"
)

// Assertion validates a response
type Assertion struct {
	Type     AssertionType `json:"type" binding:"required"`
	Target   string        `json:"target,omitempty"`   // JSONPath or header name
	Operator string        `json:"operator,omitempty"` // "eq", "contains", "gt", "lt"
	Value    interface{}   `json:"value,omitempty"`
}

// AssertionType defines what to assert
type AssertionType string

const (
	AssertionStatusCode   AssertionType = "status_code"
	AssertionResponseTime AssertionType = "response_time"
	AssertionJSONPath     AssertionType = "jsonpath"
	AssertionHeader       AssertionType = "header"
	AssertionBodyContains AssertionType = "body_contains"
)

// Condition defines a conditional execution rule
type Condition struct {
	Variable string      `json:"variable" binding:"required"`
	Operator string      `json:"operator" binding:"required"` // "eq", "ne", "exists", "not_exists"
	Value    interface{} `json:"value,omitempty"`
}

// Variables stores extracted variables during scenario execution
type Variables map[string]interface{}

// ScenarioExecution tracks the execution of a scenario
type ScenarioExecution struct {
	ID          string        `json:"id"`
	ScenarioID  string        `json:"scenario_id"`
	Status      TestRunStatus `json:"status"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	StepResults []StepResult  `json:"step_results"`
	Variables   Variables     `json:"variables"`
	Error       string        `json:"error,omitempty"`
}

// StepResult holds the result of a single step execution
type StepResult struct {
	StepName         string                 `json:"step_name"`
	Status           string                 `json:"status"` // "success", "failed", "skipped"
	StatusCode       int                    `json:"status_code,omitempty"`
	ResponseTimeMs   float64                `json:"response_time_ms"`
	Extractions      map[string]interface{} `json:"extractions,omitempty"`
	AssertionsFailed []string               `json:"assertions_failed,omitempty"`
	Error            string                 `json:"error,omitempty"`
	Skipped          bool                   `json:"skipped"`
	ExecutedAt       time.Time              `json:"executed_at"`
}

// CreateScenarioRequest represents a request to create a scenario
type CreateScenarioRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description,omitempty"`
	Steps       []Step    `json:"steps" binding:"required,min=1"`
	Variables   Variables `json:"variables,omitempty"`
}

// ExecuteScenarioRequest represents a request to execute a scenario
type ExecuteScenarioRequest struct {
	ScenarioID string    `json:"scenario_id" binding:"required"`
	Variables  Variables `json:"variables,omitempty"` // Override initial variables
}
