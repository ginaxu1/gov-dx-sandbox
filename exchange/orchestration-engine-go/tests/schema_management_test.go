package tests

import (
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/stretchr/testify/assert"
)

// TestUnifiedSchemaModel tests the UnifiedSchema model structure
func TestUnifiedSchemaModel(t *testing.T) {
	t.Run("UnifiedSchema creation", func(t *testing.T) {
		schema := &models.UnifiedSchema{
			ID:                 "test-id",
			Version:            "1.0.0",
			SDL:                "type Query { hello: String }",
			Status:             "active",
			Description:        "Test schema",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
			CreatedBy:          "test-user",
			Checksum:           "abc123",
			CompatibilityLevel: "major",
			PreviousVersion:    stringPtr("0.9.0"),
			Metadata: map[string]interface{}{
				"author": "test-user",
				"tags":   []string{"test", "example"},
			},
			IsActive:   true,
			SchemaType: "current",
		}

		// Verify all fields are set correctly
		assert.Equal(t, "test-id", schema.ID)
		assert.Equal(t, "1.0.0", schema.Version)
		assert.Equal(t, "type Query { hello: String }", schema.SDL)
		assert.Equal(t, "active", schema.Status)
		assert.Equal(t, "major", schema.CompatibilityLevel)
		assert.Equal(t, "0.9.0", *schema.PreviousVersion)
		assert.True(t, schema.IsActive)
		assert.Equal(t, "current", schema.SchemaType)
		assert.Equal(t, "test-user", schema.Metadata["author"])
	})

	t.Run("SchemaVersion model", func(t *testing.T) {
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
			CreatedAt: time.Now(),
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

	t.Run("GraphQLRequest model", func(t *testing.T) {
		req := &models.GraphQLRequest{
			Query:         "query { hello }",
			Variables:     map[string]interface{}{"name": "world"},
			OperationName: "HelloQuery",
			SchemaVersion: "1.0.0",
		}

		assert.Equal(t, "query { hello }", req.Query)
		assert.Equal(t, "world", req.Variables["name"])
		assert.Equal(t, "HelloQuery", req.OperationName)
		assert.Equal(t, "1.0.0", req.SchemaVersion)
	})

	t.Run("ValidationError model", func(t *testing.T) {
		err := &models.ValidationError{
			Field:   "query",
			Message: "query is required",
		}

		assert.Equal(t, "query", err.Field)
		assert.Equal(t, "query is required", err.Message)
		assert.Equal(t, "query is required", err.Error())
	})
}

// TestSchemaCompatibilityCheck tests the compatibility check model
func TestSchemaCompatibilityCheck(t *testing.T) {
	t.Run("Compatible schema", func(t *testing.T) {
		check := &models.SchemaCompatibilityCheck{
			Compatible:         true,
			BreakingChanges:    []string{},
			Warnings:           []string{"New field added"},
			CompatibilityLevel: "minor",
		}

		assert.True(t, check.Compatible)
		assert.Empty(t, check.BreakingChanges)
		assert.Len(t, check.Warnings, 1)
		assert.Equal(t, "minor", check.CompatibilityLevel)
	})

	t.Run("Incompatible schema", func(t *testing.T) {
		check := &models.SchemaCompatibilityCheck{
			Compatible:         false,
			BreakingChanges:    []string{"Field 'name' was removed"},
			Warnings:           []string{},
			CompatibilityLevel: "major",
		}

		assert.False(t, check.Compatible)
		assert.Len(t, check.BreakingChanges, 1)
		assert.Empty(t, check.Warnings)
		assert.Equal(t, "major", check.CompatibilityLevel)
	})
}
