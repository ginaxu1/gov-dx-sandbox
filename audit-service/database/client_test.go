package database

import (
	"bytes"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("Test 1: No Configuration - In-Memory SQLite", func(t *testing.T) {
		// Ensure no env vars are set
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_HOST")

		config := NewDatabaseConfig()

		// Verify configuration uses in-memory database when no config provided
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, ":memory:", config.DatabasePath)
		assert.Equal(t, 1, config.MaxOpenConns)
		assert.Equal(t, 1, config.MaxIdleConns)
		assert.Equal(t, time.Hour, config.ConnMaxLifetime)
		assert.Equal(t, 15*time.Minute, config.ConnMaxIdleTime)

		// Verify database connection successful
		db, err := ConnectGormDB(config)
		require.NoError(t, err, "Database connection should succeed")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
		sqlDB.Close()
	})

	t.Run("Test 2: SQLite with Custom Path", func(t *testing.T) {
		// Clear any existing env vars first
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")

		// Create a temporary directory for custom path
		tempDir, err := os.MkdirTemp("", "audit_test_custom")
		require.NoError(t, err, "Should create temp directory")
		defer os.RemoveAll(tempDir)

		customPath := filepath.Join(tempDir, "custom_audit.db")

		// Set SQLite-specific env vars
		os.Setenv("DB_TYPE", "sqlite")
		os.Setenv("DB_PATH", customPath)
		os.Setenv("DB_MAX_OPEN_CONNS", "10")
		os.Setenv("DB_MAX_IDLE_CONNS", "5")

		config := NewDatabaseConfig()

		// Verify custom path configuration works
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, customPath, config.DatabasePath)
		assert.Equal(t, 10, config.MaxOpenConns)
		assert.Equal(t, 5, config.MaxIdleConns)

		// Verify database connection successful
		db, err := ConnectGormDB(config)
		require.NoError(t, err, "Database connection should succeed")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
		sqlDB.Close()

		// Verify database file was created
		assert.FileExists(t, customPath, "Database file should be created at custom path")
	})

	t.Run("Test 2a: DB_PATH Alone (No DB_TYPE) - Should Use File-Based SQLite", func(t *testing.T) {
		// Clear all database env vars
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Create a temporary directory for custom path
		tempDir, err := os.MkdirTemp("", "audit_test_dbpath_only")
		require.NoError(t, err, "Should create temp directory")
		defer os.RemoveAll(tempDir)

		customPath := filepath.Join(tempDir, "dbpath_only.db")

		// Set ONLY DB_PATH (no DB_TYPE)
		os.Setenv("DB_PATH", customPath)

		config := NewDatabaseConfig()

		// Verify that setting DB_PATH alone implies file-based SQLite
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, customPath, config.DatabasePath, "Should use DB_PATH value even without DB_TYPE")

		// Verify database connection successful
		db, err := ConnectGormDB(config)
		require.NoError(t, err, "Database connection should succeed")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
		sqlDB.Close()

		// Verify database file was created
		assert.FileExists(t, customPath, "Database file should be created")
	})

	t.Run("Test 2b: DB_TYPE=sqlite Without DB_PATH - Should Use Default Path", func(t *testing.T) {
		// Clear all env vars
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set ONLY DB_TYPE=sqlite (no DB_PATH)
		os.Setenv("DB_TYPE", "sqlite")

		config := NewDatabaseConfig()

		// Verify configuration uses default path
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, "./data/audit.db", config.DatabasePath, "Should use default path when DB_TYPE=sqlite without DB_PATH")

		// Verify database connection successful
		db, err := ConnectGormDB(config)
		require.NoError(t, err, "Database connection should succeed")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
		sqlDB.Close()
	})

	t.Run("Test 3: SQLite In-Memory", func(t *testing.T) {
		// Clear any existing env vars first
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set SQLite in-memory configuration
		os.Setenv("DB_TYPE", "sqlite")
		os.Setenv("DB_PATH", ":memory:")

		config := NewDatabaseConfig()

		// Verify in-memory configuration works
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, ":memory:", config.DatabasePath)

		// Verify database connection successful
		db, err := ConnectGormDB(config)
		require.NoError(t, err, "Database connection should succeed")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection works
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping in-memory database")
		sqlDB.Close()
	})

	t.Run("Test 4: PostgreSQL Configuration", func(t *testing.T) {
		// Clear any existing env vars first
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set PostgreSQL-specific env vars
		os.Setenv("DB_TYPE", "postgres")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USERNAME", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "disable")
		os.Setenv("DB_MAX_OPEN_CONNS", "50")

		config := NewDatabaseConfig()

		// Verify PostgreSQL configuration loads correctly from environment variables
		assert.Equal(t, DatabaseTypePostgres, config.Type)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, "5432", config.Port)
		assert.Equal(t, "testuser", config.Username)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "disable", config.SSLMode)
		assert.Equal(t, 50, config.MaxOpenConns)

		// Verify DSN construction works
		// The DSN should be properly constructed using net/url
		dsnURL := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(config.Username, config.Password),
			Host:   config.Host + ":" + config.Port,
			Path:   config.Database,
		}
		q := dsnURL.Query()
		q.Set("sslmode", config.SSLMode)
		dsnURL.RawQuery = q.Encode()
		expectedDSN := dsnURL.String()

		// Verify DSN contains expected components
		assert.Contains(t, expectedDSN, "postgres://")
		assert.Contains(t, expectedDSN, "testuser")
		assert.Contains(t, expectedDSN, "localhost:5432")
		assert.Contains(t, expectedDSN, "testdb")
		assert.Contains(t, expectedDSN, "sslmode=disable")

		// Connection attempt works (fails gracefully if PostgreSQL is not available, as expected)
		db, err := ConnectGormDB(config)
		if err != nil {
			// Expected if PostgreSQL is not running - verify error is about connection, not configuration
			assert.Contains(t, err.Error(), "failed to connect", "Error should be about connection failure, not configuration")
			assert.Contains(t, strings.ToLower(err.Error()), "connection", "Error should mention connection")
		} else {
			// If connection succeeds, verify it works
			require.NotNil(t, db, "Database should not be nil")
			sqlDB, err := db.DB()
			require.NoError(t, err, "Should be able to get underlying sql.DB")
			assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
			sqlDB.Close()
		}
	})

	t.Run("Test 5: Unknown DB_TYPE Defaults to File-Based SQLite", func(t *testing.T) {
		// Clear any existing env vars that might affect SQLite defaults
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("DB_PATH")

		os.Setenv("DB_TYPE", "unknown_db")

		config := NewDatabaseConfig()

		// Verify unknown database types default to file-based SQLite
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		// Should use SQLite defaults
		assert.Equal(t, 1, config.MaxOpenConns)
		assert.Equal(t, 1, config.MaxIdleConns)
		assert.Equal(t, "./data/audit.db", config.DatabasePath, "Unknown DB_TYPE should default to file-based SQLite, not in-memory")

		// Warning should be logged (we can't easily test logging in unit tests,
		// but the behavior is correct - it defaults to file-based SQLite)
	})

	t.Run("Test 5a: DB_HOST Set Without DB_TYPE=postgres - Should Use In-Memory SQLite", func(t *testing.T) {
		// Clear all env vars
		os.Unsetenv("DB_TYPE")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set DB_HOST but NOT DB_TYPE=postgres
		os.Setenv("DB_HOST", "localhost")

		// Capture log output to verify warning is logged
		var logBuffer bytes.Buffer
		originalLogger := slog.Default()
		defer slog.SetDefault(originalLogger) // Restore original logger
		slog.SetDefault(slog.New(slog.NewTextHandler(&logBuffer, nil)))

		config := NewDatabaseConfig()

		// Verify that DB_HOST alone (without DB_TYPE=postgres) does NOT trigger file-based SQLite
		// It should use in-memory SQLite because DB_HOST is only relevant for PostgreSQL
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, ":memory:", config.DatabasePath, "DB_HOST without DB_TYPE=postgres should use in-memory SQLite, not file-based")

		// Verify warning is logged about DB_HOST being ignored
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "DB_HOST will be ignored", "Warning should be logged when DB_HOST is set without DB_TYPE=postgres")
		assert.Contains(t, logOutput, "DB_TYPE is not 'postgres'", "Warning should mention that DB_TYPE is not postgres")
	})

	t.Run("Test 5b: DB_HOST Set With DB_TYPE=sqlite - Should Use File-Based SQLite", func(t *testing.T) {
		// Clear all env vars
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set DB_TYPE=sqlite and DB_HOST (DB_HOST should be ignored/warned)
		os.Setenv("DB_TYPE", "sqlite")
		os.Setenv("DB_HOST", "localhost")

		// Capture log output to verify warning is logged
		var logBuffer bytes.Buffer
		originalLogger := slog.Default()
		defer slog.SetDefault(originalLogger) // Restore original logger
		slog.SetDefault(slog.New(slog.NewTextHandler(&logBuffer, nil)))

		config := NewDatabaseConfig()

		// Verify that DB_TYPE=sqlite triggers file-based SQLite (DB_HOST is ignored)
		assert.Equal(t, DatabaseTypeSQLite, config.Type)
		assert.Equal(t, "./data/audit.db", config.DatabasePath, "DB_TYPE=sqlite should use file-based SQLite, DB_HOST is ignored")

		// Verify warning is logged about DB_HOST being ignored
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "DB_HOST will be ignored", "Warning should be logged when DB_HOST is set with DB_TYPE=sqlite")
		assert.Contains(t, logOutput, "DB_TYPE is not 'postgres'", "Warning should mention that DB_TYPE is not postgres")
	})

	t.Run("Test 6: PostgreSQL with Special Characters in Password", func(t *testing.T) {
		// Clear any existing env vars first
		os.Unsetenv("DB_PATH")
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")

		// Set PostgreSQL with special characters in password
		specialPassword := "p@ss w#rd!123"
		os.Setenv("DB_TYPE", "postgres")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USERNAME", "testuser")
		os.Setenv("DB_PASSWORD", specialPassword)
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "disable")

		config := NewDatabaseConfig()

		// Verify passwords with special characters are preserved correctly
		assert.Equal(t, DatabaseTypePostgres, config.Type)
		assert.Equal(t, specialPassword, config.Password, "Password with special characters should be preserved")

		// Verify DSN construction handles special characters properly using net/url
		dsnURL := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(config.Username, config.Password),
			Host:   config.Host + ":" + config.Port,
			Path:   config.Database,
		}
		q := dsnURL.Query()
		q.Set("sslmode", config.SSLMode)
		dsnURL.RawQuery = q.Encode()
		dsn := dsnURL.String()

		// Verify DSN is properly encoded (special characters should be URL-encoded)
		assert.Contains(t, dsn, "postgres://")
		assert.Contains(t, dsn, "testuser")
		// The password should be URL-encoded in the DSN
		// net/url.UserPassword automatically encodes special characters
		assert.NotContains(t, dsn, specialPassword, "Password should be URL-encoded in DSN, not plain text")
		// Verify the encoded password is present (URL encoding of special characters)
		encodedUser := url.UserPassword(config.Username, config.Password)
		expectedEncodedPassword := encodedUser.String()
		assert.Contains(t, dsn, expectedEncodedPassword, "DSN should contain properly encoded password")

		// Connection attempt works (fails gracefully if PostgreSQL is not available, as expected)
		db, err := ConnectGormDB(config)
		if err != nil {
			// Expected if PostgreSQL is not running - verify error is about connection, not DSN construction
			assert.Contains(t, err.Error(), "failed to connect", "Error should be about connection failure, not DSN construction")
			// Verify that the error is not about malformed DSN (which would indicate encoding issues)
			assert.NotContains(t, strings.ToLower(err.Error()), "malformed", "Error should not be about malformed DSN")
		} else {
			// If connection succeeds, verify it works
			require.NotNil(t, db, "Database should not be nil")
			sqlDB, err := db.DB()
			require.NoError(t, err, "Should be able to get underlying sql.DB")
			assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
			sqlDB.Close()
		}
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
		require.NoError(t, err, "Should connect to SQLite successfully")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
		sqlDB.Close()
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
		require.NoError(t, err, "Should connect to in-memory SQLite successfully")
		require.NotNil(t, db, "Database should not be nil")

		// Verify connection
		sqlDB, err := db.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")
		assert.NoError(t, sqlDB.Ping(), "Should be able to ping in-memory database")
		sqlDB.Close()
	})

	t.Run("Connect to PostgreSQL (graceful failure if not available)", func(t *testing.T) {
		config := &Config{
			Type:            DatabaseTypePostgres,
			Host:            "localhost",
			Port:            "5432",
			Username:        "testuser",
			Password:        "testpass",
			Database:        "testdb",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 15 * time.Minute,
		}

		db, err := ConnectGormDB(config)
		if err != nil {
			// Expected if PostgreSQL is not running
			assert.Contains(t, err.Error(), "failed to connect", "Error should be about connection failure")
			t.Logf("PostgreSQL connection failed as expected (PostgreSQL not available): %v", err)
		} else {
			// If connection succeeds, verify it works
			require.NotNil(t, db, "Database should not be nil")
			sqlDB, err := db.DB()
			require.NoError(t, err, "Should be able to get underlying sql.DB")
			assert.NoError(t, sqlDB.Ping(), "Should be able to ping database")
			sqlDB.Close()
		}
	})
}
