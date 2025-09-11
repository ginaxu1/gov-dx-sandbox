package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestConsentWorkflowIntegration tests the complete consent management workflow via HTTP API
func TestConsentWorkflowIntegration(t *testing.T) {
	// Initialize the consent engine
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Test 1: Process Consent Request - Basic Flow
	t.Run("ProcessConsentRequest_BasicFlow", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:       "passport_application",
			SessionID:     "session_123",
			RedirectURL:   "https://passport-app.gov.lk/callback",
			ExpiresAt:     time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 days from now
			GrantDuration: "30d",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		// Check response status
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		// Parse response
		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields
		if response.ID == "" {
			t.Error("Expected non-empty ID")
		}
		if response.Status != StatusPending {
			t.Errorf("Expected status %s, got %s", StatusPending, response.Status)
		}
		if response.Type != ConsentTypeRealTime {
			t.Errorf("Expected type %s, got %s", ConsentTypeRealTime, response.Type)
		}
		if response.DataConsumer != "passport-app" {
			t.Errorf("Expected DataConsumer 'passport-app', got '%s'", response.DataConsumer)
		}
		if response.DataOwner != "199512345678" {
			t.Errorf("Expected DataOwner '199512345678', got '%s'", response.DataOwner)
		}
		if len(response.Fields) != 1 || response.Fields[0] != "personInfo.address" {
			t.Errorf("Expected fields ['personInfo.address'], got %v", response.Fields)
		}
		if response.SessionID != "session_123" {
			t.Errorf("Expected SessionID 'session_123', got '%s'", response.SessionID)
		}
		if response.RedirectURL != "https://passport-app.gov.lk/callback" {
			t.Errorf("Expected RedirectURL 'https://passport-app.gov.lk/callback', got '%s'", response.RedirectURL)
		}
		if response.ConsentPortalURL == "" {
			t.Error("Expected non-empty ConsentPortalURL")
		}
		if response.Metadata == nil {
			t.Error("Expected non-nil Metadata")
		} else {
			if response.Metadata["purpose"] != "passport_application" {
				t.Errorf("Expected purpose 'passport_application', got '%v'", response.Metadata["purpose"])
			}
			if response.Metadata["request_id"] == "" {
				t.Error("Expected non-empty request_id in metadata")
			}
		}
	})

	// Test 2: Process Consent Request - Multiple Data Fields
	t.Run("ProcessConsentRequest_MultipleFields", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.name", "personInfo.address", "personInfo.profession"},
				},
			},
			Purpose:       "passport_application",
			SessionID:     "session_456",
			RedirectURL:   "https://passport-app.gov.lk/callback",
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days from now
			GrantDuration: "7d",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		expectedFields := []string{"personInfo.name", "personInfo.address", "personInfo.profession"}
		if len(response.Fields) != len(expectedFields) {
			t.Errorf("Expected %d fields, got %d", len(expectedFields), len(response.Fields))
		}
		for _, expectedField := range expectedFields {
			found := false
			for _, actualField := range response.Fields {
				if actualField == expectedField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected field '%s' not found in response", expectedField)
			}
		}
	})

	// Test 3: Process Consent Request - Multiple Data Owners
	t.Run("ProcessConsentRequest_MultipleDataOwners", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.name"},
				},
				{
					OwnerType: "government",
					OwnerID:   "gov_12345",
					Fields:    []string{"personInfo.birthDate"},
				},
			},
			Purpose:       "passport_application",
			SessionID:     "session_789",
			RedirectURL:   "https://passport-app.gov.lk/callback",
			ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(), // 1 day from now
			GrantDuration: "1d",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Should use first data owner as primary
		if response.DataOwner != "199512345678" {
			t.Errorf("Expected DataOwner '199512345678', got '%s'", response.DataOwner)
		}

		expectedFields := []string{"personInfo.name", "personInfo.birthDate"}
		if len(response.Fields) != len(expectedFields) {
			t.Errorf("Expected %d fields, got %d", len(expectedFields), len(response.Fields))
		}
	})

	// Test 4: Process Consent Request - Invalid Input
	t.Run("ProcessConsentRequest_InvalidInput", func(t *testing.T) {
		// Test with empty data fields
		reqBody := ConsentRequest{
			AppID:      "passport-app",
			DataFields: []DataField{}, // Empty data fields
			Purpose:    "passport_application",
			SessionID:  "session_invalid",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		// Should return error status
		if w.Code == http.StatusCreated {
			t.Error("Expected error status, got success")
		}
	})

	// Test 5: Process Consent Request - Missing Required Fields
	t.Run("ProcessConsentRequest_MissingRequiredFields", func(t *testing.T) {
		// Test with missing app_id
		reqBody := map[string]interface{}{
			"data_fields": []map[string]interface{}{
				{
					"owner_type": "citizen",
					"owner_id":   "199512345678",
					"fields":     []string{"personInfo.address"},
				},
			},
			"purpose": "passport_application",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		// Should return error status
		if w.Code == http.StatusCreated {
			t.Error("Expected error status, got success")
		}
	})

	// Test 6: Get Consent Status
	t.Run("GetConsentStatus", func(t *testing.T) {
		// First create a consent record
		createReq := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_get_test",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(createReq)
		createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		server.processConsentRequest(createW, createReqHTTP)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create consent record: %d", createW.Code)
		}

		var createResponse ConsentResponse
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		// Now get the consent status
		getReq := httptest.NewRequest("GET", fmt.Sprintf("/consent/%s", createResponse.ID), nil)
		getW := httptest.NewRecorder()

		server.getConsentStatus(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, getW.Code)
		}

		var getResponse ConsentRecord
		if err := json.Unmarshal(getW.Body.Bytes(), &getResponse); err != nil {
			t.Fatalf("Failed to unmarshal get response: %v", err)
		}

		if getResponse.ID != createResponse.ID {
			t.Errorf("Expected ID %s, got %s", createResponse.ID, getResponse.ID)
		}
		if getResponse.Status != StatusPending {
			t.Errorf("Expected status %s, got %s", StatusPending, getResponse.Status)
		}
	})

	// Test 7: Update Consent Status
	t.Run("UpdateConsentStatus", func(t *testing.T) {
		// First create a consent record
		createReq := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_update_test",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(createReq)
		createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		server.processConsentRequest(createW, createReqHTTP)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create consent record: %d", createW.Code)
		}

		var createResponse ConsentResponse
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		// Now update the consent status to approved
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "data_owner",
			Reason:    "User approved consent",
		}

		updateJsonBody, _ := json.Marshal(updateReq)
		updateReqHTTP := httptest.NewRequest("PUT", fmt.Sprintf("/consent/%s", createResponse.ID), bytes.NewBuffer(updateJsonBody))
		updateReqHTTP.Header.Set("Content-Type", "application/json")
		updateW := httptest.NewRecorder()

		server.updateConsent(updateW, updateReqHTTP)

		if updateW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, updateW.Code)
		}

		var updateResponse ConsentRecord
		if err := json.Unmarshal(updateW.Body.Bytes(), &updateResponse); err != nil {
			t.Fatalf("Failed to unmarshal update response: %v", err)
		}

		if updateResponse.Status != StatusApproved {
			t.Errorf("Expected status %s, got %s", StatusApproved, updateResponse.Status)
		}
	})

	// Test 8: Revoke Consent
	t.Run("RevokeConsent", func(t *testing.T) {
		// First create a consent record
		createReq := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_revoke_test",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(createReq)
		createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		server.processConsentRequest(createW, createReqHTTP)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create consent record: %d", createW.Code)
		}

		var createResponse ConsentResponse
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		// First approve the consent
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "data_owner",
			Reason:    "User approved consent",
		}

		updateJsonBody, _ := json.Marshal(updateReq)
		updateReqHTTP := httptest.NewRequest("PUT", fmt.Sprintf("/consent/%s", createResponse.ID), bytes.NewBuffer(updateJsonBody))
		updateReqHTTP.Header.Set("Content-Type", "application/json")
		updateW := httptest.NewRecorder()

		server.updateConsent(updateW, updateReqHTTP)

		if updateW.Code != http.StatusOK {
			t.Fatalf("Failed to approve consent record: %d", updateW.Code)
		}

		// Now revoke the consent
		revokeReq := map[string]string{
			"reason": "User requested revocation",
		}

		revokeJsonBody, _ := json.Marshal(revokeReq)
		revokeReqHTTP := httptest.NewRequest("DELETE", fmt.Sprintf("/consent/%s", createResponse.ID), bytes.NewBuffer(revokeJsonBody))
		revokeReqHTTP.Header.Set("Content-Type", "application/json")
		revokeW := httptest.NewRecorder()

		server.revokeConsent(revokeW, revokeReqHTTP)

		if revokeW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, revokeW.Code)
		}

		var revokeResponse ConsentRecord
		if err := json.Unmarshal(revokeW.Body.Bytes(), &revokeResponse); err != nil {
			t.Fatalf("Failed to unmarshal revoke response: %v", err)
		}

		if revokeResponse.Status != StatusRevoked {
			t.Errorf("Expected status %s, got %s", StatusRevoked, revokeResponse.Status)
		}
	})

	// Test 9: Get Consents by Data Owner
	t.Run("GetConsentsByDataOwner", func(t *testing.T) {
		// Create multiple consent records for the same data owner
		dataOwner := "199512345678"

		for i := 0; i < 3; i++ {
			createReq := ConsentRequest{
				AppID: "passport-app",
				DataFields: []DataField{
					{
						OwnerType: "citizen",
						OwnerID:   dataOwner,
						Fields:    []string{fmt.Sprintf("personInfo.field%d", i)},
					},
				},
				Purpose:     "passport_application",
				SessionID:   fmt.Sprintf("session_owner_test_%d", i),
				RedirectURL: "https://passport-app.gov.lk/callback",
			}

			jsonBody, _ := json.Marshal(createReq)
			createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
			createReqHTTP.Header.Set("Content-Type", "application/json")
			createW := httptest.NewRecorder()

			server.processConsentRequest(createW, createReqHTTP)

			if createW.Code != http.StatusCreated {
				t.Fatalf("Failed to create consent record %d: %d", i, createW.Code)
			}
		}

		// Now get all consents for the data owner
		getReq := httptest.NewRequest("GET", fmt.Sprintf("/data-owner/%s", dataOwner), nil)
		getW := httptest.NewRecorder()

		server.getConsentsByDataOwner(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, getW.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(getW.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consents, ok := response["consents"].([]interface{})
		if !ok {
			t.Fatal("Expected 'consents' field to be an array")
		}

		if len(consents) < 3 {
			t.Errorf("Expected at least 3 consents, got %d", len(consents))
		}
	})

	// Test 10: Get Consents by Consumer
	t.Run("GetConsentsByConsumer", func(t *testing.T) {
		// Create multiple consent records for the same consumer
		consumer := "passport-app"

		for i := 0; i < 2; i++ {
			createReq := ConsentRequest{
				AppID: consumer,
				DataFields: []DataField{
					{
						OwnerType: "citizen",
						OwnerID:   fmt.Sprintf("19951234567%d", i),
						Fields:    []string{fmt.Sprintf("personInfo.field%d", i)},
					},
				},
				Purpose:     "passport_application",
				SessionID:   fmt.Sprintf("session_consumer_test_%d", i),
				RedirectURL: "https://passport-app.gov.lk/callback",
			}

			jsonBody, _ := json.Marshal(createReq)
			createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
			createReqHTTP.Header.Set("Content-Type", "application/json")
			createW := httptest.NewRecorder()

			server.processConsentRequest(createW, createReqHTTP)

			if createW.Code != http.StatusCreated {
				t.Fatalf("Failed to create consent record %d: %d", i, createW.Code)
			}
		}

		// Now get all consents for the consumer
		getReq := httptest.NewRequest("GET", fmt.Sprintf("/consumer/%s", consumer), nil)
		getW := httptest.NewRecorder()

		server.getConsentsByConsumer(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, getW.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(getW.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consents, ok := response["consents"].([]interface{})
		if !ok {
			t.Fatal("Expected 'consents' field to be an array")
		}

		if len(consents) < 2 {
			t.Errorf("Expected at least 2 consents, got %d", len(consents))
		}
	})

	// Test 11: Send Consent OTP
	t.Run("SendConsentOTP", func(t *testing.T) {
		// First create a consent record
		createReq := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_otp_test",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(createReq)
		createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		server.processConsentRequest(createW, createReqHTTP)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create consent record: %d", createW.Code)
		}

		var createResponse ConsentResponse
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		// Test sending OTP
		otpReq := httptest.NewRequest("POST", fmt.Sprintf("/consent/%s/otp", createResponse.ID), bytes.NewBufferString(`{"phone_number": "+1234567890"}`))
		otpReq.Header.Set("Content-Type", "application/json")
		otpW := httptest.NewRecorder()

		server.sendConsentOTP(otpW, otpReq)

		if otpW.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, otpW.Code)
		}

		var otpResponse SMSOTPResponse
		if err := json.Unmarshal(otpW.Body.Bytes(), &otpResponse); err != nil {
			t.Fatalf("Failed to unmarshal OTP response: %v", err)
		}

		if !otpResponse.Success {
			t.Error("Expected OTP send to be successful")
		}

		if otpResponse.ConsentID != createResponse.ID {
			t.Errorf("Expected ConsentID=%s, got %s", createResponse.ID, otpResponse.ConsentID)
		}
	})
}

// TestConsentWorkflowErrorCases tests various error scenarios
func TestConsentWorkflowErrorCases(t *testing.T) {
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Test 1: Get non-existent consent
	t.Run("GetNonExistentConsent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consent/non-existent-id", nil)
		w := httptest.NewRecorder()

		server.getConsentStatus(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	// Test 2: Update non-existent consent
	t.Run("UpdateNonExistentConsent", func(t *testing.T) {
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "data_owner",
			Reason:    "User approved consent",
		}

		jsonBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/consent/non-existent-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.updateConsent(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	// Test 3: Revoke non-existent consent
	t.Run("RevokeNonExistentConsent", func(t *testing.T) {
		revokeReq := map[string]string{
			"reason": "User requested revocation",
		}

		jsonBody, _ := json.Marshal(revokeReq)
		req := httptest.NewRequest("DELETE", "/consent/non-existent-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.revokeConsent(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	// Test 4: Invalid status transition
	t.Run("InvalidStatusTransition", func(t *testing.T) {
		// First create a consent record
		createReq := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_invalid_transition",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(createReq)
		createReqHTTP := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		createReqHTTP.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		server.processConsentRequest(createW, createReqHTTP)

		if createW.Code != http.StatusCreated {
			t.Fatalf("Failed to create consent record: %d", createW.Code)
		}

		var createResponse ConsentResponse
		if err := json.Unmarshal(createW.Body.Bytes(), &createResponse); err != nil {
			t.Fatalf("Failed to unmarshal create response: %v", err)
		}

		// Try to revoke directly from pending (invalid transition)
		revokeReq := map[string]string{
			"reason": "User requested revocation",
		}

		revokeJsonBody, _ := json.Marshal(revokeReq)
		revokeReqHTTP := httptest.NewRequest("DELETE", fmt.Sprintf("/consent/%s", createResponse.ID), bytes.NewBuffer(revokeJsonBody))
		revokeReqHTTP.Header.Set("Content-Type", "application/json")
		revokeW := httptest.NewRecorder()

		server.revokeConsent(revokeW, revokeReqHTTP)

		// Should return error for invalid transition
		if revokeW.Code == http.StatusOK {
			t.Error("Expected error for invalid status transition, got success")
		}
	})
}

// TestConsentWorkflowEdgeCases tests edge cases and boundary conditions
func TestConsentWorkflowEdgeCases(t *testing.T) {
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Test 1: Very long field names
	t.Run("LongFieldNames", func(t *testing.T) {
		longFieldName := "very.long.field.name.that.might.cause.issues.in.the.system.and.should.be.handled.properly"

		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{longFieldName},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_long_field",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", w.Code, http.StatusCreated)
		}

		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response.Fields) != 1 || response.Fields[0] != longFieldName {
			t.Errorf("Expected field '%s', got %v", longFieldName, response.Fields)
		}
	})

	// Test 2: Special characters in field names
	t.Run("SpecialCharactersInFieldNames", func(t *testing.T) {
		specialFields := []string{
			"personInfo.name-with-dashes",
			"personInfo.name_with_underscores",
			"personInfo.name.with.dots",
			"personInfo.name@with@symbols",
		}

		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    specialFields,
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_special_chars",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", w.Code, http.StatusCreated)
		}

		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response.Fields) != len(specialFields) {
			t.Errorf("Expected %d fields, got %d", len(specialFields), len(response.Fields))
		}
	})

	// Test 3: Empty strings in required fields
	t.Run("EmptyStringsInRequiredFields", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "", // Empty app_id
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_empty_strings",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		// Should return error for empty app_id
		if w.Code == http.StatusCreated {
			t.Error("Expected error for empty app_id, got success")
		}
	})

	// Test 4: Very large number of fields
	t.Run("LargeNumberOfFields", func(t *testing.T) {
		var fields []string
		for i := 0; i < 1000; i++ {
			fields = append(fields, fmt.Sprintf("personInfo.field%d", i))
		}

		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    fields,
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_large_fields",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.processConsentRequest(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", w.Code, http.StatusCreated)
		}

		var response ConsentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response.Fields) != 1000 {
			t.Errorf("Expected 1000 fields, got %d", len(response.Fields))
		}
	})
}
