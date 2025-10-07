package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
)

// Provider struct that represents a provider attributes.
type Provider struct {
	Client     *http.Client
	ServiceUrl string           `json:"providerUrl,omitempty"`
	ServiceKey string           `json:"providerKey,omitempty"`
	Auth       *auth.AuthConfig `json:"auth,omitempty"`
	Auth2Token *auth.Auth2TokenResponse
	Headers    map[string]string `json:"headers,omitempty"`
	tokenMu    sync.RWMutex
}

// IsTokenValid checks if the current token is valid (not expired with buffer time)
func (p *Provider) IsTokenValid() bool {
	p.tokenMu.RLock()
	defer p.tokenMu.RUnlock()

	if p.Auth2Token == nil {
		return false
	}

	// Get configuration with defaults
	tokenConfig := p.getTokenConfig()
	bufferTime := tokenConfig.ValidationBuffer
	return time.Now().Add(bufferTime).Before(p.Auth2Token.ExpiresAt)
}

// NeedsTokenRefresh checks if the token needs to be refreshed (expires within buffer time)
func (p *Provider) NeedsTokenRefresh() bool {
	p.tokenMu.RLock()
	defer p.tokenMu.RUnlock()

	if p.Auth2Token == nil {
		return true
	}

	// Get configuration with defaults
	tokenConfig := p.getTokenConfig()
	refreshBuffer := tokenConfig.RefreshBuffer
	return time.Now().Add(refreshBuffer).After(p.Auth2Token.ExpiresAt)
}

// getTokenConfig returns the token configuration with defaults applied
func (p *Provider) getTokenConfig() *auth.TokenConfig {
	if p.Auth == nil || p.Auth.TokenConfig == nil {
		return (&auth.TokenConfig{}).GetTokenConfig()
	}
	return p.Auth.TokenConfig.GetTokenConfig()
}

// GetValidToken ensures we have a valid token, refreshing if necessary
func (p *Provider) GetValidToken() error {
	// Check if we have a token
	if p.Auth2Token == nil {
		// No token - get new one
		logger.Log.Info("No token available, getting new token", "service", p.ServiceKey)
		return p.GetNewTokenWithRetry()
	}

	// Check if token needs refresh
	if p.NeedsTokenRefresh() {
		logger.Log.Info("Token needs refresh", "service", p.ServiceKey, "expires_at", p.Auth2Token.ExpiresAt)

		// Try refresh token first if available
		if p.Auth2Token.RefreshToken != "" {
			if err := p.RefreshTokenWithRetry(); err != nil {
				logger.Log.Warn("Refresh token failed, getting new token", "service", p.ServiceKey, "error", err)
				return p.GetNewTokenWithRetry()
			}
			return nil
		} else {
			// No refresh token - get new token
			logger.Log.Info("No refresh token available, getting new token", "service", p.ServiceKey)
			return p.GetNewTokenWithRetry()
		}
	}

	return nil
}

// GetNewTokenWithRetry gets a new token with retry logic
func (p *Provider) GetNewTokenWithRetry() error {
	tokenConfig := p.getTokenConfig()

	for attempt := 1; attempt <= tokenConfig.MaxRetries; attempt++ {
		if err := p.GetNewToken(); err != nil {
			logger.Log.Warn("Token request failed", "service", p.ServiceKey, "attempt", attempt, "maxRetries", tokenConfig.MaxRetries, "error", err)

			if attempt < tokenConfig.MaxRetries {
				time.Sleep(tokenConfig.RetryDelay)
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("max retries exceeded")
}

// RefreshTokenWithRetry refreshes token with retry logic
func (p *Provider) RefreshTokenWithRetry() error {
	tokenConfig := p.getTokenConfig()

	for attempt := 1; attempt <= tokenConfig.MaxRetries; attempt++ {
		if err := p.RefreshToken(); err != nil {
			logger.Log.Warn("Token refresh failed", "service", p.ServiceKey, "attempt", attempt, "maxRetries", tokenConfig.MaxRetries, "error", err)

			if attempt < tokenConfig.MaxRetries {
				time.Sleep(tokenConfig.RetryDelay)
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("max retries exceeded")
}

// RefreshToken refreshes the access token using the refresh token
func (p *Provider) RefreshToken() error {
	if p.Auth.Type != auth.AuthTypeOAuth2 {
		return fmt.Errorf("[%s] RefreshToken called on non-oauth2 provider", p.ServiceKey)
	}

	p.tokenMu.RLock()
	refreshToken := p.Auth2Token.RefreshToken
	p.tokenMu.RUnlock()

	if refreshToken == "" {
		return fmt.Errorf("[%s] no refresh token available", p.ServiceKey)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.Auth.ClientID)
	data.Set("client_secret", p.Auth.ClientSecret)

	req, err := http.NewRequest("POST", p.Auth.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Log.Error("Failed to create refresh token request", "service", p.ServiceKey, "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.Client.Do(req)
	if err != nil {
		logger.Log.Error("Failed to refresh token", "service", p.ServiceKey, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("Token refresh request failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("[%s] token refresh failed: %s", p.ServiceKey, string(body))
	}

	var res struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"` // seconds
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Log.Error("Failed to decode refresh token response", "error", err)
		return err
	}

	// Update token with thread safety
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	now := time.Now()
	p.Auth2Token = &auth.Auth2TokenResponse{
		AccessToken: res.AccessToken,
		ExpiresAt:   now.Add(time.Duration(res.ExpiresIn) * time.Second),
		TokenType:   res.TokenType,
		Scope:       res.Scope,
		IssuedAt:    now,
	}

	// Update refresh token if provided (some providers issue new refresh tokens)
	if res.RefreshToken != "" {
		p.Auth2Token.RefreshToken = res.RefreshToken
	}

	logger.Log.Info("Token refreshed successfully", "service", p.ServiceKey, "expires_at", p.Auth2Token.ExpiresAt)
	return nil
}

// GetNewToken gets a new access token using client credentials
func (p *Provider) GetNewToken() error {
	if p.Auth.Type != auth.AuthTypeOAuth2 {
		return fmt.Errorf("[%s] GetNewToken called on non-oauth2 provider", p.ServiceKey)
	}

	// Prepare request data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Add scope if provided
	if p.Auth.Scope != "" {
		data.Set("scope", p.Auth.Scope)
	}

	// Create request
	req, err := http.NewRequest("POST", p.Auth.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Log.Error("Failed to create token request", "service", p.ServiceKey, "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set authentication based on method
	authMethod := p.Auth.AuthMethod
	if authMethod == "" {
		authMethod = auth.OAuth2AuthMethodBody // Default to body method
	}

	switch authMethod {
	case auth.OAuth2AuthMethodBody:
		// Send credentials in request body (default)
		data.Set("client_id", p.Auth.ClientID)
		data.Set("client_secret", p.Auth.ClientSecret)
		req.Body = io.NopCloser(strings.NewReader(data.Encode()))

	case auth.OAuth2AuthMethodBasic:
		// Send credentials as Basic Auth header (Asgardeo style)
		credentials := base64.StdEncoding.EncodeToString([]byte(p.Auth.ClientID + ":" + p.Auth.ClientSecret))
		req.Header.Set("Authorization", "Basic "+credentials)

	case auth.OAuth2AuthMethodBearer:
		// Send credentials as Bearer token
		req.Header.Set("Authorization", "Bearer "+p.Auth.ClientSecret)

	default:
		return fmt.Errorf("[%s] unsupported OAuth2 auth method: %s", p.ServiceKey, authMethod)
	}

	logger.Log.Info("Requesting new token", "service", p.ServiceKey, "authMethod", authMethod, "tokenUrl", p.Auth.TokenURL)

	resp, err := p.Client.Do(req)
	if err != nil {
		logger.Log.Error("Failed to fetch token", "service", p.ServiceKey, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("Token request failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("[%s] token request failed: %s", p.ServiceKey, string(body))
	}

	var res struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"` // seconds
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Log.Error("Failed to decode token response", "error", err)
		return err
	}

	// Update token with thread safety
	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	now := time.Now()
	p.Auth2Token = &auth.Auth2TokenResponse{
		AccessToken:  res.AccessToken,
		ExpiresAt:    now.Add(time.Duration(res.ExpiresIn) * time.Second),
		TokenType:    res.TokenType,
		Scope:        res.Scope,
		RefreshToken: res.RefreshToken,
		IssuedAt:     now,
	}

	logger.Log.Info("New token obtained successfully", "service", p.ServiceKey, "expires_at", p.Auth2Token.ExpiresAt, "authMethod", authMethod)
	return nil
}

// PerformRequest performs the HTTP request to the provider with necessary authentication.
func (p *Provider) PerformRequest(reqBody []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", p.ServiceUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if p.Auth != nil {
		switch p.Auth.Type {
		case auth.AuthTypeOAuth2:
			// Use on-demand token management
			if err := p.GetValidToken(); err != nil {
				logger.Log.Error("Failed to get valid token", "service", p.ServiceKey, "error", err)
				return nil, err
			}

			// Get the access token with thread safety
			p.tokenMu.RLock()
			accessToken := p.Auth2Token.AccessToken
			p.tokenMu.RUnlock()

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		case auth.AuthTypeAPIKey:
			req.Header.Set(p.Auth.APIKeyName, p.Auth.APIKeyValue)
		}
	}

	return p.Client.Do(req)
}
