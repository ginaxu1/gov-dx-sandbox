package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseConfig(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("DB_CONN_MAX_IDLE_TIME")
	}()

	t.Run("Default configuration (SQLite)", func(t *testing.T) {
		// Ensure no env vars are set
		os.Unsetenv("DB_TYPE")

		config := NewDatabaseConfig()

		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, "./data/audit.db", config.DatabasePath)
		assert.Equal(t, 1, config.MaxOpenConns)
		assert.Equal(t, 1, config.MaxIdleConns)
		assert.Equal(t, time.Hour, config.ConnMaxLifetime)
		assert.Equal(t, 15*time.Minute, config.ConnMaxIdleTime)
	})

	t.Run("SQLite configuration from environment", func(t *testing.T) {
		os.Setenv("DB_TYPE", "sqlite")
		os.Setenv("DB_PATH", "./custom/audit.db")
		os.Setenv("DB_MAX_OPEN_CONNS", "10") // Should remain 1 for SQLite in our implementation, but let's see if we enforce it
		// In our implementation we do enforce 1 for defaults, but check if env var overrides
		// Actually the implementation uses parseIntOrDefault("DB_MAX_OPEN_CONNS", 1) so it takes the env var
		// Wait, the code says:
		// config.MaxOpenConns = parseIntOrDefault("DB_MAX_OPEN_CONNS", 1)
		// So if we set it to 10, it should be 10.
		// However, the comment says "SQLite best practice: Use MaxOpenConns=1"
		os.Setenv("DB_MAX_IDLE_CONNS", "5")

		config := NewDatabaseConfig()

		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, "./custom/audit.db", config.DatabasePath)
		// Verify if the code enforces 1 or allows override.
		// Based on reading the code: parseIntOrDefault("DB_MAX_OPEN_CONNS", 1)
		// It will return 10 if env var is 10.
		// Ideally we might want to enforce 1, but for now let's test what the code does.
		assert.Equal(t, 10, config.MaxOpenConns)
		assert.Equal(t, 5, config.MaxIdleConns)
	})

	t.Run("PostgreSQL configuration from environment", func(t *testing.T) {
		os.Setenv("DB_TYPE", "postgres")
		os.Setenv("DB_HOST", "db-host")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USERNAME", "user")
		os.Setenv("DB_PASSWORD", "pass")
		os.Setenv("DB_NAME", "audit")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("DB_MAX_OPEN_CONNS", "50")

		config := NewDatabaseConfig()

		assert.Equal(t, DatabaseTypePostgres, config.Type)
		assert.Equal(t, "db-host", config.Host)
		assert.Equal(t, "5432", config.Port)
		assert.Equal(t, "user", config.Username)
		assert.Equal(t, "pass", config.Password)
		assert.Equal(t, "audit", config.Database)
		assert.Equal(t, "require", config.SSLMode)
		assert.Equal(t, 50, config.MaxOpenConns)
	})

	t.Run("Unknown DB_TYPE defaults to SQLite", func(t *testing.T) {
		os.Setenv("DB_TYPE", "unknown_db")

		config := NewDatabaseConfig()

		assert.Equal(t, DatabaseTypeSQLite, config.Type)
	})
}

func TestConnectGormDB(t *testing.T) {
	// Create a temporary directory for test database
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("Connect to SQLite", func(t *testing.T) {
		config := &Config{
			Type:            DatabaseTypeSQLite,
			DatabasePath:    dbPath,
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 15 * time.Minute,
		}

		db, err := ConnectGormDB(config)
		assert.NoError(t, err)
		assert.NotNil(t, db)

		// Verify connection
		sqlDB, err := db.DB()
		assert.NoError(t, err)
		assert.NoError(t, sqlDB.Ping())
	})

	t.Run("Connect to In-Memory SQLite", func(t *testing.T) {
		config := &Config{
			Type:            DatabaseTypeSQLite,
			DatabasePath:    ":memory:",
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 15 * time.Minute,
		}

		db, err := ConnectGormDB(config)
		assert.NoError(t, err)
		assert.NotNil(t, db)

		// Verify connection
		sqlDB, err := db.DB()
		assert.NoError(t, err)
		assert.NoError(t, sqlDB.Ping())
	})
}
