package models

import "time"

// ApplicationStatus represents the status of a data consumer's application
type ApplicationStatus string

const (
	StatusPending  ApplicationStatus = "pending"
	StatusApproved ApplicationStatus = "approved"
	StatusDenied   ApplicationStatus = "denied"
)

// Consumer represents a data consumer organization
type Consumer struct {
	ConsumerID   string    `json:"consumerId"`
	ConsumerName string    `json:"consumerName"`
	ContactEmail string    `json:"contactEmail"`
	PhoneNumber  string    `json:"phoneNumber"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ConsumerApp represents a consumer's application to access specific data fields
type ConsumerApp struct {
	SubmissionID   string              `json:"submissionId"`
	ConsumerID     string              `json:"consumerId"`
	Status         ApplicationStatus   `json:"status"`
	RequiredFields map[string]bool     `json:"requiredFields"`
	CreatedAt      time.Time           `json:"createdAt"`
	Credentials    *Credentials        `json:"credentials,omitempty"`
	AsgardeoClient *AsgardeoClientInfo `json:"asgardeoClient,omitempty"`
}

// Credentials represents API credentials for a consumer
type Credentials struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

// AsgardeoClientInfo represents Asgardeo client information for a consumer app
type AsgardeoClientInfo struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret,omitempty"` // Omitted in responses for security
	ClientName   string    `json:"client_name"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
	Scopes       []string  `json:"scopes"`
}

// ClientMapping represents the mapping between consumer ID and Asgardeo client ID
type ClientMapping struct {
	ConsumerID string    `json:"consumerId"`
	ClientID   string    `json:"clientId"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CreateConsumerRequest represents the request to create a new consumer
type CreateConsumerRequest struct {
	ConsumerName string `json:"consumerName"`
	ContactEmail string `json:"contactEmail"`
	PhoneNumber  string `json:"phoneNumber"`
}

// UpdateConsumerRequest represents the request to update a consumer
type UpdateConsumerRequest struct {
	ConsumerName *string `json:"consumerName,omitempty"`
	ContactEmail *string `json:"contactEmail,omitempty"`
	PhoneNumber  *string `json:"phoneNumber,omitempty"`
}

// CreateConsumerAppRequest represents the request to create a new consumer application
type CreateConsumerAppRequest struct {
	ConsumerID     string          `json:"consumerId"`
	RequiredFields map[string]bool `json:"required_fields"`
}

// UpdateConsumerAppRequest represents the request to update a consumer application
type UpdateConsumerAppRequest struct {
	Status         *ApplicationStatus `json:"status,omitempty"`
	RequiredFields map[string]bool    `json:"required_fields,omitempty"`
}

// UpdateConsumerAppResponse represents the response when updating a consumer application
type UpdateConsumerAppResponse struct {
	*ConsumerApp
	ProviderID string `json:"providerId,omitempty"` // Only present when status is approved
}

// Token exchange models

// CredentialMapping represents the mapping between API credentials and Asgardeo credentials
type CredentialMapping struct {
	APIKey               string `json:"apiKey"`
	APISecret            string `json:"apiSecret"`
	AsgardeoClientID     string `json:"asgardeoClientId"`
	AsgardeoClientSecret string `json:"asgardeoClientSecret"`
	ConsumerID           string `json:"consumerId"`
}

// DEPRECATED: TokenExchangeRequest is no longer used in the new M2M authentication flow.
// Consumer applications now get tokens directly from Asgardeo, bypassing the API Server.
type TokenExchangeRequest struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
	Scope     string `json:"scope,omitempty"`
}

// TokenExchangeResponse represents the response from token exchange
type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
	ConsumerID  string `json:"consumerId"`
}

// Legacy models for backward compatibility
type Application = ConsumerApp
type CreateApplicationRequest = CreateConsumerAppRequest
type UpdateApplicationRequest = UpdateConsumerAppRequest
type UpdateApplicationResponse = UpdateConsumerAppResponse
