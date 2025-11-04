package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
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
	var jobs []models.PDPJob

	// Fetch pending jobs
	if err := w.db.Where("status = ?", models.PDPJobStatusPending).
		Order("created_at ASC").
		Limit(w.batchSize).
		Find(&jobs).Error; err != nil {
		slog.Error("Failed to fetch pending PDP jobs", "error", err)
		return
	}

	if len(jobs) == 0 {
		return // No jobs to process
	}

	slog.Debug("Processing PDP jobs", "count", len(jobs))

	for _, job := range jobs {
		w.processJob(&job)
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

	// Update job status
	updates := map[string]interface{}{
		"processed_at": now,
		"retry_count":  job.RetryCount + 1,
	}

	if err != nil {
		errorMsg := err.Error()
		updates["error"] = &errorMsg

		// Check if we've exceeded max retries
		if job.RetryCount+1 >= job.MaxRetries {
			updates["status"] = models.PDPJobStatusFailed
			slog.Error("PDP job failed after max retries",
				"jobID", job.JobID,
				"jobType", job.JobType,
				"retryCount", job.RetryCount+1,
				"error", err)
		} else {
			slog.Warn("PDP job failed, will retry",
				"jobID", job.JobID,
				"jobType", job.JobType,
				"retryCount", job.RetryCount+1,
				"maxRetries", job.MaxRetries,
				"error", err)
		}
	} else {
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
