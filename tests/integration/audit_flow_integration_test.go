package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestCompleteAuditFlow tests the complete audit flow against PostgreSQL database
// This test simulates a full orchestration request that generates multiple audit events
// all linked by the same traceID
func TestCompleteAuditFlow(t *testing.T) {
	// Skip if audit service is not available
	if !isAuditServiceAvailable() {
		t.Skip("Audit service is not available. Skipping complete audit flow test.")
	}

	// Setup PostgreSQL test database connection
	db := setupAuditTestDB(t)
	if db == nil {
		t.Skip("Could not connect to PostgreSQL test database")
	}

	// Generate a unique traceID for this test (must be valid UUID format)
	// Format: 8-4-4-4-12 hex digits
	testTraceID := fmt.Sprintf("550e8400-e29b-41d4-a716-%012x", time.Now().UnixNano()%0xffffffffffff)
	t.Logf("Test traceID: %s", testTraceID)

	// Simulate the complete audit flow:
	// 1. ORCHESTRATION_REQUEST_RECEIVED
	// 2. POLICY_CHECK
	// 3. CONSENT_CHECK
	// 4. PROVIDER_FETCH (multiple providers)

	// Step 1: Log ORCHESTRATION_REQUEST_RECEIVED
	t.Run("ORCHESTRATION_REQUEST_RECEIVED", func(t *testing.T) {
		req := createAuditLogRequest(t, AuditLogRequest{
			TraceID:    testTraceID,
			EventType:  "ORCHESTRATION_REQUEST_RECEIVED",
			Status:     "SUCCESS",
			ActorType:  "SERVICE",
			ActorID:    "orchestration-engine",
			TargetType: "SERVICE",
			RequestMetadata: map[string]interface{}{
				"applicationId": "test-app-123",
				"query":         "query { citizen(nic: \"123456789V\") { name } }",
			},
		})
		resp, err := testHTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Should create audit log successfully")
	})

	// Small delay to ensure different timestamps
	time.Sleep(100 * time.Millisecond)

	// Step 2: Log POLICY_CHECK
	t.Run("POLICY_CHECK", func(t *testing.T) {
		req := createAuditLogRequest(t, AuditLogRequest{
			TraceID:    testTraceID,
			EventType:  "POLICY_CHECK",
			Status:     "SUCCESS",
			ActorType:  "SERVICE",
			ActorID:    "orchestration-engine",
			TargetType: "SERVICE",
			TargetID:   stringPtr("policy-decision-point"),
			ResponseMetadata: map[string]interface{}{
				"applicationId":         "test-app-123",
				"authorized":            true,
				"consentRequired":       true,
				"accessExpired":         false,
				"consentRequiredFields": []string{"citizen.name"},
			},
		})
		resp, err := testHTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Should create audit log successfully")
	})

	time.Sleep(100 * time.Millisecond)

	// Step 3: Log CONSENT_CHECK
	t.Run("CONSENT_CHECK", func(t *testing.T) {
		req := createAuditLogRequest(t, AuditLogRequest{
			TraceID:    testTraceID,
			EventType:  "CONSENT_CHECK",
			Status:     "SUCCESS",
			ActorType:  "SERVICE",
			ActorID:    "orchestration-engine",
			TargetType: "SERVICE",
			TargetID:   stringPtr("consent-engine"),
			ResponseMetadata: map[string]interface{}{
				"applicationId": "test-app-123",
				"ownerEmail":    "test@example.com",
				"ownerId":       "test-owner-123",
				"consentId":     "consent-123",
				"status":        "APPROVED",
				"fieldsCount":   1,
			},
		})
		resp, err := testHTTPClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Should create audit log successfully")
	})

	time.Sleep(100 * time.Millisecond)

	// Step 4: Log PROVIDER_FETCH (multiple providers)
	providers := []string{"provider-1", "provider-2"}
	for _, provider := range providers {
		t.Run(fmt.Sprintf("PROVIDER_FETCH_%s", provider), func(t *testing.T) {
			req := createAuditLogRequest(t, AuditLogRequest{
				TraceID:    testTraceID,
				EventType:  "PROVIDER_FETCH",
				Status:     "SUCCESS",
				ActorType:  "SERVICE",
				ActorID:    "orchestration-engine",
				TargetType: "SERVICE",
				TargetID:   stringPtr(provider),
				ResponseMetadata: map[string]interface{}{
					"applicationId":   "test-app-123",
					"schemaId":        "test-schema-123",
					"serviceKey":      provider,
					"requestedFields": []string{"citizen.name"},
					"query":           "query { citizen(nic: \"123456789V\") { name } }",
					"hasErrors":       false,
					"dataKeys":        []string{"citizen"},
				},
			})
			resp, err := testHTTPClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusCreated, resp.StatusCode, "Should create audit log successfully")
		})
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a moment for async operations to complete
	time.Sleep(500 * time.Millisecond)

	// Verify all events are in the database with the same traceID
	t.Run("VerifyTraceIDCorrelation", func(t *testing.T) {
		// Query database directly to verify all events
		var logs []AuditLogDB
		result := db.Where("trace_id = ?", testTraceID).Order("created_at ASC").Find(&logs)
		require.NoError(t, result.Error, "Should query audit logs from database")
		require.Greater(t, result.RowsAffected, int64(0), "Should find at least one audit log")

		// Expected event types in order
		expectedEvents := []string{
			"ORCHESTRATION_REQUEST_RECEIVED",
			"POLICY_CHECK",
			"CONSENT_CHECK",
			"PROVIDER_FETCH",
			"PROVIDER_FETCH",
		}

		assert.Equal(t, len(expectedEvents), len(logs), "Should have exactly %d events", len(expectedEvents))

		// Verify all events have the same traceID
		for i, log := range logs {
			assert.NotNil(t, log.TraceID, "Log %d should have traceID", i)
			assert.Equal(t, testTraceID, *log.TraceID, "All logs should share the same traceID")
			assert.Equal(t, expectedEvents[i], *log.EventType, "Event %d should be %s", i, expectedEvents[i])
			assert.Equal(t, "SUCCESS", log.Status, "Event %d should have SUCCESS status", i)
			assert.Equal(t, "SERVICE", log.ActorType, "Event %d should have SERVICE actor type", i)
			assert.Equal(t, "orchestration-engine", log.ActorID, "Event %d should have orchestration-engine actor", i)
		}

		t.Logf("Verified %d audit events all linked by traceID: %s", len(logs), testTraceID)
	})

	// Verify via API endpoint
	t.Run("VerifyViaAPI", func(t *testing.T) {
		auditLogs, err := getAuditLogsByTraceIDWithRetry(t, testTraceID, 5*time.Second, 200*time.Millisecond)
		require.NoError(t, err)
		require.Greater(t, len(auditLogs), 0, "Should find audit logs via API")

		// Verify all events share the same traceID
		for _, log := range auditLogs {
			require.NotNil(t, log.TraceID, "Audit log should have traceID")
			assert.Equal(t, testTraceID, *log.TraceID, "All audit logs should share the same traceID")
		}

		// Count event types
		eventTypeCount := make(map[string]int)
		for _, log := range auditLogs {
			if log.EventType != nil {
				eventTypeCount[*log.EventType]++
			}
		}

		assert.GreaterOrEqual(t, eventTypeCount["ORCHESTRATION_REQUEST_RECEIVED"], 1, "Should have ORCHESTRATION_REQUEST_RECEIVED")
		assert.GreaterOrEqual(t, eventTypeCount["POLICY_CHECK"], 1, "Should have POLICY_CHECK")
		assert.GreaterOrEqual(t, eventTypeCount["CONSENT_CHECK"], 1, "Should have CONSENT_CHECK")
		assert.GreaterOrEqual(t, eventTypeCount["PROVIDER_FETCH"], 2, "Should have at least 2 PROVIDER_FETCH events")

		t.Logf("Verified %d audit events via API, all linked by traceID: %s", len(auditLogs), testTraceID)
	})

	// Cleanup: Remove test data
	t.Cleanup(func() {
		if db != nil {
			db.Where("trace_id = ?", testTraceID).Delete(&AuditLogDB{})
			t.Logf("Cleaned up test audit logs for traceID: %s", testTraceID)
		}
	})
}

// AuditLogRequest represents the request payload for creating an audit log
type AuditLogRequest struct {
	TraceID          string
	EventType        string
	Status           string
	ActorType        string
	ActorID          string
	TargetType       string
	TargetID         *string
	RequestMetadata  map[string]interface{}
	ResponseMetadata map[string]interface{}
}

// AuditLogDB represents the audit log structure in the database
type AuditLogDB struct {
	ID               string  `gorm:"type:uuid;primary_key"`
	TraceID          *string `gorm:"type:uuid"`
	Timestamp        time.Time
	EventType        *string
	EventAction      *string
	Status           string
	ActorType        string
	ActorID          string
	TargetType       string
	TargetID         *string
	RequestMetadata  *string `gorm:"type:jsonb"`
	ResponseMetadata *string `gorm:"type:jsonb"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (AuditLogDB) TableName() string {
	return "audit_logs"
}

// setupAuditTestDB creates a PostgreSQL test database connection for audit service
func setupAuditTestDB(t *testing.T) *gorm.DB {
	// Use environment variables that match docker-compose.test.yml
	host := getEnvOrDefault("TEST_AUDIT_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_AUDIT_DB_PORT", "5435") // Matches docker-compose.test.yml port mapping
	user := getEnvOrDefault("TEST_AUDIT_DB_USERNAME", "postgres")
	password := getEnvOrDefault("TEST_AUDIT_DB_PASSWORD", "password")
	database := getEnvOrDefault("TEST_AUDIT_DB_DATABASE", "audit_db")
	sslmode := getEnvOrDefault("TEST_AUDIT_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("Skipping audit flow test: could not connect to test database: %v", err)
		return nil
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		t.Skipf("Skipping audit flow test: failed to get sql.DB: %v", err)
		return nil
	}

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping audit flow test: failed to ping database: %v", err)
		return nil
	}

	t.Logf("Connected to PostgreSQL audit test database: %s@%s:%s/%s", user, host, port, database)

	return db
}

// createAuditLogRequest creates an HTTP request to create an audit log
func createAuditLogRequest(t *testing.T, req AuditLogRequest) *http.Request {
	payload := map[string]interface{}{
		"traceId":    req.TraceID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"status":     req.Status,
		"actorType":  req.ActorType,
		"actorId":    req.ActorID,
		"targetType": req.TargetType,
	}

	if req.EventType != "" {
		payload["eventType"] = req.EventType
	}
	if req.TargetID != nil {
		payload["targetId"] = *req.TargetID
	}
	if req.RequestMetadata != nil {
		payload["requestMetadata"] = req.RequestMetadata
	}
	if req.ResponseMetadata != nil {
		payload["responseMetadata"] = req.ResponseMetadata
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, "Should marshal audit log request")

	httpReq, err := http.NewRequest("POST", auditServiceURL+"/api/audit-logs", bytes.NewBuffer(jsonData))
	require.NoError(t, err, "Should create HTTP request")
	httpReq.Header.Set("Content-Type", "application/json")

	return httpReq
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
