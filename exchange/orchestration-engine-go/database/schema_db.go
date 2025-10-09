package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// SchemaDB handles database operations for schemas
type SchemaDB struct {
	db *sql.DB
}

// NewSchemaDB creates a new schema database connection
func NewSchemaDB(connectionString string) (*SchemaDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	schemaDB := &SchemaDB{db: db}

	// Create tables if they don't exist
	if err := schemaDB.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return schemaDB, nil
}

// Close closes the database connection
func (s *SchemaDB) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables
func (s *SchemaDB) createTables() error {
	// Create unified_schemas table
	createSchemasTable := `
	CREATE TABLE IF NOT EXISTS unified_schemas (
		id VARCHAR(36) PRIMARY KEY,
		version VARCHAR(50) UNIQUE NOT NULL,
		sdl TEXT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'inactive',
		description TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_by VARCHAR(100),
		checksum VARCHAR(64) NOT NULL,
		is_active BOOLEAN DEFAULT FALSE
	);`

	if _, err := s.db.Exec(createSchemasTable); err != nil {
		return fmt.Errorf("failed to create unified_schemas table: %w", err)
	}

	// Create schema_versions table for change tracking
	createVersionsTable := `
	CREATE TABLE IF NOT EXISTS schema_versions (
		id SERIAL PRIMARY KEY,
		from_version VARCHAR(50),
		to_version VARCHAR(50) NOT NULL,
		change_type VARCHAR(20) NOT NULL,
		changes JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_by VARCHAR(255) NOT NULL
	);`

	if _, err := s.db.Exec(createVersionsTable); err != nil {
		return fmt.Errorf("failed to create schema_versions table: %w", err)
	}

	return nil
}

// Schema represents a database schema record
type Schema struct {
	ID          string    `json:"id" db:"id"`
	Version     string    `json:"version" db:"version"`
	SDL         string    `json:"sdl" db:"sdl"`
	Status      string    `json:"status" db:"status"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	Checksum    string    `json:"checksum" db:"checksum"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// CreateSchema creates a new schema in the database
func (s *SchemaDB) CreateSchema(schema *Schema) error {
	query := `
		INSERT INTO unified_schemas (id, version, sdl, status, description, created_by, checksum, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(query, schema.ID, schema.Version, schema.SDL, schema.Status,
		schema.Description, schema.CreatedBy, schema.Checksum, schema.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetSchemaByVersion retrieves a schema by version
func (s *SchemaDB) GetSchemaByVersion(version string) (*Schema, error) {
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas WHERE version = $1`

	row := s.db.QueryRow(query, version)

	schema := &Schema{}
	err := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
		&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
		&schema.Checksum, &schema.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema version %s not found", version)
		}
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	return schema, nil
}

// GetActiveSchema retrieves the currently active schema
func (s *SchemaDB) GetActiveSchema() (*Schema, error) {
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas WHERE is_active = TRUE LIMIT 1`

	row := s.db.QueryRow(query)

	schema := &Schema{}
	err := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
		&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
		&schema.Checksum, &schema.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active schema
		}
		return nil, fmt.Errorf("failed to get active schema: %w", err)
	}

	return schema, nil
}

// GetAllSchemas retrieves all schemas
func (s *SchemaDB) GetAllSchemas() ([]*Schema, error) {
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*Schema
	for rows.Next() {
		schema := &Schema{}
		err := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
			&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
			&schema.Checksum, &schema.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// ActivateSchema activates a specific schema version
func (s *SchemaDB) ActivateSchema(version string) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate all schemas
	_, err = tx.Exec("UPDATE unified_schemas SET is_active = FALSE")
	if err != nil {
		return fmt.Errorf("failed to deactivate schemas: %w", err)
	}

	// Activate the specified version
	result, err := tx.Exec("UPDATE unified_schemas SET is_active = TRUE WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to activate schema: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schema version %s not found", version)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
