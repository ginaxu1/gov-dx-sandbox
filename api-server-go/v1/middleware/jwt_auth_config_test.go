package middleware

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateJWTAuthConfig_ValidConfig verifies that valid configurations pass validation
func TestValidateJWTAuthConfig_ValidConfig(t *testing.T) {
	validConfigs := []JWTAuthConfig{
		{
			JWKSURL:        "https://example.com/.well-known/jwks.json",
			ExpectedIssuer: "https://example.com/oauth2/token",
			ValidClientIDs: []string{"client-id-1"},
		},
		{
			JWKSURL:        "https://auth.example.com/jwks",
			ExpectedIssuer: "https://auth.example.com",
			ValidClientIDs: []string{"client-1", "client-2"},
			OrgName:        "test-org",
		},
		// Allow localhost HTTP for development
		{
			JWKSURL:        "http://localhost:8080/jwks",
			ExpectedIssuer: "http://localhost:8080",
			ValidClientIDs: []string{"local-client"},
		},
		{
			JWKSURL:        "http://127.0.0.1:8080/jwks",
			ExpectedIssuer: "http://127.0.0.1:8080",
			ValidClientIDs: []string{"local-client"},
		},
	}

	for i, config := range validConfigs {
		t.Run(fmt.Sprintf("valid_config_%d", i), func(t *testing.T) {
			err := ValidateJWTAuthConfig(config)
			assert.NoError(t, err, "Valid config should pass validation")
		})
	}
}

// TestValidateJWTAuthConfig_InvalidConfig verifies that invalid configurations fail validation
func TestValidateJWTAuthConfig_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  JWTAuthConfig
		wantErr string
	}{
		{
			name: "missing_jwks_url",
			config: JWTAuthConfig{
				ExpectedIssuer: "https://example.com",
				ValidClientIDs: []string{"client-id"},
			},
			wantErr: "JWKS URL is required",
		},
		{
			name: "missing_issuer",
			config: JWTAuthConfig{
				JWKSURL:        "https://example.com/jwks",
				ValidClientIDs: []string{"client-id"},
			},
			wantErr: "expected issuer is required",
		},
		{
			name: "missing_client_ids",
			config: JWTAuthConfig{
				JWKSURL:        "https://example.com/jwks",
				ExpectedIssuer: "https://example.com",
				ValidClientIDs: []string{},
			},
			wantErr: "at least one valid client ID is required",
		},
		{
			name: "empty_client_id",
			config: JWTAuthConfig{
				JWKSURL:        "https://example.com/jwks",
				ExpectedIssuer: "https://example.com",
				ValidClientIDs: []string{"", "valid-id"},
			},
			wantErr: "client ID at index 0 is empty",
		},
		{
			name: "http_not_localhost",
			config: JWTAuthConfig{
				JWKSURL:        "http://example.com/jwks",
				ExpectedIssuer: "https://example.com",
				ValidClientIDs: []string{"client-id"},
			},
			wantErr: "JWKS URL must use HTTPS",
		},
		{
			name: "http_not_127_0_0_1",
			config: JWTAuthConfig{
				JWKSURL:        "http://192.168.1.1/jwks",
				ExpectedIssuer: "https://example.com",
				ValidClientIDs: []string{"client-id"},
			},
			wantErr: "JWKS URL must use HTTPS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJWTAuthConfig(tt.config)
			require.Error(t, err, "Invalid config should fail validation")
			assert.Contains(t, err.Error(), tt.wantErr, "Error message should contain expected text")
		})
	}
}

// TestNewJWTAuthMiddleware_InvalidConfig verifies that NewJWTAuthMiddleware returns error for invalid config
func TestNewJWTAuthMiddleware_InvalidConfig(t *testing.T) {
	invalidConfig := JWTAuthConfig{
		JWKSURL:        "http://example.com/jwks", // Invalid: not HTTPS and not localhost
		ExpectedIssuer: "https://example.com",
		ValidClientIDs: []string{"client-id"},
	}

	middleware, err := NewJWTAuthMiddleware(invalidConfig)
	assert.Error(t, err, "Should return error for invalid config")
	assert.Nil(t, middleware, "Should not return middleware for invalid config")
	assert.Contains(t, err.Error(), "invalid JWT auth configuration", "Error should indicate configuration issue")
}

// TestNewJWTAuthMiddleware_ValidConfig verifies that NewJWTAuthMiddleware succeeds for valid config
func TestNewJWTAuthMiddleware_ValidConfig(t *testing.T) {
	validConfig := JWTAuthConfig{
		JWKSURL:        "https://example.com/.well-known/jwks.json",
		ExpectedIssuer: "https://example.com/oauth2/token",
		ValidClientIDs: []string{"client-id-1", "client-id-2"},
		OrgName:        "test-org",
		Timeout:        5,
	}

	middleware, err := NewJWTAuthMiddleware(validConfig)
	require.NoError(t, err, "Should not return error for valid config")
	require.NotNil(t, middleware, "Should return middleware for valid config")
	
	// Verify middleware is properly initialized
	assert.NotNil(t, middleware.httpClient, "HTTP client should be initialized")
	assert.Equal(t, validConfig.JWKSURL, middleware.jwksURL, "JWKS URL should be set")
	assert.Equal(t, validConfig.ExpectedIssuer, middleware.expectedIssuer, "Expected issuer should be set")
}

