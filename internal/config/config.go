package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Config holds application configuration
type Config struct {
	Environment             string // "development", "staging", "production"
	ServerPort              string
	LogLevel                string
	MaxWorkers              int
	DefaultTimeout          int
	MetricsEnabled          bool
	DatabaseDSN             string
	DatabaseMaxConns        int
	DatabaseMaxIdleConns    int
	JWTSecret               string
	JWTDuration             int // in hours
	AuthEnabled             bool
	RateLimitEnabled        bool
	RateLimitPerSecond      float64
	AllowedOrigins          []string
	AllowedWebSocketOrigins []string
	CORSAllowCredentials    bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	cfg := &Config{
		Environment:             getEnv("ENVIRONMENT", "development"),
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		MaxWorkers:              getEnvAsInt("MAX_WORKERS", 1000),
		DefaultTimeout:          getEnvAsInt("DEFAULT_TIMEOUT_MS", 30000),
		MetricsEnabled:          getEnvAsBool("METRICS_ENABLED", true),
		DatabaseDSN:             getEnv("DATABASE_DSN", ""),
		DatabaseMaxConns:        getEnvAsInt("DATABASE_MAX_CONNS", 25),
		DatabaseMaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		JWTDuration:             getEnvAsInt("JWT_DURATION_HOURS", 24),
		AuthEnabled:             getEnvAsBool("AUTH_ENABLED", true),
		RateLimitEnabled:        getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitPerSecond:      getEnvAsFloat("RATE_LIMIT_PER_SECOND", 10.0),
		AllowedOrigins:          getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		AllowedWebSocketOrigins: getEnvAsSlice("ALLOWED_WEBSOCKET_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		CORSAllowCredentials:    getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
	}

	// Validate and set JWT secret
	cfg.validateJWTSecret()

	return cfg
}

// validateJWTSecret ensures JWT secret is set and secure
func (c *Config) validateJWTSecret() {
	if c.JWTSecret == "" {
		// Generate a random secret if not provided (for development only)
		secret := generateRandomSecret(32)
		c.JWTSecret = secret
		zap.L().Warn("JWT_SECRET not set, generated random secret. Set JWT_SECRET environment variable in production!",
			zap.String("generated_secret_preview", secret[:8]+"..."))
	} else if len(c.JWTSecret) < 32 {
		zap.L().Warn("JWT_SECRET is less than 32 characters, consider using a longer secret for better security")
	}

	// Check for default/weak secrets
	weakSecrets := []string{
		"default-secret-change-in-production",
		"secret",
		"jwt-secret",
		"changeme",
	}
	for _, weak := range weakSecrets {
		if c.JWTSecret == weak {
			zap.L().Error("SECURITY WARNING: Using weak JWT secret. Please set a strong JWT_SECRET environment variable!")
			break
		}
	}
}

// generateRandomSecret generates a cryptographically secure random secret
func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback - this should never happen
		return fmt.Sprintf("fallback-secret-%d", os.Getpid())
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// IsOriginAllowed checks if an origin is in the allowed list
func (c *Config) IsOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// IsWebSocketOriginAllowed checks if a WebSocket origin is allowed
func (c *Config) IsWebSocketOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedWebSocketOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
