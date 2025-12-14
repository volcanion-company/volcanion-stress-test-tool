package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	listType   string
	listLimit  int
	listStatus string
)

var listCmd = &cobra.Command{
	Use:   "list [plans|runs]",
	Short: "List test plans or test runs",
	Long: `List test plans or test runs with optional filtering.
	
Examples:
  # List test plans
  volcanion list plans
  
  # List test runs
  volcanion list runs
  
  # List only running tests
  volcanion list runs --status running
  
  # List last 5 test runs
  volcanion list runs --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: listResources,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 10, "maximum number of results")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "filter by status (for runs)")
}

func listResources(cmd *cobra.Command, args []string) error {
	resourceType := args[0]

	client := NewAPIClient(GetAPIBaseURL())

	switch resourceType {
	case "plans", "plan":
		return listPlans(client)
	case "runs", "run":
		return listRuns(client)
	default:
		return fmt.Errorf("unknown resource type: %s (use 'plans' or 'runs')", resourceType)
	}
}

func listPlans(client *APIClient) error {
	printInfo("Fetching test plans...")

	plans, err := client.GetTestPlans()
	if err != nil {
		return fmt.Errorf("failed to fetch plans: %w", err)
	}

	if len(plans) == 0 {
		fmt.Println("No test plans found")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, color.New(color.Bold).Sprint("ID\tNAME\tTARGET URL\tVUs\tDURATION\tCREATED"))

	// Rows
	count := 0
	for _, plan := range plans {
		if count >= listLimit {
			break
		}

		id := plan["id"].(string)
		name := plan["name"].(string)
		targetURL := plan["target_url"].(string)
		vus := int(plan["concurrent_users"].(float64))
		duration := int(plan["duration_sec"].(float64))
		createdAt := plan["created_at"].(string)

		// Parse and format date
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			t = time.Now()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%ds\t%s\n",
			color.CyanString(id[:8]),
			color.YellowString(name),
			truncate(targetURL, 40),
			vus,
			duration,
			t.Format("2006-01-02 15:04"))

		count++
	}

	return nil
}

func listRuns(client *APIClient) error {
	printInfo("Fetching test runs...")

	runs, err := client.GetTestRuns()
	if err != nil {
		return fmt.Errorf("failed to fetch runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No test runs found")
		return nil
	}

	// Filter by status if requested
	if listStatus != "" {
		filtered := make([]map[string]interface{}, 0)
		for _, run := range runs {
			if run["status"].(string) == listStatus {
				filtered = append(filtered, run)
			}
		}
		runs = filtered
	}

	if len(runs) == 0 {
		fmt.Printf("No test runs found with status: %s\n", listStatus)
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, color.New(color.Bold).Sprint("ID\tPLAN\tSTATUS\tREQUESTS\tSUCCESS RATE\tSTARTED"))

	// Rows
	count := 0
	for _, run := range runs {
		if count >= listLimit {
			break
		}

		id := run["id"].(string)
		planName := run["test_plan_name"].(string)
		status := run["status"].(string)
		startedAt := run["started_at"].(string)

		// Parse and format date
		t, err := time.Parse(time.RFC3339, startedAt)
		if err != nil {
			t = time.Now()
		}

		// Get metrics if available
		requests := "-"
		successRate := "-"

		if run["total_requests"] != nil {
			totalReqs := int(run["total_requests"].(float64))
			successReqs := int(run["success_requests"].(float64))
			requests = fmt.Sprintf("%d", totalReqs)
			if totalReqs > 0 {
				rate := float64(successReqs) / float64(totalReqs) * 100
				successRate = fmt.Sprintf("%.1f%%", rate)
			}
		}

		// Color status
		statusStr := status
		switch status {
		case "completed":
			statusStr = color.GreenString(status)
		case "running":
			statusStr = color.BlueString(status)
		case "failed":
			statusStr = color.RedString(status)
		case "cancelled":
			statusStr = color.YellowString(status)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			color.CyanString(id[:8]),
			truncate(planName, 30),
			statusStr,
			requests,
			successRate,
			t.Format("2006-01-02 15:04"))

		count++
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
