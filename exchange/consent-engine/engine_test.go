package main

import (
	"testing"
	"time"
)

// TestConsentEngine_CreateConsent tests the core CreateConsent functionality
func TestConsentEngine_CreateConsent(t *testing.T) {
	TestWithBothEngines(t, "CreateConsent", func(t *testing.T, engine ConsentEngine) {

		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_123",
			DataFields: []DataField{
				{
					
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
	})
}

// TestConsentEngine_GetConsentStatus tests retrieving consent status
func TestConsentEngine_GetConsentStatus(t *testing.T) {
	TestWithBothEngines(t, "GetConsentStatus", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_123",
			DataFields: []DataField{
				{
					
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
	})
}

// TestConsentEngine_UpdateConsent tests updating consent status
func TestConsentEngine_UpdateConsent(t *testing.T) {
	TestWithBothEngines(t, "UpdateConsent", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_123",
			DataFields: []DataField{
				{
					
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
	})
}

// TestConsentEngine_FindExistingConsent tests finding existing consents
func TestConsentEngine_FindExistingConsent(t *testing.T) {
	TestWithBothEngines(t, "FindExistingConsent", func(t *testing.T, engine ConsentEngine) {

		// Create a consent record first
		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_123",
			DataFields: []DataField{
				{
					
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
	})
}

// TestConsentEngine_StatusTransitions tests status transition validation
func TestConsentEngine_StatusTransitions(t *testing.T) {
	// Test valid transitions
	validTransitions := []struct {
		from  ConsentStatus
		to    ConsentStatus
		valid bool
	}{
		// Pending state transitions
		{StatusPending, StatusApproved, true},
		{StatusPending, StatusRejected, true},
		{StatusPending, StatusExpired, true},

		// Approved state transitions
		{StatusApproved, StatusApproved, true}, // Direct approval
		{StatusApproved, StatusRejected, true}, // Direct rejection
		{StatusApproved, StatusRevoked, true},  // User revocation
		{StatusApproved, StatusExpired, true},  // Expiry

		// Terminal states - can only transition to expired or stay the same
		{StatusRejected, StatusExpired, true}, // Rejected -> Expired
		{StatusExpired, StatusExpired, true},  // Expired -> Expired
		{StatusRevoked, StatusExpired, true},  // Revoked -> Expired

		// Invalid transitions from terminal states
		{StatusRejected, StatusPending, false},  // Terminal state cannot go back to pending
		{StatusRejected, StatusApproved, false}, // Terminal state cannot go back to approved
		{StatusExpired, StatusPending, false},   // Terminal state cannot go back to pending
		{StatusExpired, StatusApproved, false},  // Terminal state cannot go back to approved
		{StatusRevoked, StatusPending, false},   // Terminal state cannot go back to pending
		{StatusRevoked, StatusApproved, false},  // Terminal state cannot go back to approved

		// Invalid transitions from approved
		{StatusApproved, StatusPending, false}, // Cannot go back to pending
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
			
			SessionID: "session_123",
			DataFields: []DataField{
				{
					
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

		// Check expiry - should return empty list (no records to delete)
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 0 {
			t.Errorf("Expected 0 deleted records, got %d", len(deletedRecords))
		}
	})

	t.Run("HasExpiredRecords", func(t *testing.T) {
		// Create a new engine to avoid interference
		consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
		engine := NewConsentEngine(consentPortalURL)

		// Create a consent
		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_456",
			DataFields: []DataField{
				{
					
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

		// Check expiry - should delete the expired record
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 1 {
			t.Errorf("Expected 1 deleted record, got %d", len(deletedRecords))
		}

		if deletedRecords[0].ConsentID != record.ConsentID {
			t.Errorf("Expected deleted record ID %s, got %s", record.ConsentID, deletedRecords[0].ConsentID)
		}

		// Verify the record was actually deleted from the store
		_, err = engine.GetConsentStatus(record.ConsentID)
		if err == nil {
			t.Errorf("Expected record to be deleted, but it still exists")
		}
	})

	t.Run("OnlyApprovedRecordsExpire", func(t *testing.T) {
		// Create a new engine to avoid interference
		consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
		engine := NewConsentEngine(consentPortalURL)

		// Create a consent
		req := ConsentRequest{
			AppID:     "passport-app",
			
			SessionID: "session_789",
			DataFields: []DataField{
				{
					
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

		// Check expiry - should return empty list (rejected records don't get deleted)
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 0 {
			t.Errorf("Expected 0 deleted records (rejected records don't get deleted), got %d", len(deletedRecords))
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
				
				SessionID: "session_" + string(rune('0'+i)),
				DataFields: []DataField{
					{
						
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

		// Check expiry - should delete all 3 expired records
		deletedRecords, err := engine.CheckConsentExpiry()
		if err != nil {
			t.Fatalf("CheckConsentExpiry failed: %v", err)
		}

		if len(deletedRecords) != 3 {
			t.Errorf("Expected 3 deleted records, got %d", len(deletedRecords))
		}

		// Verify all records were actually deleted from the store
		for _, consentID := range consentIDs {
			_, err = engine.GetConsentStatus(consentID)
			if err == nil {
				t.Errorf("Expected record %s to be deleted, but it still exists", consentID)
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
		
		SessionID: "session_123",
		DataFields: []DataField{
			{
				
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

// TestRejectedConsentReuseIssue tests the issue where rejected consents create new records instead of reusing
func TestRejectedConsentReuseIssue(t *testing.T) {
	// Create a new consent engine
	engine := NewConsentEngine("http://localhost:5173")

	// Create initial consent request
	req1 := ConsentRequest{
		AppID:         "test-app",
		
		SessionID:     "session123",
		GrantDuration: "1h",
		DataFields: []DataField{
			{
				
				OwnerID:    "user123",
				OwnerEmail: "user@example.com",
				Fields:     []string{"name", "email"}, // 2 fields
			},
		},
	}

	// Create the first consent record
	record1, err := engine.ProcessConsentRequest(req1)
	if err != nil {
		t.Fatalf("Failed to create first consent: %v", err)
	}

	t.Logf("First consent created: ID=%s, Status=%s", record1.ConsentID, record1.Status)

	// Reject the consent
	updateReq := UpdateConsentRequest{
		Status:    StatusRejected,
		UpdatedBy: "test-user",
		Reason:    "testing rejection",
	}

	rejectedRecord, err := engine.UpdateConsent(record1.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("Failed to reject consent: %v", err)
	}

	if rejectedRecord.Status != string(StatusRejected) {
		t.Errorf("Expected status to be rejected, got %s", rejectedRecord.Status)
	}

	t.Logf("Consent rejected: ID=%s, Status=%s", rejectedRecord.ConsentID, rejectedRecord.Status)

	// Now send the same request but with different number of fields
	req2 := ConsentRequest{
		AppID:         "test-app",
		
		SessionID:     "session456", // Different session
		GrantDuration: "1h",
		DataFields: []DataField{
			{
				
				OwnerID:    "user123",
				OwnerEmail: "user@example.com",
				Fields:     []string{"name", "email", "phone"}, // 3 fields instead of 2
			},
		},
	}

	// This should reuse the existing rejected record, not create a new one
	record2, err := engine.ProcessConsentRequest(req2)
	if err != nil {
		t.Fatalf("Failed to process second consent request: %v", err)
	}

	t.Logf("Second consent processed: ID=%s, Status=%s", record2.ConsentID, record2.Status)

	// CORRECT BEHAVIOR: When a consent is rejected, we should create a NEW record, not reuse the old one
	if record2.ConsentID == record1.ConsentID {
		t.Errorf("Expected to create NEW consent ID, but reused existing ID %s", record1.ConsentID)
		t.Logf("ISSUE: Reused existing rejected consent record when it should create a new one")
	} else {
		t.Logf("SUCCESS: Created new consent record as expected for rejected consent")
	}

	// Verify the record is now pending
	if record2.Status != string(StatusPending) {
		t.Errorf("Expected status to be pending, got %s", record2.Status)
	}

	// Verify the fields were updated
	expectedFields := []string{"name", "email", "phone"}
	if len(record2.Fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(record2.Fields))
	}

	// Stop the background process
	engine.StopBackgroundExpiryProcess()
}

// TestConsentReuseLogic tests the correct behavior for consent record reuse based on status
func TestConsentReuseLogic(t *testing.T) {
	// Create a new consent engine
	engine := NewConsentEngine("http://localhost:5173")

	baseReq := ConsentRequest{
		AppID:         "test-app",
		
		SessionID:     "session123",
		GrantDuration: "1h",
		DataFields: []DataField{
			{
				
				OwnerID:    "user123",
				OwnerEmail: "user@example.com",
				Fields:     []string{"name", "email"},
			},
		},
	}

	// Test 1: Pending records should be reused
	t.Log("=== Test 1: Pending records should be reused ===")
	record1, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to create first consent: %v", err)
	}
	t.Logf("First consent: ID=%s, Status=%s", record1.ConsentID, record1.Status)

	// Send same request - should reuse pending record
	baseReq.SessionID = "session456" // Different session
	record2, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process second request: %v", err)
	}
	t.Logf("Second request: ID=%s, Status=%s", record2.ConsentID, record2.Status)

	if record2.ConsentID != record1.ConsentID {
		t.Errorf("Expected to reuse pending consent ID %s, but got new ID %s", record1.ConsentID, record2.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused pending consent record")
	}

	// Test 2: Rejected records should NOT be reused
	t.Log("\n=== Test 2: Rejected records should NOT be reused ===")
	updateReq := UpdateConsentRequest{
		Status:    StatusRejected,
		UpdatedBy: "test-user",
		Reason:    "testing rejection",
	}

	rejectedRecord, err := engine.UpdateConsent(record1.ConsentID, updateReq)
	if err != nil {
		t.Fatalf("Failed to reject consent: %v", err)
	}
	t.Logf("Consent rejected: ID=%s, Status=%s", rejectedRecord.ConsentID, rejectedRecord.Status)

	// Send same request after rejection - should create NEW record
	baseReq.SessionID = "session789"
	record3, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process third request: %v", err)
	}
	t.Logf("Third request: ID=%s, Status=%s", record3.ConsentID, record3.Status)

	if record3.ConsentID == record1.ConsentID {
		t.Errorf("Expected to create NEW consent ID, but reused rejected ID %s", record1.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Created new consent record after rejection")
	}

	// Test 3: Approved records should NOT be reused
	t.Log("\n=== Test 3: Approved records should NOT be reused ===")
	approveReq := UpdateConsentRequest{
		Status:    StatusApproved,
		UpdatedBy: "test-user",
		Reason:    "testing approval",
	}

	approvedRecord, err := engine.UpdateConsent(record3.ConsentID, approveReq)
	if err != nil {
		t.Fatalf("Failed to approve consent: %v", err)
	}
	t.Logf("Consent approved: ID=%s, Status=%s", approvedRecord.ConsentID, approvedRecord.Status)

	// Send same request after approval - should create NEW record
	baseReq.SessionID = "session999"
	record4, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process fourth request: %v", err)
	}
	t.Logf("Fourth request: ID=%s, Status=%s", record4.ConsentID, record4.Status)

	if record4.ConsentID == record3.ConsentID {
		t.Errorf("Expected to create NEW consent ID, but reused approved ID %s", record3.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Created new consent record after approval")
	}

	// Test 4: Only ONE pending record per (appId, ownerId, ownerEmail) tuple
	t.Log("\n=== Test 4: Only ONE pending record per tuple ===")
	// Send another request - should reuse the pending record from Test 3
	baseReq.SessionID = "session1000"
	record5, err := engine.ProcessConsentRequest(baseReq)
	if err != nil {
		t.Fatalf("Failed to process fifth request: %v", err)
	}
	t.Logf("Fifth request: ID=%s, Status=%s", record5.ConsentID, record5.Status)

	if record5.ConsentID != record4.ConsentID {
		t.Errorf("Expected to reuse pending consent ID %s, but got new ID %s", record4.ConsentID, record5.ConsentID)
	} else {
		t.Logf("✓ SUCCESS: Reused pending consent record (only one pending per tuple)")
	}

	// Stop the background process
	engine.StopBackgroundExpiryProcess()
}
