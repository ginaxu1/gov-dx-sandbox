package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type ConsumerServiceDB struct {
	db *sql.DB
}

// Consumer management methods

// CreateConsumer creates a new consumer
func (s *ConsumerServiceDB) CreateConsumer(req models.CreateConsumerRequest) (*models.Consumer, error) {
	// Generate unique consumer ID
	consumerID, err := s.generateConsumerID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate consumer ID: %w", err)
	}

	consumer := &models.Consumer{
		ConsumerID:   consumerID,
		ConsumerName: req.ConsumerName,
		ContactEmail: req.ContactEmail,
		PhoneNumber:  req.PhoneNumber,
		CreatedAt:    time.Now(),
	}

	query := `INSERT INTO consumers (consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = s.db.Exec(query, consumer.ConsumerID, consumer.ConsumerName, consumer.ContactEmail,
		consumer.PhoneNumber, consumer.CreatedAt, consumer.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	slog.Info("Created new consumer", "consumerId", consumerID)
	return consumer, nil
}

// GetConsumer retrieves a specific consumer
func (s *ConsumerServiceDB) GetConsumer(id string) (*models.Consumer, error) {
	query := `SELECT consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at 
			  FROM consumers WHERE consumer_id = $1`

	row := s.db.QueryRow(query, id)

	consumer := &models.Consumer{}
	err := row.Scan(&consumer.ConsumerID, &consumer.ConsumerName, &consumer.ContactEmail,
		&consumer.PhoneNumber, &consumer.CreatedAt, &consumer.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consumer not found")
		}
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}

	return consumer, nil
}

// GetAllConsumers retrieves all consumers
func (s *ConsumerServiceDB) GetAllConsumers() ([]*models.Consumer, error) {
	query := `SELECT consumer_id, consumer_name, contact_email, phone_number, created_at, updated_at 
			  FROM consumers ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumers: %w", err)
	}
	defer rows.Close()

	consumers := make([]*models.Consumer, 0)
	for rows.Next() {
		consumer := &models.Consumer{}
		err := rows.Scan(&consumer.ConsumerID, &consumer.ConsumerName, &consumer.ContactEmail,
			&consumer.PhoneNumber, &consumer.CreatedAt, &consumer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer: %w", err)
		}
		consumers = append(consumers, consumer)
	}

	return consumers, nil
}

// UpdateConsumer updates a consumer
func (s *ConsumerServiceDB) UpdateConsumer(id string, req models.UpdateConsumerRequest) (*models.Consumer, error) {
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

	_, err = s.db.Exec(query, consumer.ConsumerName, consumer.ContactEmail,
		consumer.PhoneNumber, consumer.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer: %w", err)
	}

	slog.Info("Updated consumer", "consumerId", id)
	return consumer, nil
}

// DeleteConsumer deletes a consumer
func (s *ConsumerServiceDB) DeleteConsumer(id string) error {
	query := `DELETE FROM consumers WHERE consumer_id = $1`

	result, err := s.db.Exec(query, id)
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
func (s *ConsumerServiceDB) CreateConsumerApp(req models.CreateConsumerAppRequest) (*models.ConsumerApp, error) {
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

	application := &models.ConsumerApp{
		SubmissionID:   submissionID,
		ConsumerID:     req.ConsumerID,
		Status:         models.StatusPending,
		RequiredFields: req.RequiredFields,
		CreatedAt:      time.Now(),
	}

	// Convert RequiredFields to JSON
	requiredFieldsJSON, err := json.Marshal(application.RequiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal required fields: %w", err)
	}

	query := `INSERT INTO consumer_apps (submission_id, consumer_id, status, required_fields, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = s.db.Exec(query, application.SubmissionID, application.ConsumerID,
		string(application.Status), requiredFieldsJSON, application.CreatedAt, application.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer app: %w", err)
	}

	slog.Info("Created new consumer application", "submissionId", submissionID, "consumerId", req.ConsumerID)
	return application, nil
}

// GetConsumerApp retrieves a specific consumer application
func (s *ConsumerServiceDB) GetConsumerApp(id string) (*models.ConsumerApp, error) {
	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps WHERE submission_id = $1`

	row := s.db.QueryRow(query, id)

	application := &models.ConsumerApp{}
	var statusStr string
	var requiredFieldsJSON, credentialsJSON sql.NullString

	err := row.Scan(&application.SubmissionID, &application.ConsumerID, &statusStr,
		&requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to get consumer app: %w", err)
	}

	application.Status = models.ApplicationStatus(statusStr)

	// Parse required fields JSON
	if requiredFieldsJSON.Valid {
		err = json.Unmarshal([]byte(requiredFieldsJSON.String), &application.RequiredFields)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal required fields: %w", err)
		}
	}

	// Parse credentials JSON
	if credentialsJSON.Valid {
		var credentials models.Credentials
		err = json.Unmarshal([]byte(credentialsJSON.String), &credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
		}
		application.Credentials = &credentials
	}

	return application, nil
}

// UpdateConsumerApp updates a consumer application
func (s *ConsumerServiceDB) UpdateConsumerApp(id string, req models.UpdateConsumerAppRequest) (*models.UpdateConsumerAppResponse, error) {
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
	application.UpdatedAt = time.Now()

	// Generate credentials if status is approved and credentials don't exist
	if application.Status == models.StatusApproved && application.Credentials == nil {
		credentials, err := s.generateCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to generate credentials: %w", err)
		}
		application.Credentials = credentials
	}

	// Convert fields to JSON
	requiredFieldsJSON, err := json.Marshal(application.RequiredFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal required fields: %w", err)
	}

	var credentialsJSON sql.NullString
	if application.Credentials != nil {
		credentialsBytes, err := json.Marshal(application.Credentials)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal credentials: %w", err)
		}
		credentialsJSON = sql.NullString{String: string(credentialsBytes), Valid: true}
	}

	query := `UPDATE consumer_apps SET status = $1, required_fields = $2, credentials = $3, updated_at = $4 
			  WHERE submission_id = $5`

	_, err = s.db.Exec(query, string(application.Status), requiredFieldsJSON, credentialsJSON,
		application.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer app: %w", err)
	}

	slog.Info("Updated consumer application", "submissionId", id, "status", application.Status)
	return response, nil
}

// GetAllConsumerApps retrieves all consumer applications
func (s *ConsumerServiceDB) GetAllConsumerApps() ([]*models.ConsumerApp, error) {
	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer apps: %w", err)
	}
	defer rows.Close()

	applications := make([]*models.ConsumerApp, 0)
	for rows.Next() {
		application := &models.ConsumerApp{}
		var statusStr string
		var requiredFieldsJSON, credentialsJSON sql.NullString

		err := rows.Scan(&application.SubmissionID, &application.ConsumerID, &statusStr,
			&requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer app: %w", err)
		}

		application.Status = models.ApplicationStatus(statusStr)

		// Parse required fields JSON
		if requiredFieldsJSON.Valid {
			err = json.Unmarshal([]byte(requiredFieldsJSON.String), &application.RequiredFields)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal required fields: %w", err)
			}
		}

		// Parse credentials JSON
		if credentialsJSON.Valid {
			var credentials models.Credentials
			err = json.Unmarshal([]byte(credentialsJSON.String), &credentials)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
			}
			application.Credentials = &credentials
		}

		applications = append(applications, application)
	}

	return applications, nil
}

// GetConsumerAppsByConsumerID retrieves all applications for a specific consumer
func (s *ConsumerServiceDB) GetConsumerAppsByConsumerID(consumerID string) ([]*models.ConsumerApp, error) {
	// Verify consumer exists
	_, err := s.GetConsumer(consumerID)
	if err != nil {
		return nil, fmt.Errorf("consumer not found: %w", err)
	}

	query := `SELECT submission_id, consumer_id, status, required_fields, credentials, created_at, updated_at 
			  FROM consumer_apps WHERE consumer_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer apps: %w", err)
	}
	defer rows.Close()

	applications := make([]*models.ConsumerApp, 0)
	for rows.Next() {
		application := &models.ConsumerApp{}
		var statusStr string
		var requiredFieldsJSON, credentialsJSON sql.NullString

		err := rows.Scan(&application.SubmissionID, &application.ConsumerID, &statusStr,
			&requiredFieldsJSON, &credentialsJSON, &application.CreatedAt, &application.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer app: %w", err)
		}

		application.Status = models.ApplicationStatus(statusStr)

		// Parse required fields JSON
		if requiredFieldsJSON.Valid {
			err = json.Unmarshal([]byte(requiredFieldsJSON.String), &application.RequiredFields)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal required fields: %w", err)
			}
		}

		// Parse credentials JSON
		if credentialsJSON.Valid {
			var credentials models.Credentials
			err = json.Unmarshal([]byte(credentialsJSON.String), &credentials)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
			}
			application.Credentials = &credentials
		}

		applications = append(applications, application)
	}

	return applications, nil
}

// Legacy methods for backward compatibility

// GetAllApplications retrieves all consumer applications (legacy)
func (s *ConsumerServiceDB) GetAllApplications() ([]*models.Application, error) {
	apps, err := s.GetAllConsumerApps()
	if err != nil {
		return nil, err
	}

	// Convert ConsumerApp to Application (they're the same type now)
	applications := make([]*models.Application, len(apps))
	for i, app := range apps {
		applications[i] = (*models.Application)(app)
	}

	return applications, nil
}

// CreateApplication creates a new consumer application (legacy)
func (s *ConsumerServiceDB) CreateApplication(req models.CreateApplicationRequest) (*models.Application, error) {
	// Convert to new format - this is a legacy method that creates a consumer app without a consumer
	// This should not be used in new code

	// Create a default consumer for legacy compatibility
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Legacy Consumer",
		ContactEmail: "legacy@example.com",
		PhoneNumber:  "000-000-0000",
	}

	consumer, err := s.CreateConsumer(consumerReq)
	if err != nil {
		return nil, err
	}

	consumerAppReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: req.RequiredFields,
	}

	consumerApp, err := s.CreateConsumerApp(consumerAppReq)
	if err != nil {
		return nil, err
	}

	return (*models.Application)(consumerApp), nil
}

// GetApplication retrieves a specific consumer application
func (s *ConsumerServiceDB) GetApplication(id string) (*models.Application, error) {
	app, err := s.GetConsumerApp(id)
	if err != nil {
		return nil, err
	}
	return (*models.Application)(app), nil
}

// UpdateApplication updates a consumer application
func (s *ConsumerServiceDB) UpdateApplication(id string, req models.UpdateApplicationRequest) (*models.UpdateApplicationResponse, error) {
	// Convert to new format
	consumerAppReq := models.UpdateConsumerAppRequest{
		Status:         req.Status,
		RequiredFields: req.RequiredFields,
	}

	response, err := s.UpdateConsumerApp(id, consumerAppReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	legacyResponse := &models.UpdateApplicationResponse{
		ConsumerApp: response.ConsumerApp,
	}

	// For now, assign a default provider ID when approved
	// In a real system, this would be based on business logic
	if response.ConsumerApp.Status == models.StatusApproved {
		legacyResponse.ProviderID = "default_provider"
	}

	return legacyResponse, nil
}

// DeleteApplication deletes a consumer application
func (s *ConsumerServiceDB) DeleteApplication(id string) error {
	query := `DELETE FROM consumer_apps WHERE submission_id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete consumer app: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("application not found")
	}

	slog.Info("Deleted application", "appId", id)
	return nil
}

// ID generation methods (same as original)

// generateConsumerID generates a unique consumer ID
func (s *ConsumerServiceDB) generateConsumerID() (string, error) {
	// Use database sequence or timestamp-based approach for better uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("consumer_%d", timestamp), nil
}

// generateSubmissionID generates a unique submission ID
func (s *ConsumerServiceDB) generateSubmissionID() (string, error) {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("sub_%d", timestamp), nil
}

// generateCredentials generates API credentials
func (s *ConsumerServiceDB) generateCredentials() (*models.Credentials, error) {
	// For now, use timestamp-based approach
	// In production, use crypto/rand for better security
	timestamp := time.Now().UnixNano()
	apiKey := fmt.Sprintf("ak_%d", timestamp)
	apiSecret := fmt.Sprintf("as_%d", timestamp+1)

	return &models.Credentials{
		APIKey:    apiKey,
		APISecret: apiSecret,
	}, nil
}
