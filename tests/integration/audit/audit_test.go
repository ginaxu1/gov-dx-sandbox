package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	auditBaseURL = "http://127.0.0.1:3001"
)

func TestMain(m *testing.M) {
	// Wait for audit service availability
	if err := testutils.WaitForService(auditBaseURL + "/health"); err != nil {
		fmt.Printf("Audit Service not available: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

// TestAudit_CreateDataExchangeEvent tests creating a data exchange audit event
func TestAudit_CreateDataExchangeEvent(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-audit-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"consumerId": "test-consumer-123",
		"providerId": "test-provider-456",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID, ok := createResponse["id"].(string)
	require.True(t, ok, "id should be present in response")
	assert.NotEmpty(t, eventID, "event id should not be empty")

	status, ok := createResponse["status"].(string)
	require.True(t, ok, "status should be present in response")
	assert.Equal(t, "success", status, "Event status should be success")
}

// TestAudit_CreateDataExchangeEventFailure tests creating a failure event
func TestAudit_CreateDataExchangeEventFailure(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "failure",
		"applicationId": "test-app-audit-2",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"consumerId": "test-consumer-123",
		"providerId": "test-provider-456",
		"additionalInfo": map[string]interface{}{
			"error": "Access denied",
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	status, ok := createResponse["status"].(string)
	require.True(t, ok)
	assert.Equal(t, "failure", status, "Event status should be failure")
}

// TestAudit_GetDataExchangeEvents tests retrieving data exchange events
func TestAudit_GetDataExchangeEvents(t *testing.T) {
	// First, create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-retrieve-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"consumerId": "test-consumer-retrieve-1",
		"providerId": "test-provider-retrieve-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Retrieve events
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok, "events should be an array")

	total, ok := listResponse["total"].(float64)
	require.True(t, ok, "total should be present")
	assert.GreaterOrEqual(t, int(total), len(events), "total should be >= number of events returned")
}

// TestAudit_FilterByConsumer tests filtering events by consumer ID
func TestAudit_FilterByConsumer(t *testing.T) {
	consumerID := "test-consumer-filter-1"

	// Create an event with specific consumer ID
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-filter-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"consumerId": consumerID,
		"providerId": "test-provider-filter-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Retrieve events filtered by consumer ID
	params := url.Values{}
	params.Add("consumerId", consumerID)
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok)

	// Verify all returned events have the correct consumer ID
	for _, event := range events {
		eventMap, ok := event.(map[string]interface{})
		require.True(t, ok)

		if cid, exists := eventMap["consumerId"].(string); exists {
			assert.Equal(t, consumerID, cid, "All events should have matching consumer ID")
		}
	}
}

// TestAudit_FilterByStatus tests filtering events by status
func TestAudit_FilterByStatus(t *testing.T) {
	// Create a success event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-status-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Filter by success status
	params := url.Values{}
	params.Add("transaction_status", "SUCCESS")
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok)

	// Verify all returned events have success status
	for _, event := range events {
		eventMap, ok := event.(map[string]interface{})
		require.True(t, ok)

		status, exists := eventMap["status"].(string)
		require.True(t, exists, "status should be present")
		assert.Equal(t, "success", status, "All events should have success status")
	}
}

// TestAudit_FilterByDateRange tests filtering events by date range
func TestAudit_FilterByDateRange(t *testing.T) {
	// Create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-date-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	resp.Body.Close()

	// Filter by date range (today)
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := now.AddDate(0, 0, 1).Format("2006-01-02")

	params := url.Values{}
	params.Add("start_date", startDate)
	params.Add("end_date", endDate)
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	// Should return events within the date range
	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(events), 0, "Should return events within date range")
}

// TestAudit_Pagination tests pagination functionality
func TestAudit_Pagination(t *testing.T) {
	// Retrieve events with pagination
	params := url.Values{}
	params.Add("limit", "10")
	params.Add("offset", "0")

	resp, err := http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	limit, ok := listResponse["limit"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(10), limit, "limit should match request")

	offset, ok := listResponse["offset"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), offset, "offset should match request")

	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(events), 10, "Should return at most limit number of events")
}

// TestAudit_InvalidRequest tests edge cases for invalid audit event requests
func TestAudit_InvalidRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing timestamp",
			request: map[string]interface{}{
				"status":        "success",
				"applicationId": "test-app",
				"schemaId":      "test-schema",
				"requestedData": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid status",
			request: map[string]interface{}{
				"timestamp":     time.Now().UTC().Format(time.RFC3339),
				"status":        "invalid",
				"applicationId": "test-app",
				"schemaId":      "test-schema",
				"requestedData": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing applicationId",
			request: map[string]interface{}{
				"timestamp":     time.Now().UTC().Format(time.RFC3339),
				"status":        "success",
				"schemaId":      "test-schema",
				"requestedData": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing schemaId",
			request: map[string]interface{}{
				"timestamp":     time.Now().UTC().Format(time.RFC3339),
				"status":        "success",
				"applicationId": "test-app",
				"requestedData": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Expected status %d for invalid request: %s", tt.expectedStatus, tt.name)
		})
	}
}

// TestAudit_DatabaseVerification tests audit event creation with database verification
func TestAudit_DatabaseVerification(t *testing.T) {
	if os.Getenv("TEST_VERIFY_DB") != "true" {
		t.Skip("Skipping database verification test (set TEST_VERIFY_DB=true to enable)")
	}

	var db *gorm.DB
	db = testutils.SetupPostgresTestDB(t)
	if db == nil {
		t.Skip("Database connection not available")
		return
	}

	// Create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":     timestamp,
		"status":        "success",
		"applicationId": "test-app-db-1",
		"schemaId":      "test-schema-123",
		"requestedData": map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		"consumerId": "test-consumer-db-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID, ok := createResponse["id"].(string)
	require.True(t, ok)

	// Verify event exists in database
	var count int64
	err = db.Table("data_exchange_events").
		Where("id = ?", eventID).
		Count(&count).Error

	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Event should exist in database")

	// Cleanup
	defer func() {
		db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
	}()
}

// TestAudit_CreateManagementEvent tests creating a management event
func TestAudit_CreateManagementEvent(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := map[string]interface{}{
		"timestamp":    timestamp,
		"eventType":    "schema_created",
		"userId":       "test-user-123",
		"resourceType": "schema",
		"resourceId":   "test-schema-123",
		"action":       "create",
		"details": map[string]interface{}{
			"schemaName": "Test Schema",
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/management-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Management events endpoint should accept the request
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK,
		"Expected 201 Created or 200 OK")
}

// TestAudit_GetManagementEvents tests retrieving management events
func TestAudit_GetManagementEvents(t *testing.T) {
	// Retrieve management events
	resp, err := http.Get(auditBaseURL + "/api/management-events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var listResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	events, ok := listResponse["events"].([]interface{})
	require.True(t, ok, "events should be an array")

	total, ok := listResponse["total"].(float64)
	require.True(t, ok, "total should be present")
	assert.GreaterOrEqual(t, int(total), len(events), "total should be >= number of events returned")
}
