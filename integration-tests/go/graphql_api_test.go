package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// TestGraphQLSchema tests GraphQL schema availability
func TestGraphQLSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("unified schemas table", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for unified schemas
		rows, err := testDB.Query(ctx,
			"SELECT id, version, status, is_active, description FROM unified_schemas LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query unified schemas: %v", err)
		}
		defer rows.Close()

		type SchemaInfo struct {
			ID          string
			Version     string
			Status      string
			IsActive    bool
			Description string
		}

		schemas := make([]SchemaInfo, 0)
		for rows.Next() {
			var schema SchemaInfo
			if err := rows.Scan(&schema.ID, &schema.Version, &schema.Status, &schema.IsActive, &schema.Description); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			schemas = append(schemas, schema)
		}

		if len(schemas) == 0 {
			t.Error("Expected to find unified schemas")
		} else {
			t.Logf("Found %d unified schemas", len(schemas))
			for _, schema := range schemas {
				t.Logf("Schema: %s, Version: %s, Status: %s, Active: %v",
					schema.ID, schema.Version, schema.Status, schema.IsActive)
			}
		}
	})

	t.Run("check active schema", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for active schema
		var count int
		err := testDB.QueryRow(ctx, "SELECT COUNT(*) FROM unified_schemas WHERE is_active = true").
			Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query active schemas: %v", err)
		}

		if count == 0 {
			t.Error("Expected to find at least one active schema")
		} else {
			t.Logf("Found %d active schema(s)", count)
		}
	})
}

// TestGraphQLQueryExecution tests GraphQL query execution (if service is running)
func TestGraphQLQueryExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("simulate GraphQL query", func(t *testing.T) {
		// This would test actual GraphQL execution if the service were running

		graphqlQuery := map[string]interface{}{
			"query": `
				query {
					personInfo(nic: "123456789V") {
						fullName
						birthDate
					}
				}
			`,
		}

		// In a real test, this would be sent to the orchestration engine
		queryJSON, _ := json.Marshal(graphqlQuery)
		t.Logf("Simulated GraphQL query: %s", string(queryJSON))
	})
}

// TestGraphQLSchemaIntrospection tests schema introspection capabilities
func TestGraphQLSchemaIntrospection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("query schema SDL", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for schema SDL
		var sdl string
		err := testDB.QueryRow(ctx, "SELECT sdl FROM unified_schemas WHERE is_active = true LIMIT 1").
			Scan(&sdl)
		if err != nil {
			t.Fatalf("Failed to query schema SDL: %v", err)
		}

		if len(sdl) == 0 {
			t.Error("Expected to find schema SDL")
		} else {
			t.Logf("Schema SDL:\n%s", sdl)
		}
	})
}

// TestGraphQLFederation tests GraphQL federation capabilities
func TestGraphQLFederation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("schema version tracking", func(t *testing.T) {
		ctx := context.Background()
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Skipf("Skipping test - database not available: %v", err)
		}
		defer testDB.Close()

		// Query for schema versions
		rows, err := testDB.Query(ctx,
			"SELECT from_version, to_version, change_type FROM schema_versions LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query schema versions: %v", err)
		}
		defer rows.Close()

		type VersionInfo struct {
			FromVersion string
			ToVersion   string
			ChangeType  string
		}

		versions := make([]VersionInfo, 0)
		for rows.Next() {
			var version VersionInfo
			var fromVersion sql.NullString
			if err := rows.Scan(&fromVersion, &version.ToVersion, &version.ChangeType); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			version.FromVersion = fromVersion.String
			versions = append(versions, version)
		}

		t.Logf("Found %d schema version records", len(versions))
		for _, version := range versions {
			t.Logf("Version change: %s -> %s, Type: %s",
				version.FromVersion, version.ToVersion, version.ChangeType)
		}
	})
}

// Helper function to make HTTP requests (if needed)
func makeHTTPRequest(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}
