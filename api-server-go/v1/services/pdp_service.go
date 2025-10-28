package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/utils"
)

// PDPService handles communication with the Policy Decision Point
type PDPService struct {
	// baseURL is the endpoint of the PDP
	baseURL string
	// HTTPClient is used to make requests to the PDP
	HTTPClient *http.Client
}

// NewPDPService creates a new instance of PDPService with optimized HTTP client
func NewPDPService(baseURL string) *PDPService {
	// Create optimized HTTP client with connection pooling
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // Increased timeout for policy operations
	}

	return &PDPService{
		baseURL:    baseURL,
		HTTPClient: client,
	}
}

// executeWithRetry executes a function with exponential backoff retry logic
func (s *PDPService) executeWithRetry(operation func() error) error {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		// Don't retry on the last attempt
		if attempt == maxRetries {
			return err
		}

		// Calculate delay with exponential backoff
		delay := baseDelay * time.Duration(1<<uint(attempt))
		slog.Warn("PDP operation failed, retrying",
			"attempt", attempt+1,
			"maxRetries", maxRetries+1,
			"delay", delay,
			"error", err)

		time.Sleep(delay)
	}

	return fmt.Errorf("operation failed after %d attempts", maxRetries+1)
}

// HealthCheck checks the health of the PDP service with retry logic
func (s *PDPService) HealthCheck() error {
	return s.executeWithRetry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/health", nil)
		if err != nil {
			return fmt.Errorf("failed to create health check request: %w", err)
		}

		resp, err := s.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to execute health check request: %w", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				slog.Error("failed to close response body", "error", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("PDP service health check failed with status: %s", resp.Status)
		}

		return nil
	})
}

// CreatePolicyMetadata sends a request to create policy metadata in the PDP
func (s *PDPService) CreatePolicyMetadata(schemaId string, sdl string) (*models.PolicyMetadataCreateResponse, error) {
	// parse SDL and create policy metadata request
	handler := utils.NewGraphQLHandler()
	policyRequest, err := handler.ParseSDLToPolicyRequest(schemaId, sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(policyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/policy/metadata", s.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request to PDP
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PDP: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp.Body)

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
	var response models.PolicyMetadataCreateResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	slog.Info("Successfully created policy metadata in PDP", "schemaId", schemaId, "recordsCreated", len(response.Records))
	return &response, nil
}

// UpdateAllowList sends a request to update the allow list in the PDP
func (s *PDPService) UpdateAllowList(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
	// Marshal request to JSON
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/policy/update-allowlist", s.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	slog.Debug("Sending allow list update request to PDP", "url", url, "applicationId", request.ApplicationID)
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PDP: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp.Body)

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
	var response models.AllowListUpdateResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	slog.Info("Successfully updated allow list in PDP", "applicationId", request.ApplicationID, "recordsUpdated", len(response.Records))
	return &response, nil
}
