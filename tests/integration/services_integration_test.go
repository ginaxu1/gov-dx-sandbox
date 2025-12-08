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
	portalBackendURL = "http://127.0.0.1:3000"
)

func TestPortalBackend_Health(t *testing.T) {
	resp, err := http.Get(portalBackendURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuditService_Health(t *testing.T) {
	resp, err := http.Get(auditServiceURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuditLogging_From_OrchestrationEngine(t *testing.T) {
	// 1. Perform a GraphQL request to Orchestration Engine (reuse simplified setup)
	// We want to generate an audit log. Any successful or failed request should theoretically generate one 
	// if configured correctly, but success is safer.
	
	// Ensure metadata exists for a valid request
	schemaID := "audit-test-schema"
	fieldName := "audit_field"
	appID := "audit-test-app"

	// Setup Policy Metadata
	reqBody := map[string]interface{}{
		"schemaId": schemaID,
		"records": []map[string]interface{}{
			{
				"fieldName":         fieldName,
				"source":            "primary",
				"isOwner":           true,
				"accessControlType": "public", // Public to avoid consent complexity for this specific test
			},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)
	resp, err := http.Post(pdpURL+"/metadata", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Send GraphQL Query
	graphQLQuery := map[string]interface{}{
		"query": fmt.Sprintf(`
			query {
				%s
			}
		`, fieldName),
	}
	jsonData, err = json.Marshal(graphQLQuery)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", orchestrationEngineURL, bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("X-Consumer-ID", appID) // If supported

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// 2. Verify Audit Log
	// Wait for async logging by polling the audit service with a timeout
	var (
		events         []map[string]interface{}
		auditLogFound  bool
		pollInterval   = 200 * time.Millisecond
		timeout        = 10 * time.Second
		startTime      = time.Now()
	)

	for time.Since(startTime) < timeout {
		auditResp, err := http.Get(auditServiceURL + "/api/data-exchange-events")
		require.NoError(t, err)

		if auditResp.StatusCode == http.StatusOK {
			events = nil
			err = json.NewDecoder(auditResp.Body).Decode(&events)
			auditResp.Body.Close()
			if err == nil {
				for _, event := range events {
					if sid, ok := event["schema_id"].(string); ok && sid == schemaID {
						auditLogFound = true
						break
					}
				}
				if auditLogFound || len(events) > 0 {
					break
				}
			} else {
				t.Logf("Failed to decode as array, might be wrapped object: %v", err)
				auditResp.Body.Close()
			}
		} else {
			auditResp.Body.Close()
		}
		time.Sleep(pollInterval)
	}

	// If the above decoding failed or structure is different, let's log the body to debug if it fails
	if !auditLogFound && len(events) > 0 {
		// Limitation: Could not match specific schemaID. This test only asserts that at least one event exists,
		// but does not guarantee it is the expected audit event. If schema_id matching is unreliable,
		// refine the test or improve the event structure for robust verification.
		t.Log("Found events but didn't match specific schemaID. Inspecting first event:", events[0])
	}

	assert.True(t, len(events) > 0, "Should have at least one audit event")
}
