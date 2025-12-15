package service

import (
	"testing"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

func init() {
	// Initialize logger for tests
	if err := logger.Init("error"); err != nil {
		// Best-effort: fail fast during test init
		panic(err)
	}
}

func TestNewScenarioService(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	if service == nil {
		t.Fatal("Expected ScenarioService to be created")
	}
}

func TestCreateScenario(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	req := &model.CreateScenarioRequest{
		Name:        "Login Scenario",
		Description: "Test user login flow",
		Steps: []model.Step{
			{
				Name:    "Login",
				URL:     "http://localhost:8080/api/login",
				Method:  "POST",
				Body:    `{"username": "test", "password": "test123"}`,
				Headers: map[string]string{"Content-Type": "application/json"},
			},
			{
				Name:    "Get Profile",
				URL:     "http://localhost:8080/api/profile",
				Method:  "GET",
				Headers: map[string]string{"Authorization": "Bearer {{token}}"},
			},
		},
		Variables: model.Variables{
			"baseUrl": "http://localhost:8080",
		},
	}

	scenario, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	if scenario.ID == "" {
		t.Error("Expected scenario ID to be generated")
	}
	if scenario.Name != req.Name {
		t.Errorf("Name mismatch: expected %s, got %s", req.Name, scenario.Name)
	}
	if len(scenario.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(scenario.Steps))
	}
}

func TestGetScenario(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	req := &model.CreateScenarioRequest{
		Name:        "Test Scenario",
		Description: "A test scenario",
		Steps: []model.Step{
			{
				Name:   "Step 1",
				URL:    "http://localhost:8080/api/test",
				Method: "GET",
			},
		},
	}

	created, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	retrieved, err := service.GetScenario(created.ID)
	if err != nil {
		t.Fatalf("Failed to get scenario: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID mismatch: expected %s, got %s", created.ID, retrieved.ID)
	}
	if retrieved.Name != created.Name {
		t.Errorf("Name mismatch: expected %s, got %s", created.Name, retrieved.Name)
	}
}

func TestGetScenarioNotFound(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	_, err := service.GetScenario("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent scenario")
	}
}

func TestGetAllScenarios(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	// Create multiple scenarios
	for i := 0; i < 3; i++ {
		req := &model.CreateScenarioRequest{
			Name:        "Scenario " + string(rune('A'+i)),
			Description: "Test scenario " + string(rune('A'+i)),
			Steps: []model.Step{
				{
					Name:   "Step 1",
					URL:    "http://localhost:8080/api/test",
					Method: "GET",
				},
			},
		}
		_, err := service.CreateScenario(req)
		if err != nil {
			t.Fatalf("Failed to create scenario: %v", err)
		}
	}

	scenarios, err := service.GetAllScenarios()
	if err != nil {
		t.Fatalf("Failed to get all scenarios: %v", err)
	}

	if len(scenarios) != 3 {
		t.Errorf("Expected 3 scenarios, got %d", len(scenarios))
	}
}

func TestDeleteScenario(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	req := &model.CreateScenarioRequest{
		Name:        "Delete Test",
		Description: "A scenario to delete",
		Steps: []model.Step{
			{
				Name:   "Step 1",
				URL:    "http://localhost:8080/api/test",
				Method: "GET",
			},
		},
	}

	scenario, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	// Delete the scenario
	err = service.DeleteScenario(scenario.ID)
	if err != nil {
		t.Fatalf("Failed to delete scenario: %v", err)
	}

	// Verify it's deleted
	_, err = service.GetScenario(scenario.ID)
	if err == nil {
		t.Error("Expected error when getting deleted scenario")
	}
}

func TestCreateScenarioTimestamp(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	beforeCreate := time.Now()

	req := &model.CreateScenarioRequest{
		Name:        "Timestamp Test",
		Description: "Test timestamp",
		Steps: []model.Step{
			{
				Name:   "Step 1",
				URL:    "http://localhost:8080",
				Method: "GET",
			},
		},
	}

	scenario, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	afterCreate := time.Now()

	if scenario.CreatedAt.Before(beforeCreate) || scenario.CreatedAt.After(afterCreate) {
		t.Errorf("CreatedAt timestamp %v not within expected range", scenario.CreatedAt)
	}
}

func TestCreateScenarioWithVariables(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	req := &model.CreateScenarioRequest{
		Name:        "Variables Test",
		Description: "Scenario with variables",
		Steps: []model.Step{
			{
				Name:   "Step 1",
				URL:    "{{baseUrl}}/api/test",
				Method: "GET",
			},
		},
		Variables: model.Variables{
			"baseUrl":  "http://localhost:8080",
			"apiKey":   "test-key-123",
			"username": "testuser",
		},
	}

	scenario, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	if len(scenario.Variables) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(scenario.Variables))
	}
	if scenario.Variables["baseUrl"] != "http://localhost:8080" {
		t.Errorf("baseUrl variable mismatch")
	}
}

func TestCreateScenarioMultipleSteps(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	req := &model.CreateScenarioRequest{
		Name:        "Multi-Step Scenario",
		Description: "Scenario with multiple steps",
		Steps: []model.Step{
			{
				Name:   "Login",
				URL:    "http://localhost:8080/api/login",
				Method: "POST",
				Body:   `{"username": "admin", "password": "admin123"}`,
			},
			{
				Name:   "List Users",
				URL:    "http://localhost:8080/api/users",
				Method: "GET",
			},
			{
				Name:   "Create User",
				URL:    "http://localhost:8080/api/users",
				Method: "POST",
				Body:   `{"name": "New User", "email": "new@example.com"}`,
			},
			{
				Name:   "Delete User",
				URL:    "http://localhost:8080/api/users/123",
				Method: "DELETE",
			},
			{
				Name:   "Logout",
				URL:    "http://localhost:8080/api/logout",
				Method: "POST",
			},
		},
	}

	scenario, err := service.CreateScenario(req)
	if err != nil {
		t.Fatalf("Failed to create scenario: %v", err)
	}

	if len(scenario.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(scenario.Steps))
	}

	expectedNames := []string{"Login", "List Users", "Create User", "Delete User", "Logout"}
	for i, step := range scenario.Steps {
		if step.Name != expectedNames[i] {
			t.Errorf("Step %d name mismatch: expected %s, got %s", i, expectedNames[i], step.Name)
		}
	}
}

func TestDeleteNonExistentScenario(t *testing.T) {
	scenarioRepo := repository.NewMemoryScenarioRepository()
	executionRepo := repository.NewMemoryScenarioExecutionRepository()

	service := NewScenarioService(scenarioRepo, executionRepo, nil)

	err := service.DeleteScenario("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent scenario")
	}
}
