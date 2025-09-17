package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	apiServerURL string
	httpClient   *http.Client
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Valid      bool   `json:"valid"`
	ConsumerID string `json:"consumerId,omitempty"`
	Error      string `json:"error,omitempty"`
}

func NewClient() *Client {
	apiServerURL := os.Getenv("API_SERVER_URL")
	if apiServerURL == "" {
		apiServerURL = "http://localhost:3000"
	}

	return &Client{
		apiServerURL: apiServerURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken validates an access token with the API server
func (c *Client) ValidateToken(token string) (*ValidateTokenResponse, error) {
	reqBody := ValidateTokenRequest{
		Token: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.apiServerURL + "/auth/validate"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &ValidateTokenResponse{
			Valid: false,
			Error: fmt.Sprintf("API server returned status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// The API server returns the response directly, not wrapped in a success structure
	var validateResp ValidateTokenResponse
	if err := json.Unmarshal(body, &validateResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Debug logging
	fmt.Printf("DEBUG: Token validation response: %+v\n", validateResp)

	return &validateResp, nil
}

// GetConsumerIDFromToken extracts consumer ID from a valid token
func (c *Client) GetConsumerIDFromToken(token string) (string, error) {
	response, err := c.ValidateToken(token)
	if err != nil {
		return "", err
	}

	if !response.Valid {
		return "", fmt.Errorf("invalid token: %s", response.Error)
	}

	return response.ConsumerID, nil
}
