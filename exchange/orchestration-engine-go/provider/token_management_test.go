package provider

import (
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
)

func TestIsTokenValid(t *testing.T) {
	provider := &Provider{
		ServiceKey: "test-service",
		Auth: &auth.AuthConfig{
			Type: auth.AuthTypeOAuth2,
		},
	}

	// Test with no token
	if provider.IsTokenValid() {
		t.Error("Expected IsTokenValid to return false when no token is present")
	}

	// Test with valid token
	now := time.Now()
	provider.Auth2Token = &auth.Auth2TokenResponse{
		AccessToken: "test-token",
		ExpiresAt:   now.Add(5 * time.Minute), // Expires in 5 minutes
		IssuedAt:    now,
	}

	if !provider.IsTokenValid() {
		t.Error("Expected IsTokenValid to return true for valid token")
	}

	// Test with expired token
	provider.Auth2Token.ExpiresAt = now.Add(-1 * time.Minute) // Expired 1 minute ago

	if provider.IsTokenValid() {
		t.Error("Expected IsTokenValid to return false for expired token")
	}
}

func TestNeedsTokenRefresh(t *testing.T) {
	provider := &Provider{
		ServiceKey: "test-service",
		Auth: &auth.AuthConfig{
			Type: auth.AuthTypeOAuth2,
		},
	}

	// Test with no token
	if !provider.NeedsTokenRefresh() {
		t.Error("Expected NeedsTokenRefresh to return true when no token is present")
	}

	// Test with token that doesn't need refresh
	now := time.Now()
	provider.Auth2Token = &auth.Auth2TokenResponse{
		AccessToken: "test-token",
		ExpiresAt:   now.Add(5 * time.Minute), // Expires in 5 minutes
		IssuedAt:    now,
	}

	if provider.NeedsTokenRefresh() {
		t.Error("Expected NeedsTokenRefresh to return false for token that doesn't need refresh")
	}

	// Test with token that needs refresh (expires within 2-minute buffer)
	provider.Auth2Token.ExpiresAt = now.Add(1 * time.Minute) // Expires in 1 minute

	if !provider.NeedsTokenRefresh() {
		t.Error("Expected NeedsTokenRefresh to return true for token that needs refresh")
	}
}

func TestGetTokenConfig(t *testing.T) {
	provider := &Provider{
		ServiceKey: "test-service",
	}

	// Test with no auth config
	config := provider.getTokenConfig()
	if config.RefreshBuffer != 2*time.Minute {
		t.Errorf("Expected default RefreshBuffer to be 2 minutes, got %v", config.RefreshBuffer)
	}
	if config.ValidationBuffer != 30*time.Second {
		t.Errorf("Expected default ValidationBuffer to be 30 seconds, got %v", config.ValidationBuffer)
	}
	if config.MaxRetries != 3 {
		t.Errorf("Expected default MaxRetries to be 3, got %d", config.MaxRetries)
	}
	if config.RetryDelay != 5*time.Second {
		t.Errorf("Expected default RetryDelay to be 5 seconds, got %v", config.RetryDelay)
	}

	// Test with custom config
	provider.Auth = &auth.AuthConfig{
		Type: auth.AuthTypeOAuth2,
		TokenConfig: &auth.TokenConfig{
			RefreshBuffer:    1 * time.Minute,
			ValidationBuffer: 15 * time.Second,
			MaxRetries:       5,
			RetryDelay:       2 * time.Second,
		},
	}

	config = provider.getTokenConfig()
	if config.RefreshBuffer != 1*time.Minute {
		t.Errorf("Expected custom RefreshBuffer to be 1 minute, got %v", config.RefreshBuffer)
	}
	if config.ValidationBuffer != 15*time.Second {
		t.Errorf("Expected custom ValidationBuffer to be 15 seconds, got %v", config.ValidationBuffer)
	}
	if config.MaxRetries != 5 {
		t.Errorf("Expected custom MaxRetries to be 5, got %d", config.MaxRetries)
	}
	if config.RetryDelay != 2*time.Second {
		t.Errorf("Expected custom RetryDelay to be 2 seconds, got %v", config.RetryDelay)
	}
}
