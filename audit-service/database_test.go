package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Host)
	assert.NotEmpty(t, config.Port)
	assert.NotEmpty(t, config.Username)
	assert.NotEmpty(t, config.Password)
	assert.NotEmpty(t, config.Database)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, config.ConnMaxIdleTime)
}

func TestNewDatabaseConfig_WithEnvVars(t *testing.T) {
	os.Setenv("CHOREO_OPENDIF_DATABASE_HOSTNAME", "test-host")
	os.Setenv("CHOREO_OPENDIF_DATABASE_PORT", "5433")
	os.Setenv("CHOREO_OPENDIF_DATABASE_USERNAME", "test-user")
	os.Setenv("CHOREO_OPENDIF_DATABASE_PASSWORD", "test-pass")
	os.Setenv("CHOREO_OPENDIF_DATABASE_DATABASENAME", "test-db")
	defer func() {
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_HOSTNAME")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_PORT")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_USERNAME")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_PASSWORD")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_DATABASENAME")
	}()

	config := NewDatabaseConfig()
	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "5433", config.Port)
	assert.Equal(t, "test-user", config.Username)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.Database)
}

func TestParseIntOrDefault(t *testing.T) {
	t.Run("Returns parsed int when valid", func(t *testing.T) {
		key := "TEST_INT_VAR_12345"
		os.Setenv(key, "42")
		defer os.Unsetenv(key)

		result := parseIntOrDefault(key, 10)
		assert.Equal(t, 42, result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		key := "TEST_INT_VAR_NONEXISTENT_12345"
		os.Unsetenv(key)

		result := parseIntOrDefault(key, 10)
		assert.Equal(t, 10, result)
	})

	t.Run("Returns default when invalid", func(t *testing.T) {
		key := "TEST_INT_VAR_INVALID_12345"
		os.Setenv(key, "not-a-number")
		defer os.Unsetenv(key)

		result := parseIntOrDefault(key, 10)
		assert.Equal(t, 10, result)
	})
}

func TestParseDurationOrDefault(t *testing.T) {
	t.Run("Returns parsed duration when valid", func(t *testing.T) {
		key := "TEST_DURATION_VAR_12345"
		os.Setenv(key, "2h")
		defer os.Unsetenv(key)

		result := parseDurationOrDefault(key, "1h")
		assert.Equal(t, 2*time.Hour, result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		key := "TEST_DURATION_VAR_NONEXISTENT_12345"
		os.Unsetenv(key)

		result := parseDurationOrDefault(key, "1h")
		assert.Equal(t, 1*time.Hour, result)
	})

	t.Run("Returns default when invalid", func(t *testing.T) {
		key := "TEST_DURATION_VAR_INVALID_12345"
		os.Setenv(key, "invalid-duration")
		defer os.Unsetenv(key)

		result := parseDurationOrDefault(key, "1h")
		assert.Equal(t, 1*time.Hour, result)
	})
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("Returns env var when set", func(t *testing.T) {
		key := "TEST_ENV_VAR_12345"
		os.Setenv(key, "test-value")
		defer os.Unsetenv(key)

		result := getEnvOrDefault(key, "default")
		assert.Equal(t, "test-value", result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		key := "TEST_ENV_VAR_NONEXISTENT_12345"
		os.Unsetenv(key)

		result := getEnvOrDefault(key, "default-value")
		assert.Equal(t, "default-value", result)
	})

	t.Run("Returns default when empty string", func(t *testing.T) {
		key := "TEST_ENV_VAR_EMPTY_12345"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		result := getEnvOrDefault(key, "default")
		assert.Equal(t, "default", result)
	})
}

func TestConnectGORM_InvalidConnection(t *testing.T) {
	config := &DatabaseConfig{
		Host:          "invalid-host",
		Port:          "5432",
		Username:      "invalid-user",
		Password:      "invalid-password",
		Database:      "invalid-db",
		SSLMode:       "disable",
		RetryAttempts: 1, // Set to 1 to fail fast
		RetryDelay:    100 * time.Millisecond,
	}

	_, err := ConnectGORM(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open GORM database connection")
}
