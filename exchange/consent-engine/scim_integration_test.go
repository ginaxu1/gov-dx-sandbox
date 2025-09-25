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

	t.Run("Owner_Email_Lookup_Hardcoded_Mapping", func(t *testing.T) {
		// Test that owner email lookup uses hardcoded mapping
		// Since M2M authentication was removed, we now use hardcoded mapping for simplicity

		cleanup := SetupTestWithCleanup(t)
		defer cleanup()

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
