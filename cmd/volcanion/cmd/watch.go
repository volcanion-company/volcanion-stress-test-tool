package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

func watchTest(client *APIClient, runID string) error {
	printInfo("Watching live metrics... (Press Ctrl+C to stop watching)")
	fmt.Println()

	// Create progress bar for duration
	var bar *progressbar.ProgressBar

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	firstUpdate := true

	for {
		<-ticker.C

		// Get test run info
		run, err := client.GetTestRun(runID)
		if err != nil {
			return fmt.Errorf("failed to get test run: %w", err)
		}

		status := run["status"].(string)

		// Get live metrics
		metrics, err := client.GetLiveMetrics(runID)
		if err != nil {
			if IsVerbose() {
				printError(fmt.Sprintf("Failed to get metrics: %v", err))
			}
			continue
		}

		// Initialize progress bar on first update
		if firstUpdate && run["duration_sec"] != nil {
			duration := int(run["duration_sec"].(float64))
			bar = progressbar.NewOptions(duration,
				progressbar.OptionSetDescription("Progress"),
				progressbar.OptionSetWidth(40),
				progressbar.OptionShowCount(),
				progressbar.OptionSetPredictTime(true),
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "[green]=[reset]",
					SaucerHead:    "[green]>[reset]",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)
			firstUpdate = false
		}

		// Update progress bar
		if bar != nil && metrics["total_duration_ms"] != nil {
			elapsed := int(metrics["total_duration_ms"].(float64) / 1000)
			bar.Set(elapsed)
		}

		// Print live stats
		clearLines(6)
		printLiveStats(metrics)

		// Check if completed
		if status == "completed" || status == "failed" || status == "cancelled" {
			fmt.Println()
			printSuccess(fmt.Sprintf("Test %s!", status))
			return nil
		}
	}
}

func printLiveStats(metrics map[string]interface{}) {
	totalReqs := int(metrics["total_requests"].(float64))
	successReqs := int(metrics["success_requests"].(float64))
	failedReqs := int(metrics["failed_requests"].(float64))
	currentRPS := metrics["current_rps"].(float64)
	avgLatency := metrics["avg_latency_ms"].(float64)

	successRate := 0.0
	if totalReqs > 0 {
		successRate = float64(successReqs) / float64(totalReqs) * 100
	}

	fmt.Printf("\r  Requests:     %s | Success: %s (%.1f%%) | Failed: %s\n",
		color.CyanString("%6d", totalReqs),
		color.GreenString("%6d", successReqs),
		successRate,
		color.RedString("%6d", failedReqs))

	fmt.Printf("  RPS:          %s\n", color.MagentaString("%.2f req/s", currentRPS))
	fmt.Printf("  Avg Latency:  %s\n", color.YellowString("%.2f ms", avgLatency))

	// Print P95 and P99 if available
	if p95, ok := metrics["p95_latency_ms"]; ok {
		fmt.Printf("  P95 Latency:  %s\n", color.YellowString("%.2f ms", p95.(float64)))
	}
	if p99, ok := metrics["p99_latency_ms"]; ok {
		fmt.Printf("  P99 Latency:  %s\n", color.YellowString("%.2f ms", p99.(float64)))
	}
}

func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[F\033[K")
	}
}
