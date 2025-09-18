package models

import "time"

// AuthRequest represents the request to authenticate a consumer
type AuthRequest struct {
	ConsumerID string `json:"consumerId"`
	Secret     string `json:"secret"`
}

// AuthResponse represents the response from authentication
type AuthResponse struct {
	AccessToken  string            `json:"accessToken"`
	TokenType    string            `json:"tokenType"`
	ExpiresIn    int64             `json:"expiresIn"`
	ExpiresAt    time.Time         `json:"expiresAt"`
	ConsumerID   string            `json:"consumerId"`
	AsgardeoUser *AsgardeoUserInfo `json:"asgardeoUser,omitempty"`
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	ConsumerID string    `json:"consumerId"`
	IssuedAt   time.Time `json:"iat"`
	ExpiresAt  time.Time `json:"exp"`
	Issuer     string    `json:"iss"`
}

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
