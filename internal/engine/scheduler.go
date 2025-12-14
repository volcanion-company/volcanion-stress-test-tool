package engine

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"go.uber.org/zap"
)

// Scheduler manages the execution of a test run with workers and rate control
type Scheduler struct {
	plan    *model.TestPlan
	metrics *model.Metrics
	workers []*Worker
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	ctx     context.Context
}

// NewScheduler creates a new scheduler for a test plan
func NewScheduler(plan *model.TestPlan, metrics *model.Metrics) *Scheduler {
	return &Scheduler{
		plan:    plan,
		metrics: metrics,
		workers: make([]*Worker, 0, plan.Users),
	}
}

// Start begins the test execution
func (s *Scheduler) Start() error {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), time.Duration(s.plan.DurationSec)*time.Second)

	logger.Log.Info("Starting test execution",
		zap.String("plan_id", s.plan.ID),
		zap.Int("users", s.plan.Users),
		zap.Int("duration_sec", s.plan.DurationSec),
		zap.Int("ramp_up_sec", s.plan.RampUpSec))

	// Create request channel for rate control
	requestChan := make(chan struct{}, s.plan.Users*10)

	// Start workers with ramp-up
	go s.startWorkersWithRampUp(requestChan)

	// Start request generator
	go s.generateRequests(requestChan)

	// Start metrics reporter
	go s.reportMetrics()

	return nil
}

// startWorkersWithRampUp gradually spawns workers according to ramp-up time
func (s *Scheduler) startWorkersWithRampUp(requestChan <-chan struct{}) {
	if s.plan.RampUpSec == 0 {
		// No ramp-up, start all workers immediately
		for i := 0; i < s.plan.Users; i++ {
			worker := NewWorker(i, s.plan, s.metrics)
			s.workers = append(s.workers, worker)
			s.wg.Add(1)
			go func(w *Worker) {
				defer s.wg.Done()
				w.Run(s.ctx, requestChan)
			}(worker)
		}
		s.metrics.SetActiveWorkers(s.plan.Users)
		logger.Log.Info("All workers started immediately",
			zap.Int("workers", s.plan.Users))
		return
	}

	// Calculate workers to start per interval
	rampUpInterval := time.Second
	workersPerInterval := s.plan.Users / s.plan.RampUpSec
	if workersPerInterval < 1 {
		workersPerInterval = 1
		rampUpInterval = time.Duration(s.plan.RampUpSec*1000/s.plan.Users) * time.Millisecond
	}

	ticker := time.NewTicker(rampUpInterval)
	defer ticker.Stop()

	workerCount := 0
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Start batch of workers
			for i := 0; i < workersPerInterval && workerCount < s.plan.Users; i++ {
				worker := NewWorker(workerCount, s.plan, s.metrics)
				s.workers = append(s.workers, worker)
				s.wg.Add(1)
				go func(w *Worker) {
					defer s.wg.Done()
					w.Run(s.ctx, requestChan)
				}(worker)
				workerCount++
			}
			s.metrics.SetActiveWorkers(workerCount)
			logger.Log.Debug("Workers ramped up",
				zap.Int("active_workers", workerCount),
				zap.Int("total_workers", s.plan.Users))

			if workerCount >= s.plan.Users {
				logger.Log.Info("All workers started after ramp-up",
					zap.Int("workers", s.plan.Users))
				return
			}
		}
	}
}

// generateRequests sends requests to workers at a controlled rate
func (s *Scheduler) generateRequests(requestChan chan<- struct{}) {
	defer close(requestChan)

	// Use ticker for rate control - adjust based on concurrent users
	// This creates a constant flow of work for available workers
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			logger.Log.Info("Request generation stopped")
			return
		case <-ticker.C:
			// Non-blocking send to avoid goroutine buildup
			select {
			case requestChan <- struct{}{}:
			default:
				// Channel full, workers are busy
			}
		}
	}
}

// reportMetrics logs metrics every 5 seconds
func (s *Scheduler) reportMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// Calculate final metrics
			s.calculateFinalMetrics()
			return
		case <-ticker.C:
			snapshot := s.metrics.GetSnapshot()
			logger.Log.Info("Metrics update",
				zap.String("run_id", snapshot.RunID),
				zap.Int64("total_requests", snapshot.TotalRequests),
				zap.Int64("success", snapshot.SuccessRequests),
				zap.Int64("failed", snapshot.FailedRequests),
				zap.Float64("avg_latency_ms", snapshot.AvgLatencyMs),
				zap.Int("active_workers", snapshot.ActiveWorkers))
		}
	}
}

// calculateFinalMetrics computes percentiles and final statistics
func (s *Scheduler) calculateFinalMetrics() {
	// Collect all latencies from all workers
	allLatencies := make([]float64, 0)
	for _, worker := range s.workers {
		allLatencies = append(allLatencies, worker.GetLatencies()...)
	}

	if len(allLatencies) == 0 {
		return
	}

	// Sort for percentile calculation
	sort.Float64s(allLatencies)

	// Calculate percentiles
	s.metrics.Mu.Lock()
	s.metrics.P50LatencyMs = percentile(allLatencies, 0.50)
	s.metrics.P75LatencyMs = percentile(allLatencies, 0.75)
	s.metrics.P95LatencyMs = percentile(allLatencies, 0.95)
	s.metrics.P99LatencyMs = percentile(allLatencies, 0.99)

	// Calculate average
	sum := 0.0
	for _, lat := range allLatencies {
		sum += lat
	}
	s.metrics.AvgLatencyMs = sum / float64(len(allLatencies))

	// Calculate RPS
	if s.metrics.TotalDurationMs > 0 {
		s.metrics.RequestsPerSec = float64(s.metrics.TotalRequests) / (float64(s.metrics.TotalDurationMs) / 1000.0)
	}

	s.metrics.Mu.Unlock()

	logger.Log.Info("Final metrics calculated",
		zap.String("run_id", s.metrics.RunID),
		zap.Int64("total_requests", s.metrics.TotalRequests),
		zap.Float64("avg_latency_ms", s.metrics.AvgLatencyMs),
		zap.Float64("p50", s.metrics.P50LatencyMs),
		zap.Float64("p95", s.metrics.P95LatencyMs),
		zap.Float64("p99", s.metrics.P99LatencyMs),
		zap.Float64("rps", s.metrics.RequestsPerSec))
}

// percentile calculates the percentile value from sorted data
func percentile(sortedData []float64, p float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}
	index := int(float64(len(sortedData)) * p)
	if index >= len(sortedData) {
		index = len(sortedData) - 1
	}
	return sortedData[index]
}

// Stop cancels the test execution
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		logger.Log.Info("Stopping test execution")
		s.cancel()
	}
}

// Wait waits for all workers to finish
func (s *Scheduler) Wait() {
	s.wg.Wait()
	logger.Log.Info("All workers finished")
}
