package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/auth"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/middleware"
	"github.com/volcanion-company/volcanion-stress-test-tool/internal/storage/repository"
)

// UserRepository interface for user operations
type UserRepository interface {
	GetByUsername(username string) (*auth.User, string, error)
	GetByEmail(email string) (*auth.User, string, error)
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	jwtService      *auth.JWTService
	apiKeyService   *auth.APIKeyService
	userRepo        UserRepository
	passwordService *auth.PasswordService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *auth.JWTService, apiKeyService *auth.APIKeyService) *AuthHandler {
	return &AuthHandler{
		jwtService:      jwtService,
		apiKeyService:   apiKeyService,
		userRepo:        repository.NewMemoryUserRepository(),
		passwordService: auth.NewPasswordService(),
	}
}

// NewAuthHandlerWithRepo creates a new auth handler with custom user repository
func NewAuthHandlerWithRepo(jwtService *auth.JWTService, apiKeyService *auth.APIKeyService, userRepo UserRepository, passwordService *auth.PasswordService) *AuthHandler {
	return &AuthHandler{
		jwtService:      jwtService,
		apiKeyService:   apiKeyService,
		userRepo:        userRepo,
		passwordService: passwordService,
	}
}

// Login handles user login and returns JWT token
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate credentials against user repository
	user, passwordHash, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		// Use generic error to prevent username enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password using bcrypt
	if err := h.passwordService.VerifyPassword(req.Password, passwordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, auth.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	})
}

// CreateAPIKey creates a new API key for the authenticated user
// POST /api/auth/api-keys
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req auth.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get(middleware.AuthUserKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	// Create API key
	apiKey, err := h.apiKeyService.CreateAPIKey(userIDStr, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

// ListAPIKeys returns all API keys for the authenticated user
// GET /api/auth/api-keys
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get(middleware.AuthUserKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	keys := h.apiKeyService.ListAPIKeys(userIDStr)
	c.JSON(http.StatusOK, gin.H{
		"api_keys": keys,
		"count":    len(keys),
	})
}

// RevokeAPIKey revokes an API key
// DELETE /api/auth/api-keys/:id
func (h *AuthHandler) RevokeAPIKey(c *gin.Context) {
	keyID := c.Param("id")

	if err := h.apiKeyService.RevokeAPIKey(keyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}
