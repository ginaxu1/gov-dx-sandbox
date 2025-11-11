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

// AlertNotifier is an interface for sending high-priority alerts
type AlertNotifier interface {
	SendAlert(severity string, message string, details map[string]interface{}) error
}

// PDPWorker processes PDP jobs from the outbox table using a one-shot transactional state machine
// It does NOT retry - if PDP call fails, it compensates by deleting the main record
type PDPWorker struct {
	db            *gorm.DB
	pdpService    PDPClient
	pollInterval  time.Duration
	batchSize     int
	alertNotifier AlertNotifier // Optional alert notifier
}

// NewPDPWorker creates a new PDP worker
// alertNotifier can be nil - alerts will be logged but not sent to external systems
func NewPDPWorker(db *gorm.DB, pdpService PDPClient, alertNotifier AlertNotifier) *PDPWorker {
	return &PDPWorker{
		db:            db,
		pdpService:    pdpService,
		pollInterval:  10 * time.Second,
		batchSize:     10,
		alertNotifier: alertNotifier,
	}
}

// Start starts the background worker that processes PDP jobs
func (w *PDPWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	slog.Info("PDP worker started (one-shot mode)", "pollInterval", w.pollInterval, "batchSize", w.batchSize)

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
		// Pass the pointer to the job we already fetched (status already updated to processing in transaction)
		w.processJob(&jobs[i])
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
			compensationErrorMsg := fmt.Sprintf("PDP call failed: %v; Compensation failed: %v", err, compensationErr)
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

	// Update the job record
	if updateErr := w.db.Model(job).Updates(updates).Error; updateErr != nil {
		slog.Error("Failed to update PDP job status",
			"jobID", job.JobID,
			"error", updateErr)
	}
}

// compensate attempts to delete the main record to restore consistency
func (w *PDPWorker) compensate(job *models.PDPJob) error {
	switch job.JobType {
	case models.PDPJobTypeCreatePolicyMetadata:
		// Delete the schema record
		if job.SchemaID == nil {
			return fmt.Errorf("cannot compensate: schema_id is nil")
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
		// For UpdateAllowList, we don't delete the application (it may have been created successfully)
		// Instead, we just log that the allow list update failed
		// The application exists but without the allow list entry - this is acceptable
		// as the application can be updated later
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
