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

// LogManagementEvent logs a management event to the audit service
// This is called asynchronously (fire-and-forget) to avoid blocking
func (c *httpClient) LogManagementEvent(ctx context.Context, event ManagementEventRequest) error {
	// Generate event ID if not provided
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}

	// Generate timestamp if not provided
	if event.Timestamp == nil {
		timestamp := time.Now().UTC().Format(time.RFC3339)
		event.Timestamp = &timestamp
	}

	// Validate required fields
	if event.EventType == "" || event.Actor.Type == "" || event.Target.Resource == "" {
		return fmt.Errorf("missing required fields: eventType, actor.type, or target.resource")
	}

	// ResourceID is optional for CREATE failures (when status is FAILURE and eventType is CREATE)
	// For other operations (UPDATE, DELETE) or SUCCESS status, ResourceID should be provided
	if event.Target.ResourceID == "" {
		if event.EventType != "CREATE" || event.Status != "FAILURE" {
			return fmt.Errorf("missing required field: target.resourceId (required for UPDATE/DELETE operations or SUCCESS status)")
		}
		// Allow empty ResourceID for CREATE failures
	}

	// Validate actor fields based on type
	if event.Actor.Type == "USER" {
		if event.Actor.ID == nil || *event.Actor.ID == "" {
			return fmt.Errorf("missing actor.id for USER type management event")
		}
		if event.Actor.Role == nil || (*event.Actor.Role != "MEMBER" && *event.Actor.Role != "ADMIN") {
			return fmt.Errorf("missing or invalid actor.role for USER type management event")
		}
	}

	// Marshal request (event is already ManagementEventRequest type via type alias)
	reqBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal management event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/events", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create audit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request asynchronously (fire-and-forget)
	go func() {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			slog.Error("Failed to send management event audit log",
				"error", err,
				"eventId", event.EventID,
				"eventType", event.EventType)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			slog.Error("Audit service returned error for management event",
				"status", resp.StatusCode,
				"response", string(bodyBytes),
				"eventId", event.EventID)
			return
		}

		slog.Debug("Management event audit log sent successfully",
			"eventId", event.EventID,
			"eventType", event.EventType,
			"targetResource", event.Target.Resource,
			"targetResourceId", event.Target.ResourceID)
	}()

	return nil
}
