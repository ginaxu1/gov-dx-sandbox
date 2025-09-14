package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types"
)

// TestConsentWorkflowIntegration tests the complete consent management workflow via HTTP API
func TestConsentWorkflowIntegration(t *testing.T) {
	// Initialize the consent engine
	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Test 1: Create New Consent (Non-existing)
	t.Run("CreateNewConsent", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		// Check response status
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		// Parse response - now returns simplified format
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Validate response fields - simplified format
		if response["status"] != "pending" {
			t.Errorf("Expected status 'pending', got '%s'", response["status"])
		}
		if response["redirect_url"] == "" {
			t.Error("Expected non-empty redirect_url")
		}
		// Check that redirect_url contains consent_id
		redirectURL, ok := response["redirect_url"].(string)
		if !ok || !strings.Contains(redirectURL, "consent_id=") {
			t.Errorf("Expected redirect_url to contain consent_id parameter, got: %s", redirectURL)
		}
	})

	// Test 2: Create Consent with Existing (owner_id, app_id) Pair
	t.Run("CreateConsentWithExistingPair", func(t *testing.T) {
		// First create a consent
		reqBody1 := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "1991111111",
					Fields:    []string{"personInfo.address"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody1, _ := json.Marshal(reqBody1)
		req1 := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()

		server.consentHandler(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w1.Code)
		}

		var response1 map[string]interface{}
		if err := json.Unmarshal(w1.Body.Bytes(), &response1); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		firstConsentID := extractConsentIDFromURL(response1["redirect_url"].(string))

		// Now try to create another consent with the same (owner_id, app_id) pair
		reqBody2 := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "1991111111",
					Fields:    []string{"personInfo.birthDate"},
				},
			},
			Purpose:     "passport_application",
			SessionID:   "session_456",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody2, _ := json.Marshal(reqBody2)
		req2 := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()

		server.consentHandler(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w2.Code)
		}

		var response2 map[string]interface{}
		if err := json.Unmarshal(w2.Body.Bytes(), &response2); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		secondConsentID := extractConsentIDFromURL(response2["redirect_url"].(string))

		// Should return the same consent ID (existing record)
		if firstConsentID != secondConsentID {
			t.Errorf("Expected same consent ID for existing (owner_id, app_id) pair, got %s and %s", firstConsentID, secondConsentID)
		}
	})

	// Test 3: Approve Consent
	t.Run("ApproveConsent", func(t *testing.T) {
		// Create a consent first
		consentID := createTestConsent(t, server, "1992222222")

		// Approve the consent
		approveData := map[string]string{
			"status":   "approved",
			"owner_id": "1992222222",
			"message":  "Approved via consent portal",
		}

		jsonBody, _ := json.Marshal(approveData)
		req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the consent status
		verifyConsentStatus(t, server, consentID, "approved")
	})

	// Test 4: Reject Consent
	t.Run("RejectConsent", func(t *testing.T) {
		// Create a consent first
		consentID := createTestConsent(t, server, "1993333333")

		// Reject the consent
		rejectData := map[string]string{
			"status":   "rejected",
			"owner_id": "1993333333",
			"message":  "Rejected via consent portal",
		}

		jsonBody, _ := json.Marshal(rejectData)
		req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the consent status
		verifyConsentStatus(t, server, consentID, "rejected")
	})

	// Test 5: OTP Verification - Correct OTP
	t.Run("OTPVerificationCorrect", func(t *testing.T) {
		// Create and approve a consent first
		consentID := createTestConsent(t, server, "1994444444")
		approveConsent(t, server, consentID, "1994444444")

		// Verify OTP with correct code
		otpData := map[string]string{
			"otp_code": "123456",
		}

		jsonBody, _ := json.Marshal(otpData)
		req := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the consent status remains approved
		verifyConsentStatus(t, server, consentID, "approved")
	})

	// Test 6: OTP Verification - Incorrect OTP (3 attempts)
	t.Run("OTPVerificationIncorrect", func(t *testing.T) {
		// Create and approve a consent first
		consentID := createTestConsent(t, server, "1995555555")
		approveConsent(t, server, consentID, "1995555555")

		// Try wrong OTP 3 times
		wrongOTPData := map[string]string{
			"otp_code": "000000",
		}

		for i := 1; i <= 3; i++ {
			jsonBody, _ := json.Marshal(wrongOTPData)
			req := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.consentHandler(w, req)

			if i < 3 {
				// First two attempts should return 400 with retry message
				if w.Code != http.StatusBadRequest {
					t.Errorf("Attempt %d: Expected status %d, got %d", i, http.StatusBadRequest, w.Code)
				}
			} else {
				// Third attempt should return 400 with rejection message
				if w.Code != http.StatusBadRequest {
					t.Errorf("Attempt %d: Expected status %d, got %d", i, http.StatusBadRequest, w.Code)
				}
			}
		}

		// Verify the consent status is now rejected
		verifyConsentStatus(t, server, consentID, "rejected")
	})

	// Test 7: OTP on Pending Consent (should fail)
	t.Run("OTPOnPendingConsent", func(t *testing.T) {
		// Create a consent but don't approve it
		consentID := createTestConsent(t, server, "1996666666")

		// Try OTP on pending consent
		otpData := map[string]string{
			"otp_code": "123456",
		}

		jsonBody, _ := json.Marshal(otpData)
		req := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Test 8: OTP on Rejected Consent (should fail)
	t.Run("OTPOnRejectedConsent", func(t *testing.T) {
		// Create and reject a consent
		consentID := createTestConsent(t, server, "1997777777")
		rejectConsent(t, server, consentID, "1997777777")

		// Try OTP on rejected consent
		otpData := map[string]string{
			"otp_code": "123456",
		}

		jsonBody, _ := json.Marshal(otpData)
		req := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// Test 9: Get Non-existent Consent
	t.Run("GetNonExistentConsent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consents/non-existent-id", nil)
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	// Test 10: Update Non-existent Consent
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

	// Test 11: Revoke Consent
	t.Run("RevokeConsent", func(t *testing.T) {
		// Create and approve a consent first
		consentID := createTestConsent(t, server, "1999999999")
		approveConsent(t, server, consentID, "1999999999")

		// Revoke the consent
		revokeData := map[string]string{
			"reason": "User requested revocation",
		}

		jsonBody, _ := json.Marshal(revokeData)
		req := httptest.NewRequest("DELETE", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the consent status is revoked
		verifyConsentStatus(t, server, consentID, "revoked")
	})
}

// Helper functions for DRY testing

func createTestConsent(t *testing.T, server *apiServer, ownerID string) string {
	reqBody := ConsentRequest{
		AppID: "passport-app",
		DataFields: []types.DataField{
			{
				OwnerType: "citizen",
				OwnerID:   ownerID,
				Fields:    []string{"personInfo.address"},
			},
		},
		Purpose:     "passport_application",
		SessionID:   "session_test",
		RedirectURL: "https://passport-app.gov.lk/callback",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.consentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return extractConsentIDFromURL(response["redirect_url"].(string))
}

func approveConsent(t *testing.T, server *apiServer, consentID, ownerID string) {
	approveData := map[string]string{
		"status":   "approved",
		"owner_id": ownerID,
		"message":  "Approved via consent portal",
	}

	jsonBody, _ := json.Marshal(approveData)
	req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.consentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func rejectConsent(t *testing.T, server *apiServer, consentID, ownerID string) {
	rejectData := map[string]string{
		"status":   "rejected",
		"owner_id": ownerID,
		"message":  "Rejected via consent portal",
	}

	jsonBody, _ := json.Marshal(rejectData)
	req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.consentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func verifyConsentStatus(t *testing.T, server *apiServer, consentID, expectedStatus string) {
	req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
	w := httptest.NewRecorder()

	server.consentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != expectedStatus {
		t.Errorf("Expected status %s, got %s", expectedStatus, response["status"])
	}
}

func extractConsentIDFromURL(redirectURL string) string {
	// Extract consent_id from URL like "http://localhost:5173/?consent_id=consent_abc123"
	parts := strings.Split(redirectURL, "consent_id=")
	if len(parts) > 1 {
		return strings.Split(parts[1], "&")[0]
	}
	return ""
}
