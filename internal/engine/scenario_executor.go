package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

// ScenarioExecutor executes multi-step scenarios
type ScenarioExecutor struct {
	client *http.Client
}

const (
	statusSkipped = "skipped"
	statusFailed  = "failed"
	statusSuccess = "success"
)

// NewScenarioExecutor creates a new scenario executor
func NewScenarioExecutor() *ScenarioExecutor {
	// Create shared HTTP client with optimized transport
	sharedTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	return &ScenarioExecutor{
		client: &http.Client{
			Transport: sharedTransport,
		},
	}
}

// Execute runs a scenario and returns the execution result
func (e *ScenarioExecutor) Execute(scenario *model.Scenario, initialVars model.Variables) (*model.ScenarioExecution, error) {
	execution := &model.ScenarioExecution{
		ID:          generateExecutionID(),
		ScenarioID:  scenario.ID,
		Status:      model.StatusRunning,
		StartedAt:   time.Now(),
		StepResults: make([]model.StepResult, 0),
		Variables:   make(model.Variables),
	}

	// Initialize variables
	for k, v := range scenario.Variables {
		execution.Variables[k] = v
	}
	for k, v := range initialVars {
		execution.Variables[k] = v // Override with initial vars
	}

	logger.Log.Info("Starting scenario execution",
		zap.String("scenario_id", scenario.ID),
		zap.String("execution_id", execution.ID),
		zap.Int("steps", len(scenario.Steps)))

	// Execute each step
	for i, step := range scenario.Steps {
		logger.Log.Debug("Executing step",
			zap.Int("step_index", i+1),
			zap.String("step_name", step.Name))

		stepResult, err := e.executeStep(&step, execution.Variables)
		execution.StepResults = append(execution.StepResults, *stepResult)

		if err != nil {
			execution.Status = model.StatusFailed
			execution.Error = fmt.Sprintf("Step '%s' failed: %v", step.Name, err)
			now := time.Now()
			execution.CompletedAt = &now
			return execution, err
		}

		// If step was skipped, continue to next
		if stepResult.Skipped {
			logger.Log.Info("Step skipped", zap.String("step_name", step.Name))
			continue
		}

		// Check if any assertions failed
		if len(stepResult.AssertionsFailed) > 0 {
			execution.Status = model.StatusFailed
			execution.Error = fmt.Sprintf("Step '%s' assertions failed: %v", step.Name, stepResult.AssertionsFailed)
			now := time.Now()
			execution.CompletedAt = &now
			return execution, fmt.Errorf("assertions failed: %v", stepResult.AssertionsFailed)
		}

		// Think time between steps
		if step.ThinkTimeMs > 0 {
			time.Sleep(time.Duration(step.ThinkTimeMs) * time.Millisecond)
		}
	}

	// All steps completed successfully
	execution.Status = model.StatusCompleted
	now := time.Now()
	execution.CompletedAt = &now

	logger.Log.Info("Scenario execution completed",
		zap.String("scenario_id", scenario.ID),
		zap.String("execution_id", execution.ID),
		zap.String("status", string(execution.Status)))

	return execution, nil
}

// executeStep executes a single step and returns the result
func (e *ScenarioExecutor) executeStep(step *model.Step, vars model.Variables) (*model.StepResult, error) {
	result := &model.StepResult{
		StepName:    step.Name,
		ExecutedAt:  time.Now(),
		Extractions: make(map[string]interface{}),
	}

	// Check skip condition
	if step.SkipIf != nil {
		if e.evaluateCondition(step.SkipIf, vars) {
			result.Status = statusSkipped
			result.Skipped = true
			return result, nil
		}
	}

	// Apply variable substitution to URL, headers, and body
	url := e.substituteVariables(step.URL, vars)
	headers := make(map[string]string)
	for k, v := range step.Headers {
		headers[k] = e.substituteVariables(v, vars)
	}
	body := e.substituteVariables(step.Body, vars)

	// Create HTTP request
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(step.Method, url, bodyReader)
	if err != nil {
		result.Status = statusFailed
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set timeout
	timeout := 30 * time.Second
	if step.TimeoutMs > 0 {
		timeout = time.Duration(step.TimeoutMs) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request
	startTime := time.Now()
	resp, err := e.client.Do(req)
	responseTime := time.Since(startTime).Milliseconds()
	result.ResponseTimeMs = float64(responseTime)

	if err != nil {
		result.Status = statusFailed
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Status = statusFailed
		result.Error = fmt.Sprintf("failed to read response: %v", err)
		return result, err
	}

	result.StatusCode = resp.StatusCode

	// Extract variables
	if len(step.Extractions) > 0 {
		for _, extraction := range step.Extractions {
			value, err := e.extractVariable(&extraction, resp, responseBody)
			if err != nil {
				logger.Log.Warn("Failed to extract variable",
					zap.String("variable", extraction.Name),
					zap.Error(err))
				continue
			}
			vars[extraction.Name] = value
			result.Extractions[extraction.Name] = value
		}
	}

	// Run assertions
	result.AssertionsFailed = make([]string, 0)
	for _, assertion := range step.Assertions {
		if !e.evaluateAssertion(&assertion, resp, responseBody, responseTime) {
			failMsg := fmt.Sprintf("%s %s %v", assertion.Type, assertion.Operator, assertion.Value)
			result.AssertionsFailed = append(result.AssertionsFailed, failMsg)
		}
	}

	if len(result.AssertionsFailed) > 0 {
		result.Status = statusFailed
	} else {
		result.Status = statusSuccess
	}

	return result, nil
}

// substituteVariables replaces {{variable}} placeholders with values
func (e *ScenarioExecutor) substituteVariables(template string, vars model.Variables) string {
	result := template
	for k, v := range vars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		replacement := fmt.Sprintf("%v", v)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// extractVariable extracts a value from the response
func (e *ScenarioExecutor) extractVariable(extraction *model.VariableExtraction, resp *http.Response, body []byte) (interface{}, error) {
	switch extraction.Type {
	case model.ExtractionHeader:
		return resp.Header.Get(extraction.Path), nil

	case model.ExtractionStatus:
		return resp.StatusCode, nil

	case model.ExtractionRegex:
		re, err := regexp.Compile(extraction.Path)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		matches := re.FindStringSubmatch(string(body))
		if len(matches) > 1 {
			return matches[1], nil // Return first capturing group
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
		return nil, fmt.Errorf("regex pattern not found")

	case model.ExtractionJSONPath:
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("invalid JSON response: %w", err)
		}
		// Simple JSONPath implementation (supports basic dot notation)
		value := e.extractJSONPath(data, extraction.Path)
		if value == nil {
			return nil, fmt.Errorf("JSONPath not found: %s", extraction.Path)
		}
		return value, nil

	default:
		return nil, fmt.Errorf("unsupported extraction type: %s", extraction.Type)
	}
}

// extractJSONPath extracts value using simple dot notation (e.g., "data.id")
func (e *ScenarioExecutor) extractJSONPath(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}

	return current
}

// evaluateAssertion checks if an assertion passes
func (e *ScenarioExecutor) evaluateAssertion(assertion *model.Assertion, resp *http.Response, body []byte, responseTimeMs int64) bool {
	switch assertion.Type {
	case model.AssertionStatusCode:
		expected, ok := assertion.Value.(float64) // JSON numbers are float64
		if !ok {
			return false
		}
		return resp.StatusCode == int(expected)

	case model.AssertionResponseTime:
		maxTime, ok := assertion.Value.(float64)
		if !ok {
			return false
		}
		return responseTimeMs <= int64(maxTime)

	case model.AssertionBodyContains:
		expected, ok := assertion.Value.(string)
		if !ok {
			return false
		}
		return strings.Contains(string(body), expected)

	case model.AssertionHeader:
		headerValue := resp.Header.Get(assertion.Target)
		expected := fmt.Sprintf("%v", assertion.Value)
		return e.compareValues(headerValue, assertion.Operator, expected)

	case model.AssertionJSONPath:
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return false
		}
		value := e.extractJSONPath(data, assertion.Target)
		expected := assertion.Value
		return e.compareValues(value, assertion.Operator, expected)

	default:
		return false
	}
}

// evaluateCondition checks if a condition is met
func (e *ScenarioExecutor) evaluateCondition(cond *model.Condition, vars model.Variables) bool {
	value, exists := vars[cond.Variable]

	switch cond.Operator {
	case "exists":
		return exists
	case "not_exists":
		return !exists
	case "eq":
		if !exists {
			return false
		}
		return e.compareValues(value, "eq", cond.Value)
	case "ne":
		if !exists {
			return true
		}
		return !e.compareValues(value, "eq", cond.Value)
	default:
		return false
	}
}

// compareValues compares two values using an operator
func (e *ScenarioExecutor) compareValues(left interface{}, operator string, right interface{}) bool {
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch operator {
	case "eq":
		return leftStr == rightStr
	case "ne":
		return leftStr != rightStr
	case "contains":
		return strings.Contains(leftStr, rightStr)
	case "gt":
		return leftStr > rightStr
	case "lt":
		return leftStr < rightStr
	default:
		return false
	}
}

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
