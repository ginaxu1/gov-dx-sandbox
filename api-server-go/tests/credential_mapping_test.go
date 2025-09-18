package tests

import (
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

// TestCredentialMapping tests the credential mapping functionality
func TestCredentialMapping(t *testing.T) {
	// Setup
	asgardeoService := services.NewAsgardeoService("https://api.asgardeo.io/t/test-tenant")
	consumerService := services.NewConsumerServiceWithAsgardeo(asgardeoService)

	// Set test environment variables
	os.Setenv("ASGARDEO_CLIENT_ID", "test_client_id")
	os.Setenv("ASGARDEO_CLIENT_SECRET", "test_client_secret")
	defer os.Unsetenv("ASGARDEO_CLIENT_ID")
	defer os.Unsetenv("ASGARDEO_CLIENT_SECRET")

	t.Run("CreateCredentialMapping", func(t *testing.T) {
		mapping, err := consumerService.CreateCredentialMapping(
			"test_consumer_123",
			"test_asgardeo_client_id",
			"test_asgardeo_client_secret",
		)

		if err != nil {
			t.Fatalf("Failed to create credential mapping: %v", err)
		}

		if mapping.ConsumerID != "test_consumer_123" {
			t.Errorf("Expected consumer ID 'test_consumer_123', got '%s'", mapping.ConsumerID)
		}

		if mapping.AsgardeoClientID != "test_asgardeo_client_id" {
			t.Errorf("Expected Asgardeo client ID 'test_asgardeo_client_id', got '%s'", mapping.AsgardeoClientID)
		}

		if mapping.AsgardeoClientSecret != "test_asgardeo_client_secret" {
			t.Errorf("Expected Asgardeo client secret 'test_asgardeo_client_secret', got '%s'", mapping.AsgardeoClientSecret)
		}

		if mapping.APIKey == "" {
			t.Error("Expected API key to be generated")
		}

		if mapping.APISecret == "" {
			t.Error("Expected API secret to be generated")
		}
	})

	t.Run("ValidateCredentialMapping", func(t *testing.T) {
		// Create a mapping first
		mapping, err := consumerService.CreateCredentialMapping(
			"test_consumer_456",
			"test_asgardeo_client_id",
			"test_asgardeo_client_secret",
		)

		if err != nil {
			t.Fatalf("Failed to create credential mapping: %v", err)
		}

		// Test valid credentials
		validMapping, err := consumerService.ValidateAndGetMapping(mapping.APIKey, mapping.APISecret)
		if err != nil {
			t.Errorf("Expected valid credentials to pass validation, got error: %v", err)
		}

		if validMapping.ConsumerID != "test_consumer_456" {
			t.Errorf("Expected consumer ID 'test_consumer_456', got '%s'", validMapping.ConsumerID)
		}

		// Test invalid API key
		_, err = consumerService.ValidateAndGetMapping("invalid_key", mapping.APISecret)
		if err == nil {
			t.Error("Expected invalid API key to fail validation")
		}

		// Test invalid API secret
		_, err = consumerService.ValidateAndGetMapping(mapping.APIKey, "invalid_secret")
		if err == nil {
			t.Error("Expected invalid API secret to fail validation")
		}
	})

	t.Run("IsCredentialMappingConfigured", func(t *testing.T) {
		// Test with environment variables set
		os.Setenv("ASGARDEO_CLIENT_ID", "test_client_id")
		os.Setenv("ASGARDEO_CLIENT_SECRET", "test_client_secret")

		if !consumerService.IsCredentialMappingConfigured() {
			t.Error("Expected credential mapping to be configured with environment variables")
		}

		// Test without environment variables
		os.Unsetenv("ASGARDEO_CLIENT_ID")
		os.Unsetenv("ASGARDEO_CLIENT_SECRET")

		if consumerService.IsCredentialMappingConfigured() {
			t.Error("Expected credential mapping to not be configured without environment variables")
		}
	})
}

// TestTokenExchangeWithValidCredentials tests token exchange with valid credentials
func TestTokenExchangeWithValidCredentials(t *testing.T) {
	// Setup
	asgardeoService := services.NewAsgardeoService("https://api.asgardeo.io/t/test-tenant")
	consumerService := services.NewConsumerServiceWithAsgardeo(asgardeoService)

	// Set test environment variables
	os.Setenv("ASGARDEO_CLIENT_ID", "test_client_id")
	os.Setenv("ASGARDEO_CLIENT_SECRET", "test_client_secret")
	defer os.Unsetenv("ASGARDEO_CLIENT_ID")
	defer os.Unsetenv("ASGARDEO_CLIENT_SECRET")

	// Create a credential mapping
	mapping, err := consumerService.CreateCredentialMapping(
		"test_consumer_789",
		"test_asgardeo_client_id",
		"test_asgardeo_client_secret",
	)

	if err != nil {
		t.Fatalf("Failed to create credential mapping: %v", err)
	}

	t.Run("ValidTokenExchange", func(t *testing.T) {
		req := models.TokenExchangeRequest{
			APIKey:    mapping.APIKey,
			APISecret: mapping.APISecret,
			Scope:     "gov-dx-api",
		}

		// This will fail with Asgardeo connection error, but should not fail with credential validation
		_, err := consumerService.ExchangeCredentialsForToken(req)

		// Should fail with Asgardeo connection error, not credential validation error
		if err != nil && err.Error() == "invalid credentials: invalid API key" {
			t.Errorf("Expected Asgardeo connection error, got credential validation error: %v", err)
		}
	})

	t.Run("InvalidTokenExchange", func(t *testing.T) {
		req := models.TokenExchangeRequest{
			APIKey:    "invalid_key",
			APISecret: "invalid_secret",
			Scope:     "gov-dx-api",
		}

		_, err := consumerService.ExchangeCredentialsForToken(req)

		if err == nil {
			t.Error("Expected invalid credentials to fail")
		}

		if err.Error() != "invalid credentials: invalid API key" {
			t.Errorf("Expected 'invalid credentials: invalid API key' error, got: %v", err)
		}
	})

	t.Run("MissingFields", func(t *testing.T) {
		req := models.TokenExchangeRequest{
			APIKey: "test_key",
			// Missing APISecret
		}

		_, err := consumerService.ExchangeCredentialsForToken(req)

		if err == nil {
			t.Error("Expected missing fields to fail")
		}

		if err.Error() != "API secret is required" {
			t.Errorf("Expected 'API secret is required' error, got: %v", err)
		}
	})
}

// TestCredentialMappingConcurrency tests concurrent access to credential mappings
func TestCredentialMappingConcurrency(t *testing.T) {
	asgardeoService := services.NewAsgardeoService("https://api.asgardeo.io/t/test-tenant")
	consumerService := services.NewConsumerServiceWithAsgardeo(asgardeoService)

	// Set test environment variables
	os.Setenv("ASGARDEO_CLIENT_ID", "test_client_id")
	os.Setenv("ASGARDEO_CLIENT_SECRET", "test_client_secret")
	defer os.Unsetenv("ASGARDEO_CLIENT_ID")
	defer os.Unsetenv("ASGARDEO_CLIENT_SECRET")

	t.Run("ConcurrentCredentialCreation", func(t *testing.T) {
		// Create multiple credential mappings concurrently
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				_, err := consumerService.CreateCredentialMapping(
					"test_consumer_"+string(rune(id)),
					"test_asgardeo_client_id",
					"test_asgardeo_client_secret",
				)

				if err != nil {
					t.Errorf("Failed to create credential mapping %d: %v", id, err)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("ConcurrentCredentialValidation", func(t *testing.T) {
		// Create a credential mapping
		mapping, err := consumerService.CreateCredentialMapping(
			"test_consumer_concurrent",
			"test_asgardeo_client_id",
			"test_asgardeo_client_secret",
		)

		if err != nil {
			t.Fatalf("Failed to create credential mapping: %v", err)
		}

		// Validate credentials concurrently
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				_, err := consumerService.ValidateAndGetMapping(mapping.APIKey, mapping.APISecret)

				if err != nil {
					t.Errorf("Failed to validate credentials: %v", err)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
