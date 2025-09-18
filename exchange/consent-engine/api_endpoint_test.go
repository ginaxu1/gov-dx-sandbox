package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPOSTConsentsEndpoint tests the POST /consents endpoint
func TestPOSTConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := NewConsentEngine("http://localhost:5173")
	server := &apiServer{engine: engine}

	t.Run("CreateNewConsent_Success", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType:  "citizen",
					OwnerID:    "199512345678",
					OwnerEmail: "199512345678@example.com",
					Fields:     []string{"person.permanentAddress", "person.birthDate"},
				},
			},
			Purpose:   "passport_application",
			SessionID: "session_123",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		// Verify response
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields
		if response["status"] != "pending" {
			t.Errorf("Expected status 'pending', got '%s'", response["status"])
		}
		if response["owner_email"] != "199512345678@example.com" {
			t.Errorf("Expected owner_email '199512345678@example.com', got '%s'", response["owner_email"])
		}
		if response["redirect_url"] == "" {
			t.Error("Expected non-empty redirect_url")
		}
	})

	t.Run("CreateConsent_InvalidRequest", func(t *testing.T) {
		// Test with empty data fields
		reqBody := ConsentRequest{
			AppID:      "passport-app",
			DataFields: []DataField{},
			Purpose:    "passport_application",
			SessionID:  "session_123",
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
}

// TestPUTConsentsEndpoint tests the PUT /consents/{id} endpoint
func TestPUTConsentsEndpoint(t *testing.T) {
	// Create a test server
	engine := NewConsentEngine("http://localhost:5173")
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
	engine := NewConsentEngine("http://localhost:5173")
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
	engine := NewConsentEngine("http://localhost:5173")
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
	engine := NewConsentEngine("http://localhost:5173")
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
		engine := NewConsentEngine("http://localhost:5173")
		server := &apiServer{engine: engine}

		// Create a consent
		req := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType:  "citizen",
					OwnerID:    "user123",
					OwnerEmail: "user123@example.com",
					Fields:     []string{"person.permanentAddress"},
				},
			},
			Purpose:   "passport_application",
			SessionID: "session_123",
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

		// Manually set the expiry time to the past
		record.ExpiresAt = time.Now().Add(-1 * time.Hour)
		engineImpl := engine.(*consentEngineImpl)
		engineImpl.consentRecords[record.ConsentID] = record

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

		if expiredRecord["status"] != "expired" {
			t.Errorf("Expected status 'expired', got %s", expiredRecord["status"])
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
	engine := NewConsentEngine("http://localhost:5173")
	server := &apiServer{engine: engine}

	// First create a consent
	createReq := ConsentRequest{
		AppID: "passport-app",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "33333",
				OwnerEmail: "33333@example.com",
				Fields:     []string{"personInfo.permanentAddress"},
			},
		},
		Purpose:   "passport_application",
		SessionID: "session_123",
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
	req = httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.consentHandler(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response fields
	if response["consent_id"] != consentID {
		t.Errorf("Expected consent_id %s, got %s", consentID, response["consent_id"])
	}

	if response["status"] != "approved" {
		t.Errorf("Expected status 'approved', got %s", response["status"])
	}

	if response["grant_duration"] != "1m" {
		t.Errorf("Expected grant_duration '1m', got %s", response["grant_duration"])
	}

	if response["owner_id"] != "33333" {
		t.Errorf("Expected owner_id '33333', got %s", response["owner_id"])
	}

	if response["owner_email"] != "33333@example.com" {
		t.Errorf("Expected owner_email '33333@example.com', got %s", response["owner_email"])
	}

	if response["app_id"] != "passport-app" {
		t.Errorf("Expected app_id 'passport-app', got %s", response["app_id"])
	}

	// Verify expires_at was recalculated
	expiresAtStr, ok := response["expires_at"].(string)
	if !ok {
		t.Error("Expected expires_at to be a string")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		t.Fatalf("Failed to parse expires_at: %v", err)
	}

	// Should be approximately 1 minute from now
	expectedExpiry := time.Now().Add(1 * time.Minute)
	timeDiff := expiresAt.Sub(expectedExpiry)
	if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
		t.Errorf("Expected expires_at to be approximately 1 minute from now, got %v", expiresAt)
	}
}
