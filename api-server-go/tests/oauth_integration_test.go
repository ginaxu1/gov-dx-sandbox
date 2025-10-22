package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/middleware"
	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	_ "github.com/mattn/go-sqlite3"
)

// TestCompleteOAuth2Flow tests the complete OAuth 2.0 Authorization Code Flow
func TestCompleteOAuth2Flow(t *testing.T) {
	// Setup test database (in a real test, this would be a test database)
	db := setupIntegrationTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)
	oauthHandler := handlers.NewOAuth2Handler(oauthService)

	// Create a test OAuth 2.0 client
	clientReq := models.CreateOAuth2ClientRequest{
		Name:        "Integration Test Client",
		Description: "A client for integration testing",
		RedirectURI: "https://integration-test-app.com/auth/callback",
		Scopes:      []string{"read:data"},
	}

	client, err := oauthService.CreateClient(clientReq)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	t.Run("CompleteAuthorizationCodeFlow", func(t *testing.T) {
		// Step 1: Client initiates authorization request
		authURL := fmt.Sprintf("/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=read:data&state=integration-test-state",
			client.ClientID, url.QueryEscape(client.RedirectURI))

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		// Verify authorization response
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302 (redirect), got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location == "" {
			t.Fatal("Expected redirect location header")
		}

		// Parse redirect URL to extract authorization code
		redirectURL, err := url.Parse(location)
		if err != nil {
			t.Fatalf("Failed to parse redirect URL: %v", err)
		}

		code := redirectURL.Query().Get("code")
		if code == "" {
			t.Fatal("Expected authorization code in redirect URL")
		}

		state := redirectURL.Query().Get("state")
		if state != "integration-test-state" {
			t.Errorf("Expected state 'integration-test-state', got '%s'", state)
		}

		// Step 2: Client exchanges authorization code for access token
		formData := url.Values{}
		formData.Set("grant_type", "authorization_code")
		formData.Set("code", code)
		formData.Set("redirect_uri", client.RedirectURI)
		formData.Set("client_id", client.ClientID)
		formData.Set("client_secret", client.ClientSecret)

		req = httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		oauthHandler.HandleToken(w, req)

		// Verify token response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var tokenResponse models.TokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &tokenResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal token response: %v", err)
		}

		if tokenResponse.AccessToken == "" {
			t.Fatal("Expected access token in response")
		}
		if tokenResponse.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got '%s'", tokenResponse.TokenType)
		}
		if tokenResponse.ExpiresIn != 3600 {
			t.Errorf("Expected expires in 3600 seconds, got %d", tokenResponse.ExpiresIn)
		}
		if tokenResponse.RefreshToken == "" {
			t.Fatal("Expected refresh token in response")
		}
		if tokenResponse.Scope != "read:data" {
			t.Errorf("Expected scope 'read:data', got '%s'", tokenResponse.Scope)
		}

		// Step 3: Client uses access token to access protected resource
		dataReq := models.DataRequest{
			Fields: []string{"person.fullName", "person.email", "birthinfo.birthDate"},
		}

		reqBody, _ := json.Marshal(dataReq)
		req = httptest.NewRequest("POST", "/api/v1/data", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
		w = httptest.NewRecorder()

		oauthHandler.HandleDataAccess(w, req)

		// Verify data access response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var dataResponse models.DataResponse
		err = json.Unmarshal(w.Body.Bytes(), &dataResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal data response: %v", err)
		}

		if dataResponse.UserID == "" {
			t.Fatal("Expected user ID in data response")
		}
		if len(dataResponse.Data) == 0 {
			t.Fatal("Expected data in response")
		}
		if len(dataResponse.Fields) != len(dataReq.Fields) {
			t.Errorf("Expected %d fields, got %d", len(dataReq.Fields), len(dataResponse.Fields))
		}

		// Verify that the data is user-specific
		if dataResponse.UserID != "user_123" {
			t.Errorf("Expected user ID 'user_123', got '%s'", dataResponse.UserID)
		}

		// Step 4: Test refresh token flow
		refreshFormData := url.Values{}
		refreshFormData.Set("grant_type", "refresh_token")
		refreshFormData.Set("refresh_token", tokenResponse.RefreshToken)
		refreshFormData.Set("client_id", client.ClientID)
		refreshFormData.Set("client_secret", client.ClientSecret)

		req = httptest.NewRequest("POST", "/oauth2/refresh", strings.NewReader(refreshFormData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		oauthHandler.HandleRefresh(w, req)

		// Verify refresh token response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		var refreshResponse models.TokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &refreshResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal refresh response: %v", err)
		}

		if refreshResponse.AccessToken == "" {
			t.Fatal("Expected access token in refresh response")
		}
		if refreshResponse.AccessToken == tokenResponse.AccessToken {
			t.Error("Expected new access token, got same token")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid client ID
		authURL := "/oauth2/authorize?response_type=code&client_id=invalid-client&redirect_uri=https://test.com/callback"
		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid client, got %d", w.Code)
		}

		// Test invalid authorization code
		formData := url.Values{}
		formData.Set("grant_type", "authorization_code")
		formData.Set("code", "invalid-code")
		formData.Set("redirect_uri", client.RedirectURI)
		formData.Set("client_id", client.ClientID)
		formData.Set("client_secret", client.ClientSecret)

		req = httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		oauthHandler.HandleToken(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid code, got %d", w.Code)
		}

		// Test invalid access token
		dataReq := models.DataRequest{
			Fields: []string{"person.fullName"},
		}

		reqBody, _ := json.Marshal(dataReq)
		req = httptest.NewRequest("POST", "/api/v1/data", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token")
		w = httptest.NewRecorder()

		oauthHandler.HandleDataAccess(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for invalid token, got %d", w.Code)
		}
	})

	t.Run("ScopeValidation", func(t *testing.T) {
		// Create a client with limited scope
		limitedClientReq := models.CreateOAuth2ClientRequest{
			Name:        "Limited Scope Client",
			Description: "A client with limited scope",
			RedirectURI: "https://limited-test-app.com/callback",
			Scopes:      []string{"read:basic"},
		}

		limitedClient, err := oauthService.CreateClient(limitedClientReq)
		if err != nil {
			t.Fatalf("Failed to create limited scope client: %v", err)
		}

		// Get authorization code
		authURL := fmt.Sprintf("/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=read:basic",
			limitedClient.ClientID, url.QueryEscape(limitedClient.RedirectURI))

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		// Extract authorization code
		location := w.Header().Get("Location")
		redirectURL, _ := url.Parse(location)
		code := redirectURL.Query().Get("code")

		// Exchange code for token
		formData := url.Values{}
		formData.Set("grant_type", "authorization_code")
		formData.Set("code", code)
		formData.Set("redirect_uri", limitedClient.RedirectURI)
		formData.Set("client_id", limitedClient.ClientID)
		formData.Set("client_secret", limitedClient.ClientSecret)

		req = httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()

		oauthHandler.HandleToken(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var tokenResponse models.TokenResponse
		json.Unmarshal(w.Body.Bytes(), &tokenResponse)

		// Try to access data with limited scope
		dataReq := models.DataRequest{
			Fields: []string{"person.fullName"},
		}

		reqBody, _ := json.Marshal(dataReq)
		req = httptest.NewRequest("POST", "/api/v1/data", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
		w = httptest.NewRecorder()

		oauthHandler.HandleDataAccess(w, req)

		// This should fail due to insufficient scope
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for insufficient scope, got %d", w.Code)
		}
	})
}

// TestOAuth2MiddlewareIntegration tests the OAuth 2.0 middleware with real requests
func TestOAuth2MiddlewareIntegration(t *testing.T) {
	// Setup test database
	db := setupIntegrationTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)
	oauthMiddleware := middleware.NewOAuth2Middleware(oauthService)

	// Create a test client and access token
	clientReq := models.CreateOAuth2ClientRequest{
		Name:        "Middleware Test Client",
		Description: "A client for middleware testing",
		RedirectURI: "https://middleware-test-app.com/callback",
		Scopes:      []string{"read:data", "write:data"},
	}

	client, err := oauthService.CreateClient(clientReq)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	accessToken, err := oauthService.CreateAccessToken(client.ClientID, "user_123", []string{"read:data", "write:data"})
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Create a test handler that uses the middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo, ok := middleware.UserInfoFromContext(r.Context())
		if !ok {
			http.Error(w, "No user info in context", http.StatusInternalServerError)
			return
		}

		clientID, ok := middleware.ClientIDFromContext(r.Context())
		if !ok {
			http.Error(w, "No client ID in context", http.StatusInternalServerError)
			return
		}

		scopes, ok := middleware.ScopesFromContext(r.Context())
		if !ok {
			http.Error(w, "No scopes in context", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"user_id":   userInfo.UserID,
			"client_id": clientID,
			"scopes":    scopes,
			"message":   "Protected resource accessed successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	t.Run("ValidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(testHandler)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["user_id"] != "user_123" {
			t.Errorf("Expected user_id 'user_123', got '%v'", response["user_id"])
		}
		if response["client_id"] != client.ClientID {
			t.Errorf("Expected client_id '%s', got '%v'", client.ClientID, response["client_id"])
		}
	})

	t.Run("ScopeValidation", func(t *testing.T) {
		// Test with required scope
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireScope("read:data")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test with missing scope
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w = httptest.NewRecorder()

		handler = oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireScope("admin:access")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("AnyScopeValidation", func(t *testing.T) {
		// Test with any of the required scopes
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireAnyScope("read:data", "write:data")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test with none of the required scopes
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w = httptest.NewRecorder()

		handler = oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireAnyScope("admin:access", "super:user")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("AllScopesValidation", func(t *testing.T) {
		// Test with all required scopes
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireAllScopes("read:data", "write:data")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test with missing one of the required scopes
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w = httptest.NewRecorder()

		handler = oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireAllScopes("read:data", "admin:access")(testHandler),
		)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})
}

// Helper function to setup test database for integration tests
func setupIntegrationTestDB(t *testing.T) *sql.DB {
	// For testing, we'll use an in-memory SQLite database
	// This avoids the need for a real database setup
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create OAuth 2.0 tables
	createTables := `
	CREATE TABLE IF NOT EXISTS oauth2_clients (
		client_id VARCHAR(255) PRIMARY KEY,
		client_secret VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		redirect_uri VARCHAR(255) NOT NULL,
		scopes TEXT,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
		code VARCHAR(255) PRIMARY KEY,
		client_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		redirect_uri VARCHAR(255) NOT NULL,
		scopes TEXT,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT true,
		FOREIGN KEY (client_id) REFERENCES oauth2_clients(client_id)
	);

		CREATE TABLE IF NOT EXISTS oauth2_tokens (
			token VARCHAR(255) PRIMARY KEY,
			token_type VARCHAR(20) NOT NULL CHECK (token_type IN ('access', 'refresh')),
			client_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			scopes TEXT,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT true,
			related_token VARCHAR(255),
			parent_token VARCHAR(255),
			FOREIGN KEY (client_id) REFERENCES oauth2_clients(client_id)
		);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create OAuth 2.0 tables: %v", err)
	}

	return db
}
