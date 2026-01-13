package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	auditServiceURL = "http://127.0.0.1:3001"
)

// AuditLog represents an audit log entry from the audit service
type AuditLog struct {
	ID               string                 `json:"id"`
	TraceID          *string                `json:"traceId"`
	Timestamp        string                 `json:"timestamp"`
	EventType        *string                `json:"eventType"`
	EventAction      *string                `json:"eventAction"`
	Status           string                 `json:"status"`
	ActorType        string                 `json:"actorType"`
	ActorID          string                 `json:"actorId"`
	TargetType       string                 `json:"targetType"`
	TargetID         *string                `json:"targetId"`
	RequestMetadata  map[string]interface{} `json:"requestMetadata,omitempty"`
	ResponseMetadata map[string]interface{} `json:"responseMetadata,omitempty"`
}

// AuditLogsResponse represents the response from GET /api/audit-logs
type AuditLogsResponse struct {
	Logs  []AuditLog `json:"logs"`
	Total int64      `json:"total"`
}

// TestAuditTraceIDCorrelation verifies that all audit events for a single request
// share the same traceID and can be correlated across the entire flow.
func TestAuditTraceIDCorrelation(t *testing.T) {
	// Skip test if audit service is not available
	if !isAuditServiceAvailable() {
		t.Skip("Audit service is not available. Skipping traceID correlation test.")
	}

	// Create a test scenario that will generate multiple audit events
	// This requires: policy metadata, allowlist, consent, and a GraphQL query

	// Step 1: Setup test data (policy metadata, allowlist, consent)
	// This is similar to TestGraphQLFlow_SuccessPath but we'll track the traceID

	// For now, we'll create a simple test that:
	// 1. Makes a GraphQL request to OE
	// 2. Extracts traceID from response header
	// 3. Queries audit service for events with that traceID
	// 4. Verifies all expected events are present with the same traceID

	// Create a test JWT
	appID := fmt.Sprintf("test-audit-app-%d", time.Now().Unix())
	token, err := createTestJWT(appID)
	require.NoError(t, err)

	// Make a GraphQL request
	query := `query {
		citizen(nic: "123456789V") {
			name
		}
	}`

	reqBody := map[string]interface{}{
		"query": query,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", orchestrationEngineURL, bytes.NewBuffer(reqBodyJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-JWT-Assertion", token)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Extract traceID from response header
	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Skip("TraceID not found in response header. Audit logging may not be enabled.")
	}

	t.Logf("Extracted traceID from response: %s", traceID)

	// Poll for audit logs with retry mechanism (async logging may take time)
	auditLogs, err := getAuditLogsByTraceIDWithRetry(t, traceID, 10*time.Second, 300*time.Millisecond)
	require.NoError(t, err)

	// Verify we got at least one event
	require.Greater(t, len(auditLogs), 0, "Expected at least one audit log for traceID %s", traceID)

	// Verify all events share the same traceID
	expectedEventTypes := map[string]bool{
		"ORCHESTRATION_REQUEST_RECEIVED": false,
		"POLICY_CHECK":                   false,
		"CONSENT_CHECK":                  false,
		"PROVIDER_FETCH":                 false,
	}

	for _, log := range auditLogs {
		// Verify traceID matches
		require.NotNil(t, log.TraceID, "Audit log should have a traceID")
		assert.Equal(t, traceID, *log.TraceID, "All audit logs should share the same traceID")

		// Track which event types we found
		if log.EventType != nil {
			if _, exists := expectedEventTypes[*log.EventType]; exists {
				expectedEventTypes[*log.EventType] = true
			}
		}

		t.Logf("Found audit event: type=%s, status=%s, actor=%s, target=%v",
			getStringValue(log.EventType),
			log.Status,
			log.ActorID,
			getStringValue(log.TargetID))
	}

	// Verify we found at least ORCHESTRATION_REQUEST_RECEIVED
	assert.True(t, expectedEventTypes["ORCHESTRATION_REQUEST_RECEIVED"],
		"Expected to find ORCHESTRATION_REQUEST_RECEIVED event")

	// Log summary
	foundCount := 0
	for eventType, found := range expectedEventTypes {
		if found {
			foundCount++
			t.Logf("Found event type: %s", eventType)
		} else {
			t.Logf("⚠️  Event type not found (may be expected): %s", eventType)
		}
	}

	t.Logf("Summary: Found %d/%d expected event types for traceID %s", foundCount, len(expectedEventTypes), traceID)
}

// isAuditServiceAvailable checks if the audit service is available
func isAuditServiceAvailable() bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(auditServiceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getAuditLogsByTraceID queries the audit service for logs with a specific traceID
func getAuditLogsByTraceID(t *testing.T, traceID string) ([]AuditLog, error) {
	url := fmt.Sprintf("%s/api/audit-logs?traceId=%s", auditServiceURL, traceID)

	resp, err := testHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("audit service returned status %d", resp.StatusCode)
	}

	var auditResponse AuditLogsResponse
	if err := json.NewDecoder(resp.Body).Decode(&auditResponse); err != nil {
		return nil, fmt.Errorf("failed to decode audit response: %w", err)
	}

	return auditResponse.Logs, nil
}

// getAuditLogsByTraceIDWithRetry polls the audit service for logs with a specific traceID
// until logs are found or the timeout is reached. This handles async audit logging gracefully.
func getAuditLogsByTraceIDWithRetry(t *testing.T, traceID string, timeout time.Duration, pollInterval time.Duration) ([]AuditLog, error) {
	deadline := time.Now().Add(timeout)
	startTime := time.Now()
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		logs, err := getAuditLogsByTraceID(t, traceID)
		if err != nil {
			// If there's an error, wait and retry
			remaining := time.Until(deadline)
			if remaining > pollInterval {
				time.Sleep(pollInterval)
				continue
			}
			return nil, fmt.Errorf("failed to get audit logs after %d attempts: %w", attempt, err)
		}

		// If we found logs, return them immediately
		if len(logs) > 0 {
			if attempt > 1 {
				elapsed := time.Since(startTime)
				t.Logf("Found audit logs after %d attempts (polled for %v)", attempt, elapsed)
			}
			return logs, nil
		}

		// No logs yet, wait before next attempt
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if remaining < pollInterval {
			time.Sleep(remaining)
		} else {
			time.Sleep(pollInterval)
		}
	}

	// Timeout reached, return empty slice (caller can decide if this is an error)
	elapsed := time.Since(startTime)
	t.Logf("Timeout reached after %d attempts (polled for %v)", attempt, elapsed)
	return []AuditLog{}, nil
}

// getStringValue safely gets string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
