package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gov-dx-sandbox/exchange/shared/config"
	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// OwnerIDToEmailMapping represents the mapping from owner_id to owner_email
type OwnerIDToEmailMapping struct {
	OwnerID    string `json:"owner_id"`
	OwnerEmail string `json:"owner_email"`
}

// M2M Token structures
type M2MTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}

type M2MTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// SCIM User structures
type SCIMUser struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
	Emails   []struct {
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
		Type    string `json:"type"`
	} `json:"emails"`
	Schemas []string `json:"schemas"`
	Meta    struct {
		ResourceType string `json:"resourceType"`
	} `json:"meta"`
}

type SCIMResponse struct {
	TotalResults int        `json:"totalResults"`
	ItemsPerPage int        `json:"itemsPerPage"`
	StartIndex   int        `json:"startIndex"`
	Resources    []SCIMUser `json:"Resources"`
	Schemas      []string   `json:"schemas"`
}

// AsgardeoSCIMClient handles SCIM API interactions
type AsgardeoSCIMClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

// NewAsgardeoSCIMClient creates a new SCIM client
func NewAsgardeoSCIMClient(baseURL, clientID, clientSecret string) *AsgardeoSCIMClient {
	return &AsgardeoSCIMClient{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// getM2MToken fetches an M2M access token from Asgardeo
func (c *AsgardeoSCIMClient) getM2MToken() error {
	// Check if we have a valid token
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	tokenURL := fmt.Sprintf("%s/oauth2/token", c.baseURL)

	// Prepare the token request
	tokenReq := M2MTokenRequest{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		GrantType:    "client_credentials",
		Scope:        "internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update",
	}

	// Convert to form data
	formData := url.Values{}
	formData.Set("client_id", tokenReq.ClientID)
	formData.Set("client_secret", tokenReq.ClientSecret)
	formData.Set("grant_type", tokenReq.GrantType)
	formData.Set("scope", tokenReq.Scope)

	// Make the request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp M2MTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	slog.Info("M2M token obtained successfully", "expires_in", tokenResp.ExpiresIn)
	return nil
}

// getUserByNIC fetches user information by NIC using SCIM API
func (c *AsgardeoSCIMClient) getUserByNIC(nic string) (*SCIMUser, error) {
	// Ensure we have a valid token
	if err := c.getM2MToken(); err != nil {
		return nil, fmt.Errorf("failed to get M2M token: %w", err)
	}

	// Try SCIM API first
	user, err := c.getUserByNICSCIM(nic)
	if err == nil {
		return user, nil
	}

	// If SCIM fails, try User Management API as fallback
	slog.Warn("SCIM API failed, trying User Management API", "error", err)
	return c.getUserByNICUserMgmt(nic)
}

// getUserByNICSCIM tries to get user via SCIM API
func (c *AsgardeoSCIMClient) getUserByNICSCIM(nic string) (*SCIMUser, error) {
	// Construct SCIM query URL
	scimURL := fmt.Sprintf("%s/scim2/Users", c.baseURL)

	// Create query parameters for NIC search
	params := url.Values{}
	params.Set("filter", fmt.Sprintf("urn:scim:schemas:extension:custom:User:nic eq \"%s\"", nic))
	params.Set("attributes", "id,userName,emails")

	queryURL := fmt.Sprintf("%s?%s", scimURL, params.Encode())

	// Make the SCIM request
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCIM request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Accept", "application/scim+json")
	req.Header.Set("Content-Type", "application/scim+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make SCIM request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SCIM request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var scimResp SCIMResponse
	if err := json.NewDecoder(resp.Body).Decode(&scimResp); err != nil {
		return nil, fmt.Errorf("failed to decode SCIM response: %w", err)
	}

	if scimResp.TotalResults == 0 {
		return nil, fmt.Errorf("no user found with NIC: %s", nic)
	}

	if len(scimResp.Resources) == 0 {
		return nil, fmt.Errorf("no user resources found for NIC: %s", nic)
	}

	user := scimResp.Resources[0]

	// Validate that the user has an email
	if len(user.Emails) == 0 {
		return nil, fmt.Errorf("user with NIC %s has no email address", nic)
	}

	// Find the primary email or use the first one
	var email string
	for _, e := range user.Emails {
		if e.Primary {
			email = e.Value
			break
		}
	}
	if email == "" {
		email = user.Emails[0].Value
	}

	user.Emails = []struct {
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
		Type    string `json:"type"`
	}{{
		Value:   email,
		Primary: true,
		Type:    "work",
	}}

	slog.Info("User found via SCIM", "nic", nic, "email", email, "username", user.UserName)
	return &user, nil
}

// getUserByNICUserMgmt tries to get user via User Management API
func (c *AsgardeoSCIMClient) getUserByNICUserMgmt(nic string) (*SCIMUser, error) {
	// Construct User Management API query URL
	userMgmtURL := fmt.Sprintf("%s/scim2/Users", c.baseURL)

	// Create query parameters for NIC search
	params := url.Values{}
	params.Set("filter", fmt.Sprintf("urn:scim:schemas:extension:custom:User:nic eq \"%s\"", nic))
	params.Set("attributes", "id,userName,emails")

	queryURL := fmt.Sprintf("%s?%s", userMgmtURL, params.Encode())

	// Make the request
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create User Management request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make User Management request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user management request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var scimResp SCIMResponse
	if err := json.NewDecoder(resp.Body).Decode(&scimResp); err != nil {
		return nil, fmt.Errorf("failed to decode User Management response: %w", err)
	}

	if scimResp.TotalResults == 0 {
		return nil, fmt.Errorf("no user found with NIC: %s", nic)
	}

	if len(scimResp.Resources) == 0 {
		return nil, fmt.Errorf("no user resources found for NIC: %s", nic)
	}

	user := scimResp.Resources[0]

	// Validate that the user has an email
	if len(user.Emails) == 0 {
		return nil, fmt.Errorf("user with NIC %s has no email address", nic)
	}

	// Find the primary email or use the first one
	var email string
	for _, e := range user.Emails {
		if e.Primary {
			email = e.Value
			break
		}
	}
	if email == "" {
		email = user.Emails[0].Value
	}

	user.Emails = []struct {
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
		Type    string `json:"type"`
	}{{
		Value:   email,
		Primary: true,
		Type:    "work",
	}}

	slog.Info("User found via User Management API", "nic", nic, "email", email, "username", user.UserName)
	return &user, nil
}

// Global SCIM client instance with thread-safe initialization
var (
	scimClient *AsgardeoSCIMClient
	scimOnce   sync.Once
)

// getOwnerEmailByID looks up the owner_email for a given owner_id using SCIM API
func getOwnerEmailByID(ownerID string) (string, error) {
	// Initialize SCIM client thread-safely using sync.Once
	scimOnce.Do(func() {
		baseURL := getEnvOrDefault("ASGARDEO_BASE_URL", "https://api.asgardeo.io/t/YOUR_TENANT")
		clientID := getEnvOrDefault("ASGARDEO_M2M_CLIENT_ID", "")
		clientSecret := getEnvOrDefault("ASGARDEO_M2M_CLIENT_SECRET", "")

		if clientID == "" || clientSecret == "" {
			// Fallback to hardcoded mapping for development/testing
			slog.Warn("M2M credentials not configured, using hardcoded mapping", "owner_id", ownerID)
			// Don't initialize scimClient if credentials are missing
			return
		}

		scimClient = NewAsgardeoSCIMClient(baseURL, clientID, clientSecret)
	})

	// If SCIM client is not initialized (missing credentials), use hardcoded mapping
	if scimClient == nil {
		if email, exists := ownerIDToEmailMap[ownerID]; exists {
			return email, nil
		}
		return "", fmt.Errorf("no email mapping found for owner_id: %s", ownerID)
	}

	// Try SCIM lookup first
	user, err := scimClient.getUserByNIC(ownerID)
	if err != nil {
		slog.Warn("SCIM lookup failed, falling back to hardcoded mapping", "owner_id", ownerID, "error", err)
		// Fallback to hardcoded mapping
		if email, exists := ownerIDToEmailMap[ownerID]; exists {
			return email, nil
		}
		return "", fmt.Errorf("no email found for owner_id: %s (SCIM error: %v)", ownerID, err)
	}

	return user.Emails[0].Value, nil
}

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify your frontend domain
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// UserTokenValidationConfig holds configuration for user token validation
type UserTokenValidationConfig struct {
	ExpectedIssuer   string
	ExpectedAudience string
	ExpectedOrgName  string
	RequiredScopes   []string
}

// Context key types to avoid collisions
type contextKey string

const (
	userEmailKey contextKey = "user_email"
	authTypeKey  contextKey = "auth_type"
	tokenInfoKey contextKey = "token_info"
)

// Token type detection
type TokenType string

const (
	TokenTypeUser    TokenType = "user"
	TokenTypeM2M     TokenType = "m2m"
	TokenTypeUnknown TokenType = "unknown"
)

// TokenInfo contains information about a verified token
type TokenInfo struct {
	Type     TokenType
	Subject  string
	Email    string
	ClientID string
	Issuer   string
	Audience []string
	Scopes   []string
	AuthType string
}

// detectTokenType determines if a token is a user token or M2M token based on issuer and claims
func detectTokenType(claims jwt.MapClaims, userIssuer, choreoIssuer string) TokenType {
	iss, ok := claims["iss"].(string)
	if !ok {
		return TokenTypeUnknown
	}

	// Check if it's a Choreo M2M token
	if iss == choreoIssuer {
		// Additional checks for M2M token characteristics
		if _, hasClientID := claims["client_id"]; hasClientID {
			return TokenTypeM2M
		}
		// Check for M2M-specific claims
		if _, hasScope := claims["scope"]; hasScope {
			return TokenTypeM2M
		}
	}

	// Check if it's a user token
	if iss == userIssuer {
		// User tokens typically have email or sub claims
		if _, hasEmail := claims["email"]; hasEmail {
			return TokenTypeUser
		}
		if _, hasSub := claims["sub"]; hasSub {
			return TokenTypeUser
		}
	}

	return TokenTypeUnknown
}

// Hybrid authentication middleware that handles both user JWT and Choreo M2M JWT authentication
func hybridAuthMiddleware(userJWTVerifier *JWTVerifier, choreoJWTVerifier *ChoreoJWTVerifier, engine ConsentEngine, userTokenConfig UserTokenValidationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract consent ID from the URL path
			consentID := strings.TrimPrefix(r.URL.Path, "/consents/")
			if consentID == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
				return
			}

			// Get the consent record to check permissions
			consentRecord, err := engine.GetConsentStatus(consentID)
			if err != nil {
				utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
				return
			}

			// Extract the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Authorization header is required"})
				return
			}

			// Check if it's a Bearer token
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid authorization format. Expected 'Bearer <token>'"})
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
			if tokenString == "" {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Token is required"})
				return
			}

			// First, try to decode the token to determine its type
			parts := strings.Split(tokenString, ".")
			if len(parts) != 3 {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid token format"})
				return
			}

			// Decode claims to determine token type
			claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Failed to decode token claims"})
				return
			}

			var claims jwt.MapClaims
			if err := json.Unmarshal(claimsBytes, &claims); err != nil {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Failed to parse token claims"})
				return
			}

			// Determine token type based on issuer and claims
			userIssuer := userTokenConfig.ExpectedIssuer
			choreoIssuer := getEnvOrDefault("CHOREO_ISSUER", "https://sts.choreo.dev/oauth2/token")
			tokenType := detectTokenType(claims, userIssuer, choreoIssuer)

			if tokenType == TokenTypeUnknown {
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Unknown token type. Token must be from a recognized issuer"})
				return
			}

			// Handle M2M tokens (from Choreo)
			if tokenType == TokenTypeM2M {
				// Verify the token using Choreo JWT verifier
				token, err := choreoJWTVerifier.VerifyToken(tokenString)
				if err != nil {
					slog.Warn("Choreo M2M token verification failed", "error", err, "consent_id", consentID)
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid M2M token"})
					return
				}

				// Extract token information
				tokenInfo, err := extractTokenInfo(token)
				if err != nil {
					slog.Warn("Failed to extract M2M token info", "error", err, "consent_id", consentID)
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid M2M token format"})
					return
				}

				// Add M2M auth type to the request context
				ctx := context.WithValue(r.Context(), authTypeKey, "m2m")
				ctx = context.WithValue(ctx, tokenInfoKey, tokenInfo)
				r = r.WithContext(ctx)

				slog.Info("M2M authentication successful",
					"client_id", tokenInfo.ClientID,
					"consent_id", consentID)

			} else if tokenType == TokenTypeUser {
				// Handle user tokens (from Asgardeo)
				// Verify the token using user JWT verifier
				token, err := userJWTVerifier.VerifyToken(tokenString)
				if err != nil {
					slog.Warn("User token verification failed", "error", err, "consent_id", consentID)
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid user token"})
					return
				}

				// Extract email from token
				email, err := userJWTVerifier.ExtractEmailFromToken(token)
				if err != nil {
					slog.Warn("Failed to extract email from user token", "error", err, "consent_id", consentID)
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Token missing email claim"})
					return
				}

				// Check if the email matches the consent owner email
				if consentRecord.OwnerEmail != email {
					slog.Warn("User email does not match consent owner email",
						"user_email", email,
						"consent_owner_email", consentRecord.OwnerEmail,
						"consent_id", consentID)
					utils.RespondWithJSON(w, http.StatusForbidden, utils.ErrorResponse{Error: "Access denied: email does not match consent owner"})
					return
				}

				// Add user auth type and email to the request context
				ctx := context.WithValue(r.Context(), authTypeKey, "user")
				ctx = context.WithValue(ctx, userEmailKey, email)
				r = r.WithContext(ctx)

				slog.Info("User authentication successful",
					"email", email,
					"consent_id", consentID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTokenInfo extracts token information and determines the token type
func extractTokenInfo(token *jwt.Token) (*TokenInfo, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	tokenInfo := &TokenInfo{}

	// Extract basic claims
	if sub, ok := claims["sub"].(string); ok {
		tokenInfo.Subject = sub
	}

	if aud, ok := claims["aud"]; ok {
		switch v := aud.(type) {
		case string:
			tokenInfo.Audience = []string{v}
		case []interface{}:
			for _, a := range v {
				if s, ok := a.(string); ok {
					tokenInfo.Audience = append(tokenInfo.Audience, s)
				}
			}
		}
	}

	if iss, ok := claims["iss"].(string); ok {
		tokenInfo.Issuer = iss
	}

	if clientID, ok := claims["client_id"].(string); ok {
		tokenInfo.ClientID = clientID
	}

	// Extract scopes
	if scope, ok := claims["scope"].(string); ok {
		tokenInfo.Scopes = strings.Fields(scope)
	}

	// Extract email
	emailFields := []string{"email", "preferred_username"}
	for _, field := range emailFields {
		if email, ok := claims[field].(string); ok && email != "" {
			tokenInfo.Email = email
			break
		}
	}

	// Extract auth type
	if authType, ok := claims["aut"].(string); ok {
		tokenInfo.AuthType = authType
	}

	// Determine token type based on claims
	tokenInfo.Type = determineTokenType(claims)

	return tokenInfo, nil
}

// determineTokenType determines if a token is a user token or M2M token
func determineTokenType(claims jwt.MapClaims) TokenType {
	// Check for auth type first
	if authType, ok := claims["aut"].(string); ok {
		if authType == "APPLICATION_USER" {
			return TokenTypeUser
		}
		if authType == "APPLICATION" {
			return TokenTypeM2M
		}
	}

	// Fallback to legacy logic
	clientID, hasClientID := claims["client_id"].(string)
	sub, hasSub := claims["sub"].(string)
	scope, hasScope := claims["scope"].(string)

	// If it has client_id and either no sub or sub matches client_id, it's likely M2M
	if hasClientID && (!hasSub || sub == clientID) {
		// Additional check: M2M tokens usually have scopes like "consent:read consent:write"
		if hasScope && strings.Contains(scope, "consent:") {
			return TokenTypeM2M
		}
	}

	// If it has a sub claim that doesn't match client_id, it's likely a user token
	if hasSub && (clientID == "" || sub != clientID) {
		return TokenTypeUser
	}

	// If it has an email claim, it's likely a user token
	if email, ok := claims["email"].(string); ok && email != "" {
		return TokenTypeUser
	}

	// Default to user token if we can't determine
	return TokenTypeUser
}

// apiServer holds dependencies for the HTTP handlers
type apiServer struct {
	engine ConsentEngine
}

// ConsentPortalCreateRequest represents the request format for creating consent via portal
type ConsentPortalCreateRequest struct {
	AppID      string      `json:"app_id"`
	DataFields []DataField `json:"data_fields"`
	Purpose    string      `json:"purpose"`
	SessionID  string      `json:"session_id"`
}

// ConsentPortalUpdateRequest represents the request format for updating consent via portal
type ConsentPortalUpdateRequest struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updated_by"`
	Reason    string `json:"reason,omitempty"`
}

// Consent handlers - organized for better readability
func (s *apiServer) handleConsentPost(w http.ResponseWriter, r *http.Request) {
	// POST /consents should only create new consent records
	// The engine will handle reuse logic internally
	s.createConsent(w, r)
}

func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	var req ConsentRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate that all required fields are present and not empty
	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if req.SessionID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "session_id is required and cannot be empty"})
		return
	}
	if len(req.DataFields) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "data_fields is required and cannot be empty"})
		return
	}

	// Validate each data field and populate owner_email from owner_id mapping
	for i, dataField := range req.DataFields {
		if dataField.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if len(dataField.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields is required and cannot be empty", i)})
			return
		}
		// Validate that no field is empty
		for j, field := range dataField.Fields {
			if field == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields[%d] cannot be empty", i, j)})
				return
			}
		}

		// Look up owner_email from owner_id mapping
		ownerEmail, err := getOwnerEmailByID(dataField.OwnerID)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_id '%s' not found in mapping: %v", i, dataField.OwnerID, err)})
			return
		}

		// Set the owner_email in the data field
		req.DataFields[i].OwnerEmail = ownerEmail
	}

	// Process consent request using the engine
	response, err := s.engine.ProcessConsentRequest(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "create consent", "id", response.ConsentID, "owner", response.OwnerID, "existing", false)

	// Return the ConsentRecord
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (s *apiServer) updateConsent(w http.ResponseWriter, r *http.Request) {
	var req UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.UpdateConsent(id, req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentUpdateFailed+": %w", err)
		}

		// Return the ConsentRecord directly
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) revokeConsent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Reason string }
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.RevokeConsent(id, req.Reason)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentRevokeFailed+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

// revokeConsentByID handles DELETE /consents/{id} - revoke consent by ID
func (s *apiServer) revokeConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct{ Reason string }

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	record, err := s.engine.RevokeConsent(consentID, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to revoke consent: " + err.Error()})
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, record)
}

// patchConsentByID handles PATCH /consents/{id} - partial update of consent resource
func (s *apiServer) patchConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status        string   `json:"status,omitempty"`
		UpdatedBy     string   `json:"updated_by,omitempty"`
		Reason        string   `json:"reason,omitempty"`
		GrantDuration string   `json:"grant_duration,omitempty"`
		Fields        []string `json:"fields,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Get the existing record first
	existingRecord, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to get consent record: " + err.Error()})
		}
		return
	}

	// Apply partial updates
	updateReq := UpdateConsentRequest{
		Status:    ConsentStatus(existingRecord.Status), // Keep existing status by default
		UpdatedBy: existingRecord.OwnerID,               // Keep existing updated_by by default
		Reason:    "",                                   // Will be set if provided
	}

	// Update only provided fields
	if req.Status != "" {
		updateReq.Status = ConsentStatus(req.Status)
	}
	if req.UpdatedBy != "" {
		updateReq.UpdatedBy = req.UpdatedBy
	}
	if req.Reason != "" {
		updateReq.Reason = req.Reason
	}
	if req.GrantDuration != "" {
		updateReq.GrantDuration = req.GrantDuration
	}
	if len(req.Fields) > 0 {
		updateReq.Fields = req.Fields
	}

	// Update the record
	updatedRecord, err := s.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record: " + err.Error()})
		}
		return
	}
	slog.Info("Consent record updated", "consent_id", updatedRecord.ConsentID, "owner_id", updatedRecord.OwnerID, "owner_email", updatedRecord.OwnerEmail, "app_id", updatedRecord.AppID, "status", updatedRecord.Status, "type", updatedRecord.Type, "created_at", updatedRecord.CreatedAt, "updated_at", updatedRecord.UpdatedAt, "expires_at", updatedRecord.ExpiresAt, "grant_duration", updatedRecord.GrantDuration, "fields", updatedRecord.Fields, "session_id", updatedRecord.SessionID, "consent_portal_url", updatedRecord.ConsentPortalURL)
	utils.RespondWithJSON(w, http.StatusOK, updatedRecord)
}

func (s *apiServer) getConsentPortalInfo(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		consentID, err := utils.ExtractQueryParam(r, "consent_id")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.GetConsentStatus(consentID)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(ErrConsentNotFound+": %w", err)
		}

		// Convert to the user-facing ConsentPortalView
		consentView := record.ToConsentPortalView()

		return consentView, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByDataOwner(dataOwner)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"owner_id": dataOwner,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consumer/", func(consumer string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByConsumer(consumer)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"consumer": consumer,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (s *apiServer) checkConsentExpiry(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		expiredRecords, err := s.engine.CheckConsentExpiry()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentExpiryFailed+": %w", err)
		}

		// Log the operation
		slog.Info("Operation successful",
			"operation", OpCheckConsentExpiry,
			"expired_count", len(expiredRecords),
		)

		// Ensure expired_records is always an array, never null
		expiredRecordsList := make([]*ConsentRecord, 0)
		if expiredRecords != nil {
			expiredRecordsList = expiredRecords
		}

		return map[string]interface{}{
			"expired_records": expiredRecordsList,
			"count":           len(expiredRecordsList),
			"checked_at":      time.Now(),
		}, http.StatusOK, nil
	})
}

// Route handlers - organized for better readability
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	switch {
	case path == "" && r.Method == http.MethodPost:
		// POST /consents - create new consent record
		s.handleConsentPost(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		// GET /consents/{id} - get consent by ID
		consentID := strings.TrimPrefix(path, "/")
		s.getConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		// PUT /consents/{id} - replace entire consent resource (idempotent)
		consentID := strings.TrimPrefix(path, "/")
		s.updateConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPatch:
		// PATCH /consents/{id} - partial update of consent resource
		consentID := strings.TrimPrefix(path, "/")
		s.patchConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		// DELETE /consents/{id} - revoke consent
		consentID := strings.TrimPrefix(path, "/")
		s.revokeConsentByID(w, r, consentID)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) consentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.processConsentPortalRequest(w, r)
	case http.MethodPut:
		s.processConsentPortalUpdate(w, r)
	case http.MethodGet:
		s.getConsentPortalInfo(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// processConsentPortalRequest handles POST requests to the consent portal
func (s *apiServer) processConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentPortalCreateRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate required fields
	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if req.Purpose == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "purpose is required and cannot be empty"})
		return
	}
	if req.SessionID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "session_id is required and cannot be empty"})
		return
	}
	if len(req.DataFields) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "data_fields is required and cannot be empty"})
		return
	}

	// Validate each data field
	for i, dataField := range req.DataFields {
		if dataField.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if dataField.OwnerEmail == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_email is required and cannot be empty", i)})
			return
		}
		if len(dataField.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields is required and cannot be empty", i)})
			return
		}
		// Validate that no field is empty
		for j, field := range dataField.Fields {
			if field == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields[%d] cannot be empty", i, j)})
				return
			}
		}
	}

	// Convert to ConsentRequest format
	consentReq := ConsentRequest{
		AppID:      req.AppID,
		DataFields: req.DataFields,
		SessionID:  req.SessionID,
	}

	// Process consent request using the engine
	response, err := s.engine.ProcessConsentRequest(consentReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "create consent via portal", "id", response.ConsentID, "owner", response.OwnerID, "existing", false)

	// Return the ConsentRecord
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

// processConsentPortalUpdate handles PUT requests to the consent portal
func (s *apiServer) processConsentPortalUpdate(w http.ResponseWriter, r *http.Request) {
	// Extract consent ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/consent-portal/")
	if path == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	var req ConsentPortalUpdateRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate required fields for update
	if req.Status == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status is required and cannot be empty"})
		return
	}
	if req.UpdatedBy == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "updated_by is required and cannot be empty"})
		return
	}

	// Convert to UpdateConsentRequest format
	updateReq := UpdateConsentRequest{
		Status:    ConsentStatus(req.Status),
		UpdatedBy: req.UpdatedBy,
		Reason:    req.Reason,
	}

	// Process consent update using the engine
	response, err := s.engine.UpdateConsent(path, updateReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "update consent via portal", "id", response.ConsentID, "status", response.Status, "updated_by", req.UpdatedBy)

	// Return the updated ConsentRecord
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *apiServer) dataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByDataOwner(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) consumerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByConsumer(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) getConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Convert to the user-facing ConsentPortalView
	consentView := record.ToConsentPortalView()

	// Return only the UI-necessary fields
	utils.RespondWithJSON(w, http.StatusOK, consentView)
}

func (s *apiServer) getDataInfo(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Return only owner_id and owner_email
	dataInfo := map[string]interface{}{
		"owner_id":    record.OwnerID,
		"owner_email": record.OwnerEmail,
	}

	utils.RespondWithJSON(w, http.StatusOK, dataInfo)
}

func (s *apiServer) updateConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status        string `json:"status"`
		UpdatedBy     string `json:"updated_by,omitempty"`
		GrantDuration string `json:"grant_duration,omitempty"`
		Reason        string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Get the existing consent record to extract owner information
	existingRecord, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to get consent record: " + err.Error()})
		}
		return
	}

	// Validate status if provided
	var newStatus ConsentStatus
	if req.Status != "" {
		// Validate that the status is one of the valid consent statuses
		validStatuses := []string{"pending", "approved", "rejected", "expired", "revoked"}
		isValid := false
		for _, validStatus := range validStatuses {
			if req.Status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status must be one of: pending, approved, rejected, expired, revoked"})
			return
		}
		newStatus = ConsentStatus(req.Status)
	} else {
		// Keep existing status if not provided
		newStatus = ConsentStatus(existingRecord.Status)
	}

	// Set default reason if not provided
	reason := req.Reason
	if reason == "" {
		switch newStatus {
		case StatusApproved:
			reason = "Consent approved via API"
		case StatusRejected:
			reason = "Consent rejected via API"
		case StatusExpired:
			reason = "Consent expired via API"
		case StatusRevoked:
			reason = "Consent revoked via API"
		case StatusPending:
			reason = "Consent reset to pending via API"
		default:
			reason = "Consent updated via API"
		}
	}

	// Update the record
	updateReq := UpdateConsentRequest{
		Status:        newStatus,
		UpdatedBy:     existingRecord.OwnerID, // Use existing owner ID
		Reason:        reason,
		GrantDuration: req.GrantDuration, // Will be empty string if not provided
	}

	updatedRecord, err := s.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record: " + err.Error()})
		}
		return
	}

	// Log the operation
	slog.Info("Consent updated via PUT", "consentId", consentID, "status", string(newStatus), "ownerId", existingRecord.OwnerID, "grantDuration", req.GrantDuration)

	utils.RespondWithJSON(w, http.StatusOK, updatedRecord)
}

func (s *apiServer) consentWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the consent website HTML file
	http.ServeFile(w, r, "consent-website.html")
}

func (s *apiServer) dataInfoHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/data-info/")
	if path == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	if r.Method == http.MethodGet {
		s.getDataInfo(w, r, path)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) adminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		s.checkConsentExpiry(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}
func main() {
	// Load configuration using flags
	cfg := config.LoadConfig("consent-engine")

	// Setup logging
	utils.SetupLogging(cfg.Logging.Format, cfg.Logging.Level)

	slog.Info("Starting consent engine",
		"environment", cfg.Environment,
		"port", cfg.Service.Port,
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	// Initialize database connection
	dbConfig := NewDatabaseConfig()
	db, err := ConnectDB(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize database tables
	if err := InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize consent engine
	consentPortalUrl := getEnvOrDefault("CONSENT_PORTAL_URL", "http://localhost:5173")

	slog.Info("Using consent portal URL", "url", consentPortalUrl)

	engine := NewPostgresConsentEngine(db, consentPortalUrl)
	server := &apiServer{engine: engine}

	// Start background expiry process with context cancellation
	expiryInterval := getEnvOrDefault("CONSENT_EXPIRY_CHECK_INTERVAL", "24h")
	interval, err := time.ParseDuration(expiryInterval)
	if err != nil {
		slog.Warn("Invalid expiry check interval, using default 24h", "interval", expiryInterval, "error", err)
		interval = 24 * time.Hour
	}

	// Create a context for the background process that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This will cancel the background process when main exits

	engine.StartBackgroundExpiryProcess(ctx, interval)

	// Initialize user JWT verifier with Asgardeo JWKS endpoint
	userJwksURL := getEnvOrDefault("ASGARDEO_JWKS_URL", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/jwks")
	userIssuer := getEnvOrDefault("ASGARDEO_ISSUER", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token")
	userAudience := getEnvOrDefault("ASGARDEO_AUDIENCE", "YOUR_AUDIENCE")
	userJWTVerifier := NewJWTVerifier(userJwksURL, userIssuer, userAudience)
	slog.Info("Initialized user JWT verifier", "jwks_url", userJwksURL, "issuer", userIssuer, "audience", userAudience)

	// Initialize Choreo M2M JWT verifier
	choreoJwksURL := getEnvOrDefault("CHOREO_JWKS_URL", "https://sts.choreo.dev/oauth2/jwks")
	choreoIssuer := getEnvOrDefault("CHOREO_ISSUER", "https://sts.choreo.dev/oauth2/token")
	choreoAudience := getEnvOrDefault("CHOREO_AUDIENCE", "YOUR_CHOREO_AUDIENCE")
	choreoJWTVerifier := NewChoreoJWTVerifier(choreoJwksURL, choreoIssuer, choreoAudience)
	slog.Info("Initialized Choreo M2M JWT verifier", "jwks_url", choreoJwksURL, "issuer", choreoIssuer, "audience", choreoAudience)

	// Configure user token validation
	userTokenConfig := UserTokenValidationConfig{
		ExpectedIssuer:   getEnvOrDefault("ASGARDEO_ISSUER", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token"),
		ExpectedAudience: getEnvOrDefault("ASGARDEO_AUDIENCE", "YOUR_AUDIENCE"),
		ExpectedOrgName:  getEnvOrDefault("ASGARDEO_ORG_NAME", "YOUR_ORG_NAME"),
		RequiredScopes:   []string{}, // No required scopes for basic consent access
	}

	// Setup routes using utils
	mux := http.NewServeMux()

	// Routes that don't require authentication
	mux.Handle("/consents", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consent-portal/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentPortalHandler)))
	mux.Handle("/consent-website", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentWebsiteHandler)))
	mux.Handle("/data-owner/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.dataOwnerHandler)))
	mux.Handle("/consumer/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consumerHandler)))
	mux.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.adminHandler)))
	mux.Handle("/data-info/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.dataInfoHandler)))
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("consent-engine")))

	// Routes that require hybrid authentication (both user and M2M tokens)
	mux.Handle("/consents/", utils.PanicRecoveryMiddleware(hybridAuthMiddleware(userJWTVerifier, choreoJWTVerifier, engine, userTokenConfig)(http.HandlerFunc(server.consentHandler))))

	// Create server using utils
	serverConfig := &utils.ServerConfig{
		Port:         cfg.Service.Port,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  60 * time.Second,
	}

	// Wrap the mux with CORS middleware
	handler := corsMiddleware(mux)
	httpServer := utils.CreateServer(serverConfig, handler)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(httpServer, "consent-engine"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
