package models

import (
	"time"
)

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
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message   string                 `json:"message"`
	Locations []GraphQLErrorLocation `json:"locations,omitempty"`
	Path      []interface{}          `json:"path,omitempty"`
}

// GraphQLErrorLocation represents the location of a GraphQL error
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}
