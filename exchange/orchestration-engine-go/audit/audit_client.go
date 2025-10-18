package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

// AuditClient handles communication with the audit service
type AuditClient struct {
	baseURL    string
	httpClient *http.Client
}

// AuditLogRequest represents the request structure for creating audit logs
type AuditLogRequest struct {
	Status        string `json:"status"`
	RequestedData string `json:"requestedData"`
	ApplicationID string `json:"applicationId"`
	SchemaID      string `json:"schemaId"`
}

// AuditLogResponse represents the response from the audit service
type AuditLogResponse struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	RequestedData string    `json:"requestedData"`
	ApplicationID string    `json:"applicationId"`
	SchemaID      string    `json:"schemaId"`
	ConsumerID    string    `json:"consumerId"`
	ProviderID    string    `json:"providerId"`
}

// NewAuditClient creates a new audit client
func NewAuditClient() *AuditClient {
	baseURL := os.Getenv("AUDIT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://audit-service:3001" // Default for Choreo service-to-service communication
	}

	return &AuditClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// BaseURL returns the base URL of the audit service
func (c *AuditClient) BaseURL() string {
	return c.baseURL
}

// LogQuery logs a GraphQL query execution to the audit service
func (c *AuditClient) LogQuery(query string, status string, consumerID string, providerID string) error {
	// If consumerID or providerID are empty, use default values
	if consumerID == "" {
		consumerID = "consumer-123"
	}
	if providerID == "" {
		providerID = "provider-456"
	}

	auditRequest := AuditLogRequest{
		Status:        status,
		RequestedData: query,
		ApplicationID: consumerID,
		SchemaID:      providerID,
	}

	jsonData, err := json.Marshal(auditRequest)
	if err != nil {
		logger.Log.Error("Failed to marshal audit request", "error", err)
		return err
	}

	url := fmt.Sprintf("%s/api/logs", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.Error("Failed to create audit request", "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Failed to send audit request", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read audit response", "error", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("Audit service returned error", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("audit service returned status %d: %s", resp.StatusCode, string(body))
	}

	var auditResponse AuditLogResponse
	if err := json.Unmarshal(body, &auditResponse); err != nil {
		logger.Log.Error("Failed to unmarshal audit response", "error", err)
		return err
	}

	logger.Log.Info("Successfully logged query to audit service",
		"audit_id", auditResponse.ID,
		"status", status,
		"consumer_id", consumerID,
		"provider_id", providerID)

	return nil
}

// LogQueryAsync logs a query asynchronously to avoid blocking the main request
func (c *AuditClient) LogQueryAsync(query string, status string, consumerID string, providerID string) {
	go func() {
		if err := c.LogQuery(query, status, consumerID, providerID); err != nil {
			logger.Log.Error("Failed to log query asynchronously", "error", err)
		}
	}()
}
