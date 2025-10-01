package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/stretchr/testify/assert"
)

func TestSchemaDatabaseNewFields(t *testing.T) {
	// This test verifies that the new database fields work correctly
	// Note: This is a unit test that doesn't require a real database connection

	t.Run("UnifiedSchema model has new fields", func(t *testing.T) {
		schema := &models.UnifiedSchema{
			ID:                 "test-id",
			Version:            "1.0.0",
			SDL:                "type Query { hello: String }",
			Status:             "active",
			Description:        "Test schema",
			CompatibilityLevel: "major",
			PreviousVersion:    stringPtr("0.9.0"),
			Metadata: map[string]interface{}{
				"author": "test-user",
				"tags":   []string{"test", "example"},
			},
			IsActive:   true,
			SchemaType: "current",
		}

		// Verify all new fields are present
		assert.Equal(t, "major", schema.CompatibilityLevel)
		assert.Equal(t, "0.9.0", *schema.PreviousVersion)
		assert.Equal(t, "test-user", schema.Metadata["author"])
		assert.True(t, schema.IsActive)
		assert.Equal(t, "current", schema.SchemaType)
	})

	t.Run("SchemaVersion model structure", func(t *testing.T) {
		version := &models.SchemaVersion{
			ID:          1,
			FromVersion: "1.0.0",
			ToVersion:   "1.1.0",
			ChangeType:  "minor",
			Changes: map[string]interface{}{
				"description":     "Added new field",
				"fields_added":    []string{"newField"},
				"fields_removed":  []string{},
				"fields_modified": []string{},
			},
			CreatedBy: "test-user",
		}

		// Verify schema version structure
		assert.Equal(t, 1, version.ID)
		assert.Equal(t, "1.0.0", version.FromVersion)
		assert.Equal(t, "1.1.0", version.ToVersion)
		assert.Equal(t, "minor", version.ChangeType)
		assert.Equal(t, "Added new field", version.Changes["description"])
		assert.Equal(t, "test-user", version.CreatedBy)
	})

	t.Run("Database table creation methods exist", func(t *testing.T) {
		// Verify that the database methods exist
		var schemaDB *database.SchemaDB

		// These methods should exist (we're not calling them, just checking they exist)
		_ = schemaDB.CreateSchemaTable
		_ = schemaDB.CreateSchemaVersionsTable
		_ = schemaDB.CreateSchemaVersion
		_ = schemaDB.GetSchemaVersionsByVersion
		_ = schemaDB.GetAllSchemaVersions

		// If we get here without compilation errors, the methods exist
		assert.True(t, true, "All required database methods exist")
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
