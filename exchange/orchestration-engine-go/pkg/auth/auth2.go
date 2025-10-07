package auth

import "time"

type AuthType string

const (
	AuthTypeAPIKey AuthType = "apiKey"
	AuthTypeOAuth2 AuthType = "oauth2"
)

type OAuth2AuthMethod string

const (
	OAuth2AuthMethodBody   OAuth2AuthMethod = "body"   // Send credentials in request body (default)
	OAuth2AuthMethodBasic  OAuth2AuthMethod = "basic"  // Send credentials as Basic Auth header
	OAuth2AuthMethodBearer OAuth2AuthMethod = "bearer" // Send credentials as Bearer token
)

type Auth2TokenResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IssuedAt     time.Time `json:"issued_at,omitempty"` // Track when token was issued
}

type AuthConfig struct {
	Type AuthType `json:"type"` // "apiKey" or "oauth2"

	// For API Key auth
	APIKeyName  string `json:"apiKeyName,omitempty"`
	APIKeyValue string `json:"apiKeyValue,omitempty"`

	// For OAuth2 auth
	TokenURL     string           `json:"tokenUrl,omitempty"`
	ClientID     string           `json:"clientId,omitempty"`
	ClientSecret string           `json:"clientSecret,omitempty"`
	Scope        string           `json:"scope,omitempty"`
	AuthMethod   OAuth2AuthMethod `json:"authMethod,omitempty"` // How to send credentials (default: "body")

	// Token management configuration
	TokenConfig *TokenConfig `json:"tokenConfig,omitempty"`
}

// TokenConfig holds configuration for token management
type TokenConfig struct {
	RefreshBuffer    time.Duration `json:"refreshBuffer,omitempty"`    // When to start refreshing (default: 2 minutes)
	ValidationBuffer time.Duration `json:"validationBuffer,omitempty"` // Buffer for validation (default: 30 seconds)
	MaxRetries       int           `json:"maxRetries,omitempty"`       // Max retry attempts (default: 3)
	RetryDelay       time.Duration `json:"retryDelay,omitempty"`       // Delay between retries (default: 5 seconds)
}

// GetTokenConfig returns the token configuration with defaults applied
func (tc *TokenConfig) GetTokenConfig() *TokenConfig {
	if tc == nil {
		tc = &TokenConfig{}
	}

	// Apply defaults
	if tc.RefreshBuffer == 0 {
		tc.RefreshBuffer = 2 * time.Minute
	}
	if tc.ValidationBuffer == 0 {
		tc.ValidationBuffer = 30 * time.Second
	}
	if tc.MaxRetries == 0 {
		tc.MaxRetries = 3
	}
	if tc.RetryDelay == 0 {
		tc.RetryDelay = 5 * time.Second
	}

	return tc
}
