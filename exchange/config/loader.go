// Package config provides configuration loading utilities
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadFromFile loads configuration from a .env file
func LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set environment variable
		os.Setenv(key, value)
	}

	return scanner.Err()
}

// LoadConfigForEnvironment loads configuration for a specific environment
func LoadConfigForEnvironment(serviceName, environment string) (*Config, error) {
	configDir := getConfigDir()

	// Try service-specific environment file first
	serviceEnvFile := filepath.Join(configDir, fmt.Sprintf("%s-%s.env", serviceName, environment))
	if _, err := os.Stat(serviceEnvFile); err == nil {
		if err := LoadFromFile(serviceEnvFile); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", serviceEnvFile, err)
		}
	}

	// Load configuration (will use environment variables)
	config := LoadConfig(serviceName)

	// Basic validation
	if config.Service.Port == "" {
		return nil, fmt.Errorf("service port is required")
	}

	return config, nil
}

// getConfigDir returns the configuration directory
func getConfigDir() string {
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return dir
	}
	return "config"
}
