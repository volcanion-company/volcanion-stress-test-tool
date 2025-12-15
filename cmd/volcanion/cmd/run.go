package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	planFile   string
	planID     string
	watch      bool
	outputFile string
	noColor    bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a stress test",
	Long: `Run a stress test from a plan file or existing plan ID.
	
Examples:
  # Run from YAML file
  volcanion run -f plan.yaml
  
  # Run from JSON file
  volcanion run -f plan.json
  
  # Run existing plan by ID
  volcanion run --plan-id abc123
  
  # Run and watch live metrics
  volcanion run -f plan.yaml --watch
  
  # Run and save results to file
  volcanion run -f plan.yaml -o results.json`,
	RunE: runTest,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&planFile, "file", "f", "", "test plan file (YAML or JSON)")
	runCmd.Flags().StringVar(&planID, "plan-id", "", "existing test plan ID")
	runCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch live metrics during test")
	runCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file for results (JSON)")
	runCmd.Flags().BoolVar(&noColor, "no-color", false, "disable colored output")

	if err := runCmd.MarkFlagFilename("file", "yaml", "yml", "json"); err != nil {
		panic(err)
	}
}

func runTest(_ *cobra.Command, _ []string) error {
	if noColor {
		color.NoColor = true
	}

	// Validate flags
	if planFile == "" && planID == "" {
		return fmt.Errorf("either --file or --plan-id must be specified")
	}

	if planFile != "" && planID != "" {
		return fmt.Errorf("cannot specify both --file and --plan-id")
	}

	client := NewAPIClient(GetAPIBaseURL())

	var testRunID string

	if planFile != "" {
		// Load plan from file and create
		var err error
		testRunID, err = runFromFile(client, planFile)
		if err != nil {
			return err
		}
	} else {
		// Run existing plan
		var err error
		testRunID, err = runFromPlanID(client, planID)
		if err != nil {
			return err
		}
	}

	printSuccess(fmt.Sprintf("Test started successfully! Run ID: %s", testRunID))

	// Watch live metrics if requested
	if watch {
		if err := watchTest(client, testRunID); err != nil {
			return err
		}
	} else {
		// Just wait for completion
		if err := waitForCompletion(client, testRunID); err != nil {
			return err
		}
	}

	// Fetch final results
	results, err := client.GetTestRunMetrics(testRunID)
	if err != nil {
		return fmt.Errorf("failed to fetch results: %w", err)
	}

	// Print summary
	printTestSummary(results)

	// Save to file if requested
	if outputFile != "" {
		if err := saveResults(results, outputFile); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
		printSuccess(fmt.Sprintf("Results saved to %s", outputFile))
	}

	return nil
}

func runFromFile(client *APIClient, filename string) (string, error) {
	printInfo(fmt.Sprintf("Loading test plan from %s...", filename))

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse based on extension
	var plan map[string]interface{}
	ext := filepath.Ext(filename)

	switch ext {
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, &plan); err != nil {
			return "", fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err = json.Unmarshal(data, &plan); err != nil {
			return "", fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported file format: %s (use .yaml, .yml, or .json)", ext)
	}

	// Create test plan
	printInfo("Creating test plan...")
	planID, err := client.CreateTestPlan(plan)
	if err != nil {
		return "", fmt.Errorf("failed to create test plan: %w", err)
	}

	printSuccess(fmt.Sprintf("Test plan created: %s", planID))

	// Start test
	printInfo("Starting test...")
	runID, err := client.StartTest(planID)
	if err != nil {
		return "", fmt.Errorf("failed to start test: %w", err)
	}

	return runID, nil
}

func runFromPlanID(client *APIClient, planID string) (string, error) {
	printInfo(fmt.Sprintf("Starting test from plan %s...", planID))

	runID, err := client.StartTest(planID)
	if err != nil {
		return "", fmt.Errorf("failed to start test: %w", err)
	}

	return runID, nil
}

func waitForCompletion(client *APIClient, runID string) error {
	printInfo("Waiting for test completion...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		run, err := client.GetTestRun(runID)
		if err != nil {
			return fmt.Errorf("failed to check test status: %w", err)
		}

		status := run["status"].(string)
		//nolint:misspell // domain uses British spelling 'cancelled' for stored status values
		if status == StatusCompleted || status == StatusFailed || status == StatusCanceled {
			printSuccess(fmt.Sprintf("Test %s", status))
			return nil
		}

		if IsVerbose() {
			printInfo(fmt.Sprintf("Test status: %s", status))
		}
	}
}

func saveResults(results map[string]interface{}, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o600)
}

func printTestSummary(results map[string]interface{}) {
	fmt.Println()
	printHeader("Test Results Summary")
	fmt.Println()

	// Extract metrics
	totalReqs := int(results["total_requests"].(float64))
	successReqs := int(results["success_requests"].(float64))
	failedReqs := int(results["failed_requests"].(float64))
	avgLatency := results["avg_latency_ms"].(float64)
	p95Latency := results["p95_latency_ms"].(float64)
	p99Latency := results["p99_latency_ms"].(float64)
	rps := results["requests_per_sec"].(float64)

	successRate := float64(successReqs) / float64(totalReqs) * 100

	// Print metrics
	fmt.Printf("  Total Requests:    %s\n", color.CyanString("%d", totalReqs))
	fmt.Printf("  Successful:        %s (%s)\n",
		color.GreenString("%d", successReqs),
		color.GreenString("%.2f%%", successRate))
	fmt.Printf("  Failed:            %s\n", color.RedString("%d", failedReqs))
	fmt.Println()
	fmt.Printf("  Avg Response Time: %s\n", color.YellowString("%.2f ms", avgLatency))
	fmt.Printf("  P95 Response Time: %s\n", color.YellowString("%.2f ms", p95Latency))
	fmt.Printf("  P99 Response Time: %s\n", color.YellowString("%.2f ms", p99Latency))
	fmt.Println()
	fmt.Printf("  Throughput:        %s\n", color.MagentaString("%.2f req/s", rps))
	fmt.Println()
}

func printInfo(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", color.BlueString("ℹ"), msg)
}

func printSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", color.GreenString("✓"), msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", color.RedString("✗"), msg)
}

func printHeader(msg string) {
	fmt.Println(color.New(color.Bold, color.Underline).Sprint(msg))
}
