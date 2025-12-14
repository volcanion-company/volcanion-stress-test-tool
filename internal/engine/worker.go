package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"go.uber.org/zap"
)

// Worker represents a single worker that executes HTTP requests
type Worker struct {
	ID             int
	plan           *model.TestPlan
	client         *http.Client
	metrics        *model.Metrics
	latencyBuffer  *RingBuffer
	collector      *metrics.Collector
	templateEngine *TemplateEngine
}

// NewWorker creates a new worker instance
func NewWorker(id int, plan *model.TestPlan, metrics *model.Metrics, sharedClient *http.Client, collector *metrics.Collector) *Worker {
	// Use the shared client but create a wrapper with timeout for this plan
	client := &http.Client{
		Transport: sharedClient.Transport,
		Timeout:   time.Duration(plan.TimeoutMs) * time.Millisecond,
	}

	// Create ring buffer: store last 10,000 latencies per worker
	latencyBuffer := NewRingBuffer(10000)

	return &Worker{
		ID:             id,
		plan:           plan,
		client:         client,
		metrics:        metrics,
		latencyBuffer:  latencyBuffer,
		collector:      collector,
		templateEngine: NewTemplateEngine(),
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

	// Apply template substitution to body and headers
	processedBody := w.templateEngine.Process(w.plan.Body)
	processedHeaders := w.templateEngine.ProcessMap(w.plan.Headers)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, w.plan.Method, w.plan.TargetURL, bytes.NewBufferString(processedBody))
	if err != nil {
		latency := float64(time.Since(startTime).Milliseconds())
		w.metrics.RecordRequest(false, latency, 0, err)
		logger.Log.Error("Failed to create request",
			zap.Int("worker_id", w.ID),
			zap.Error(err))
		return
	}

	// Add headers with template substitution
	for key, value := range processedHeaders {
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

	// Store latency in ring buffer for percentile calculation
	w.latencyBuffer.Add(latency)

	// Record to Prometheus
	status := fmt.Sprintf("%d", resp.StatusCode)
	w.collector.RecordRequest(w.metrics.RunID, w.plan.Method, status, latency/1000.0, !success)
}

// GetLatencies returns all recorded latencies from the ring buffer
func (w *Worker) GetLatencies() []float64 {
	return w.latencyBuffer.GetAll()
}
