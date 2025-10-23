package models

// PolicyMetadataCreateRequest represents the request to create policy metadata
type PolicyMetadataCreateRequest struct {
	FieldName         string           `json:"field_name" validate:"required"`
	DisplayName       string           `json:"display_name" validate:"required"`
	Description       string           `json:"description"`
	Source            string           `json:"source" validate:"required"`
	IsOwner           bool             `json:"is_owner"`
	AccessControlType string           `json:"access_control_type" validate:"required"`
	AllowList         []AllowListEntry `json:"allow_list"`
}

// PolicyMetadataCreateResponse represents the response from policy metadata creation
type PolicyMetadataCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
}

// AllowListUpdateRequest represents the request to update allow list
type AllowListUpdateRequest struct {
	FieldName     string `json:"field_name" validate:"required"`
	ApplicationID string `json:"application_id" validate:"required"`
	ExpiresAt     string `json:"expires_at" validate:"required"`
}

// AllowListUpdateResponse represents the response from allow list update
type AllowListUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// AllowListEntry represents an entry in the allow list
type AllowListEntry struct {
	ApplicationID string `json:"application_id"`
	ExpiresAt     int64  `json:"expires_at"`
}

// PolicyMetadata represents a policy metadata record from the database
type PolicyMetadata struct {
	ID                string           `json:"id"`
	SchemaID          *string          `json:"schema_id,omitempty"`
	FieldName         string           `json:"field_name"`
	DisplayName       *string          `json:"display_name,omitempty"`
	Description       *string          `json:"description,omitempty"`
	Source            string           `json:"source"`
	IsOwner           bool             `json:"is_owner"`
	Owner             string           `json:"owner"`
	AccessControlType string           `json:"access_control_type"`
	AllowList         []AllowListEntry `json:"allow_list"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
}

// PolicyDecisionRequest represents a policy decision request
type PolicyDecisionRequest struct {
	ConsumerID     string                 `json:"consumer_id"`
	AppID          string                 `json:"app_id"`
	RequestID      string                 `json:"request_id"`
	RequiredFields []string               `json:"required_fields"`
	Context        map[string]interface{} `json:"context,omitempty"`
}
