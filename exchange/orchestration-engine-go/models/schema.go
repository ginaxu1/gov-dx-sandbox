package models

import (
	"time"
)

<<<<<<< HEAD
// UnifiedSchema represents a versioned GraphQL schema
type UnifiedSchema struct {
	ID                 string                 `json:"id" db:"id"`
	Version            string                 `json:"version" db:"version"`
	SDL                string                 `json:"sdl" db:"sdl"`
	Status             string                 `json:"status" db:"status"` // active, inactive, deprecated
	Description        string                 `json:"description" db:"description"`
	CreatedAt          time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time              `json:"updatedAt" db:"updated_at"`
	CreatedBy          string                 `json:"createdBy" db:"created_by"`
	Checksum           string                 `json:"checksum" db:"checksum"`
	CompatibilityLevel string                 `json:"compatibilityLevel" db:"compatibility_level"`
	PreviousVersion    *string                `json:"previousVersion,omitempty" db:"previous_version"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	IsActive           bool                   `json:"isActive" db:"is_active"`
	SchemaType         string                 `json:"schemaType" db:"schema_type"`
}

// CreateSchemaRequest represents a request to create a new schema version
type CreateSchemaRequest struct {
	SDL         string `json:"sdl" validate:"required"`
	Description string `json:"description"`
	Version     string `json:"version,omitempty"` // Optional, will be auto-generated if not provided
}

// UpdateSchemaStatusRequest represents a request to update schema status
type UpdateSchemaStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive deprecated"`
}

// SchemaCompatibilityCheck represents the result of a compatibility check
type SchemaCompatibilityCheck struct {
	Compatible         bool     `json:"compatible"`
	BreakingChanges    []string `json:"breakingChanges,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
	CompatibilityLevel string   `json:"compatibilityLevel"` // major, minor, patch
}

// SchemaVersionInfo represents information about a schema version
type SchemaVersionInfo struct {
	Version     string    `json:"version"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
	Checksum    string    `json:"checksum"`
}

// SchemaVersion represents a schema version change record
type SchemaVersion struct {
	ID          int                    `json:"id" db:"id"`
	FromVersion string                 `json:"fromVersion" db:"from_version"`
	ToVersion   string                 `json:"toVersion" db:"to_version"`
	ChangeType  string                 `json:"changeType" db:"change_type"` // major, minor, patch
	Changes     map[string]interface{} `json:"changes" db:"changes"`
	CreatedAt   time.Time              `json:"createdAt" db:"created_at"`
	CreatedBy   string                 `json:"createdBy" db:"created_by"`
}

// GraphQLRequest represents a GraphQL query request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
	SchemaVersion string                 `json:"schemaVersion,omitempty"` // Optional schema version
}

// GraphQLResponse represents a GraphQL query response
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []interface{}          `json:"errors,omitempty"`
=======
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
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
}

// SchemaService interface defines the contract for schema management
type SchemaService interface {
<<<<<<< HEAD
	// Schema CRUD operations
	CreateSchema(req *CreateSchemaRequest) (*UnifiedSchema, error)
	GetSchemaByVersion(version string) (*UnifiedSchema, error)
	GetActiveSchema() (*UnifiedSchema, error)
	GetAllSchemas() ([]*UnifiedSchema, error)
	UpdateSchemaStatus(version string, req *UpdateSchemaStatusRequest) error
	DeleteSchema(version string) error

	// Versioning operations
	ActivateVersion(version string) error
	DeactivateVersion(version string) error
	GetSchemaVersions() ([]*SchemaVersionInfo, error)

	// Compatibility operations
	CheckCompatibility(sdl string) (*SchemaCompatibilityCheck, error)
	ValidateSDL(sdl string) error

	// Query operations
	ExecuteQuery(req *GraphQLRequest) (*GraphQLResponse, error)
=======
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
<<<<<<< HEAD
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
=======

	// Multi-Version Schema Support Methods

	// LoadSchema loads schema from database into memory
	LoadSchema() error

	// GetSchemaForVersion returns schema for specific version
	GetSchemaForVersion(version string) (interface{}, error)

	// RouteQuery routes GraphQL query to appropriate schema version
	RouteQuery(query string, version string) (interface{}, error)

	// GetDefaultSchema returns the currently active default schema
	GetDefaultSchema() (interface{}, error)

	// ReloadSchemasInMemory reloads all schemas from database into memory
	ReloadSchemasInMemory() error

	// GetSchemaVersions returns all loaded schema versions
	GetSchemaVersions() map[string]interface{}

	// IsSchemaVersionLoaded checks if a specific schema version is loaded in memory
	IsSchemaVersionLoaded(version string) bool

	// GetActiveSchemaVersion returns the version string of the currently active schema
	GetActiveSchemaVersion() (string, error)
>>>>>>> ac208af (added multi-version schema support)
}
