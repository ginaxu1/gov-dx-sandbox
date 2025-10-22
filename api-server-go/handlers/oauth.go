package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/api-server-go/shared"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
)

// OAuth2Handler handles OAuth 2.0 endpoints
type OAuth2Handler struct {
	oauthService *services.OAuth2Service
}

// NewOAuth2Handler creates a new OAuth 2.0 handler
func NewOAuth2Handler(oauthService *services.OAuth2Service) *OAuth2Handler {
	return &OAuth2Handler{oauthService: oauthService}
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

	// For testing purposes, we'll create an authorization code with a dummy user ID
	// In a real implementation, this would come from user authentication
	userID := "user_123" // This should come from the authenticated user session
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
	err := r.ParseForm()
	if err != nil {
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

	// Validate request
	if req.GrantType != "authorization_code" {
		h.respondWithOAuth2Error(w, "unsupported_grant_type", "Only 'authorization_code' grant type is supported", "")
		return
	}

	if req.Code == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "code is required", "")
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

	if req.RedirectURI == "" {
		h.respondWithOAuth2Error(w, "invalid_request", "redirect_uri is required", "")
		return
	}

	// Validate client credentials
	_, clientErr := h.oauthService.ValidateClient(req.ClientID, req.ClientSecret, req.RedirectURI)
	if clientErr != nil {
		h.respondWithOAuth2Error(w, "invalid_client", "Invalid client credentials", "")
		return
	}

	// Exchange code for token using the oauth2 package with PKCE
	ctx := context.Background()
	token, err := h.oauthService.ExchangeCodeForToken(ctx, req.ClientID, req.Code, req.RedirectURI, req.CodeVerifier)
	if err != nil {
		slog.Error("Failed to exchange code for token", "error", err, "client_id", req.ClientID)
		h.respondWithOAuth2Error(w, "invalid_grant", "Invalid or expired authorization code", "")
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
	response := models.TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: token.RefreshToken,
		Scope:        "read:data", // Default scope
	}

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
	response := models.TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: token.RefreshToken,
		Scope:        "read:data", // Default scope
	}

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

// Exported wrapper methods for testing
func (h *OAuth2Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	h.handleAuthorize(w, r)
}

func (h *OAuth2Handler) HandleToken(w http.ResponseWriter, r *http.Request) {
	h.handleToken(w, r)
}

func (h *OAuth2Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	h.handleRefresh(w, r)
}

func (h *OAuth2Handler) HandleDataAccess(w http.ResponseWriter, r *http.Request) {
	h.handleDataAccess(w, r)
}

func (h *OAuth2Handler) HandleClients(w http.ResponseWriter, r *http.Request) {
	h.handleClients(w, r)
}

func (h *OAuth2Handler) HandleClientByID(w http.ResponseWriter, r *http.Request) {
	h.handleClientByID(w, r)
}
