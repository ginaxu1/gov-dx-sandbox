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
	db *sql.DB
}

func NewConsumerService(db *sql.DB) *ConsumerService {
	return &ConsumerService{
		db: db,
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

	// Generate unique consumer ID
	consumerID, err := s.generateConsumerID()
	if err != nil {
		slog.Error("Failed to generate consumer ID", "error", err)
		return nil, errors.InternalErrorWithCause("failed to generate consumer ID", err)
	}

	now := time.Now()
	consumer := &models.Consumer{
		ConsumerID:   consumerID,
		ConsumerName: req.ConsumerName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `INSERT INTO consumers (consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing database query", "query", "INSERT INTO consumers", "consumerId", consumerID)
	start := time.Now()
	_, err = s.db.ExecContext(ctx, query, consumer.ConsumerID, consumer.ConsumerName, consumer.ContactEmail, consumer.PhoneNumber, consumer.CreatedAt, consumer.UpdatedAt)
	duration := time.Since(start)

	if err != nil {
		slog.Error("Database query failed", "error", err, "query", "INSERT INTO consumers", "consumerId", consumerID, "duration", duration)
		return nil, errors.HandleDatabaseError(err, "create consumer")
	}

	slog.Info("Successfully created consumer", "consumerId", consumerID, "consumerName", req.ConsumerName, "duration", duration)
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

	query := `SELECT consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at 
			  FROM consumers WHERE consumer_id = $1`

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

	query := `SELECT consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at 
			  FROM consumers ORDER BY created_at DESC`

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

	query := `UPDATE consumers SET consumer_name = $1, contact_email = $2, phone_number = $3, updated_at = $4 
			  WHERE consumer_id = $5`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = s.db.ExecContext(ctx, query, consumer.ConsumerName, consumer.ContactEmail, consumer.PhoneNumber, consumer.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer: %w", err)
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
	requiredFieldsJSON, err := json.Marshal(req.RequiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize required fields: %w", err)
	}

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

	_, err = s.db.ExecContext(ctx, query, application.SubmissionID, application.ConsumerID, application.Status, string(requiredFieldsJSON), application.CreatedAt, application.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer application: %w", err)
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

	_, err = s.db.Exec(query, application.Status, application.RequiredFields, credentialsJSON, application.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer application: %w", err)
	}

	slog.Info("Updated consumer application", "submissionId", id, "status", application.Status)
	return response, nil
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
