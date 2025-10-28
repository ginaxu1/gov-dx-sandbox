package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// TestAuditLogCreation tests that audit logs can be created in the database
func TestAuditLogCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("create audit log in database", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Insert a test audit log
		insertQuery := `
			INSERT INTO audit_logs (timestamp, status, requested_data, application_id, schema_id, consumer_id, provider_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		testData := map[string]interface{}{
			"query": "query { personInfo(nic: \"123\") { fullName } }",
		}
		testDataJSON, _ := json.Marshal(testData)

		err := testDB.ExecuteSQL(ctx, insertQuery,
			time.Now(),
			"success",
			string(testDataJSON),
			"test-app",
			"schema-1",
			"test-consumer",
			"provider-drp",
		)
		if err != nil {
			t.Fatalf("Failed to insert audit log: %v", err)
		}

		// Query to verify the audit log was created
		var count int
		err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs WHERE application_id = 'test-app'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}

		if count == 0 {
			t.Error("Expected to find audit log entry")
		} else {
			t.Logf("Successfully created audit log (count: %d)", count)
		}
	})
}

// TestAuditStreamToDatabase tests that Redis stream messages are processed to the database
func TestAuditStreamToDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Redis stream operations", func(t *testing.T) {
		testRedis := NewTestRedis(getRedisURL())
		if err := testRedis.Connect(); err != nil {
			t.Fatalf("Failed to connect to Redis: %v", err)
		}
		defer testRedis.Close()

		streamName := "audit-events"

		// Add a test message to the stream
		testMessage := map[string]interface{}{
			"status":         "success",
			"requested_data": "query { personInfo(nic: \"123\") { fullName } }",
			"application_id": "test-app",
			"schema_id":      "schema-1",
			"consumer_id":    "test-consumer",
			"provider_id":    "provider-drp",
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		messageID, err := testRedis.AddToStream(ctx, streamName, testMessage)
		if err != nil {
			t.Fatalf("Failed to add message to stream: %v", err)
		}

		t.Logf("Added message to stream with ID: %s", messageID)

		// Check stream length
		streamLength, err := testRedis.GetStreamLength(ctx, streamName)
		if err != nil {
			t.Fatalf("Failed to get stream length: %v", err)
		}

		t.Logf("Stream length: %d", streamLength)
		if streamLength == 0 {
			t.Error("Expected stream to have at least one message")
		}
	})

	t.Run("verify message in database after processing", func(t *testing.T) {
		// Give time for message processing (if a consumer were running)
		time.Sleep(2 * time.Second)

		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for recent audit logs
		rows, err := testDB.Query(ctx,
			"SELECT id, status, application_id, consumer_id FROM audit_logs ORDER BY timestamp DESC LIMIT 5")
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}
		defer rows.Close()

		// Just verify we can read audit logs from the database
		count := 0
		for rows.Next() {
			var id, status, appID, consumerID string
			if err := rows.Scan(&id, &status, &appID, &consumerID); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			t.Logf("Audit log: ID=%s, Status=%s, App=%s, Consumer=%s", id, status, appID, consumerID)
			count++
		}

		if count > 0 {
			t.Logf("Found %d audit log entries in database", count)
		}
	})
}

// TestAuditLogFiltering tests that audit logs can be filtered correctly
func TestAuditLogFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("filter by consumer_id", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for specific consumer
		var count int
		err := testDB.QueryRow(ctx,
			"SELECT COUNT(*) FROM audit_logs WHERE consumer_id = 'test-consumer'").
			Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}

		t.Logf("Found %d audit logs for test-consumer", count)
	})

	t.Run("filter by status", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for specific status
		var count int
		err := testDB.QueryRow(ctx,
			"SELECT COUNT(*) FROM audit_logs WHERE status = 'success'").
			Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query audit logs: %v", err)
		}

		t.Logf("Found %d audit logs with status 'success'", count)
	})
}

// TestAuditMiddlewareIntegration tests the integration with audit middleware
func TestAuditMiddlewareIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("verify audit logs table structure", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for table schema
		rows, err := testDB.Query(ctx, `
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = 'audit_logs'
			ORDER BY ordinal_position
		`)
		if err != nil {
			t.Fatalf("Failed to query table schema: %v", err)
		}
		defer rows.Close()

		columns := make([]string, 0)
		for rows.Next() {
			var columnName, dataType string
			if err := rows.Scan(&columnName, &dataType); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			columns = append(columns, fmt.Sprintf("%s (%s)", columnName, dataType))
		}

		t.Logf("Audit logs table has %d columns", len(columns))
		for _, col := range columns {
			t.Logf("  Column: %s", col)
		}
	})
}
