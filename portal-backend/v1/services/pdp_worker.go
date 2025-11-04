package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PDPWorker processes PDP jobs from the outbox table
type PDPWorker struct {
	db           *gorm.DB
	pdpService   PDPClient
	pollInterval time.Duration
	batchSize    int
}

// NewPDPWorker creates a new PDP worker
func NewPDPWorker(db *gorm.DB, pdpService PDPClient) *PDPWorker {
	return &PDPWorker{
		db:           db,
		pdpService:   pdpService,
		pollInterval: 10 * time.Second,
		batchSize:    10,
	}
}

// Start starts the background worker that processes PDP jobs
func (w *PDPWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	slog.Info("PDP worker started", "pollInterval", w.pollInterval, "batchSize", w.batchSize)

	for {
		select {
		case <-ctx.Done():
			slog.Info("PDP worker stopped")
			return
		case <-ticker.C:
			w.processJobs()
		}
	}
}

// processJobs processes a batch of pending PDP jobs
func (w *PDPWorker) processJobs() {
	now := time.Now()

	// Fetch pending jobs that are ready for retry (using SELECT FOR UPDATE to prevent concurrent processing)
	// Only select jobs where next_retry_at is NULL or has passed
	var jobs []models.PDPJob

	// Clean up jobs stuck in "processing" status (e.g., from crashed workers)
	// Reset them to "pending" if they've been processing for more than 5 minutes
	stuckThreshold := now.Add(-5 * time.Minute)
	if err := w.db.Model(&models.PDPJob{}).
		Where("status = ?", models.PDPJobStatusProcessing).
		Where("updated_at < ?", stuckThreshold).
		Update("status", models.PDPJobStatusPending).Error; err != nil {
		slog.Warn("Failed to clean up stuck processing jobs", "error", err)
	}

	// Use a transaction with row-level locking to prevent concurrent processing
	err := w.db.Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE with SKIP LOCKED to avoid blocking other workers unnecessarily
		// This ensures only one worker can claim a job
		if err := tx.Where("status = ?", models.PDPJobStatusPending).
			Where("(next_retry_at IS NULL OR next_retry_at <= ?)", now).
			Order("created_at ASC").
			Limit(w.batchSize).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Find(&jobs).Error; err != nil {
			return err
		}

		// Atomically mark jobs as processing to prevent other workers from picking them up
		if len(jobs) > 0 {
			jobIDs := make([]string, len(jobs))
			for i := range jobs {
				jobIDs[i] = jobs[i].JobID
			}

			// Update status to processing within the same transaction
			if err := tx.Model(&models.PDPJob{}).
				Where("job_id IN ?", jobIDs).
				Update("status", models.PDPJobStatusProcessing).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("Failed to fetch pending PDP jobs", "error", err)
		return
	}

	if len(jobs) == 0 {
		return // No jobs to process
	}

	slog.Debug("Processing PDP jobs", "count", len(jobs))

	for i := range jobs {
		// Pass the pointer to the job we already fetched (status already updated to processing in transaction)
		w.processJob(&jobs[i])
	}
}

// processJob processes a single PDP job
func (w *PDPWorker) processJob(job *models.PDPJob) {
	var err error
	now := time.Now()

	switch job.JobType {
	case models.PDPJobTypeCreatePolicyMetadata:
		err = w.processCreatePolicyMetadata(job)
	case models.PDPJobTypeUpdateAllowList:
		err = w.processUpdateAllowList(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	// Calculate new retry count
	newRetryCount := job.RetryCount + 1

	// Update job status
	updates := map[string]interface{}{
		"processed_at": now,
		"retry_count":  newRetryCount,
	}

	if err != nil {
		errorMsg := err.Error()
		updates["error"] = &errorMsg

		// Check if we've exceeded max retries
		// newRetryCount represents the number of attempts made (including this one)
		// MaxRetries is the maximum number of retry attempts allowed
		// If newRetryCount > MaxRetries, we've exceeded the limit
		// Note: RetryCount starts at 0, so with MaxRetries=5:
		//   - RetryCount=0 (1st attempt) -> newRetryCount=1, allowed
		//   - RetryCount=4 (5th attempt) -> newRetryCount=5, allowed
		//   - RetryCount=5 (6th attempt) -> newRetryCount=6, exceeded
		if newRetryCount > job.MaxRetries {
			updates["status"] = models.PDPJobStatusFailed
			updates["next_retry_at"] = nil // Clear retry time since we're giving up
			slog.Error("PDP job failed after max retries",
				"jobID", job.JobID,
				"jobType", job.JobType,
				"retryCount", newRetryCount,
				"maxRetries", job.MaxRetries,
				"error", err)
		} else {
			// Exponential backoff: base delay 1 minute, doubled for each retry
			baseDelay := time.Minute
			backoffDelay := baseDelay * time.Duration(1<<job.RetryCount)
			nextRetryAt := now.Add(backoffDelay)
			updates["next_retry_at"] = &nextRetryAt
			updates["status"] = models.PDPJobStatusPending // Reset to pending for next retry

			slog.Warn("PDP job failed, will retry",
				"jobID", job.JobID,
				"jobType", job.JobType,
				"retryCount", newRetryCount,
				"maxRetries", job.MaxRetries,
				"error", err,
				"nextRetryAt", nextRetryAt)
		}
	} else {
		updates["status"] = models.PDPJobStatusCompleted
		updates["error"] = nil
		updates["next_retry_at"] = nil // Clear retry time on success
		slog.Info("PDP job completed successfully",
			"jobID", job.JobID,
			"jobType", job.JobType)
	}

	// Update the job record
	if updateErr := w.db.Model(job).Updates(updates).Error; updateErr != nil {
		slog.Error("Failed to update PDP job status",
			"jobID", job.JobID,
			"error", updateErr)
	}
}

// processCreatePolicyMetadata processes a create policy metadata job
func (w *PDPWorker) processCreatePolicyMetadata(job *models.PDPJob) error {
	if job.SchemaID == nil || job.SDL == nil {
		return fmt.Errorf("missing required fields for create policy metadata job")
	}

	_, err := w.pdpService.CreatePolicyMetadata(*job.SchemaID, *job.SDL)
	return err
}

// processUpdateAllowList processes an update allow list job
func (w *PDPWorker) processUpdateAllowList(job *models.PDPJob) error {
	if job.ApplicationID == nil || job.SelectedFields == nil {
		return fmt.Errorf("missing required fields for update allow list job")
	}

	// Parse SelectedFields from JSON string
	var selectedFields []models.SelectedFieldRecord
	if err := json.Unmarshal([]byte(*job.SelectedFields), &selectedFields); err != nil {
		return fmt.Errorf("failed to unmarshal selected fields: %w", err)
	}

	// Determine grant duration
	grantDuration := models.GrantDurationTypeOneMonth
	if job.GrantDuration != nil {
		grantDuration = models.GrantDurationType(*job.GrantDuration)
	}

	// Create allow list update request
	policyReq := models.AllowListUpdateRequest{
		ApplicationID: *job.ApplicationID,
		Records:       selectedFields,
		GrantDuration: grantDuration,
	}

	_, err := w.pdpService.UpdateAllowList(policyReq)
	return err
}
