package models

// ProviderMetadataUpdateRequest represents the request to update provider metadata
type ProviderMetadataUpdateRequest struct {
	ApplicationID string               `json:"application_id" validate:"required"`
	Fields        []ProviderFieldGrant `json:"fields" validate:"required,min=1"`
}

// ProviderFieldGrant represents a field grant for a specific consumer
type ProviderFieldGrant struct {
	FieldName     string `json:"fieldName" validate:"required"`
	GrantDuration string `json:"grantDuration" validate:"required"`
}

// ProviderMetadataField represents a field in the provider metadata JSON
type ProviderMetadataField struct {
	Owner             string              `json:"owner"`
	Provider          string              `json:"provider"`
	ConsentRequired   bool                `json:"consent_required"`
	AccessControlType string              `json:"access_control_type"`
	AllowList         []PDPAllowListEntry `json:"allow_list"`
}

// PDPAllowListEntry represents an entry in the allow list for PDP metadata
type PDPAllowListEntry struct {
	ConsumerID    string `json:"consumer_id"`
	ExpiresAt     int64  `json:"expires_at"`
	GrantDuration string `json:"grant_duration"`
}

// ProviderMetadata represents the complete provider metadata structure
type ProviderMetadata struct {
	Fields map[string]ProviderMetadataField `json:"fields"`
}

// ProviderMetadataUpdateResponse represents the response from PDP metadata update
type ProviderMetadataUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Updated int    `json:"updated,omitempty"`
}
