package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware returns a custom recovery middleware that logs panics
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log the panic with full context
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.ByteString("stack", stack),
				)

				// Return 500 error to client
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred. Please try again later.",
				})
			}
		}()
		c.Next()
	}
}

// RecoveryWithCallback returns recovery middleware with custom callback
func RecoveryWithCallback(logger *zap.Logger, callback func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", stack),
				)

				// Call custom callback if provided
				if callback != nil {
					callback(c, err)
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// PanicInfo holds information about a recovered panic
type PanicInfo struct {
	Error      string `json:"error"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	ClientIP   string `json:"client_ip"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// GetPanicInfo extracts panic information (for testing/debugging)
func GetPanicInfo(err interface{}, c *gin.Context) *PanicInfo {
	return &PanicInfo{
		Error:      fmt.Sprintf("%v", err),
		Path:       c.Request.URL.Path,
		Method:     c.Request.Method,
		ClientIP:   c.ClientIP(),
		StackTrace: string(debug.Stack()),
	}
}
