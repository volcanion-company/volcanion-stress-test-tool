package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportRunID  string
)

var exportCmd = &cobra.Command{
	Use:   "export <run-id>",
	Short: "Export test results",
	Long: `Export test results in various formats (JSON, CSV, HTML).
	
Examples:
  # Export as JSON
  volcanion export abc123 --format json -o results.json
  
  # Export as CSV
  volcanion export abc123 --format csv -o results.csv
  
  # Export as HTML report
  volcanion export abc123 --format html -o report.html`,
	Args: cobra.ExactArgs(1),
	RunE: exportResults,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "export format (json, csv, html)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (required)")
	exportCmd.MarkFlagRequired("output")
}

func exportResults(cmd *cobra.Command, args []string) error {
	runID := args[0]

	client := NewAPIClient(GetAPIBaseURL())

	printInfo(fmt.Sprintf("Fetching test results for run %s...", runID))

	// Get test run and metrics
	run, err := client.GetTestRun(runID)
	if err != nil {
		return fmt.Errorf("failed to fetch test run: %w", err)
	}

	metrics, err := client.GetTestRunMetrics(runID)
	if err != nil {
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// Export based on format
	switch exportFormat {
	case "json":
		err = exportJSON(run, metrics, outputFile)
	case "csv":
		err = exportCSV(metrics, outputFile)
	case "html":
		err = exportHTML(run, metrics, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s (use json, csv, or html)", exportFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	printSuccess(fmt.Sprintf("Results exported to %s", outputFile))
	return nil
}

func exportJSON(run, metrics map[string]interface{}, filename string) error {
	data := map[string]interface{}{
		"run":     run,
		"metrics": metrics,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func exportCSV(metrics map[string]interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Metric", "Value"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write metrics
	rows := [][]string{
		{"Total Requests", fmt.Sprintf("%.0f", metrics["total_requests"].(float64))},
		{"Success Requests", fmt.Sprintf("%.0f", metrics["success_requests"].(float64))},
		{"Failed Requests", fmt.Sprintf("%.0f", metrics["failed_requests"].(float64))},
		{"Avg Latency (ms)", fmt.Sprintf("%.2f", metrics["avg_latency_ms"].(float64))},
		{"Min Latency (ms)", fmt.Sprintf("%.2f", metrics["min_latency_ms"].(float64))},
		{"Max Latency (ms)", fmt.Sprintf("%.2f", metrics["max_latency_ms"].(float64))},
		{"P50 Latency (ms)", fmt.Sprintf("%.2f", metrics["p50_latency_ms"].(float64))},
		{"P95 Latency (ms)", fmt.Sprintf("%.2f", metrics["p95_latency_ms"].(float64))},
		{"P99 Latency (ms)", fmt.Sprintf("%.2f", metrics["p99_latency_ms"].(float64))},
		{"Requests Per Second", fmt.Sprintf("%.2f", metrics["requests_per_sec"].(float64))},
		{"Total Duration (ms)", fmt.Sprintf("%.0f", metrics["total_duration_ms"].(float64))},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func exportHTML(run, metrics map[string]interface{}, filename string) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Test Results - {{.Run.TestPlanName}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 8px;
            padding: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .meta {
            color: #666;
            margin-bottom: 30px;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 30px 0;
        }
        .metric-card {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 6px;
            border-left: 4px solid #4CAF50;
        }
        .metric-card.warning {
            border-left-color: #FF9800;
        }
        .metric-card.danger {
            border-left-color: #F44336;
        }
        .metric-label {
            font-size: 14px;
            color: #666;
            margin-bottom: 5px;
        }
        .metric-value {
            font-size: 28px;
            font-weight: bold;
            color: #333;
        }
        .status {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 500;
        }
        .status.completed {
            background: #4CAF50;
            color: white;
        }
        .status.failed {
            background: #F44336;
            color: white;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 30px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background: #f5f5f5;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Load Test Results</h1>
        <div class="meta">
            <strong>Test Plan:</strong> {{.Run.TestPlanName}}<br>
            <strong>Status:</strong> <span class="status {{.Run.Status}}">{{.Run.Status}}</span><br>
            <strong>Started:</strong> {{.Run.StartedAt}}<br>
            {{if .Run.CompletedAt}}<strong>Completed:</strong> {{.Run.CompletedAt}}<br>{{end}}
        </div>

        <h2>Performance Metrics</h2>
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Total Requests</div>
                <div class="metric-value">{{.Metrics.TotalRequests}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Success Rate</div>
                <div class="metric-value">{{.Metrics.SuccessRate}}%</div>
            </div>
            <div class="metric-card {{if gt .Metrics.AvgLatency 1000.0}}warning{{end}}">
                <div class="metric-label">Avg Response Time</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.AvgLatency}} ms</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Throughput</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.RPS}} req/s</div>
            </div>
        </div>

        <h2>Response Time Percentiles</h2>
        <table>
            <tr>
                <th>Percentile</th>
                <th>Response Time</th>
            </tr>
            <tr>
                <td>P50 (Median)</td>
                <td>{{printf "%.2f" .Metrics.P50}} ms</td>
            </tr>
            <tr>
                <td>P95</td>
                <td>{{printf "%.2f" .Metrics.P95}} ms</td>
            </tr>
            <tr>
                <td>P99</td>
                <td>{{printf "%.2f" .Metrics.P99}} ms</td>
            </tr>
        </table>

        <p style="margin-top: 30px; color: #666; font-size: 14px;">
            Generated by Volcanion Stress Test Tool on {{.GeneratedAt}}
        </p>
    </div>
</body>
</html>`

	data := struct {
		Run         HTMLRun
		Metrics     HTMLMetrics
		GeneratedAt string
	}{
		Run: HTMLRun{
			TestPlanName: run["test_plan_name"].(string),
			Status:       run["status"].(string),
			StartedAt:    run["started_at"].(string),
			CompletedAt:  getStringOrEmpty(run, "completed_at"),
		},
		Metrics: HTMLMetrics{
			TotalRequests: int(metrics["total_requests"].(float64)),
			SuccessRate:   calculateSuccessRate(metrics),
			AvgLatency:    metrics["avg_latency_ms"].(float64),
			P50:           metrics["p50_latency_ms"].(float64),
			P95:           metrics["p95_latency_ms"].(float64),
			P99:           metrics["p99_latency_ms"].(float64),
			RPS:           metrics["requests_per_sec"].(float64),
		},
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

type HTMLRun struct {
	TestPlanName string
	Status       string
	StartedAt    string
	CompletedAt  string
}

type HTMLMetrics struct {
	TotalRequests int
	SuccessRate   float64
	AvgLatency    float64
	P50           float64
	P95           float64
	P99           float64
	RPS           float64
}

func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}

func calculateSuccessRate(metrics map[string]interface{}) float64 {
	total := metrics["total_requests"].(float64)
	success := metrics["success_requests"].(float64)
	if total == 0 {
		return 0
	}
	return (success / total) * 100
}
