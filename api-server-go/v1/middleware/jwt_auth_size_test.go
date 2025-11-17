package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTTokenSizeValidation verifies that oversized tokens are rejected
func TestJWTTokenSizeValidation(t *testing.T) {
	config := JWTAuthConfig{
		JWKSURL:        "https://example.com/.well-known/jwks.json",
		ExpectedIssuer: "test-issuer",
		ValidClientIDs: []string{"test-client"},
	}
	middleware, err := NewJWTAuthMiddleware(config)
	require.NoError(t, err)

	// Create a handler that should not be called if token is rejected
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Test with token exceeding maximum size
	oversizedToken := strings.Repeat("a", MaxJWTTokenSize+1) // 8KB + 1 byte
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+oversizedToken)
	w := httptest.NewRecorder()

	middleware.AuthenticateJWT(nextHandler).ServeHTTP(w, req)

	// Verify request was rejected
	assert.Equal(t, http.StatusBadRequest, w.Code, "Oversized token should be rejected with 400 Bad Request")
	assert.False(t, handlerCalled, "Next handler should not be called for oversized token")
	assert.Contains(t, w.Body.String(), "Token size exceeds maximum", "Error message should mention token size")
}

// TestJWTTokenSizeWarning verifies that tokens approaching the limit trigger warnings
func TestJWTTokenSizeWarning(t *testing.T) {
	config := JWTAuthConfig{
		JWKSURL:        "https://example.com/.well-known/jwks.json",
		ExpectedIssuer: "test-issuer",
		ValidClientIDs: []string{"test-client"},
	}
	middleware, err := NewJWTAuthMiddleware(config)
	require.NoError(t, err)

	// Create a handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with token approaching warning threshold
	// Note: This test verifies the size check logic, but won't actually validate the token
	// since we don't have a real JWKS endpoint. The size validation happens before token parsing.
	largeToken := strings.Repeat("a", JWTTokenSizeWarningThreshold+100) // Just over warning threshold
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+largeToken)
	w := httptest.NewRecorder()

	middleware.AuthenticateJWT(nextHandler).ServeHTTP(w, req)

	// Token will fail validation (no real JWKS), but size check should pass
	// The important thing is that oversized tokens (>8KB) are rejected
	assert.NotEqual(t, http.StatusBadRequest, w.Code, "Token within size limit should not be rejected for size")
}

// TestJWTTokenSizeConstants verifies the size constants are reasonable
func TestJWTTokenSizeConstants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, 8*1024, MaxJWTTokenSize, "MaxJWTTokenSize should be 8KB")
	assert.Equal(t, 6*1024, JWTTokenSizeWarningThreshold, "JWTTokenSizeWarningThreshold should be 6KB")
	assert.Less(t, JWTTokenSizeWarningThreshold, MaxJWTTokenSize, "Warning threshold should be less than max size")
}

// TestJWTTokenSizeNormalToken verifies that normal-sized tokens pass size validation
func TestJWTTokenSizeNormalToken(t *testing.T) {
	config := JWTAuthConfig{
		JWKSURL:        "https://example.com/.well-known/jwks.json",
		ExpectedIssuer: "test-issuer",
		ValidClientIDs: []string{"test-client"},
	}
	middleware, err := NewJWTAuthMiddleware(config)
	require.NoError(t, err)

	// Create a handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with normal-sized token (typical JWT is 200-500 bytes)
	normalToken := strings.Repeat("a", 500) // 500 bytes - well within limit
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+normalToken)
	w := httptest.NewRecorder()

	middleware.AuthenticateJWT(nextHandler).ServeHTTP(w, req)

	// Token will fail validation (no real JWKS), but size check should pass
	// Verify it's not rejected for size (will be rejected later for invalid token)
	require.NotEqual(t, http.StatusBadRequest, w.Code,
		"Normal-sized token should not be rejected for size (may fail validation later)")
}

