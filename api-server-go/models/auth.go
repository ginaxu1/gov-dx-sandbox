package models

// DEPRECATED: These models are no longer used in the new M2M authentication flow.
// Choreo Gateway now handles JWT validation directly, eliminating the need for
// these token validation endpoints in the API Server.

// ValidateTokenRequest represents the request to validate a token
type ValidateTokenRequest struct {
	Token string `json:"token"`
}

// ValidateTokenResponse represents the response from token validation
type ValidateTokenResponse struct {
	Valid      bool   `json:"valid"`
	ConsumerID string `json:"consumerId,omitempty"`
	Error      string `json:"error,omitempty"`
}

// AsgardeoUserInfo represents user information from Asgardeo
type AsgardeoUserInfo struct {
	Sub               string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}
