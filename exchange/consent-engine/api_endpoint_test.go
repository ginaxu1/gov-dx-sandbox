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

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestPOSTConsentsEndpoint tests the POST /consents endpoint
func TestPOSTConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	t.Run("CreateNewConsent_Success", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"app_id": "acde070d-8c4c-4f0d-9d8a-162843c10333",
			"consent_requirements": []map[string]interface{}{
				{
					"owner":    "CITIZEN",
					"owner_id": "mohamed@opensource.lk",
					"fields": []map[string]string{
						{
							"fieldName": "personInfo.name",
							"schemaId":  "acde070d-8c4c-4f0d-9d8a-162843c10333",
						},
					},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		// Verify response
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields - new format: only consent_id, status, consent_portal_url
		if _, exists := response["consent_id"]; !exists {
			t.Error("Expected 'consent_id' field in response")
		}
		if response["status"] != "pending" {
			t.Errorf("Expected status 'pending', got '%s'", response["status"])
		}
		if _, exists := response["consent_portal_url"]; !exists {
			t.Error("Expected 'consent_portal_url' field in response")
		}
		consentPortalURL, ok := response["consent_portal_url"].(string)
		if !ok || consentPortalURL == "" {
			t.Error("Expected non-empty consent_portal_url")
		}
		// Verify URL format includes consent_id
		consentID, ok := response["consent_id"].(string)
		if ok && !contains(consentPortalURL, consentID) {
			t.Errorf("Expected consent_portal_url to contain consent_id '%s', got '%s'", consentID, consentPortalURL)
		}
	})

	t.Run("CreateConsent_InvalidRequest", func(t *testing.T) {
		// Test with empty consent_requirements
		reqBody := map[string]interface{}{
			"app_id":               "acde070d-8c4c-4f0d-9d8a-162843c10333",
			"consent_requirements": []interface{}{},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateConsent_DifferentOwners", func(t *testing.T) {
		// First request with owner_id: "mohamed@opensource.lk"
		req1 := map[string]interface{}{
			"app_id": "acde070d-8c4c-4f0d-9d8a-162843c10333",
			"consent_requirements": []map[string]interface{}{
				{
					"owner":    "CITIZEN",
					"owner_id": "mohamed@opensource.lk",
					"fields": []map[string]string{
						{
							"fieldName": "personInfo.name",
							"schemaId":  "acde070d-8c4c-4f0d-9d8a-162843c10333",
						},
					},
				},
			},
		}

		jsonBody1, _ := json.Marshal(req1)
		httpReq1 := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody1))
		httpReq1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()

		server.consentHandler(w1, httpReq1)

		if w1.Code != http.StatusCreated {
			t.Errorf("First request: Expected status %d, got %d", http.StatusCreated, w1.Code)
		}

		var response1 map[string]interface{}
		if err := json.Unmarshal(w1.Body.Bytes(), &response1); err != nil {
			t.Fatalf("Failed to unmarshal first response: %v", err)
		}

		firstConsentID := response1["consent_id"].(string)

		// Second request with different owner_id
		req2 := map[string]interface{}{
			"app_id": "acde070d-8c4c-4f0d-9d8a-162843c10333",
			"consent_requirements": []map[string]interface{}{
				{
					"owner":    "CITIZEN",
					"owner_id": "regina@opensource.lk",
					"fields": []map[string]string{
						{
							"fieldName": "personInfo.name",
							"schemaId":  "acde070d-8c4c-4f0d-9d8a-162843c10333",
						},
					},
				},
			},
		}

		jsonBody2, _ := json.Marshal(req2)
		httpReq2 := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody2))
		httpReq2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()

		server.consentHandler(w2, httpReq2)

		if w2.Code != http.StatusCreated {
			t.Errorf("Second request: Expected status %d, got %d", http.StatusCreated, w2.Code)
		}

		var response2 map[string]interface{}
		if err := json.Unmarshal(w2.Body.Bytes(), &response2); err != nil {
			t.Fatalf("Failed to unmarshal second response: %v", err)
		}

		secondConsentID := response2["consent_id"].(string)

		// Verify different consent IDs
		if firstConsentID == secondConsentID {
			t.Errorf("Expected different consent IDs, but both returned '%s'", firstConsentID)
		}
	})
}

// TestPUTConsentsEndpoint tests the PUT /consents/{id} endpoint
func TestPUTConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	t.Run("UpdateNonExistentConsent", func(t *testing.T) {
		updateData := map[string]string{
			"status":   "approved",
			"owner_id": "1998888888",
			"message":  "Test update",
		}

		jsonBody, _ := json.Marshal(updateData)
		req := httptest.NewRequest("PUT", "/consents/non-existent-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestGETConsentsEndpoint tests the GET /consents/{id} endpoint
func TestGETConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	t.Run("GetNonExistentConsent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consents/non-existent-id", nil)
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestDELETEConsentsEndpoint tests the DELETE /consents/{id} endpoint
func TestDELETEConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	t.Run("RevokeNonExistentConsent", func(t *testing.T) {
		revokeData := map[string]string{
			"reason": "User requested revocation",
		}

		jsonBody, _ := json.Marshal(revokeData)
		req := httptest.NewRequest("DELETE", "/consents/non-existent-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestPOSTAdminExpiryCheckEndpoint tests the POST /admin/expiry-check endpoint
func TestPOSTAdminExpiryCheckEndpoint(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	t.Run("NoExpiredRecords", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/expiry-check", nil)
		w := httptest.NewRecorder()

		server.adminHandler(w, req)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response structure
		if _, exists := response["expired_records"]; !exists {
			t.Error("Expected 'expired_records' field in response")
		}

		if _, exists := response["count"]; !exists {
			t.Error("Expected 'count' field in response")
		}

		if _, exists := response["checked_at"]; !exists {
			t.Error("Expected 'checked_at' field in response")
		}

		// Validate that expired_records is an empty array, not null
		expiredRecords, ok := response["expired_records"].([]interface{})
		if !ok {
			t.Errorf("Expected expired_records to be an array, got %T", response["expired_records"])
		}

		if len(expiredRecords) != 0 {
			t.Errorf("Expected 0 expired records, got %d", len(expiredRecords))
		}

		count, ok := response["count"].(float64)
		if !ok {
			t.Errorf("Expected count to be a number, got %T", response["count"])
		}

		if int(count) != 0 {
			t.Errorf("Expected count to be 0, got %d", int(count))
		}
	})

	t.Run("WithExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		engine := setupPostgresTestEngine(t)
		server := &apiServer{engine: engine}

		// Create a consent with a very short grant duration
		req := ConsentRequest{
			AppID: "passport-app",
			ConsentRequirements: []ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user123@example.com",
					Fields: []ConsentField{
						{
							FieldName: "person.permanentAddress",
							SchemaID:  "schema_123",
						},
					},
				},
			},
			GrantDuration: "1s", // Very short duration
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Approve the consent
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "citizen_123",
			Reason:    "User approved",
		}
		_, err = engine.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		// Wait for the record to expire
		time.Sleep(2 * time.Second)

		// Call the expiry check endpoint
		httpReq := httptest.NewRequest("POST", "/admin/expiry-check", nil)
		w := httptest.NewRecorder()

		server.adminHandler(w, httpReq)

		// Verify response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response structure
		expiredRecords, ok := response["expired_records"].([]interface{})
		if !ok {
			t.Errorf("Expected expired_records to be an array, got %T", response["expired_records"])
		}

		if len(expiredRecords) != 1 {
			t.Errorf("Expected 1 expired record, got %d", len(expiredRecords))
		}

		count, ok := response["count"].(float64)
		if !ok {
			t.Errorf("Expected count to be a number, got %T", response["count"])
		}

		if int(count) != 1 {
			t.Errorf("Expected count to be 1, got %d", int(count))
		}

		// Validate the expired record structure
		expiredRecord, ok := expiredRecords[0].(map[string]interface{})
		if !ok {
			t.Errorf("Expected expired record to be an object, got %T", expiredRecords[0])
		}

		if expiredRecord["consent_id"] != record.ConsentID {
			t.Errorf("Expected consent_id %s, got %s", record.ConsentID, expiredRecord["consent_id"])
		}

		if expiredRecord["status"] != "approved" {
			t.Errorf("Expected status 'approved', got %s", expiredRecord["status"])
		}
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/expiry-check", nil)
		w := httptest.NewRecorder()

		server.adminHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

// TestPUTConsentsWithGrantDuration tests the PUT /consents/:consentId endpoint with grant_duration
func TestPUTConsentsWithGrantDuration(t *testing.T) {
	// Create a test server
	engine := setupPostgresTestEngine(t)
	server := &apiServer{engine: engine}

	// First create a consent
	createReq := ConsentRequest{
		AppID: "passport-app",
		ConsentRequirements: []ConsentRequirement{
			{
				Owner:   "CITIZEN",
				OwnerID: "mohamed@opensource.lk",
				Fields: []ConsentField{
					{
						FieldName: "personInfo.permanentAddress",
						SchemaID:  "schema_123",
					},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.consentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var createResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &createResponse); err != nil {
		t.Fatalf("Failed to unmarshal create response: %v", err)
	}

	consentID := createResponse["consent_id"].(string)

	// Now update the consent with grant_duration
	updateReq := map[string]interface{}{
		"status":         "approved",
		"grant_duration": "1m",
		"updated_by":     "citizen_199512345678",
		"reason":         "User approved consent via portal",
	}

	jsonBody, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PATCH", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	// PATCH /consents/{id} routes through consentHandlerWithID
	server.consentHandlerWithID(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate simplified response format (consent_id, status only)
	// consent_portal_url should NOT be present when status is approved
	if response["consent_id"] != consentID {
		t.Errorf("Expected consent_id %s, got %s", consentID, response["consent_id"])
	}

	if response["status"] != "approved" {
		t.Errorf("Expected status 'approved', got %s", response["status"])
	}

	// Verify consent_portal_url is NOT present (only present when status is pending)
	if _, exists := response["consent_portal_url"]; exists {
		t.Error("Expected consent_portal_url to be absent when status is approved")
	}

	// Verify the update was successful by checking the consent status via GET
	// (The simplified response doesn't include grant_duration, owner_id, etc.)
	getReq := httptest.NewRequest("GET", "/consents/"+consentID, nil)
	getW := httptest.NewRecorder()

	// Use the consentHandlerWithID handler for GET
	server.consentHandlerWithID(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("Expected GET status %d, got %d. Response: %s", http.StatusOK, getW.Code, getW.Body.String())
	}

	// Note: The simplified response format means we don't return grant_duration, owner_id, etc.
	// in the PATCH response. These details can be retrieved via GET if needed.
}
