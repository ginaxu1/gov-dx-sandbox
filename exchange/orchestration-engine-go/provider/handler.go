package provider

import (
	"net/http"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/auth"
)

// Handler is the main struct that holds all the provider handling information
type Handler struct {
	Providers  map[string]*Provider
	HttpClient *http.Client
}

// NewProviderHandler creates a new ProviderHandler with the given providers.
func NewProviderHandler(providers []*Provider) *Handler {
	providerMap := make(map[string]*Provider)

	// Create an http client with a 10 seconds timeout
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: http.DefaultTransport,
	}

	for _, p := range providers {
		if p != nil {
			providerMap[p.ServiceKey] = p
			p.Client = httpClient
		}
	}

	return &Handler{
		Providers:  providerMap,
		HttpClient: httpClient,
	}
}

// GetProvider retrieves a provider by its service key.
func (h *Handler) GetProvider(serviceKey string) (*Provider, bool) {
	p, exists := h.Providers[serviceKey]
	return p, exists
}

// AddProvider adds a new provider to the handler.
func (h *Handler) AddProvider(serviceKey string, provider *Provider) {
	h.Providers[serviceKey] = provider
	provider.Client = h.HttpClient
}

// StartTokenRefreshProcess starts the token refresh process for all providers that use OAuth2 authentication.
func (h *Handler) StartTokenRefreshProcess() {
	for _, p := range h.Providers {
		if p != nil && p.Auth != nil && p.Auth.Type == auth.AuthTypeOAuth2 {
			p.StartTokenRefresh()
		}
	}
}
