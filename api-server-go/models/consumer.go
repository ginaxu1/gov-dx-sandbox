package models

// ApplicationStatus represents the status of a data consumer's application
type ApplicationStatus string

const (
	StatusPending  ApplicationStatus = "pending"
	StatusApproved ApplicationStatus = "approved"
	StatusDenied   ApplicationStatus = "denied"
)

// Application represents a data consumer's application to access specific data fields
type Application struct {
	AppID          string                 `json:"appId"`
	Status         ApplicationStatus      `json:"status"`
	RequiredFields map[string]interface{} `json:"requiredFields"`
	Credentials    *Credentials           `json:"credentials,omitempty"`
}

// Credentials represents API credentials for a consumer
type Credentials struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

// CreateApplicationRequest represents the request to create a new application
type CreateApplicationRequest struct {
	RequiredFields map[string]interface{} `json:"requiredFields"`
}

// UpdateApplicationRequest represents the request to update an application
type UpdateApplicationRequest struct {
	Status         *ApplicationStatus     `json:"status,omitempty"`
	RequiredFields map[string]interface{} `json:"requiredFields,omitempty"`
}

// UpdateApplicationResponse represents the response when updating a consumer application
type UpdateApplicationResponse struct {
	*Application
	ProviderID string `json:"providerId,omitempty"` // Only present when status is approved
}
