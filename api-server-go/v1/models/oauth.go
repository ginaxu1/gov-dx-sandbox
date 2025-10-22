package models

import (
	"time"
)

// OAuth2Client represents a registered OAuth 2.0 client application
type OAuth2Client struct {
	ClientID     string    `json:"client_id" db:"client_id"`
	ClientSecret string    `json:"client_secret" db:"client_secret"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	RedirectURI  string    `json:"redirect_uri" db:"redirect_uri"`
	Scopes       []string  `json:"scopes" db:"scopes"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// OAuth2AuthorizationCode represents a temporary authorization code
type OAuth2AuthorizationCode struct {
	Code                string    `json:"code" db:"code"`
	ClientID            string    `json:"client_id" db:"client_id"`
	UserID              string    `json:"user_id" db:"user_id"`
	RedirectURI         string    `json:"redirect_uri" db:"redirect_uri"`
	Scopes              []string  `json:"scopes" db:"scopes"`
	CodeChallenge       string    `json:"code_challenge" db:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method" db:"code_challenge_method"`
	ExpiresAt           time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	Used                bool      `json:"used" db:"used"`
}

// OAuth2AccessToken represents an access token issued to a client
type OAuth2AccessToken struct {
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ClientID     string    `json:"client_id" db:"client_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Scopes       []string  `json:"scopes" db:"scopes"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// OAuth2RefreshToken represents a refresh token for obtaining new access tokens
type OAuth2RefreshToken struct {
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ClientID     string    `json:"client_id" db:"client_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Scopes       []string  `json:"scopes" db:"scopes"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// OAuth2Token represents a consolidated token (access or refresh) issued to a client
type OAuth2Token struct {
	Token        string    `json:"token" db:"token"`
	TokenType    string    `json:"token_type" db:"token_type"` // 'access' or 'refresh'
	ClientID     string    `json:"client_id" db:"client_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Scopes       []string  `json:"scopes" db:"scopes"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	RelatedToken *string   `json:"related_token,omitempty" db:"related_token"` // For access tokens, links to refresh token
	ParentToken  *string   `json:"parent_token,omitempty" db:"parent_token"`   // For refresh tokens, links to access token
}

// Request/Response DTOs

// CreateOAuth2ClientRequest represents the request to create a new OAuth 2.0 client
type CreateOAuth2ClientRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	RedirectURI string   `json:"redirect_uri" validate:"required,url"`
	Scopes      []string `json:"scopes"`
}

// CreateOAuth2ClientResponse represents the response when creating a new OAuth 2.0 client
type CreateOAuth2ClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	RedirectURI  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
	CreatedAt    string   `json:"created_at"`
}

// AuthorizationRequest represents the OAuth 2.0 authorization request
type AuthorizationRequest struct {
	ResponseType        string `json:"response_type" validate:"required,eq=code"`
	ClientID            string `json:"client_id" validate:"required"`
	RedirectURI         string `json:"redirect_uri" validate:"required,url"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// AuthorizationResponse represents the OAuth 2.0 authorization response
type AuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty"`
}

// TokenRequest represents the OAuth 2.0 token request
type TokenRequest struct {
	GrantType    string `json:"grant_type" validate:"required,oneof=authorization_code client_credentials"`
	Code         string `json:"code"`         // Required for authorization_code, not for client_credentials
	RedirectURI  string `json:"redirect_uri"` // Required for authorization_code, not for client_credentials
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	CodeVerifier string `json:"code_verifier"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// RefreshTokenRequest represents the OAuth 2.0 refresh token request
type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type" validate:"required,eq=refresh_token"`
	RefreshToken string `json:"refresh_token" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// UserInfo represents user information extracted from a JWT token
type UserInfo struct {
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Scopes    []string `json:"scopes"`
	ClientID  string   `json:"client_id"`
}

// DataRequest represents a request for user data
type DataRequest struct {
	Fields []string `json:"fields" validate:"required,min=1"`
}

// DataResponse represents the response containing user data
type DataResponse struct {
	UserID string                 `json:"user_id"`
	Data   map[string]interface{} `json:"data"`
	Fields []string               `json:"fields"`
}

// ErrorResponse represents an OAuth 2.0 error response
type OAuth2ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	State            string `json:"state,omitempty"`
}
