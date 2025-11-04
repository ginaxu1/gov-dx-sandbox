package services

import (
	"testing"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplicationService_CreateApplication_TransactionalOutbox tests that CreateApplication creates both application and job atomically
func TestApplicationService_CreateApplication_TransactionalOutbox(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewApplicationService(db, mockPDPService)

	desc := "Test Description"
	req := &models.CreateApplicationRequest{
		ApplicationName:        "Test App",
		ApplicationDescription: &desc,
		SelectedFields: []models.SelectedFieldRecord{
			{
				FieldName: "person.name",
				SchemaID:  "schema_123",
			},
		},
		MemberID: "member_123",
	}

	// Create application
	response, err := service.CreateApplication(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.ApplicationID)

	// Verify application was created
	var application models.Application
	err = db.First(&application, "application_id = ?", response.ApplicationID).Error
	require.NoError(t, err)
	assert.Equal(t, req.ApplicationName, application.ApplicationName)
	assert.Equal(t, len(req.SelectedFields), len(application.SelectedFields))

	// Verify PDP job was created atomically
	var job models.PDPJob
	err = db.Where("application_id = ?", response.ApplicationID).
		Where("job_type = ?", models.PDPJobTypeUpdateAllowList).
		First(&job).Error
	require.NoError(t, err)
	assert.Equal(t, models.PDPJobStatusPending, job.Status)
	assert.Equal(t, response.ApplicationID, *job.ApplicationID)
	assert.NotNil(t, job.SelectedFields) // Should contain serialized JSON
	assert.NotNil(t, job.GrantDuration)
	assert.Equal(t, 0, job.RetryCount)
	assert.Equal(t, 5, job.MaxRetries)
}

// TestApplicationService_CreateApplication_TransactionRollbackOnApplicationError tests rollback when application creation fails
func TestApplicationService_CreateApplication_TransactionRollbackOnApplicationError(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewApplicationService(db, mockPDPService)

	// Drop the Application table to simulate creation failure
	db.Migrator().DropTable(&models.Application{})

	desc := "Test Description"
	req := &models.CreateApplicationRequest{
		ApplicationName:        "Test App",
		ApplicationDescription: &desc,
		SelectedFields: []models.SelectedFieldRecord{
			{
				FieldName: "person.name",
				SchemaID:  "schema_123",
			},
		},
		MemberID: "member_123",
	}

	_, err := service.CreateApplication(req)
	require.Error(t, err)

	// Verify no job was created
	var jobCount int64
	db.Model(&models.PDPJob{}).Count(&jobCount)
	assert.Equal(t, int64(0), jobCount)
}

// TestApplicationService_CreateApplication_TransactionRollbackOnJobError tests rollback when job creation fails
func TestApplicationService_CreateApplication_TransactionRollbackOnJobError(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewApplicationService(db, mockPDPService)

	// Drop the PDPJob table to simulate job creation failure
	db.Migrator().DropTable(&models.PDPJob{})

	desc := "Test Description"
	req := &models.CreateApplicationRequest{
		ApplicationName:        "Test App",
		ApplicationDescription: &desc,
		SelectedFields: []models.SelectedFieldRecord{
			{
				FieldName: "person.name",
				SchemaID:  "schema_123",
			},
		},
		MemberID: "member_123",
	}

	_, err := service.CreateApplication(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create PDP job")

	// Verify application was NOT created
	var appCount int64
	db.Model(&models.Application{}).Count(&appCount)
	assert.Equal(t, int64(0), appCount)
}

// TestApplicationService_CreateApplication_SelectedFieldsSerialization tests that SelectedFields are properly serialized in the job
func TestApplicationService_CreateApplication_SelectedFieldsSerialization(t *testing.T) {
	db := setupTestDB(t)
	mockPDPService := NewPDPService("http://localhost:8082", "test-key")
	service := NewApplicationService(db, mockPDPService)

	desc := "Test Description"
	req := &models.CreateApplicationRequest{
		ApplicationName:        "Test App",
		ApplicationDescription: &desc,
		SelectedFields: []models.SelectedFieldRecord{
			{
				FieldName: "person.name",
				SchemaID:  "schema_123",
			},
			{
				FieldName: "person.email",
				SchemaID:  "schema_456",
			},
		},
		MemberID: "member_123",
	}

	response, err := service.CreateApplication(req)
	require.NoError(t, err)

	// Verify job contains serialized SelectedFields
	var job models.PDPJob
	err = db.Where("application_id = ?", response.ApplicationID).First(&job).Error
	require.NoError(t, err)
	assert.NotNil(t, job.SelectedFields)
	assert.Contains(t, *job.SelectedFields, "person.name")
	assert.Contains(t, *job.SelectedFields, "person.email")
	assert.Contains(t, *job.SelectedFields, "schema_123")
}
