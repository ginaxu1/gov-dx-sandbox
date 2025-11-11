package services

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaService_CreateSchema_TransactionalOutbox tests that CreateSchema creates both schema and job atomically
func TestSchemaService_CreateSchema_TransactionalOutbox(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewSchemaService(db, mockPDPService)

	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	// Create schema
	response, err := service.CreateSchema(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.SchemaID)

	// Verify schema was created
	var schema models.Schema
	err = db.First(&schema, "schema_id = ?", response.SchemaID).Error
	require.NoError(t, err)
	assert.Equal(t, req.SchemaName, schema.SchemaName)
	assert.Equal(t, req.SDL, schema.SDL)

	// Verify PDP job was created atomically
	var job models.PDPJob
	err = db.Where("schema_id = ?", response.SchemaID).
		Where("job_type = ?", models.PDPJobTypeCreatePolicyMetadata).
		First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)
	assert.Equal(t, response.SchemaID, *job.SchemaID)
	assert.Equal(t, req.SDL, *job.SDL)
}

// TestSchemaService_CreateSchema_TransactionRollbackOnSchemaError tests that transaction rolls back when schema creation fails
func TestSchemaService_CreateSchema_TransactionRollbackOnSchemaError(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewSchemaService(db, mockPDPService)

	// Create a schema with invalid member_id (assuming foreign key constraint)
	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "invalid_member",
	}

	// Try to create schema - this should fail if there's a constraint
	// For this test, we'll manually delete the schema table to simulate failure
	db.Migrator().DropTable(&models.Schema{})

	_, err := service.CreateSchema(req)
	require.Error(t, err)

	// Verify no job was created (transaction should have rolled back)
	var jobCount int64
	db.Model(&models.PDPJob{}).Count(&jobCount)
	assert.Equal(t, int64(0), jobCount, "No job should be created if schema creation fails")
}

// TestSchemaService_CreateSchema_TransactionRollbackOnJobError tests that transaction rolls back when job creation fails
func TestSchemaService_CreateSchema_TransactionRollbackOnJobError(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewSchemaService(db, mockPDPService)

	// Drop the PDPJob table to simulate job creation failure
	db.Migrator().DropTable(&models.PDPJob{})

	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	_, err := service.CreateSchema(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create PDP job")

	// Verify schema was NOT created (transaction should have rolled back)
	var schemaCount int64
	db.Model(&models.Schema{}).Count(&schemaCount)
	assert.Equal(t, int64(0), schemaCount, "Schema should not be created if job creation fails")
}

// TestSchemaService_CreateSchema_AtomicityOnCommitFailure tests that both operations are rolled back if transaction fails
func TestSchemaService_CreateSchema_AtomicityOnCommitFailure(t *testing.T) {
	// This test verifies that if transaction fails (commit or start), nothing is persisted
	// In a real scenario, transaction failures are rare, but we should handle them
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewSchemaService(db, mockPDPService)

	// Close the database connection to simulate transaction failure
	sqlDB, _ := db.DB()
	sqlDB.Close()

	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	_, err := service.CreateSchema(req)
	require.Error(t, err)
	// Transaction failures (start or commit) should prevent any data persistence
	assert.Contains(t, err.Error(), "transaction", "Error should mention transaction failure")

	// Verify nothing was persisted (no schema or job should exist)
	var schemaCount int64
	db.Model(&models.Schema{}).Count(&schemaCount)
	assert.Equal(t, int64(0), schemaCount, "No schema should be created when transaction fails")

	var jobCount int64
	db.Model(&models.PDPJob{}).Count(&jobCount)
	assert.Equal(t, int64(0), jobCount, "No job should be created when transaction fails")
}

// TestSchemaService_CreateSchema_ReturnsImmediately tests that CreateSchema returns immediately without waiting for PDP
func TestSchemaService_CreateSchema_ReturnsImmediately(t *testing.T) {
	db := setupTestDB(t)
	// Use a mock PDP service that would fail if called (but shouldn't be called)
	mockPDPService := NewPDPService("http://invalid-url:9999", "test-key")
	service := NewSchemaService(db, mockPDPService)

	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	// This should return immediately without calling PDP
	response, err := service.CreateSchema(req)
	require.NoError(t, err)
	assert.NotNil(t, response)

	// Verify job is pending (not processed yet)
	var job models.PDPJob
	err = db.Where("schema_id = ?", response.SchemaID).First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)
}
