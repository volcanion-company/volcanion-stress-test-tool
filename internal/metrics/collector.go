package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector holds Prometheus metrics for the stress test tool
type Collector struct {
	// Test execution metrics
	RequestDuration *prometheus.HistogramVec
	RequestsTotal   *prometheus.CounterVec
	RequestsFailed  *prometheus.CounterVec
	ActiveTests     prometheus.Gauge
	ActiveWorkers   prometheus.Gauge

	// API server metrics
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestsInFlight prometheus.Gauge

	// WebSocket metrics
	WebSocketConnections   prometheus.Gauge
	WebSocketMessagesTotal *prometheus.CounterVec

	// Go runtime metrics
	GoGoroutines prometheus.GaugeFunc
	GoMemAlloc   prometheus.GaugeFunc
	GoMemSys     prometheus.GaugeFunc
	GoGCDuration prometheus.Summary

	// Application info
	BuildInfo *prometheus.GaugeVec
}

var (
	defaultCollector *Collector
	once             sync.Once
)

// NewCollector creates a new metrics collector with Prometheus metrics
// For tests, use GetCollector() to avoid duplicate registration
func NewCollector() *Collector {
	once.Do(func() {
		defaultCollector = &Collector{
			// Test execution metrics
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

			// API server metrics
			HTTPRequestDuration: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "api_http_request_duration_seconds",
					Help:    "API HTTP request latency in seconds",
					Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
				},
				[]string{"method", "path", "status"},
			),
			HTTPRequestsTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "api_http_requests_total",
					Help: "Total number of API HTTP requests",
				},
				[]string{"method", "path", "status"},
			),
			HTTPRequestsInFlight: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "api_http_requests_in_flight",
					Help: "Number of API HTTP requests currently being processed",
				},
			),

			// WebSocket metrics
			WebSocketConnections: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "websocket_connections_active",
					Help: "Number of active WebSocket connections",
				},
			),
			WebSocketMessagesTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "websocket_messages_total",
					Help: "Total number of WebSocket messages sent",
				},
				[]string{"type", "direction"},
			),

			// Go runtime metrics
			GoGoroutines: promauto.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: "go_goroutines_count",
					Help: "Number of goroutines currently existing",
				},
				func() float64 {
					return float64(runtime.NumGoroutine())
				},
			),
			GoMemAlloc: promauto.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: "go_memory_alloc_bytes",
					Help: "Bytes of allocated heap objects",
				},
				func() float64 {
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					return float64(m.Alloc)
				},
			),
			GoMemSys: promauto.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: "go_memory_sys_bytes",
					Help: "Total bytes of memory obtained from the OS",
				},
				func() float64 {
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					return float64(m.Sys)
				},
			),
			GoGCDuration: promauto.NewSummary(
				prometheus.SummaryOpts{
					Name:       "app_gc_duration_seconds",
					Help:       "Application GC invocation durations",
					Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
					MaxAge:     10 * time.Minute,
				},
			),

			// Build info
			BuildInfo: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "app_build_info",
					Help: "Application build information",
				},
				[]string{"version", "go_version", "build_time"},
			),
		}
	})
	return defaultCollector
}

// GetCollector returns the singleton collector (same as NewCollector)
func GetCollector() *Collector {
	return NewCollector()
}

// ResetForTesting resets the collector for testing purposes
// This should only be used in tests to avoid duplicate registration errors
func ResetForTesting() {
	once = sync.Once{}
	defaultCollector = nil
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

// RecordHTTPRequest records an API HTTP request metric
func (c *Collector) RecordHTTPRequest(method, path, status string, durationSec float64) {
	c.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(durationSec)
	c.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// IncrementHTTPRequestsInFlight increments the in-flight counter
func (c *Collector) IncrementHTTPRequestsInFlight() {
	c.HTTPRequestsInFlight.Inc()
}

// DecrementHTTPRequestsInFlight decrements the in-flight counter
func (c *Collector) DecrementHTTPRequestsInFlight() {
	c.HTTPRequestsInFlight.Dec()
}

// SetWebSocketConnections sets the number of active WebSocket connections
func (c *Collector) SetWebSocketConnections(count int) {
	c.WebSocketConnections.Set(float64(count))
}

// RecordWebSocketMessage records a WebSocket message
func (c *Collector) RecordWebSocketMessage(msgType, direction string) {
	c.WebSocketMessagesTotal.WithLabelValues(msgType, direction).Inc()
}

// RecordGCDuration records a GC duration
func (c *Collector) RecordGCDuration(durationSec float64) {
	c.GoGCDuration.Observe(durationSec)
}

// SetBuildInfo sets the build information
func (c *Collector) SetBuildInfo(version, goVersion, buildTime string) {
	c.BuildInfo.WithLabelValues(version, goVersion, buildTime).Set(1)
}
