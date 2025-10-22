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

// TestOAuth2Service tests the OAuth 2.0 service functionality
func TestOAuth2Service(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)

	t.Run("CreateClient", func(t *testing.T) {
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client",
			Description: "A test client",
			RedirectURI: "https://test-app.com/callback",
			Scopes:      []string{"read:data"},
		}

		client, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		if client.ClientID == "" {
			t.Error("Expected client ID to be generated")
		}
		if client.ClientSecret == "" {
			t.Error("Expected client secret to be generated")
		}
		if client.Name != req.Name {
			t.Errorf("Expected name %s, got %s", req.Name, client.Name)
		}
	})

	t.Run("GetClient", func(t *testing.T) {
		// First create a client
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 2",
			Description: "Another test client",
			RedirectURI: "https://test-app2.com/callback",
			Scopes:      []string{"read:data", "write:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Now retrieve it
		client, err := oauthService.GetClient(createdClient.ClientID)
		if err != nil {
			t.Fatalf("Failed to get client: %v", err)
		}

		if client.ClientID != createdClient.ClientID {
			t.Errorf("Expected client ID %s, got %s", createdClient.ClientID, client.ClientID)
		}
		if client.Name != req.Name {
			t.Errorf("Expected name %s, got %s", req.Name, client.Name)
		}
	})

	t.Run("ValidateClient", func(t *testing.T) {
		// Create a client
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 3",
			Description: "Another test client",
			RedirectURI: "https://test-app3.com/callback",
			Scopes:      []string{"read:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Test valid credentials
		client, err := oauthService.ValidateClient(createdClient.ClientID, createdClient.ClientSecret, req.RedirectURI)
		if err != nil {
			t.Fatalf("Failed to validate client: %v", err)
		}

		if client.ClientID != createdClient.ClientID {
			t.Errorf("Expected client ID %s, got %s", createdClient.ClientID, client.ClientID)
		}

		// Test invalid credentials
		_, err = oauthService.ValidateClient(createdClient.ClientID, "wrong-secret", req.RedirectURI)
		if err == nil {
			t.Error("Expected error for invalid client secret")
		}
	})

	t.Run("CreateAuthorizationCode", func(t *testing.T) {
		// Create a client first
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 4",
			Description: "Another test client",
			RedirectURI: "https://test-app4.com/callback",
			Scopes:      []string{"read:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Create authorization code
		authCode, err := oauthService.CreateAuthorizationCode(createdClient.ClientID, "user_123", req.RedirectURI, []string{"read:data"})
		if err != nil {
			t.Fatalf("Failed to create authorization code: %v", err)
		}

		if authCode.Code == "" {
			t.Error("Expected authorization code to be generated")
		}
		if authCode.ClientID != createdClient.ClientID {
			t.Errorf("Expected client ID %s, got %s", createdClient.ClientID, authCode.ClientID)
		}
		if authCode.UserID != "user_123" {
			t.Errorf("Expected user ID user_123, got %s", authCode.UserID)
		}
	})

	t.Run("ValidateAuthorizationCode", func(t *testing.T) {
		// Create a client first
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 5",
			Description: "Another test client",
			RedirectURI: "https://test-app5.com/callback",
			Scopes:      []string{"read:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Create authorization code
		authCode, err := oauthService.CreateAuthorizationCode(createdClient.ClientID, "user_123", req.RedirectURI, []string{"read:data"})
		if err != nil {
			t.Fatalf("Failed to create authorization code: %v", err)
		}

		// Validate the code
		validatedCode, err := oauthService.ValidateAuthorizationCode(authCode.Code, createdClient.ClientID, req.RedirectURI)
		if err != nil {
			t.Fatalf("Failed to validate authorization code: %v", err)
		}

		if validatedCode.Code != authCode.Code {
			t.Errorf("Expected code %s, got %s", authCode.Code, validatedCode.Code)
		}
		if !validatedCode.Used {
			t.Error("Expected authorization code to be marked as used")
		}

		// Try to validate the same code again (should fail)
		_, err = oauthService.ValidateAuthorizationCode(authCode.Code, createdClient.ClientID, req.RedirectURI)
		if err == nil {
			t.Error("Expected error when validating used authorization code")
		}
	})

	t.Run("CreateAccessToken", func(t *testing.T) {
		// Create a client first
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 6",
			Description: "Another test client",
			RedirectURI: "https://test-app6.com/callback",
			Scopes:      []string{"read:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Create access token
		accessToken, err := oauthService.CreateAccessToken(createdClient.ClientID, "user_123", []string{"read:data"})
		if err != nil {
			t.Fatalf("Failed to create access token: %v", err)
		}

		if accessToken.AccessToken == "" {
			t.Error("Expected access token to be generated")
		}
		if accessToken.RefreshToken == "" {
			t.Error("Expected refresh token to be generated")
		}
		if accessToken.ClientID != createdClient.ClientID {
			t.Errorf("Expected client ID %s, got %s", createdClient.ClientID, accessToken.ClientID)
		}
		if accessToken.UserID != "user_123" {
			t.Errorf("Expected user ID user_123, got %s", accessToken.UserID)
		}
	})

	t.Run("ValidateAccessToken", func(t *testing.T) {
		// Create a client first
		req := models.CreateOAuth2ClientRequest{
			Name:        "Test Client 7",
			Description: "Another test client",
			RedirectURI: "https://test-app7.com/callback",
			Scopes:      []string{"read:data"},
		}

		createdClient, err := oauthService.CreateClient(req)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// Create access token
		accessToken, err := oauthService.CreateAccessToken(createdClient.ClientID, "user_123", []string{"read:data"})
		if err != nil {
			t.Fatalf("Failed to create access token: %v", err)
		}

		// Validate the token
		userInfo, err := oauthService.ValidateAccessToken(accessToken.AccessToken)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}

		if userInfo.UserID != "user_123" {
			t.Errorf("Expected user ID user_123, got %s", userInfo.UserID)
		}
		if userInfo.ClientID != createdClient.ClientID {
			t.Errorf("Expected client ID %s, got %s", createdClient.ClientID, userInfo.ClientID)
		}
	})

	t.Run("GetUserData", func(t *testing.T) {
		// Test getting user data
		fields := []string{"person.fullName", "person.email", "birthinfo.birthDate"}
		data, err := oauthService.GetUserData("user_123", fields)
		if err != nil {
			t.Fatalf("Failed to get user data: %v", err)
		}

		if data.UserID != "user_123" {
			t.Errorf("Expected user ID user_123, got %s", data.UserID)
		}
		if len(data.Data) == 0 {
			t.Error("Expected user data to be returned")
		}
		if len(data.Fields) != len(fields) {
			t.Errorf("Expected %d fields, got %d", len(fields), len(data.Fields))
		}
	})
}

// TestOAuth2Handler tests the OAuth 2.0 handler functionality
func TestOAuth2Handler(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)
	oauthHandler := handlers.NewOAuth2Handler(oauthService)

	// Create a test client
	req := models.CreateOAuth2ClientRequest{
		Name:        "Test Client",
		Description: "A test client",
		RedirectURI: "https://test-app.com/callback",
		Scopes:      []string{"read:data"},
	}

	client, err := oauthService.CreateClient(req)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	t.Run("CreateClient", func(t *testing.T) {
		createReq := models.CreateOAuth2ClientRequest{
			Name:        "Handler Test Client",
			Description: "A test client for handler testing",
			RedirectURI: "https://handler-test-app.com/callback",
			Scopes:      []string{"read:data"},
		}

		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/oauth2/clients", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		oauthHandler.HandleClients(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var response models.CreateOAuth2ClientResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Name != createReq.Name {
			t.Errorf("Expected name %s, got %s", createReq.Name, response.Name)
		}
	})

	t.Run("GetClient", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/oauth2/clients/"+client.ClientID, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleClientByID(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response models.OAuth2Client
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.ClientID != client.ClientID {
			t.Errorf("Expected client ID %s, got %s", client.ClientID, response.ClientID)
		}
	})

	t.Run("AuthorizationFlow", func(t *testing.T) {
		// Test authorization endpoint
		authURL := fmt.Sprintf("/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=read:data&state=test-state",
			client.ClientID, url.QueryEscape(client.RedirectURI))

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		// Check redirect URL contains authorization code
		location := w.Header().Get("Location")
		if location == "" {
			t.Error("Expected redirect location")
		}

		redirectURL, err := url.Parse(location)
		if err != nil {
			t.Fatalf("Failed to parse redirect URL: %v", err)
		}

		code := redirectURL.Query().Get("code")
		if code == "" {
			t.Error("Expected authorization code in redirect URL")
		}

		state := redirectURL.Query().Get("state")
		if state != "test-state" {
			t.Errorf("Expected state test-state, got %s", state)
		}
	})

	t.Run("TokenExchange", func(t *testing.T) {
		// First get an authorization code
		authURL := fmt.Sprintf("/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=read:data",
			client.ClientID, url.QueryEscape(client.RedirectURI))

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}

		// Extract authorization code from redirect
		location := w.Header().Get("Location")
		redirectURL, _ := url.Parse(location)
		code := redirectURL.Query().Get("code")

		// Now exchange code for token
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

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response models.TokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token")
		}
		if response.TokenType != "Bearer" {
			t.Errorf("Expected token type Bearer, got %s", response.TokenType)
		}
		if response.ExpiresIn < 3590 || response.ExpiresIn > 3600 {
			t.Errorf("Expected expires in ~3600, got %d", response.ExpiresIn)
		}
	})

	t.Run("DataAccess", func(t *testing.T) {
		// First get an access token
		authURL := fmt.Sprintf("/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=read:data",
			client.ClientID, url.QueryEscape(client.RedirectURI))

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		oauthHandler.HandleAuthorize(w, req)

		// Extract authorization code
		location := w.Header().Get("Location")
		redirectURL, _ := url.Parse(location)
		code := redirectURL.Query().Get("code")

		// Exchange code for token
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

		var tokenResponse models.TokenResponse
		json.Unmarshal(w.Body.Bytes(), &tokenResponse)

		// Now test data access
		dataReq := models.DataRequest{
			Fields: []string{"person.fullName", "person.email"},
		}

		reqBody, _ := json.Marshal(dataReq)
		req = httptest.NewRequest("POST", "/api/v1/data", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
		w = httptest.NewRecorder()

		oauthHandler.HandleDataAccess(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var dataResponse models.DataResponse
		err := json.Unmarshal(w.Body.Bytes(), &dataResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if dataResponse.UserID == "" {
			t.Error("Expected user ID in response")
		}
		if len(dataResponse.Data) == 0 {
			t.Error("Expected data in response")
		}
	})
}

// TestOAuth2Middleware tests the OAuth 2.0 middleware functionality
func TestOAuth2Middleware(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)
	oauthMiddleware := middleware.NewOAuth2Middleware(oauthService)

	// Create a test client and access token
	req := models.CreateOAuth2ClientRequest{
		Name:        "Test Client",
		Description: "A test client",
		RedirectURI: "https://test-app.com/callback",
		Scopes:      []string{"read:data"},
	}

	client, err := oauthService.CreateClient(req)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	accessToken, err := oauthService.CreateAccessToken(client.ClientID, "user_123", []string{"read:data"})
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	t.Run("RequireOAuth2Token", func(t *testing.T) {
		// Test with valid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo, ok := middleware.UserInfoFromContext(r.Context())
			if !ok {
				t.Error("Expected user info in context")
			}
			if userInfo.UserID != "user_123" {
				t.Errorf("Expected user ID user_123, got %s", userInfo.UserID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test with invalid token
		req = httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w = httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		// Test with missing token
		req = httptest.NewRequest("GET", "/protected", nil)
		w = httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("RequireScope", func(t *testing.T) {
		// Test with valid scope
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
		w := httptest.NewRecorder()

		handler := oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireScope("read:data")(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test with invalid scope
		handler = oauthMiddleware.RequireOAuth2Token(
			oauthMiddleware.RequireScope("write:data")(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			),
		)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *sql.DB {
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
		code_challenge VARCHAR(255),
		code_challenge_method VARCHAR(10) DEFAULT 'S256',
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		used BOOLEAN DEFAULT false,
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

	CREATE TABLE IF NOT EXISTS oauth2_users (
		user_id VARCHAR(255) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create OAuth 2.0 tables: %v", err)
	}

	return db
}
