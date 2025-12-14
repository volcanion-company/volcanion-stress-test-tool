package cmd

import (
	"bytes"
	"encoding/json"
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

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/test-plans",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// GetTestPlans fetches all test plans
func (c *APIClient) GetTestPlans() ([]map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/test-plans")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var plans []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&plans); err != nil {
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

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/test-runs/start",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// GetTestRuns fetches all test runs
func (c *APIClient) GetTestRuns() ([]map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/test-runs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var runs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, err
	}

	return runs, nil
}

// GetTestRun fetches a specific test run
func (c *APIClient) GetTestRun(runID string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/test-runs/" + runID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var run map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return run, nil
}

// GetTestRunMetrics fetches metrics for a test run
func (c *APIClient) GetTestRunMetrics(runID string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/test-runs/" + runID + "/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetLiveMetrics fetches live metrics for a running test
func (c *APIClient) GetLiveMetrics(runID string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/test-runs/" + runID + "/live")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
