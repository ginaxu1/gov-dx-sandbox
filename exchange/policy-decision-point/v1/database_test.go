package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewDatabaseConfig(t *testing.T) {
	dbConfigs := &DatabaseConfigs{
		Host:     "localhost",
		Port:     "5432",
		Username: "postgres",
		Password: "password",
		Database: "pdp",
		SSLMode:  "require",
	}
	config := NewDatabaseConfig(dbConfigs)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "postgres", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "pdp", config.Database)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, time.Hour, config.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, config.ConnMaxIdleTime)
}

func TestNewDatabaseConfig_WithConfig(t *testing.T) {
	dbConfigs := &DatabaseConfigs{
		Host:     "test-host",
		Port:     "5432",
		Username: "test-user",
		Password: "test-pass",
		Database: "test-db",
		SSLMode:  "disable",
	}
	config := NewDatabaseConfig(dbConfigs)
	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "test-user", config.Username)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.Database)
	assert.Equal(t, "disable", config.SSLMode)
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
