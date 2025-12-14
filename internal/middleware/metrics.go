package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/metrics"
)

// MetricsMiddleware records API endpoint metrics
func MetricsMiddleware(collector *metrics.Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track in-flight requests
		collector.IncrementHTTPRequestsInFlight()
		defer collector.DecrementHTTPRequestsInFlight()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get path template (use the registered route, not the actual path with params)
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		collector.RecordHTTPRequest(method, path, status, duration)
	}
}

// MetricsMiddlewareConfig allows custom configuration
type MetricsMiddlewareConfig struct {
	Collector *metrics.Collector
	SkipPaths []string
}

// MetricsMiddlewareWithConfig creates a metrics middleware with custom config
func MetricsMiddlewareWithConfig(config MetricsMiddlewareConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip metrics for specified paths (like /metrics itself)
		if skipPaths[path] {
			c.Next()
			return
		}

		// Track in-flight requests
		config.Collector.IncrementHTTPRequestsInFlight()
		defer config.Collector.DecrementHTTPRequestsInFlight()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get path template
		fullPath := c.FullPath()
		if fullPath == "" {
			fullPath = "unknown"
		}

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		config.Collector.RecordHTTPRequest(method, fullPath, status, duration)
	}
}
