package store

import (

	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "postgres", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "consent_engine", config.Database)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
}

func TestNewDatabaseConfig_WithEnvVars(t *testing.T) {
	os.Setenv("CHOREO_DB_CE_HOSTNAME", "test-host")
	os.Setenv("CHOREO_DB_CE_PORT", "5433")
	os.Setenv("CHOREO_DB_CE_USERNAME", "test-user")
	os.Setenv("CHOREO_DB_CE_PASSWORD", "test-pass")
	os.Setenv("CHOREO_DB_CE_DATABASENAME", "test-db")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_CONN_MAX_LIFETIME", "2h")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1h")
	os.Setenv("DB_QUERY_TIMEOUT", "60s")
	os.Setenv("DB_CONNECT_TIMEOUT", "20s")
	os.Setenv("DB_RETRY_ATTEMPTS", "5")
	os.Setenv("DB_RETRY_DELAY", "2s")
	defer func() {
		os.Unsetenv("CHOREO_DB_CE_HOSTNAME")
		os.Unsetenv("CHOREO_DB_CE_PORT")
		os.Unsetenv("CHOREO_DB_CE_USERNAME")
		os.Unsetenv("CHOREO_DB_CE_PASSWORD")
		os.Unsetenv("CHOREO_DB_CE_DATABASENAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("DB_CONN_MAX_IDLE_TIME")
		os.Unsetenv("DB_QUERY_TIMEOUT")
		os.Unsetenv("DB_CONNECT_TIMEOUT")
		os.Unsetenv("DB_RETRY_ATTEMPTS")
		os.Unsetenv("DB_RETRY_DELAY")
	}()

	config := NewDatabaseConfig()
	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "5433", config.Port)
	assert.Equal(t, "test-user", config.Username)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.Database)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 2*time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 1*time.Hour, config.ConnMaxIdleTime)
	assert.Equal(t, 60*time.Second, config.QueryTimeout)
	assert.Equal(t, 20*time.Second, config.ConnectTimeout)
	assert.Equal(t, 5, config.RetryAttempts)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
}

func TestParseIntOrDefault(t *testing.T) {
	t.Run("WithValidValue", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")
		result := parseIntOrDefault("TEST_INT", 10)
		assert.Equal(t, 42, result)
	})

	t.Run("WithInvalidValue", func(t *testing.T) {
		os.Setenv("TEST_INT", "invalid")
		defer os.Unsetenv("TEST_INT")
		result := parseIntOrDefault("TEST_INT", 10)
		assert.Equal(t, 10, result) // Should return default
	})

	t.Run("WithoutValue", func(t *testing.T) {
		os.Unsetenv("TEST_INT")
		result := parseIntOrDefault("TEST_INT", 10)
		assert.Equal(t, 10, result)
	})
}

func TestParseDurationOrDefault(t *testing.T) {
	t.Run("WithValidValue", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "2h")
		defer os.Unsetenv("TEST_DURATION")
		result := parseDurationOrDefault("TEST_DURATION", "1h")
		assert.Equal(t, 2*time.Hour, result)
	})

	t.Run("WithInvalidValue", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "invalid")
		defer os.Unsetenv("TEST_DURATION")
		result := parseDurationOrDefault("TEST_DURATION", "1h")
		assert.Equal(t, 1*time.Hour, result) // Should return default parsed
	})

	t.Run("WithoutValue", func(t *testing.T) {
		os.Unsetenv("TEST_DURATION")
		result := parseDurationOrDefault("TEST_DURATION", "30m")
		assert.Equal(t, 30*time.Minute, result)
	})
}

// Note: Integration tests for ConnectDB, ExecuteWithTimeout, etc. require a running DB
// and are better suited for integration test suites or mocked DB tests.
// For unit tests, we can mock sql.DB if needed, but here we focus on config parsing.
