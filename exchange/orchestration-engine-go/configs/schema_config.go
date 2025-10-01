package configs

import (
	"os"
	"strconv"
)

// SchemaConfig holds configuration for schema management
type SchemaConfig struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	Schema   SchemaSettings `json:"schema"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// SchemaSettings holds schema-specific settings
type SchemaSettings struct {
	MaxVersions        int    `json:"maxVersions"`
	CompatibilityCheck bool   `json:"compatibilityCheck"`
	AutoActivate       bool   `json:"autoActivate"`
	DefaultVersion     string `json:"defaultVersion"`
}

// LoadSchemaConfig loads configuration from environment variables
func LoadSchemaConfig() *SchemaConfig {
	return &SchemaConfig{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "orchestration_engine"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SCHEMA_SERVER_PORT", "8081"),
			Host: getEnv("SCHEMA_SERVER_HOST", "0.0.0.0"),
		},
		Schema: SchemaSettings{
			MaxVersions:        getEnvAsInt("SCHEMA_MAX_VERSIONS", 10),
			CompatibilityCheck: getEnvAsBool("SCHEMA_COMPATIBILITY_CHECK", true),
			AutoActivate:       getEnvAsBool("SCHEMA_AUTO_ACTIVATE", false),
			DefaultVersion:     getEnv("SCHEMA_DEFAULT_VERSION", "latest"),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
