package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
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
func TestNewWorker(t *testing.T) {
	plan := &model.TestPlan{
		ID:        "test-plan-1",
		Name:      "Test Plan",
		TargetURL: "http://localhost:8080",
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-1")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	if worker == nil {
		t.Fatal("Expected worker to be created")
	}
	if worker.ID != 1 {
		t.Errorf("Worker ID mismatch: expected 1, got %d", worker.ID)
	}
	if worker.plan != plan {
		t.Error("Worker plan mismatch")
	}
}

func TestWorkerExecuteRequest(t *testing.T) {
	var requestReceived atomic.Bool
	var receivedMethod string
	var receivedPath string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived.Store(true)
		mu.Lock()
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-exec",
		Name:      "Execute Request Test",
		TargetURL: server.URL + "/api/test",
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-exec")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 1)
	requestChan <- struct{}{}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	// Wait for request to be processed
	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	if !requestReceived.Load() {
		t.Error("Expected request to be sent to server")
	}

	mu.Lock()
	if receivedMethod != "GET" {
		t.Errorf("Expected GET method, got %s", receivedMethod)
	}
	if receivedPath != "/api/test" {
		t.Errorf("Expected path /api/test, got %s", receivedPath)
	}
	mu.Unlock()

	m.Mu.RLock()
	totalRequests := m.TotalRequests
	successRequests := m.SuccessRequests
	m.Mu.RUnlock()

	if totalRequests == 0 {
		t.Error("Expected total requests to be recorded")
	}
	if successRequests == 0 {
		t.Error("Expected success requests to be recorded")
	}
}

func TestWorkerPOSTWithBody(t *testing.T) {
	var receivedBody string
	var receivedContentType string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedContentType = r.Header.Get("Content-Type")
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		receivedBody = string(body)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-post",
		Name:      "POST Test",
		TargetURL: server.URL + "/api/users",
		Method:    "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:      `{"name": "John", "email": "john@example.com"}`,
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-post")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 1)
	requestChan <- struct{}{}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if receivedContentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", receivedContentType)
	}
	if receivedBody != `{"name": "John", "email": "john@example.com"}` {
		t.Errorf("Body mismatch: got %s", receivedBody)
	}
}

func TestWorkerCustomHeaders(t *testing.T) {
	var receivedHeaders http.Header
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedHeaders = r.Header.Clone()
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{})
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-headers",
		Name:      "Headers Test",
		TargetURL: server.URL,
		Method:    "GET",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-Custom":      "custom-value",
			"Accept":        "application/json",
		},
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-headers")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 1)
	requestChan <- struct{}{}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if receivedHeaders.Get("Authorization") != "Bearer token123" {
		t.Errorf("Authorization header mismatch: got %s", receivedHeaders.Get("Authorization"))
	}
	if receivedHeaders.Get("X-Custom") != "custom-value" {
		t.Errorf("X-Custom header mismatch: got %s", receivedHeaders.Get("X-Custom"))
	}
	if receivedHeaders.Get("Accept") != "application/json" {
		t.Errorf("Accept header mismatch: got %s", receivedHeaders.Get("Accept"))
	}
}

func TestWorkerTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Delay longer than timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-timeout",
		Name:      "Timeout Test",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 100, // Very short timeout
	}
	m := model.NewMetrics("run-timeout")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 1)
	requestChan <- struct{}{}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	wg.Wait()

	m.Mu.RLock()
	failedRequests := m.FailedRequests
	m.Mu.RUnlock()

	if failedRequests == 0 {
		t.Error("Expected failed request due to timeout")
	}
}

func TestWorkerServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-error",
		Name:      "Server Error Test",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-error")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 1)
	requestChan <- struct{}{}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	m.Mu.RLock()
	statusCodes := m.StatusCodes
	m.Mu.RUnlock()

	if statusCodes[500] == 0 {
		t.Error("Expected 500 status code to be recorded")
	}
}

func TestWorkerMultipleRequests(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-multi",
		Name:      "Multiple Requests Test",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-multi")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		requestChan <- struct{}{}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	// Wait for all requests
	time.Sleep(500 * time.Millisecond)
	cancel()
	wg.Wait()

	count := atomic.LoadInt64(&requestCount)
	if count < 10 {
		t.Errorf("Expected at least 10 requests, got %d", count)
	}
}

func TestWorkerLatencyRecording(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond) // Known latency
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-latency",
		Name:      "Latency Test",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-latency")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	requestChan := make(chan struct{}, 5)
	for i := 0; i < 5; i++ {
		requestChan <- struct{}{}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	time.Sleep(1 * time.Second)
	cancel()
	wg.Wait()

	m.Mu.RLock()
	minLatency := m.MinLatencyMs
	maxLatency := m.MaxLatencyMs
	totalReqs := m.TotalRequests
	m.Mu.RUnlock()

	t.Logf("Latency - Min: %.2fms, Max: %.2fms, Total Requests: %d", minLatency, maxLatency, totalReqs)

	// Should be at least 50ms due to server delay
	if minLatency < 40 { // Allow some variance
		t.Errorf("Min latency %.2fms is less than expected ~50ms", minLatency)
	}
	// Verify we actually made requests
	if totalReqs == 0 {
		t.Error("Expected some requests to be made")
	}
}

func TestWorkerContextCancellation(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		time.Sleep(10 * time.Millisecond) // Add delay to ensure cancellation works
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	plan := &model.TestPlan{
		ID:        "test-cancel",
		Name:      "Cancel Test",
		TargetURL: server.URL,
		Method:    "GET",
		TimeoutMs: 5000,
	}
	m := model.NewMetrics("run-cancel")
	collector := getSharedTestCollector()

	worker := NewWorker(1, plan, m, http.DefaultClient, collector)

	ctx, cancel := context.WithCancel(context.Background())

	requestChan := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		requestChan <- struct{}{}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker.Run(ctx, requestChan)
	}()

	// Process a few requests then cancel
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	count := atomic.LoadInt64(&requestCount)
	// Should have processed fewer than 50 requests in 50ms with 10ms delay each
	t.Logf("Processed %d requests before cancellation", count)
	if count > 50 {
		t.Error("Worker should have stopped relatively quickly after cancellation")
	}
}
