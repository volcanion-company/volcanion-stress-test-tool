package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware logs all HTTP requests with structured fields
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		requestID := GetRequestID(c)
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Build log entry with all fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("body_size", c.Writer.Size()),
		}

		// Add user info if authenticated
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Add error if present
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// Log at appropriate level based on status code
		msg := "HTTP request"
		switch {
		case statusCode >= 500:
			logger.Error(msg, fields...)
		case statusCode >= 400:
			logger.Warn(msg, fields...)
		case statusCode >= 300:
			logger.Info(msg, fields...)
		default:
			logger.Info(msg, fields...)
		}
	}
}

// LoggingMiddlewareWithConfig allows custom configuration
type LoggingConfig struct {
	SkipPaths []string // Paths to skip logging (e.g., /health, /metrics)
	Logger    *zap.Logger
}

// LoggingMiddlewareWithConfig creates a logging middleware with custom config
func LoggingMiddlewareWithConfig(config LoggingConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip logging for specified paths
		if skipPaths[path] {
			c.Next()
			return
		}

		start := time.Now()
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		requestID := GetRequestID(c)

		// Build log entry
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		}

		// Add user info if authenticated
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Log at appropriate level
		msg := "HTTP request"
		switch {
		case statusCode >= 500:
			config.Logger.Error(msg, fields...)
		case statusCode >= 400:
			config.Logger.Warn(msg, fields...)
		default:
			config.Logger.Info(msg, fields...)
		}
	}
}
