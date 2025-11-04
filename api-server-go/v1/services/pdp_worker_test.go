package services

import (
	"context"
	"encoding/json"
	"errors"
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
	worker := NewPDPWorker(db, mockPDP)

	// Create a pending job
	job := models.PDPJob{
		JobID:      "job_123",
		JobType:    models.PDPJobTypeCreatePolicyMetadata,
		SchemaID:   stringPtr("schema_123"),
		SDL:        stringPtr("type Person { name: String }"),
		Status:     models.PDPJobStatusPending,
		RetryCount: 0,
		MaxRetries: 5,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job was completed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1, updatedJob.RetryCount)
	assert.NotNil(t, updatedJob.ProcessedAt)
	assert.Nil(t, updatedJob.Error)
}

// TestPDPWorker_ProcessUpdateAllowListJob tests processing of update allow list jobs
func TestPDPWorker_ProcessUpdateAllowListJob(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP)

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
		RetryCount:     0,
		MaxRetries:     5,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job was completed
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 1, updatedJob.RetryCount)
}

// TestPDPWorker_ProcessJob_RetryOnFailure tests that jobs are retried on failure
func TestPDPWorker_ProcessJob_RetryOnFailure(t *testing.T) {
	db := setupTestDB(t)
	callCount := 0
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			callCount++
			if callCount < 3 {
				return nil, errors.New("PDP service temporarily unavailable")
			}
			return &models.PolicyMetadataCreateResponse{Records: []models.PolicyMetadataResponse{}}, nil
		},
	}
	worker := NewPDPWorker(db, mockPDP)

	// Create a pending job
	job := models.PDPJob{
		JobID:      "job_retry",
		JobType:    models.PDPJobTypeCreatePolicyMetadata,
		SchemaID:   stringPtr("schema_123"),
		SDL:        stringPtr("type Person { name: String }"),
		Status:     models.PDPJobStatusPending,
		RetryCount: 0,
		MaxRetries: 5,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job (first attempt - should fail)
	worker.processJob(&job)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusPending, updatedJob.Status) // Still pending
	assert.Equal(t, 1, updatedJob.RetryCount)
	assert.NotNil(t, updatedJob.Error)

	// Process again (second attempt - should fail)
	worker.processJob(&updatedJob)
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusPending, updatedJob.Status)
	assert.Equal(t, 2, updatedJob.RetryCount)

	// Process again (third attempt - should succeed)
	worker.processJob(&updatedJob)
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusCompleted, updatedJob.Status)
	assert.Equal(t, 3, updatedJob.RetryCount)
	assert.Equal(t, 3, callCount)
}

// TestPDPWorker_ProcessJob_MaxRetriesExceeded tests that jobs fail after max retries
func TestPDPWorker_ProcessJob_MaxRetriesExceeded(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{
		createPolicyMetadataFunc: func(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error) {
			return nil, errors.New("PDP service permanently unavailable")
		},
	}
	worker := NewPDPWorker(db, mockPDP)

	// Create a pending job with retry count at max (next attempt will exceed)
	job := models.PDPJob{
		JobID:      "job_max_retries",
		JobType:    models.PDPJobTypeCreatePolicyMetadata,
		SchemaID:   stringPtr("schema_123"),
		SDL:        stringPtr("type Person { name: String }"),
		Status:     models.PDPJobStatusPending,
		RetryCount: 5, // At max retries (5 attempts made), next will exceed
		MaxRetries: 5,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job (should exceed max retries)
	worker.processJob(&job)

	// Verify job was marked as failed (6th attempt exceeded MaxRetries=5)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusFailed, updatedJob.Status)
	assert.Equal(t, 6, updatedJob.RetryCount) // 6 attempts made (exceeded MaxRetries=5)
	assert.NotNil(t, updatedJob.Error)
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
	worker := NewPDPWorker(db, mockPDP)
	worker.batchSize = 3

	// Create multiple pending jobs
	for i := 0; i < 5; i++ {
		job := models.PDPJob{
			JobID:      "job_batch_" + string(rune(i)),
			JobType:    models.PDPJobTypeCreatePolicyMetadata,
			SchemaID:   stringPtr("schema_123"),
			SDL:        stringPtr("type Person { name: String }"),
			Status:     models.PDPJobStatusPending,
			RetryCount: 0,
			MaxRetries: 5,
		}
		require.NoError(t, db.Create(&job).Error)
	}

	// Process jobs
	worker.processJobs()

	// Verify only batchSize jobs were processed
	assert.Equal(t, worker.batchSize, processedCount)
}

// TestPDPWorker_ProcessJobs_NoJobs tests that worker handles empty job queue gracefully
func TestPDPWorker_ProcessJobs_NoJobs(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP)

	// Process jobs when there are none
	worker.processJobs()

	// Should not panic or error
	var jobCount int64
	db.Model(&models.PDPJob{}).Count(&jobCount)
	assert.Equal(t, int64(0), jobCount)
}

// TestPDPWorker_Start_StopsOnContextCancel tests that worker stops gracefully
func TestPDPWorker_Start_StopsOnContextCancel(t *testing.T) {
	db := setupTestDB(t)
	mockPDP := &mockPDPService{}
	worker := NewPDPWorker(db, mockPDP)
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
	worker := NewPDPWorker(db, mockPDP)

	// Create a job with invalid type
	job := models.PDPJob{
		JobID:      "job_invalid",
		JobType:    "invalid_type",
		Status:     models.PDPJobStatusPending,
		RetryCount: 0,
		MaxRetries: 5,
	}
	require.NoError(t, db.Create(&job).Error)

	// Process the job
	worker.processJob(&job)

	// Verify job has error but is still pending (will retry)
	var updatedJob models.PDPJob
	require.NoError(t, db.First(&updatedJob, "job_id = ?", job.JobID).Error)
	assert.Equal(t, models.PDPJobStatusPending, updatedJob.Status)
	assert.NotNil(t, updatedJob.Error)
	assert.Contains(t, *updatedJob.Error, "unknown job type")
}
