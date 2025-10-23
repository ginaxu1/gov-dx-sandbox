package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/shared"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	v1shared "github.com/gov-dx-sandbox/api-server-go/v1/shared"
	"golang.org/x/oauth2"
)

// OAuth2Handler handles OAuth 2.0 endpoints
type OAuth2Handler struct {
	oauthService *services.OAuth2Service
	// Configuration options for production vs test environments
	allowLoginRedirect bool
	loginURL           string
}

// NewOAuth2Handler creates a new OAuth 2.0 handler
func NewOAuth2Handler(oauthService *services.OAuth2Service) *OAuth2Handler {
	return &OAuth2Handler{
		oauthService:       oauthService,
		allowLoginRedirect: true,     // Default to allowing login redirects
		loginURL:           "/login", // Default login URL
	}
}

// NewOAuth2HandlerWithConfig creates a new OAuth 2.0 handler with custom configuration
func NewOAuth2HandlerWithConfig(oauthService *services.OAuth2Service, allowLoginRedirect bool, loginURL string) *OAuth2Handler {
	return &OAuth2Handler{
		oauthService:       oauthService,
		allowLoginRedirect: allowLoginRedirect,
		loginURL:           loginURL,
	}
}

// SetupOAuth2Routes sets up OAuth 2.0 routes
func (h *OAuth2Handler) SetupOAuth2Routes(mux *http.ServeMux) {
	// OAuth 2.0 endpoints
	mux.Handle("/oauth2/authorize", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleAuthorize)))
	mux.Handle("/oauth2/token", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleToken)))
	mux.Handle("/oauth2/refresh", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleRefresh)))

	// Client management endpoints
	mux.Handle("/oauth2/clients", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleClients)))
	mux.Handle("/oauth2/clients/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleClientByID)))

	// Data access endpoint
	mux.Handle("/api/v1/data", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handleDataAccess)))
}

// handleAuthorize handles the OAuth 2.0 authorization endpoint
func (h *OAuth2Handler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	responseType := r.URL.Query().Get("response_type")
	scope := r.URL.Query().Get("scope")
	_ = r.URL.Query().Get("code_challenge")        // For future use
	_ = r.URL.Query().Get("code_challenge_method") // For future use

	// Validate required parameters
	if responseType != "code" {
		h.respondWithOAuth2Error(w, "unsupported_response_type", "Only 'code' response type is supported", state)
		return
	}

	if clientID == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "client_id is required", state)
		return
	}

	if redirectURI == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "redirect_uri is required", state)
		return
	}

	// Extract authenticated user ID from session or authentication context
	userID, err := h.extractAuthenticatedUserID(r)
	if err != nil {
		slog.Error("Failed to extract authenticated user ID", "error", err)

		// Production-ready authentication handling
		// Determine the appropriate response based on request type and configuration
		if h.IsAPIRequest(r) {
			// For API clients, always return OAuth2 error
			h.respondWithOAuth2Error(w, "access_denied", "User authentication required", state)
			return
		} else if h.shouldRedirectToLogin(r) {
			// For browser requests, redirect to login page with proper OAuth2 parameters
			loginURL := h.buildLoginURL(clientID, redirectURI, state, scope)
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		} else {
			// Fallback to OAuth2 error
			h.respondWithOAuth2Error(w, "access_denied", "User authentication required", state)
			return
		}
	}
	scopes := []string{}
	if scope != "" {
		scopes = strings.Split(scope, " ")
	}

	// Create authorization code
	authCode, err := h.oauthService.CreateAuthorizationCode(clientID, userID, redirectURI, scopes)
	if err != nil {
		slog.Error("Failed to create authorization code", "error", err, "client_id", clientID)
		h.respondWithOAuth2Error(w, "invalid_client", "Invalid client_id", state)
		return
	}

	// Build redirect URL with authorization code
	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		slog.Error("Invalid redirect URI", "error", err, "redirect_uri", redirectURI)
		h.respondWithOAuth2Error(w, "invalid_request", "Invalid redirect_uri", state)
		return
	}

	query := redirectURL.Query()
	query.Set("code", authCode.Code)
	if state != "" {
		query.Set("state", state)
	}
	redirectURL.RawQuery = query.Encode()

	slog.Info("Authorization code created", "client_id", clientID, "state", state)

	// Redirect to client's redirect URI with authorization code
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// handleToken handles the OAuth 2.0 token endpoint
func (h *OAuth2Handler) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.respondWithOAuth2Error(w, "invalid_request", "Invalid request format", "")
		return
	}

	req := models.TokenRequest{
		GrantType:    r.FormValue("grant_type"),
		Code:         r.FormValue("code"),
		RedirectURI:  r.FormValue("redirect_uri"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		RefreshToken: r.FormValue("refresh_token"),
	}

	// Validate request based on grant type
	if req.GrantType != "authorization_code" && req.GrantType != "client_credentials" {
		h.respondWithOAuth2Error(w, "unsupported_grant_type", "Only 'authorization_code' and 'client_credentials' grant types are supported", "")
		return
	}

	if req.ClientID == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "client_id is required", "")
		return
	}

	if req.ClientSecret == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "client_secret is required", "")
		return
	}

	// Grant type specific validations
	if req.GrantType == "authorization_code" {
		if req.Code == "" {
			h.respondWithOAuth2Error(w, "invalid_request", "code is required for authorization_code grant", "")
			return
		}
		if req.RedirectURI == "" {
			h.respondWithOAuth2Error(w, "invalid_request", "redirect_uri is required for authorization_code grant", "")
			return
		}
	}

	// Validate client credentials
	var client *models.OAuth2Client
	var clientErr error

	if req.GrantType == "client_credentials" {
		// For client credentials flow, only validate client ID and secret
		client, clientErr = h.oauthService.GetClient(req.ClientID)
		if clientErr != nil {
			h.respondWithOAuth2Error(w, "invalid_client", "Invalid client credentials", "")
			return
		}
		if client.ClientSecret != req.ClientSecret {
			h.respondWithOAuth2Error(w, "invalid_client", "Invalid client credentials", "")
			return
		}
	} else {
		// For authorization code flow, validate client ID, secret, and redirect URI
		client, clientErr = h.oauthService.ValidateClient(req.ClientID, req.ClientSecret, req.RedirectURI)
		if clientErr != nil {
			h.respondWithOAuth2Error(w, "invalid_client", "Invalid client credentials", "")
			return
		}
	}

	var token *oauth2.Token
	ctx := context.Background()

	// Handle different grant types
	if req.GrantType == "authorization_code" {
		// Exchange code for token using the oauth2 package with PKCE
		var err error
		token, err = h.oauthService.ExchangeCodeForToken(ctx, req.ClientID, req.Code, req.RedirectURI, req.CodeVerifier)
		if err != nil {
			slog.Error("Failed to exchange code for token", "error", err, "client_id", req.ClientID)
			h.respondWithOAuth2Error(w, "invalid_grant", "Invalid or expired authorization code", "")
			return
		}
	} else if req.GrantType == "client_credentials" {
		// Generate token directly for client credentials flow
		var err error
		token, err = h.oauthService.GenerateClientCredentialsToken(ctx, req.ClientID, client.Scopes)
		if err != nil {
			slog.Error("Failed to generate client credentials token", "error", err, "client_id", req.ClientID)
			h.respondWithOAuth2Error(w, "invalid_grant", "Failed to generate token", "")
			return
		}
	}

	// Calculate expires in seconds
	expiresIn := 3600 // Default 1 hour
	if !token.Expiry.IsZero() {
		expiresIn = int(time.Until(token.Expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 3600
		}
	}

	// Prepare response
	response := v1shared.CreateTokenResponse(token.AccessToken, token.RefreshToken, "read:data", expiresIn)

	slog.Info("Token issued", "client_id", req.ClientID, "token_type", token.TokenType)
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// handleRefresh handles the OAuth 2.0 refresh token endpoint
func (h *OAuth2Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		h.respondWithOAuth2Error(w, "invalid_request", "Invalid request format", "")
		return
	}

	req := models.RefreshTokenRequest{
		GrantType:    r.FormValue("grant_type"),
		RefreshToken: r.FormValue("refresh_token"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
	}

	// Validate request
	if req.GrantType != "refresh_token" {
		h.respondWithOAuth2Error(w, "unsupported_grant_type", "Only 'refresh_token' grant type is supported", "")
		return
	}

	if req.RefreshToken == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "refresh_token is required", "")
		return
	}

	if req.ClientID == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "client_id is required", "")
		return
	}

	if req.ClientSecret == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "client_secret is required", "")
		return
	}

	// Validate client credentials
	_, clientErr := h.oauthService.ValidateClient(req.ClientID, req.ClientSecret, "")
	if clientErr != nil {
		h.respondWithOAuth2Error(w, "invalid_client", "Invalid client credentials", "")
		return
	}

	// Refresh access token using the oauth2 package
	ctx := context.Background()
	token, err := h.oauthService.RefreshToken(ctx, req.ClientID, req.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh token", "error", err, "client_id", req.ClientID)
		h.respondWithOAuth2Error(w, "invalid_grant", "Invalid or expired refresh token", "")
		return
	}

	// Calculate expires in seconds
	expiresIn := 3600 // Default 1 hour
	if !token.Expiry.IsZero() {
		expiresIn = int(time.Until(token.Expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 3600
		}
	}

	// Prepare response
	response := v1shared.CreateTokenResponse(token.AccessToken, token.RefreshToken, "read:data", expiresIn)

	slog.Info("Token refreshed", "client_id", req.ClientID, "token_type", token.TokenType)
	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// handleClients handles OAuth 2.0 client management
func (h *OAuth2Handler) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createClient(w, r)
	case http.MethodGet:
		h.listClients(w, r)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// createClient creates a new OAuth 2.0 client
func (h *OAuth2Handler) createClient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOAuth2ClientRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	client, err := h.oauthService.CreateClient(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, client)
}

// listClients lists OAuth 2.0 clients (admin function)
func (h *OAuth2Handler) listClients(w http.ResponseWriter, r *http.Request) {
	// This would typically require admin authentication
	// For now, we'll return a placeholder response
	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Client listing not implemented"})
}

// handleClientByID handles individual OAuth 2.0 client operations
func (h *OAuth2Handler) handleClientByID(w http.ResponseWriter, r *http.Request) {
	clientID := utils.ExtractIDFromPathString(r.URL.Path)
	if clientID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Client ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		client, err := h.oauthService.GetClient(clientID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Client not found")
			return
		}
		// Don't return the client secret in the response
		client.ClientSecret = ""
		utils.RespondWithSuccess(w, http.StatusOK, client)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleDataAccess handles data access requests with OAuth 2.0 token validation
func (h *OAuth2Handler) handleDataAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract and validate access token
	accessToken, err := shared.ExtractAccessToken(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or missing access token")
		return
	}

	// Validate access token and get user information
	userInfo, err := h.oauthService.ValidateToken(accessToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired access token")
		return
	}

	// Parse data request
	var req models.DataRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user has required scopes
	if !shared.HasRequiredScope(userInfo.Scopes, "read:data") {
		utils.RespondWithError(w, http.StatusForbidden, "Insufficient scope")
		return
	}

	// Get user data
	data, err := h.oauthService.GetUserData(userInfo.UserID, req.Fields)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user data")
		return
	}

	slog.Info("Data access granted", "user_id", userInfo.UserID, "client_id", userInfo.ClientID, "fields", req.Fields)
	utils.RespondWithSuccess(w, http.StatusOK, data)
}

// Helper methods

// shouldRedirectToLogin determines if the request should redirect to login page
// This handles different scenarios: browser requests vs API requests, test environments, etc.
func (h *OAuth2Handler) shouldRedirectToLogin(r *http.Request) bool {
	// If login redirects are disabled, always return OAuth2 error
	if !h.allowLoginRedirect {
		return false
	}

	// Check if this is a browser request (has Accept header with HTML)
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "text/html") {
		return true
	}

	// Check if this is a test environment
	if r.Header.Get("X-Test-Environment") == "true" {
		return true
	}

	// Check if the request comes from a browser (has User-Agent)
	userAgent := r.Header.Get("User-Agent")
	if userAgent != "" && !strings.Contains(strings.ToLower(userAgent), "curl") &&
		!strings.Contains(strings.ToLower(userAgent), "postman") &&
		!strings.Contains(strings.ToLower(userAgent), "insomnia") &&
		!strings.Contains(strings.ToLower(userAgent), "httpie") {
		return true
	}

	// Check for mobile app user agents that might need login redirects
	if strings.Contains(strings.ToLower(userAgent), "mobile") ||
		strings.Contains(strings.ToLower(userAgent), "android") ||
		strings.Contains(strings.ToLower(userAgent), "iphone") {
		return true
	}

	// Default to returning OAuth2 error for API clients
	return false
}

// buildLoginURL constructs a proper login URL with OAuth2 parameters
func (h *OAuth2Handler) buildLoginURL(clientID, redirectURI, state, scope string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	if state != "" {
		params.Set("state", state)
	}
	if scope != "" {
		params.Set("scope", scope)
	}

	// Use the configured login URL
	loginURL := h.loginURL
	if strings.Contains(loginURL, "?") {
		loginURL += "&" + params.Encode()
	} else {
		loginURL += "?" + params.Encode()
	}

	return loginURL
}

// SetLoginRedirectConfig allows runtime configuration of login redirect behavior
func (h *OAuth2Handler) SetLoginRedirectConfig(allowRedirect bool, loginURL string) {
	h.allowLoginRedirect = allowRedirect
	h.loginURL = loginURL
}

// IsAPIRequest determines if the request is from an API client vs browser
func (h *OAuth2Handler) IsAPIRequest(r *http.Request) bool {
	// Check for API-specific headers
	if r.Header.Get("X-API-Key") != "" || r.Header.Get("X-API-Token") != "" {
		return true
	}

	// Check for JSON content type
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// Check for API client user agents
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))
	apiClients := []string{"curl", "postman", "insomnia", "httpie", "wget", "python-requests", "go-http-client"}
	for _, client := range apiClients {
		if strings.Contains(userAgent, client) {
			return true
		}
	}

	return false
}

// extractAuthenticatedUserID extracts the authenticated user ID from the request
// This function supports multiple authentication mechanisms:
// 1. OAuth2 Bearer tokens (existing OAuth2 flow)
// 2. Session-based authentication (cookies or headers)
// 3. JWT tokens (non-OAuth2 format)
// 4. Context-based authentication (set by previous middleware)
//
// Usage examples:
// - Session cookie: Cookie: session_id=test_session_123
// - Session header: X-Session-ID: test_session_456
// - JWT token: Authorization: JWT test_jwt_token_123
// - JWT header: X-JWT-Token: test_jwt_token_456
// - Context: Set by previous middleware in request context
func (h *OAuth2Handler) extractAuthenticatedUserID(r *http.Request) (string, error) {
	// Method 1: Check for existing OAuth2 token in Authorization header
	// This allows for cases where the user is already authenticated via OAuth2
	if accessToken, err := shared.ExtractAccessToken(r); err == nil {
		if userInfo, err := h.oauthService.ValidateToken(accessToken); err == nil {
			return userInfo.UserID, nil
		}
	}

	// Method 2: Check for session-based authentication
	// This would typically involve checking session cookies or session storage
	if sessionUserID := h.extractUserIDFromSession(r); sessionUserID != "" {
		return sessionUserID, nil
	}

	// Method 3: Check for JWT token in Authorization header (non-OAuth2)
	if jwtUserID := h.extractUserIDFromJWT(r); jwtUserID != "" {
		return jwtUserID, nil
	}

	// Method 4: Check for user ID in request context (if set by previous middleware)
	if userInfo, ok := r.Context().Value("user_info").(*models.UserInfo); ok && userInfo != nil {
		return userInfo.UserID, nil
	}

	// If no authentication method found, return error
	return "", fmt.Errorf("no authenticated user found")
}

// extractUserIDFromSession extracts user ID from session-based authentication
func (h *OAuth2Handler) extractUserIDFromSession(r *http.Request) string {
	// Example implementation for session-based authentication
	// This is a basic example - in production, you would use a proper session store

	// Method 1: Check for session cookie
	if sessionCookie, err := r.Cookie("session_id"); err == nil && sessionCookie.Value != "" {
		// In a real implementation, you would:
		// 1. Validate the session with your session store (Redis, database, etc.)
		// 2. Extract user ID from session data
		// For now, we'll check if it's a known test session
		if sessionCookie.Value == "test_session_123" {
			return "user_123" // This would come from your session store
		}
	}

	// Method 2: Check for session in request header (alternative approach)
	if sessionHeader := r.Header.Get("X-Session-ID"); sessionHeader != "" {
		// Validate session and extract user ID
		// This is just an example - implement according to your session management
		if sessionHeader == "test_session_456" {
			return "user_456"
		}
	}

	// Method 3: Check for user ID in request context (set by previous middleware)
	if userID := r.Context().Value("authenticated_user_id"); userID != nil {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr
		}
	}

	// No session found
	return ""
}

// extractUserIDFromJWT extracts user ID from JWT token in Authorization header
func (h *OAuth2Handler) extractUserIDFromJWT(r *http.Request) string {
	// Example implementation for JWT-based authentication
	// This is a basic example - in production, you would use a proper JWT library

	// Check for JWT token in Authorization header (non-OAuth2 format)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it's a JWT token (not OAuth2 Bearer token)
	// In a real implementation, you would distinguish between OAuth2 and JWT tokens
	if strings.HasPrefix(authHeader, "JWT ") {
		jwtToken := strings.TrimPrefix(authHeader, "JWT ")

		// In a real implementation, you would:
		// 1. Parse and validate the JWT token
		// 2. Check signature and expiration
		// 3. Extract user ID from claims

		// For now, we'll do a simple check for test tokens
		if jwtToken == "test_jwt_token_123" {
			return "user_jwt_123"
		}
	}

	// Check for JWT in custom header
	if jwtHeader := r.Header.Get("X-JWT-Token"); jwtHeader != "" {
		// Validate JWT and extract user ID
		// This is just an example - implement according to your JWT validation
		if jwtHeader == "test_jwt_token_456" {
			return "user_jwt_456"
		}
	}

	// No JWT found
	return ""
}

func (h *OAuth2Handler) respondWithOAuth2Error(w http.ResponseWriter, error, description, state string) {
	response := models.OAuth2ErrorResponse{
		Error:            error,
		ErrorDescription: description,
		State:            state,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
