package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestDB creates a PostgreSQL test database connection
// Similar to consent-engine and api-server-go test utilities
func setupTestDB(t *testing.T) *SchemaDB {
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

	db, err := NewSchemaDB(dsn)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: failed to connect to database: %v", err)
		return nil
	}

	return db
}

func TestNewSchemaDB_InvalidConnection(t *testing.T) {
	// PostgreSQL connection will fail on ping, not on open
	_, err := NewSchemaDB("host=invalid port=5432 user=test password=test dbname=test sslmode=disable")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestNewSchemaDB_InvalidConnectionString(t *testing.T) {
	// Invalid connection string format
	_, err := NewSchemaDB("invalid connection string")
	assert.Error(t, err)
}

func TestSchemaDB_CreateSchema(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	schema := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  false,
	}

	err := db.CreateSchema(schema)
	assert.NoError(t, err)

	// Verify schema was created
	retrieved, err := db.GetSchemaByVersion("1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, schema.ID, retrieved.ID)
	assert.Equal(t, schema.Version, retrieved.Version)
	assert.Equal(t, schema.SDL, retrieved.SDL)
}

func TestSchemaDB_CreateSchema_DuplicateVersion(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	schema1 := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  false,
	}

	err := db.CreateSchema(schema1)
	assert.NoError(t, err)

	// Try to create duplicate version
	schema2 := &Schema{
		ID:        "test-id-2",
		Version:   "1.0.0", // Same version
		SDL:       "type Query { test2: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "def456",
		IsActive:  false,
	}

	err = db.CreateSchema(schema2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create schema")
}

func TestSchemaDB_GetSchemaByVersion(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	schema := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  false,
	}

	err := db.CreateSchema(schema)
	require.NoError(t, err)

	// Get existing schema
	retrieved, err := db.GetSchemaByVersion("1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, schema.ID, retrieved.ID)
	assert.Equal(t, schema.Version, retrieved.Version)

	// Get non-existent schema
	_, err = db.GetSchemaByVersion("2.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSchemaDB_GetActiveSchema(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// No active schema initially
	active, err := db.GetActiveSchema()
	assert.NoError(t, err)
	assert.Nil(t, active)

	// Create inactive schema
	schema1 := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  false,
	}
	err = db.CreateSchema(schema1)
	require.NoError(t, err)

	// Still no active schema
	active, err = db.GetActiveSchema()
	assert.NoError(t, err)
	assert.Nil(t, active)

	// Create active schema
	schema2 := &Schema{
		ID:        "test-id-2",
		Version:   "2.0.0",
		SDL:       "type Query { test2: String }",
		Status:    "active",
		CreatedBy: "test-user",
		Checksum:  "def456",
		IsActive:  true,
	}
	err = db.CreateSchema(schema2)
	require.NoError(t, err)

	// Now should have active schema
	active, err = db.GetActiveSchema()
	assert.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, "2.0.0", active.Version)
	assert.True(t, active.IsActive)
}

func TestSchemaDB_GetAllSchemas(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// Initially empty
	schemas, err := db.GetAllSchemas()
	assert.NoError(t, err)
	assert.Len(t, schemas, 0)

	// Create multiple schemas with explicit timestamps to ensure they're different
	now := time.Now()
	schema1 := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = db.CreateSchema(schema1)
	require.NoError(t, err)

	schema2 := &Schema{
		ID:        "test-id-2",
		Version:   "2.0.0",
		SDL:       "type Query { test2: String }",
		Status:    "active",
		CreatedBy: "test-user",
		Checksum:  "def456",
		IsActive:  true,
		CreatedAt: now.Add(1 * time.Second),
		UpdatedAt: now.Add(1 * time.Second),
	}
	err = db.CreateSchema(schema2)
	require.NoError(t, err)

	// Get all schemas
	schemas, err = db.GetAllSchemas()
	assert.NoError(t, err)
	assert.Len(t, schemas, 2)
}

func TestSchemaDB_ActivateSchema(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// Create two schemas
	schema1 := &Schema{
		ID:        "test-id-1",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "abc123",
		IsActive:  true, // Initially active
	}
	err := db.CreateSchema(schema1)
	require.NoError(t, err)

	schema2 := &Schema{
		ID:        "test-id-2",
		Version:   "2.0.0",
		SDL:       "type Query { test2: String }",
		Status:    "inactive",
		CreatedBy: "test-user",
		Checksum:  "def456",
		IsActive:  false,
	}
	err = db.CreateSchema(schema2)
	require.NoError(t, err)

	// Activate version 2.0.0
	err = db.ActivateSchema("2.0.0")
	assert.NoError(t, err)

	// Verify schema2 is now active
	active, err := db.GetActiveSchema()
	assert.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, "2.0.0", active.Version)
	assert.True(t, active.IsActive)

	// Verify schema1 is now inactive
	schema1Retrieved, err := db.GetSchemaByVersion("1.0.0")
	assert.NoError(t, err)
	assert.False(t, schema1Retrieved.IsActive)
}

func TestSchemaDB_ActivateSchema_NotFound(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	err := db.ActivateSchema("non-existent-version")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSchemaDB_Close(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	err := db.Close()
	assert.NoError(t, err)

	// Try to use closed database
	err = db.CreateSchema(&Schema{
		ID:        "test-id",
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		CreatedBy: "test-user",
		Checksum:  "abc123",
	})
	assert.Error(t, err)
}
