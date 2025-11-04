package models

import (
	"time"

	"gorm.io/gorm"
)

// PDPJobType represents the type of PDP operation
type PDPJobType string

const (
	PDPJobTypeCreatePolicyMetadata PDPJobType = "create_policy_metadata"
	PDPJobTypeUpdateAllowList      PDPJobType = "update_allow_list"
)

// PDPJobStatus represents the status of a PDP job
type PDPJobStatus string

const (
	PDPJobStatusPending    PDPJobStatus = "pending"
	PDPJobStatusProcessing PDPJobStatus = "processing" // Job is currently being processed by a worker
	PDPJobStatusCompleted  PDPJobStatus = "completed"
	PDPJobStatusFailed     PDPJobStatus = "failed"
)

// PDPJob represents a job to be processed by the PDP worker
type PDPJob struct {
	JobID          string         `gorm:"primaryKey;type:varchar(255)" json:"job_id"`
	JobType        PDPJobType     `gorm:"type:varchar(50);not null" json:"job_type"`
	SchemaID       *string        `gorm:"type:varchar(255)" json:"schema_id,omitempty"`      // For policy metadata jobs
	ApplicationID  *string        `gorm:"type:varchar(255)" json:"application_id,omitempty"` // For allow list jobs
	SDL            *string        `gorm:"type:text" json:"sdl,omitempty"`                    // For policy metadata jobs
	SelectedFields *string        `gorm:"type:text" json:"selected_fields,omitempty"`        // For allow list jobs (JSON stored as TEXT for SQLite compatibility)
	GrantDuration  *string        `gorm:"type:varchar(50)" json:"grant_duration,omitempty"`  // For allow list jobs
	Status         PDPJobStatus   `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	RetryCount     int            `gorm:"not null;default:0" json:"retry_count"`
	MaxRetries     int            `gorm:"not null;default:5" json:"max_retries"`
	Error          *string        `gorm:"type:text" json:"error,omitempty"`
	NextRetryAt    *time.Time     `gorm:"type:timestamp" json:"next_retry_at,omitempty"` // When to retry (for exponential backoff)
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ProcessedAt    *time.Time     `json:"processed_at,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for PDPJob
func (PDPJob) TableName() string {
	return "pdp_jobs"
}
