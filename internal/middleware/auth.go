package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
	"go.uber.org/zap"
)

const (
	AuthUserKey = "auth_user"
	AuthRoleKey = "auth_role"
)

// AuthMiddleware creates authentication middleware supporting both JWT and API keys
func AuthMiddleware(jwtService *auth.JWTService, apiKeyService *auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT token first (Bearer token)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				zap.L().Warn("Invalid JWT token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				c.Abort()
				return
			}

			// Set user info in context
			c.Set(AuthUserKey, claims.UserID)
			c.Set(AuthRoleKey, claims.Role)
			c.Next()
			return
		}

		// Try API key (X-API-Key header)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			keyInfo, err := apiKeyService.ValidateAPIKey(apiKey)
			if err != nil {
				zap.L().Warn("Invalid API key", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired API key"})
				c.Abort()
				return
			}

			// Set user info in context
			c.Set(AuthUserKey, keyInfo.UserID)
			c.Set(AuthRoleKey, keyInfo.Role)
			c.Next()
			return
		}

		// No authentication provided
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		c.Abort()
	}
}

// RequireRole creates middleware that checks for specific role(s)
func RequireRole(allowedRoles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(AuthRoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userRole, ok := role.(auth.Role)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid role format"})
			c.Abort()
			return
		}

		// Admin has access to everything
		if userRole == auth.RoleAdmin {
			c.Next()
			return
		}

		// Check if user's role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

// OptionalAuth middleware allows both authenticated and unauthenticated requests
func OptionalAuth(jwtService *auth.JWTService, apiKeyService *auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)
			if err == nil {
				c.Set(AuthUserKey, claims.UserID)
				c.Set(AuthRoleKey, claims.Role)
			}
		}

		// Try API key if JWT didn't work
		if _, exists := c.Get(AuthUserKey); !exists {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey != "" {
				keyInfo, err := apiKeyService.ValidateAPIKey(apiKey)
				if err == nil {
					c.Set(AuthUserKey, keyInfo.UserID)
					c.Set(AuthRoleKey, keyInfo.Role)
				}
			}
		}

		c.Next()
	}
}
