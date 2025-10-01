package models

import (
	"time"
)

// SchemaStatus represents the status of a schema version
type SchemaStatus string

const (
	SchemaStatusActive     SchemaStatus = "active"
	SchemaStatusInactive   SchemaStatus = "inactive"
	SchemaStatusDeprecated SchemaStatus = "deprecated"
)

// VersionChangeType represents the type of change in a schema version
type VersionChangeType string

const (
	VersionChangeTypeMajor VersionChangeType = "major"
	VersionChangeTypeMinor VersionChangeType = "minor"
	VersionChangeTypePatch VersionChangeType = "patch"
)

// UnifiedSchema represents a schema version in the database
type UnifiedSchema struct {
	ID                int               `json:"id" db:"id"`
	Version           string            `json:"version" db:"version"`
	SDL               string            `json:"sdl" db:"sdl"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	CreatedBy         string            `json:"created_by" db:"created_by"`
	Status            SchemaStatus      `json:"status" db:"status"`
	ChangeType        VersionChangeType `json:"change_type" db:"change_type"`
	Notes             *string           `json:"notes,omitempty" db:"notes"`
	PreviousVersionID *int              `json:"previous_version_id,omitempty" db:"previous_version_id"`
}

// CreateSchemaRequest represents the request to create a new schema version
type CreateSchemaRequest struct {
	Version    string            `json:"version" validate:"required"`
	SDL        string            `json:"sdl" validate:"required"`
	CreatedBy  string            `json:"createdBy" validate:"required"`
	ChangeType VersionChangeType `json:"changeType" validate:"required,oneof=major minor patch"`
	Notes      *string           `json:"notes,omitempty"`
}

// UpdateSchemaStatusRequest represents the request to update schema status
type UpdateSchemaStatusRequest struct {
	IsActive bool    `json:"is_active"`
	Reason   *string `json:"reason,omitempty"`
}

// SchemaVersionResponse represents a schema version in API responses
type SchemaVersionResponse struct {
	ID                int               `json:"id"`
	Version           string            `json:"version"`
	SDL               string            `json:"sdl"`
	CreatedAt         time.Time         `json:"created_at"`
	CreatedBy         string            `json:"created_by"`
	Status            SchemaStatus      `json:"status"`
	ChangeType        VersionChangeType `json:"change_type"`
	Notes             *string           `json:"notes,omitempty"`
	PreviousVersionID *int              `json:"previous_version_id,omitempty"`
}

// SchemaVersionsListResponse represents the response for listing schema versions
type SchemaVersionsListResponse struct {
	Versions []SchemaVersionResponse `json:"versions"`
	Total    int                     `json:"total"`
}

// SchemaCompatibilityCheck represents a compatibility check result
type SchemaCompatibilityCheck struct {
	IsCompatible bool     `json:"is_compatible"`
	Issues       []string `json:"issues,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// UpdateSchemaRequest represents the request to update schema
type UpdateSchemaRequest struct {
	Version   string `json:"version" validate:"required,version"`
	SDL       string `json:"sdl" validate:"required"`
	CreatedAt string `json:"createdAt,omitempty"`
	CreatedBy string `json:"createdBy" validate:"required"`
}

// UpdateSchemaResponse represents the response from schema update
type UpdateSchemaResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Version    string `json:"version"`
	SchemaType string `json:"schema_type"`
	IsActive   bool   `json:"is_active"`
	UpdatedAt  string `json:"updated_at"`
}

// SchemaVersion represents a schema version
type SchemaVersion struct {
	ID                 int                    `json:"id"`
	Version            string                 `json:"version"`
	SDL                string                 `json:"sdl"`
	CreatedAt          time.Time              `json:"created_at"`
	CreatedBy          string                 `json:"created_by"`
	IsActive           bool                   `json:"is_active"`
	SchemaType         string                 `json:"schema_type"`
	CompatibilityLevel string                 `json:"compatibility_level"`
	PreviousVersion    string                 `json:"previous_version,omitempty"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// VersionCompatibility represents version compatibility info
type VersionCompatibility struct {
	IsCompatible    bool     `json:"is_compatible"`
	ChangeType      string   `json:"change_type"` // 'major', 'minor', 'patch'
	BreakingChanges []string `json:"breaking_changes,omitempty"`
	NewFields       []string `json:"new_fields,omitempty"`
	RemovedFields   []string `json:"removed_fields,omitempty"`
	ModifiedFields  []string `json:"modified_fields,omitempty"`
}

// SchemaService interface defines the contract for schema management
type SchemaService interface {
	// CreateSchema creates a new schema version
	CreateSchema(req *CreateSchemaRequest) (*UnifiedSchema, error)

	// GetSchemaVersion retrieves a specific schema version by version string
	GetSchemaVersion(version string) (*UnifiedSchema, error)

	// GetAllSchemaVersions retrieves all schema versions with optional filtering
	GetAllSchemaVersions(status *SchemaStatus, limit, offset int) ([]*UnifiedSchema, int, error)

	// UpdateSchemaStatus updates the status of a schema version
	UpdateSchemaStatus(version string, isActive bool, reason *string) error

	// GetActiveSchema retrieves the currently active schema
	GetActiveSchema() (*UnifiedSchema, error)

	// CheckCompatibility checks if a new schema is compatible with existing ones
	CheckCompatibility(newSDL string, previousVersionID *int) (*SchemaCompatibilityCheck, error)

	// GetPreviousVersionID retrieves the ID of the previous version for a given version
	GetPreviousVersionID(version string) (*int, error)
}
