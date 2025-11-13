package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDPWorker_ProcessCreatePolicyMetadataJob tests processing of create policy metadata jobs
func TestPDPWorker_ProcessCreatePolicyMetadataJob(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP, nil)

	// Create a pending job
	job := models.PDPJob{
		JobID:    "job_123",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: stringPtr("schema_123"),
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job was completed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.NotNil(t, updatedJob.ProcessedAt)
	assert.Nil(t, updatedJob.Error)
}

// TestPDPWorker_ProcessUpdateAllowListJob tests processing of update allow list jobs
func TestPDPWorker_ProcessUpdateAllowListJob(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP, nil)

	// Serialize SelectedFields
	selectedFields := []models.SelectedFieldRecord{
		{FieldName: "person.name", SchemaID: "schema_123"},
	}
	fieldsJSON, _ := json.Marshal(selectedFields)
	fieldsStr := string(fieldsJSON)
	grantDuration := string(models.GrantDurationTypeOneMonth)

	// Create a pending job
	job := models.PDPJob{
		JobID:          "job_456",
		JobType:        models.PDPJobTypeUpdateAllowList,
		ApplicationID:  stringPtr("app_123"),
		SelectedFields: &fieldsStr,
		GrantDuration:  &grantDuration,
		Status:         models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job was completed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
}

// TestPDPWorker_OneShot_Success tests Scenario A: PDP call succeeds
func TestPDPWorker_OneShot_Success(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create a schema and job
	schemaID := "schema_123"
	schema := models.Schema{
		SchemaID:   schemaID,
		SchemaName: "Test Schema",
		SDL:        "type Person { name: String }",
		MemberID:   "member_123",
	}
	require.NoError(t, db.Create(&schema).Error)

	job := models.PDPJob{
		JobID:    "job_success",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job status is completed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.Nil(t, updatedJob.Error)
	assert.NotNil(t, updatedJob.ProcessedAt)

	// Verify schema still exists (not deleted)
	var updatedSchema models.Schema
	err := db.First(&updatedSchema, "schema_id = ?", schemaID).Error
	require.NoError(t, err)
	assert.Equal(t, schemaID, updatedSchema.SchemaID)

	// Verify no alerts were sent
	assert.Equal(t, 0, len(alertNotifier.alerts))
}

// TestPDPWorker_OneShot_FailureWithCompensation tests Scenario B.1: PDP fails, compensation succeeds
func TestPDPWorker_OneShot_FailureWithCompensation(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP service unavailable")
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create a schema and job
	schemaID := "schema_456"
	schema := models.Schema{
		SchemaID:   schemaID,
		SchemaName: "Test Schema",
		SDL:        "type Person { name: String }",
		MemberID:   "member_123",
	}
	require.NoError(t, db.Create(&schema).Error)

	job := models.PDPJob{
		JobID:    "job_compensated",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job status is compensated
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompensated, updatedJob.Status)
	assert.NotNil(t, updatedJob.Error)
	assert.Contains(t, *updatedJob.Error, "PDP service unavailable")
	assert.NotNil(t, updatedJob.ProcessedAt)

	// Verify schema was deleted (compensation succeeded)
	var deletedSchema models.Schema
	err := db.First(&deletedSchema, "schema_id = ?", schemaID).Error
	assert.Error(t, err, "Schema should have been deleted")

	// Verify no critical alerts were sent (compensation succeeded)
	assert.Equal(t, 0, len(alertNotifier.alerts))
}

// TestPDPWorker_OneShot_CompensationFailure tests Scenario B.2: Both PDP and compensation fail
func TestPDPWorker_OneShot_CompensationFailure(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP service unavailable")
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create a schema and job
	schemaID := "schema_789"
	schema := models.Schema{
		SchemaID:   schemaID,
		SchemaName: "Test Schema",
		SDL:        "type Person { name: String }",
		MemberID:   "member_123",
	}
	require.NoError(t, db.Create(&schema).Error)

	job := models.PDPJob{
		JobID:    "job_compensation_failed",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Delete the schema BEFORE processing to simulate compensation failure
	// (schema doesn't exist when compensation tries to delete it)
	db.Where("schema_id = ?", schemaID).Delete(&models.Schema{})

	// Process the job
	worker.processJob(&job)

	// Verify job status is compensation_failed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompensationFailed, updatedJob.Status)
	assert.NotNil(t, updatedJob.Error)
	assert.Contains(t, *updatedJob.Error, "PDP call failed")
	assert.Contains(t, *updatedJob.Error, "compensation failed")
	assert.NotNil(t, updatedJob.ProcessedAt)

	// Verify critical alert was sent
	require.Equal(t, 1, len(alertNotifier.alerts))
	alert := alertNotifier.alerts[0]
	assert.Equal(t, "critical", alert.severity)
	assert.Contains(t, alert.message, "compensation failed")
	assert.Contains(t, alert.details["jobID"].(string), "job_compensation_failed")
}

// TestPDPWorker_OneShot_NoRetries tests that jobs are NOT retried
func TestPDPWorker_OneShot_NoRetries(t *testing.T) {
	db := setupTestDB(t)
	callCount := 0
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			callCount++
			return nil, errors.New("PDP service unavailable")
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create a schema and job
	schemaID := "schema_no_retry"
	schema := models.Schema{
		SchemaID:   schemaID,
		SchemaName: "Test Schema",
		SDL:        "type Person { name: String }",
		MemberID:   "member_123",
	}
	require.NoError(t, db.Create(&schema).Error)

	job := models.PDPJob{
		JobID:    "job_no_retry",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job once - should call PDP once and compensate
	worker.processJob(&job)

	// Verify PDP was called exactly once (no retries)
	assert.Equal(t, 1, callCount, "PDP should be called exactly once, no retries")

	// Verify job is compensated (not pending for retry)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompensated, updatedJob.Status, "Job should be compensated, not pending for retry")

	// Verify schema was deleted
	var deletedSchema models.Schema
	err := db.First(&deletedSchema, "schema_id = ?", schemaID).Error
	assert.Error(t, err, "Schema should have been deleted")

	// Note: In practice, the worker only picks up jobs with status='pending',
	// so a compensated job won't be reprocessed. The processJob method itself
	// doesn't check status (it's called by the worker after status is set to 'processing'),
	// but the worker's processJobs() method only selects pending jobs.
}

// TestPDPWorker_OneShot_UpdateAllowList_NoCompensation tests UpdateAllowList doesn't need compensation
func TestPDPWorker_OneShot_UpdateAllowList_NoCompensation(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		updateAllowListFunc: func(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
			return nil, errors.New("PDP service unavailable")
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create an application
	applicationID := "app_123"
	application := models.Application{
		ApplicationID:   applicationID,
		ApplicationName: "Test App",
		MemberID:        "member_123",
		SelectedFields:  models.SelectedFieldRecords{},
		Version:         string(models.ActiveVersion),
	}
	require.NoError(t, db.Create(&application).Error)

	// Create job
	selectedFieldsJSON := `[{"fieldName":"person.name","schemaId":"schema_123"}]`
	job := models.PDPJob{
		JobID:          "job_allowlist",
		JobType:        models.PDPJobTypeUpdateAllowList,
		ApplicationID:  &applicationID,
		SelectedFields: &selectedFieldsJSON,
		Status:         models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job status is compensated (compensation succeeds immediately for UpdateAllowList)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompensated, updatedJob.Status)

	// Verify application still exists (not deleted - UpdateAllowList doesn't delete)
	var updatedApp models.Application
	err := db.First(&updatedApp, "application_id = ?", applicationID).Error
	require.NoError(t, err)
	assert.Equal(t, applicationID, updatedApp.ApplicationID)
}

// TestPDPWorker_OneShot_StateMachineTransitions tests all state transitions
func TestPDPWorker_OneShot_StateMachineTransitions(t *testing.T) {
	db := setupTestDB(t)
	alertNotifier := &mockAlertNotifier{}

	t.Run("pending -> processing -> completed", func(t *testing.T) {
		mockPDP := &mockPDPService{
			createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
				return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
			},
		}
		worker := NewPDPWorker(db, mockPDP, alertNotifier)

		schemaID := "schema_state_1"
		schema := models.Schema{SchemaID: schemaID, SchemaName: "Test", MemberID: "member_123"}
		require.NoError(t, db.Create(&schema).Error)

		job := models.PDPJob{
			JobID:    "job_state_1",
			JobType:  models.PDPJobTypeCreatePolicyMetadata,
			SchemaID: &schemaID,
			SDL:      stringPtr("type Person { name: String }"),
			Status:   models.PDPJobStatusPending,
		}
		require.NoError(t, db.Create(&job).Error)

		// Mark as processing (simulating worker fetch)
		db.Model(&job).Update("status", models.PDPJobStatusProcessing)

		// Process
		worker.processJob(&job)

		var updatedJob models.PDPJob
		require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
		assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	})

	t.Run("pending -> processing -> compensated", func(t *testing.T) {
		mockPDP := &mockPDPService{
			createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
				return nil, errors.New("PDP failed")
			},
		}
		worker := NewPDPWorker(db, mockPDP, alertNotifier)

		schemaID := "schema_state_2"
		schema := models.Schema{SchemaID: schemaID, SchemaName: "Test", MemberID: "member_123"}
		require.NoError(t, db.Create(&schema).Error)

		job := models.PDPJob{
			JobID:    "job_state_2",
			JobType:  models.PDPJobTypeCreatePolicyMetadata,
			SchemaID: &schemaID,
			SDL:      stringPtr("type Person { name: String }"),
			Status:   models.PDPJobStatusPending,
		}
		require.NoError(t, db.Create(&job).Error)

		db.Model(&job).Update("status", models.PDPJobStatusProcessing)
		worker.processJob(&job)

		var updatedJob models.PDPJob
		require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
		assert.Equal(t, models.PDPJobStatusCompensated, updatedJob.Status)
	})

	t.Run("pending -> processing -> compensation_failed", func(t *testing.T) {
		mockPDP := &mockPDPService{
			createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
				return nil, errors.New("PDP failed")
			},
		}
		worker := NewPDPWorker(db, mockPDP, alertNotifier)

		schemaID := "schema_state_3"
		job := models.PDPJob{
			JobID:    "job_state_3",
			JobType:  models.PDPJobTypeCreatePolicyMetadata,
			SchemaID: &schemaID,
			SDL:      stringPtr("type Person { name: String }"),
			Status:   models.PDPJobStatusPending,
		}
		require.NoError(t, db.Create(&job).Error)

		// Don't create schema - compensation will fail
		db.Model(&job).Update("status", models.PDPJobStatusProcessing)
		worker.processJob(&job)

		var updatedJob models.PDPJob
		require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
		assert.Equal(t, models.PDPJobStatusCompensationFailed, updatedJob.Status)
	})
}

// TestPDPWorker_OneShot_AlertNotifierNil tests that nil alert notifier doesn't crash
func TestPDPWorker_OneShot_AlertNotifierNil(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP failed")
		},
	}
	// Pass nil alert notifier
	worker := NewPDPWorker(db, mockPDP, nil)

	schemaID := "schema_nil_alert"
	job := models.PDPJob{
		JobID:    "job_nil_alert",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Don't create schema - compensation will fail
	// Should not panic even with nil alert notifier
	worker.processJob(&job)

	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompensationFailed, updatedJob.Status)
}

// TestPDPWorker_OneShot_CompensationDeletesCorrectRecord tests compensation deletes the right schema
func TestPDPWorker_OneShot_CompensationDeletesCorrectRecord(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP failed")
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	// Create two schemas
	schemaID1 := "schema_1"
	schemaID2 := "schema_2"
	schema1 := models.Schema{SchemaID: schemaID1, SchemaName: "Schema 1", MemberID: "member_123"}
	schema2 := models.Schema{SchemaID: schemaID2, SchemaName: "Schema 2", MemberID: "member_123"}
	require.NoError(t, db.Create(&schema1).Error)
	require.NoError(t, db.Create(&schema2).Error)

	// Create job for schema1
	job := models.PDPJob{
		JobID:    "job_selective",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID1,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify only schema1 was deleted
	var deletedSchema1 models.Schema
	err1 := db.First(&deletedSchema1, "schema_id = ?", schemaID1).Error
	assert.Error(t, err1, "Schema 1 should be deleted")

	var existingSchema2 models.Schema
	err2 := db.First(&existingSchema2, "schema_id = ?", schemaID2).Error
	require.NoError(t, err2, "Schema 2 should still exist")
	assert.Equal(t, schemaID2, existingSchema2.SchemaID)
}

// TestPDPWorker_OneShot_ErrorDetailsStored tests that error details are properly stored
func TestPDPWorker_OneShot_ErrorDetailsStored(t *testing.T) {
	db := setupTestDB(t)
	pdpError := errors.New("PDP connection timeout")
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, pdpError
		},
	}
	alertNotifier := &mockAlertNotifier{}
	worker := NewPDPWorker(db, mockPDP, alertNotifier)

	schemaID := "schema_error_details"
	schema := models.Schema{SchemaID: schemaID, SchemaName: "Test", MemberID: "member_123"}
	require.NoError(t, db.Create(&schema).Error)

	job := models.PDPJob{
		JobID:    "job_error_details",
		JobType:  models.PDPJobTypeCreatePolicyMetadata,
		SchemaID: &schemaID,
		SDL:      stringPtr("type Person { name: String }"),
		Status:   models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	worker.processJob(&job)

	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.NotNil(t, updatedJob.Error)
	assert.Contains(t, *updatedJob.Error, "PDP connection timeout")
}

// TestPDPWorker_ProcessJobs_BatchProcessing tests that worker processes jobs in batches
func TestPDPWorker_ProcessJobs_BatchProcessing(t *testing.T) {
	db := setupTestDB(t)
	processedCount := 0
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			processedCount++
			return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
		},
	}
	worker := NewPDPWorker(db, mockPDP, nil)
	worker.batchSize = 3

	// Create multiple pending jobs
	for i := 0; i < 5; i++ {
		job := models.PDPJob{
			JobID:    fmt.Sprintf("job_batch_%d", i),
			JobType:  models.PDPJobTypeCreatePolicyMetadata,
			SchemaID: stringPtr("schema_123"),
			SDL:      stringPtr("type Person { name: String }"),
			Status:   models.PDPJobStatusPending,
		}
		require.NoError(t, db.Create(&job).Error)
	}

	// Process jobs
	worker.processJobs(context.Background())

	// Verify only batchSize jobs were processed
	assert.Equal(t, worker.batchSize, processedCount)
}

// TestPDPWorker_ProcessJobs_NoJobs tests that worker handles empty job queue gracefully
func TestPDPWorker_ProcessJobs_NoJobs(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP, nil)

	// Process jobs when there are none
	worker.processJobs(context.Background())

	// Should not panic or error
	var jobCount int64
	db.Model(&models.PDPJob{}).Count(&jobCount)
	assert.Equal(t, int64(0), jobCount)
}

// TestPDPWorker_Start_StopsOnContextCancel tests that worker stops gracefully
func TestPDPWorker_Start_StopsOnContextCancel(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP, nil)
	worker.pollInterval = 100 * time.Millisecond // Faster for testing

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker in goroutine
	done := make(chan bool)
	go func() {
		worker.Start(ctx)
		done <- true
	}()

	// Cancel context after short delay
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not stop within timeout")
	}
}

// TestPDPWorker_ProcessJob_InvalidJobType tests handling of unknown job types
func TestPDPWorker_ProcessJob_InvalidJobType(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP, nil)

	// Create a job with invalid type
	job := models.PDPJob{
		JobID:   "job_invalid",
		JobType: "invalid_type",
		Status:  models.PDPJobStatusPending,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job has error and is compensated (unknown job type triggers compensation attempt)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	// Unknown job type will fail compensation too (no schema to delete), so status is compensation_failed
	assert.Equal(t, models.PDPJobStatusCompensationFailed, updatedJob.Status)
	assert.NotNil(t, updatedJob.Error)
	assert.Contains(t, *updatedJob.Error, "unknown job type")
}
