package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/pkg/errors"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/shared"
	"golang.org/x/oauth2"
)

// OAuth2Service handles OAuth 2.0 operations
type OAuth2Service struct {
	db           *sql.DB
	baseURL      string
	oauth2Config *oauth2.Config
}

// NewOAuth2Service creates a new OAuth 2.0 service
func NewOAuth2Service(db *sql.DB, baseURL ...string) *OAuth2Service {
	url := "http://localhost:3000"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  url + "/oauth2/authorize",
			TokenURL: url + "/oauth2/token",
		},
	}

	return &OAuth2Service{
		db:           db,
		baseURL:      url,
		oauth2Config: config,
	}
}

// PKCE helper functions

// GenerateCodeVerifier generates a cryptographically random code verifier for PKCE
func GenerateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

// GenerateCodeChallenge generates a code challenge from a code verifier using S256 method
func GenerateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}

// VerifyCodeChallenge verifies a code challenge against a code verifier
func VerifyCodeChallenge(codeVerifier, codeChallenge, method string) bool {
	if method != "S256" {
		return false
	}
	expectedChallenge := GenerateCodeChallenge(codeVerifier)
	return codeChallenge == expectedChallenge
}

// CreateClient creates a new OAuth 2.0 client
func (s *OAuth2Service) CreateClient(req models.CreateOAuth2ClientRequest) (*models.CreateOAuth2ClientResponse, error) {
	// Generate client credentials
	clientID := "client_" + uuid.New().String()
	clientSecret := s.generateClientSecret()

	// Default scopes if none provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:data"}
	}

	now := time.Now()
	client := &models.OAuth2Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         req.Name,
		Description:  req.Description,
		RedirectURI:  req.RedirectURI,
		Scopes:       scopes,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert client into database
	query := `INSERT INTO oauth2_clients (client_id, client_secret, name, description, redirect_uri, scopes, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	scopesJSON := s.scopesToJSON(scopes)
	_, err := s.db.Exec(query, client.ClientID, client.ClientSecret, client.Name, client.Description,
		client.RedirectURI, scopesJSON, client.IsActive, client.CreatedAt, client.UpdatedAt)
	if err != nil {
		slog.Error("Failed to create OAuth2 client", "error", err)
		return nil, errors.HandleDatabaseError(err, "create OAuth2 client")
	}

	response := &models.CreateOAuth2ClientResponse{
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Name:         client.Name,
		Description:  client.Description,
		RedirectURI:  client.RedirectURI,
		Scopes:       client.Scopes,
		CreatedAt:    client.CreatedAt.Format(time.RFC3339),
	}

	slog.Info("Created OAuth2 client", "client_id", client.ClientID, "name", client.Name)
	return response, nil
}

// GetClient retrieves an OAuth 2.0 client by ID
func (s *OAuth2Service) GetClient(clientID string) (*models.OAuth2Client, error) {
	query := `SELECT client_id, client_secret, name, description, redirect_uri, scopes, is_active, created_at, updated_at 
			  FROM oauth2_clients WHERE client_id = ? AND is_active = true`

	var client models.OAuth2Client
	var scopesJSON string

	err := s.db.QueryRow(query, clientID).Scan(
		&client.ClientID, &client.ClientSecret, &client.Name, &client.Description,
		&client.RedirectURI, &scopesJSON, &client.IsActive, &client.CreatedAt, &client.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundError("OAuth2 client not found")
		}
		slog.Error("Failed to get OAuth2 client", "error", err, "client_id", clientID)
		return nil, errors.HandleDatabaseError(err, "get OAuth2 client")
	}

	client.Scopes = s.jsonToScopes(scopesJSON)
	return &client, nil
}

// ValidateClient validates client credentials and redirect URI
func (s *OAuth2Service) ValidateClient(clientID, clientSecret, redirectURI string) (*models.OAuth2Client, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, errors.UnauthorizedError("Invalid client credentials")
	}

	if client.RedirectURI != redirectURI {
		return nil, fmt.Errorf("invalid redirect URI")
	}

	return client, nil
}

// ValidateClientCredentials validates only client ID and secret (for refresh tokens and client credentials)
func (s *OAuth2Service) ValidateClientCredentials(clientID, clientSecret string) (*models.OAuth2Client, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, errors.UnauthorizedError("Invalid client credentials")
	}

	return client, nil
}

// CreateAuthorizationCode creates a new authorization code
func (s *OAuth2Service) CreateAuthorizationCode(clientID, userID, redirectURI string, scopes []string) (*models.OAuth2AuthorizationCode, error) {
	code := s.generateAuthorizationCode()
	expiresAt := time.Now().Add(10 * time.Minute) // Authorization codes expire in 10 minutes

	authCode := &models.OAuth2AuthorizationCode{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		Used:        false,
	}

	query := `INSERT INTO oauth2_authorization_codes (code, client_id, user_id, redirect_uri, scopes, expires_at, created_at, used) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	scopesJSON := s.scopesToJSON(scopes)
	_, err := s.db.Exec(query, authCode.Code, authCode.ClientID, authCode.UserID, authCode.RedirectURI,
		scopesJSON, authCode.ExpiresAt, authCode.CreatedAt, authCode.Used)
	if err != nil {
		slog.Error("Failed to create authorization code", "error", err)
		return nil, errors.HandleDatabaseError(err, "create authorization code")
	}

	slog.Info("Created authorization code", "code", code, "client_id", clientID, "user_id", userID)
	return authCode, nil
}

// ValidateAuthorizationCode validates and consumes an authorization code
func (s *OAuth2Service) ValidateAuthorizationCode(code, clientID, redirectURI string) (*models.OAuth2AuthorizationCode, error) {
	query := `SELECT code, client_id, user_id, redirect_uri, scopes, expires_at, created_at, used 
			  FROM oauth2_authorization_codes WHERE code = ? AND client_id = ? AND redirect_uri = ?`

	var authCode models.OAuth2AuthorizationCode
	var scopesJSON string

	err := s.db.QueryRow(query, code, clientID, redirectURI).Scan(
		&authCode.Code, &authCode.ClientID, &authCode.UserID, &authCode.RedirectURI,
		&scopesJSON, &authCode.ExpiresAt, &authCode.CreatedAt, &authCode.Used,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UnauthorizedError("Invalid authorization code")
		}
		slog.Error("Failed to validate authorization code", "error", err)
		return nil, errors.HandleDatabaseError(err, "validate authorization code")
	}

	authCode.Scopes = s.jsonToScopes(scopesJSON)

	// Check if code is expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, errors.UnauthorizedError("Authorization code has expired")
	}

	// Check if code has already been used
	if authCode.Used {
		return nil, errors.UnauthorizedError("Authorization code has already been used")
	}

	// Mark code as used
	updateQuery := `UPDATE oauth2_authorization_codes SET used = true WHERE code = ?`
	_, err = s.db.Exec(updateQuery, code)
	if err != nil {
		slog.Error("Failed to mark authorization code as used", "error", err)
		return nil, errors.HandleDatabaseError(err, "mark authorization code as used")
	}

	authCode.Used = true
	return &authCode, nil
}

// CreateAccessToken creates a new access token
func (s *OAuth2Service) CreateAccessToken(clientID, userID string, scopes []string) (*models.OAuth2AccessToken, error) {
	accessToken := s.generateAccessToken()
	refreshToken := s.generateRefreshToken()
	expiresAt := time.Now().Add(1 * time.Hour) // Access tokens expire in 1 hour

	token := &models.OAuth2AccessToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		UserID:       userID,
		Scopes:       scopes,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	// Start a transaction to ensure both tokens are stored atomically
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Store access token
	accessQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	scopesJSON := s.scopesToJSON(scopes)
	_, err = tx.Exec(accessQuery, token.AccessToken, "access", token.ClientID, token.UserID,
		scopesJSON, token.ExpiresAt, token.CreatedAt, token.IsActive)
	if err != nil {
		slog.Error("Failed to create access token", "error", err)
		return nil, errors.HandleDatabaseError(err, "create access token")
	}

	// Store refresh token
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	refreshQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active, parent_token)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(refreshQuery, token.RefreshToken, "refresh", token.ClientID, token.UserID,
		scopesJSON, refreshExpiresAt, token.CreatedAt, token.IsActive, token.AccessToken)
	if err != nil {
		slog.Error("Failed to create refresh token", "error", err)
		return nil, errors.HandleDatabaseError(err, "create refresh token")
	}

	// Update access token to link to refresh token
	updateQuery := `UPDATE oauth2_tokens SET related_token = ? WHERE token = ?`
	_, err = tx.Exec(updateQuery, token.RefreshToken, token.AccessToken)
	if err != nil {
		slog.Error("Failed to link tokens", "error", err)
		return nil, errors.HandleDatabaseError(err, "link tokens")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	slog.Info("Created access token", "client_id", clientID, "user_id", userID)
	return token, nil
}

// ValidateAccessToken validates an access token and returns user information
// This is an alias for ValidateToken to maintain backward compatibility
func (s *OAuth2Service) ValidateAccessToken(accessToken string) (*models.UserInfo, error) {
	return s.ValidateToken(accessToken)
}

// RefreshAccessToken creates a new access token using a refresh token
func (s *OAuth2Service) RefreshAccessToken(refreshToken, clientID, clientSecret string) (*models.OAuth2AccessToken, error) {
	// Validate client credentials (no redirect URI needed for refresh tokens)
	_, err := s.ValidateClientCredentials(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	// Validate refresh token using the consolidated oauth2_tokens table
	query := `SELECT token, client_id, user_id, scopes, expires_at, is_active 
			  FROM oauth2_tokens WHERE token = ? AND client_id = ? AND token_type = 'refresh'`

	var refresh models.OAuth2RefreshToken
	var scopesJSON string

	err = s.db.QueryRow(query, refreshToken, clientID).Scan(
		&refresh.RefreshToken, &refresh.ClientID, &refresh.UserID, &scopesJSON, &refresh.ExpiresAt, &refresh.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UnauthorizedError("Invalid refresh token")
		}
		slog.Error("Failed to validate refresh token", "error", err)
		return nil, errors.HandleDatabaseError(err, "validate refresh token")
	}

	refresh.Scopes = s.jsonToScopes(scopesJSON)

	// Check if refresh token is expired
	if time.Now().After(refresh.ExpiresAt) {
		return nil, errors.UnauthorizedError("Refresh token has expired")
	}

	// Check if refresh token is active
	if !refresh.IsActive {
		return nil, errors.UnauthorizedError("Refresh token is not active")
	}

	// Deactivate old access tokens for this client and user
	deactivateQuery := `UPDATE oauth2_tokens SET is_active = false WHERE client_id = ? AND user_id = ? AND token_type = 'access'`
	_, err = s.db.Exec(deactivateQuery, clientID, refresh.UserID)
	if err != nil {
		slog.Error("Failed to deactivate old access tokens", "error", err)
		return nil, errors.HandleDatabaseError(err, "deactivate old access tokens")
	}

	// Create new access token
	return s.CreateAccessToken(clientID, refresh.UserID, refresh.Scopes)
}

// GetUserData retrieves user-specific data based on requested fields
func (s *OAuth2Service) GetUserData(userID string, fields []string) (*models.DataResponse, error) {
	// This is a mock implementation - in a real system, this would query the actual data sources
	// For now, we'll return mock data based on the user ID and requested fields

	mockData := map[string]interface{}{
		"person.fullName":              "John Doe",
		"person.email":                 "john.doe@example.com",
		"person.photo":                 "https://example.com/photos/john-doe.jpg",
		"birthinfo.birthDate":          "1990-01-15",
		"birthinfo.birthCertificateID": "BC123456789",
		"address.street":               "123 Main Street",
		"address.city":                 "Anytown",
		"address.country":              "USA",
	}

	// Filter data based on requested fields
	filteredData := make(map[string]interface{})
	for _, field := range fields {
		if value, exists := mockData[field]; exists {
			filteredData[field] = value
		}
	}

	response := &models.DataResponse{
		UserID: userID,
		Data:   filteredData,
		Fields: fields,
	}

	slog.Info("Retrieved user data", "user_id", userID, "fields", fields)
	return response, nil
}

// Helper methods

func (s *OAuth2Service) generateClientSecret() string {
	token, _ := shared.GenerateToken()
	return token
}

func (s *OAuth2Service) generateAuthorizationCode() string {
	token, _ := shared.GenerateToken()
	return token
}

func (s *OAuth2Service) generateAccessToken() string {
	// In a real implementation, this would be a JWT token
	// For now, we'll generate a random string
	token, _ := shared.GenerateToken()
	return token
}

func (s *OAuth2Service) generateRefreshToken() string {
	token, _ := shared.GenerateToken()
	return token
}

func (s *OAuth2Service) scopesToJSON(scopes []string) string {
	return shared.ScopesToJSON(scopes)
}

func (s *OAuth2Service) jsonToScopes(scopesJSON string) []string {
	return shared.JSONToScopes(scopesJSON)
}

func (s *OAuth2Service) getUserInfo(userID string) (*models.UserInfo, error) {
	// This is a mock implementation - in a real system, this would query the user database
	// For now, we'll return mock user information
	userInfo := &models.UserInfo{
		UserID:    userID,
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	return userInfo, nil
}

// GenerateAuthCodeURL generates the authorization URL for a client using oauth2 package with PKCE support
func (s *OAuth2Service) GenerateAuthCodeURL(clientID, state string, codeChallenge string) (string, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return "", err
	}

	// Create a temporary oauth2 config for this client
	config := &oauth2.Config{
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		RedirectURL:  client.RedirectURI,
		Scopes:       client.Scopes,
		Endpoint:     s.oauth2Config.Endpoint,
	}

	// Generate the authorization URL with PKCE
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("code_challenge", codeChallenge), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	return authURL, nil
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (s *OAuth2Service) ExchangeCodeForToken(ctx context.Context, clientID, code, redirectURI, codeVerifier string) (*oauth2.Token, error) {
	// Validate and consume the authorization code
	authCode, err := s.ValidateAuthorizationCode(code, clientID, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired authorization code")
	}

	// Verify PKCE if code challenge is present
	if authCode.CodeChallenge != "" {
		if !VerifyCodeChallenge(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			return nil, fmt.Errorf("invalid code verifier")
		}
	}

	// Create access and refresh tokens
	accessToken := s.generateAccessToken()
	refreshToken := s.generateRefreshToken()
	expiresAt := time.Now().Add(1 * time.Hour) // Access tokens expire in 1 hour

	// Store both tokens in the database
	err = s.storeOAuth2TokensDirect(clientID, authCode.UserID, accessToken, refreshToken, authCode.Scopes, expiresAt)
	if err != nil {
		slog.Error("Failed to store tokens", "error", err)
		return nil, fmt.Errorf("failed to store tokens: %w", err)
	}

	// Create oauth2.Token response
	token := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		Expiry:       expiresAt,
	}

	slog.Info("Successfully exchanged authorization code for token", "client_id", clientID, "user_id", authCode.UserID)
	return token, nil
}

// RefreshToken refreshes an access token using a refresh token with oauth2 package
func (s *OAuth2Service) RefreshToken(ctx context.Context, clientID, refreshToken string) (*oauth2.Token, error) {
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	// Create a temporary oauth2 config for this client
	config := &oauth2.Config{
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		RedirectURL:  client.RedirectURI,
		Scopes:       client.Scopes,
		Endpoint:     s.oauth2Config.Endpoint,
	}

	// Create a token source with the refresh token
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		slog.Error("Failed to refresh token", "error", err, "client_id", clientID)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Store the new token
	err = s.storeOAuth2Token(clientID, newToken)
	if err != nil {
		slog.Error("Failed to store refreshed token", "error", err)
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return newToken, nil
}

// getAuthorizationCode retrieves an authorization code from the database
func (s *OAuth2Service) getAuthorizationCode(code string) (*models.OAuth2AuthorizationCode, error) {
	query := `SELECT code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at, created_at, used
			  FROM oauth2_authorization_codes WHERE code = ?`

	var authCode models.OAuth2AuthorizationCode
	var scopesJSON string

	err := s.db.QueryRow(query, code).Scan(
		&authCode.Code, &authCode.ClientID, &authCode.UserID, &authCode.RedirectURI,
		&scopesJSON, &authCode.CodeChallenge, &authCode.CodeChallengeMethod,
		&authCode.ExpiresAt, &authCode.CreatedAt, &authCode.Used,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("authorization code not found")
		}
		return nil, err
	}

	// Parse scopes
	authCode.Scopes = s.jsonToScopes(scopesJSON)

	// Check if code is expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, fmt.Errorf("authorization code has expired")
	}

	// Check if code is already used
	if authCode.Used {
		return nil, fmt.Errorf("authorization code has already been used")
	}

	return &authCode, nil
}

// markAuthorizationCodeAsUsed marks an authorization code as used
func (s *OAuth2Service) markAuthorizationCodeAsUsed(code string) error {
	query := `UPDATE oauth2_authorization_codes SET used = true WHERE code = ?`
	_, err := s.db.Exec(query, code)
	return err
}

// extractUserInfoFromRefreshToken extracts user ID and scopes from an existing refresh token
func (s *OAuth2Service) extractUserInfoFromRefreshToken(refreshToken string) (string, []string, error) {
	query := `SELECT user_id, scopes FROM oauth2_tokens WHERE token = ? AND token_type = 'refresh' AND is_active = true`

	var userID, scopesJSON string
	err := s.db.QueryRow(query, refreshToken).Scan(&userID, &scopesJSON)
	if err != nil {
		return "", nil, fmt.Errorf("refresh token not found or invalid: %w", err)
	}

	// Parse scopes from JSON
	scopes := shared.JSONToScopes(scopesJSON)
	return userID, scopes, nil
}

// storeOAuth2Token stores an oauth2.Token in our consolidated database
func (s *OAuth2Service) storeOAuth2Token(clientID string, token *oauth2.Token) error {
	// Extract user ID and scopes from the existing refresh token in the database
	userID, scopes, err := s.extractUserInfoFromRefreshToken(token.RefreshToken)
	if err != nil {
		slog.Error("Failed to extract user info from refresh token", "error", err)
		return fmt.Errorf("failed to extract user info from refresh token: %w", err)
	}
	now := time.Now()

	// Calculate access token expiry
	accessExpiresAt := token.Expiry
	if accessExpiresAt.IsZero() {
		accessExpiresAt = now.Add(1 * time.Hour)
	}

	// Start a transaction to ensure both tokens are stored atomically
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Store access token
	accessTokenQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active)
						 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(accessTokenQuery, token.AccessToken, "access", clientID, userID,
		s.scopesToJSON(scopes), accessExpiresAt, now, true)
	if err != nil {
		return err
	}

	// Store refresh token if present
	if token.RefreshToken != "" {
		refreshExpiresAt := now.Add(30 * 24 * time.Hour) // 30 days

		refreshTokenQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active, parent_token)
							  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = tx.Exec(refreshTokenQuery, token.RefreshToken, "refresh", clientID, userID,
			s.scopesToJSON(scopes), refreshExpiresAt, now, true, token.AccessToken)
		if err != nil {
			return err
		}

		// Update access token to link to refresh token
		updateAccessTokenQuery := `UPDATE oauth2_tokens SET related_token = ? WHERE token = ?`
		_, err = tx.Exec(updateAccessTokenQuery, token.RefreshToken, token.AccessToken)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// storeOAuth2TokensDirect stores access and refresh tokens directly
func (s *OAuth2Service) storeOAuth2TokensDirect(clientID, userID, accessToken, refreshToken string, scopes []string, expiresAt time.Time) error {
	now := time.Now()
	refreshExpiresAt := now.Add(24 * 7 * time.Hour) // Refresh tokens expire in 7 days

	// Start a transaction to ensure both tokens are stored atomically
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Store access token
	accessTokenQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active)
						 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(accessTokenQuery, accessToken, "access", clientID, userID,
		s.scopesToJSON(scopes), expiresAt, now, true)
	if err != nil {
		return err
	}

	// Store refresh token
	refreshTokenQuery := `INSERT INTO oauth2_tokens (token, token_type, client_id, user_id, scopes, expires_at, created_at, is_active, related_token)
						  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(refreshTokenQuery, refreshToken, "refresh", clientID, userID,
		s.scopesToJSON(scopes), refreshExpiresAt, now, true, accessToken)
	if err != nil {
		return err
	}

	// Update access token with related refresh token
	updateQuery := `UPDATE oauth2_tokens SET related_token = ? WHERE token = ?`
	_, err = tx.Exec(updateQuery, refreshToken, accessToken)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// ValidateToken validates an access token and returns user information
func (s *OAuth2Service) ValidateToken(accessToken string) (*models.UserInfo, error) {
	return s.validateAccessToken(accessToken)
}

// validateAccessToken is the shared token validation logic
func (s *OAuth2Service) validateAccessToken(accessToken string) (*models.UserInfo, error) {
	// First check if it's a JWT token (our custom format)
	if s.isJWTToken(accessToken) {
		return s.validateJWTToken(accessToken)
	}

	// Otherwise, check our database for stored tokens
	return s.validateStoredToken(accessToken)
}

// isJWTToken checks if a token is a JWT
func (s *OAuth2Service) isJWTToken(token string) bool {
	return shared.IsJWTToken(token)
}

// validateJWTToken validates a JWT token
func (s *OAuth2Service) validateJWTToken(accessToken string) (*models.UserInfo, error) {
	// This would parse and validate the JWT token
	// For now, return a mock user
	userInfo := &models.UserInfo{
		UserID:    "user_123",
		Email:     "john.doe@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Scopes:    []string{"read:data"},
		ClientID:  "client_123",
	}
	return userInfo, nil
}

// validateStoredToken validates a stored access token using the consolidated table
func (s *OAuth2Service) validateStoredToken(accessToken string) (*models.UserInfo, error) {
	query := `SELECT token, token_type, client_id, user_id, scopes, expires_at, is_active
			  FROM oauth2_tokens WHERE token = ? AND token_type = 'access'`

	var token models.OAuth2Token
	var scopesJSON string

	err := s.db.QueryRow(query, accessToken).Scan(
		&token.Token, &token.TokenType, &token.ClientID, &token.UserID, &scopesJSON, &token.ExpiresAt, &token.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid access token")
		}
		slog.Error("Failed to validate access token", "error", err)
		return nil, fmt.Errorf("failed to validate access token: %w", err)
	}

	token.Scopes = s.jsonToScopes(scopesJSON)

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("access token has expired")
	}

	// Check if token is active
	if !token.IsActive {
		return nil, fmt.Errorf("access token is not active")
	}

	// Get user information
	userInfo, err := s.getUserInfo(token.UserID)
	if err != nil {
		return nil, err
	}

	userInfo.Scopes = token.Scopes
	userInfo.ClientID = token.ClientID

	return userInfo, nil
}

// RevokeToken revokes a token and all related tokens (access/refresh token pairs)
func (s *OAuth2Service) RevokeToken(token string) error {
	// First, find the token and its related tokens
	query := `SELECT token, token_type, related_token, parent_token 
			  FROM oauth2_tokens WHERE token = ? AND is_active = true`

	var tokenInfo struct {
		Token        string
		TokenType    string
		RelatedToken *string
		ParentToken  *string
	}

	err := s.db.QueryRow(query, token).Scan(
		&tokenInfo.Token, &tokenInfo.TokenType, &tokenInfo.RelatedToken, &tokenInfo.ParentToken,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("token not found")
		}
		return err
	}

	// Start a transaction to revoke all related tokens
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Revoke the main token
	revokeQuery := `UPDATE oauth2_tokens SET is_active = false WHERE token = ?`
	_, err = tx.Exec(revokeQuery, token)
	if err != nil {
		return err
	}

	// Revoke related tokens
	if tokenInfo.TokenType == "access" && tokenInfo.RelatedToken != nil {
		// If this is an access token, revoke its refresh token
		_, err = tx.Exec(revokeQuery, *tokenInfo.RelatedToken)
		if err != nil {
			return err
		}
	} else if tokenInfo.TokenType == "refresh" && tokenInfo.ParentToken != nil {
		// If this is a refresh token, revoke its access token
		_, err = tx.Exec(revokeQuery, *tokenInfo.ParentToken)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (s *OAuth2Service) RevokeAllUserTokens(userID string) error {
	query := `UPDATE oauth2_tokens SET is_active = false WHERE user_id = ?`
	_, err := s.db.Exec(query, userID)
	return err
}

// RevokeAllClientTokens revokes all tokens for a specific client
func (s *OAuth2Service) RevokeAllClientTokens(clientID string) error {
	query := `UPDATE oauth2_tokens SET is_active = false WHERE client_id = ?`
	_, err := s.db.Exec(query, clientID)
	return err
}

// GenerateClientCredentialsToken generates an access token for client credentials flow
func (s *OAuth2Service) GenerateClientCredentialsToken(ctx context.Context, clientID string, scopes []string) (*oauth2.Token, error) {
	// Generate access token
	accessToken := s.generateAccessToken()
	expiresAt := time.Now().Add(1 * time.Hour) // Access tokens expire in 1 hour

	// For client credentials flow, we don't have a user ID, so we use a system user ID
	systemUserID := "system_" + clientID

	// Store access token in database
	err := s.storeOAuth2TokensDirect(clientID, systemUserID, accessToken, "", scopes, expiresAt)
	if err != nil {
		slog.Error("Failed to store client credentials token", "error", err, "client_id", clientID)
		return nil, fmt.Errorf("failed to store client credentials token: %w", err)
	}

	// Create oauth2.Token response (no refresh token for client credentials)
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		Expiry:      expiresAt,
	}

	slog.Info("Generated client credentials token", "client_id", clientID, "scopes", scopes)
	return token, nil
}
