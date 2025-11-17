package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *SchemaDB {
	// Use SQLite for testing instead of PostgreSQL
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to open test database")

	schemaDB := &SchemaDB{db: db}

	// Create tables
	createSchemasTable := `
	CREATE TABLE IF NOT EXISTS unified_schemas (
		id VARCHAR(36) PRIMARY KEY,
		version VARCHAR(50) UNIQUE NOT NULL,
		sdl TEXT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'inactive',
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(100),
		checksum VARCHAR(64) NOT NULL,
		is_active BOOLEAN DEFAULT 0
	);`

	_, err = db.Exec(createSchemasTable)
	require.NoError(t, err, "Failed to create schemas table")

	createVersionsTable := `
	CREATE TABLE IF NOT EXISTS schema_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_version VARCHAR(50),
		to_version VARCHAR(50) NOT NULL,
		change_type VARCHAR(20) NOT NULL,
		changes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(255) NOT NULL
	);`

	_, err = db.Exec(createVersionsTable)
	require.NoError(t, err, "Failed to create versions table")

	return schemaDB
}

func TestNewSchemaDB_InvalidConnection(t *testing.T) {
	// PostgreSQL connection will fail on ping, not on open
	_, err := NewSchemaDB("host=invalid port=5432 user=test password=test dbname=test sslmode=disable")
	assert.Error(t, err)
	// Error could be from open or ping
	assert.True(t, err != nil, "Should return error for invalid connection")
}

func TestSchemaDB_CreateSchema(t *testing.T) {
	db := setupTestDB(t)
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
	defer db.Close()

	// Initially empty
	schemas, err := db.GetAllSchemas()
	assert.NoError(t, err)
	assert.Len(t, schemas, 0)

	// Create multiple schemas
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

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

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

	// Get all schemas
	schemas, err = db.GetAllSchemas()
	assert.NoError(t, err)
	assert.Len(t, schemas, 2)
}

func TestSchemaDB_ActivateSchema(t *testing.T) {
	db := setupTestDB(t)
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
	defer db.Close()

	err := db.ActivateSchema("non-existent-version")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSchemaDB_Close(t *testing.T) {
	db := setupTestDB(t)
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
