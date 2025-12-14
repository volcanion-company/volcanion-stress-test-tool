package model

import (
	"sync"
	"time"
)

// Metrics holds the results of a test run
type Metrics struct {
	RunID           string           `json:"run_id"`
	TotalRequests   int64            `json:"total_requests"`
	SuccessRequests int64            `json:"success_requests"`
	FailedRequests  int64            `json:"failed_requests"`
	TotalDurationMs int64            `json:"total_duration_ms"`
	MinLatencyMs    float64          `json:"min_latency_ms"`
	MaxLatencyMs    float64          `json:"max_latency_ms"`
	AvgLatencyMs    float64          `json:"avg_latency_ms"`
	P50LatencyMs    float64          `json:"p50_latency_ms"`
	P75LatencyMs    float64          `json:"p75_latency_ms"`
	P95LatencyMs    float64          `json:"p95_latency_ms"`
	P99LatencyMs    float64          `json:"p99_latency_ms"`
	RequestsPerSec  float64          `json:"requests_per_sec"`
	CurrentRPS      float64          `json:"current_rps"`
	ActiveWorkers   int              `json:"active_workers"`
	StatusCodes     map[int]int64    `json:"status_codes"`
	Errors          map[string]int64 `json:"errors,omitempty"`
	LastUpdated     time.Time        `json:"last_updated"`
	Mu              sync.RWMutex     `json:"-"`
}

// NewMetrics creates a new Metrics instance
func NewMetrics(runID string) *Metrics {
	return &Metrics{
		RunID:        runID,
		MinLatencyMs: -1,
		StatusCodes:  make(map[int]int64),
		Errors:       make(map[string]int64),
		LastUpdated:  time.Now(),
	}
}

// RecordRequest records a single request result
func (m *Metrics) RecordRequest(success bool, latencyMs float64, statusCode int, err error) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	m.TotalRequests++

	if success {
		m.SuccessRequests++
	} else {
		m.FailedRequests++
		if err != nil {
			m.Errors[err.Error()]++
		}
	}

	if statusCode > 0 {
		m.StatusCodes[statusCode]++
	}

	// Update latency stats
	if m.MinLatencyMs < 0 || latencyMs < m.MinLatencyMs {
		m.MinLatencyMs = latencyMs
	}
	if latencyMs > m.MaxLatencyMs {
		m.MaxLatencyMs = latencyMs
	}

	m.LastUpdated = time.Now()
}

// SetActiveWorkers updates the number of active workers
func (m *Metrics) SetActiveWorkers(count int) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.ActiveWorkers = count
}

// GetSnapshot returns a copy of current metrics (thread-safe)
func (m *Metrics) GetSnapshot() *Metrics {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	snapshot := &Metrics{
		RunID:           m.RunID,
		TotalRequests:   m.TotalRequests,
		SuccessRequests: m.SuccessRequests,
		FailedRequests:  m.FailedRequests,
		TotalDurationMs: m.TotalDurationMs,
		MinLatencyMs:    m.MinLatencyMs,
		MaxLatencyMs:    m.MaxLatencyMs,
		AvgLatencyMs:    m.AvgLatencyMs,
		P50LatencyMs:    m.P50LatencyMs,
		P75LatencyMs:    m.P75LatencyMs,
		P95LatencyMs:    m.P95LatencyMs,
		P99LatencyMs:    m.P99LatencyMs,
		RequestsPerSec:  m.RequestsPerSec,
		CurrentRPS:      m.CurrentRPS,
		ActiveWorkers:   m.ActiveWorkers,
		StatusCodes:     make(map[int]int64),
		Errors:          make(map[string]int64),
		LastUpdated:     m.LastUpdated,
	}

	for k, v := range m.StatusCodes {
		snapshot.StatusCodes[k] = v
	}
	for k, v := range m.Errors {
		snapshot.Errors[k] = v
	}

	return snapshot
}

// LatencyRecord holds individual request latency for percentile calculation
type LatencyRecord struct {
	Timestamp time.Time
	Latency   float64
}
