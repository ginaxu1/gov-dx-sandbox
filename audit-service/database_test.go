package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.DatabasePath)
	assert.Equal(t, "./data/audit.db", config.DatabasePath) // Default path
	assert.Equal(t, 1, config.MaxOpenConns)                 // SQLite default: 1 to prevent lock contention
	assert.Equal(t, 1, config.MaxIdleConns)                 // SQLite default: 1 (matches MaxOpenConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)      // Default: 1 hour
	assert.Equal(t, 15*time.Minute, config.ConnMaxIdleTime) // Default: 15 minutes
}

func TestNewDatabaseConfig_WithEnvVars(t *testing.T) {
	testPath := "/tmp/test-audit.db"
	os.Setenv("DB_PATH", testPath)
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	defer func() {
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
	}()

	config := NewDatabaseConfig()
	assert.Equal(t, testPath, config.DatabasePath)
	assert.Equal(t, 50, config.MaxOpenConns)
	assert.Equal(t, 10, config.MaxIdleConns)
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

		result := parseDurationOrDefault(key, time.Hour)
		assert.Equal(t, 2*time.Hour, result)
	})

	t.Run("Returns default when not set", func(t *testing.T) {
		key := "TEST_DURATION_VAR_NONEXISTENT_12345"
		os.Unsetenv(key)

		result := parseDurationOrDefault(key, 30*time.Minute)
		assert.Equal(t, 30*time.Minute, result)
	})

	t.Run("Returns default when invalid", func(t *testing.T) {
		key := "TEST_DURATION_VAR_INVALID_12345"
		os.Setenv(key, "not-a-duration")
		defer os.Unsetenv(key)

		result := parseDurationOrDefault(key, 15*time.Minute)
		assert.Equal(t, 15*time.Minute, result)
	})

	t.Run("Handles various valid duration formats", func(t *testing.T) {
		testCases := []struct {
			value    string
			expected time.Duration
		}{
			{"30m", 30 * time.Minute},
			{"1h30m", 90 * time.Minute},
			{"15s", 15 * time.Second},
			{"2h", 2 * time.Hour},
		}

		for _, tc := range testCases {
			t.Run(tc.value, func(t *testing.T) {
				key := "TEST_DURATION_VAR_FORMAT_12345"
				os.Setenv(key, tc.value)
				defer os.Unsetenv(key)

				result := parseDurationOrDefault(key, time.Hour)
				assert.Equal(t, tc.expected, result)
			})
		}
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

func TestConnectGORM_Success(t *testing.T) {
	// Use a temporary database file for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &DatabaseConfig{
		DatabasePath:    dbPath,
		MaxOpenConns:    1, // SQLite best practice: use 1 to prevent lock contention
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	db, err := ConnectGORM(config)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify we can ping the database
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
}
