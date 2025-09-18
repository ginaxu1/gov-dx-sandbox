package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
}

// StartTokenRefresh starts a goroutine to refresh the OAuth2 token periodically.
func (p *Provider) StartTokenRefresh() {
	go func(prov *Provider) {
		for {
			// If no token or expired, fetch a new one
			if prov.Auth2Token == nil || time.Now().After(prov.Auth2Token.ExpiresAt) {
				logger.Log.Info("Refreshing Token", "service", prov.ServiceKey)
				if err := prov.fetchToken(); err != nil {
					logger.Log.Error("Failed to refresh token", "service", prov.ServiceKey, "error", err)
					// Wait a bit before retrying on failure
					time.Sleep(10 * time.Second)
					continue
				}
				// print the token response
				logger.Log.Info("Refreshing Token", "service", prov.ServiceKey)
				logger.Log.Info("Token Refreshed", "Token Expires At:", prov.Auth2Token.ExpiresAt)
			}

			// Sleep until just before expiry (or re-check every 30s if very long expiry)
			sleepFor := time.Until(prov.Auth2Token.ExpiresAt.Add(-1 * time.Minute))
			if sleepFor < 30*time.Second {
				sleepFor = 30 * time.Second
			}
			time.Sleep(sleepFor)
		}
	}(p)
}

func (p *Provider) fetchToken() error {
	if p.Auth.Type != auth.AuthTypeOAuth2 {
		return fmt.Errorf("[%s] fetchToken called on non-oauth2 provider", p.ServiceKey)
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", p.Auth.ClientID)
	data.Set("client_secret", p.Auth.ClientSecret)

	req, err := http.NewRequest("POST", p.Auth.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Log.Error("Failed to create token request", "service", p.ServiceKey, "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"` // seconds
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Log.Error("Failed to decode token response", "error", err)
		return err
	}

	p.Auth2Token = &auth.Auth2TokenResponse{
		AccessToken: res.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(res.ExpiresIn) * time.Second),
		TokenType:   res.TokenType,
	}

	return nil
}

// PerformRequest performs the HTTP request to the provider with necessary authentication.
func (p *Provider) PerformRequest(reqBody []byte) (*http.Response, error) {
	// Add auth headers if needed

	req, err := http.NewRequest("POST", p.ServiceUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if p.Auth != nil {
		switch p.Auth.Type {
		case auth.AuthTypeOAuth2:
			if p.Auth2Token == nil || p.Auth2Token.AccessToken == "" {
				if err := p.fetchToken(); err != nil {
					return nil, err
				}
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Auth2Token.AccessToken))
		case auth.AuthTypeAPIKey:
			req.Header.Set(p.Auth.APIKeyName, p.Auth.APIKeyValue)
		}
	}

	return p.Client.Do(req)
}
