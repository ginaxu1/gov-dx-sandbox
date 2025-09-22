package main

import (
	"testing"
)

// TestSCIMIntegration demonstrates the SCIM integration functionality
func TestSCIMIntegration(t *testing.T) {
	t.Run("SCIM_Client_Initialization", func(t *testing.T) {
		// Test SCIM client initialization
		baseURL := "https://api.asgardeo.io/t/lankasoftwarefoundation"
		clientID := "test-client-id"
		clientSecret := "test-client-secret"

		client := NewAsgardeoSCIMClient(baseURL, clientID, clientSecret)

		if client == nil {
			t.Fatal("SCIM client should not be nil")
		}

		if client.baseURL != baseURL {
			t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
		}

		if client.clientID != clientID {
			t.Errorf("Expected clientID %s, got %s", clientID, client.clientID)
		}

		if client.clientSecret != clientSecret {
			t.Errorf("Expected clientSecret %s, got %s", clientSecret, client.clientSecret)
		}

		if client.httpClient == nil {
			t.Error("HTTP client should not be nil")
		}
	})

	t.Run("SCIM_User_Structure", func(t *testing.T) {
		// Test SCIM user structure
		user := SCIMUser{
			ID:       "test-user-id",
			UserName: "testuser",
			Emails: []struct {
				Value   string `json:"value"`
				Primary bool   `json:"primary"`
				Type    string `json:"type"`
			}{
				{
					Value:   "test@example.com",
					Primary: true,
					Type:    "work",
				},
			},
			Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
			Meta: struct {
				ResourceType string `json:"resourceType"`
			}{
				ResourceType: "User",
			},
		}

		if user.ID != "test-user-id" {
			t.Errorf("Expected ID %s, got %s", "test-user-id", user.ID)
		}

		if len(user.Emails) != 1 {
			t.Errorf("Expected 1 email, got %d", len(user.Emails))
		}

		if user.Emails[0].Value != "test@example.com" {
			t.Errorf("Expected email %s, got %s", "test@example.com", user.Emails[0].Value)
		}
	})

	t.Run("SCIM_Response_Structure", func(t *testing.T) {
		// Test SCIM response structure
		response := SCIMResponse{
			TotalResults: 1,
			ItemsPerPage: 10,
			StartIndex:   1,
			Resources: []SCIMUser{
				{
					ID:       "user-123",
					UserName: "john.doe",
					Emails: []struct {
						Value   string `json:"value"`
						Primary bool   `json:"primary"`
						Type    string `json:"type"`
					}{
						{
							Value:   "john.doe@example.com",
							Primary: true,
							Type:    "work",
						},
					},
				},
			},
			Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:ListResponse"},
		}

		if response.TotalResults != 1 {
			t.Errorf("Expected TotalResults %d, got %d", 1, response.TotalResults)
		}

		if len(response.Resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(response.Resources))
		}

		if response.Resources[0].UserName != "john.doe" {
			t.Errorf("Expected username %s, got %s", "john.doe", response.Resources[0].UserName)
		}
	})

	t.Run("M2M_Token_Request_Structure", func(t *testing.T) {
		// Test M2M token request structure
		tokenReq := M2MTokenRequest{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			GrantType:    "client_credentials",
			Scope:        "internal_user_mgt_view",
		}

		if tokenReq.ClientID != "test-client-id" {
			t.Errorf("Expected ClientID %s, got %s", "test-client-id", tokenReq.ClientID)
		}

		if tokenReq.GrantType != "client_credentials" {
			t.Errorf("Expected GrantType %s, got %s", "client_credentials", tokenReq.GrantType)
		}

		if tokenReq.Scope != "internal_user_mgt_view" {
			t.Errorf("Expected Scope %s, got %s", "internal_user_mgt_view", tokenReq.Scope)
		}
	})

	t.Run("M2M_Token_Response_Structure", func(t *testing.T) {
		// Test M2M token response structure
		tokenResp := M2MTokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "internal_user_mgt_view",
		}

		if tokenResp.AccessToken != "test-access-token" {
			t.Errorf("Expected AccessToken %s, got %s", "test-access-token", tokenResp.AccessToken)
		}

		if tokenResp.TokenType != "Bearer" {
			t.Errorf("Expected TokenType %s, got %s", "Bearer", tokenResp.TokenType)
		}

		if tokenResp.ExpiresIn != 3600 {
			t.Errorf("Expected ExpiresIn %d, got %d", 3600, tokenResp.ExpiresIn)
		}
	})

	t.Run("Owner_Email_Lookup_Fallback", func(t *testing.T) {
		// Test that owner email lookup falls back to hardcoded mapping when M2M credentials are not configured
		// This test ensures backward compatibility

		// Reset global SCIM client to ensure clean state
		scimClient = nil

		// Test with a known owner_id from hardcoded mapping
		email, err := getOwnerEmailByID("199512345678")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedEmail := "regina@opensource.lk"
		if email != expectedEmail {
			t.Errorf("Expected email %s, got %s", expectedEmail, email)
		}
	})

	t.Run("Owner_Email_Lookup_Unknown_ID", func(t *testing.T) {
		// Test that owner email lookup returns error for unknown owner_id

		// Reset global SCIM client to ensure clean state
		scimClient = nil

		// Test with an unknown owner_id
		email, err := getOwnerEmailByID("unknown-id")
		if err == nil {
			t.Error("Expected error for unknown owner_id, got nil")
		}

		if email != "" {
			t.Errorf("Expected empty email for unknown owner_id, got %s", email)
		}
	})
}
