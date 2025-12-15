package engine

import (
	"net/http"
	"net/http/httptest"
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

func TestNewScheduler(t *testing.T) {
	plan := &model.TestPlan{
		ID:          "test-plan-1",
		Name:        "Test Plan",
		TargetURL:   "http://localhost:8080",
		Method:      "GET",
		Users:       10,
		DurationSec: 5,
		RampUpSec:   2,
	}
	m := model.NewMetrics("run-1")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)

	if scheduler == nil {
		t.Fatal("Expected scheduler to be created")
	}
	if scheduler.plan != plan {
		t.Error("Scheduler plan mismatch")
	}
	if scheduler.metrics != m {
		t.Error("Scheduler metrics mismatch")
	}
}

func TestSchedulerStartAndStop(t *testing.T) {
	// Create a test server that responds immediately
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-plan-1",
		Name:        "Test Plan",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       5,
		DurationSec: 2,
		RampUpSec:   0,
		TargetRPS:   100,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("run-1")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Let it run for a short time
	time.Sleep(500 * time.Millisecond)

	// Stop the scheduler
	scheduler.Stop()

	// Verify some requests were made
	count := atomic.LoadInt64(&requestCount)
	if count == 0 {
		t.Error("Expected some requests to be made")
	}
	t.Logf("Made %d requests in 500ms", count)
}

func TestSchedulerRampUp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-plan-rampup",
		Name:        "Ramp Up Test",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       10,
		DurationSec: 5,
		RampUpSec:   2, // 2 second ramp up
		TargetRPS:   50,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("run-rampup")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Check workers after 500ms (should be ~2-3 workers)
	time.Sleep(500 * time.Millisecond)
	m.Mu.RLock()
	workersAt500ms := m.ActiveWorkers
	m.Mu.RUnlock()

	// Check workers after 2.5s (should be all 10 workers)
	time.Sleep(2 * time.Second)
	m.Mu.RLock()
	workersAt2500ms := m.ActiveWorkers
	m.Mu.RUnlock()

	scheduler.Stop()

	t.Logf("Workers at 500ms: %d, Workers at 2500ms: %d", workersAt500ms, workersAt2500ms)

	if workersAt500ms >= plan.Users {
		t.Errorf("Expected fewer workers during ramp up, got %d", workersAt500ms)
	}
	if workersAt2500ms < plan.Users/2 {
		t.Errorf("Expected more workers after ramp up, got %d", workersAt2500ms)
	}
}

func TestSchedulerMetricsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond) // Add some latency
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-plan-metrics",
		Name:        "Metrics Test",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       3,
		DurationSec: 2,
		RampUpSec:   0,
		TargetRPS:   20,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("run-metrics")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	time.Sleep(1 * time.Second)
	scheduler.Stop()

	m.Mu.RLock()
	totalRequests := m.TotalRequests
	successRequests := m.SuccessRequests
	failedRequests := m.FailedRequests
	minLatency := m.MinLatencyMs
	maxLatency := m.MaxLatencyMs
	m.Mu.RUnlock()

	t.Logf("Total: %d, Success: %d, Failed: %d, Min: %.2fms, Max: %.2fms",
		totalRequests, successRequests, failedRequests, minLatency, maxLatency)

	if totalRequests == 0 {
		t.Error("Expected some requests to be recorded")
	}
	if successRequests == 0 {
		t.Error("Expected some successful requests")
	}
	if minLatency < 0 {
		t.Error("Expected min latency to be recorded")
	}
	if minLatency < 10 {
		t.Logf("Warning: Min latency %.2fms is less than server delay", minLatency)
	}
}

func TestSchedulerLoadPatternConstant(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-constant",
		Name:        "Constant Load",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       5,
		DurationSec: 3,
		RampUpSec:   0,
		TargetRPS:   50, // 50 requests per second
		TimeoutMs:   5000,
		RatePattern: model.RatePatternFixed,
	}
	m := model.NewMetrics("run-constant")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	time.Sleep(2 * time.Second)
	scheduler.Stop()

	count := atomic.LoadInt64(&requestCount)
	expectedMin := int64(50 * 2 * 0.7) // 70% of expected
	expectedMax := int64(50 * 2 * 1.3) // 130% of expected

	t.Logf("Request count: %d (expected range: %d-%d)", count, expectedMin, expectedMax)

	if count < expectedMin {
		t.Errorf("Too few requests: got %d, expected at least %d", count, expectedMin)
	}
	if count > expectedMax {
		t.Errorf("Too many requests: got %d, expected at most %d", count, expectedMax)
	}
}

func TestSchedulerContextCancellation(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-cancel",
		Name:        "Cancel Test",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       5,
		DurationSec: 10, // Long duration
		RampUpSec:   0,
		TargetRPS:   20,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("run-cancel")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Let it run briefly then cancel
	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()

	countBefore := atomic.LoadInt64(&requestCount)

	// Wait a bit to ensure no more requests are made
	time.Sleep(300 * time.Millisecond)
	countAfter := atomic.LoadInt64(&requestCount)

	// Allow for some in-flight requests to complete
	if countAfter > countBefore+5 {
		t.Errorf("Requests continued after stop: before=%d, after=%d", countBefore, countAfter)
	}
}

func TestSchedulerWait(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:          "test-wait",
		Name:        "Wait Test",
		TargetURL:   server.URL,
		Method:      "GET",
		Users:       2,
		DurationSec: 1, // Short duration
		RampUpSec:   0,
		TargetRPS:   10,
		TimeoutMs:   5000,
	}
	m := model.NewMetrics("run-wait")
	collector := getSharedTestCollector()

	scheduler := NewScheduler(plan, m, http.DefaultClient, collector)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		scheduler.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Good - Wait completed
	case <-time.After(3 * time.Second):
		t.Error("Wait timed out - should have completed within duration")
	}
}
