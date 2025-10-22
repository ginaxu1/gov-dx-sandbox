package tests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/handlers"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	_ "github.com/mattn/go-sqlite3"
)

// TestOAuth2BasicFlow tests the basic OAuth 2.0 flow
func TestOAuth2BasicFlow(t *testing.T) {
	// Setup test database
	db := setupSimpleTestDB(t)
	defer db.Close()

	oauthService := services.NewOAuth2Service(db)
	oauthHandler := handlers.NewOAuth2Handler(oauthService)

	// Create a test client
	clientReq := models.CreateOAuth2ClientRequest{
		Name:        "Test Client",
		Description: "A test client",
		RedirectURI: "https://test.com/callback",
		Scopes:      []string{"read:data"},
	}

	client, err := oauthService.CreateClient(clientReq)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify client exists in database
	retrievedClient, err := oauthService.GetClient(client.ClientID)
	if err != nil {
		t.Fatalf("Failed to retrieve client: %v", err)
	}
	if retrievedClient == nil {
		t.Fatal("Client not found in database")
	}

	t.Run("ClientCredentialsFlow", func(t *testing.T) {
		// Test client credentials flow
		formData := url.Values{}
		formData.Set("grant_type", "client_credentials")
		formData.Set("client_id", client.ClientID)
		formData.Set("client_secret", client.ClientSecret)

		req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		// Use HTTP router
		mux := http.NewServeMux()
		oauthHandler.SetupOAuth2Routes(mux)
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}
	})

	t.Run("AuthorizationFlow", func(t *testing.T) {
		// Test authorization endpoint
		authURL := "/oauth2/authorize?response_type=code&client_id=" + client.ClientID + "&redirect_uri=" + url.QueryEscape(client.RedirectURI) + "&scope=read:data&state=test-state"

		req := httptest.NewRequest("GET", authURL, nil)
		w := httptest.NewRecorder()

		// Use HTTP router
		mux := http.NewServeMux()
		oauthHandler.SetupOAuth2Routes(mux)
		mux.ServeHTTP(w, req)

		// Authorization endpoint should redirect (302) to login page
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d", w.Code)
		}
	})
}

// setupSimpleTestDB creates an in-memory SQLite database for testing
func setupSimpleTestDB(t *testing.T) *sql.DB {
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
		used BOOLEAN DEFAULT false
	);

	CREATE TABLE IF NOT EXISTS oauth2_tokens (
		token_id VARCHAR(255) PRIMARY KEY,
		token VARCHAR(255) NOT NULL,
		token_type VARCHAR(50) NOT NULL,
		client_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		scopes TEXT,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT true,
		related_token VARCHAR(255),
		parent_token VARCHAR(255)
	);
	`

	if _, err := db.Exec(createTables); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return db
}
