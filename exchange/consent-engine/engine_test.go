package main

import (
	"testing"
	"time"
)

// TestConsentEngine_CreateConsent tests the core CreateConsent functionality
func TestConsentEngine_CreateConsent(t *testing.T) {
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
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

	if record.OwnerEmail != req.DataFields[0].OwnerEmail {
		t.Errorf("Expected OwnerEmail=%s, got %s", req.DataFields[0].OwnerEmail, record.OwnerEmail)
	}

}

// TestConsentEngine_GetConsentStatus tests retrieving consent status
func TestConsentEngine_GetConsentStatus(t *testing.T) {
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	// Create a consent record first
	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
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
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	// Create a consent record first
	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
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
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	// Create a consent record first
	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
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
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	// Create a consent record first
	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
			},
		},
	}

	createdRecord, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Update the record directly
	createdRecord.UpdatedAt = time.Now()

	err = engine.UpdateConsentRecord(createdRecord)
	if err != nil {
		t.Fatalf("UpdateConsentRecord failed: %v", err)
	}

	// Verify the update
	_, err = engine.GetConsentStatus(createdRecord.ConsentID)
	if err != nil {
		t.Fatalf("GetConsentStatus failed: %v", err)
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
		{StatusApproved, StatusApproved, true}, // Direct approval
		{StatusApproved, StatusRejected, true}, // Direct rejection
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

// TestConsentEngine_CheckConsentExpiry tests the CheckConsentExpiry functionality
func TestConsentEngine_CheckConsentExpiry(t *testing.T) {
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	t.Run("NoExpiredRecords", func(t *testing.T) {
		// Create a consent that won't expire soon
		req := ConsentRequest{
			AppID:     "passport-app",
			Purpose:   "passport_application",
			SessionID: "session_123",
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

		// Check expiry - should return empty list
		expiredRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(expiredRecords) != 0 {
			t.Errorf("Expected 0 expired records, got %d", len(expiredRecords))
		}
	})

	t.Run("HasExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
		engine := NewConsentEngine(consentPortalURL)

		// Create a consent
		req := ConsentRequest{
			AppID:     "passport-app",
			Purpose:   "passport_application",
			SessionID: "session_456",
			DataFields: []DataField{
				{
					OwnerType:  "citizen",
					OwnerID:    "user456",
					OwnerEmail: "user456@example.com",
					Fields:     []string{"person.permanentAddress"},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Approve the consent
		updateReq := UpdateConsentRequest{
			Status:    StatusApproved,
			UpdatedBy: "citizen_456",
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

		// Check expiry - should return the expired record
		expiredRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(expiredRecords) != 1 {
			t.Errorf("Expected 1 expired record, got %d", len(expiredRecords))
		}

		if expiredRecords[0].ConsentID != record.ConsentID {
			t.Errorf("Expected expired record ID %s, got %s", record.ConsentID, expiredRecords[0].ConsentID)
		}

		if expiredRecords[0].Status != string(StatusExpired) {
			t.Errorf("Expected expired record status %s, got %s", string(StatusExpired), expiredRecords[0].Status)
		}
	})

	t.Run("OnlyApprovedRecordsExpire", func(t *testing.T) {
		// Create a new engine to avoid interference
		consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
		engine := NewConsentEngine(consentPortalURL)

		// Create a consent
		req := ConsentRequest{
			AppID:     "passport-app",
			Purpose:   "passport_application",
			SessionID: "session_789",
			DataFields: []DataField{
				{
					OwnerType:  "citizen",
					OwnerID:    "user789",
					OwnerEmail: "user789@example.com",
					Fields:     []string{"person.permanentAddress"},
				},
			},
		}

		record, err := engine.CreateConsent(req)
		if err != nil {
			t.Fatalf("CreateConsent failed: %v", err)
		}

		// Reject the consent (not approved)
		updateReq := UpdateConsentRequest{
			Status:    StatusRejected,
			UpdatedBy: "citizen_789",
			Reason:    "User rejected",
		}
		_, err = engine.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			t.Fatalf("UpdateConsent failed: %v", err)
		}

		// Manually set the expiry time to the past
		record.ExpiresAt = time.Now().Add(-1 * time.Hour)
		engineImpl := engine.(*consentEngineImpl)
		engineImpl.consentRecords[record.ConsentID] = record

		// Check expiry - should return empty list (rejected records don't expire)
		expiredRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(expiredRecords) != 0 {
			t.Errorf("Expected 0 expired records (rejected records don't expire), got %d", len(expiredRecords))
		}
	})

	t.Run("MultipleExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
		engine := NewConsentEngine(consentPortalURL)

		// Create multiple consents
		consentIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			req := ConsentRequest{
				AppID:     "passport-app",
				Purpose:   "passport_application",
				SessionID: "session_" + string(rune('0'+i)),
				DataFields: []DataField{
					{
						OwnerType:  "citizen",
						OwnerID:    "user" + string(rune('0'+i)),
						OwnerEmail: "user" + string(rune('0'+i)) + "@example.com",
						Fields:     []string{"person.permanentAddress"},
					},
				},
			}

			record, err := engine.CreateConsent(req)
			if err != nil {
				t.Fatalf("CreateConsent failed: %v", err)
			}

			// Approve the consent
			updateReq := UpdateConsentRequest{
				Status:    StatusApproved,
				UpdatedBy: "citizen_" + string(rune('0'+i)),
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
			consentIDs[i] = record.ConsentID
		}

		// Check expiry - should return all 3 expired records
		expiredRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(expiredRecords) != 3 {
			t.Errorf("Expected 3 expired records, got %d", len(expiredRecords))
		}

		// Verify all records are marked as expired
		for _, record := range expiredRecords {
			if record.Status != string(StatusExpired) {
				t.Errorf("Expected expired record status %s, got %s", string(StatusExpired), record.Status)
			}
		}
	})
}

// TestConsentEngine_UpdateConsentWithGrantDuration tests updating consent with grant_duration
func TestConsentEngine_UpdateConsentWithGrantDuration(t *testing.T) {
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalURL)

	// Create a consent
	req := ConsentRequest{
		AppID:     "passport-app",
		Purpose:   "passport_application",
		SessionID: "session_123",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "user123",
				OwnerEmail: "user123@example.com",
				Fields:     []string{"person.permanentAddress"},
			},
		},
	}

	record, err := engine.CreateConsent(req)
	if err != nil {
		t.Fatalf("CreateConsent failed: %v", err)
	}

	// Update consent with grant_duration
	updateReq := UpdateConsentRequest{
		Status:        StatusApproved,
		GrantDuration: "1m",
		UpdatedBy:     "citizen_123",
		Reason:        "User approved with custom duration",
	}

	updatedRecord, err := engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("UpdateConsent failed: %v", err)
	}

	// Verify grant_duration was updated
	if updatedRecord.GrantDuration != "1m" {
		t.Errorf("Expected grant_duration '1m', got %s", updatedRecord.GrantDuration)
	}

	// Verify expires_at was recalculated (should be approximately 1 minute from now)
	expectedExpiry := time.Now().Add(1 * time.Minute)
	timeDiff := updatedRecord.ExpiresAt.Sub(expectedExpiry)
	if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
		t.Errorf("Expected expires_at to be approximately 1 minute from now, got %v", updatedRecord.ExpiresAt)
	}

	// Verify status was updated
	if updatedRecord.Status != string(StatusApproved) {
		t.Errorf("Expected status %s, got %s", string(StatusApproved), updatedRecord.Status)
	}
}
