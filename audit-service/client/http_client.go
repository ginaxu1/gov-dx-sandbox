package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// httpClient implements AuditClient using HTTP calls to the audit service
type httpClient struct {
	baseURL    string
	httpClient *http.Client
}

// newHTTPClient creates a new HTTP client with appropriate timeout
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second, // Short timeout to avoid blocking
	}
}

// LogDataExchange logs a data exchange event to the audit service
// This is called asynchronously (fire-and-forget) to avoid blocking
func (c *httpClient) LogDataExchange(ctx context.Context, event DataExchangeEvent) error {
	// Generate event ID if not provided
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}

	// Generate timestamp if not provided
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Validate required fields
	if event.ConsumerAppID == "" || event.ProviderSchemaID == "" || event.Status == "" {
		return fmt.Errorf("missing required fields: consumerAppId, providerSchemaId, or status")
	}

	// ConsumerID and ProviderID are optional (can be NULL in database)
	// They can be looked up later or tracked via ApplicationID/SchemaID

	// Marshal request
	reqBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal data exchange event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/audit/exchange", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create audit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request asynchronously (fire-and-forget)
	go func() {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to send data exchange audit log", "error", err, "eventId", event.EventID)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			slog.Error("Audit service returned error for data exchange event",
				"status", resp.StatusCode,
				"response", string(bodyBytes),
				"eventId", event.EventID)
			return
		}

		slog.Debug("Data exchange audit log sent successfully",
			"eventId", event.EventID,
			"consumerAppId", event.ConsumerAppID,
			"providerSchemaId", event.ProviderSchemaID,
			"status", event.Status)
	}()

	return nil
}
