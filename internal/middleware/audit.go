package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/audit"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
)

// AuditMiddleware creates middleware that logs all API requests
func AuditMiddleware(auditLogger *audit.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Get user info from context (if authenticated)
		userID, _ := c.Get(AuthUserKey)
		role, _ := c.Get(AuthRoleKey)

		userIDStr := ""
		roleStr := ""

		if userID != nil {
			if uid, ok := userID.(string); ok {
				userIDStr = uid
			}
		}

		if role != nil {
			roleStr = string(role.(auth.Role))
		}

		// Create audit event
		event := audit.AuditEvent{
			UserID:     userIDStr,
			Role:       roleStr,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			Duration:   duration,
		}

		// Add error if request failed
		if len(c.Errors) > 0 {
			event.Error = c.Errors.String()
		}

		// Determine event type based on path and method
		event.EventType = determineEventType(c.Request.Method, c.Request.URL.Path, c.Writer.Status())

		// Log the event
		auditLogger.Log(event)
	}
}

// determineEventType infers event type from HTTP method and path
func determineEventType(method, path string, statusCode int) audit.EventType {
	// Check for unauthorized access
	if statusCode == 401 || statusCode == 403 {
		return audit.EventUnauthorizedAccess
	}

	// Map common patterns
	const (
		methodPost   = "POST"
		methodDelete = "DELETE"
	)

	switch {
	case method == methodPost && contains(path, "/test-plans"):
		return audit.EventTestPlanCreated
	case method == methodDelete && contains(path, "/test-plans"):
		return audit.EventTestPlanDeleted
	case method == methodPost && contains(path, "/test-runs/start"):
		return audit.EventTestRunStarted
	case method == methodPost && contains(path, "/test-runs") && contains(path, "/stop"):
		return audit.EventTestRunStopped
	case method == methodPost && contains(path, "/scenarios") && !contains(path, "/execute"):
		return audit.EventScenarioCreated
	case method == methodPost && contains(path, "/scenarios/execute"):
		return audit.EventScenarioExecuted
	case method == methodDelete && contains(path, "/scenarios"):
		return audit.EventScenarioDeleted
	default:
		return audit.EventType("api.request")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
