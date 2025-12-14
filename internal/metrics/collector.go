package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector holds Prometheus metrics for the stress test tool
type Collector struct {
	RequestDuration *prometheus.HistogramVec
	RequestsTotal   *prometheus.CounterVec
	RequestsFailed  *prometheus.CounterVec
	ActiveTests     prometheus.Gauge
	ActiveWorkers   prometheus.Gauge
}

// NewCollector creates a new metrics collector with Prometheus metrics
func NewCollector() *Collector {
	return &Collector{
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"run_id", "method", "status"},
		),
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"run_id", "method", "status"},
		),
		RequestsFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_failed_total",
				Help: "Total number of failed HTTP requests",
			},
			[]string{"run_id", "method"},
		),
		ActiveTests: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "stress_test_active_tests",
				Help: "Number of currently active stress tests",
			},
		),
		ActiveWorkers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "stress_test_active_workers",
				Help: "Number of currently active workers across all tests",
			},
		),
	}
}

// RecordRequest records a request metric
func (c *Collector) RecordRequest(runID, method, status string, durationSec float64, failed bool) {
	c.RequestDuration.WithLabelValues(runID, method, status).Observe(durationSec)
	c.RequestsTotal.WithLabelValues(runID, method, status).Inc()

	if failed {
		c.RequestsFailed.WithLabelValues(runID, method).Inc()
	}
}

// SetActiveTests sets the number of active tests
func (c *Collector) SetActiveTests(count int) {
	c.ActiveTests.Set(float64(count))
}

// SetActiveWorkers sets the number of active workers
func (c *Collector) SetActiveWorkers(count int) {
	c.ActiveWorkers.Set(float64(count))
}
