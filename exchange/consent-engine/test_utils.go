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

// TestHelper provides common test utilities
type TestHelper struct {
	engine *consentEngineImpl
	server *apiServer
}

// NewTestHelper creates a new test helper with fresh engine and server
func NewTestHelper(t *testing.T) *TestHelper {
	engine := NewConsentEngine().(*consentEngineImpl)
	server := &apiServer{engine: engine}
	return &TestHelper{
		engine: engine,
		server: server,
	}
}

// CreateTestConsent creates a consent and returns the consent ID
func (th *TestHelper) CreateTestConsent(t *testing.T, ownerID string) string {
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

	th.server.consentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	return extractConsentIDFromURL(response["redirect_url"].(string))
}

// ApproveConsent approves a consent
func (th *TestHelper) ApproveConsent(t *testing.T, consentID, ownerID string) {
	approveData := map[string]string{
		"status":   "approved",
		"owner_id": ownerID,
		"message":  "Approved via consent portal",
	}

	jsonBody, _ := json.Marshal(approveData)
	req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	th.server.consentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// RejectConsent rejects a consent
func (th *TestHelper) RejectConsent(t *testing.T, consentID, ownerID string) {
	rejectData := map[string]string{
		"status":   "rejected",
		"owner_id": ownerID,
		"message":  "Rejected via consent portal",
	}

	jsonBody, _ := json.Marshal(rejectData)
	req := httptest.NewRequest("PUT", "/consents/"+consentID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	th.server.consentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// VerifyConsentStatus verifies a consent's status
func (th *TestHelper) VerifyConsentStatus(t *testing.T, consentID, expectedStatus string) {
	req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
	w := httptest.NewRecorder()

	th.server.consentHandler(w, req)

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

// VerifyOTP verifies an OTP code
func (th *TestHelper) VerifyOTP(t *testing.T, consentID, otpCode string) int {
	otpData := map[string]string{
		"otp_code": otpCode,
	}

	jsonBody, _ := json.Marshal(otpData)
	req := httptest.NewRequest("POST", "/consents/"+consentID+"/otp", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	th.server.consentHandler(w, req)

	return w.Code
}

// ExtractConsentIDFromURL extracts consent ID from redirect URL
func extractConsentIDFromURL(redirectURL string) string {
	parts := strings.Split(redirectURL, "consent_id=")
	if len(parts) > 1 {
		return strings.Split(parts[1], "&")[0]
	}
	return ""
}
