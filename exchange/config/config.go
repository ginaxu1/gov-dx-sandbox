// Package config provides simplified configuration management
package config

import (
	"flag"
	"os"
	"time"
)

// Config holds all configuration for a service
type Config struct {
	Environment string
	Service     ServiceConfig
	Logging     LoggingConfig
	Security    SecurityConfig
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name    string
	Port    string
	Host    string
	Timeout time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret  string
	EnableCORS bool
	RateLimit  int
}

// LoadConfig loads configuration from flags and environment variables
func LoadConfig(serviceName string) *Config {
	// Define flags
	env := flag.String("env", getEnvOrDefault("ENVIRONMENT", "local"), "Environment: local or production")
	port := flag.String("port", getDefaultPort(serviceName), "Service port")
	host := flag.String("host", getEnvOrDefault("HOST", "0.0.0.0"), "Host address")
	timeout := flag.Duration("timeout", 10*time.Second, "Request timeout")
	logLevel := flag.String("log-level", getDefaultLogLevel(*env), "Log level")
	logFormat := flag.String("log-format", getDefaultLogFormat(*env), "Log format")
	jwtSecret := flag.String("jwt-secret", getDefaultJWTSecret(*env), "JWT secret")
	enableCORS := flag.Bool("cors", getDefaultCORS(*env), "Enable CORS")
	rateLimit := flag.Int("rate-limit", getDefaultRateLimit(*env), "Rate limit per minute")

	// Parse flags
	flag.Parse()

	return &Config{
		Environment: *env,
		Service: ServiceConfig{
			Name:    serviceName,
			Port:    *port,
			Host:    *host,
			Timeout: *timeout,
		},
		Logging: LoggingConfig{
			Level:  *logLevel,
			Format: *logFormat,
		},
		Security: SecurityConfig{
			JWTSecret:  *jwtSecret,
			EnableCORS: *enableCORS,
			RateLimit:  *rateLimit,
		},
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

func getDefaultLogLevel(env string) string {
	if env == "production" {
		return "warn"
	}
	return "debug"
}

func getDefaultLogFormat(env string) string {
	if env == "production" {
		return "json"
	}
	return "text"
}

func getDefaultJWTSecret(env string) string {
	if env == "production" {
		return ""
	}
	return "local-secret-key"
}

func getDefaultCORS(env string) bool {
	return env != "production"
}

func getDefaultRateLimit(env string) int {
	if env == "production" {
		return 100
	}
	return 1000
}
