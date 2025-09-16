package main

import (
	"testing"
	"time"
)

// TestConsentEngine_CreateConsent tests the core CreateConsent functionality
func TestConsentEngine_CreateConsent(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	req := ConsentRequest{
		AppID:       "passport-app",
		Purpose:     "passport_application",
		SessionID:   "session_123",
		RedirectURL: "https://passport-app.gov.lk",
		DataFields: []DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "user123",
				Fields:    []string{"person.permanentAddress"},
			},
		},
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	if record.ConsentID == "" {
		t.Error("Expected non-empty consent ID")
	}

	if record.Status != string(StatusPending) {
		t.Errorf("Expected status=%s, got %s", string(StatusPending), record.Status)
	}

	if record.AppID != req.AppID {
		t.Errorf("Expected AppID=%s, got %s", req.AppID, record.AppID)
	}

	if record.OwnerID != req.DataFields[0].OwnerID {
		t.Errorf("Expected OwnerID=%s, got %s", req.DataFields[0].OwnerID, record.OwnerID)
	}

	if record.OTPAttempts != 0 {
		t.Errorf("Expected OTPAttempts=0, got %d", record.OTPAttempts)
	}
}

// TestConsentEngine_GetConsentStatus tests retrieving consent status
func TestConsentEngine_GetConsentStatus(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	// Create a consent record first
	req := ConsentRequest{
		AppID:       "passport-app",
		Purpose:     "passport_application",
		SessionID:   "session_123",
		RedirectURL: "https://passport-app.gov.lk",
		DataFields: []DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "user123",
				Fields:    []string{"person.permanentAddress"},
			},
		},
	}

	createdRecord, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Test getting the consent status
	record, err := engine.GetConsentStatus(createdRecord.ConsentID)
	if err != nil {
		t.Fatalf("GetConsentStatus failed: %v", err)
	}

	if record.ConsentID != createdRecord.ConsentID {
		t.Errorf("Expected ConsentID=%s, got %s", createdRecord.ConsentID, record.ConsentID)
	}

	// Test getting non-existent consent
	_, err = engine.GetConsentStatus("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent consent ID")
	}
}

// TestConsentEngine_UpdateConsent tests updating consent status
func TestConsentEngine_UpdateConsent(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	// Create a consent record first
	req := ConsentRequest{
		AppID:       "passport-app",
		Purpose:     "passport_application",
		SessionID:   "session_123",
		RedirectURL: "https://passport-app.gov.lk",
		DataFields: []DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "user123",
				Fields:    []string{"person.permanentAddress"},
			},
		},
	}

	createdRecord, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Test updating consent status
	updateReq := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "user123",
		Reason:    "User approved consent",
	}

	updatedRecord, err := engine.UpdateConsent(createdRecord.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	if updatedRecord.Status != string(StatusApproved) {
		t.Errorf("Expected status=%s, got %s", string(StatusApproved), updatedRecord.Status)
	}

	// Test invalid status transition (approved -> pending is not allowed)
	invalidUpdateReq := UpdateConsentRequest{
		Status:    StatusPending,
		UpdatedBy: "user123",
		Reason:    "Invalid transition",
	}

	_, err = engine.UpdateConsent(createdRecord.ConsentID, invalidUpdateReq)
	if err == nil {
		t.Error("Expected error for invalid status transition from approved to pending")
	}
}

// TestConsentEngine_FindExistingConsent tests finding existing consents
func TestConsentEngine_FindExistingConsent(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	// Create a consent record first
	req := ConsentRequest{
		AppID:       "passport-app",
		Purpose:     "passport_application",
		SessionID:   "session_123",
		RedirectURL: "https://passport-app.gov.lk",
		DataFields: []DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "user123",
				Fields:    []string{"person.permanentAddress"},
			},
		},
	}

	createdRecord, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Test finding existing consent
	foundRecord := engine.FindExistingConsent("passport-app", "user123")
	if foundRecord == nil {
		t.Error("Expected to find existing consent record")
	}

	if foundRecord.ConsentID != createdRecord.ConsentID {
		t.Errorf("Expected ConsentID=%s, got %s", createdRecord.ConsentID, foundRecord.ConsentID)
	}

	// Test finding non-existent consent
	notFoundRecord := engine.FindExistingConsent("different-app", "user123")
	if notFoundRecord != nil {
		t.Error("Expected not to find consent record for different app")
	}
}

// TestConsentEngine_UpdateConsentRecord tests direct record updates
func TestConsentEngine_UpdateConsentRecord(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	// Create a consent record first
	req := ConsentRequest{
		AppID:       "passport-app",
		Purpose:     "passport_application",
		SessionID:   "session_123",
		RedirectURL: "https://passport-app.gov.lk",
		DataFields: []DataField{
			{
				OwnerType: "citizen",
				OwnerID:   "user123",
				Fields:    []string{"person.permanentAddress"},
			},
		},
	}

	createdRecord, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Update the record directly
	createdRecord.OTPAttempts = 2
	createdRecord.UpdatedAt = time.Now()

	err = engine.UpdateConsentRecord(createdRecord)
	if err != nil {
		t.Fatalf("UpdateConsentRecord failed: %v", err)
	}

	// Verify the update
	updatedRecord, err := engine.GetConsentStatus(createdRecord.ConsentID)
	if err != nil {
		t.Fatalf("GetConsentStatus failed: %v", err)
	}

	if updatedRecord.OTPAttempts != 2 {
		t.Errorf("Expected OTPAttempts=2, got %d", updatedRecord.OTPAttempts)
	}
}

// TestConsentEngine_DefaultRecord tests the default record functionality
func TestConsentEngine_DefaultRecord(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")

	// Test that the default record exists
	defaultRecord, err := engine.GetConsentStatus("consent_03c134ae")
	if err != nil {
		t.Fatalf("Expected default record to exist: %v", err)
	}

	if defaultRecord.OwnerID != "199512345678" {
		t.Errorf("Expected OwnerID=199512345678, got %s", defaultRecord.OwnerID)
	}

	if defaultRecord.AppID != "passport-app" {
		t.Errorf("Expected AppID=passport-app, got %s", defaultRecord.AppID)
	}

	if defaultRecord.Status != string(StatusPending) {
		t.Errorf("Expected Status=%s, got %s", string(StatusPending), defaultRecord.Status)
	}

	if defaultRecord.OTPAttempts != 0 {
		t.Errorf("Expected OTPAttempts=0, got %d", defaultRecord.OTPAttempts)
	}
}

// TestConsentEngine_StatusTransitions tests status transition validation
func TestConsentEngine_StatusTransitions(t *testing.T) {
	// Test valid transitions
	validTransitions := []struct {
		from  ConsentStatus
		to    ConsentStatus
		valid bool
	}{
		{StatusPending, StatusApproved, true},
		{StatusPending, StatusRejected, true},
		{StatusApproved, StatusApproved, true}, // OTP flow
		{StatusApproved, StatusRejected, true}, // OTP failure
		{StatusRejected, StatusPending, true},  // Retry
		{StatusApproved, StatusPending, false}, // Invalid
		{StatusRejected, StatusApproved, true}, // Valid - allow direct approval from rejected
	}

	for _, tt := range validTransitions {
		result := isValidStatusTransition(tt.from, tt.to)
		if result != tt.valid {
			t.Errorf("isValidStatusTransition(%v, %v) = %v, want %v", tt.from, tt.to, result, tt.valid)
		}
	}
}
