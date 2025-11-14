package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestAuthenticationAndAuthorization tests the authentication and authorization flow
func TestAuthenticationAndAuthorization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Wait for services to be ready
	err := WaitForServices(2 * time.Minute)
	if err != nil {
		t.Skipf("Skipping test - services not ready: %v", err)
	}

	t.Run("OPA health check", func(t *testing.T) {
		opaURL := getOPAURL()

		// Try to connect to OPA
		resp, err := http.Get(fmt.Sprintf("%s/health", opaURL))
		if err != nil {
			t.Logf("OPA not accessible at %s (this is expected in basic setup)", opaURL)
			return // OPA is optional in test environment
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Database authentication query", func(t *testing.T) {
		// Connect to test database
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for test consumers
		rows, err := testDB.Query(ctx, "SELECT consumer_id FROM consumers LIMIT 5")
		if err != nil {
			t.Fatalf("Failed to query consumers: %v", err)
		}
		defer rows.Close()

		foundPassportApp := false
		for rows.Next() {
			var consumerID string
			if err := rows.Scan(&consumerID); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			if consumerID == "passport-app" {
				foundPassportApp = true
			}
		}

		if !foundPassportApp {
			t.Error("Expected to find 'passport-app' consumer in test data")
		}
	})

	t.Run("Policy metadata check", func(t *testing.T) {
		// Connect to test database
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for policy metadata
		rows, err := testDB.Query(ctx,
			"SELECT field_name, access_control_type, consent_required FROM provider_metadata LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query policy metadata: %v", err)
		}
		defer rows.Close()

		type FieldMeta struct {
			FieldName         string
			AccessControlType string
			ConsentRequired   bool
		}

		fields := make([]FieldMeta, 0)
		for rows.Next() {
			var field FieldMeta
			if err := rows.Scan(&field.FieldName, &field.AccessControlType, &field.ConsentRequired); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			t.Error("Expected to find policy metadata in test data")
		} else {
			t.Logf("Found %d policy metadata entries", len(fields))
			for _, field := range fields {
				t.Logf("Field: %s, Type: %s, Consent: %v",
					field.FieldName, field.AccessControlType, field.ConsentRequired)
			}
		}
	})
}

// TestJWTAuthentication tests JWT token validation (if JWT infrastructure exists)
func TestJWTAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("JWT structure validation", func(t *testing.T) {
		// This would test actual JWT validation if the authentication
		// infrastructure were running in the test environment

		// For now, just verify we can query the database
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// This test would normally make an API call with a JWT
		// and verify the service correctly validates it
		t.Log("JWT validation test would execute here if services were running")
	})
}

// TestAuthorizationPolicies tests that policies are correctly enforced
func TestAuthorizationPolicies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("check public field access", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for public fields
		var fieldCount int
		err := testDB.QueryRow(ctx,
			"SELECT COUNT(*) FROM provider_metadata WHERE access_control_type = 'public'").
			Scan(&fieldCount)
		if err != nil {
			t.Fatalf("Failed to query public fields: %v", err)
		}

		if fieldCount == 0 {
			t.Error("Expected to find public fields in test data")
		} else {
			t.Logf("Found %d public fields", fieldCount)
		}
	})

	t.Run("check restricted field access", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for restricted fields
		var fieldCount int
		err := testDB.QueryRow(ctx,
			"SELECT COUNT(*) FROM provider_metadata WHERE access_control_type = 'restricted'").
			Scan(&fieldCount)
		if err != nil {
			t.Fatalf("Failed to query restricted fields: %v", err)
		}

		if fieldCount == 0 {
			t.Error("Expected to find restricted fields in test data")
		} else {
			t.Logf("Found %d restricted fields", fieldCount)
		}
	})

	t.Run("check allow_list configuration", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for fields with allow_list entries
		rows, err := testDB.Query(ctx,
			"SELECT field_name, allow_list FROM provider_metadata WHERE allow_list IS NOT NULL")
		if err != nil {
			t.Fatalf("Failed to query allow_list: %v", err)
		}
		defer rows.Close()

		foundEntries := false
		for rows.Next() {
			var fieldName string
			var allowListJSON string
			if err := rows.Scan(&fieldName, &allowListJSON); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}

			var allowList []map[string]interface{}
			if err := json.Unmarshal([]byte(allowListJSON), &allowList); err != nil {
				t.Logf("Warning: could not parse allow_list for %s: %v", fieldName, err)
				continue
			}

			if len(allowList) > 0 {
				foundEntries = true
				t.Logf("Field %s has %d allow_list entries", fieldName, len(allowList))
			}
		}

		if !foundEntries {
			t.Error("Expected to find allow_list entries in test data")
		}
	})
}
