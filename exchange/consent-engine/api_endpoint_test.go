package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types"
)

// TestPOSTConsentsEndpoint tests the POST /consents endpoint
func TestPOSTConsentsEndpoint(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("CreateNewConsent_Success", func(t *testing.T) {
		reqBody := ConsentRequest{
			AppID: "passport-app",
			DataFields: []types.DataField{
				{
					OwnerType: "citizen",
					OwnerID:   "199512345678",
					Fields:    []string{"person.permanentAddress", "person.birthDate"},
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

		th.server.consentHandler(w, req)

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

	t.Run("CreateConsentWithExistingPair", func(t *testing.T) {
		// First create a consent
		consentID1 := th.CreateTestConsent(t, "1991111111")

		// Try to create another consent with the same (owner_id, app_id) pair
		reqBody := ConsentRequest{
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

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.server.consentHandler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		consentID2 := extractConsentIDFromURL(response["redirect_url"].(string))

		// Should return the same consent ID (existing record)
		if consentID1 != consentID2 {
			t.Errorf("Expected same consent ID for existing (owner_id, app_id) pair, got %s and %s", consentID1, consentID2)
		}
	})

	t.Run("CreateConsent_InvalidRequest", func(t *testing.T) {
		// Test with empty data fields
		reqBody := ConsentRequest{
			AppID:       "passport-app",
			DataFields:  []types.DataField{},
			Purpose:     "passport_application",
			SessionID:   "session_123",
			RedirectURL: "https://passport-app.gov.lk/callback",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/consents", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.server.consentHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// TestPUTConsentsEndpoint tests the PUT /consents/{id} endpoint
func TestPUTConsentsEndpoint(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("ApproveConsent_Success", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1992222222")
		th.ApproveConsent(t, consentID, "1992222222")
		th.VerifyConsentStatus(t, consentID, "approved")
	})

	t.Run("RejectConsent_Success", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1993333333")
		th.RejectConsent(t, consentID, "1993333333")
		th.VerifyConsentStatus(t, consentID, "rejected")
	})

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

		th.server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestGETConsentsEndpoint tests the GET /consents/{id} endpoint
func TestGETConsentsEndpoint(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("GetConsent_Success", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1994444444")

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		w := httptest.NewRecorder()

		th.server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["consent_uuid"] != consentID {
			t.Errorf("Expected consent_uuid %s, got %s", consentID, response["consent_uuid"])
		}
	})

	t.Run("GetNonExistentConsent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consents/non-existent-id", nil)
		w := httptest.NewRecorder()

		th.server.consentHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// TestOTPEndpoint tests the POST /consents/{id}/otp endpoint
func TestOTPEndpoint(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("OTPVerification_CorrectOTP", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1995555555")
		th.ApproveConsent(t, consentID, "1995555555")

		// Verify OTP with correct code
		code := th.VerifyOTP(t, consentID, "123456")
		if code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, code)
		}

		th.VerifyConsentStatus(t, consentID, "approved")
	})

	t.Run("OTPVerification_IncorrectOTP", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1996666666")
		th.ApproveConsent(t, consentID, "1996666666")

		// Try wrong OTP 3 times
		for i := 1; i <= 3; i++ {
			code := th.VerifyOTP(t, consentID, "000000")
			if code != http.StatusBadRequest {
				t.Errorf("Attempt %d: Expected status %d, got %d", i, http.StatusBadRequest, code)
			}
		}

		// Verify the consent status is now rejected
		th.VerifyConsentStatus(t, consentID, "rejected")
	})

	t.Run("OTPOnPendingConsent", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1997777777")

		// Try OTP on pending consent
		code := th.VerifyOTP(t, consentID, "123456")
		if code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, code)
		}
	})

	t.Run("OTPOnRejectedConsent", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1998888888")
		th.RejectConsent(t, consentID, "1998888888")

		// Try OTP on rejected consent
		code := th.VerifyOTP(t, consentID, "123456")
		if code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, code)
		}
	})
}

// TestDELETEConsentsEndpoint tests the DELETE /consents/{id} endpoint
func TestDELETEConsentsEndpoint(t *testing.T) {
	th := NewTestHelper(t)

	t.Run("RevokeConsent_Success", func(t *testing.T) {
		consentID := th.CreateTestConsent(t, "1999999999")
		th.ApproveConsent(t, consentID, "1999999999")

		// Revoke the consent
		revokeData := map[string]string{
			"reason": "User requested revocation",
		}

		jsonBody, _ := json.Marshal(revokeData)
		req := httptest.NewRequest("DELETE", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		th.server.consentHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the consent status is revoked
		th.VerifyConsentStatus(t, consentID, "revoked")
	})
}
