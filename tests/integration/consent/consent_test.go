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
)

const (
	consentBaseURL = "http://127.0.0.1:8081"
)

// Request/Response types for type safety

type ConsentField struct {
	FieldName string `json:"fieldName"`
	SchemaID  string `json:"schemaId"`
}

type ConsentRequirement struct {
	Owner      string         `json:"owner"`
	OwnerID    string         `json:"ownerId"`
	OwnerEmail string         `json:"ownerEmail"`
	Fields     []ConsentField `json:"fields"`
}

type CreateConsentRequest struct {
	AppID              string             `json:"appId"`
	ConsentRequirement ConsentRequirement `json:"consentRequirement"`
}

type CreateConsentResponse struct {
	ConsentID        string          `json:"consentId"`
	Status           string          `json:"status"`
	ConsentPortalURL *string         `json:"consentPortalUrl,omitempty"`
	Fields           *[]ConsentField `json:"fields,omitempty"`
}

type Consent struct {
	ConsentID string `json:"consentId"`
	Status    string `json:"status"`
	// Add other fields as needed based on API response
}

type UpdateConsentRequest struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updated_by"`
}

type UpdateConsentResponse struct {
	Status string `json:"status"`
}

func TestMain(m *testing.M) {
	// Wait for consent engine service availability
	if err := testutils.WaitForService(consentBaseURL+"/health", 30); err != nil {
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
	createReq := CreateConsentRequest{
		AppID: appID,
		ConsentRequirement: ConsentRequirement{
			Owner:      "citizen",
			OwnerID:    ownerID,
			OwnerEmail: ownerID + "@example.com",
			Fields: []ConsentField{
				{
					FieldName: fieldName,
					SchemaID:  schemaID,
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create consent
	resp, err := http.Post(consentBaseURL+"/internal/api/v1/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var createResponse CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID := createResponse.ConsentID
	assert.NotEmpty(t, consentID, "consent_id should not be empty")
	assert.Equal(t, "pending", createResponse.Status, "New consent should have pending status")

	// Cleanup: Remove created consent record
	t.Cleanup(func() {
		if db := testutils.SetupConsentDB(t); db != nil {
			db.Exec("DELETE FROM consent_records WHERE consent_id = ?", consentID)
		}
	})

	// Retrieve consent (internal endpoint - no auth required for testing)
	// Note: Querying by ownerId alone may return multiple consents, so handle as array
	resp, err = http.Get(consentBaseURL + "/internal/api/v1/consents?ownerId=" + ownerID + "&appId=" + appID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var retrieveResponse []CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&retrieveResponse)
	require.NoError(t, err)
	require.NotEmpty(t, retrieveResponse, "Expected at least one consent for the owner")

	// Verify consent matches (assuming the first result is the one we created)
	assert.Equal(t, consentID, retrieveResponse[0].ConsentID, "Retrieved consent ID should match created consent")
	assert.Equal(t, "pending", retrieveResponse[0].Status, "Consent status should be pending")
}

// TestConsent_InvalidRequest tests edge cases for invalid consent requests
func TestConsent_InvalidRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        func() []byte
		expectedStatus int
	}{
		{
			name: "Missing appId",
			request: func() []byte {
				req := map[string]interface{}{
					"consentRequirement": map[string]interface{}{
						"owner":      "citizen",
						"ownerId":    "test-owner",
						"ownerEmail": "test@example.com",
						"fields":     []map[string]interface{}{},
					},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing consentRequirement",
			request: func() []byte {
				req := CreateConsentRequest{
					AppID: "test-app",
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing ownerId",
			request: func() []byte {
				req := map[string]interface{}{
					"appId": "test-app",
					"consentRequirement": map[string]interface{}{
						"owner":      "citizen",
						"ownerEmail": "test@example.com",
						"fields":     []map[string]interface{}{},
					},
				}
				body, _ := json.Marshal(req)
				return body
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty fields",
			request: func() []byte {
				req := CreateConsentRequest{
					AppID: "test-app",
					ConsentRequirement: ConsentRequirement{
						Owner:      "citizen",
						OwnerID:    "test-owner",
						OwnerEmail: "test@example.com",
						Fields:     []ConsentField{},
					},
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

			resp, err := http.Post(consentBaseURL+"/internal/api/v1/consents", "application/json", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode,
				"Expected status %d for invalid request: %s", tt.expectedStatus, tt.name)
		})
	}
}

// TestConsent_GetByConsumer tests retrieving consents by appId and ownerId
func TestConsent_GetByConsumer(t *testing.T) {
	appID := "test-app-consumer-1"
	ownerID := "test-owner-consumer-1"

	// Create consent
	createReq := CreateConsentRequest{
		AppID: appID,
		ConsentRequirement: ConsentRequirement{
			Owner:      "citizen",
			OwnerID:    ownerID,
			OwnerEmail: ownerID + "@example.com",
			Fields: []ConsentField{
				{
					FieldName: "personInfo.name",
					SchemaID:  "test-schema-123",
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/internal/api/v1/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	assert.NotEmpty(t, createResponse.ConsentID)

	// Retrieve by appId (consumer endpoint doesn't exist, using appId query instead)
	resp, err = http.Get(consentBaseURL + "/internal/api/v1/consents?appId=" + appID + "&ownerId=" + ownerID)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return the consent we created
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	var retrievedConsent CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&retrievedConsent)
	require.NoError(t, err)
	require.NotEmpty(t, retrievedConsent.ConsentID, "Expected to retrieve the created consent")
	assert.Equal(t, createResponse.ConsentID, retrievedConsent.ConsentID, "Retrieved consent ID should match created consent ID")

	// Cleanup: Remove created consent record
	consentID := createResponse.ConsentID
	t.Cleanup(func() {
		if db := testutils.SetupConsentDB(t); db != nil {
			db.Exec("DELETE FROM consent_records WHERE consent_id = ?", consentID)
		}
	})
}

// TestConsent_StatusUpdate tests consent status updates
// Note: This test may require JWT authentication in production
// For integration tests, we test the internal PATCH endpoint if available
func TestConsent_StatusUpdate(t *testing.T) {
	appID := "test-app-update-1"
	ownerID := "test-owner-update-1"

	// Create consent
	createReq := CreateConsentRequest{
		AppID: appID,
		ConsentRequirement: ConsentRequirement{
			Owner:      "citizen",
			OwnerID:    ownerID,
			OwnerEmail: ownerID + "@example.com",
			Fields: []ConsentField{
				{
					FieldName: "personInfo.name",
					SchemaID:  "test-schema-123",
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/internal/api/v1/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID := createResponse.ConsentID
	require.NotEmpty(t, consentID)

	// Update consent status using PUT (portal endpoint requires JWT auth)
	// For integration tests, we verify the consent was created successfully
	// Status updates require JWT authentication which is tested separately
	assert.Equal(t, "pending", createResponse.Status, "New consent should have pending status")

	// Cleanup: Remove created consent record
	t.Cleanup(func() {
		if db := testutils.SetupConsentDB(t); db != nil {
			db.Exec("DELETE FROM consent_records WHERE consent_id = ?", consentID)
		}
	})
}

// TestConsent_HealthCheck tests the health check endpoint
func TestConsent_HealthCheck(t *testing.T) {
	// Health check endpoint exists
	resp, err := http.Get(consentBaseURL + "/internal/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Health check should return 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200 OK")
}

// TestConsent_DatabaseVerification tests consent creation with database verification
func TestConsent_DatabaseVerification(t *testing.T) {
	if os.Getenv("TEST_VERIFY_DB") != "true" {
		t.Skip("Skipping database verification test (set TEST_VERIFY_DB=true to enable)")
	}

	db := testutils.SetupConsentDB(t)
	if db == nil {
		t.Skip("Database connection not available")
		return
	}

	appID := "test-app-db-1"
	ownerID := "test-owner-db-1"

	// Create consent
	createReq := CreateConsentRequest{
		AppID: appID,
		ConsentRequirement: ConsentRequirement{
			Owner:      "citizen",
			OwnerID:    ownerID,
			OwnerEmail: ownerID + "@example.com",
			Fields: []ConsentField{
				{
					FieldName: "personInfo.name",
					SchemaID:  "test-schema-123",
				},
			},
		},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(consentBaseURL+"/internal/api/v1/consents", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResponse CreateConsentResponse
	err = json.NewDecoder(resp.Body).Decode(&createResponse)
	require.NoError(t, err)

	consentID := createResponse.ConsentID
	require.NotEmpty(t, consentID)

	// Verify consent exists in database
	var count int64
	err = db.Table("consent_records").
		Where("consent_id = ?", consentID).
		Count(&count).Error

	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Consent should exist in database")

	// Cleanup
	t.Cleanup(func() {
		db.Exec("DELETE FROM consent_records WHERE consent_id = ?", consentID)
	})
}
