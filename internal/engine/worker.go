package engine

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

// Worker represents a single worker that executes HTTP requests
type Worker struct {
	ID        int
	plan      *model.TestPlan
	client    *http.Client
	metrics   *model.Metrics
	latencies []float64
}

// NewWorker creates a new worker instance
func NewWorker(id int, plan *model.TestPlan, metrics *model.Metrics) *Worker {
	// Create custom HTTP client with timeout and keep-alive
	client := &http.Client{
		Timeout: time.Duration(plan.TimeoutMs) * time.Millisecond,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	return &Worker{
		ID:        id,
		plan:      plan,
		client:    client,
		metrics:   metrics,
		latencies: make([]float64, 0, 1000),
	}
}

// Run executes the worker's request loop until context is cancelled
func (w *Worker) Run(ctx context.Context, requestChan <-chan struct{}) {
	logger.Log.Debug("Worker started",
		zap.Int("worker_id", w.ID),
		zap.String("target_url", w.plan.TargetURL))

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("Worker stopped",
				zap.Int("worker_id", w.ID))
			return
		case _, ok := <-requestChan:
			if !ok {
				return
			}
			w.executeRequest(ctx)
		}
	}
}

// executeRequest performs a single HTTP request and records metrics
func (w *Worker) executeRequest(ctx context.Context) {
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, w.plan.Method, w.plan.TargetURL, bytes.NewBufferString(w.plan.Body))
	if err != nil {
		latency := float64(time.Since(startTime).Milliseconds())
		w.metrics.RecordRequest(false, latency, 0, err)
		logger.Log.Error("Failed to create request",
			zap.Int("worker_id", w.ID),
			zap.Error(err))
		return
	}

	// Add headers
	for key, value := range w.plan.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := w.client.Do(req)
	latency := float64(time.Since(startTime).Milliseconds())

	if err != nil {
		w.metrics.RecordRequest(false, latency, 0, err)
		logger.Log.Debug("Request failed",
			zap.Int("worker_id", w.ID),
			zap.Error(err))
		return
	}
	defer resp.Body.Close()

	// Read and discard response body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	// Record success/failure based on status code
	success := resp.StatusCode >= 200 && resp.StatusCode < 400
	w.metrics.RecordRequest(success, latency, resp.StatusCode, nil)

	// Store latency for percentile calculation
	w.latencies = append(w.latencies, latency)
}

// GetLatencies returns all recorded latencies
func (w *Worker) GetLatencies() []float64 {
	return w.latencies
}
