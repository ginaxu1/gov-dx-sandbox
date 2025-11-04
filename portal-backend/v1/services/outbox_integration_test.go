package services

import (
	"errors"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutboxPattern_EndToEnd_Schema tests the complete flow from schema creation to job processing
func TestOutboxPattern_EndToEnd_Schema(t *testing.T) {
	db := setupTestDB(t)
	successful := false
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			successful = true
			return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
		},
	}

	schemaService := NewSchemaService(db, mockPDP)
	worker := NewPDPWorker(db, mockPDP)
	worker.pollInterval = 100 * time.Millisecond // Faster polling for test

	// Step 1: Create schema (should create job atomically)
	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	response, err := schemaService.CreateSchema(req)
	require.NoError(t, err)
	assert.NotEmpty(t, response.SchemaID)

	// Step 2: Verify job exists and is pending
	var job models.PDPJob
	err = db.Where("schema_id = ?", response.SchemaID).First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)

	// Step 3: Process the job
	worker.processJob(&job)

	// Step 4: Verify job was completed
	var updatedJob models.PDPJob
	err = db.First(&updatedJob, "job_id = ?", job.JobID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.True(t, successful, "PDP service should have been called")
}

// TestOutboxPattern_EndToEnd_Application tests the complete flow from application creation to job processing
func TestOutboxPattern_EndToEnd_Application(t *testing.T) {
	db := setupTestDB(t)
	successful := false
	var actualApplicationID string
	mockPDP := &mockPDPService{
		updateAllowListFunc: func(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
			successful = true
			actualApplicationID = request.ApplicationID
			assert.NotEmpty(t, request.ApplicationID)
			assert.Equal(t, models.GrantDurationTypeOneMonth, request.GrantDuration)
			return &models.AllowListUpdateResponse{Records: []models.AllowListUpdateResponseRecord{}}, nil
		},
	}

	appService := NewApplicationService(db, mockPDP)
	worker := NewPDPWorker(db, mockPDP)

	// Step 1: Create application (should create job atomically)
	desc := "Test Description"
	req := &models.CreateApplicationRequest{
		ApplicationName:        "Test App",
		ApplicationDescription: &desc,
		SelectedFields: []models.SelectedFieldRecord{
			{FieldName: "person.name", SchemaID: "schema_123"},
		},
		MemberID: "member_123",
	}

	response, err := appService.CreateApplication(req)
	require.NoError(t, err)
	assert.NotEmpty(t, response.ApplicationID)

	// Step 2: Verify job exists and is pending
	var job models.PDPJob
	err = db.Where("application_id = ?", response.ApplicationID).First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)

	// Step 3: Process the job
	worker.processJob(&job)

	// Step 4: Verify job was completed
	var updatedJob models.PDPJob
	err = db.First(&updatedJob, "job_id = ?", job.JobID).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.True(t, successful, "PDP service should have been called")
	assert.Equal(t, response.ApplicationID, actualApplicationID, "PDP service should have been called with the correct application ID")
}

// TestOutboxPattern_Resilience tests that system recovers from PDP failures
func TestOutboxPattern_Resilience(t *testing.T) {
	db := setupTestDB(t)
	callCount := 0
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			callCount++
			// Fail first 2 times, succeed on 3rd
			if callCount < 3 {
				return nil, errors.New("PDP service temporarily down")
			}
			return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
		},
	}

	schemaService := NewSchemaService(db, mockPDP)
	worker := NewPDPWorker(db, mockPDP)

	// Create schema
	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	response, err := schemaService.CreateSchema(req)
	require.NoError(t, err)

	// Get the job
	var job models.PDPJob
	err = db.Where("schema_id = ?", response.SchemaID).First(&job).Error
	require.NoError(t, err)

	// Process job (first attempt - fails)
	worker.processJob(&job)
	db.First(&job, "job_id = ?", job.JobID)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)
	assert.Equal(t, 1, job.RetryCount)

	// Process job (second attempt - fails)
	worker.processJob(&job)
	db.First(&job, "job_id = ?", job.JobID)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)
	assert.Equal(t, 2, job.RetryCount)

	// Process job (third attempt - succeeds)
	worker.processJob(&job)
	db.First(&job, "job_id = ?", job.JobID)
	assert.Equal(t, models.PDPJobStatusCompleted, job.Status)
	assert.Equal(t, 3, job.RetryCount)
	assert.Equal(t, 3, callCount)
}

// TestOutboxPattern_NoCompensationNeeded tests that no compensation is needed with outbox pattern
func TestOutboxPattern_NoCompensationNeeded(t *testing.T) {
	db := setupTestDB(t)
	// Create a failing PDP service - should not affect schema creation
	failingPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP service is down")
		},
	}

	schemaService := NewSchemaService(db, failingPDP)

	desc := "Test Description"
	req := &models.CreateSchemaRequest{
		SchemaName:        "Test Schema",
		SchemaDescription: &desc,
		SDL:               "type Person { name: String }",
		Endpoint:          "http://example.com/graphql",
		MemberID:          "member_123",
	}

	// Schema creation should succeed immediately (PDP call happens asynchronously)
	response, err := schemaService.CreateSchema(req)
	require.NoError(t, err)
	assert.NotEmpty(t, response.SchemaID)

	// Verify schema exists
	var schema models.Schema
	err = db.First(&schema, "schema_id = ?", response.SchemaID).Error
	require.NoError(t, err)
	assert.Equal(t, req.SchemaName, schema.SchemaName)

	// Verify job exists (will be processed later by worker)
	var job models.PDPJob
	err = db.Where("schema_id = ?", response.SchemaID).First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)

	// Schema is NOT deleted - this is the key difference from the old pattern!
	// The worker will retry the PDP call until it succeeds
}
