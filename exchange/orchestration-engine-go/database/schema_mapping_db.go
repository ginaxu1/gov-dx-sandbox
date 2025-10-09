package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// SchemaMappingDB handles database operations for schema mapping
type SchemaMappingDB struct {
	db *sql.DB
}

// NewSchemaMappingDB creates a new schema mapping database connection
func NewSchemaMappingDB(connectionString string) (*SchemaMappingDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	schemaMappingDB := &SchemaMappingDB{db: db}

	// Run migration
	if err := schemaMappingDB.runMigration(); err != nil {
		return nil, fmt.Errorf("failed to run migration: %w", err)
	}

	return schemaMappingDB, nil
}

// Close closes the database connection
func (s *SchemaMappingDB) Close() error {
	return s.db.Close()
}

// runMigration runs the schema mapping migration
func (s *SchemaMappingDB) runMigration() error {
	migrationSQL := `
	-- Create unified_schemas table
	CREATE TABLE IF NOT EXISTS unified_schemas (
		id VARCHAR(36) PRIMARY KEY,
		version VARCHAR(20) NOT NULL UNIQUE,
		sdl TEXT NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_by VARCHAR(255) NOT NULL,
		status VARCHAR(20) DEFAULT 'draft'
	);

	-- Create provider_schemas table
	CREATE TABLE IF NOT EXISTS provider_schemas (
		id VARCHAR(36) PRIMARY KEY,
		provider_id VARCHAR(255) NOT NULL,
		schema_name VARCHAR(255) NOT NULL,
		sdl TEXT NOT NULL,
		is_active BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Create field_mappings table
	CREATE TABLE IF NOT EXISTS field_mappings (
		id VARCHAR(36) PRIMARY KEY,
		unified_schema_id VARCHAR(36) NOT NULL REFERENCES unified_schemas(id) ON DELETE CASCADE,
		unified_field_path VARCHAR(500) NOT NULL,
		provider_id VARCHAR(255) NOT NULL,
		provider_field_path VARCHAR(500) NOT NULL,
		field_type VARCHAR(50) NOT NULL,
		is_required BOOLEAN DEFAULT FALSE,
		directives JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Create schema_change_history table
	CREATE TABLE IF NOT EXISTS schema_change_history (
		id VARCHAR(36) PRIMARY KEY,
		unified_schema_id VARCHAR(36) NOT NULL REFERENCES unified_schemas(id) ON DELETE CASCADE,
		change_type VARCHAR(50) NOT NULL,
		unified_field_path VARCHAR(500),
		provider_field_path VARCHAR(500),
		old_value JSONB,
		new_value JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		created_by VARCHAR(255) NOT NULL
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_unified_schemas_version ON unified_schemas(version);
	CREATE INDEX IF NOT EXISTS idx_unified_schemas_is_active ON unified_schemas(is_active);
	CREATE INDEX IF NOT EXISTS idx_provider_schemas_provider_id ON provider_schemas(provider_id);
	CREATE INDEX IF NOT EXISTS idx_provider_schemas_is_active ON provider_schemas(is_active);
	CREATE INDEX IF NOT EXISTS idx_field_mappings_unified_schema_id ON field_mappings(unified_schema_id);
	CREATE INDEX IF NOT EXISTS idx_field_mappings_provider_id ON field_mappings(provider_id);
	CREATE INDEX IF NOT EXISTS idx_schema_change_history_unified_schema_id ON schema_change_history(unified_schema_id);

	-- Add constraints (only if they don't exist)
	DO $$ 
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_unified_schemas_status') THEN
			ALTER TABLE unified_schemas ADD CONSTRAINT chk_unified_schemas_status 
				CHECK (status IN ('draft', 'pending_approval', 'active', 'deprecated'));
		END IF;
	END $$;

	-- Ensure only one active unified schema at a time
	CREATE UNIQUE INDEX IF NOT EXISTS idx_unified_schemas_single_active 
		ON unified_schemas(is_active) WHERE is_active = TRUE;
	`

	_, err := s.db.Exec(migrationSQL)
	return err
}

// Unified Schema Operations

// CreateUnifiedSchema creates a new unified schema
func (s *SchemaMappingDB) CreateUnifiedSchema(schema *models.UnifiedSchema) error {
	query := `
		INSERT INTO unified_schemas (id, version, sdl, is_active, notes, created_by, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Exec(query, schema.ID, schema.Version, schema.SDL, schema.IsActive,
		schema.Notes, schema.CreatedBy, schema.Status)

	if err != nil {
		return fmt.Errorf("failed to create unified schema: %w", err)
	}

	return nil
}

// GetUnifiedSchemaByVersion retrieves a unified schema by version
func (s *SchemaMappingDB) GetUnifiedSchemaByVersion(version string) (*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, is_active, notes, created_at, created_by, status
			  FROM unified_schemas WHERE version = $1`

	row := s.db.QueryRow(query, version)

	schema := &models.UnifiedSchema{}
	err := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.IsActive,
		&schema.Notes, &schema.CreatedAt, &schema.CreatedBy, &schema.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unified schema version %s not found", version)
		}
		return nil, fmt.Errorf("failed to get unified schema: %w", err)
	}

	return schema, nil
}

// GetActiveUnifiedSchema retrieves the currently active unified schema
func (s *SchemaMappingDB) GetActiveUnifiedSchema() (*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, is_active, notes, created_at, created_by, status
			  FROM unified_schemas WHERE is_active = TRUE LIMIT 1`

	row := s.db.QueryRow(query)

	schema := &models.UnifiedSchema{}
	err := row.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.IsActive,
		&schema.Notes, &schema.CreatedAt, &schema.CreatedBy, &schema.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active schema
		}
		return nil, fmt.Errorf("failed to get active unified schema: %w", err)
	}

	return schema, nil
}

// GetAllUnifiedSchemas retrieves all unified schemas
func (s *SchemaMappingDB) GetAllUnifiedSchemas() ([]*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, is_active, notes, created_at, created_by, status
			  FROM unified_schemas ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unified schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*models.UnifiedSchema
	for rows.Next() {
		schema := &models.UnifiedSchema{}
		err := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.IsActive,
			&schema.Notes, &schema.CreatedAt, &schema.CreatedBy, &schema.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unified schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// ActivateUnifiedSchema activates a specific unified schema version
func (s *SchemaMappingDB) ActivateUnifiedSchema(version string) error {
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
	result, err := tx.Exec("UPDATE unified_schemas SET is_active = TRUE, status = 'active' WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to activate schema: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("unified schema version %s not found", version)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Provider Schema Operations

// CreateProviderSchema creates a new provider schema
func (s *SchemaMappingDB) CreateProviderSchema(schema *models.ProviderSchema) error {
	query := `
		INSERT INTO provider_schemas (id, provider_id, schema_name, sdl, is_active)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, schema.ID, schema.ProviderID, schema.SchemaName, schema.SDL, schema.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create provider schema: %w", err)
	}

	return nil
}

// GetAllProviderSchemas retrieves all active provider schemas
func (s *SchemaMappingDB) GetAllProviderSchemas() (map[string]*models.ProviderSchema, error) {
	query := `SELECT id, provider_id, schema_name, sdl, is_active, created_at
			  FROM provider_schemas WHERE is_active = TRUE ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider schemas: %w", err)
	}
	defer rows.Close()

	schemas := make(map[string]*models.ProviderSchema)
	for rows.Next() {
		schema := &models.ProviderSchema{}
		err := rows.Scan(&schema.ID, &schema.ProviderID, &schema.SchemaName, &schema.SDL, &schema.IsActive, &schema.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider schema: %w", err)
		}
		schemas[schema.ProviderID] = schema
	}

	return schemas, nil
}

// Field Mapping Operations

// CreateFieldMapping creates a new field mapping
func (s *SchemaMappingDB) CreateFieldMapping(mapping *models.FieldMapping) error {
	query := `
		INSERT INTO field_mappings (id, unified_schema_id, unified_field_path, provider_id, provider_field_path, field_type, is_required, directives)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// Convert directives to JSON string for PostgreSQL
	var directivesJSON string
	if mapping.Directives != nil {
		directivesBytes, err := json.Marshal(mapping.Directives)
		if err != nil {
			return fmt.Errorf("failed to marshal directives: %w", err)
		}
		directivesJSON = string(directivesBytes)
	}

	_, err := s.db.Exec(query, mapping.ID, mapping.UnifiedSchemaID, mapping.UnifiedFieldPath,
		mapping.ProviderID, mapping.ProviderFieldPath, mapping.FieldType, mapping.IsRequired, directivesJSON)

	if err != nil {
		return fmt.Errorf("failed to create field mapping: %w", err)
	}

	return nil
}

// GetFieldMappingsBySchemaID retrieves all field mappings for a unified schema
func (s *SchemaMappingDB) GetFieldMappingsBySchemaID(schemaID string) ([]*models.FieldMapping, error) {
	query := `SELECT id, unified_schema_id, unified_field_path, provider_id, provider_field_path, field_type, is_required, directives, created_at
			  FROM field_mappings WHERE unified_schema_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get field mappings: %w", err)
	}
	defer rows.Close()

	var mappings []*models.FieldMapping
	for rows.Next() {
		mapping := &models.FieldMapping{}
		var directivesJSON sql.NullString
		err := rows.Scan(&mapping.ID, &mapping.UnifiedSchemaID, &mapping.UnifiedFieldPath,
			&mapping.ProviderID, &mapping.ProviderFieldPath, &mapping.FieldType, &mapping.IsRequired, &directivesJSON, &mapping.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field mapping: %w", err)
		}

		// Parse directives JSON
		if directivesJSON.Valid && directivesJSON.String != "" {
			err := json.Unmarshal([]byte(directivesJSON.String), &mapping.Directives)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal directives: %w", err)
			}
		} else {
			mapping.Directives = make(map[string]interface{})
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// UpdateFieldMapping updates an existing field mapping
func (s *SchemaMappingDB) UpdateFieldMapping(mapping *models.FieldMapping) error {
	query := `
		UPDATE field_mappings 
		SET provider_id = $2, provider_field_path = $3, field_type = $4, is_required = $5, directives = $6
		WHERE id = $1`

	// Convert directives to JSON string for PostgreSQL
	var directivesJSON string
	if mapping.Directives != nil {
		directivesBytes, err := json.Marshal(mapping.Directives)
		if err != nil {
			return fmt.Errorf("failed to marshal directives: %w", err)
		}
		directivesJSON = string(directivesBytes)
	}

	result, err := s.db.Exec(query, mapping.ID, mapping.ProviderID, mapping.ProviderFieldPath,
		mapping.FieldType, mapping.IsRequired, directivesJSON)

	if err != nil {
		return fmt.Errorf("failed to update field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field mapping with id %s not found", mapping.ID)
	}

	return nil
}

// DeleteFieldMapping deletes a field mapping
func (s *SchemaMappingDB) DeleteFieldMapping(mappingID string) error {
	query := `DELETE FROM field_mappings WHERE id = $1`

	result, err := s.db.Exec(query, mappingID)
	if err != nil {
		return fmt.Errorf("failed to delete field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field mapping with id %s not found", mappingID)
	}

	return nil
}

// Schema Change History Operations

// CreateSchemaChangeHistory creates a new schema change history entry
func (s *SchemaMappingDB) CreateSchemaChangeHistory(change *models.SchemaChangeHistory) error {
	query := `
		INSERT INTO schema_change_history (id, unified_schema_id, change_type, unified_field_path, provider_field_path, old_value, new_value, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(query, change.ID, change.UnifiedSchemaID, change.ChangeType,
		change.UnifiedFieldPath, change.ProviderFieldPath, change.OldValue, change.NewValue, change.CreatedBy)

	if err != nil {
		return fmt.Errorf("failed to create schema change history: %w", err)
	}

	return nil
}

// GetSchemaChangeHistory retrieves change history for a unified schema
func (s *SchemaMappingDB) GetSchemaChangeHistory(schemaID string) ([]*models.SchemaChangeHistory, error) {
	query := `SELECT id, unified_schema_id, change_type, unified_field_path, provider_field_path, old_value, new_value, created_at, created_by
			  FROM schema_change_history WHERE unified_schema_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema change history: %w", err)
	}
	defer rows.Close()

	var changes []*models.SchemaChangeHistory
	for rows.Next() {
		change := &models.SchemaChangeHistory{}
		err := rows.Scan(&change.ID, &change.UnifiedSchemaID, &change.ChangeType,
			&change.UnifiedFieldPath, &change.ProviderFieldPath, &change.OldValue, &change.NewValue, &change.CreatedAt, &change.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema change history: %w", err)
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// Utility functions

// GenerateID generates a new UUID
func GenerateID() string {
	return uuid.New().String()
}
