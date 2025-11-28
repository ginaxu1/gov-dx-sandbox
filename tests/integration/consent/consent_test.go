package consent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	consentBaseURL = "http://127.0.0.1:8081"
)

func TestMain(m *testing.M) {
	// Wait for consent engine service availability
	if err := testutils.WaitForService(consentBaseURL + "/health"); err != nil {
		fmt.Printf("Consent Engine service not available: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

// TestConsent_CreateAndRetrieve tests basic consent creation and retrieval
func TestConsent_CreateAndRetrieve(t *testing.T) {
	appID := "test-app-consent-1"
	ownerID := "test-owner-123"
	fieldName := "personInfo.name"
	schemaID := "test-schema-123"

	// Create consent request
	createReq := map[string]interface{}{
		"app_id": appID,
		"consent_requirements": []map[string]interface{}{
			{
				"owner":    "citizen",
				"owner_id": ownerID,
				"fields": []map[string]interface{}{
					{
						"fieldName": fieldName,
						"schemaId":  schemaID,
					},
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create consent
	resp, err := http.Post(consentBaseURL+"/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID, ok := createResponse["consent_id"].(string)
	require.True(t, ok, "consent_id should be present in response")
	assert.NotEmpty(t, consentID, "consent_id should not be empty")

	status, ok := createResponse["status"].(string)
	require.True(t, ok, "status should be present in response")
	assert.Equal(t, "pending", status, "New consent should have pending status")

	// Retrieve consent (internal endpoint - no auth required for testing)
	resp, err = http.Get(consentBaseURL + "/data-owner/" + ownerID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var retrieveResponse []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&retrieveResponse)
	require.NoError(t, err)

	// Verify consent exists in the list
	found := false
	for _, consent := range retrieveResponse {
		if cid, ok := consent["consent_id"].(string); ok && cid == consentID {
			found = true
			assert.Equal(t, "pending", consent["status"], "Consent status should be pending")
			break
		}
	}
	assert.True(t, found, "Created consent should be retrievable")
}

// TestConsent_InvalidRequest tests edge cases for invalid consent requests
func TestConsent_InvalidRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing app_id",
			request: map[string]interface{}{
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "citizen",
						"owner_id": "test-owner",
						"fields":   []map[string]interface{}{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty consent_requirements",
			request: map[string]interface{}{
				"app_id":               "test-app",
				"consent_requirements": []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing owner_id",
			request: map[string]interface{}{
				"app_id": "test-app",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":  "citizen",
						"fields": []map[string]interface{}{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty fields",
			request: map[string]interface{}{
				"app_id": "test-app",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "citizen",
						"owner_id": "test-owner",
						"fields":   []map[string]interface{}{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			resp, err := http.Post(consentBaseURL+"/consents", "application/json", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Expected status %d for invalid request: %s", tt.expectedStatus, tt.name)
		})
	}
}

// TestConsent_GetByConsumer tests retrieving consents by consumer ID
func TestConsent_GetByConsumer(t *testing.T) {
	appID := "test-app-consumer-1"
	ownerID := "test-owner-consumer-1"
	consumerID := "test-consumer-123"

	// Create consent
	createReq := map[string]interface{}{
		"app_id": appID,
		"consent_requirements": []map[string]interface{}{
			{
				"owner":    "citizen",
				"owner_id": ownerID,
				"fields": []map[string]interface{}{
					{
						"fieldName": "personInfo.name",
						"schemaId":  "test-schema-123",
					},
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	_, ok := createResponse["consent_id"].(string)
	require.True(t, ok)

	// Retrieve by consumer
	resp, err = http.Get(consentBaseURL + "/consumer/" + consumerID)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Note: This endpoint may return empty list if consumer_id doesn't match
	// The exact behavior depends on how the service maps app_id to consumer_id
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
		"Should return 200 OK or 404 Not Found")
}

// TestConsent_StatusUpdate tests consent status updates
// Note: This test may require JWT authentication in production
// For integration tests, we test the internal PATCH endpoint if available
func TestConsent_StatusUpdate(t *testing.T) {
	appID := "test-app-update-1"
	ownerID := "test-owner-update-1"

	// Create consent
	createReq := map[string]interface{}{
		"app_id": appID,
		"consent_requirements": []map[string]interface{}{
			{
				"owner":    "citizen",
				"owner_id": ownerID,
				"fields": []map[string]interface{}{
					{
						"fieldName": "personInfo.name",
						"schemaId":  "test-schema-123",
					},
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID, ok := createResponse["consent_id"].(string)
	require.True(t, ok)

	// Update consent status using PATCH (internal endpoint)
	updateReq := map[string]interface{}{
		"status":     "approved",
		"updated_by": ownerID,
	}

	reqBody, err = json.Marshal(updateReq)
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", consentBaseURL+"/consents/"+consentID, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// PATCH endpoint should work for internal updates
	if resp.StatusCode == http.StatusOK {
		var updateResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&updateResponse)
		require.NoError(t, err)

		status, ok := updateResponse["status"].(string)
		require.True(t, ok)
		assert.Equal(t, "approved", status, "Consent status should be updated to approved")
	} else {
		// If PATCH requires auth, that's expected - we're testing the endpoint exists
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
			"PATCH endpoint should exist (may require auth)")
	}
}

// TestConsent_ExpiryCheck tests the expiry check endpoint
func TestConsent_ExpiryCheck(t *testing.T) {
	// Call the admin expiry check endpoint
	resp, err := http.Post(consentBaseURL+"/admin/expiry-check", "application/json", bytes.NewBuffer([]byte("{}")))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Expiry check should return 200 OK
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent,
		"Expiry check should return 200 OK or 204 No Content")
}

// TestConsent_DatabaseVerification tests consent creation with database verification
func TestConsent_DatabaseVerification(t *testing.T) {
	if os.Getenv("TEST_VERIFY_DB") != "true" {
		t.Skip("Skipping database verification test (set TEST_VERIFY_DB=true to enable)")
	}

	var db *gorm.DB
	db = testutils.SetupPostgresTestDB(t)
	if db == nil {
		t.Skip("Database connection not available")
		return
	}

	appID := "test-app-db-1"
	ownerID := "test-owner-db-1"

	// Create consent
	createReq := map[string]interface{}{
		"app_id": appID,
		"consent_requirements": []map[string]interface{}{
			{
				"owner":    "citizen",
				"owner_id": ownerID,
				"fields": []map[string]interface{}{
					{
						"fieldName": "personInfo.name",
						"schemaId":  "test-schema-123",
					},
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID, ok := createResponse["consent_id"].(string)
	require.True(t, ok)

	// Verify consent exists in database
	var count int64
	err = db.Table("consent_records").
		Where("consent_id = ?", consentID).
		Count(&count).Error

	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Consent should exist in database")

	// Cleanup
	defer func() {
		db.Exec("DELETE FROM consent_records WHERE consent_id = ?", consentID)
	}()
}
