package models

// ConsumerGrant represents a consumer's approved field access
type ConsumerGrant struct {
	ConsumerID     string   `json:"consumerId"`
	ApprovedFields []string `json:"approvedFields"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// ConsumerGrantsData represents the complete consumer grants data structure
type ConsumerGrantsData struct {
	LegacyConsumerGrants map[string]ConsumerGrant `json:"legacy_consumer_grants"`
}

// CreateConsumerGrantRequest represents the request to create a new consumer grant
type CreateConsumerGrantRequest struct {
	ConsumerID     string   `json:"consumerId" validate:"required"`
	ApprovedFields []string `json:"approvedFields" validate:"required,min=1"`
}

// UpdateConsumerGrantRequest represents the request to update a consumer grant
type UpdateConsumerGrantRequest struct {
	ApprovedFields []string `json:"approvedFields" validate:"required,min=1"`
}

// ProviderField represents a single field in provider metadata
type ProviderField struct {
	Owner             string                 `json:"owner" validate:"required"`
	Provider          string                 `json:"provider" validate:"required"`
	ConsentRequired   bool                   `json:"consent_required"`
	AccessControlType string                 `json:"access_control_type" validate:"required,oneof=public restricted"`
	AllowList         []AllowListEntry       `json:"allow_list"`
	Description       string                 `json:"description,omitempty"`
	ExpiryTime        string                 `json:"expiry_time,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// AllowListEntry represents an entry in the allow list for restricted fields
type AllowListEntry struct {
	ConsumerID string `json:"consumerId" validate:"required"`
	ExpiryTime string `json:"expiry_time" validate:"required"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// ProviderMetadataData represents the complete provider metadata structure
type ProviderMetadataData struct {
	Fields map[string]ProviderField `json:"fields"`
}

// CreateProviderFieldRequest represents the request to create a new provider field
type CreateProviderFieldRequest struct {
	FieldName        string           `json:"fieldName" validate:"required"`
	Owner            string           `json:"owner" validate:"required"`
	Provider         string           `json:"provider" validate:"required"`
	ConsentRequired  bool             `json:"consent_required"`
	AccessControlType string          `json:"access_control_type" validate:"required,oneof=public restricted"`
	AllowList        []AllowListEntry `json:"allow_list"`
	Description      string           `json:"description,omitempty"`
	ExpiryTime       string           `json:"expiry_time,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateProviderFieldRequest represents the request to update a provider field
type UpdateProviderFieldRequest struct {
	Owner             *string          `json:"owner,omitempty"`
	Provider          *string          `json:"provider,omitempty"`
	ConsentRequired   *bool            `json:"consent_required,omitempty"`
	AccessControlType *string          `json:"access_control_type,omitempty" validate:"omitempty,oneof=public restricted"`
	AllowList         []AllowListEntry `json:"allow_list,omitempty"`
	Description       *string          `json:"description,omitempty"`
	ExpiryTime        *string          `json:"expiry_time,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// SchemaConversionRequest represents the request to convert GraphQL SDL to provider metadata
type SchemaConversionRequest struct {
	ProviderID string `json:"providerId" validate:"required"`
	SDL        string `json:"sdl" validate:"required"`
}

// SchemaConversionResponse represents the response from schema conversion
type SchemaConversionResponse struct {
	ProviderID string                 `json:"providerId"`
	Fields     map[string]ProviderField `json:"fields"`
	ConvertedAt string                `json:"convertedAt"`
}
