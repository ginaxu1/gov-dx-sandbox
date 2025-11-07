package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

// APIServerClient handles communication with the API Server
type APIServerClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// SchemaResponse represents a schema from the API Server
type SchemaResponse struct {
	SchemaID          string  `json:"schemaId"`
	MemberID          string  `json:"memberId"`
	SchemaName        string  `json:"schemaName"`
	SDL               string  `json:"sdl"`
	Endpoint          string  `json:"endpoint"`
	Version           string  `json:"version"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

// NewAPIServerClient creates a new API Server client
func NewAPIServerClient(baseURL, apiKey string) *APIServerClient {
	return &APIServerClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSchema fetches a schema by ID from the API Server
func (c *APIServerClient) GetSchema(schemaID string) (*SchemaResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("API Server URL not configured")
	}

	url := fmt.Sprintf("%s/api/v1/schemas/%s", c.baseURL, schemaID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var schema SchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &schema, nil
}

// RegisterSchema registers a new unified schema in the API Server
func (c *APIServerClient) RegisterSchema(schemaID, schemaName, sdl, version string) (*SchemaResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("API Server URL not configured")
	}

	// Create schema request
	createReq := map[string]interface{}{
		"schemaId":   schemaID,
		"schemaName": schemaName,
		"sdl":        sdl,
		"version":    version,
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/schemas", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var schema SchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Log.Info("Schema registered in API Server", "schemaId", schemaID, "version", version)
	return &schema, nil
}

// UpdateSchema updates an existing schema in the API Server
func (c *APIServerClient) UpdateSchema(schemaID, sdl, version string) (*SchemaResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("API Server URL not configured")
	}

	// Update schema request
	updateReq := map[string]interface{}{
		"sdl":     sdl,
		"version": version,
	}

	reqBody, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/schemas/%s", c.baseURL, schemaID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var schema SchemaResponse
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Log.Info("Schema updated in API Server", "schemaId", schemaID, "version", version)
	return &schema, nil
}
