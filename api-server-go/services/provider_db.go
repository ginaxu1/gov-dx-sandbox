package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ProviderServiceDB struct {
	db              *sql.DB
	schemaConverter *SchemaConverter
}

// Provider Submission methods

func (s *ProviderServiceDB) GetAllProviderSubmissions() ([]*models.ProviderSubmission, error) {
	query := `SELECT submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at 
			  FROM provider_submissions ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider submissions: %w", err)
	}
	defer rows.Close()

	submissions := make([]*models.ProviderSubmission, 0)
	for rows.Next() {
		submission := &models.ProviderSubmission{}
		var statusStr string

		err := rows.Scan(&submission.SubmissionID, &submission.ProviderName, &submission.ContactEmail,
			&submission.PhoneNumber, &submission.ProviderType, &statusStr, &submission.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider submission: %w", err)
		}

		submission.Status = models.ProviderSubmissionStatus(statusStr)
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

func (s *ProviderServiceDB) CreateProviderSubmission(req models.CreateProviderSubmissionRequest) (*models.ProviderSubmission, error) {
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
	checkQuery := `SELECT COUNT(*) FROM provider_submissions WHERE provider_name = $1 AND status = 'pending'`
	var count int
	err := s.db.QueryRow(checkQuery, req.ProviderName).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate submissions: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("a pending submission for '%s' already exists", req.ProviderName)
	}

	// Generate unique submission ID
	submissionID, err := s.generateSubmissionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate submission ID: %w", err)
	}

	submission := &models.ProviderSubmission{
		SubmissionID: submissionID,
		ProviderName: req.ProviderName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		ProviderType: req.ProviderType,
		Status:       models.SubmissionStatusPending,
		CreatedAt:    time.Now(),
	}

	query := `INSERT INTO provider_submissions (submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(query, submission.SubmissionID, submission.ProviderName, submission.ContactEmail,
		submission.PhoneNumber, string(submission.ProviderType), string(submission.Status),
		submission.CreatedAt, submission.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider submission: %w", err)
	}

	slog.Info("Created new provider submission", "submissionId", submissionID)
	return submission, nil
}

func (s *ProviderServiceDB) GetProviderSubmission(id string) (*models.ProviderSubmission, error) {
	query := `SELECT submission_id, provider_name, contact_email, phone_number, provider_type, status, created_at 
			  FROM provider_submissions WHERE submission_id = $1`

	row := s.db.QueryRow(query, id)

	submission := &models.ProviderSubmission{}
	var statusStr string

	err := row.Scan(&submission.SubmissionID, &submission.ProviderName, &submission.ContactEmail,
		&submission.PhoneNumber, &submission.ProviderType, &statusStr, &submission.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider submission not found")
		}
		return nil, fmt.Errorf("failed to get provider submission: %w", err)
	}

	submission.Status = models.ProviderSubmissionStatus(statusStr)
	return submission, nil
}

func (s *ProviderServiceDB) UpdateProviderSubmission(id string, req models.UpdateProviderSubmissionRequest) (*models.UpdateProviderSubmissionResponse, error) {
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

			// Insert provider profile
			profileQuery := `INSERT INTO provider_profiles (provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at) 
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

			_, err = s.db.Exec(profileQuery, profile.ProviderID, profile.ProviderName, profile.ContactEmail,
				profile.PhoneNumber, string(profile.ProviderType), profile.ApprovedAt, profile.ApprovedAt, profile.ApprovedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to create provider profile: %w", err)
			}

			response.ProviderID = profile.ProviderID
		}

		// Update submission
		updateQuery := `UPDATE provider_submissions SET status = $1, updated_at = $2 WHERE submission_id = $3`
		_, err = s.db.Exec(updateQuery, string(submission.Status), submission.UpdatedAt, id)
		if err != nil {
			return nil, fmt.Errorf("failed to update provider submission: %w", err)
		}
	}

	slog.Info("Updated provider submission", "submissionId", id, "status", submission.Status)
	return response, nil
}

// Provider Profile methods

func (s *ProviderServiceDB) GetAllProviderProfiles() ([]*models.ProviderProfile, error) {
	query := `SELECT provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at 
			  FROM provider_profiles ORDER BY approved_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider profiles: %w", err)
	}
	defer rows.Close()

	profiles := make([]*models.ProviderProfile, 0)
	for rows.Next() {
		profile := &models.ProviderProfile{}

		err := rows.Scan(&profile.ProviderID, &profile.ProviderName, &profile.ContactEmail,
			&profile.PhoneNumber, &profile.ProviderType, &profile.ApprovedAt, &profile.CreatedAt, &profile.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider profile: %w", err)
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (s *ProviderServiceDB) GetProviderProfile(id string) (*models.ProviderProfile, error) {
	query := `SELECT provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at 
			  FROM provider_profiles WHERE provider_id = $1`

	row := s.db.QueryRow(query, id)

	profile := &models.ProviderProfile{}

	err := row.Scan(&profile.ProviderID, &profile.ProviderName, &profile.ContactEmail,
		&profile.PhoneNumber, &profile.ProviderType, &profile.ApprovedAt, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider profile not found")
		}
		return nil, fmt.Errorf("failed to get provider profile: %w", err)
	}

	return profile, nil
}

// Provider Schema methods

func (s *ProviderServiceDB) GetAllProviderSchemas() ([]*models.ProviderSchema, error) {
	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider schemas: %w", err)
	}
	defer rows.Close()

	schemas := make([]*models.ProviderSchema, 0)
	for rows.Next() {
		schema, err := s.scanProviderSchema(rows)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (s *ProviderServiceDB) CreateProviderSchema(req models.CreateProviderSchemaRequest) (*models.ProviderSchema, error) {
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

	schema := &models.ProviderSchema{
		SubmissionID:        schemaID,
		ProviderID:          req.ProviderID,
		Status:              models.SchemaStatusPending,
		SchemaInput:         req.SchemaInput,
		FieldConfigurations: req.FieldConfigurations,
	}

	// Convert fields to JSON
	schemaInputJSON, err := json.Marshal(schema.SchemaInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema input: %w", err)
	}

	fieldConfigurationsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field configurations: %w", err)
	}

	query := `INSERT INTO provider_schemas (submission_id, provider_id, status, schema_input, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(query, schema.SubmissionID, schema.ProviderID, string(schema.Status),
		schemaInputJSON, fieldConfigurationsJSON, schema.CreatedAt, schema.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider schema: %w", err)
	}

	slog.Info("Created new provider schema", "submissionId", schemaID)
	return schema, nil
}

func (s *ProviderServiceDB) CreateProviderSchemaSDL(providerID string, req models.CreateProviderSchemaSDLRequest) (*models.ProviderSchema, error) {
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

	schema := &models.ProviderSchema{
		SubmissionID:        schemaID,
		ProviderID:          providerID,
		Status:              models.SchemaStatusPending,
		SDL:                 req.SDL,
		FieldConfigurations: make(models.FieldConfigurations),
	}

	fieldConfigurationsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field configurations: %w", err)
	}

	query := `INSERT INTO provider_schemas (submission_id, provider_id, status, sdl, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(query, schema.SubmissionID, schema.ProviderID, string(schema.Status),
		schema.SDL, fieldConfigurationsJSON, schema.CreatedAt, schema.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider schema: %w", err)
	}

	slog.Info("Created new provider schema from SDL", "submissionId", schemaID, "providerId", providerID)
	return schema, nil
}

func (s *ProviderServiceDB) CreateProviderSchemaSubmission(providerID string, req models.CreateProviderSchemaSubmissionRequest) (*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	// If schema_id is provided, this is a modification of existing schema
	if req.SchemaID != nil && *req.SchemaID != "" {
		// Verify the existing schema belongs to this provider
		existingSchema, err := s.GetProviderSchema(*req.SchemaID)
		if err != nil {
			return nil, fmt.Errorf("schema not found: %w", err)
		}
		if existingSchema.ProviderID != providerID {
			return nil, fmt.Errorf("schema does not belong to this provider")
		}

		// Create a new submission for the modification
		schemaID, err := s.generateSchemaID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema ID: %w", err)
		}

		schema := &models.ProviderSchema{
			SubmissionID:        schemaID,
			ProviderID:          providerID,
			Status:              models.SchemaStatusDraft,
			SDL:                 req.SDL,
			FieldConfigurations: make(models.FieldConfigurations),
		}

		fieldConfigurationsJSON, err := json.Marshal(schema.FieldConfigurations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field configurations: %w", err)
		}

		query := `INSERT INTO provider_schemas (submission_id, provider_id, status, sdl, field_configurations, created_at, updated_at) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err = s.db.Exec(query, schema.SubmissionID, schema.ProviderID, string(schema.Status),
			schema.SDL, fieldConfigurationsJSON, schema.CreatedAt, schema.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create schema modification: %w", err)
		}

		slog.Info("Created schema modification submission", "submissionId", schemaID, "providerId", providerID, "originalSchemaId", *req.SchemaID)
		return schema, nil
	}

	// This is a new schema submission
	schemaID, err := s.generateSchemaID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema ID: %w", err)
	}

	schema := &models.ProviderSchema{
		SubmissionID:        schemaID,
		ProviderID:          providerID,
		Status:              models.SchemaStatusDraft,
		SDL:                 req.SDL,
		FieldConfigurations: make(models.FieldConfigurations),
	}

	fieldConfigurationsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field configurations: %w", err)
	}

	query := `INSERT INTO provider_schemas (submission_id, provider_id, status, sdl, field_configurations, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(query, schema.SubmissionID, schema.ProviderID, string(schema.Status),
		schema.SDL, fieldConfigurationsJSON, schema.CreatedAt, schema.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider schema: %w", err)
	}

	slog.Info("Created new schema submission", "submissionId", schemaID, "providerId", providerID)
	return schema, nil
}

func (s *ProviderServiceDB) SubmitSchemaForReview(schemaID string) (*models.ProviderSchema, error) {
	// Get existing schema
	schema, err := s.GetProviderSchema(schemaID)
	if err != nil {
		return nil, err
	}

	if schema.Status != models.SchemaStatusDraft {
		return nil, fmt.Errorf("only draft schemas can be submitted for review")
	}

	schema.Status = models.SchemaStatusPending
	schema.UpdatedAt = time.Now()

	query := `UPDATE provider_schemas SET status = $1, updated_at = $2 WHERE submission_id = $3`

	_, err = s.db.Exec(query, string(schema.Status), schema.UpdatedAt, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to update schema status: %w", err)
	}

	slog.Info("Schema submitted for review", "submissionId", schemaID, "providerId", schema.ProviderID)
	return schema, nil
}

func (s *ProviderServiceDB) GetApprovedSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error) {
	// Verify provider exists
	_, err := s.GetProviderProfile(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider profile not found: %w", err)
	}

	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas WHERE provider_id = $1 AND status = 'approved' AND schema_id IS NOT NULL 
			  ORDER BY created_at DESC`

	rows, err := s.db.Query(query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved schemas: %w", err)
	}
	defer rows.Close()

	var approvedSchemas []*models.ProviderSchema
	for rows.Next() {
		schema, err := s.scanProviderSchema(rows)
		if err != nil {
			return nil, err
		}
		approvedSchemas = append(approvedSchemas, schema)
	}

	return approvedSchemas, nil
}

func (s *ProviderServiceDB) GetProviderSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error) {
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

	schemas := make([]*models.ProviderSchema, 0)
	for rows.Next() {
		schema, err := s.scanProviderSchema(rows)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func (s *ProviderServiceDB) GetProviderSchema(id string) (*models.ProviderSchema, error) {
	query := `SELECT submission_id, provider_id, schema_id, status, schema_input, sdl, field_configurations, created_at, updated_at 
			  FROM provider_schemas WHERE submission_id = $1`

	row := s.db.QueryRow(query, id)

	schema, err := s.scanProviderSchemaFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider schema not found")
		}
		return nil, err
	}

	return schema, nil
}

func (s *ProviderServiceDB) UpdateProviderSchema(id string, req models.UpdateProviderSchemaRequest) (*models.ProviderSchema, error) {
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
		schema.UpdatedAt = time.Now()

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

	// Convert fields to JSON
	fieldConfigurationsJSON, err := json.Marshal(schema.FieldConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal field configurations: %w", err)
	}

	query := `UPDATE provider_schemas SET status = $1, schema_id = $2, field_configurations = $3, updated_at = $4 
			  WHERE submission_id = $5`

	_, err = s.db.Exec(query, string(schema.Status), schema.SchemaID, fieldConfigurationsJSON,
		schema.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider schema: %w", err)
	}

	slog.Info("Updated provider schema", "submissionId", id, "status", schema.Status, "schemaId", schema.SchemaID)
	return schema, nil
}

// Helper methods

func (s *ProviderServiceDB) generateSubmissionID() (string, error) {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("sub_prov_%d", timestamp), nil
}

func (s *ProviderServiceDB) generateSchemaID() (string, error) {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("schema_%d", timestamp), nil
}

func (s *ProviderServiceDB) generateProviderID() (string, error) {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("prov_%d", timestamp), nil
}

func (s *ProviderServiceDB) createProviderProfile(submission *models.ProviderSubmission) (*models.ProviderProfile, error) {
	providerID, err := s.generateProviderID()
	if err != nil {
		return nil, err
	}

	profile := &models.ProviderProfile{
		ProviderID:   providerID,
		ProviderName: submission.ProviderName,
		ContactEmail: submission.ContactEmail,
		PhoneNumber:  submission.PhoneNumber,
		ProviderType: submission.ProviderType,
		ApprovedAt:   time.Now(),
	}

	return profile, nil
}

func (s *ProviderServiceDB) CreateProviderProfileForTesting(providerName, contactEmail, phoneNumber, providerType string) (*models.ProviderProfile, error) {
	providerID, err := s.generateProviderID()
	if err != nil {
		return nil, err
	}

	profile := &models.ProviderProfile{
		ProviderID:   providerID,
		ProviderName: providerName,
		ContactEmail: contactEmail,
		PhoneNumber:  phoneNumber,
		ProviderType: models.ProviderType(providerType),
		ApprovedAt:   time.Now(),
	}

	query := `INSERT INTO provider_profiles (provider_id, provider_name, contact_email, phone_number, provider_type, approved_at, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(query, profile.ProviderID, profile.ProviderName, profile.ContactEmail,
		profile.PhoneNumber, string(profile.ProviderType), profile.ApprovedAt, profile.ApprovedAt, profile.ApprovedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider profile: %w", err)
	}

	return profile, nil
}

// Helper function to scan provider schema from rows
func (s *ProviderServiceDB) scanProviderSchema(rows *sql.Rows) (*models.ProviderSchema, error) {
	schema := &models.ProviderSchema{}
	var statusStr string
	var schemaID sql.NullString
	var schemaInputJSON, sdlJSON, fieldConfigurationsJSON sql.NullString

	err := rows.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &statusStr,
		&schemaInputJSON, &sdlJSON, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan provider schema: %w", err)
	}

	schema.Status = models.ProviderSchemaStatus(statusStr)
	if schemaID.Valid {
		schema.SchemaID = &schemaID.String
	}

	// Parse schema input JSON
	if schemaInputJSON.Valid {
		err = json.Unmarshal([]byte(schemaInputJSON.String), &schema.SchemaInput)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema input: %w", err)
		}
	}

	// Parse SDL
	if sdlJSON.Valid {
		schema.SDL = sdlJSON.String
	}

	// Parse field configurations JSON
	if fieldConfigurationsJSON.Valid {
		err = json.Unmarshal([]byte(fieldConfigurationsJSON.String), &schema.FieldConfigurations)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal field configurations: %w", err)
		}
	}

	return schema, nil
}

// Helper function to scan provider schema from a single row
func (s *ProviderServiceDB) scanProviderSchemaFromRow(row *sql.Row) (*models.ProviderSchema, error) {
	schema := &models.ProviderSchema{}
	var statusStr string
	var schemaID sql.NullString
	var schemaInputJSON, sdlJSON, fieldConfigurationsJSON sql.NullString

	err := row.Scan(&schema.SubmissionID, &schema.ProviderID, &schemaID, &statusStr,
		&schemaInputJSON, &sdlJSON, &fieldConfigurationsJSON, &schema.CreatedAt, &schema.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan provider schema: %w", err)
	}

	schema.Status = models.ProviderSchemaStatus(statusStr)
	if schemaID.Valid {
		schema.SchemaID = &schemaID.String
	}

	// Parse schema input JSON
	if schemaInputJSON.Valid {
		err = json.Unmarshal([]byte(schemaInputJSON.String), &schema.SchemaInput)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal schema input: %w", err)
		}
	}

	// Parse SDL
	if sdlJSON.Valid {
		schema.SDL = sdlJSON.String
	}

	// Parse field configurations JSON
	if fieldConfigurationsJSON.Valid {
		err = json.Unmarshal([]byte(fieldConfigurationsJSON.String), &schema.FieldConfigurations)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal field configurations: %w", err)
		}
	}

	return schema, nil
}
