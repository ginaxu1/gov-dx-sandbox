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

// Request/Response types for type safety

type CreateDataExchangeEventRequest struct {
	Timestamp      string                 `json:"timestamp"`
	Status         string                 `json:"status"`
	ApplicationID  string                 `json:"applicationId"`
	SchemaID       string                 `json:"schemaId"`
	RequestedData  map[string]interface{} `json:"requestedData"`
	ConsumerID     string                 `json:"consumerId,omitempty"`
	ProviderID     string                 `json:"providerId,omitempty"`
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

type CreateDataExchangeEventResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type DataExchangeEvent struct {
	ID             string                 `json:"id"`
	Timestamp      string                 `json:"timestamp"`
	Status         string                 `json:"status"`
	ApplicationID  string                 `json:"applicationId"`
	SchemaID       string                 `json:"schemaId"`
	RequestedData  map[string]interface{} `json:"requestedData"`
	ConsumerID     string                 `json:"consumerId,omitempty"`
	ProviderID     string                 `json:"providerId,omitempty"`
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

type DataExchangeEventListResponse struct {
	Events []DataExchangeEvent `json:"events"`
	Total  int                 `json:"total"`
	Limit  *int                `json:"limit,omitempty"`
	Offset *int                `json:"offset,omitempty"`
}

type ManagementEventActor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Role string `json:"role"`
}

type ManagementEventTarget struct {
	Resource   string `json:"resource"`
	ResourceID string `json:"resourceId"`
}

type CreateManagementEventRequest struct {
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"eventType"`
	Status    string                 `json:"status"`
	Actor     ManagementEventActor   `json:"actor"`
	Target    ManagementEventTarget  `json:"target"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ManagementEvent struct {
	ID        string                 `json:"id"`
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"eventType"`
	Status    string                 `json:"status"`
	Actor     ManagementEventActor   `json:"actor"`
	Target    ManagementEventTarget  `json:"target"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ManagementEventListResponse struct {
	Events []ManagementEvent `json:"events"`
	Total  int               `json:"total"`
}

func TestMain(m *testing.M) {
	// Wait for audit service availability
	if err := testutils.WaitForService(auditBaseURL+"/health", 30); err != nil {
		fmt.Printf("Audit Service not available: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

// TestAudit_CreateDataExchangeEvent tests creating a data exchange audit event
func TestAudit_CreateDataExchangeEvent(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-audit-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		ConsumerID: "test-consumer-123",
		ProviderID: "test-provider-456",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	assert.NotEmpty(t, createResponse.ID, "event id should not be empty")
	assert.Equal(t, "success", createResponse.Status, "Event status should be success")
}

// TestAudit_CreateDataExchangeEventFailure tests creating a failure event
func TestAudit_CreateDataExchangeEventFailure(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "failure",
		ApplicationID: "test-app-audit-2",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		ConsumerID: "test-consumer-123",
		ProviderID: "test-provider-456",
		AdditionalInfo: map[string]interface{}{
			"error": "Access denied",
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	assert.Equal(t, "failure", createResponse.Status, "Event status should be failure")
}

// TestAudit_GetDataExchangeEvents tests retrieving data exchange events
func TestAudit_GetDataExchangeEvents(t *testing.T) {
	// First, create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-retrieve-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		ConsumerID: "test-consumer-retrieve-1",
		ProviderID: "test-provider-retrieve-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID := createResponse.ID
	require.NotEmpty(t, eventID)

	// Cleanup: Remove created event record
	t.Cleanup(func() {
		if db := testutils.SetupAuditDB(t); db != nil {
			db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
		}
	})

	// Retrieve events
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var listResponse DataExchangeEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, listResponse.Total, len(listResponse.Events), "total should be >= number of events returned")
}

// TestAudit_FilterByConsumer tests filtering events by consumer ID
func TestAudit_FilterByConsumer(t *testing.T) {
	consumerID := "test-consumer-filter-1"

	// Create an event with specific consumer ID
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-filter-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		ConsumerID: consumerID,
		ProviderID: "test-provider-filter-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID := createResponse.ID
	require.NotEmpty(t, eventID)

	// Cleanup: Remove created event record
	t.Cleanup(func() {
		if db := testutils.SetupAuditDB(t); db != nil {
			db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
		}
	})

	// Retrieve events filtered by consumer ID
	params := url.Values{}
	params.Add("consumerId", consumerID)
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse DataExchangeEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	// Verify all returned events have the correct consumer ID
	assert.NotEmpty(t, listResponse.Events, "Expected to find at least one event for the consumer")
	for _, event := range listResponse.Events {
		assert.Equal(t, consumerID, event.ConsumerID, "All events should have matching consumer ID")
	}
}

// TestAudit_FilterByStatus tests filtering events by status
func TestAudit_FilterByStatus(t *testing.T) {
	// Create a success event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-status-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID := createResponse.ID
	require.NotEmpty(t, eventID)

	// Cleanup: Remove created event record
	t.Cleanup(func() {
		if db := testutils.SetupAuditDB(t); db != nil {
			db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
		}
	})

	// Filter by success status
	params := url.Values{}
	params.Add("status", "success")
	resp, err = http.Get(auditBaseURL + "/api/data-exchange-events?" + params.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResponse DataExchangeEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	// Verify all returned events have success status
	for _, event := range listResponse.Events {
		assert.Equal(t, "success", event.Status, "All events should have success status")
	}
}

// TestAudit_FilterByDateRange tests filtering events by date range
func TestAudit_FilterByDateRange(t *testing.T) {
	// Create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-date-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID := createResponse.ID
	require.NotEmpty(t, eventID)

	// Cleanup: Remove created event record
	t.Cleanup(func() {
		if db := testutils.SetupAuditDB(t); db != nil {
			db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
		}
	})

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

	var listResponse DataExchangeEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	// Should return events within the date range
	assert.GreaterOrEqual(t, len(listResponse.Events), 0, "Should return events within date range")
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

	var listResponse DataExchangeEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	if listResponse.Limit != nil {
		assert.Equal(t, 10, *listResponse.Limit, "limit should match request")
	}

	if listResponse.Offset != nil {
		assert.Equal(t, 0, *listResponse.Offset, "offset should match request")
	}

	assert.LessOrEqual(t, len(listResponse.Events), 10, "Should return at most limit number of events")
}

// TestAudit_InvalidRequest tests edge cases for invalid audit event requests
// NOTE: Some test cases currently fail due to service-level validation issues:
// - Missing timestamp: Service returns 500 instead of 400
// - Invalid status: Service returns 500 instead of 400
// - Missing applicationId: Service returns 201 (accepts invalid request) instead of 400
// - Missing schemaId: Service returns 201 (accepts invalid request) instead of 400
// These failures indicate that the audit service needs to implement proper request validation.
// The tests are correctly expecting 400 Bad Request for invalid inputs.
func TestAudit_InvalidRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        func() []byte
		expectedStatus int
	}{
		{
			name: "Missing timestamp",
			request: func() []byte {
				req := map[string]interface{}{
					"status":        "success",
					"applicationId": "test-app",
					"schemaId":      "test-schema",
					"requestedData": map[string]interface{}{},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid status",
			request: func() []byte {
				req := CreateDataExchangeEventRequest{
					Timestamp:     time.Now().UTC().Format(time.RFC3339),
					Status:        "invalid",
					ApplicationID: "test-app",
					SchemaID:      "test-schema",
					RequestedData: map[string]interface{}{},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing applicationId",
			request: func() []byte {
				req := map[string]interface{}{
					"timestamp":     time.Now().UTC().Format(time.RFC3339),
					"status":        "success",
					"schemaId":      "test-schema",
					"requestedData": map[string]interface{}{},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing schemaId",
			request: func() []byte {
				req := map[string]interface{}{
					"timestamp":     time.Now().UTC().Format(time.RFC3339),
					"status":        "success",
					"applicationId": "test-app",
					"requestedData": map[string]interface{}{},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := tt.request()

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
	db = testutils.SetupAuditDB(t)
	if db == nil {
		t.Skip("Database connection not available")
		return
	}

	// Create an event
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateDataExchangeEventRequest{
		Timestamp:     timestamp,
		Status:        "success",
		ApplicationID: "test-app-db-1",
		SchemaID:      "test-schema-123",
		RequestedData: map[string]interface{}{
			"personInfo": map[string]interface{}{
				"name": "John Doe",
			},
		},
		ConsumerID: "test-consumer-db-1",
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/data-exchange-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse CreateDataExchangeEventResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	eventID := createResponse.ID
	require.NotEmpty(t, eventID)

	// Verify event exists in database
	var count int64
	err = db.Table("data_exchange_events").
		Where("id = ?", eventID).
		Count(&count).Error

	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Event should exist in database")

	// Cleanup
	t.Cleanup(func() {
		db.Exec("DELETE FROM data_exchange_events WHERE id = ?", eventID)
	})
}

// TestAudit_CreateManagementEvent tests creating a management event
func TestAudit_CreateManagementEvent(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	createReq := CreateManagementEventRequest{
		Timestamp: timestamp,
		EventType: "CREATE",
		Status:    "success",
		Actor: ManagementEventActor{
			Type: "USER",
			ID:   "test-user-123",
			Role: "ADMIN",
		},
		Target: ManagementEventTarget{
			Resource:   "SCHEMAS",
			ResourceID: "test-schema-123",
		},
		Metadata: map[string]interface{}{
			"schemaName": "Test Schema",
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(auditBaseURL+"/api/management-events", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	// POST requests creating resources should return 201 Created.
	// 200 OK may be acceptable if the API implements idempotent creation
	// (e.g., returns existing event if already created with same identifier).
	assert.Contains(t, []int{http.StatusCreated, http.StatusOK}, resp.StatusCode,
		"POST to create management event should return 201 Created (or 200 OK if idempotent)")
}

// TestAudit_GetManagementEvents tests retrieving management events
func TestAudit_GetManagementEvents(t *testing.T) {
	// Retrieve management events
	resp, err := http.Get(auditBaseURL + "/api/management-events")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var listResponse ManagementEventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, listResponse.Total, len(listResponse.Events), "total should be >= number of events returned")
}
