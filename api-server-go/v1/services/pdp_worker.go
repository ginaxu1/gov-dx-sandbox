package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AlertNotifier is an interface for sending high-priority alerts
type AlertNotifier interface {
	SendAlert(severity string, message string, details map[string]interface{}) error
}

// PDPWorker processes PDP jobs from the outbox table using a one-shot transactional state machine
// It does NOT retry - if PDP call fails, it compensates by deleting the main record
type PDPWorker struct {
	db                *gorm.DB
	pdpService        PDPClient
	pollInterval      time.Duration
	batchSize         int
	alertNotifier     AlertNotifier   // Optional alert notifier
	stuckJobThreshold time.Duration   // Threshold for cleaning up stuck jobs in "processing" status
	cleanupCounter    int             // Counter to track when to run stuck job cleanup
	cleanupInterval   int             // Run cleanup every N polls (default: 10)
	inFlightJobs      sync.WaitGroup  // Tracks in-flight jobs for graceful shutdown
	shutdownCtx       context.Context // Context for graceful shutdown
	shutdownCancel    context.CancelFunc
	mu                sync.Mutex // Protects shutdown state
	isShuttingDown    bool
}

// NewPDPWorker creates a new PDP worker
// alertNotifier can be nil - alerts will be logged but not sent to external systems
// stuckJobThreshold defaults to 5 minutes if not provided (0 or negative value)
func NewPDPWorker(db *gorm.DB, pdpService PDPClient, alertNotifier AlertNotifier) *PDPWorker {
	return NewPDPWorkerWithConfig(db, pdpService, alertNotifier, 0)
}

// NewPDPWorkerWithConfig creates a new PDP worker with configurable thresholds
// If stuckJobThreshold is 0 or negative, it defaults to 5 minutes
// cleanupInterval controls how often stuck job cleanup runs (default: 10, meaning every 10th poll)
func NewPDPWorkerWithConfig(db *gorm.DB, pdpService PDPClient, alertNotifier AlertNotifier, stuckJobThreshold time.Duration) *PDPWorker {
	if stuckJobThreshold <= 0 {
		stuckJobThreshold = 5 * time.Minute
	}
	cleanupInterval := 10 // Run cleanup every 10th poll by default
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	return &PDPWorker{
		db:                db,
		pdpService:        pdpService,
		pollInterval:      10 * time.Second,
		batchSize:         10,
		alertNotifier:     alertNotifier,
		stuckJobThreshold: stuckJobThreshold,
		cleanupCounter:    0,
		cleanupInterval:   cleanupInterval,
		shutdownCtx:       shutdownCtx,
		shutdownCancel:    shutdownCancel,
	}
}

// Start starts the background worker that processes PDP jobs
func (w *PDPWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	slog.Info("PDP worker started (one-shot mode)", "pollInterval", w.pollInterval, "batchSize", w.batchSize)

	// Combine the provided context with the shutdown context
	combinedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Cancel combined context when shutdown is requested
	go func() {
		select {
		case <-w.shutdownCtx.Done():
			cancel()
		case <-combinedCtx.Done():
		}
	}()

	for {
		select {
		case <-combinedCtx.Done():
			slog.Info("PDP worker stopping, waiting for in-flight jobs to complete...")
			// Wait for all in-flight jobs to complete before returning
			w.inFlightJobs.Wait()
			slog.Info("PDP worker stopped (all jobs completed)")
			return
		case <-ticker.C:
			// Check if we're shutting down before processing new jobs
			w.mu.Lock()
			shuttingDown := w.isShuttingDown
			w.mu.Unlock()
			if shuttingDown {
				continue // Skip processing new jobs if shutting down
			}
			w.processJobs(combinedCtx)
		}
	}
}

// Shutdown gracefully stops the worker, waiting for in-flight jobs to complete
// Returns after all in-flight jobs are done or the timeout expires
func (w *PDPWorker) Shutdown(timeout time.Duration) error {
	w.mu.Lock()
	if w.isShuttingDown {
		w.mu.Unlock()
		// Already shutting down, just wait for completion
		done := make(chan struct{})
		go func() {
			w.inFlightJobs.Wait()
			close(done)
		}()
		select {
		case <-done:
			return nil
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for in-flight jobs to complete")
		}
	}
	w.isShuttingDown = true
	w.mu.Unlock()

	// Signal shutdown
	w.shutdownCancel()

	// Wait for in-flight jobs with timeout
	done := make(chan struct{})
	go func() {
		w.inFlightJobs.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("PDP worker shutdown complete (all jobs finished)")
		return nil
	case <-time.After(timeout):
		slog.Warn("PDP worker shutdown timeout - some jobs may still be in-flight")
		return fmt.Errorf("timeout waiting for in-flight jobs to complete")
	}
}

// processJobs processes a batch of pending PDP jobs
func (w *PDPWorker) processJobs(ctx context.Context) {
	now := time.Now()

	// Clean up jobs stuck in "processing" status (e.g., from crashed workers)
	// Run cleanup less frequently to reduce database load (every N polls)
	// Note: For optimal performance with large job volumes, add a composite index:
	// CREATE INDEX idx_pdp_jobs_status_updated_at ON pdp_jobs(status, updated_at);
	w.cleanupCounter++
	if w.cleanupCounter >= w.cleanupInterval {
		w.cleanupCounter = 0
		stuckThreshold := now.Add(-w.stuckJobThreshold)
		if err := w.db.Model(&models.PDPJob{}).
			Where("status = ?", models.PDPJobStatusProcessing).
			Where("updated_at < ?", stuckThreshold).
			Update("status", models.PDPJobStatusPending).Error; err != nil {
			slog.Warn("Failed to clean up stuck processing jobs", "error", err)
		}
	}

	// Use a transaction with row-level locking to prevent concurrent processing
	var jobs []models.PDPJob
	err := w.db.Transaction(func(tx *gorm.DB) error {
		// SELECT FOR UPDATE with SKIP LOCKED to avoid blocking other workers unnecessarily
		// This ensures only one worker can claim a job
		if err := tx.Where("status = ?", models.PDPJobStatusPending).
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
		// Check if we should stop processing new jobs
		select {
		case <-ctx.Done():
			slog.Info("Stopping job processing due to context cancellation")
			return
		default:
		}

		// Track in-flight job for graceful shutdown
		w.inFlightJobs.Add(1)
		// Process job synchronously (original behavior)
		// The WaitGroup ensures we wait for completion during shutdown
		func(job *models.PDPJob) {
			defer w.inFlightJobs.Done()
			// Pass the pointer to the job we already fetched (status already updated to processing in transaction)
			w.processJob(job)
		}(&jobs[i])
	}
}

// processJob processes a single PDP job using one-shot logic with compensation
func (w *PDPWorker) processJob(job *models.PDPJob) {
	now := time.Now()
	var err error

	// Attempt the PDP call exactly once
	switch job.JobType {
	case models.PDPJobTypeCreatePolicyMetadata:
		err = w.processCreatePolicyMetadata(job)
	case models.PDPJobTypeUpdateAllowList:
		err = w.processUpdateAllowList(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	updates := map[string]interface{}{
		"processed_at": now,
	}

	if err != nil {
		// Scenario B: PDP Call FAILED - Move to compensation
		errorMsg := err.Error()
		updates["error"] = &errorMsg

		slog.Warn("PDP job failed, attempting compensation",
			"jobID", job.JobID,
			"jobType", job.JobType,
			"error", err)

		// Attempt compensation (delete the main record)
		compensationErr := w.compensate(job)
		if compensationErr != nil {
			// Scenario B.2: Compensation also failed - CRITICAL ALERT
			updates["status"] = models.PDPJobStatusCompensationFailed
			// Use errors.Join() to properly combine errors (Go 1.20+)
			combinedErr := errors.Join(
				fmt.Errorf("PDP call failed: %w", err),
				fmt.Errorf("compensation failed: %w", compensationErr),
			)
			compensationErrorMsg := combinedErr.Error()
			updates["error"] = &compensationErrorMsg

			slog.Error("CRITICAL: Both PDP call and compensation failed",
				"jobID", job.JobID,
				"jobType", job.JobType,
				"pdpError", err,
				"compensationError", compensationErr)

			// Fire high-priority alert
			if w.alertNotifier != nil {
				alertErr := w.alertNotifier.SendAlert("critical",
					fmt.Sprintf("PDP job compensation failed for job %s", job.JobID),
					map[string]interface{}{
						"jobID":             job.JobID,
						"jobType":           job.JobType,
						"pdpError":          err.Error(),
						"compensationError": compensationErr.Error(),
						"schemaID":          job.SchemaID,
						"applicationID":     job.ApplicationID,
					})
				if alertErr != nil {
					slog.Error("Failed to send alert", "error", alertErr)
				}
			}
		} else {
			// Scenario B.1: Compensation succeeded
			updates["status"] = models.PDPJobStatusCompensated
			slog.Info("PDP job failed, compensation successful",
				"jobID", job.JobID,
				"jobType", job.JobType)
		}
	} else {
		// Scenario A: PDP Call SUCCEEDED
		updates["status"] = models.PDPJobStatusCompleted
		updates["error"] = nil
		slog.Info("PDP job completed successfully",
			"jobID", job.JobID,
			"jobType", job.JobType)
	}

	// Update the job record using a separate transaction to ensure it succeeds
	// This is critical because if the update fails, the job will remain in "processing" state
	// even though the work has been completed (or compensated)
	updateErr := w.updateJobStatusWithRetry(job, updates)
	if updateErr != nil {
		slog.Error("Failed to update PDP job status after retries",
			"jobID", job.JobID,
			"jobType", job.JobType,
			"error", updateErr)
		slog.Error("CRITICAL: PDP job status update failed",
			"jobID", job.JobID,
			"jobType", job.JobType,
			"intendedStatus", updates["status"],
			"error", updateErr)
		if w.alertNotifier != nil {
			alertDetails := map[string]interface{}{
				"jobID":          job.JobID,
				"jobType":        job.JobType,
				"intendedStatus": updates["status"],
				"error":          updateErr.Error(),
			}
			_ = w.alertNotifier.SendAlert("critical", "Failed to update PDP job status", alertDetails)
		}
	}
}

// updateJobStatusWithRetry attempts to update the job status with retry logic
// Uses a separate transaction to ensure the update succeeds even if the main transaction fails
func (w *PDPWorker) updateJobStatusWithRetry(job *models.PDPJob, updates map[string]interface{}) error {
	const maxRetries = 3
	const retryDelay = 100 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Use a separate transaction to ensure the status update succeeds
		err := w.db.Transaction(func(tx *gorm.DB) error {
			// Re-fetch the job to ensure we have the latest state
			var currentJob models.PDPJob
			if err := tx.Where("job_id = ?", job.JobID).First(&currentJob).Error; err != nil {
				return fmt.Errorf("failed to fetch job for update: %w", err)
			}

			// Update the job status
			if err := tx.Model(&currentJob).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update job status: %w", err)
			}

			return nil
		})

		if err == nil {
			return nil // Success
		}

		lastErr = err
		if attempt < maxRetries-1 {
			// Wait before retrying (exponential backoff)
			time.Sleep(retryDelay * time.Duration(1<<attempt))
			slog.Warn("Retrying PDP job status update",
				"jobID", job.JobID,
				"attempt", attempt+1,
				"maxRetries", maxRetries,
				"error", err)
		}
	}

	return fmt.Errorf("failed to update job status after %d attempts: %w", maxRetries, lastErr)
}

// compensate attempts to delete the main record to restore consistency
func (w *PDPWorker) compensate(job *models.PDPJob) error {
	switch job.JobType {
	case models.PDPJobTypeCreatePolicyMetadata:
		// Delete the schema record
		if job.SchemaID == nil {
			return fmt.Errorf("cannot compensate job %s: schema_id is nil", job.JobID)
		}
		result := w.db.Where("schema_id = ?", *job.SchemaID).Delete(&models.Schema{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete schema: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("schema not found for compensation: %s", *job.SchemaID)
		}
		slog.Info("Compensated by deleting schema", "schemaID", *job.SchemaID)
		return nil

	case models.PDPJobTypeUpdateAllowList:
		// For UpdateAllowList, we don't delete the application (it may have been created successfully).
		// Instead, we just log that the allow list update failed.
		// See documentation in docs/pdp_consistency.md for further details on handling and resolution of this state.
		slog.Info("Compensation not needed for UpdateAllowList - application remains",
			"applicationID", job.ApplicationID)
		return nil

	default:
		return fmt.Errorf("unknown job type for compensation: %s", job.JobType)
	}
}

// processCreatePolicyMetadata processes a create policy metadata job
func (w *PDPWorker) processCreatePolicyMetadata(job *models.PDPJob) error {
	if job.SchemaID == nil || job.SDL == nil {
		return fmt.Errorf("missing required fields for create policy metadata job (schema_id: %v, SDL: %v)", job.SchemaID != nil, job.SDL != nil)
	}

	_, err := w.pdpService.CreatePolicyMetadata(*job.SchemaID, *job.SDL)
	return err
}

// processUpdateAllowList processes an update allow list job
func (w *PDPWorker) processUpdateAllowList(job *models.PDPJob) error {
	if job.ApplicationID == nil || job.SelectedFields == nil {
		return fmt.Errorf("missing required fields for update allow list job (application_id: %v, selected_fields: %v)", job.ApplicationID != nil, job.SelectedFields != nil)
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
