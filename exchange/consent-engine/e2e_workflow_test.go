package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types"
)

// TestEndToEndPassportApplicationWorkflow tests the complete passport application flow
// This test simulates the entire journey from user clicking "Apply" to final consent decision
func TestEndToEndPassportApplicationWorkflow(t *testing.T) {
	// Initialize the consent engine
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	t.Run("CompletePassportApplicationFlow", func(t *testing.T) {
		// Step 1: User clicks "Apply" button in PassportApp
		// This triggers a POST /getData call to orchestration-engine-go
		t.Log("Step 1: User clicks 'Apply' button in PassportApp")

		// Step 2: Orchestration-engine-go calls policy-decision-point
		// This is simulated by checking if consent is required
		t.Log("Step 2: Orchestration-engine-go calls policy-decision-point")

		// Simulate policy decision point response
		policyDecision := map[string]interface{}{
			"allow":                   true,
			"consent_required":        true,
			"consent_required_fields": []string{"person.permanentAddress", "person.birthDate"},
		}

		t.Logf("Policy Decision: %+v", policyDecision)

		// Step 3: Orchestration-engine-go calls consent-engine with POST /consents
		t.Log("Step 3: Orchestration-engine-go calls consent-engine with POST /consents")

		consentRequest := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.permanentAddress"},
				},
				{
					OwnerType: "citizen",
					OwnerID:   "1995000000",
					Fields:    []string{"personInfo.birthDate"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		// Create consent records
		jsonBody, _ := json.Marshal(consentRequest)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var consentResponse map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &consentResponse); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		t.Logf("Consent Response: %+v", consentResponse)

		// Step 4: PassportApp is redirected to ConsentPortal (popup window)
		t.Log("Step 4: PassportApp opens ConsentPortal popup window")

		// In a real scenario, we would have separate consent records for each owner
		// For this test, we'll create separate consent records
		consentID1 := extractConsentIDFromURL(consentResponse["redirect_url"].(string))

		// Create second consent record for the second owner
		consentRequest2 := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "1995000000",
					Fields:    []string{"personInfo.birthDate"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_456",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody2, _ := json.Marshal(consentRequest2)
		req2 := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()

		server.consentHandler(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w2.Code)
		}

		var consentResponse2 map[string]interface{}
		if err := json.Unmarshal(w2.Body.Bytes(), &consentResponse2); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consentID2 := extractConsentIDFromURL(consentResponse2["redirect_url"].(string))

		// Step 5: User approves first consent (permanentAddress)
		t.Log("Step 5: User approves first consent (permanentAddress)")

		approveData1 := map[string]string{
			"status":   "approved",
			"owner_id": "199512345678",
			"message":  "Approved via consent portal",
		}

		jsonBody1, _ := json.Marshal(approveData1)
		req1 := httptest.NewRequest("PUT", "/consents/"+consentID1, bytes.NewBuffer(jsonBody1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()

		server.consentHandler(w1, req1)

		if w1.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w1.Code)
		}

		// Step 6: OTP workflow for first consent
		t.Log("Step 6: OTP workflow for first consent")

		// Simulate OTP verification
		otpData1 := map[string]string{
			"otp_code": "123456",
		}

		jsonOTP1, _ := json.Marshal(otpData1)
		reqOTP1 := httptest.NewRequest("POST", "/consents/"+consentID1+"/otp", bytes.NewBuffer(jsonOTP1))
		reqOTP1.Header.Set("Content-Type", "application/json")
		wOTP1 := httptest.NewRecorder()

		server.consentHandler(wOTP1, reqOTP1)

		if wOTP1.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, wOTP1.Code)
		}

		// Step 7: User rejects second consent (birthDate)
		t.Log("Step 7: User rejects second consent (birthDate)")

		rejectData2 := map[string]string{
			"status":   "rejected",
			"owner_id": "1995000000",
			"message":  "Denied via consent portal",
		}

		jsonReject2, _ := json.Marshal(rejectData2)
		reqReject2 := httptest.NewRequest("PUT", "/consents/"+consentID2, bytes.NewBuffer(jsonReject2))
		reqReject2.Header.Set("Content-Type", "application/json")
		wReject2 := httptest.NewRecorder()

		server.consentHandler(wReject2, reqReject2)

		if wReject2.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, wReject2.Code)
		}

		// Step 8: Verify final consent statuses
		t.Log("Step 8: Verify final consent statuses")

		// Check first consent (should be approved)
		reqGet1 := httptest.NewRequest("GET", "/consents/"+consentID1, nil)
		wGet1 := httptest.NewRecorder()
		server.consentHandler(wGet1, reqGet1)

		if wGet1.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, wGet1.Code)
		}

		var consent1 map[string]interface{}
		if err := json.Unmarshal(wGet1.Body.Bytes(), &consent1); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if consent1["status"] != "approved" {
			t.Errorf("Expected status 'approved', got '%s'", consent1["status"])
		}

		// Check second consent (should be rejected)
		reqGet2 := httptest.NewRequest("GET", "/consents/"+consentID2, nil)
		wGet2 := httptest.NewRecorder()
		server.consentHandler(wGet2, reqGet2)

		if wGet2.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, wGet2.Code)
		}

		var consent2 map[string]interface{}
		if err := json.Unmarshal(wGet2.Body.Bytes(), &consent2); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if consent2["status"] != "rejected" {
			t.Errorf("Expected status 'rejected', got '%s'", consent2["status"])
		}

		// Step 9: Orchestration-engine-go processes consent results
		t.Log("Step 9: Orchestration-engine-go processes consent results")

		// Simulate orchestration engine decision
		consentResults := []map[string]interface{}{
			{
				"consent_id":   consentID1,
				"status":       "approved",
				"redirect_url": "https://passport-app.gov.lk/callback",
				"owner_id":     "199512345678",
				"fields":       []string{"personInfo.permanentAddress"},
			},
			{
				"consent_id":   consentID2,
				"status":       "rejected",
				"redirect_url": "https://passport-app.gov.lk/callback",
				"owner_id":     "1995000000",
				"fields":       []string{"personInfo.birthDate"},
			},
		}

		t.Logf("Consent Results: %+v", consentResults)

		// Step 10: PassportApp shows appropriate message to user
		t.Log("Step 10: PassportApp shows appropriate message to user")

		// Simulate PassportApp behavior based on consent results
		approvedFields := []string{}
		rejectedFields := []string{}

		for _, result := range consentResults {
			if result["status"] == "approved" {
				if fields, ok := result["fields"].([]string); ok {
					approvedFields = append(approvedFields, fields...)
				}
			} else {
				if fields, ok := result["fields"].([]string); ok {
					rejectedFields = append(rejectedFields, fields...)
				}
			}
		}

		if len(approvedFields) > 0 {
			t.Logf("✅ Approved fields: %v", approvedFields)
		}

		if len(rejectedFields) > 0 {
			t.Logf("❌ Rejected fields: %v", rejectedFields)
			t.Log("PassportApp shows: 'Request failed. Please manually enter these fields'")
		}

		// Step 11: Final validation
		t.Log("Step 11: Final validation")

		// Verify that we have the expected number of consent records
		// In a real scenario, this would be done by the orchestration engine
		expectedConsentCount := 2
		actualConsentCount := len(engine.(*consentEngineImpl).consentRecords)

		if actualConsentCount != expectedConsentCount {
			t.Errorf("Expected %d consent records, got %d", expectedConsentCount, actualConsentCount)
		}

		// Verify that the consent records have the correct statuses
		consent1Record, exists1 := engine.(*consentEngineImpl).consentRecords[consentID1]
		if !exists1 {
			t.Errorf("Consent record %s not found", consentID1)
		} else if consent1Record.Status != StatusApproved {
			t.Errorf("Expected consent %s to be approved, got %s", consentID1, consent1Record.Status)
		}

		consent2Record, exists2 := engine.(*consentEngineImpl).consentRecords[consentID2]
		if !exists2 {
			t.Errorf("Consent record %s not found", consentID2)
		} else if consent2Record.Status != StatusRejected {
			t.Errorf("Expected consent %s to be rejected, got %s", consentID2, consent2Record.Status)
		}

		t.Log("✅ End-to-end workflow completed successfully!")
	})
}

// TestEndToEndPassportApplicationWorkflowWithOTPFailure tests the complete flow with OTP failure
func TestEndToEndPassportApplicationWorkflowWithOTPFailure(t *testing.T) {
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	t.Run("CompletePassportApplicationFlowWithOTPFailure", func(t *testing.T) {
		// Step 1-4: Same as above (create consent, redirect to portal)
		t.Log("Step 1-4: Create consent and redirect to portal")

		consentRequest := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.permanentAddress"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(consentRequest)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var consentResponse map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &consentResponse); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consentID := extractConsentIDFromURL(consentResponse["redirect_url"].(string))

		// Step 5: User approves consent
		t.Log("Step 5: User approves consent")

		approveData := map[string]string{
			"status":   "approved",
			"owner_id": "199512345678",
			"message":  "Approved via consent portal",
		}

		jsonBody, _ = json.Marshal(approveData)
		req = httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Step 6: OTP workflow with 3 failed attempts
		t.Log("Step 6: OTP workflow with 3 failed attempts")

		wrongOTPData := map[string]string{
			"otp_code": "000000",
		}

		// Try wrong OTP 3 times
		for i := 1; i <= 3; i++ {
			t.Logf("OTP Attempt %d/3", i)

			jsonOTP, _ := json.Marshal(wrongOTPData)
			reqOTP := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonOTP))
			reqOTP.Header.Set("Content-Type", "application/json")
			wOTP := httptest.NewRecorder()

			server.consentHandler(wOTP, reqOTP)

			if i < 3 {
				// First two attempts should return 400 with retry message
				if wOTP.Code != http.StatusBadRequest {
					t.Errorf("Attempt %d: Expected status %d, got %d", i, http.StatusBadRequest, wOTP.Code)
				}
			} else {
				// Third attempt should return 400 with rejection message
				if wOTP.Code != http.StatusBadRequest {
					t.Errorf("Attempt %d: Expected status %d, got %d", i, http.StatusBadRequest, wOTP.Code)
				}
			}
		}

		// Step 7: Verify consent is now rejected due to OTP failure
		t.Log("Step 7: Verify consent is now rejected due to OTP failure")

		reqGet := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		wGet := httptest.NewRecorder()
		server.consentHandler(wGet, reqGet)

		if wGet.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, wGet.Code)
		}

		var consent map[string]interface{}
		if err := json.Unmarshal(wGet.Body.Bytes(), &consent); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if consent["status"] != "rejected" {
			t.Errorf("Expected status 'rejected' after OTP failure, got '%s'", consent["status"])
		}

		// Step 8: PassportApp shows failure message
		t.Log("Step 8: PassportApp shows failure message")
		t.Log("PassportApp shows: 'Request failed. Please manually enter these fields'")

		t.Log("✅ End-to-end workflow with OTP failure completed successfully!")
	})
}

// TestEndToEndPassportApplicationWorkflowWithExistingConsent tests the flow when consent already exists
func TestEndToEndPassportApplicationWorkflowWithExistingConsent(t *testing.T) {
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	t.Run("CompletePassportApplicationFlowWithExistingConsent", func(t *testing.T) {
		// Step 1: Create initial consent
		t.Log("Step 1: Create initial consent")

		consentRequest1 := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.permanentAddress"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(consentRequest1)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response1 map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response1); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consentID1 := extractConsentIDFromURL(response1["redirect_url"].(string))

		// Step 2: Try to create consent again with same (app_id, owner_id) pair
		t.Log("Step 2: Try to create consent again with same (app_id, owner_id) pair")

		consentRequest2 := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.birthDate"}, // Different field
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_456", // Different session
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ = json.Marshal(consentRequest2)
		req = httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response2 map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response2); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consentID2 := extractConsentIDFromURL(response2["redirect_url"].(string))

		// Step 3: Verify that the same consent ID is returned
		t.Log("Step 3: Verify that the same consent ID is returned")

		if consentID1 != consentID2 {
			t.Errorf("Expected same consent ID for existing (app_id, owner_id) pair, got %s and %s", consentID1, consentID2)
		}

		// Step 4: Verify that only one consent record exists
		t.Log("Step 4: Verify that only one consent record exists")

		consentCount := len(engine.(*consentEngineImpl).consentRecords)
		if consentCount != 1 {
			t.Errorf("Expected 1 consent record, got %d", consentCount)
		}

		t.Log("✅ End-to-end workflow with existing consent completed successfully!")
	})
}
