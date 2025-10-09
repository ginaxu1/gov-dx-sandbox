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
	"github.com/gov-dx-sandbox/api-server-go/pkg/errors"
)

type ConsumerService struct {
	db        *sql.DB
	pdpClient *PDPClient
}

func NewConsumerService(db *sql.DB, pdpClient *PDPClient) *ConsumerService {
	return &ConsumerService{
		db:        db,
		pdpClient: pdpClient,
	}
}

// validateDBConnection checks if the database connection is valid
func (s *ConsumerService) validateDBConnection() error {
	if s.db == nil {
		return errors.InternalError("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return errors.HandleDatabaseError(err, "connection validation")
	}
	return nil
}

// CreateConsumer creates a new consumer
func (s *ConsumerService) CreateConsumer(req models.CreateConsumerRequest) (*models.Consumer, error) {
	slog.Info("Starting consumer creation", "consumerName", req.ConsumerName, "contactEmail", req.ContactEmail)

	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		slog.Error("Database connection validation failed", "error", err)
		return nil, errors.InternalErrorWithCause("database connection validation failed", err)
	}

	// Generate unique IDs
	consumerID, err := s.generateConsumerID()
	if err != nil {
		slog.Error("Failed to generate consumer ID", "error", err)
		return nil, errors.InternalErrorWithCause("failed to generate consumer ID", err)
	}

	entityID, err := s.generateEntityID()
	if err != nil {
		slog.Error("Failed to generate entity ID", "error", err)
		return nil, errors.InternalErrorWithCause("failed to generate entity ID", err)
	}

	now := time.Now()

	// Start transaction
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return nil, errors.HandleDatabaseError(err, "begin transaction")
	}
	defer tx.Rollback()

	// Insert entity first
	entityQuery := `INSERT INTO entities (entity_id, entity_name, contact_email, phone_number, entity_type, created_at, updated_at) 
					VALUES ($1, $2, $3, $4, $5, $6, $7)`

	slog.Debug("Executing entity insert", "entityId", entityID)
	_, err = tx.ExecContext(context.Background(), entityQuery, entityID, req.ConsumerName, req.ContactEmail, req.PhoneNumber, "consumer", now, now)
	if err != nil {
		slog.Error("Failed to insert entity", "error", err, "entityId", entityID, "query", entityQuery)
		return nil, errors.HandleDatabaseError(err, "create entity")
	}

	// Insert consumer
	consumerQuery := `INSERT INTO consumers (consumer_id, entity_id, created_at, updated_at) 
					  VALUES ($1, $2, $3, $4)`

	slog.Debug("Executing consumer insert", "consumerId", consumerID, "entityId", entityID)
	_, err = tx.ExecContext(context.Background(), consumerQuery, consumerID, entityID, now, now)
	if err != nil {
		slog.Error("Failed to insert consumer", "error", err, "consumerId", consumerID, "query", consumerQuery)
		return nil, errors.HandleDatabaseError(err, "create consumer")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return nil, errors.HandleDatabaseError(err, "commit transaction")
	}

	// Create response object
	consumer := &models.Consumer{
		ConsumerID:   consumerID,
		ConsumerName: req.ConsumerName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	slog.Info("Successfully created consumer", "consumerId", consumerID, "entityId", entityID, "consumerName", req.ConsumerName)
	return consumer, nil
}

// GetConsumer retrieves a specific consumer
func (s *ConsumerService) GetConsumer(id string) (*models.Consumer, error) {
	slog.Info("Starting consumer retrieval", "consumerId", id)

	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		slog.Error("Database connection validation failed", "error", err, "consumerId", id)
		return nil, errors.InternalErrorWithCause("database connection validation failed", err)
	}

	query := `SELECT c.consumer_id, e.entity_name, e.contact_email, e.phone_number, c.created_at, c.updated_at 
			  FROM consumers c 
			  JOIN entities e ON c.entity_id = e.entity_id 
			  WHERE c.consumer_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing database query", "query", "SELECT FROM consumers WHERE consumer_id", "consumerId", id)
	start := time.Now()
	row := s.db.QueryRowContext(ctx, query, id)

	consumer := &models.Consumer{}
	err := row.Scan(&consumer.ConsumerID, &consumer.ConsumerName, &consumer.ContactEmail, &consumer.PhoneNumber, &consumer.CreatedAt, &consumer.UpdatedAt)
	duration := time.Since(start)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Consumer not found", "consumerId", id, "duration", duration)
			return nil, errors.NotFoundError("consumer")
		}
		slog.Error("Database query failed", "error", err, "query", "SELECT FROM consumers", "consumerId", id, "duration", duration)
		return nil, errors.HandleDatabaseError(err, "get consumer")
	}

	slog.Info("Successfully retrieved consumer", "consumerId", id, "consumerName", consumer.ConsumerName, "duration", duration)
	return consumer, nil
}

// GetAllConsumers retrieves all consumers
func (s *ConsumerService) GetAllConsumers() ([]*models.Consumer, error) {
	slog.Info("Starting retrieval of all consumers")

	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		slog.Error("Database connection validation failed", "error", err)
		return nil, errors.HandleDatabaseError(err, "validate connection")
	}

	query := `SELECT c.consumer_id, e.entity_name, e.contact_email, e.phone_number, c.created_at, c.updated_at 
			  FROM consumers c 
			  JOIN entities e ON c.entity_id = e.entity_id 
			  ORDER BY c.created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing database query", "query", "SELECT FROM consumers ORDER BY created_at DESC")
	start := time.Now()
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("Database query failed", "error", err, "query", query, "duration", time.Since(start))
		return nil, errors.HandleDatabaseError(err, "get all consumers")
	}
	defer rows.Close()

	var consumers []*models.Consumer
	rowCount := 0
	for rows.Next() {
		consumer := &models.Consumer{}
		err := rows.Scan(&consumer.ConsumerID, &consumer.ConsumerName, &consumer.ContactEmail, &consumer.PhoneNumber, &consumer.CreatedAt, &consumer.UpdatedAt)
		if err != nil {
			slog.Error("Failed to scan consumer row", "error", err, "rowCount", rowCount)
			return nil, errors.HandleDatabaseError(err, "scan consumer row")
		}
		consumers = append(consumers, consumer)
		rowCount++
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		slog.Error("Error during row iteration", "error", err, "rowCount", rowCount)
		return nil, errors.HandleDatabaseError(err, "iterate consumers")
	}

	duration := time.Since(start)
	slog.Info("Successfully retrieved all consumers", "count", len(consumers), "duration", duration)
	return consumers, nil
}

// UpdateConsumer updates a consumer
func (s *ConsumerService) UpdateConsumer(id string, req models.UpdateConsumerRequest) (*models.Consumer, error) {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	// First get the existing consumer
	consumer, err := s.GetConsumer(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.ConsumerName != nil {
		consumer.ConsumerName = *req.ConsumerName
	}
	if req.ContactEmail != nil {
		consumer.ContactEmail = *req.ContactEmail
	}
	if req.PhoneNumber != nil {
		consumer.PhoneNumber = *req.PhoneNumber
	}

	consumer.UpdatedAt = time.Now()

	// Update the entity record instead of consumer record
	query := `UPDATE entities SET entity_name = $1, contact_email = $2, phone_number = $3, updated_at = $4 
			  WHERE entity_id = (SELECT entity_id FROM consumers WHERE consumer_id = $5)`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing entity update", "consumerId", id, "query", query)
	_, err = s.db.ExecContext(ctx, query, consumer.ConsumerName, consumer.ContactEmail, consumer.PhoneNumber, consumer.UpdatedAt, id)
	if err != nil {
		slog.Error("Failed to update entity", "error", err, "consumerId", id, "query", query)
		return nil, errors.HandleDatabaseError(err, "update consumer entity")
	}

	slog.Info("Updated consumer", "consumerId", id)
	return consumer, nil
}

// DeleteConsumer deletes a consumer
func (s *ConsumerService) DeleteConsumer(id string) error {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return fmt.Errorf("database connection validation failed: %w", err)
	}

	query := `DELETE FROM consumers WHERE consumer_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete consumer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("consumer not found")
	}

	slog.Info("Deleted consumer", "consumerId", id)
	return nil
}

// ConsumerApp management methods

// CreateConsumerApp creates a new consumer application
func (s *ConsumerService) CreateConsumerApp(req models.CreateConsumerAppRequest) (*models.ConsumerApp, error) {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	// Verify consumer exists
	_, err := s.GetConsumer(req.ConsumerID)
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	// Generate unique submission ID
	submissionID, err := s.generateSubmissionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate submission ID: %w", err)
	}

	// Serialize required fields to JSON
	slog.Debug("Serializing required fields", "fields", req.RequiredFields)
	requiredFieldsJSON, err := json.Marshal(req.RequiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize required fields: %w", err)
	}
	slog.Debug("Serialized required fields JSON", "json", string(requiredFieldsJSON))

	now := time.Now()
	application := &models.ConsumerApp{
		SubmissionID:   submissionID,
		ConsumerID:     req.ConsumerID,
		Status:         models.StatusPending,
		RequiredFields: req.RequiredFields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	query := `INSERT INTO consumer_apps (submission_id, consumer_id, status, required_fields, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing consumer app insert", "submissionId", application.SubmissionID, "consumerId", application.ConsumerID)
	_, err = s.db.ExecContext(ctx, query, application.SubmissionID, application.ConsumerID, application.Status, string(requiredFieldsJSON), application.CreatedAt, application.UpdatedAt)
	if err != nil {
		slog.Error("Failed to insert consumer app", "error", err, "submissionId", application.SubmissionID, "query", query)
		return nil, errors.HandleDatabaseError(err, "create consumer application")
	}

	slog.Info("Created new consumer application", "submissionId", submissionID, "consumerId", req.ConsumerID)
	return application, nil
}

// GetConsumerApp retrieves a specific consumer application
func (s *ConsumerService) GetConsumerApp(id string) (*models.ConsumerApp, error) {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps WHERE submission_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, id)

	application := &models.ConsumerApp{}
	var requiredFieldsJSON string
	var credentialsJSON sql.NullString

	err := row.Scan(&application.SubmissionID, &application.ConsumerID, &application.Status, &requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundError("consumer application not found")
		}
		return nil, errors.HandleDatabaseError(err, "failed to get consumer application")
	}

	// Deserialize required fields from JSON
	slog.Debug("Raw required fields JSON from database", "submissionId", id, "json", requiredFieldsJSON)
	if err := json.Unmarshal([]byte(requiredFieldsJSON), &application.RequiredFields); err != nil {
		slog.Error("Failed to deserialize required fields", "error", err, "submissionId", id, "json", requiredFieldsJSON)
		return nil, fmt.Errorf("failed to deserialize required fields: %w", err)
	}
	slog.Debug("Deserialized required fields", "submissionId", id, "fields", application.RequiredFields)

	// Parse credentials if present
	if credentialsJSON.Valid && credentialsJSON.String != "" {
		var credentials models.Credentials
		if err := json.Unmarshal([]byte(credentialsJSON.String), &credentials); err != nil {
			slog.Warn("Failed to parse credentials JSON", "error", err, "submissionId", application.SubmissionID)
		} else {
			application.Credentials = &credentials
		}
	}

	slog.Info("Retrieved consumer application", "submissionId", id, "consumerId", application.ConsumerID)
	return application, nil
}

// UpdateConsumerApp updates a consumer application
func (s *ConsumerService) UpdateConsumerApp(id string, req models.UpdateConsumerAppRequest) (*models.UpdateConsumerAppResponse, error) {
	// First get the existing application
	application, err := s.GetConsumerApp(id)
	if err != nil {
		return nil, err
	}

	response := &models.UpdateConsumerAppResponse{
		ConsumerApp: application,
	}

	// Update fields if provided
	if req.Status != nil {
		application.Status = *req.Status
	}
	if req.RequiredFields != nil {
		application.RequiredFields = req.RequiredFields
	}

	// Generate credentials if status is approved and credentials don't exist
	if application.Status == models.StatusApproved && application.Credentials == nil {
		credentials, err := s.generateCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to generate credentials: %w", err)
		}
		application.Credentials = credentials
	}

	application.UpdatedAt = time.Now()

	query := `UPDATE consumer_apps SET status = $1, required_fields = $2, credentials = $3, updated_at = $4 
			  WHERE submission_id = $5`

	// Serialize required fields to JSON
	requiredFieldsJSON, err := json.Marshal(application.RequiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize required fields: %w", err)
	}

	// Serialize credentials to JSON
	var credentialsJSON sql.NullString
	if application.Credentials != nil {
		credentialsBytes, err := json.Marshal(application.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize credentials: %w", err)
		}
		credentialsJSON.String = string(credentialsBytes)
		credentialsJSON.Valid = true
	}

	slog.Debug("Executing consumer app update", "submissionId", id, "query", query)
	_, err = s.db.Exec(query, application.Status, string(requiredFieldsJSON), credentialsJSON, application.UpdatedAt, id)
	if err != nil {
		slog.Error("Failed to update consumer app", "error", err, "submissionId", id, "query", query)
		return nil, errors.HandleDatabaseError(err, "update consumer application")
	}

	// If the application was approved, update provider metadata in PDP
	if application.Status == models.StatusApproved && s.pdpClient != nil {
		if err := s.updateProviderMetadataForApprovedApp(application); err != nil {
			slog.Warn("Failed to update provider metadata in PDP", "error", err, "submissionId", id)
			// Don't fail the entire operation if PDP update fails
		}
	}

	slog.Info("Updated consumer application", "submissionId", id, "status", application.Status)
	return response, nil
}

// updateProviderMetadataForApprovedApp updates provider metadata in PDP for approved consumer app
func (s *ConsumerService) updateProviderMetadataForApprovedApp(app *models.ConsumerApp) error {
	// Convert required fields to provider field grants
	fields := make([]models.ProviderFieldGrant, 0, len(app.RequiredFields))

	for _, field := range app.RequiredFields {
		// Default grant duration to 30 days if not specified
		grantDuration := "P30D"
		if field.GrantDuration != "" {
			grantDuration = field.GrantDuration
		}

		fields = append(fields, models.ProviderFieldGrant{
			FieldName:     field.FieldName,
			GrantDuration: grantDuration,
		})
	}

	// Create metadata update request
	req := models.ProviderMetadataUpdateRequest{
		ApplicationID: app.ConsumerID, // Use consumer ID as application ID
		Fields:        fields,
	}

	// Send request to PDP
	response, err := s.pdpClient.UpdateProviderMetadata(req)
	if err != nil {
		return fmt.Errorf("failed to update provider metadata: %w", err)
	}

	slog.Info("Successfully updated provider metadata in PDP",
		"applicationId", req.ApplicationID,
		"updated", response.Updated,
		"fields", len(fields))

	return nil
}

// GetAllConsumerApps retrieves all consumer applications
func (s *ConsumerService) GetAllConsumerApps() ([]*models.ConsumerApp, error) {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer applications: %w", err)
	}
	defer rows.Close()

	var applications []*models.ConsumerApp
	for rows.Next() {
		application := &models.ConsumerApp{}
		var requiredFieldsJSON string
		var credentialsJSON sql.NullString

		err := rows.Scan(&application.SubmissionID, &application.ConsumerID, &application.Status, &requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer application: %w", err)
		}

		// Deserialize required fields from JSON
		if err := json.Unmarshal([]byte(requiredFieldsJSON), &application.RequiredFields); err != nil {
			return nil, fmt.Errorf("failed to deserialize required fields: %w", err)
		}

		// Parse credentials if present
		if credentialsJSON.Valid && credentialsJSON.String != "" {
			var credentials models.Credentials
			if err := json.Unmarshal([]byte(credentialsJSON.String), &credentials); err != nil {
				slog.Warn("Failed to parse credentials JSON", "error", err, "submissionId", application.SubmissionID)
			} else {
				application.Credentials = &credentials
			}
		}
		applications = append(applications, application)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate consumer applications: %w", err)
	}

	return applications, nil
}

// GetConsumerAppsByConsumerID retrieves all applications for a specific consumer
func (s *ConsumerService) GetConsumerAppsByConsumerID(consumerID string) ([]*models.ConsumerApp, error) {
	// Validate database connection
	if err := s.validateDBConnection(); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	// Verify consumer exists
	_, err := s.GetConsumer(consumerID)
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps WHERE consumer_id = $1 ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer applications: %w", err)
	}
	defer rows.Close()

	var applications []*models.ConsumerApp
	for rows.Next() {
		application := &models.ConsumerApp{}
		var requiredFieldsJSON string
		var credentialsJSON sql.NullString

		err := rows.Scan(&application.SubmissionID, &application.ConsumerID, &application.Status, &requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer application: %w", err)
		}

		// Deserialize required fields from JSON
		if err := json.Unmarshal([]byte(requiredFieldsJSON), &application.RequiredFields); err != nil {
			return nil, fmt.Errorf("failed to deserialize required fields: %w", err)
		}

		// Parse credentials if present
		if credentialsJSON.Valid && credentialsJSON.String != "" {
			var credentials models.Credentials
			if err := json.Unmarshal([]byte(credentialsJSON.String), &credentials); err != nil {
				slog.Warn("Failed to parse credentials JSON", "error", err, "submissionId", application.SubmissionID)
			} else {
				application.Credentials = &credentials
			}
		}
		applications = append(applications, application)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate consumer applications: %w", err)
	}

	return applications, nil
}

// ID generation methods

// generateConsumerID generates a unique consumer ID
func (s *ConsumerService) generateConsumerID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "consumer_" + id.String(), nil
}

// generateEntityID generates a unique entity ID
func (s *ConsumerService) generateEntityID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "entity_" + id.String(), nil
}

// generateSubmissionID generates a unique submission ID
func (s *ConsumerService) generateSubmissionID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return "sub_" + id.String(), nil
}

// generateCredentials generates API credentials
func (s *ConsumerService) generateCredentials() (*models.Credentials, error) {
	apiKey, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	apiSecret, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API secret: %w", err)
	}

	return &models.Credentials{
		APIKey:    "key_" + apiKey.String(),
		APISecret: "secret_" + apiSecret.String(),
	}, nil
}
