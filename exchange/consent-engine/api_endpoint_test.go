package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"person.permanentAddress", "person.birthDate"},
				},
			},
			Purpose:          "passport_application",
			SessionID:        "session_123",
			ConsentPortalURL: "https://passport-app.gov.lk/callback",
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
		if response["redirect_url"] == "" {
			t.Error("Expected non-empty redirect_url")
		}
	})

	t.Run("CreateConsent_InvalidRequest", func(t *testing.T) {
		// Test with empty data fields
		reqBody := ConsentRequest{
			AppID:            "passport-app",
			DataFields:       []DataField{},
			Purpose:          "passport_application",
			SessionID:        "session_123",
			ConsentPortalURL: "https://passport-app.gov.lk/callback",
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
