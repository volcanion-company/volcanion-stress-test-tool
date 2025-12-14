package reporting

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
)

// ExportFormat represents the export format type
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
	FormatHTML ExportFormat = "html"
)

// TestRunExport represents data structure for test run export
type TestRunExport struct {
	TestRun      *model.TestRun  `json:"test_run"`
	TestPlan     *model.TestPlan `json:"test_plan"`
	Metrics      *model.Metrics  `json:"metrics,omitempty"`
	ExportedAt   time.Time       `json:"exported_at"`
	ExportFormat string          `json:"export_format"`
}

// Exporter handles export of test results in various formats
type Exporter struct{}

// NewExporter creates a new Exporter
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportTestRun exports a single test run to the specified format
func (e *Exporter) ExportTestRun(writer io.Writer, format ExportFormat, testRun *model.TestRun, testPlan *model.TestPlan, metrics *model.Metrics) error {
	exportData := &TestRunExport{
		TestRun:      testRun,
		TestPlan:     testPlan,
		Metrics:      metrics,
		ExportedAt:   time.Now(),
		ExportFormat: string(format),
	}

	switch format {
	case FormatJSON:
		return e.exportJSON(writer, exportData)
	case FormatCSV:
		return e.exportCSV(writer, exportData)
	case FormatHTML:
		return e.exportHTML(writer, exportData)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON exports test run as JSON
func (e *Exporter) exportJSON(writer io.Writer, data *TestRunExport) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportCSV exports test run metrics as CSV
func (e *Exporter) exportCSV(writer io.Writer, data *TestRunExport) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	headers := []string{
		"Test Run ID", "Test Plan ID", "Test Plan Name", "Status",
		"Started At", "Ended At", "Duration (s)",
		"Total Requests", "Successful Requests", "Failed Requests", "Success Rate (%)",
		"Min Response Time (ms)", "Max Response Time (ms)", "Avg Response Time (ms)",
		"P50 (ms)", "P75 (ms)", "P95 (ms)", "P99 (ms)",
		"Requests/Second", "Concurrent Users",
	}
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// Calculate duration
	duration := float64(0)
	if data.TestRun.EndAt != nil {
		duration = data.TestRun.EndAt.Sub(data.TestRun.StartAt).Seconds()
	}

	// Calculate success rate
	successRate := float64(0)
	if data.Metrics != nil && data.Metrics.TotalRequests > 0 {
		successRate = float64(data.Metrics.SuccessRequests) / float64(data.Metrics.TotalRequests) * 100
	}

	// Write data row
	row := []string{
		data.TestRun.ID,
		data.TestRun.PlanID,
		data.TestPlan.Name,
		string(data.TestRun.Status),
		data.TestRun.StartAt.Format(time.RFC3339),
		formatTimePtr(data.TestRun.EndAt),
		fmt.Sprintf("%.2f", duration),
	}

	if data.Metrics != nil {
		row = append(row,
			fmt.Sprintf("%d", data.Metrics.TotalRequests),
			fmt.Sprintf("%d", data.Metrics.SuccessRequests),
			fmt.Sprintf("%d", data.Metrics.FailedRequests),
			fmt.Sprintf("%.2f", successRate),
			fmt.Sprintf("%.2f", data.Metrics.MinLatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.MaxLatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.AvgLatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.P50LatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.P75LatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.P95LatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.P99LatencyMs),
			fmt.Sprintf("%.2f", data.Metrics.RequestsPerSec),
			fmt.Sprintf("%d", data.TestPlan.Users),
		)
	} else {
		row = append(row, "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A", "N/A")
	}

	return csvWriter.Write(row)
}

// exportHTML exports test run as HTML report
func (e *Exporter) exportHTML(writer io.Writer, data *TestRunExport) error {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Run Report - {{.TestPlan.Name}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #007bff;
            padding-bottom: 10px;
        }
        h2 {
            color: #555;
            margin-top: 30px;
            border-bottom: 2px solid #e0e0e0;
            padding-bottom: 8px;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .metric-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            border-left: 4px solid #007bff;
        }
        .metric-card.success {
            border-left-color: #28a745;
        }
        .metric-card.warning {
            border-left-color: #ffc107;
        }
        .metric-card.danger {
            border-left-color: #dc3545;
        }
        .metric-label {
            font-size: 12px;
            color: #666;
            text-transform: uppercase;
            font-weight: 600;
        }
        .metric-value {
            font-size: 28px;
            font-weight: bold;
            color: #333;
            margin-top: 5px;
        }
        .metric-unit {
            font-size: 14px;
            color: #888;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #e0e0e0;
        }
        th {
            background: #f8f9fa;
            font-weight: 600;
            color: #555;
        }
        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
        }
        .status-running {
            background: #fff3cd;
            color: #856404;
        }
        .status-completed {
            background: #d4edda;
            color: #155724;
        }
        .status-failed {
            background: #f8d7da;
            color: #721c24;
        }
        .status-stopped {
            background: #d1ecf1;
            color: #0c5460;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e0e0e0;
            text-align: center;
            color: #888;
            font-size: 14px;
        }
        .progress-bar {
            background: #e0e0e0;
            border-radius: 4px;
            height: 24px;
            overflow: hidden;
            margin-top: 10px;
        }
        .progress-fill {
            background: #28a745;
            height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 12px;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Test Run Report</h1>
        
        <h2>Test Configuration</h2>
        <table>
            <tr>
                <th>Test Plan Name</th>
                <td>{{.TestPlan.Name}}</td>
            </tr>
            <tr>
                <th>Target URL</th>
                <td>{{.TestPlan.TargetURL}}</td>
            </tr>
            <tr>
                <th>HTTP Method</th>
                <td>{{.TestPlan.Method}}</td>
            </tr>
            <tr>
                <th>Concurrent Users</th>
                <td>{{.TestPlan.Users}}</td>
            </tr>
            <tr>
                <th>Ramp-Up Period</th>
                <td>{{.TestPlan.RampUpSec}} seconds</td>
            </tr>
            <tr>
                <th>Test Duration</th>
                <td>{{.TestPlan.DurationSec}} seconds</td>
            </tr>
            <tr>
                <th>Status</th>
                <td><span class="status-badge status-{{.TestRun.Status}}">{{.TestRun.Status}}</span></td>
            </tr>
            <tr>
                <th>Started At</th>
                <td>{{.TestRun.StartAt.Format "2006-01-02 15:04:05 MST"}}</td>
            </tr>
            {{if .TestRun.EndAt}}
            <tr>
                <th>Ended At</th>
                <td>{{.TestRun.EndAt.Format "2006-01-02 15:04:05 MST"}}</td>
            </tr>
            <tr>
                <th>Actual Duration</th>
                <td>{{printf "%.2f" (.TestRun.EndAt.Sub .TestRun.StartAt).Seconds}} seconds</td>
            </tr>
            {{end}}
        </table>

        {{if .Metrics}}
        <h2>Performance Metrics</h2>
        
        <div class="summary-grid">
            <div class="metric-card success">
                <div class="metric-label">Total Requests</div>
                <div class="metric-value">{{.Metrics.TotalRequests}}</div>
            </div>
            <div class="metric-card success">
                <div class="metric-label">Successful Requests</div>
                <div class="metric-value">{{.Metrics.SuccessRequests}}</div>
            </div>
            <div class="metric-card danger">
                <div class="metric-label">Failed Requests</div>
                <div class="metric-value">{{.Metrics.FailedRequests}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Requests/Second</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.RequestsPerSec}}</div>
            </div>
        </div>

        <div class="metric-card">
            <div class="metric-label">Success Rate</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: {{if gt .Metrics.TotalRequests 0}}{{printf "%.0f" (div (mul (float64 .Metrics.SuccessRequests) 100.0) (float64 .Metrics.TotalRequests))}}{{else}}0{{end}}%">
                    {{if gt .Metrics.TotalRequests 0}}{{printf "%.2f" (div (mul (float64 .Metrics.SuccessRequests) 100.0) (float64 .Metrics.TotalRequests))}}%{{else}}0%{{end}}
                </div>
            </div>
        </div>

        <h2>Response Times</h2>
        <div class="summary-grid">
            <div class="metric-card">
                <div class="metric-label">Average</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.AvgLatencyMs}} <span class="metric-unit">ms</span></div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Minimum</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.MinLatencyMs}} <span class="metric-unit">ms</span></div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Maximum</div>
                <div class="metric-value">{{printf "%.2f" .Metrics.MaxLatencyMs}} <span class="metric-unit">ms</span></div>
            </div>
        </div>

        <h2>Percentiles</h2>
        <table>
            <thead>
                <tr>
                    <th>Percentile</th>
                    <th>Response Time (ms)</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>P50 (Median)</td>
                    <td>{{printf "%.2f" .Metrics.P50LatencyMs}}</td>
                </tr>
                <tr>
                    <td>P75</td>
                    <td>{{printf "%.2f" .Metrics.P75LatencyMs}}</td>
                </tr>
                <tr>
                    <td>P95</td>
                    <td>{{printf "%.2f" .Metrics.P95LatencyMs}}</td>
                </tr>
                <tr>
                    <td>P99</td>
                    <td>{{printf "%.2f" .Metrics.P99LatencyMs}}</td>
                </tr>
            </tbody>
        </table>

        {{if .Metrics.Errors}}
        <h2>Error Distribution</h2>
        <table>
            <thead>
                <tr>
                    <th>Error Type</th>
                    <th>Count</th>
                </tr>
            </thead>
            <tbody>
                {{range $key, $value := .Metrics.Errors}}
                <tr>
                    <td>{{$key}}</td>
                    <td>{{$value}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{end}}

        {{if .Metrics.StatusCodes}}
        <h2>HTTP Status Code Distribution</h2>
        <table>
            <thead>
                <tr>
                    <th>Status Code</th>
                    <th>Count</th>
                </tr>
            </thead>
            <tbody>
                {{range $key, $value := .Metrics.StatusCodes}}
                <tr>
                    <td>{{$key}}</td>
                    <td>{{$value}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{end}}
        {{end}}

        <div class="footer">
            <p>Generated by Volcanion Stress Test Tool on {{.ExportedAt.Format "2006-01-02 15:04:05 MST"}}</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("report").Funcs(template.FuncMap{
		"float64": func(i int64) float64 { return float64(i) },
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mul": func(a, b float64) float64 { return a * b },
	}).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(writer, data)
}

// formatTimePtr formats a time pointer or returns "N/A"
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	return t.Format(time.RFC3339)
}
