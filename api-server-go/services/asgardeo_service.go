package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

// AsgardeoService handles all Asgardeo operations
type AsgardeoService struct {
	baseURL            string
	httpClient         *http.Client
	tokenEndpoint      string
	introspectEndpoint string
}

// TokenResponse represents OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// TokenIntrospection represents token introspection response
type TokenIntrospection struct {
	Active   bool   `json:"active"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// NewAsgardeoService creates a new Asgardeo service
func NewAsgardeoService(baseURL string) *AsgardeoService {
	return &AsgardeoService{
		baseURL:            baseURL,
		httpClient:         &http.Client{Timeout: 30 * time.Second},
		tokenEndpoint:      baseURL + "/oauth2/token",
		introspectEndpoint: baseURL + "/oauth2/introspect",
	}
}

// ExchangeCredentialsForToken exchanges API key/secret for Asgardeo access token
func (s *AsgardeoService) ExchangeCredentialsForToken(apiKey, apiSecret string) (*TokenResponse, error) {
	formData := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     apiKey,
		"client_secret": apiSecret,
		"scope":         "gov-dx-api",
	}

	reqBody := s.createFormData(formData)
	req, err := http.NewRequest("POST", s.tokenEndpoint, bytes.NewBufferString(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	slog.Info("Successfully exchanged credentials for Asgardeo token",
		"apiKey", apiKey,
		"tokenType", tokenResp.TokenType,
		"expiresIn", tokenResp.ExpiresIn)

	return &tokenResp, nil
}

// ValidateToken validates an access token using introspection
func (s *AsgardeoService) ValidateToken(accessToken string) (*models.ValidateTokenResponse, error) {
	formData := map[string]string{
		"token": accessToken,
	}

	reqBody := s.createFormData(formData)
	req, err := http.NewRequest("POST", s.introspectEndpoint, bytes.NewBufferString(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add basic authentication for introspection
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	if clientID != "" && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make introspection request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("introspection request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var introResp TokenIntrospection
	if err := json.NewDecoder(resp.Body).Decode(&introResp); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	if !introResp.Active {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Token is not active",
		}, nil
	}

	// Check if token is expired
	if introResp.Exp > 0 && time.Now().Unix() > introResp.Exp {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Token has expired",
		}, nil
	}

	return &models.ValidateTokenResponse{
		Valid:      true,
		ConsumerID: introResp.ClientID, // This will be the Asgardeo client ID
	}, nil
}

// createFormData creates URL-encoded form data from a map
func (s *AsgardeoService) createFormData(data map[string]string) string {
	var formData bytes.Buffer
	first := true

	for key, value := range data {
		if !first {
			formData.WriteString("&")
		}
		formData.WriteString(fmt.Sprintf("%s=%s", key, value))
		first = false
	}

	return formData.String()
}
