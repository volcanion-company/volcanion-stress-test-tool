package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIClient handles communication with the Volcanion API
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// CreateTestPlan creates a new test plan
func (c *APIClient) CreateTestPlan(plan map[string]interface{}) (string, error) {
	data, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/api/test-plans", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("API error: status %d (failed to read body): %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	idVal, ok := result["id"]
	if !ok {
		return "", errors.New("response missing id field")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return "", errors.New("id field is not a string")
	}

	return idStr, nil
}

// helper to perform GET and decode JSON into target
func (c *APIClient) getJSON(path string, target interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API error: status %d (failed to read body): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("API error: %s", string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

// GetTestPlans fetches all test plans
func (c *APIClient) GetTestPlans() ([]map[string]interface{}, error) {
	var plans []map[string]interface{}
	if err := c.getJSON("/api/test-plans", &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

// StartTest starts a test run from a plan
func (c *APIClient) StartTest(planID string) (string, error) {
	data, err := json.Marshal(map[string]string{
		"test_plan_id": planID,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/api/test-runs/start", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("API error: status %d (failed to read body): %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	idVal, ok := result["id"]
	if !ok {
		return "", errors.New("response missing id field")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return "", errors.New("id field is not a string")
	}

	return idStr, nil
}

// GetTestRuns fetches all test runs
func (c *APIClient) GetTestRuns() ([]map[string]interface{}, error) {
	var runs []map[string]interface{}
	if err := c.getJSON("/api/test-runs", &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

// GetTestRun fetches a specific test run
func (c *APIClient) GetTestRun(runID string) (map[string]interface{}, error) {
	var run map[string]interface{}
	if err := c.getJSON("/api/test-runs/"+runID, &run); err != nil {
		return nil, err
	}
	return run, nil
}

// GetTestRunMetrics fetches metrics for a test run
func (c *APIClient) GetTestRunMetrics(runID string) (map[string]interface{}, error) {
	var metrics map[string]interface{}
	if err := c.getJSON("/api/test-runs/"+runID+"/metrics", &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

// GetLiveMetrics fetches live metrics for a running test
func (c *APIClient) GetLiveMetrics(runID string) (map[string]interface{}, error) {
	var metrics map[string]interface{}
	if err := c.getJSON("/api/test-runs/"+runID+"/live", &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
