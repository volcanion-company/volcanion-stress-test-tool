package engine

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/volcanion-company/volcanion-stress-test-tool/internal/domain/model"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/logger"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
	"go.uber.org/zap"
)

// Scheduler manages the execution of a test run with workers and rate control
type Scheduler struct {
	plan         *model.TestPlan
	metrics      *model.Metrics
	workers      []*Worker
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	ctx          context.Context
	sharedClient *http.Client
	collector    *metrics.Collector
}

// NewScheduler creates a new scheduler for a test plan
func NewScheduler(plan *model.TestPlan, metrics *model.Metrics, sharedClient *http.Client, collector *metrics.Collector) *Scheduler {
	return &Scheduler{
		plan:         plan,
		metrics:      metrics,
		workers:      make([]*Worker, 0, plan.Users),
		sharedClient: sharedClient,
		collector:    collector,
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
	go s.generateRequestsWithPattern(requestChan)

	// Start metrics reporter
	go s.reportMetrics()

	return nil
}

// startWorkersWithRampUp gradually spawns workers according to ramp-up time
func (s *Scheduler) startWorkersWithRampUp(requestChan <-chan struct{}) {
	if s.plan.RampUpSec == 0 {
		// No ramp-up, start all workers immediately
		for i := 0; i < s.plan.Users; i++ {
			worker := NewWorker(i, s.plan, s.metrics, s.sharedClient, s.collector)
			s.workers = append(s.workers, worker)
			s.wg.Add(1)
			go func(w *Worker) {
				defer s.wg.Done()
				w.Run(s.ctx, requestChan)
			}(worker)
		}
		s.metrics.SetActiveWorkers(s.plan.Users)
		s.collector.SetActiveWorkers(s.plan.Users)
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
				worker := NewWorker(workerCount, s.plan, s.metrics, s.sharedClient, s.collector)
				s.workers = append(s.workers, worker)
				s.wg.Add(1)
				go func(w *Worker) {
					defer s.wg.Done()
					w.Run(s.ctx, requestChan)
				}(worker)
				workerCount++
			}
			s.metrics.SetActiveWorkers(workerCount)
			s.collector.SetActiveWorkers(workerCount)
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

// generateRequestsWithPattern sends requests based on rate pattern
func (s *Scheduler) generateRequestsWithPattern(requestChan chan<- struct{}) {
	defer close(requestChan)

	switch s.plan.RatePattern {
	case model.RatePatternStep:
		s.generateStepPattern(requestChan)
	case model.RatePatternSpike:
		s.generateSpikePattern(requestChan)
	case model.RatePatternRamp:
		s.generateRampPattern(requestChan)
	default: // RatePatternFixed or empty
		s.generateFixedRate(requestChan)
	}
}

// generateFixedRate generates requests at a fixed rate
func (s *Scheduler) generateFixedRate(requestChan chan<- struct{}) {
	// Calculate request interval based on target RPS
	var ticker *time.Ticker
	if s.plan.TargetRPS > 0 {
		// Calculate interval: 1 second / target RPS
		intervalNs := int64(1e9) / int64(s.plan.TargetRPS)
		ticker = time.NewTicker(time.Duration(intervalNs))
		logger.Log.Info("Rate control enabled (fixed)",
			zap.Int("target_rps", s.plan.TargetRPS),
			zap.Duration("interval", time.Duration(intervalNs)))
	} else {
		// No rate limit: use 1ms ticker for fast generation
		ticker = time.NewTicker(time.Millisecond)
		logger.Log.Info("Rate control disabled (unlimited RPS)")
	}
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

// generateStepPattern generates requests with step increases
func (s *Scheduler) generateStepPattern(requestChan chan<- struct{}) {
	if len(s.plan.RateSteps) == 0 {
		logger.Log.Warn("No rate steps defined, falling back to fixed rate")
		s.generateFixedRate(requestChan)
		return
	}

	logger.Log.Info("Starting step rate pattern",
		zap.Int("steps", len(s.plan.RateSteps)))

	for stepIdx, step := range s.plan.RateSteps {
		logger.Log.Info("Step rate change",
			zap.Int("step", stepIdx+1),
			zap.Int("rps", step.RPS),
			zap.Int("duration_sec", step.DurationSec))

		s.runRateForDuration(requestChan, step.RPS, step.DurationSec)

		select {
		case <-s.ctx.Done():
			logger.Log.Info("Step pattern stopped early")
			return
		default:
		}
	}

	logger.Log.Info("Step pattern completed, maintaining last rate")
	// Maintain last rate for remaining duration
	if len(s.plan.RateSteps) > 0 {
		lastStep := s.plan.RateSteps[len(s.plan.RateSteps)-1]
		s.runRateIndefinitely(requestChan, lastStep.RPS)
	}
}

// generateSpikePattern generates a spike then returns to base
func (s *Scheduler) generateSpikePattern(requestChan chan<- struct{}) {
	if len(s.plan.RateSteps) < 2 {
		logger.Log.Warn("Spike pattern requires at least 2 steps (base, spike), falling back to fixed")
		s.generateFixedRate(requestChan)
		return
	}

	baseRate := s.plan.RateSteps[0]
	spikeRate := s.plan.RateSteps[1]

	logger.Log.Info("Starting spike pattern",
		zap.Int("base_rps", baseRate.RPS),
		zap.Int("spike_rps", spikeRate.RPS),
		zap.Int("spike_duration_sec", spikeRate.DurationSec))

	// Base rate
	s.runRateForDuration(requestChan, baseRate.RPS, baseRate.DurationSec)

	// Spike
	select {
	case <-s.ctx.Done():
		return
	default:
	}
	s.runRateForDuration(requestChan, spikeRate.RPS, spikeRate.DurationSec)

	// Back to base
	select {
	case <-s.ctx.Done():
		return
	default:
	}
	s.runRateIndefinitely(requestChan, baseRate.RPS)
}

// generateRampPattern linearly increases rate over duration
func (s *Scheduler) generateRampPattern(requestChan chan<- struct{}) {
	startRPS := 1
	endRPS := s.plan.TargetRPS
	if endRPS == 0 {
		endRPS = 100 // Default if not specified
	}

	logger.Log.Info("Starting ramp pattern",
		zap.Int("start_rps", startRPS),
		zap.Int("end_rps", endRPS),
		zap.Int("duration_sec", s.plan.DurationSec))

	// Ramp up over 50% of duration, then maintain
	rampDuration := s.plan.DurationSec / 2
	if rampDuration < 1 {
		rampDuration = 1
	}

	// Increase RPS every second
	step := float64(endRPS-startRPS) / float64(rampDuration)
	currentRPS := float64(startRPS)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	elapsed := 0

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			elapsed++
			if elapsed <= rampDuration {
				currentRPS += step
				logger.Log.Debug("Ramp rate change",
					zap.Float64("current_rps", currentRPS))
			}

			// Generate requests for this second
			targetRPS := int(currentRPS)
			if targetRPS > 0 {
				interval := time.Second / time.Duration(targetRPS)
				s.sendRequestsForInterval(requestChan, interval, time.Second)
			}
		}
	}
}

// runRateForDuration runs at specified RPS for duration
func (s *Scheduler) runRateForDuration(requestChan chan<- struct{}, rps int, durationSec int) {
	if rps <= 0 {
		time.Sleep(time.Duration(durationSec) * time.Second)
		return
	}

	intervalNs := int64(1e9) / int64(rps)
	ticker := time.NewTicker(time.Duration(intervalNs))
	defer ticker.Stop()

	timeout := time.After(time.Duration(durationSec) * time.Second)

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
			select {
			case requestChan <- struct{}{}:
			default:
			}
		}
	}
}

// runRateIndefinitely runs at specified RPS until context done
func (s *Scheduler) runRateIndefinitely(requestChan chan<- struct{}, rps int) {
	if rps <= 0 {
		<-s.ctx.Done()
		return
	}

	intervalNs := int64(1e9) / int64(rps)
	ticker := time.NewTicker(time.Duration(intervalNs))
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			select {
			case requestChan <- struct{}{}:
			default:
			}
		}
	}
}

// sendRequestsForInterval sends requests at interval for duration
func (s *Scheduler) sendRequestsForInterval(requestChan chan<- struct{}, interval, duration time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
			select {
			case requestChan <- struct{}{}:
			default:
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
			// Update live metrics (RPS and duration)
			s.metrics.UpdateLiveMetrics()

			snapshot := s.metrics.GetSnapshot()
			logger.Log.Info("Metrics update",
				zap.String("run_id", snapshot.RunID),
				zap.Int64("total_requests", snapshot.TotalRequests),
				zap.Int64("success", snapshot.SuccessRequests),
				zap.Int64("failed", snapshot.FailedRequests),
				zap.Float64("avg_latency_ms", snapshot.AvgLatencyMs),
				zap.Float64("current_rps", snapshot.CurrentRPS),
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
