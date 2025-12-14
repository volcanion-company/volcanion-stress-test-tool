package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	ServerPort     string
	LogLevel       string
	MaxWorkers     int
	DefaultTimeout int
	MetricsEnabled bool
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MaxWorkers:     getEnvAsInt("MAX_WORKERS", 1000),
		DefaultTimeout: getEnvAsInt("DEFAULT_TIMEOUT_MS", 30000),
		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
	}
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
