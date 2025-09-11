package models

import "time"

// ProviderType represents the type of a data provider
type ProviderType string

const (
	ProviderTypeGovernment ProviderType = "government"
	ProviderTypeBoard      ProviderType = "board"
	ProviderTypeBusiness   ProviderType = "business"
)

// ProviderSubmissionStatus represents the approval status of a provider's registration
type ProviderSubmissionStatus string

const (
	SubmissionStatusPending  ProviderSubmissionStatus = "pending"
	SubmissionStatusApproved ProviderSubmissionStatus = "approved"
	SubmissionStatusRejected ProviderSubmissionStatus = "rejected"
)

// ProviderSchemaStatus represents the status of a data provider's schema submission
type ProviderSchemaStatus string

const (
	SchemaStatusDraft    ProviderSchemaStatus = "draft"
	SchemaStatusPending  ProviderSchemaStatus = "pending"
	SchemaStatusApproved ProviderSchemaStatus = "approved"
	SchemaStatusRejected ProviderSchemaStatus = "rejected"
)

// ProviderSubmission represents a temporary submission from a potential new provider
type ProviderSubmission struct {
	SubmissionID string                   `json:"submissionId"`
	ProviderName string                   `json:"providerName"`
	ContactEmail string                   `json:"contactEmail"`
	PhoneNumber  string                   `json:"phoneNumber"`
	ProviderType ProviderType             `json:"providerType"`
	Status       ProviderSubmissionStatus `json:"status"`
	CreatedAt    time.Time                `json:"createdAt"`
}

// ProviderProfile represents the official, approved profile of a Data Provider
type ProviderProfile struct {
	ProviderID   string       `json:"providerId"`
	ProviderName string       `json:"providerName"`
	ContactEmail string       `json:"contactEmail"`
	PhoneNumber  string       `json:"phoneNumber"`
	ProviderType ProviderType `json:"providerType"`
	ApprovedAt   time.Time    `json:"approvedAt"`
}

// FieldConfiguration defines the metadata for a single field in a provider's schema
type FieldConfiguration struct {
	Source      string `json:"source"` // 'authoritative' | 'fallback' | 'other'
	IsOwner     bool   `json:"isOwner"`
	Description string `json:"description"`
}

// FieldConfigurations represents the nested structure of field configurations, grouped by GraphQL Type
type FieldConfigurations map[string]map[string]FieldConfiguration

// ProviderSchema represents a data provider's complete schema submission
type ProviderSchema struct {
	SubmissionID        string               `json:"submissionId"`
	ProviderID          string               `json:"providerId"`
	SchemaID            *string              `json:"schemaId,omitempty"` // Only set when status is approved
	Status              ProviderSchemaStatus `json:"status"`
	SchemaInput         *SchemaInput         `json:"schemaInput,omitempty"`
	FieldConfigurations FieldConfigurations  `json:"fieldConfigurations"`
	SDL                 string               `json:"sdl,omitempty"` // Store SDL directly
}

// SchemaInput represents the original schema source
type SchemaInput struct {
	Type  string `json:"type"` // 'endpoint' | 'json' | 'sdl'
	Value string `json:"value"`
}

// CreateProviderSubmissionRequest represents the request to create a new provider submission
type CreateProviderSubmissionRequest struct {
	ProviderName string       `json:"providerName"`
	ContactEmail string       `json:"contactEmail"`
	PhoneNumber  string       `json:"phoneNumber"`
	ProviderType ProviderType `json:"providerType"`
}

// CreateProviderSchemaRequest represents the request to create a new provider schema
type CreateProviderSchemaRequest struct {
	ProviderID          string              `json:"providerId"`
	SchemaInput         *SchemaInput        `json:"schemaInput,omitempty"`
	FieldConfigurations FieldConfigurations `json:"fieldConfigurations"`
}

// CreateProviderSchemaSDLRequest represents the request to create a provider schema with SDL
type CreateProviderSchemaSDLRequest struct {
	SDL string `json:"sdl" validate:"required"`
}

// CreateProviderSchemaSubmissionRequest represents the request to create a new schema submission or modify an existing one
type CreateProviderSchemaSubmissionRequest struct {
	SDL      string  `json:"sdl" validate:"required"`
	SchemaID *string `json:"schema_id,omitempty"` // Optional: if provided, this is a modification of existing schema
}

// UpdateProviderSubmissionRequest represents the request to update a provider submission
type UpdateProviderSubmissionRequest struct {
	Status *ProviderSubmissionStatus `json:"status,omitempty"`
}

// UpdateProviderSchemaRequest represents the request to update a provider schema
type UpdateProviderSchemaRequest struct {
	Status              *ProviderSchemaStatus `json:"status,omitempty"`
	FieldConfigurations FieldConfigurations   `json:"fieldConfigurations,omitempty"`
}

// UpdateProviderSubmissionResponse represents the response when updating a provider submission
type UpdateProviderSubmissionResponse struct {
	*ProviderSubmission
	ProviderID string `json:"providerId,omitempty"` // Only present when status is approved
}
