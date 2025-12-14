package domain

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

// Validator provides validation logic for domain models
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateTestPlan validates a test plan request
func (v *Validator) ValidateTestPlan(req *model.CreateTestPlanRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return NewValidationError("name", "name is required")
	}

	// Validate URL format
	if strings.TrimSpace(req.TargetURL) == "" {
		return NewValidationError("target_url", "target URL is required")
	}
	if _, err := url.ParseRequestURI(req.TargetURL); err != nil {
		return NewValidationError("target_url", "invalid URL format")
	}

	// Validate HTTP method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[strings.ToUpper(req.Method)] {
		return NewValidationError("method", "invalid HTTP method")
	}

	// Validate positive numbers
	if req.Users <= 0 {
		return NewValidationError("users", "users must be greater than 0")
	}
	if req.Users > 10000 {
		return NewValidationError("users", "users cannot exceed 10000")
	}

	if req.RampUpSec < 0 {
		return NewValidationError("ramp_up_sec", "ramp_up_sec cannot be negative")
	}

	if req.DurationSec <= 0 {
		return NewValidationError("duration_sec", "duration_sec must be greater than 0")
	}
	if req.DurationSec > 86400 {
		return NewValidationError("duration_sec", "duration_sec cannot exceed 86400 (24 hours)")
	}

	if req.TimeoutMs < 0 {
		return NewValidationError("timeout_ms", "timeout_ms cannot be negative")
	}
	if req.TimeoutMs > 300000 {
		return NewValidationError("timeout_ms", "timeout_ms cannot exceed 300000 (5 minutes)")
	}

	if req.TargetRPS < 0 {
		return NewValidationError("target_rps", "target_rps cannot be negative")
	}

	return nil
}

// ValidateStartTestRequest validates a start test request
func (v *Validator) ValidateStartTestRequest(req *model.StartTestRequest) error {
	if strings.TrimSpace(req.PlanID) == "" {
		return NewValidationError("plan_id", "plan_id is required")
	}
	return nil
}

// ValidateSLAConfig validates SLA configuration
func (v *Validator) ValidateSLAConfig(sla *model.SLAConfig) error {
	if sla == nil {
		return nil
	}

	if sla.MaxP95Latency < 0 {
		return NewValidationError("max_p95_latency", "max_p95_latency cannot be negative")
	}

	if sla.MaxP99Latency < 0 {
		return NewValidationError("max_p99_latency", "max_p99_latency cannot be negative")
	}

	if sla.MaxErrorRate < 0 || sla.MaxErrorRate > 100 {
		return NewValidationError("max_error_rate", "max_error_rate must be between 0 and 100")
	}

	if sla.MinRPS < 0 {
		return NewValidationError("min_rps", "min_rps cannot be negative")
	}

	return nil
}

// ValidateRatePattern validates rate pattern configuration
func (v *Validator) ValidateRatePattern(pattern string, steps []model.RateStep) error {
	validPatterns := map[string]bool{
		"fixed": true, "step": true, "ramp": true, "spike": true,
	}

	if !validPatterns[pattern] {
		return NewValidationError("rate_pattern", fmt.Sprintf("invalid rate pattern: %s (must be: fixed, step, ramp, or spike)", pattern))
	}

	if pattern == "step" && len(steps) == 0 {
		return NewValidationError("rate_steps", "rate_steps are required for step pattern")
	}

	for i, step := range steps {
		if step.RPS < 0 {
			return NewValidationError("rate_steps", fmt.Sprintf("step %d: RPS cannot be negative", i))
		}
		if step.DurationSec <= 0 {
			return NewValidationError("rate_steps", fmt.Sprintf("step %d: duration_sec must be greater than 0", i))
		}
	}

	return nil
}
