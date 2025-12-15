package consent

// ConsentField represents a field that requires consent
// Matches PolicyDecisionResponseFieldRecord DTO structure from PolicyDecisionPoint
type ConsentField struct {
	FieldName   string    `json:"fieldName"`
	SchemaID    string    `json:"schemaId"`
	DisplayName *string   `json:"displayName,omitempty"`
	Description *string   `json:"description,omitempty"`
	Owner       OwnerType `json:"owner"`
}

// ConsentRequirement represents a consent requirement for a specific owner
type ConsentRequirement struct {
	Owner   OwnerType      `json:"owner"`
	OwnerID string         `json:"owner_id"`
	Fields  []ConsentField `json:"fields"`
}

// CreateConsentRequest defines the structure for creating a consent record
// GrantDuration is optional - nil means not provided and will use default value
type CreateConsentRequest struct {
	AppID               string               `json:"app_id"`
	AppName             *string              `json:"app_name,omitempty"`
	ConsentRequirements []ConsentRequirement `json:"consent_requirements"`
	GrantDuration       *string              `json:"grant_duration,omitempty"`
	ConsentType         *ConsentType         `json:"consent_type,omitempty"`
}

// ConsentResponseInternalView represents a simplified consent response structure for Internal API Responses
type ConsentResponseInternalView struct {
	ConsentID        string          `json:"consent_id"`
	Status           ConsentStatus   `json:"status"`
	ConsentPortalURL *string         `json:"consent_portal_url,omitempty"` // Only present when status is pending
	Fields           *[]ConsentField `json:"fields,omitempty"`             // Included for internal view
}
