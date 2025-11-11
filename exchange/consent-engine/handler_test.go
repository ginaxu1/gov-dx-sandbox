package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// createTestConsent is a helper function to create a consent record for testing
// It reduces code duplication across test functions and ensures consistent test data setup
func createTestConsent(t *testing.T, engine ConsentEngine, appID, ownerID string) *ConsentRecord {
	if appID == "" {
		appID = "test-app"
	}
	if ownerID == "" {
		ownerID = "user@example.com"
	}

	createReq := ConsentRequest{
		AppID: appID,
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: ownerID,
				Fields: []ConsentField{
					{
						FieldName: "personInfo.name",
						SchemaID:  "schema-123",
					},
				},
			},
		},
	}
	record, err := engine.ProcessConsentRequest(createReq)
	if err != nil {
		t.Fatalf("Failed to create test consent: %v", err)
	}
	return record
}

// makeJSONRequest is a simple helper to marshal request body and create an HTTP request
// It handles json.Marshal errors to prevent tests from proceeding with invalid data
func makeJSONRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// TestPOSTConsents tests POST /consents endpoint
func TestPOSTConsents(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "CreateNewConsent_Success",
			requestBody: map[string]interface{}{
				"app_id": "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "CITIZEN",
						"owner_id": "mohamed@opensource.lk",
						"fields": []map[string]string{
							{
								"fieldName": "personInfo.name",
								"schemaId":  "acde070d-8c4c-4f0d-9d8a-000000a111111",
							},
						},
					},
				},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if _, exists := response["consent_id"]; !exists {
					t.Error("Expected 'consent_id' field in response")
				}
				if response["status"] != "pending" {
					t.Errorf("Expected status 'pending', got '%s'", response["status"])
				}
				if _, exists := response["consent_portal_url"]; !exists {
					t.Error("Expected 'consent_portal_url' field in response")
				}
			},
		},
		{
			name: "CreateConsent_InvalidRequest_EmptyAppID",
			requestBody: map[string]interface{}{
				"app_id":               "",
				"consent_requirements": []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "CreateConsent_InvalidRequest_EmptyConsentRequirements",
			requestBody: map[string]interface{}{
				"app_id":               "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "CreateConsent_InvalidRequest_EmptyOwnerID",
			requestBody: map[string]interface{}{
				"app_id": "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "CITIZEN",
						"owner_id": "",
						"fields":   []map[string]string{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "CreateConsent_InvalidRequest_EmptyFields",
			requestBody: map[string]interface{}{
				"app_id": "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "CITIZEN",
						"owner_id": "mohamed@opensource.lk",
						"fields":   []interface{}{},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "CreateConsent_InvalidRequest_MissingFieldName",
			requestBody: map[string]interface{}{
				"app_id": "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "CITIZEN",
						"owner_id": "mohamed@opensource.lk",
						"fields": []map[string]string{
							{
								"schemaId": "acde070d-8c4c-4f0d-9d8a-000000a111111",
							},
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "CreateConsent_InvalidRequest_MissingSchemaID",
			requestBody: map[string]interface{}{
				"app_id": "acde070d-8c4c-4f0d-9d8a-000000a111111",
				"consent_requirements": []map[string]interface{}{
					{
						"owner":    "CITIZEN",
						"owner_id": "mohamed@opensource.lk",
						"fields": []map[string]string{
							{
								"fieldName": "personInfo.name",
							},
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := makeJSONRequest(t, http.MethodPost, "/consents", tt.requestBody)
			w := httptest.NewRecorder()

			server.consentHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestGETConsentsByID tests GET /consents/{id} endpoint
func TestGETConsentsByID(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	// Create a consent first
	record := createTestConsent(t, engine, "test-app", "user@example.com")

	tests := []struct {
		name           string
		consentID      string
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "GetConsent_Success",
			consentID:      record.ConsentID,
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if response["status"] == nil {
					t.Error("Expected 'status' field in response")
				}
				if response["fields"] == nil {
					t.Error("Expected 'fields' field in response")
				}
			},
		},
		{
			name:           "GetConsent_NotFound",
			consentID:      "non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/consents/"+tt.consentID, nil)
			w := httptest.NewRecorder()

			server.consentHandlerWithID(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestPATCHConsentsByID tests PATCH /consents/{id} endpoint
func TestPATCHConsentsByID(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		setupConsent   bool // Whether to create a new consent for this test
		consentID      string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{}, consentID string)
	}{
		{
			name:         "PatchConsent_UpdateStatus",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"status": "approved",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}, consentID string) {
				if response["status"] != "approved" {
					t.Errorf("Expected status 'approved', got '%s'", response["status"])
				}
				if response["consent_id"] != consentID {
					t.Errorf("Expected consent_id '%s', got '%s'", consentID, response["consent_id"])
				}
			},
		},
		{
			name:         "PatchConsent_UpdateGrantDuration",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"status":         "approved",
				"grant_duration": "1m",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}, consentID string) {
				if response["status"] != "approved" {
					t.Errorf("Expected status 'approved', got '%s'", response["status"])
				}
			},
		},
		{
			name:         "PatchConsent_NotFound",
			setupConsent: false,
			consentID:    "non-existent-id",
			requestBody: map[string]interface{}{
				"status": "approved",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "PatchConsent_InvalidJSON",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			expectedStatus: http.StatusOK, // PATCH accepts partial updates
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consentID := tt.consentID
			if tt.setupConsent {
				record := createTestConsent(t, engine, "test-app", "user@example.com")
				consentID = record.ConsentID
			}

			req := makeJSONRequest(t, http.MethodPatch, "/consents/"+consentID, tt.requestBody)
			w := httptest.NewRecorder()

			server.consentHandlerWithID(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response, consentID)
			}
		})
	}
}

// TestDELETEConsentsByID tests DELETE /consents/{id} endpoint
func TestDELETEConsentsByID(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		setupConsent   bool // Whether to create a new consent for this test
		consentID      string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:         "DeleteConsent_Success",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"reason": "User requested revocation",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if response["status"] != "revoked" {
					t.Errorf("Expected status 'revoked', got '%s'", response["status"])
				}
			},
		},
		{
			name:           "DeleteConsent_NotFound",
			setupConsent:   false,
			consentID:      "non-existent-id",
			requestBody:    map[string]interface{}{"reason": "test"},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consentID := tt.consentID
			if tt.setupConsent {
				record := createTestConsent(t, engine, "test-app", "user@example.com")
				consentID = record.ConsentID
			}

			req := makeJSONRequest(t, http.MethodDelete, "/consents/"+consentID, tt.requestBody)
			w := httptest.NewRecorder()

			server.consentHandlerWithID(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestGETConsentsByDataOwner tests GET /data-owner/{ownerId} endpoint
func TestGETConsentsByDataOwner(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	// Create consents for a specific owner
	_ = createTestConsent(t, engine, "test-app", "owner123@example.com")

	tests := []struct {
		name           string
		ownerID        string
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "GetConsentsByDataOwner_Success",
			ownerID:        "owner123@example.com",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if _, exists := response["owner_id"]; !exists {
					t.Error("Expected 'owner_id' field in response")
				}
				if _, exists := response["consents"]; !exists {
					t.Error("Expected 'consents' field in response")
				}
				if _, exists := response["count"]; !exists {
					t.Error("Expected 'count' field in response")
				}
			},
		},
		{
			name:           "GetConsentsByDataOwner_NoConsents",
			ownerID:        "nonexistent@example.com",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				count, ok := response["count"].(float64)
				if !ok {
					t.Error("Expected 'count' to be a number")
				}
				if int(count) != 0 {
					t.Errorf("Expected count 0, got %d", int(count))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/data-owner/"+tt.ownerID, nil)
			w := httptest.NewRecorder()

			server.dataOwnerHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestGETConsentsByConsumer tests GET /consumer/{consumerId} endpoint
func TestGETConsentsByConsumer(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	// Create consents for a specific consumer app
	_ = createTestConsent(t, engine, "consumer-app-123", "user@example.com")

	tests := []struct {
		name           string
		consumerID     string
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "GetConsentsByConsumer_Success",
			consumerID:     "consumer-app-123",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if _, exists := response["consumer"]; !exists {
					t.Error("Expected 'consumer' field in response")
				}
				if _, exists := response["consents"]; !exists {
					t.Error("Expected 'consents' field in response")
				}
				if _, exists := response["count"]; !exists {
					t.Error("Expected 'count' field in response")
				}
			},
		},
		{
			name:           "GetConsentsByConsumer_NoConsents",
			consumerID:     "nonexistent-app",
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				count, ok := response["count"].(float64)
				if !ok {
					t.Error("Expected 'count' to be a number")
				}
				if int(count) != 0 {
					t.Errorf("Expected count 0, got %d", int(count))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/consumer/"+tt.consumerID, nil)
			w := httptest.NewRecorder()

			server.consumerHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestGETDataInfo tests GET /data-info/{consentId} endpoint
func TestGETDataInfo(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	// Create a consent first
	record := createTestConsent(t, engine, "test-app", "owner@example.com")

	tests := []struct {
		name           string
		consentID      string
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "GetDataInfo_Success",
			consentID:      record.ConsentID,
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if _, exists := response["owner_id"]; !exists {
					t.Error("Expected 'owner_id' field in response")
				}
				if _, exists := response["owner_email"]; !exists {
					t.Error("Expected 'owner_email' field in response")
				}
				if response["owner_id"] != "owner@example.com" {
					t.Errorf("Expected owner_id 'owner@example.com', got '%s'", response["owner_id"])
				}
			},
		},
		{
			name:           "GetDataInfo_NotFound",
			consentID:      "non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/data-info/"+tt.consentID, nil)
			w := httptest.NewRecorder()

			server.dataInfoHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestPOSTAdminExpiryCheck tests POST /admin/expiry-check endpoint
func TestPOSTAdminExpiryCheck(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		method         string
		setupFunc      func(t *testing.T, engine ConsentEngine) string
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:   "ExpiryCheck_NoExpiredRecords",
			method: http.MethodPost,
			setupFunc: func(t *testing.T, engine ConsentEngine) string {
				// Create a consent that won't expire soon
				createReq := ConsentRequest{
					AppID: "test-app",
					ConsentRequirements: []ConsentRequirement{
						{
							Owner:   "CITIZEN",
							OwnerID: "user@example.com",
							Fields: []ConsentField{
								{
									FieldName: "personInfo.name",
									SchemaID:  "schema-123",
								},
							},
						},
					},
				}
				record, err := engine.ProcessConsentRequest(createReq)
				if err != nil {
					t.Fatalf("Failed to create consent: %v", err)
				}
				return record.ConsentID
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if _, exists := response["expired_records"]; !exists {
					t.Error("Expected 'expired_records' field in response")
				}
				if _, exists := response["count"]; !exists {
					t.Error("Expected 'count' field in response")
				}
				expiredRecords, ok := response["expired_records"].([]interface{})
				if !ok {
					t.Errorf("Expected expired_records to be an array, got %T", response["expired_records"])
				}
				if len(expiredRecords) != 0 {
					t.Errorf("Expected 0 expired records, got %d", len(expiredRecords))
				}
			},
		},
		{
			name:   "ExpiryCheck_WithExpiredRecords",
			method: http.MethodPost,
			setupFunc: func(t *testing.T, engine ConsentEngine) string {
				// Create a consent with very short grant duration
				createReq := ConsentRequest{
					AppID: "test-app",
					ConsentRequirements: []ConsentRequirement{
						{
							Owner:   "CITIZEN",
							OwnerID: "user@example.com",
							Fields: []ConsentField{
								{
									FieldName: "personInfo.name",
									SchemaID:  "schema-123",
								},
							},
						},
					},
					GrantDuration: "1s",
				}
				record, err := engine.ProcessConsentRequest(createReq)
				if err != nil {
					t.Fatalf("Failed to create consent: %v", err)
				}

				// Approve the consent
				updateReq := UpdateConsentRequest{
					Status:    StatusApproved,
					UpdatedBy: "user@example.com",
					Reason:    "User approved",
				}
				_, err = engine.UpdateConsent(record.ConsentID, updateReq)
				if err != nil {
					t.Fatalf("Failed to approve consent: %v", err)
				}

				// Wait for expiry
				time.Sleep(2 * time.Second)
				return record.ConsentID
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				expiredRecords, ok := response["expired_records"].([]interface{})
				if !ok {
					t.Errorf("Expected expired_records to be an array, got %T", response["expired_records"])
				}
				if len(expiredRecords) < 1 {
					t.Errorf("Expected at least 1 expired record, got %d", len(expiredRecords))
				}
			},
		},
		{
			name:           "ExpiryCheck_InvalidMethod",
			method:         http.MethodGet,
			setupFunc:      func(t *testing.T, engine ConsentEngine) string { return "" },
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t, engine)
			}

			req := httptest.NewRequest(tt.method, "/admin/expiry-check", nil)
			w := httptest.NewRecorder()

			server.adminHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestPUTConsentsByID tests PUT /consents/{id} endpoint
func TestPUTConsentsByID(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		setupConsent   bool // Whether to create a new consent for this test
		consentID      string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:         "PUTConsent_UpdateStatus",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"status": "approved",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if response["status"] != "approved" {
					t.Errorf("Expected status 'approved', got '%s'", response["status"])
				}
			},
		},
		{
			name:         "PUTConsent_UpdateWithGrantDuration",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"status":         "approved",
				"grant_duration": "1m",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				if response["status"] != "approved" {
					t.Errorf("Expected status 'approved', got '%s'", response["status"])
				}
			},
		},
		{
			name:         "PUTConsent_InvalidStatus",
			setupConsent: true,
			requestBody: map[string]interface{}{
				"status": "invalid-status",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUTConsent_NotFound",
			setupConsent:   false,
			consentID:      "non-existent-id",
			requestBody:    map[string]interface{}{"status": "approved"},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consentID := tt.consentID
			if tt.setupConsent {
				record := createTestConsent(t, engine, "test-app", "user@example.com")
				consentID = record.ConsentID
			}

			req := makeJSONRequest(t, http.MethodPut, "/consents/"+consentID, tt.requestBody)
			w := httptest.NewRecorder()

			server.consentHandlerWithID(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
				return
			}

			if tt.validateFunc != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.validateFunc(t, response)
			}
		})
	}
}

// TestConsentHandlerRouting tests routing logic for consent handlers
func TestConsentHandlerRouting(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST /consents",
			method:         http.MethodPost,
			path:           "/consents",
			expectedStatus: http.StatusBadRequest, // Invalid body, but endpoint exists
		},
		{
			name:           "GET /consents - Method not allowed",
			method:         http.MethodGet,
			path:           "/consents",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET /consents/{id}",
			method:         http.MethodGet,
			path:           "/consents/test-id",
			expectedStatus: http.StatusNotFound, // ID doesn't exist, but endpoint exists
		},
		{
			name:           "PUT /consents/{id}",
			method:         http.MethodPut,
			path:           "/consents/test-id",
			expectedStatus: http.StatusNotFound, // ID doesn't exist, but endpoint exists
		},
		{
			name:           "PATCH /consents/{id}",
			method:         http.MethodPatch,
			path:           "/consents/test-id",
			expectedStatus: http.StatusNotFound, // ID doesn't exist, but endpoint exists
		},
		{
			name:           "DELETE /consents/{id}",
			method:         http.MethodDelete,
			path:           "/consents/test-id",
			expectedStatus: http.StatusNotFound, // ID doesn't exist, but endpoint exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost || tt.method == http.MethodPut || tt.method == http.MethodPatch || tt.method == http.MethodDelete {
				req = makeJSONRequest(t, tt.method, tt.path, map[string]interface{}{})
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()

			if strings.HasSuffix(tt.path, "/consents") {
				server.consentHandler(w, req)
			} else {
				server.consentHandlerWithID(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
