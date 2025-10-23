package tests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/handlers"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	_ "github.com/mattn/go-sqlite3"
)

// TestSetup contains common test setup components
type TestSetup struct {
	DB           *sql.DB
	OAuthService *services.OAuth2Service
	OAuthHandler *handlers.OAuth2Handler
	Client       *models.CreateOAuth2ClientResponse
}

// SetupTestEnvironment creates a complete test environment with database, service, and handler
func SetupTestEnvironment(t *testing.T) *TestSetup {
	db := setupTestDB(t)
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
		t.Fatalf("Failed to create test client: %v", err)
	}

	return &TestSetup{
		DB:           db,
		OAuthService: oauthService,
		OAuthHandler: oauthHandler,
		Client:       client,
	}
}

// CreateBrowserRequest creates an HTTP request that simulates a browser
func CreateBrowserRequest(method, url string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	return req
}

// CreateAPIRequest creates an HTTP request that simulates an API client
func CreateAPIRequest(method, url string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "curl/7.68.0")
	return req
}

// CreateTestMux creates a test HTTP mux with OAuth routes
func (ts *TestSetup) CreateTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	ts.OAuthHandler.SetupOAuth2Routes(mux)
	return mux
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
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
