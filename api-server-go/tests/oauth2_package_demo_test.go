package tests

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// TestOAuth2PackageUsage demonstrates how to use golang.org/x/oauth2 package
// with our OAuth 2.0 server implementation
func TestOAuth2PackageUsage(t *testing.T) {
	// This test demonstrates how a client application would use the oauth2 package
	// to interact with our OAuth 2.0 server

	baseURL := "http://localhost:3000"

	// Create oauth2.Config for client application
	config := &oauth2.Config{
		ClientID:     "demo-client-123",
		ClientSecret: "demo-secret-456",
		RedirectURL:  "https://demo-app.com/auth/callback",
		Scopes:       []string{"read:data"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + "/oauth2/authorize",
			TokenURL: baseURL + "/oauth2/token",
		},
	}

	t.Run("ClientSideAuthorizationFlow", func(t *testing.T) {
		// Step 1: Generate authorization URL (client side)
		state := "client-side-state-123"
		authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

		// Verify the URL is properly formatted
		parsedURL, err := url.Parse(authURL)
		if err != nil {
			t.Fatalf("Failed to parse auth URL: %v", err)
		}

		query := parsedURL.Query()
		if query.Get("response_type") != "code" {
			t.Error("Expected response_type=code")
		}
		if query.Get("client_id") != config.ClientID {
			t.Errorf("Expected client_id=%s, got %s", config.ClientID, query.Get("client_id"))
		}
		if query.Get("redirect_uri") != config.RedirectURL {
			t.Errorf("Expected redirect_uri=%s, got %s", config.RedirectURL, query.Get("redirect_uri"))
		}
		if query.Get("state") != state {
			t.Errorf("Expected state=%s, got %s", state, query.Get("state"))
		}
		if query.Get("scope") != "read:data" {
			t.Errorf("Expected scope=read:data, got %s", query.Get("scope"))
		}

		t.Logf("Authorization URL: %s", authURL)
	})

	t.Run("TokenSourceUsage", func(t *testing.T) {
		// This demonstrates how a client would use TokenSource for automatic token refresh
		ctx := context.Background()

		// Create a token source (this would typically be done after getting initial token)
		token := &oauth2.Token{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour),
		}

		tokenSource := config.TokenSource(ctx, token)

		// Verify token source is created
		if tokenSource == nil {
			t.Fatal("Expected token source to be created")
		}

		// In a real application, the client would use this token source
		// to automatically refresh tokens when they expire
		t.Log("Token source created successfully")
	})

	t.Run("ClientCredentialsFlow", func(t *testing.T) {
		// This demonstrates how to use client credentials flow
		// (though our server doesn't implement this yet)

		clientConfig := &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     config.Endpoint,
		}

		// In a real implementation, this would use client credentials
		// For now, we just verify the config is set up correctly
		if clientConfig.ClientID == "" {
			t.Error("Expected client ID to be set")
		}
		if clientConfig.ClientSecret == "" {
			t.Error("Expected client secret to be set")
		}
		if clientConfig.Endpoint.AuthURL == "" {
			t.Error("Expected auth URL to be set")
		}
		if clientConfig.Endpoint.TokenURL == "" {
			t.Error("Expected token URL to be set")
		}

		t.Log("Client credentials config created successfully")
	})
}

// DemonstrateOAuth2PackageUsage demonstrates how a real client application would use our OAuth 2.0 server
func DemonstrateOAuth2PackageUsage() {
	// This is an example of how a client application would integrate with our OAuth 2.0 server

	baseURL := "http://localhost:3000"

	// Step 1: Configure OAuth 2.0 client
	config := &oauth2.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURL:  "https://your-app.com/auth/callback",
		Scopes:       []string{"read:data"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + "/oauth2/authorize",
			TokenURL: baseURL + "/oauth2/token",
		},
	}

	// Step 2: Generate authorization URL
	state := "random-state-string"
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Step 3: Redirect user to authURL
	fmt.Printf("Redirect user to: %s\n", authURL)

	// Step 4: After user authorizes, exchange code for token
	ctx := context.Background()
	// In a real application, you would get the code from the callback URL
	code := "authorization-code-from-callback"

	token, err := config.Exchange(ctx, code)
	if err != nil {
		fmt.Printf("Token exchange failed: %v\n", err)
		return
	}

	// Step 5: Use the token to make authenticated requests
	fmt.Printf("Access token: %s\n", token.AccessToken)
	fmt.Printf("Token type: %s\n", token.TokenType)
	fmt.Printf("Expires at: %v\n", token.Expiry)

	// Step 6: Use TokenSource for automatic token refresh
	tokenSource := config.TokenSource(ctx, token)

	// The token source will automatically refresh the token when needed
	// You can use it with http.Client
	client := oauth2.NewClient(ctx, tokenSource)

	// Use the client to make authenticated requests
	// resp, err := client.Get("http://localhost:3000/api/v1/data")
	_ = client // Use client for authenticated requests
}
