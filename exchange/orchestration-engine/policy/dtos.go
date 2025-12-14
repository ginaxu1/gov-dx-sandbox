package policy

// RequiredField represents a field that requires policy decision
type RequiredField struct {
	FieldName string `json:"fieldName"`
	SchemaId  string `json:"schemaId"`
}

// PdpRequest represents a policy decision request
type PdpRequest struct {
	AppId          string          `json:"applicationId"`
	RequiredFields []RequiredField `json:"requiredFields"`
}

// ConsentRequiredField represents a field that requires consent
// Matches PolicyDecisionResponseFieldRecord DTO structure from PolicyDecisionPoint
type ConsentRequiredField struct {
	FieldName   string     `json:"fieldName"`
	SchemaID    string     `json:"schemaId"`
	DisplayName *string    `json:"displayName,omitempty"`
	Description *string    `json:"description,omitempty"`
	Owner       *OwnerType `json:"owner,omitempty"`
}

// PdpResponse represents a policy decision response
type PdpResponse struct {
	AppAuthorized           bool                   `json:"appAuthorized"`
	UnauthorizedFields      []ConsentRequiredField `json:"unauthorizedFields"`
	AppAccessExpired        bool                   `json:"appAccessExpired"`
	ExpiredFields           []ConsentRequiredField `json:"expiredFields"`
	AppRequiresOwnerConsent bool                   `json:"appRequiresOwnerConsent"`
	ConsentRequiredFields   []ConsentRequiredField `json:"consentRequiredFields"`
}
