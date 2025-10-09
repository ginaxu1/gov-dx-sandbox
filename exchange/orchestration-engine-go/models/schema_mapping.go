package models

import "time"

// UnifiedSchema represents a unified schema version
type UnifiedSchema struct {
	ID          string    `json:"id" db:"id"`
	Version     string    `json:"version" db:"version"`
	SDL         string    `json:"sdl" db:"sdl"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	Status      string    `json:"status" db:"status"` // draft, pending_approval, active, deprecated
}

// ProviderSchema represents a provider-specific schema
type ProviderSchema struct {
	ID         string    `json:"id" db:"id"`
	ProviderID string    `json:"provider_id" db:"provider_id"`
	SchemaName string    `json:"schema_name" db:"schema_name"`
	SDL        string    `json:"sdl" db:"sdl"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// FieldMapping represents a mapping between unified and provider fields
type FieldMapping struct {
	ID                string                 `json:"id" db:"id"`
	UnifiedSchemaID   string                 `json:"unified_schema_id" db:"unified_schema_id"`
	UnifiedFieldPath  string                 `json:"unified_field_path" db:"unified_field_path"`
	ProviderID        string                 `json:"provider_id" db:"provider_id"`
	ProviderFieldPath string                 `json:"provider_field_path" db:"provider_field_path"`
	FieldType         string                 `json:"field_type" db:"field_type"`
	IsRequired        bool                   `json:"is_required" db:"is_required"`
	Directives        map[string]interface{} `json:"directives" db:"directives"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
}

// SchemaChangeHistory represents changes made to schemas
type SchemaChangeHistory struct {
	ID                  string                 `json:"id" db:"id"`
	UnifiedSchemaID     string                 `json:"unified_schema_id" db:"unified_schema_id"`
	ChangeType          string                 `json:"change_type" db:"change_type"`
	UnifiedFieldPath    string                 `json:"unified_field_path" db:"unified_field_path"`
	ProviderFieldPath   string                 `json:"provider_field_path" db:"provider_field_path"`
	OldValue            map[string]interface{} `json:"old_value" db:"old_value"`
	NewValue            map[string]interface{} `json:"new_value" db:"new_value"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	CreatedBy           string                 `json:"created_by" db:"created_by"`
}

// API Request/Response models
type CreateUnifiedSchemaRequest struct {
	Version   string `json:"version" validate:"required"`
	SDL       string `json:"sdl" validate:"required"`
	Notes     string `json:"notes"`
	CreatedBy string `json:"createdBy" validate:"required"`
}

type CreateUnifiedSchemaResponse struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	SDL       string `json:"sdl"`
	IsActive  bool   `json:"is_active"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

type ActivateSchemaResponse struct {
	Message string `json:"message"`
}

type ProviderSchemaResponse struct {
	ProviderID string `json:"provider_id"`
	SchemaName string `json:"schema_name"`
	SDL        string `json:"sdl"`
	IsActive   bool   `json:"is_active"`
}

type FieldMappingRequest struct {
	UnifiedFieldPath  string                 `json:"unified_field_path" validate:"required"`
	ProviderID        string                 `json:"provider_id" validate:"required"`
	ProviderFieldPath string                 `json:"provider_field_path" validate:"required"`
	FieldType         string                 `json:"field_type" validate:"required"`
	IsRequired        bool                   `json:"is_required"`
	Directives        map[string]interface{} `json:"directives"`
}

type FieldMappingResponse struct {
	ID                string                 `json:"id"`
	UnifiedFieldPath  string                 `json:"unified_field_path"`
	ProviderID        string                 `json:"provider_id"`
	ProviderFieldPath string                 `json:"provider_field_path"`
	FieldType         string                 `json:"field_type"`
	IsRequired        bool                   `json:"is_required"`
	Directives        map[string]interface{} `json:"directives"`
	CreatedAt         string                 `json:"created_at"`
}

type CompatibilityResult struct {
	Compatible      bool     `json:"compatible"`
	BreakingChanges []string `json:"breaking_changes"`
	Warnings        []string `json:"warnings"`
}

type CompatibilityCheckRequest struct {
	OldVersion string `json:"old_version"`
	NewSDL     string `json:"new_sdl"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}
