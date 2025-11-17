package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/telemetry"
	_ "github.com/lib/pq"
)

// SchemaDB handles database operations for schemas
type SchemaDB struct {
	db *sql.DB
}

// NewSchemaDB creates a new schema database connection
func NewSchemaDB(connectionString string) (*SchemaDB, error) {
	start := time.Now()
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		telemetry.RecordExternalCall(context.Background(), "postgres", "connect", time.Since(start), err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if pingErr := db.Ping(); pingErr != nil {
		telemetry.RecordExternalCall(context.Background(), "postgres", "connect", time.Since(start), pingErr)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	schemaDB := &SchemaDB{db: db}

	// Create tables if they don't exist
	if err := schemaDB.createTables(); err != nil {
		telemetry.RecordExternalCall(context.Background(), "postgres", "connect", time.Since(start), err)
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	telemetry.RecordExternalCall(context.Background(), "postgres", "connect", time.Since(start), nil)
	return schemaDB, nil
}

// Close closes the database connection
func (s *SchemaDB) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables
func (s *SchemaDB) createTables() error {
	start := time.Now()
	var err error
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "createTables", time.Since(start), err)
	}()
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

	if _, execErr := s.db.Exec(createSchemasTable); execErr != nil {
		err = fmt.Errorf("failed to create unified_schemas table: %w", execErr)
		return err
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

	if _, execErr := s.db.Exec(createVersionsTable); execErr != nil {
		err = fmt.Errorf("failed to create schema_versions table: %w", execErr)
		return err
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
func (s *SchemaDB) CreateSchema(schema *Schema) (err error) {
	start := time.Now()
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "CreateSchema", time.Since(start), err)
	}()
	query := `
		INSERT INTO unified_schemas (id, version, sdl, status, description, created_by, checksum, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, execErr := s.db.Exec(query, schema.ID, schema.Version, schema.SDL, schema.Status,
		schema.Description, schema.CreatedBy, schema.Checksum, schema.IsActive)

	if execErr != nil {
		err = fmt.Errorf("failed to create schema: %w", execErr)
		return err
	}

	return nil
}

// GetSchemaByVersion retrieves a schema by version
func (s *SchemaDB) GetSchemaByVersion(version string) (_ *Schema, err error) {
	start := time.Now()
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "GetSchemaByVersion", time.Since(start), err)
	}()
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas WHERE version = $1`

	row := s.db.QueryRow(query, version)

	schema := &Schema{}
	scanErr := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
		&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
		&schema.Checksum, &schema.IsActive)

	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			err = fmt.Errorf("schema version %s not found", version)
			return nil, err
		}
		err = fmt.Errorf("failed to get schema: %w", scanErr)
		return nil, err
	}

	return schema, nil
}

// GetActiveSchema retrieves the currently active schema
func (s *SchemaDB) GetActiveSchema() (_ *Schema, err error) {
	start := time.Now()
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "GetActiveSchema", time.Since(start), err)
	}()
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas WHERE is_active = TRUE LIMIT 1`

	row := s.db.QueryRow(query)

	schema := &Schema{}
	scanErr := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
		&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
		&schema.Checksum, &schema.IsActive)

	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, nil // No active schema
		}
		err = fmt.Errorf("failed to get active schema: %w", scanErr)
		return nil, err
	}

	return schema, nil
}

// GetAllSchemas retrieves all schemas
func (s *SchemaDB) GetAllSchemas() (_ []*Schema, err error) {
	start := time.Now()
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "GetAllSchemas", time.Since(start), err)
	}()
	query := `SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, is_active
			  FROM unified_schemas ORDER BY created_at DESC`

	rows, queryErr := s.db.Query(query)
	if queryErr != nil {
		err = fmt.Errorf("failed to get schemas: %w", queryErr)
		return nil, err
	}
	defer rows.Close()

	var schemas []*Schema
	for rows.Next() {
		schema := &Schema{}
		scanErr := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.Status,
			&schema.Description, &schema.CreatedAt, &schema.UpdatedAt, &schema.CreatedBy,
			&schema.Checksum, &schema.IsActive)
		if scanErr != nil {
			err = fmt.Errorf("failed to scan schema: %w", scanErr)
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// ActivateSchema activates a specific schema version
func (s *SchemaDB) ActivateSchema(version string) (err error) {
	start := time.Now()
	defer func() {
		telemetry.RecordExternalCall(context.Background(), "postgres", "ActivateSchema", time.Since(start), err)
	}()
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return err
	}
	defer tx.Rollback()

	// Deactivate all schemas
	_, err = tx.Exec("UPDATE unified_schemas SET is_active = FALSE")
	if err != nil {
		err = fmt.Errorf("failed to deactivate schemas: %w", err)
		return err
	}

	// Activate the specified version
	result, err := tx.Exec("UPDATE unified_schemas SET is_active = TRUE WHERE version = $1", version)
	if err != nil {
		err = fmt.Errorf("failed to activate schema: %w", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		err = fmt.Errorf("failed to get rows affected: %w", err)
		return err
	}

	if rowsAffected == 0 {
		err = fmt.Errorf("schema version %s not found", version)
		return err
	}

	// Commit transaction
	if commitErr := tx.Commit(); commitErr != nil {
		err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		return err
	}

	return nil
}
