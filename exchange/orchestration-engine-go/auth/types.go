package auth

import "time"

// AuthType represents the type of authentication method
type AuthType string

const (
	// AuthTypeAPIKey represents API key authentication
	AuthTypeAPIKey AuthType = "apiKey"
	// AuthTypeOAuth2 represents OAuth2 authentication
	AuthTypeOAuth2 AuthType = "oauth2"
)

// Auth2TokenResponse represents the response from an OAuth2 token request
type Auth2TokenResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_in,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// AuthConfig represents the configuration for provider authentication
type AuthConfig struct {
	Type AuthType `json:"type"` // "apiKey" or "oauth2"

	// For API Key auth
	APIKeyName  string `json:"apiKeyName,omitempty"`
	APIKeyValue string `json:"apiKeyValue,omitempty"`

	// For OAuth2 auth
	TokenURL     string `json:"tokenUrl,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}
