package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// SchemaService handles schema-related operations
type SchemaService struct {
	db *gorm.DB
}

// NewSchemaService creates a new schema service
func NewSchemaService(db *gorm.DB) *SchemaService {
	return &SchemaService{db: db}
}

// CreateSchema creates a new schema
func (s *SchemaService) CreateSchema(req *models.CreateSchemaRequest) (*models.SchemaResponse, error) {
	schema := models.Schema{
		SchemaID:          "sch_" + uuid.New().String(),
		SchemaName:        req.SchemaName,
		SchemaDescription: req.SchemaDescription,
		SDL:               req.SDL,
		Endpoint:          req.Endpoint,
		ProviderID:        req.ProviderID,
	}
	if err := s.db.Create(&schema).Error; err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	response := &models.SchemaResponse{
		SchemaID:          schema.SchemaID,
		SchemaName:        schema.SchemaName,
		SchemaDescription: schema.SchemaDescription,
		SDL:               schema.SDL,
		Endpoint:          schema.Endpoint,
		ProviderID:        schema.ProviderID,
		CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateSchema updates an existing schema
func (s *SchemaService) UpdateSchema(schemaID string, req *models.UpdateSchemaRequest) (*models.SchemaResponse, error) {
	var schema models.Schema
	err := s.db.First(&schema, "schema_id = ?", schemaID).Error
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}

	// Update fields if provided
	if req.SchemaName != nil {
		schema.SchemaName = *req.SchemaName
	}
	if req.SchemaDescription != nil {
		schema.SchemaDescription = req.SchemaDescription
	}
	if req.SDL != nil {
		schema.SDL = *req.SDL
	}
	if req.Endpoint != nil {
		schema.Endpoint = *req.Endpoint
	}
	if req.Version != nil {
		schema.Version = *req.Version
	}

	if err := s.db.Save(&schema).Error; err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}

	response := &models.SchemaResponse{
		SchemaID:          schema.SchemaID,
		SchemaName:        schema.SchemaName,
		SchemaDescription: schema.SchemaDescription,
		SDL:               schema.SDL,
		Endpoint:          schema.Endpoint,
		ProviderID:        schema.ProviderID,
		CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetSchema retrieves a schema by ID
func (s *SchemaService) GetSchema(schemaID string) (*models.SchemaResponse, error) {
	var schema models.Schema
	err := s.db.First(&schema, "schema_id = ?", schemaID).Error
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}

	response := &models.SchemaResponse{
		SchemaID:          schema.SchemaID,
		SchemaName:        schema.SchemaName,
		SchemaDescription: schema.SchemaDescription,
		SDL:               schema.SDL,
		Endpoint:          schema.Endpoint,
		ProviderID:        schema.ProviderID,
		CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetSchemas Get all schemas and filter by provider ID if given
func (s *SchemaService) GetSchemas(providerID *string) ([]*models.SchemaResponse, error) {
	var schemas []models.Schema
	query := s.db
	if providerID != nil && *providerID != "" {
		query = query.Where("provider_id = ?", *providerID)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	err := query.Find(&schemas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schemas: %w", err)
	}

	var responses []*models.SchemaResponse
	for _, schema := range schemas {
		responses = append(responses, &models.SchemaResponse{
			SchemaID:          schema.SchemaID,
			SchemaName:        schema.SchemaName,
			SchemaDescription: schema.SchemaDescription,
			SDL:               schema.SDL,
			Endpoint:          schema.Endpoint,
			ProviderID:        schema.ProviderID,
			CreatedAt:         schema.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         schema.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

// CreateSchemaSubmission creates a new schema
func (s *SchemaService) CreateSchemaSubmission(req *models.CreateSchemaSubmissionRequest) (*models.SchemaSubmissionResponse, error) {
	// Check if provider exists
	var provider models.Provider
	if err := s.db.First(&provider, "provider_id = ?", req.ProviderID).Error; err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// If PreviousSchemaID is provided, check if it exists
	if req.PreviousSchemaID != nil {
		var previousSchema models.Schema
		if err := s.db.First(&previousSchema, "schema_id = ?", *req.PreviousSchemaID).Error; err != nil {
			return nil, fmt.Errorf("previous schema not found: %w", err)
		}
	}

	// Create submission
	submission := models.SchemaSubmission{
		SubmissionID:      "sub_" + uuid.New().String(),
		PreviousSchemaID:  req.PreviousSchemaID,
		SchemaName:        req.SchemaName,
		SchemaDescription: req.SchemaDescription,
		SDL:               req.SDL,
		SchemaEndpoint:    req.SchemaEndpoint,
		Status:            models.StatusPending,
		ProviderID:        req.ProviderID,
	}
	if err := s.db.Create(&submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create schema submission: %w", err)
	}

	response := &models.SchemaSubmissionResponse{
		SubmissionID:      submission.SubmissionID,
		PreviousSchemaID:  submission.PreviousSchemaID,
		SchemaName:        submission.SchemaName,
		SchemaDescription: submission.SchemaDescription,
		SDL:               submission.SDL,
		SchemaEndpoint:    submission.SchemaEndpoint,
		Status:            submission.Status,
		ProviderID:        submission.ProviderID,
		CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateSchemaSubmission updates an existing schema submission
func (s *SchemaService) UpdateSchemaSubmission(submissionID string, req *models.UpdateSchemaSubmissionRequest) (*models.SchemaSubmissionResponse, error) {
	var submission models.SchemaSubmission

	// Start transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Find the submission
		if err := tx.First(&submission, "submission_id = ?", submissionID).Error; err != nil {
			return fmt.Errorf("schema submission not found: %w", err)
		}

		// Update fields if provided
		if req.SchemaName != nil {
			submission.SchemaName = *req.SchemaName
		}
		if req.SchemaDescription != nil {
			submission.SchemaDescription = req.SchemaDescription
		}
		if req.SDL != nil {
			submission.SDL = *req.SDL
		}
		if req.SchemaEndpoint != nil {
			submission.SchemaEndpoint = *req.SchemaEndpoint
		}
		// If status is provided and is approved, create a new schema
		if req.Status != nil {
			submission.Status = *req.Status
			if *req.Status == models.StatusApproved {
				newSchema := models.Schema{
					SchemaID:          "sch_" + uuid.New().String(),
					SchemaName:        submission.SchemaName,
					SchemaDescription: submission.SchemaDescription,
					SDL:               submission.SDL,
					Endpoint:          submission.SchemaEndpoint,
					ProviderID:        submission.ProviderID,
				}
				if err := tx.Create(&newSchema).Error; err != nil {
					return fmt.Errorf("failed to create schema: %w", err)
				}
			}
		}
		if req.PreviousSchemaID != nil {
			// Check if the new PreviousSchemaID exists
			var previousSchema models.Schema
			if err := tx.First(&previousSchema, "schema_id = ?", *req.PreviousSchemaID).Error; err != nil {
				return fmt.Errorf("previous schema not found: %w", err)
			}
			submission.PreviousSchemaID = req.PreviousSchemaID
		}

		// Save the updated submission
		if err := tx.Save(&submission).Error; err != nil {
			return fmt.Errorf("failed to update schema submission: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &models.SchemaSubmissionResponse{
		SubmissionID:      submission.SubmissionID,
		PreviousSchemaID:  submission.PreviousSchemaID,
		SchemaName:        submission.SchemaName,
		SchemaDescription: submission.SchemaDescription,
		SDL:               submission.SDL,
		SchemaEndpoint:    submission.SchemaEndpoint,
		Status:            submission.Status,
		ProviderID:        submission.ProviderID,
		CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetSchemaSubmission retrieves a schema submission by ID
func (s *SchemaService) GetSchemaSubmission(submissionID string) (*models.SchemaSubmissionResponse, error) {
	var submission models.SchemaSubmission
	err := s.db.First(&submission, "submission_id = ?", submissionID).Error
	if err != nil {
		return nil, fmt.Errorf("schema submission not found: %w", err)
	}

	response := &models.SchemaSubmissionResponse{
		SubmissionID:      submission.SubmissionID,
		PreviousSchemaID:  submission.PreviousSchemaID,
		SchemaName:        submission.SchemaName,
		SchemaDescription: submission.SchemaDescription,
		SDL:               submission.SDL,
		SchemaEndpoint:    submission.SchemaEndpoint,
		Status:            submission.Status,
		ProviderID:        submission.ProviderID,
		CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetSchemaSubmissions Get all schema submissions and filter by provider ID OR Status Array if given
func (s *SchemaService) GetSchemaSubmissions(providerID *string, statusFilter *[]string) ([]*models.SchemaSubmissionResponse, error) {
	var submissions []models.SchemaSubmission
	query := s.db.Preload("PreviousSchema").Preload("Provider")
	if providerID != nil && *providerID != "" {
		query = query.Where("provider_id = ?", *providerID)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	if statusFilter != nil && len(*statusFilter) > 0 {
		query = query.Where("status IN ?", *statusFilter)
	}

	err := query.Find(&submissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schema submissions: %w", err)
	}

	var responses []*models.SchemaSubmissionResponse
	for _, submission := range submissions {
		responses = append(responses, &models.SchemaSubmissionResponse{
			SubmissionID:      submission.SubmissionID,
			PreviousSchemaID:  submission.PreviousSchemaID,
			SchemaName:        submission.SchemaName,
			SchemaDescription: submission.SchemaDescription,
			SDL:               submission.SDL,
			SchemaEndpoint:    submission.SchemaEndpoint,
			Status:            submission.Status,
			ProviderID:        submission.ProviderID,
			CreatedAt:         submission.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         submission.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}
