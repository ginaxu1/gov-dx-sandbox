package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

// PDPClient handles communication with the Policy Decision Point
type PDPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPDPClient creates a new PDP client
func NewPDPClient(baseURL string) *PDPClient {
	return &PDPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdateProviderMetadata updates the provider metadata in the PDP
func (c *PDPClient) UpdateProviderMetadata(req models.ProviderMetadataUpdateRequest) (*models.ProviderMetadataUpdateResponse, error) {
	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/metadata/update"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	slog.Debug("Sending metadata update request to PDP", "url", url, "applicationId", req.ApplicationID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PDP: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		slog.Error("PDP returned error", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("PDP returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var response models.ProviderMetadataUpdateResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	slog.Info("Successfully updated provider metadata in PDP", "applicationId", req.ApplicationID, "updated", response.Updated)
	return &response, nil
}

// HealthCheck checks if the PDP is available
func (c *PDPClient) HealthCheck() error {
	url := c.baseURL + "/health"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check PDP health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PDP health check failed with status %d", resp.StatusCode)
	}

	return nil
}
