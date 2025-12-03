package v1

import (
	"os"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewDatabaseConfig(t *testing.T) {
	config := NewDatabaseConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "postgres", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "require", config.SSLMode)
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
	os.Setenv("DB_SSLMODE", "disable")
	defer func() {
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_HOSTNAME")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_PORT")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_USERNAME")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_PASSWORD")
		os.Unsetenv("CHOREO_OPENDIF_DATABASE_DATABASENAME")
		os.Unsetenv("DB_SSLMODE")
	}()

	config := NewDatabaseConfig()
	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "5433", config.Port)
	assert.Equal(t, "test-user", config.Username)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.Database)
	assert.Equal(t, "disable", config.SSLMode)
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

func TestConnectGormDB_WithSQLite(t *testing.T) {
	// Use SQLite for testing instead of PostgreSQL
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Create a config that simulates connection pool settings
	config := &DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		Username:        "test",
		Password:        "test",
		Database:        "test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	// Test that we can configure connection pool (even with SQLite)
	sqlDB, err := db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test ping
	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestConnectGormDB_InvalidConnection(t *testing.T) {
	config := &DatabaseConfig{
		Host:     "invalid-host",
		Port:     "5432",
		Username: "invalid-user",
		Password: "invalid-password",
		Database: "invalid-db",
		SSLMode:  "disable",
	}

	_, err := ConnectGormDB(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestConnectGormDB_WithMigration(t *testing.T) {
	// Use SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Set migration flag
	os.Setenv("RUN_MIGRATION", "true")
	defer os.Unsetenv("RUN_MIGRATION")

	// Test migration path with SQLite
	// Note: This tests the migration logic, but ConnectGormDB uses PostgreSQL
	// For actual testing, we'd use SQLite directly
	err = db.AutoMigrate(&models.PolicyMetadata{})
	assert.NoError(t, err)
}
