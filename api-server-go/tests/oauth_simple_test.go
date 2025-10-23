package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestOAuth2BasicFlow tests the basic OAuth 2.0 flow
func TestOAuth2BasicFlow(t *testing.T) {
	// Setup test environment
	setup := SetupTestEnvironment(t)
	defer setup.DB.Close()

	// Verify client exists in database
	retrievedClient, err := setup.OAuthService.GetClient(setup.Client.ClientID)
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
		formData.Set("client_id", setup.Client.ClientID)
		formData.Set("client_secret", setup.Client.ClientSecret)

		req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		// Use HTTP router
		mux := setup.CreateTestMux()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}
	})

	t.Run("AuthorizationFlow", func(t *testing.T) {
		// Test authorization endpoint
		authURL := "/oauth2/authorize?response_type=code&client_id=" + setup.Client.ClientID + "&redirect_uri=" + url.QueryEscape(setup.Client.RedirectURI) + "&scope=read:data&state=test-state"

		req := CreateBrowserRequest("GET", authURL)
		w := httptest.NewRecorder()

		// Use HTTP router
		mux := setup.CreateTestMux()
		mux.ServeHTTP(w, req)

		// Authorization endpoint should redirect (302) to login page
		if w.Code != http.StatusFound {
			t.Errorf("Expected status 302, got %d. Response: %s", w.Code, w.Body.String())
		}
	})
}
