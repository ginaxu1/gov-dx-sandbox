package services

import (
	"strings"
	"testing"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestSchemaService_UpdateSchema(t *testing.T) {
	t.Run("UpdateSchema_Success", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		// Use a real PDPService but it will fail on HTTP calls - we're testing DB operations
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a schema first
		desc := "Original Description"
		schema := models.Schema{
			SchemaID:          "sch_123",
			SchemaName:        "Original Name",
			SchemaDescription: &desc,
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a schema
		schema := models.Schema{
			SchemaID:          "sch_123",
			SchemaName:        "Test Schema",
			SchemaDescription: func() *string { s := "Test Description"; return &s }(),
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
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

		desc := "Test Description"
		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Submission",
			SchemaDescription: &desc,
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
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		desc := "Test Description"
		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Submission",
			SchemaDescription: &desc,
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

func TestSchemaService_UpdateSchemaSubmission(t *testing.T) {
	t.Run("UpdateSchemaSubmission_Success", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create member and submission
		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.SchemaSubmission{
			SubmissionID:   "sub_123",
			SchemaName:     "Original",
			SDL:            "type Query { original: String }",
			SchemaEndpoint: "http://original.com",
			MemberID:       member.MemberID,
			Status:         string(models.StatusPending),
		}
		db.Create(&submission)

		newName := "Updated"
		newSDL := "type Query { updated: String }"
		req := &models.UpdateSchemaSubmissionRequest{
			SchemaName: &newName,
			SDL:        &newSDL,
		}

		result, err := service.UpdateSchemaSubmission(submission.SubmissionID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.SchemaName)
		assert.Equal(t, newSDL, result.SDL)
	})

	t.Run("UpdateSchemaSubmission_NotFound", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		updatedName := "Updated"
		req := &models.UpdateSchemaSubmissionRequest{SchemaName: &updatedName}
		result, err := service.UpdateSchemaSubmission("non-existent", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "schema submission not found")
	})

	t.Run("UpdateSchemaSubmission_EmptySDL", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.SchemaSubmission{
			SubmissionID:   "sub_123",
			SchemaName:     "Test",
			SDL:            "type Query { test: String }",
			SchemaEndpoint: "http://example.com",
			MemberID:       member.MemberID,
			Status:         string(models.StatusPending),
		}
		db.Create(&submission)

		emptySDL := ""
		req := &models.UpdateSchemaSubmissionRequest{SDL: &emptySDL}
		result, err := service.UpdateSchemaSubmission(submission.SubmissionID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "SDL field cannot be empty")
	})
}

func TestSchemaService_GetSchemaSubmission(t *testing.T) {
	t.Run("GetSchemaSubmission_Success", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.SchemaSubmission{
			SubmissionID:   "sub_123",
			SchemaName:     "Test Submission",
			SDL:            "type Query { test: String }",
			SchemaEndpoint: "http://example.com",
			MemberID:       member.MemberID,
			Status:         string(models.StatusPending),
		}
		db.Create(&submission)

		result, err := service.GetSchemaSubmission(submission.SubmissionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, submission.SubmissionID, result.SubmissionID)
		assert.Equal(t, submission.SchemaName, result.SchemaName)
	})

	t.Run("GetSchemaSubmission_NotFound", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		result, err := service.GetSchemaSubmission("non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "schema submission not found")
	})
}

func TestSchemaService_GetSchemaSubmissions(t *testing.T) {
	t.Run("GetSchemaSubmissions_NoFilter", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submissions := []models.SchemaSubmission{
			{SubmissionID: "sub_1", SchemaName: "Sub 1", SDL: "type Query { test1: String }", SchemaEndpoint: "http://example.com", MemberID: member.MemberID, Status: string(models.StatusPending)},
			{SubmissionID: "sub_2", SchemaName: "Sub 2", SDL: "type Query { test2: String }", SchemaEndpoint: "http://example.com", MemberID: member.MemberID, Status: string(models.StatusPending)},
		}
		for _, s := range submissions {
			db.Create(&s)
		}

		result, err := service.GetSchemaSubmissions(nil, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GetSchemaSubmissions_WithMemberIDFilter", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		memberID := "member-123"
		member := models.Member{MemberID: memberID, Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.SchemaSubmission{
			SubmissionID:   "sub_1",
			SchemaName:     "Sub 1",
			SDL:            "type Query { test1: String }",
			SchemaEndpoint: "http://example.com",
			MemberID:       memberID,
			Status:         string(models.StatusPending),
		}
		db.Create(&submission)

		result, err := service.GetSchemaSubmissions(&memberID, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)
	})

	t.Run("GetSchemaSubmissions_WithStatusFilter", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submissions := []models.SchemaSubmission{
			{SubmissionID: "sub_1", SchemaName: "Sub 1", SDL: "type Query { test1: String }", SchemaEndpoint: "http://example.com", MemberID: member.MemberID, Status: string(models.StatusPending)},
			{SubmissionID: "sub_2", SchemaName: "Sub 2", SDL: "type Query { test2: String }", SchemaEndpoint: "http://example.com", MemberID: member.MemberID, Status: string(models.StatusApproved)},
		}
		for _, s := range submissions {
			db.Create(&s)
		}

		statusFilter := []string{string(models.StatusApproved)}
		result, err := service.GetSchemaSubmissions(nil, &statusFilter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, string(models.StatusApproved), result[0].Status)
	})
}

func TestSchemaService_CreateSchema_EdgeCases(t *testing.T) {
	t.Run("CreateSchema_EmptySDL", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		req := &models.CreateSchemaRequest{
			SchemaName: "Test Schema",
			SDL:        "",
		}

		_, err := service.CreateSchema(req)

		// Should fail validation or PDP call
		assert.Error(t, err)
	})

	t.Run("CreateSchema_WithOptionalFields", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		endpoint := "http://example.com/graphql"
		memberID := "member-123"
		req := &models.CreateSchemaRequest{
			SchemaName: "Test Schema",
			SDL:        "type Query { test: String }",
			Endpoint:   endpoint,
			MemberID:   memberID,
		}

		// Will fail on PDP call but tests the request structure
		_, err := service.CreateSchema(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create policy metadata")
		// Verify schema was deleted (compensation)
		var count int64
		db.Model(&models.Schema{}).Where("schema_name = ?", req.SchemaName).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CreateSchema_CompensationFailure", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		req := &models.CreateSchemaRequest{
			SchemaName: "Test Schema",
			SDL:        "type Query { test: String }",
			Endpoint:   "http://example.com/graphql",
			MemberID:   "member-123",
		}

		// Create schema manually first to simulate a scenario where deletion might fail
		schema := models.Schema{
			SchemaID:   "sch_manual",
			SchemaName: req.SchemaName,
			SDL:        req.SDL,
			Endpoint:   req.Endpoint,
			MemberID:   req.MemberID,
			Version:    string(models.ActiveVersion),
		}
		db.Create(&schema)

		// Now try to create with same name - will fail on duplicate or PDP
		_, err := service.CreateSchema(req)
		// This tests the compensation path when PDP fails
		assert.Error(t, err)
	})

	t.Run("CreateSchema_WithDescription", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		desc := "Test Description"
		req := &models.CreateSchemaRequest{
			SchemaName:        "Test Schema With Desc",
			SchemaDescription: &desc,
			SDL:               "type Query { test: String }",
			Endpoint:          "http://example.com/graphql",
			MemberID:          "member-123",
		}

		// Will fail on PDP call but tests description handling
		_, err := service.CreateSchema(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create policy metadata")
		// Verify schema was deleted (compensation)
		var count int64
		db.Model(&models.Schema{}).Where("schema_name = ?", req.SchemaName).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CreateSchema_DatabaseError", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		// Create a schema with invalid data that will cause DB error
		// Use a very long name that might exceed DB constraints
		longName := string(make([]byte, 10000)) // Very long string
		req := &models.CreateSchemaRequest{
			SchemaName: longName,
			SDL:        "type Query { test: String }",
			Endpoint:   "http://example.com/graphql",
			MemberID:   "member-123",
		}

		_, err := service.CreateSchema(req)
		// Should fail - either on database create or PDP call
		assert.Error(t, err)
		// Error could be from database (SQLite might allow long strings) or PDP (connection refused)
		errMsg := err.Error()
		assert.True(t,
			strings.Contains(errMsg, "failed to create schema") ||
				strings.Contains(errMsg, "failed to create policy metadata"),
			"Error should mention schema creation or policy metadata creation, got: %s", errMsg)
	})
}

func TestSchemaService_UpdateSchema_EdgeCases(t *testing.T) {
	t.Run("UpdateSchema_PartialUpdate", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		desc := "Original Description"
		schema := models.Schema{
			SchemaID:          "sch_123",
			SchemaName:        "Original Name",
			SchemaDescription: &desc,
			SDL:               "type Query { original: String }",
			Endpoint:          "http://original.com",
			MemberID:          "member-123",
			Version:           string(models.ActiveVersion),
		}
		db.Create(&schema)

		// Only update name, leave other fields unchanged
		newName := "Updated Name Only"
		req := &models.UpdateSchemaRequest{
			SchemaName: &newName,
		}

		result, err := service.UpdateSchema(schema.SchemaID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.SchemaName)
		// Original description should remain
		if schema.SchemaDescription != nil {
			assert.NotNil(t, result.SchemaDescription)
			assert.Equal(t, *schema.SchemaDescription, *result.SchemaDescription)
		}
	})

	t.Run("UpdateSchema_AllFields", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		schema := models.Schema{
			SchemaID:   "sch_123",
			SchemaName: "Original",
			SDL:        "type Query { original: String }",
			Endpoint:   "http://original.com",
			MemberID:   "member-123",
			Version:    string(models.ActiveVersion),
		}
		db.Create(&schema)

		newName := "Updated"
		newSDL := "type Query { updated: String }"
		newEndpoint := "http://updated.com"
		newVersion := "v2.0"
		req := &models.UpdateSchemaRequest{
			SchemaName: &newName,
			SDL:        &newSDL,
			Endpoint:   &newEndpoint,
			Version:    &newVersion,
		}

		result, err := service.UpdateSchema(schema.SchemaID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.SchemaName)
		assert.Equal(t, newSDL, result.SDL)
		assert.Equal(t, newEndpoint, result.Endpoint)
		assert.Equal(t, newVersion, result.Version)
	})
}

func TestSchemaService_CreateSchemaSubmission_EdgeCases(t *testing.T) {
	t.Run("CreateSchemaSubmission_WithPreviousSchemaID", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)

		// Create a previous schema
		previousSchema := models.Schema{
			SchemaID:   "sch_prev",
			SchemaName: "Previous Schema",
			SDL:        "type Query { prev: String }",
			Endpoint:   "http://prev.com",
			MemberID:   member.MemberID,
			Version:    string(models.ActiveVersion),
		}
		db.Create(&previousSchema)

		previousSchemaID := previousSchema.SchemaID
		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:       "New Submission",
			SDL:              "type Query { new: String }",
			SchemaEndpoint:   "http://new.com",
			MemberID:         member.MemberID,
			PreviousSchemaID: &previousSchemaID,
		}

		result, err := service.CreateSchemaSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, previousSchemaID, *result.PreviousSchemaID)
	})

	t.Run("CreateSchemaSubmission_InvalidPreviousSchemaID", func(t *testing.T) {
		db := SetupSQLiteTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewSchemaService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)

		invalidSchemaID := "non-existent-schema"
		req := &models.CreateSchemaSubmissionRequest{
			SchemaName:       "New Submission",
			SDL:              "type Query { new: String }",
			SchemaEndpoint:   "http://new.com",
			MemberID:         member.MemberID,
			PreviousSchemaID: &invalidSchemaID,
		}

		result, err := service.CreateSchemaSubmission(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "previous schema not found")
	})
}
