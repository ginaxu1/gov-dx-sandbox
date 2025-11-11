package services

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBForSchema creates an in-memory SQLite database for schema testing
func setupTestDBForSchema(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&models.Schema{}, &models.SchemaSubmission{}, &models.Member{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestSchemaService_UpdateSchema(t *testing.T) {
	t.Run("UpdateSchema_Success", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		// Use a real PDPService but it will fail on HTTP calls - we're testing DB operations
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a schema first
		schema := models.Schema{
			SchemaID:          "sch_123",
			SchemaName:        "Original Name",
			SchemaDescription: stringPtr("Original Description"),
			SDL:               "type Query { original: String }",
			Endpoint:          "http://original.com",
			MemberID:          "member-123",
			Version:           string(models.ActiveVersion),
		}
		db.Create(&schema)

		newName := "Updated Name"
		newSDL := "type Query { updated: String }"
		req := &models.UpdateSchemaRequest{
			SchemaName: &newName,
			SDL:        &newSDL,
		}

		result, err := service.UpdateSchema(schema.SchemaID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.SchemaName)
		assert.Equal(t, newSDL, result.SDL)

		// Verify database was updated
		var updatedSchema models.Schema
		db.Where("schema_id = ?", schema.SchemaID).First(&updatedSchema)
		assert.Equal(t, newName, updatedSchema.SchemaName)
		assert.Equal(t, newSDL, updatedSchema.SDL)
	})

	t.Run("UpdateSchema_NotFound", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		newName := "Updated Name"
		req := &models.UpdateSchemaRequest{
			SchemaName: &newName,
		}

		result, err := service.UpdateSchema("non-existent-id", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "schema not found")
	})
}

func TestSchemaService_GetSchema(t *testing.T) {
	t.Run("GetSchema_Success", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a schema
		schema := models.Schema{
			SchemaID:          "sch_123",
			SchemaName:        "Test Schema",
			SchemaDescription: stringPtr("Test Description"),
			SDL:               "type Query { test: String }",
			Endpoint:          "http://example.com",
			MemberID:          "member-123",
			Version:           string(models.ActiveVersion),
		}
		db.Create(&schema)

		result, err := service.GetSchema(schema.SchemaID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, schema.SchemaID, result.SchemaID)
		assert.Equal(t, schema.SchemaName, result.SchemaName)
	})

	t.Run("GetSchema_NotFound", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		result, err := service.GetSchema("non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "schema not found")
	})
}

func TestSchemaService_GetSchemas(t *testing.T) {
	t.Run("GetSchemas_NoFilter", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create multiple schemas
		schemas := []models.Schema{
			{SchemaID: "sch_1", SchemaName: "Schema 1", SDL: "type Query { test1: String }", MemberID: "member-1", Endpoint: "http://example.com", Version: string(models.ActiveVersion)},
			{SchemaID: "sch_2", SchemaName: "Schema 2", SDL: "type Query { test2: String }", MemberID: "member-2", Endpoint: "http://example.com", Version: string(models.ActiveVersion)},
		}
		for _, s := range schemas {
			db.Create(&s)
		}

		result, err := service.GetSchemas(nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GetSchemas_WithMemberIDFilter", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create schemas
		memberID := "member-123"
		schema := models.Schema{
			SchemaID:   "sch_1",
			SchemaName: "Schema 1",
			SDL:        "type Query { test1: String }",
			MemberID:   memberID,
			Endpoint:   "http://example.com",
			Version:    string(models.ActiveVersion),
		}
		db.Create(&schema)

		result, err := service.GetSchemas(&memberID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)
	})
}

func TestSchemaService_CreateSchemaSubmission(t *testing.T) {
	t.Run("CreateSchemaSubmission_Success", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a member first
		member := models.Member{
			MemberID:    "member-123",
			Name:        "Test Member",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}
		db.Create(&member)

		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Submission",
			SchemaDescription: stringPtr("Test Description"),
			SDL:               "type Query { test: String }",
			SchemaEndpoint:    "http://example.com",
			MemberID:          member.MemberID,
		}

		result, err := service.CreateSchemaSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.SchemaName, result.SchemaName)
		assert.NotEmpty(t, result.SubmissionID)
		assert.Equal(t, string(models.StatusPending), result.Status)
	})

	t.Run("CreateSchemaSubmission_MemberNotFound", func(t *testing.T) {
		db := setupTestDBForSchema(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Submission",
			SchemaDescription: stringPtr("Test Description"),
			SDL:               "type Query { test: String }",
			SchemaEndpoint:    "http://example.com",
			MemberID:          "non-existent-member",
		}

		result, err := service.CreateSchemaSubmission(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "member not found")
	})
}

func stringPtr(s string) *string {
	return &s
}
