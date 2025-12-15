package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
)

func init() {
	if err := logger.Init("error"); err != nil {
		panic(err)
	}
}

// Benchmark scheduler throughput
func BenchmarkSchedulerThroughput(b *testing.B) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "bench-plan",
		Name:        "Benchmark Plan",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       10,
		DurationSec: 60,
		RampUpSec:   0,
		TargetRPS:   1000,
		TimeoutMs:   5000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := model.NewMetrics("run-" + string(rune('0'+i)))
		collector := getSharedTestCollector()
		scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
		if err := scheduler.Start(); err != nil {
			b.Fatalf("Failed to start scheduler: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		scheduler.Stop()
	}
	b.StopTimer()

	b.ReportMetric(float64(atomic.LoadInt64(&requestCount))/float64(b.N), "requests/iteration")
}

// Benchmark worker request execution
func BenchmarkWorkerRequestExecution(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "bench-worker",
		Name:      "Benchmark Worker",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("bench-run")
	collector := getSharedTestCollector()
	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithCancel(context.Background())
	requestChan := make(chan struct{}, b.N)

	for i := 0; i < b.N; i++ {
		requestChan <- struct{}{}
	}

	b.ResetTimer()
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, requestChan)
		close(done)
	}()

	time.Sleep(5 * time.Second) // Allow time for requests
	cancel()
	<-done
	b.StopTimer()
}

// Benchmark scheduler with different user counts
func BenchmarkSchedulerUserScaling(b *testing.B) {
	userCounts := []int{1, 5, 10, 25, 50}

	for _, users := range userCounts {
		b.Run("users_"+string(rune('0'+users/10))+string(rune('0'+users%10)), func(b *testing.B) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			plan := &model.TestPlan{
				ID:          "scale-plan",
				Name:        "Scale Plan",
				TargetURL:   server.URL,
				Method:      "GET",
				Users:       users,
				DurationSec: 60,
				RampUpSec:   0,
				TargetRPS:   100,
				TimeoutMs:   5000,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := model.NewMetrics("run")
				collector := getSharedTestCollector()
				scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
				if err := scheduler.Start(); err != nil {
					b.Fatalf("Failed to start scheduler: %v", err)
				}
				time.Sleep(50 * time.Millisecond)
				scheduler.Stop()
			}
		})
	}
}

// Benchmark metrics recording
func BenchmarkMetricsRecording(b *testing.B) {
	m := model.NewMetrics("bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordRequest(true, 50.0, 200, nil)
	}
}

// Benchmark concurrent metrics recording
func BenchmarkConcurrentMetricsRecording(b *testing.B) {
	m := model.NewMetrics("bench")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.RecordRequest(true, 50.0, 200, nil)
		}
	})
}

// Benchmark worker with POST requests
func BenchmarkWorkerPOSTRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "bench-post",
		Name:      "Benchmark POST",
		TargetURL: server.URL,
		Method:    "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:      `{"name": "test", "value": 123}`,
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("bench-run")
	collector := getSharedTestCollector()
	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithCancel(context.Background())
	requestChan := make(chan struct{}, b.N)

	for i := 0; i < b.N; i++ {
		requestChan <- struct{}{}
	}

	b.ResetTimer()
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, requestChan)
		close(done)
	}()

	time.Sleep(5 * time.Second)
	cancel()
	<-done
	b.StopTimer()
}

// Benchmark scheduler ramp-up
func BenchmarkSchedulerRampUp(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "ramp-bench",
		Name:        "Ramp-up Benchmark",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       20,
		DurationSec: 60,
		RampUpSec:   5,
		TargetRPS:   100,
		TimeoutMs:   5000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := model.NewMetrics("run")
		collector := getSharedTestCollector()
		scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
		if err := scheduler.Start(); err != nil {
			b.Fatalf("Failed to start scheduler: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		scheduler.Stop()
	}
}

// Test load generator stability under sustained load
func TestSchedulerSustainedLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
	}
	// These sustained load tests are resource/time intensive and can be flaky
	// in constrained CI environments. Require an explicit env var to run them.
	if os.Getenv("RUN_SUSTAINED_TESTS") != "1" {
		t.Skip("Skipping sustained load test (set RUN_SUSTAINED_TESTS=1 to run)")
	}

	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "sustained-test",
		Name:        "Sustained Load Test",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       10,
		DurationSec: 10, // 10 seconds sustained
		RampUpSec:   1,
		TargetRPS:   100,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("sustained")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for test to complete
	scheduler.Wait()

	count := atomic.LoadInt64(&requestCount)
	expectedMin := int64(100 * 9 * 0.7) // 70% of expected (accounting for ramp-up)

	t.Logf("Total requests: %d (expected min: %d)", count, expectedMin)

	if count < expectedMin {
		t.Errorf("Too few requests: got %d, expected at least %d", count, expectedMin)
	}

	m.Mu.RLock()
	totalRequests := m.TotalRequests
	errorRate := float64(m.FailedRequests) / float64(m.TotalRequests) * 100
	m.Mu.RUnlock()

	t.Logf("Metrics - Total: %d, Error Rate: %.2f%%", totalRequests, errorRate)

	if errorRate > 5 {
		t.Errorf("Error rate too high: %.2f%%", errorRate)
	}
}

// Test memory stability under load
func TestSchedulerMemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory stability test in short mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Run multiple short tests to check for memory leaks
	for i := 0; i < 5; i++ {
		plan := &model.TestPlan{
			ID:          "memory-test-" + string(rune('0'+i)),
			Name:        "Memory Test",
			TargetURL:   server.URL,
			Method:      "GET",
			Users:       5,
			DurationSec: 3,
			RampUpSec:   0,
			TargetRPS:   50,
			TimeoutMs:   5000,
		}
		m := model.NewMetrics("memory-run")
		collector := getSharedTestCollector()

		scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
		if err := scheduler.Start(); err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}
		scheduler.Wait()

		t.Logf("Iteration %d completed", i+1)
	}
}
