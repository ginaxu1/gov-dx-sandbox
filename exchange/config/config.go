// Package config provides configuration management for all services
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the deployment environment
type Environment string

const (
	Local      Environment = "local"
	Production Environment = "production"
)

// Config holds all configuration for a service
type Config struct {
	Environment Environment    `json:"environment"`
	Service     ServiceConfig  `json:"service"`
	Database    DatabaseConfig `json:"database"`
	Logging     LoggingConfig  `json:"logging"`
	Security    SecurityConfig `json:"security"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name         string        `json:"name"`
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
	MaxConns int    `json:"max_connections"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret      string   `json:"jwt_secret"`
	EnableCORS     bool     `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`
	RateLimit      int      `json:"rate_limit_per_minute"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig(serviceName string) *Config {
	env := getEnvironment()

	return &Config{
		Environment: env,
		Service: ServiceConfig{
			Name:         serviceName,
			Port:         getEnvOrDefault("PORT", getDefaultPort(serviceName)),
			Host:         getEnvOrDefault("HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnvOrDefault("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnvOrDefault("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnvOrDefault("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", getDefaultDBHost(env)),
			Port:     getIntEnvOrDefault("DB_PORT", 5432),
			User:     getEnvOrDefault("DB_USER", getDefaultDBUser(env)),
			Password: getEnvOrDefault("DB_PASSWORD", getDefaultDBPassword(env)),
			Name:     getEnvOrDefault("DB_NAME", fmt.Sprintf("%s_%s", serviceName, env)),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", getDefaultSSLMode(env)),
			MaxConns: getIntEnvOrDefault("DB_MAX_CONNS", getDefaultMaxConns(env)),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", getDefaultLogLevel(env)),
			Format: getEnvOrDefault("LOG_FORMAT", getDefaultLogFormat(env)),
		},
		Security: SecurityConfig{
			JWTSecret:      getEnvOrDefault("JWT_SECRET", getDefaultJWTSecret(env)),
			EnableCORS:     getBoolEnvOrDefault("ENABLE_CORS", env == Local),
			AllowedOrigins: getStringSliceEnvOrDefault("ALLOWED_ORIGINS", getDefaultOrigins(env)),
			RateLimit:      getIntEnvOrDefault("RATE_LIMIT_PER_MINUTE", getDefaultRateLimit(env)),
		},
	}
}

// Helper functions
func getEnvironment() Environment {
	env := strings.ToLower(getEnvOrDefault("ENVIRONMENT", "local"))
	if env == "production" || env == "prod" {
		return Production
	}
	return Local
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnvOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Default value functions
func getDefaultPort(serviceName string) string {
	ports := map[string]string{
		"consent-engine":        "8081",
		"policy-decision-point": "8082",
	}
	if port, exists := ports[serviceName]; exists {
		return port
	}
	return "8080"
}

func getDefaultDBHost(env Environment) string {
	if env == Local {
		return "localhost"
	}
	return "prod-db.example.com"
}

func getDefaultDBUser(env Environment) string {
	if env == Local {
		return "postgres"
	}
	return "app_user"
}

func getDefaultDBPassword(env Environment) string {
	if env == Local {
		return "password"
	}
	return ""
}

func getDefaultSSLMode(env Environment) string {
	if env == Local {
		return "disable"
	}
	return "require"
}

func getDefaultMaxConns(env Environment) int {
	if env == Local {
		return 5
	}
	return 50
}

func getDefaultLogLevel(env Environment) string {
	if env == Local {
		return "debug"
	}
	return "warn"
}

func getDefaultLogFormat(env Environment) string {
	if env == Local {
		return "text"
	}
	return "json"
}

func getDefaultJWTSecret(env Environment) string {
	if env == Local {
		return "local-secret-key"
	}
	return ""
}

func getDefaultOrigins(env Environment) []string {
	if env == Local {
		return []string{"http://localhost:3000", "http://localhost:3001"}
	}
	return []string{"https://example.com"}
}

func getDefaultRateLimit(env Environment) int {
	if env == Local {
		return 1000
	}
	return 100
}
