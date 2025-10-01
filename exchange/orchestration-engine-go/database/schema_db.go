package database

import (
	"database/sql"
	"fmt"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	_ "github.com/lib/pq"
)

type SchemaDB struct {
	db *sql.DB
}

// NewSchemaDB creates a new schema database instance
func NewSchemaDB(db *sql.DB) *SchemaDB {
	return &SchemaDB{db: db}
}

// CreateSchemaTable creates the unified_schemas table if it doesn't exist
func (s *SchemaDB) CreateSchemaTable() error {
	query := `
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
			compatibility_level VARCHAR(10) DEFAULT 'major',
			previous_version VARCHAR(50),
			metadata JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT FALSE,
			schema_type VARCHAR(20) DEFAULT 'current'
		);
		
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_status ON unified_schemas(status);
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_version ON unified_schemas(version);
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_created_at ON unified_schemas(created_at);
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_active ON unified_schemas(is_active);
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_type ON unified_schemas(schema_type);
		CREATE INDEX IF NOT EXISTS idx_unified_schemas_compatibility ON unified_schemas(compatibility_level);
	`

	_, err := s.db.Exec(query)
	return err
}

// CreateSchemaVersionsTable creates the schema_versions table for change tracking
func (s *SchemaDB) CreateSchemaVersionsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_versions (
			id SERIAL PRIMARY KEY,
			from_version VARCHAR(50) NOT NULL,
			to_version VARCHAR(50) NOT NULL,
			change_type VARCHAR(20) NOT NULL,
			changes JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_schema_versions_from ON schema_versions(from_version);
		CREATE INDEX IF NOT EXISTS idx_schema_versions_to ON schema_versions(to_version);
		CREATE INDEX IF NOT EXISTS idx_schema_versions_change_type ON schema_versions(change_type);
		CREATE INDEX IF NOT EXISTS idx_schema_versions_created_at ON schema_versions(created_at);
	`

	_, err := s.db.Exec(query)
	return err
}

// CreateSchema inserts a new schema into the database
func (s *SchemaDB) CreateSchema(schema *models.UnifiedSchema) error {
	query := `
		INSERT INTO unified_schemas (id, version, sdl, status, description, created_at, updated_at, created_by, checksum, compatibility_level, previous_version, metadata, is_active, schema_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := s.db.Exec(query,
		schema.ID,
		schema.Version,
		schema.SDL,
		schema.Status,
		schema.Description,
		schema.CreatedAt,
		schema.UpdatedAt,
		schema.CreatedBy,
		schema.Checksum,
		schema.CompatibilityLevel,
		schema.PreviousVersion,
		schema.Metadata,
		schema.IsActive,
		schema.SchemaType)

	return err
}

// GetSchemaByVersion retrieves a schema by version
func (s *SchemaDB) GetSchemaByVersion(version string) (*models.UnifiedSchema, error) {
	query := `
		SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, compatibility_level, previous_version, metadata, is_active, schema_type
		FROM unified_schemas
		WHERE version = $1
	`

	row := s.db.QueryRow(query, version)
	schema := &models.UnifiedSchema{}

	err := row.Scan(
		&schema.ID,
		&schema.Version,
		&schema.SDL,
		&schema.Status,
		&schema.Description,
		&schema.CreatedAt,
		&schema.UpdatedAt,
		&schema.CreatedBy,
		&schema.Checksum,
		&schema.CompatibilityLevel,
		&schema.PreviousVersion,
		&schema.Metadata,
		&schema.IsActive,
		&schema.SchemaType,
	)

	if err != nil {
		return nil, err
	}

	return schema, nil
}

// GetActiveSchema retrieves the currently active schema
func (s *SchemaDB) GetActiveSchema() (*models.UnifiedSchema, error) {
	query := `
		SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, compatibility_level, previous_version, metadata, is_active, schema_type
		FROM unified_schemas
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := s.db.QueryRow(query)
	schema := &models.UnifiedSchema{}

	err := row.Scan(
		&schema.ID,
		&schema.Version,
		&schema.SDL,
		&schema.Status,
		&schema.Description,
		&schema.CreatedAt,
		&schema.UpdatedAt,
		&schema.CreatedBy,
		&schema.Checksum,
		&schema.CompatibilityLevel,
		&schema.PreviousVersion,
		&schema.Metadata,
		&schema.IsActive,
		&schema.SchemaType,
	)

	if err != nil {
		return nil, err
	}

	return schema, nil
}

// GetAllSchemas retrieves all schemas
func (s *SchemaDB) GetAllSchemas() ([]*models.UnifiedSchema, error) {
	query := `
		SELECT id, version, sdl, status, description, created_at, updated_at, created_by, checksum, compatibility_level, previous_version, metadata, is_active, schema_type
		FROM unified_schemas
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []*models.UnifiedSchema

	for rows.Next() {
		schema := &models.UnifiedSchema{}
		err := rows.Scan(
			&schema.ID,
			&schema.Version,
			&schema.SDL,
			&schema.Status,
			&schema.Description,
			&schema.CreatedAt,
			&schema.UpdatedAt,
			&schema.CreatedBy,
			&schema.Checksum,
			&schema.CompatibilityLevel,
			&schema.PreviousVersion,
			&schema.Metadata,
			&schema.IsActive,
			&schema.SchemaType,
		)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// UpdateSchemaStatus updates the status of a schema
func (s *SchemaDB) UpdateSchemaStatus(version string, status string) error {
	query := `
		UPDATE unified_schemas 
		SET status = $1, updated_at = NOW()
		WHERE version = $2
	`

	result, err := s.db.Exec(query, status, version)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schema with version %s not found", version)
	}

	return nil
}

// DeleteSchema deletes a schema by version
func (s *SchemaDB) DeleteSchema(version string) error {
	query := `DELETE FROM unified_schemas WHERE version = $1`

	result, err := s.db.Exec(query, version)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schema with version %s not found", version)
	}

	return nil
}

// GetSchemaVersions retrieves all schema versions with basic info
func (s *SchemaDB) GetSchemaVersions() ([]*models.SchemaVersionInfo, error) {
	query := `
		SELECT version, status, created_at, description, checksum
		FROM unified_schemas
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*models.SchemaVersionInfo

	for rows.Next() {
		version := &models.SchemaVersionInfo{}
		err := rows.Scan(
			&version.Version,
			&version.Status,
			&version.CreatedAt,
			&version.Description,
			&version.Checksum,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// DeactivateAllSchemas deactivates all schemas (used when activating a new one)
func (s *SchemaDB) DeactivateAllSchemas() error {
	query := `UPDATE unified_schemas SET status = 'inactive', updated_at = NOW()`
	_, err := s.db.Exec(query)
	return err
}

// CreateSchemaVersion creates a new schema version record for change tracking
func (s *SchemaDB) CreateSchemaVersion(version *models.SchemaVersion) error {
	query := `
		INSERT INTO schema_versions (from_version, to_version, change_type, changes, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(query,
		version.FromVersion,
		version.ToVersion,
		version.ChangeType,
		version.Changes,
		version.CreatedBy)

	return err
}

// GetSchemaVersionsByVersion retrieves schema version records for a specific version
func (s *SchemaDB) GetSchemaVersionsByVersion(version string) ([]*models.SchemaVersion, error) {
	query := `
		SELECT id, from_version, to_version, change_type, changes, created_at, created_by
		FROM schema_versions
		WHERE from_version = $1 OR to_version = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*models.SchemaVersion

	for rows.Next() {
		version := &models.SchemaVersion{}
		err := rows.Scan(
			&version.ID,
			&version.FromVersion,
			&version.ToVersion,
			&version.ChangeType,
			&version.Changes,
			&version.CreatedAt,
			&version.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetAllSchemaVersions retrieves all schema version records
func (s *SchemaDB) GetAllSchemaVersions() ([]*models.SchemaVersion, error) {
	query := `
		SELECT id, from_version, to_version, change_type, changes, created_at, created_by
		FROM schema_versions
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*models.SchemaVersion

	for rows.Next() {
		version := &models.SchemaVersion{}
		err := rows.Scan(
			&version.ID,
			&version.FromVersion,
			&version.ToVersion,
			&version.ChangeType,
			&version.Changes,
			&version.CreatedAt,
			&version.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}
