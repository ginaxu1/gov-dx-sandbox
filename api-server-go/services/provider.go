package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ProviderService struct {
	db              *sql.DB
	schemaConverter *SchemaConverter
}

func NewProviderService(db *sql.DB) *ProviderService {
	return &ProviderService{
		db:              db,
		schemaConverter: NewSchemaConverter(),
	}
}

// parseSchemaJSONFields parses JSON fields from database into the schema struct
func (s *ProviderService) parseSchemaJSONFields(schema *models.ProviderSchema, schemaInputJSON, fieldConfigurationsJSON sql.NullString) {
	// Parse schema input JSON
	if schemaInputJSON.Valid && schemaInputJSON.String != "" {
		if err := json.Unmarshal([]byte(schemaInputJSON.String), &schema.SchemaInput); err != nil {
			slog.Warn("Failed to parse schema input JSON", "error", err, "schemaId", schema.SchemaID)
		}
	}

	// Parse field configurations JSON
	if fieldConfigurationsJSON.Valid && fieldConfigurationsJSON.String != "" {
		if err := json.Unmarshal([]byte(fieldConfigurationsJSON.String), &schema.FieldConfigurations); err != nil {
			slog.Warn("Failed to parse field configurations JSON", "error", err, "schemaId", schema.SchemaID)
		}
	}
}

// validateDBConnection checks if the database connection is valid
func (s *ProviderService) validateDBConnection() error {
	if s.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

// Provider Submission methods
func (s *ProviderService) GetAllProviderSubmissions() ([]*models.ProviderSubmission, error) {
	return s.GetProviderSubmissionsByStatus("")
}

func (s *ProviderService) GetProviderSubmissionsByStatus(status string) ([]*models.ProviderSubmission, error) {
	slog.Info("Starting retrieval of provider submissions", "status_filter", status)

	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		slog.Error("Database connection validation failed", "error", err)
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	query := `SELECT submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at, updated_at 
			  FROM provider_submissions`
	args := []interface{}{}

	// Add status filter if provided
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing database query", "query", query, "args", args)
	start := time.Now()
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("Database query failed", "error", err, "query", query, "args", args, "duration", time.Since(start))
		return nil, fmt.Errorf("failed to get provider submissions: %w", err)
	}
	defer rows.Close()

	var submissions []*models.ProviderSubmission
	rowCount := 0
	for rows.Next() {
		submission := &models.ProviderSubmission{}
		err := rows.Scan(&submission.SubmissionID, &submission.ProviderName, &submission.ContactEmail, &submission.PhoneNumber, &submission.ProviderType, &submission.Status, &submission.CreatedAt, &submission.UpdatedAt)
		if err != nil {
			slog.Error("Failed to scan provider submission row", "error", err, "rowCount", rowCount)
			return nil, fmt.Errorf("failed to scan provider submission: %w", err)
		}
		submissions = append(submissions, submission)
		rowCount++
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		slog.Error("Error during row iteration", "error", err, "rowCount", rowCount)
		return nil, fmt.Errorf("failed to iterate provider submissions: %w", err)
	}

	duration := time.Since(start)
	slog.Info("Successfully retrieved all provider submissions", "count", len(submissions), "duration", duration)
	return submissions, nil
}

func (s *ProviderService) CreateProviderSubmission(req models.CreateProviderSubmissionRequest) (*models.ProviderSubmission, error) {
	// Validate required fields
	if req.ProviderName == "" {
		return nil, fmt.Errorf("providerName is required")
	}
	if req.ContactEmail == "" {
		return nil, fmt.Errorf("contactEmail is required")
	}
	if req.PhoneNumber == "" {
		return nil, fmt.Errorf("phoneNumber is required")
	}
	if req.ProviderType == "" {
		return nil, fmt.Errorf("providerType is required")
	}

	// Check for duplicate pending submission
	query := `SELECT COUNT(*) FROM provider_submissions WHERE provider_name = $1 AND status = $2`
	var count int
	err := s.db.QueryRow(query, req.ProviderName, models.SubmissionStatusPending).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate submission: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("a pending submission for '%s' already exists", req.ProviderName)
	}

	// Generate unique submission ID
	submissionID, err := s.generateSubmissionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate submission ID: %w", err)
	}

	now := time.Now()
	submission := &models.ProviderSubmission{
		SubmissionID: submissionID,
		ProviderName: req.ProviderName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		ProviderType: req.ProviderType,
		Status:       models.SubmissionStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query = `INSERT INTO provider_submissions (submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at, updated_at) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(query, submission.SubmissionID, submission.ProviderName, submission.ContactEmail, submission.PhoneNumber, submission.ProviderType, submission.Status, submission.CreatedAt, submission.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider submission: %w", err)
	}

	slog.Info("Created new provider submission", "submissionId", submissionID)
	return submission, nil
}

func (s *ProviderService) GetProviderSubmission(id string) (*models.ProviderSubmission, error) {
	query := `SELECT submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at, updated_at 
			  FROM provider_submissions WHERE submission_id = $1`

	row := s.db.QueryRow(query, id)

	submission := &models.ProviderSubmission{}
	err := row.Scan(&submission.SubmissionID, &submission.ProviderName, &submission.ContactEmail, &submission.PhoneNumber, &submission.ProviderType, &submission.Status, &submission.CreatedAt, &submission.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider submission not found")
		}
		return nil, fmt.Errorf("failed to get provider submission: %w", err)
	}

	return submission, nil
}

func (s *ProviderService) UpdateProviderSubmission(id string, req models.UpdateProviderSubmissionRequest) (*models.UpdateProviderSubmissionResponse, error) {
	// First get the existing submission
	submission, err := s.GetProviderSubmission(id)
	if err != nil {
		return nil, err
	}

	response := &models.UpdateProviderSubmissionResponse{
		ProviderSubmission: submission,
	}

	// Update status if provided
	if req.Status != nil {
		submission.Status = *req.Status
		submission.UpdatedAt = time.Now()

		// If approved, create provider profile
		if *req.Status == models.SubmissionStatusApproved {
			profile, err := s.createProviderProfile(submission)
			if err != nil {
				return nil, fmt.Errorf("failed to create provider profile: %w", err)
			}
			response.ProviderID = profile.ProviderID
		}

		// Update the submission
		query := `UPDATE provider_submissions SET status = $1, updated_at = $2 WHERE submission_id = $3`
		_, err = s.db.Exec(query, submission.Status, submission.UpdatedAt, id)
		if err != nil {
			return nil, fmt.Errorf("failed to update provider submission: %w", err)
		}
	}

	slog.Info("Updated provider submission", "submissionId", id, "status", submission.Status)
	return response, nil
}

// Provider Profile methods
func (s *ProviderService) GetAllProviderProfiles() ([]*models.ProviderProfile, error) {
	query := `SELECT provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at 
			  FROM provider_profiles ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*models.ProviderProfile
	for rows.Next() {
		profile := &models.ProviderProfile{}
		err := rows.Scan(&profile.ProviderID, &profile.ProviderName, &profile.ContactEmail, &profile.PhoneNumber, &profile.ProviderType, &profile.ApprovedAt, &profile.CreatedAt, &profile.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (s *ProviderService) GetProviderProfile(id string) (*models.ProviderProfile, error) {
	query := `SELECT provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at 
			  FROM provider_profiles WHERE provider_id = $1`

	row := s.db.QueryRow(query, id)

	profile := &models.ProviderProfile{}
	err := row.Scan(&profile.ProviderID, &profile.ProviderName, &profile.ContactEmail, &profile.PhoneNumber, &profile.ProviderType, &profile.ApprovedAt, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider profile not found")
		}
		return nil, fmt.Errorf("failed to get provider profile: %w", err)
	}

	return profile, nil
}

// Provider Schema methods
func (s *ProviderService) GetAllProviderSchemas() ([]*models.ProviderSchema, error) {
	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*models.ProviderSchema
	for rows.Next() {
		schema := &models.ProviderSchema{}
		var schemaID sql.NullString
		var schemaInputJSON sql.NullString
		var sdl sql.NullString
		var fieldConfigurationsJSON sql.NullString

		err := rows.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &schema.Status, &schemaInputJSON, &sdl, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider schema: %w", err)
		}

		if schemaID.Valid {
			schema.SchemaID = &schemaID.String
		}
		if sdl.Valid {
			schema.SDL = sdl.String
		}
		// Parse JSON fields
		s.parseSchemaJSONFields(schema, schemaInputJSON, fieldConfigurationsJSON)

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (s *ProviderService) CreateProviderSchema(req models.CreateProviderSchemaRequest) (*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	// Generate unique schema submission ID
	schemaID, err := s.generateSchemaID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema ID: %w", err)
	}

	now := time.Now()
	schema := &models.ProviderSchema{
		SchemaID:            &schemaID,
		ProviderID:          req.ProviderID,
		Status:              models.SchemaStatusPending,
		SchemaInput:         req.SchemaInput,
		FieldConfigurations: req.FieldConfigurations,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	query := `INSERT INTO provider_schemas (schema_id, provider_id, status, schema_input, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Serialize JSON fields
	schemaInputJSON, err := json.Marshal(schema.SchemaInput)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize schema input: %w", err)
	}

	fieldConfigsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize field configurations: %w", err)
	}

	_, err = s.db.Exec(query, *schema.SchemaID, schema.ProviderID, schema.Status, string(schemaInputJSON), string(fieldConfigsJSON), schema.CreatedAt, schema.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider schema: %w", err)
	}

	slog.Info("Created new provider schema", "submissionId", schemaID)
	return schema, nil
}

// CreateProviderSchemaSDL creates a new provider schema from SDL
func (s *ProviderService) CreateProviderSchemaSDL(providerID string, req models.CreateProviderSchemaSDLRequest) (*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	// Generate unique schema submission ID
	schemaID, err := s.generateSchemaID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema ID: %w", err)
	}

	now := time.Now()
	schema := &models.ProviderSchema{
		SchemaID:            &schemaID,
		ProviderID:          providerID,
		Status:              models.SchemaStatusPending,
		SDL:                 req.SDL,
		FieldConfigurations: make(models.FieldConfigurations),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	query := `INSERT INTO provider_schemas (schema_id, provider_id, status, sdl, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Serialize field configurations to JSON
	fieldConfigsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize field configurations: %w", err)
	}

	_, err = s.db.Exec(query, *schema.SchemaID, schema.ProviderID, schema.Status, schema.SDL, string(fieldConfigsJSON), schema.CreatedAt, schema.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider schema from SDL: %w", err)
	}

	slog.Info("Created new provider schema from SDL", "submissionId", schemaID, "providerId", providerID)
	return schema, nil
}

// CreateProviderSchemaSubmission creates a new schema submission or modifies an existing one
func (s *ProviderService) CreateProviderSchemaSubmission(providerID string, req models.CreateProviderSchemaSubmissionRequest) (*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	// If schema_id is provided, this is a modification of existing schema
	if req.SchemaID != nil && *req.SchemaID != "" {
		// Verify the existing schema belongs to this provider
		query := `SELECT provider_id FROM provider_schemas WHERE schema_id = $1`
		var existingProviderID string
		err := s.db.QueryRow(query, *req.SchemaID).Scan(&existingProviderID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("schema not found")
			}
			return nil, fmt.Errorf("failed to check schema: %w", err)
		}
		if existingProviderID != providerID {
			return nil, fmt.Errorf("schema does not belong to this provider")
		}
	}

	// Create a new submission for the modification
	schemaID, err := s.generateSchemaID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema ID: %w", err)
	}

	now := time.Now()
	schema := &models.ProviderSchema{
		SchemaID:            &schemaID,
		ProviderID:          providerID,
		Status:              models.SchemaStatusDraft,
		SDL:                 req.SDL,
		FieldConfigurations: make(models.FieldConfigurations),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	query := `INSERT INTO provider_schemas (schema_id, provider_id, status, sdl, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Serialize field configurations to JSON
	fieldConfigsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize field configurations: %w", err)
	}

	_, err = s.db.Exec(query, *schema.SchemaID, schema.ProviderID, schema.Status, schema.SDL, string(fieldConfigsJSON), schema.CreatedAt, schema.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema submission: %w", err)
	}

	slog.Info("Created schema submission", "submissionId", schemaID, "providerId", providerID)
	return schema, nil
}

// SubmitSchemaForReview changes schema status from draft to pending for admin review
func (s *ProviderService) SubmitSchemaForReview(schemaID string) (*models.ProviderSchema, error) {
	query := `UPDATE provider_schemas SET status = $1, updated_at = $2 WHERE submission_id = $3 AND status = $4`

	result, err := s.db.Exec(query, models.SchemaStatusPending, time.Now(), schemaID, models.SchemaStatusDraft)
	if err != nil {
		return nil, fmt.Errorf("failed to submit schema for review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("only draft schemas can be submitted for review")
	}

	// Get the updated schema
	schema, err := s.GetProviderSchema(schemaID)
	if err != nil {
		return nil, err
	}

	slog.Info("Schema submitted for review", "submissionId", schemaID, "providerId", schema.ProviderID)
	return schema, nil
}

// GetApprovedSchemasByProviderID gets all approved schemas for a specific provider
func (s *ProviderService) GetApprovedSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas WHERE provider_id = $1 AND status = $2 AND schema_id IS NOT NULL`

	rows, err := s.db.Query(query, providerID, models.SchemaStatusApproved)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved schemas: %w", err)
	}
	defer rows.Close()

	var approvedSchemas []*models.ProviderSchema
	for rows.Next() {
		schema := &models.ProviderSchema{}
		var schemaID sql.NullString
		var schemaInputJSON sql.NullString
		var sdl sql.NullString
		var fieldConfigurationsJSON sql.NullString

		err := rows.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &schema.Status, &schemaInputJSON, &sdl, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider schema: %w", err)
		}

		if schemaID.Valid {
			schema.SchemaID = &schemaID.String
		}
		if sdl.Valid {
			schema.SDL = sdl.String
		}
		// Parse JSON fields
		s.parseSchemaJSONFields(schema, schemaInputJSON, fieldConfigurationsJSON)

		approvedSchemas = append(approvedSchemas, schema)
	}

	return approvedSchemas, nil
}

// GetProviderSchemasByProviderID gets all schema submissions for a specific provider
func (s *ProviderService) GetProviderSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas WHERE provider_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*models.ProviderSchema
	for rows.Next() {
		schema := &models.ProviderSchema{}
		var schemaID sql.NullString
		var schemaInputJSON sql.NullString
		var sdl sql.NullString
		var fieldConfigurationsJSON sql.NullString

		err := rows.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &schema.Status, &schemaInputJSON, &sdl, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider schema: %w", err)
		}

		if schemaID.Valid {
			schema.SchemaID = &schemaID.String
		}
		if sdl.Valid {
			schema.SDL = sdl.String
		}
		// Parse JSON fields
		s.parseSchemaJSONFields(schema, schemaInputJSON, fieldConfigurationsJSON)

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (s *ProviderService) GetProviderSchema(id string) (*models.ProviderSchema, error) {
	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas WHERE submission_id = $1`

	row := s.db.QueryRow(query, id)

	schema := &models.ProviderSchema{}
	var schemaID sql.NullString
	var schemaInputJSON sql.NullString
	var sdl sql.NullString
	var fieldConfigurationsJSON sql.NullString

	err := row.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &schema.Status, &schemaInputJSON, &sdl, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider schema not found")
		}
		return nil, fmt.Errorf("failed to get provider schema: %w", err)
	}

	if schemaID.Valid {
		schema.SchemaID = &schemaID.String
	}
	if sdl.Valid {
		schema.SDL = sdl.String
	}
	// TODO: Parse JSON fields

	return schema, nil
}

func (s *ProviderService) UpdateProviderSchema(id string, req models.UpdateProviderSchemaRequest) (*models.ProviderSchema, error) {
	// First get the existing schema
	schema, err := s.GetProviderSchema(id)
	if err != nil {
		return nil, err
	}

	// Validate status transitions
	if req.Status != nil {
		// Only allow admin to approve pending schemas
		if *req.Status == models.SchemaStatusApproved && schema.Status != models.SchemaStatusPending {
			return nil, fmt.Errorf("only pending schemas can be approved")
		}

		// Only allow admin to reject pending schemas
		if *req.Status == models.SchemaStatusRejected && schema.Status != models.SchemaStatusPending {
			return nil, fmt.Errorf("only pending schemas can be rejected")
		}

		schema.Status = *req.Status

		// If schema is approved, set schema_id and update provider-metadata.json
		if *req.Status == models.SchemaStatusApproved {
			// Generate a schema_id for approved schemas
			schemaID, err := s.generateSchemaID()
			if err != nil {
				return nil, fmt.Errorf("failed to generate schema ID: %w", err)
			}
			schema.SchemaID = &schemaID

			// Update provider-metadata.json
			if schema.SDL != "" {
				if err := s.schemaConverter.UpdateProviderMetadataFile(schema.ProviderID, schema.SDL); err != nil {
					slog.Error("Failed to update provider-metadata.json", "error", err, "providerId", schema.ProviderID)
					// Don't fail the update, just log the error
				} else {
					slog.Info("Updated provider-metadata.json from approved schema", "providerId", schema.ProviderID, "schemaId", schemaID)
				}
			}
		}
	}

	if req.FieldConfigurations != nil {
		schema.FieldConfigurations = req.FieldConfigurations
	}

	schema.UpdatedAt = time.Now()

	// Update the schema
	query := `UPDATE provider_schemas SET status = $1, schema_id = $2, field_configurations = $3, updated_at = $4 WHERE submission_id = $5`

	// Serialize field configurations to JSON
	fieldConfigsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize field configurations: %w", err)
	}

	_, err = s.db.Exec(query, schema.Status, *schema.SchemaID, string(fieldConfigsJSON), schema.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider schema: %w", err)
	}

	slog.Info("Updated provider schema", "submissionId", id, "status", schema.Status, "schemaId", *schema.SchemaID)
	return schema, nil
}

// Helper methods
func (s *ProviderService) generateSubmissionID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "sub_prov_" + id.String(), nil
}

func (s *ProviderService) generateSchemaID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "schema_" + id.String(), nil
}

func (s *ProviderService) generateProviderID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "prov_" + id.String(), nil
}

func (s *ProviderService) createProviderProfile(submission *models.ProviderSubmission) (*models.ProviderProfile, error) {
	providerID, err := s.generateProviderID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	profile := &models.ProviderProfile{
		ProviderID:   providerID,
		ProviderName: submission.ProviderName,
		ContactEmail: submission.ContactEmail,
		PhoneNumber:  submission.PhoneNumber,
		ProviderType: submission.ProviderType,
		ApprovedAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `INSERT INTO provider_profiles (provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(query, profile.ProviderID, profile.ProviderName, profile.ContactEmail, profile.PhoneNumber, profile.ProviderType, profile.ApprovedAt, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider profile: %w", err)
	}

	return profile, nil
}

// CreateProviderProfileForTesting creates a provider profile directly for testing purposes
func (s *ProviderService) CreateProviderProfileForTesting(providerName, contactEmail, phoneNumber, providerType string) (*models.ProviderProfile, error) {
	providerID, err := s.generateProviderID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	profile := &models.ProviderProfile{
		ProviderID:   providerID,
		ProviderName: providerName,
		ContactEmail: contactEmail,
		PhoneNumber:  phoneNumber,
		ProviderType: models.ProviderType(providerType),
		ApprovedAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `INSERT INTO provider_profiles (provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(query, profile.ProviderID, profile.ProviderName, profile.ContactEmail, profile.PhoneNumber, profile.ProviderType, profile.ApprovedAt, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider profile: %w", err)
	}

	return profile, nil
}
