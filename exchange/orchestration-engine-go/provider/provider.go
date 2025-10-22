package provider

import (
	"bytes"
	"context"
	"net/http"
	"sync"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// Provider struct that represents a provider attributes.
type Provider struct {
	Client       *http.Client
	ServiceUrl   string           `json:"providerUrl,omitempty"`
	ServiceKey   string           `json:"providerKey,omitempty"`
	Auth         *auth.AuthConfig `json:"auth,omitempty"`
	OAuth2Config *clientcredentials.Config
	Headers      map[string]string `json:"headers,omitempty"`
	tokenMu      sync.RWMutex
}

// PerformRequest performs the HTTP request to the provider with necessary authentication.
func (p *Provider) PerformRequest(ctx context.Context, reqBody []byte) (*http.Response, error) {
	// 1. Create Request
	req, err := http.NewRequest("POST", p.ServiceUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if p.Auth != nil {
		switch p.Auth.Type {
		case auth.AuthTypeOAuth2:
			client := p.OAuth2Config.Client(ctx)
			return client.Do(req.WithContext(ctx)) // Use context with request
		case auth.AuthTypeAPIKey:
			req.Header.Set(p.Auth.APIKeyName, p.Auth.APIKeyValue)
		}
	}

	// Default client execution (for API Key or no auth)
	return p.Client.Do(req.WithContext(ctx))
}
