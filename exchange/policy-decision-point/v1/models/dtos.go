package models

// Request and Response DTOs

// PolicyMetadataCreateRequestRecord represents the request to create policy metadata
type PolicyMetadataCreateRequestRecord struct {
	FieldName         string            `json:"field_name" validate:"required"`
	DisplayName       *string           `json:"display_name,omitempty"`
	Description       *string           `json:"description,omitempty"`
	Source            Source            `json:"source" validate:"required,source_enum"`
	IsOwner           bool              `json:"is_owner" validate:"required"`
	AccessControlType AccessControlType `json:"access_control_type" validate:"required,access_control_type_enum"`
	Owner             *Owner            `json:"owner,omitempty" validate:"omitempty,owner_enum"`
}

// PolicyMetadataCreateRequest represents the request to create policy metadata
type PolicyMetadataCreateRequest struct {
	SchemaID string                              `json:"schema_id" validate:"required"`
	Records  []PolicyMetadataCreateRequestRecord `json:"records" validate:"required,dive"`
}

// PolicyMetadataResponse represents the response from policy metadata operations
type PolicyMetadataResponse struct {
	ID                string            `json:"id"`
	SchemaID          string            `json:"schema_id"`
	FieldName         string            `json:"field_name"`
	DisplayName       *string           `json:"display_name,omitempty"`
	Description       *string           `json:"description,omitempty"`
	Source            Source            `json:"source"`
	IsOwner           bool              `json:"is_owner"`
	AccessControlType AccessControlType `json:"access_control_type"`
	AllowList         AllowList         `json:"allow_list"`
	Owner             *Owner            `json:"owner,omitempty"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// PolicyMetadataCreateResponse represents the response from policy metadata creation
type PolicyMetadataCreateResponse struct {
	Records []PolicyMetadataResponse `json:"records"`
}

// AllowListUpdateRequestRecord represents the one record of request to update allow list
type AllowListUpdateRequestRecord struct {
	FieldName string `json:"field_name" validate:"required"`
	SchemaID  string `json:"schema_id" validate:"required"`
}

// AllowListUpdateRequest represents the request to update allow list
type AllowListUpdateRequest struct {
	ApplicationID string                         `json:"application_id" validate:"required"`
	Records       []AllowListUpdateRequestRecord `json:"records" validate:"required,dive"`
	GrantDuration GrantDurationType              `json:"grant_duration" validate:"required,grant_duration_type_enum"`
}

// AllowListUpdateResponseRecord represents one record in the allow list update response
type AllowListUpdateResponseRecord struct {
	FieldName string `json:"field_name"`
	SchemaID  string `json:"schema_id"`
	ExpiresAt string `json:"expires_at"`
	UpdatedAt string `json:"updated_at"`
}

// AllowListUpdateResponse represents the response from allow list update
type AllowListUpdateResponse struct {
	Records []AllowListUpdateResponseRecord `json:"records"`
}

// PolicyDecisionRequestRecord represents a policy decision request record
type PolicyDecisionRequestRecord struct {
	FieldName string `json:"field_name"`
	SchemaID  string `json:"schema_id"`
}

// PolicyDecisionRequest represents a policy decision request
type PolicyDecisionRequest struct {
	ApplicationID  string                        `json:"application_id" validate:"required"`
	RequiredFields []PolicyDecisionRequestRecord `json:"required_fields" validate:"required,dive"`
}

// PolicyDecisionResponseFieldRecord represents a policy decision response record
type PolicyDecisionResponseFieldRecord struct {
	FieldName   string  `json:"field_name"`
	SchemaID    string  `json:"schema_id"`
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Owner       *Owner  `json:"owner,omitempty"`
}

// PolicyDecisionResponse represents a policy decision response
type PolicyDecisionResponse struct {
	AppNotAuthorized        bool                                `json:"app_not_authorized"`
	UnAuthorizedFields      []PolicyDecisionResponseFieldRecord `json:"unauthorized_fields"`
	AppAccessExpired        bool                                `json:"app_access_expired"`
	ExpiredFields           []PolicyDecisionResponseFieldRecord `json:"expired_fields"`
	AppRequiresOwnerConsent bool                                `json:"app_requires_owner_consent"`
	ConsentRequiredFields   []PolicyDecisionResponseFieldRecord `json:"consent_required_fields"`
}
