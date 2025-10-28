package integration

import (
	"context"
	"encoding/json"
	"testing"
)

// TestDataExchangeFlow tests the happy path of data exchange
func TestDataExchangeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("provider metadata query", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for provider metadata
		rows, err := testDB.Query(ctx,
			"SELECT field_name, owner, provider, access_control_type FROM provider_metadata LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query provider metadata: %v", err)
		}
		defer rows.Close()

		type FieldInfo struct {
			FieldName         string
			Owner             string
			Provider          string
			AccessControlType string
		}

		fields := make([]FieldInfo, 0)
		for rows.Next() {
			var field FieldInfo
			if err := rows.Scan(&field.FieldName, &field.Owner, &field.Provider, &field.AccessControlType); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			t.Error("Expected to find provider metadata")
		} else {
			t.Logf("Found %d provider metadata entries", len(fields))
			for _, field := range fields {
				t.Logf("Field: %s, Owner: %s, Provider: %s, Type: %s",
					field.FieldName, field.Owner, field.Provider, field.AccessControlType)
			}
		}
	})

	t.Run("consumer grants query", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for consumer grants
		rows, err := testDB.Query(ctx,
			"SELECT consumer_id, approved_fields FROM consumer_grants LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query consumer grants: %v", err)
		}
		defer rows.Close()

		type GrantInfo struct {
			ConsumerID     string
			ApprovedFields string
		}

		grants := make([]GrantInfo, 0)
		for rows.Next() {
			var grant GrantInfo
			if err := rows.Scan(&grant.ConsumerID, &grant.ApprovedFields); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}

			// Parse JSON
			var fields []string
			if err := json.Unmarshal([]byte(grant.ApprovedFields), &fields); err != nil {
				t.Logf("Warning: could not parse approved_fields: %v", err)
				fields = []string{grant.ApprovedFields}
			}

			grants = append(grants, grant)
			t.Logf("Consumer: %s, Fields: %v", grant.ConsumerID, fields)
		}

		if len(grants) == 0 {
			t.Error("Expected to find consumer grants")
		} else {
			t.Logf("Found %d consumer grants", len(grants))
		}
	})

	t.Run("provider schemas query", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for provider schemas
		rows, err := testDB.Query(ctx,
			"SELECT submission_id, provider_id, schema_id, status, schema_endpoint FROM provider_schemas LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query provider schemas: %v", err)
		}
		defer rows.Close()

		type SchemaInfo struct {
			SubmissionID   string
			ProviderID     string
			SchemaID       string
			Status         string
			SchemaEndpoint string
		}

		schemas := make([]SchemaInfo, 0)
		for rows.Next() {
			var schema SchemaInfo
			if err := rows.Scan(&schema.SubmissionID, &schema.ProviderID, &schema.SchemaID, &schema.Status, &schema.SchemaEndpoint); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			schemas = append(schemas, schema)
		}

		if len(schemas) == 0 {
			t.Error("Expected to find provider schemas")
		} else {
			t.Logf("Found %d provider schemas", len(schemas))
			for _, schema := range schemas {
				t.Logf("Schema: %s, Provider: %s, Status: %s", schema.SchemaID, schema.ProviderID, schema.Status)
			}
		}
	})
}

// TestConsentWorkflow tests consent management
func TestConsentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("consent table structure", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		// Query for consents
		rows, err := testDB.Query(ctx,
			"SELECT consent_id, consumer_id, provider_id, status FROM consents LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query consents: %v", err)
		}
		defer rows.Close()

		type ConsentInfo struct {
			ConsentID  string
			ConsumerID string
			ProviderID string
			Status     string
		}

		consents := make([]ConsentInfo, 0)
		for rows.Next() {
			var consent ConsentInfo
			if err := rows.Scan(&consent.ConsentID, &consent.ConsumerID, &consent.ProviderID, &consent.Status); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			consents = append(consents, consent)
		}

		t.Logf("Found %d consent records", len(consents))
		for _, consent := range consents {
			t.Logf("Consent: %s, Consumer: %s, Provider: %s, Status: %s",
				consent.ConsentID, consent.ConsumerID, consent.ProviderID, consent.Status)
		}
	})

	t.Run("check pending consents", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		var count int
		err := testDB.QueryRow(ctx, "SELECT COUNT(*) FROM consents WHERE status = 'pending'").
			Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query pending consents: %v", err)
		}

		t.Logf("Found %d pending consents", count)
	})

	t.Run("check approved consents", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		var count int
		err := testDB.QueryRow(ctx, "SELECT COUNT(*) FROM consents WHERE status = 'approved'").
			Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query approved consents: %v", err)
		}

		t.Logf("Found %d approved consents", count)
	})
}

// TestProviderDataRetrieval tests that services can retrieve data from the database
func TestProviderDataRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("provider profiles", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		rows, err := testDB.Query(ctx,
			"SELECT provider_id, provider_name, provider_type FROM provider_profiles LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query provider profiles: %v", err)
		}
		defer rows.Close()

		type ProviderInfo struct {
			ProviderID   string
			ProviderName string
			ProviderType string
		}

		providers := make([]ProviderInfo, 0)
		for rows.Next() {
			var provider ProviderInfo
			if err := rows.Scan(&provider.ProviderID, &provider.ProviderName, &provider.ProviderType); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			providers = append(providers, provider)
		}

		if len(providers) == 0 {
			t.Error("Expected to find provider profiles")
		} else {
			t.Logf("Found %d provider profiles", len(providers))
			for _, provider := range providers {
				t.Logf("Provider: %s, Name: %s, Type: %s",
					provider.ProviderID, provider.ProviderName, provider.ProviderType)
			}
		}
	})

	t.Run("entity records", func(t *testing.T) {
		testDB := NewTestDB(getPostgresURL())
		if err := testDB.Connect(); err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer testDB.Close()

		rows, err := testDB.Query(ctx,
			"SELECT entity_id, entity_name, entity_type FROM entities LIMIT 10")
		if err != nil {
			t.Fatalf("Failed to query entities: %v", err)
		}
		defer rows.Close()

		type EntityInfo struct {
			EntityID   string
			EntityName string
			EntityType string
		}

		entities := make([]EntityInfo, 0)
		for rows.Next() {
			var entity EntityInfo
			if err := rows.Scan(&entity.EntityID, &entity.EntityName, &entity.EntityType); err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}
			entities = append(entities, entity)
		}

		if len(entities) == 0 {
			t.Error("Expected to find entities")
		} else {
			t.Logf("Found %d entities", len(entities))
			for _, entity := range entities {
				t.Logf("Entity: %s, Name: %s, Type: %s",
					entity.EntityID, entity.EntityName, entity.EntityType)
			}
		}
	})
}
