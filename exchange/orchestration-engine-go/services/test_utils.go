package services

import (
	"fmt"
	"os"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
)

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupPostgresTestDB creates a PostgreSQL test database connection
// Similar to portal-backend and consent-engine test utilities
func SetupPostgresTestDB(t *testing.T) *database.SchemaDB {
	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USERNAME")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_DATABASE")
	sslmode := os.Getenv("TEST_DB_SSLMODE")

	// Use safe defaults for non-sensitive values only
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	// Require sensitive credentials from environment - no defaults
	if user == "" {
		t.Skip("Skipping PostgreSQL test: TEST_DB_USERNAME environment variable not set")
		return nil
	}
	if password == "" {
		t.Skip("Skipping PostgreSQL test: TEST_DB_PASSWORD environment variable not set")
		return nil
	}
	if dbname == "" {
		t.Skip("Skipping PostgreSQL test: TEST_DB_DATABASE environment variable not set")
		return nil
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := database.NewSchemaDB(dsn)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to connect to database: %v", err)
		return nil
	}

	return db
}
