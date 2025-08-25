package main

import (
	"testing"
	"time"
)

func TestConsentEngine_CreateConsent(t *testing.T) {
	engine := NewConsentEngine()

	req := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
		SessionID:    "sess_123",
		ExpiryTime:   "30d",
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	if record.ConsentID == "" {
		t.Error("Expected non-empty consent ID")
	}

	if record.Status != StatusPending {
		t.Errorf("Expected status=StatusPending, got %v", record.Status)
	}

	if record.DataConsumer != req.DataConsumer {
		t.Errorf("Expected DataConsumer=%s, got %s", req.DataConsumer, record.DataConsumer)
	}

	if record.OwnerID != req.DataOwner {
		t.Errorf("Expected DataOwner=%s, got %s", req.DataOwner, record.OwnerID)
	}
}

func TestConsentEngine_GetConsentStatus(t *testing.T) {
	engine := NewConsentEngine()

	// Create a consent record first
	req := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Test getting the consent status
	retrieved, err := engine.GetConsentStatus(record.ConsentID)
	if err != nil {
		t.Fatalf("GetConsentStatus failed: %v", err)
	}

	if retrieved.ConsentID != record.ConsentID {
		t.Errorf("Expected ID=%s, got %s", record.ConsentID, retrieved.ConsentID)
	}

	if retrieved.Status != StatusPending {
		t.Errorf("Expected Status=StatusPending, got %v", retrieved.Status)
	}
}

func TestConsentEngine_UpdateConsent(t *testing.T) {
	engine := NewConsentEngine()

	// Create a consent record first
	req := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Update the consent status
	updateReq := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "user123",
		Reason:    "User approved via portal",
	}

	updated, err := engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	if updated.Status != StatusApproved {
		t.Errorf("Expected Status=StatusApproved, got %v", updated.Status)
	}

	if updated.UpdatedAt.Before(record.UpdatedAt) {
		t.Error("Expected UpdatedAt to be after original UpdatedAt")
	}
}

func TestConsentEngine_RevokeConsent(t *testing.T) {
	engine := NewConsentEngine()

	// Create a consent record first
	req := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// First approve the consent (required before revocation)
	updateReq := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "user123",
		Reason:    "User approved consent",
	}

	approved, err := engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	// Now revoke the approved consent
	revoked, err := engine.RevokeConsent(approved.ConsentID, "User requested revocation")
	if err != nil {
		t.Fatalf("RevokeConsent failed: %v", err)
	}

	if revoked.Status != StatusRevoked {
		t.Errorf("Expected Status=StatusRevoked, got %v", revoked.Status)
	}
}

func TestConsentEngine_GetConsentsByDataOwner(t *testing.T) {
	engine := NewConsentEngine()

	// Create multiple consent records for the same data owner
	req1 := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
	}

	req2 := CreateConsentRequest{
		DataConsumer: "other-app",
		DataOwner:    "user123",
		Fields:       []string{"person.birthDate"},
		Type:         ConsentTypeRealTime,
	}

	_, err := engine.CreateConsent(req1)
	if err != nil {
		t.Fatalf("CreateConsent 1 failed: %v", err)
	}

	_, err = engine.CreateConsent(req2)
	if err != nil {
		t.Fatalf("CreateConsent 2 failed: %v", err)
	}

	// Get consents by data owner
	consents, err := engine.GetConsentsByDataOwner("user123")
	if err != nil {
		t.Fatalf("GetConsentsByDataOwner failed: %v", err)
	}

	if len(consents) != 2 {
		t.Errorf("Expected 2 consents, got %d", len(consents))
	}
}

func TestConsentEngine_GetConsentsByConsumer(t *testing.T) {
	engine := NewConsentEngine()

	// Create multiple consent records for the same consumer
	req1 := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
	}

	req2 := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user456",
		Fields:       []string{"person.birthDate"},
		Type:         ConsentTypeRealTime,
	}

	_, err := engine.CreateConsent(req1)
	if err != nil {
		t.Fatalf("CreateConsent 1 failed: %v", err)
	}

	_, err = engine.CreateConsent(req2)
	if err != nil {
		t.Fatalf("CreateConsent 2 failed: %v", err)
	}

	// Get consents by consumer
	consents, err := engine.GetConsentsByConsumer("passport-app")
	if err != nil {
		t.Fatalf("GetConsentsByConsumer failed: %v", err)
	}

	if len(consents) != 2 {
		t.Errorf("Expected 2 consents, got %d", len(consents))
	}
}

func TestConsentEngine_CheckConsentExpiry(t *testing.T) {
	engine := NewConsentEngine()

	// Create a consent record with short expiry
	req := CreateConsentRequest{
		DataConsumer: "passport-app",
		DataOwner:    "user123",
		Fields:       []string{"person.permanentAddress"},
		Type:         ConsentTypeRealTime,
		ExpiryTime:   "1s", // Very short expiry for testing
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// First approve the consent (required for expiry check)
	updateReq := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "user123",
		Reason:    "User approved consent",
	}

	approved, err := engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Check expiry
	expiredRecords, err := engine.CheckConsentExpiry()
	if err != nil {
		t.Fatalf("CheckConsentExpiry failed: %v", err)
	}

	if len(expiredRecords) == 0 {
		t.Error("Expected at least one expired record")
	}

	// Verify the record is now expired
	retrieved, err := engine.GetConsentStatus(approved.ConsentID)
	if err != nil {
		t.Fatalf("GetConsentStatus failed: %v", err)
	}

	if retrieved.Status != StatusExpired {
		t.Errorf("Expected Status=StatusExpired, got %v", retrieved.Status)
	}
}
