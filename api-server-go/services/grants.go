package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/pkg/errors"
	"github.com/gov-dx-sandbox/api-server-go/shared/database"
)

type GrantsService struct {
	db *sql.DB
}

func NewGrantsService(db *sql.DB) *GrantsService {
	return &GrantsService{
		db: db,
	}
}

// Consumer Grants Management

// GetAllConsumerGrants retrieves all consumer grants
func (s *GrantsService) GetAllConsumerGrants() (*models.ConsumerGrantsData, error) {
	slog.Info("Starting retrieval of all consumer grants")

	// Validate database connection
	if err := database.ValidateDBConnection(s.db); err != nil {
		slog.Error("Database connection validation failed", "error", err)
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	query := `SELECT consumer_id, approved_fields, created_at, updated_at FROM consumer_grants ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Debug("Executing database query", "query", "SELECT FROM consumer_grants ORDER BY created_at DESC")
	start := time.Now()
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("Database query failed", "error", err, "query", "SELECT FROM consumer_grants", "duration", time.Since(start))
		return nil, fmt.Errorf("failed to get consumer grants: %w", err)
	}
	defer rows.Close()

	grants := make(map[string]models.ConsumerGrant)
	rowCount := 0
	for rows.Next() {
		grant := models.ConsumerGrant{}
		var approvedFieldsJSON string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&grant.ConsumerID, &approvedFieldsJSON, &createdAt, &updatedAt)
		if err != nil {
			slog.Error("Failed to scan consumer grant row", "error", err, "rowCount", rowCount)
			return nil, fmt.Errorf("failed to scan consumer grant: %w", err)
		}

		// Parse approved fields JSON
		err = json.Unmarshal([]byte(approvedFieldsJSON), &grant.ApprovedFields)
		if err != nil {
			slog.Error("Failed to parse approved fields JSON", "error", err, "consumerId", grant.ConsumerID, "rowCount", rowCount)
			return nil, fmt.Errorf("failed to parse approved fields: %w", err)
		}

		grant.CreatedAt = createdAt.Format(time.RFC3339)
		grant.UpdatedAt = updatedAt.Format(time.RFC3339)

		grants[grant.ConsumerID] = grant
		rowCount++
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		slog.Error("Error during row iteration", "error", err, "rowCount", rowCount)
		return nil, fmt.Errorf("failed to iterate consumer grants: %w", err)
	}

	duration := time.Since(start)
	slog.Info("Successfully retrieved all consumer grants", "count", len(grants), "duration", duration)
	return &models.ConsumerGrantsData{
		ConsumerGrants: grants,
	}, nil
}

// GetConsumerGrant retrieves a specific consumer grant
func (s *GrantsService) GetConsumerGrant(consumerID string) (*models.ConsumerGrant, error) {
	query := `SELECT consumer_id, approved_fields, created_at, updated_at FROM consumer_grants WHERE consumer_id = $1`

	row := s.db.QueryRow(query, consumerID)

	grant := &models.ConsumerGrant{}
	var approvedFieldsJSON string
	var createdAt, updatedAt time.Time

	err := row.Scan(&grant.ConsumerID, &approvedFieldsJSON, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consumer grant not found")
		}
		return nil, fmt.Errorf("failed to get consumer grant: %w", err)
	}

	// Parse approved fields JSON
	err = json.Unmarshal([]byte(approvedFieldsJSON), &grant.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse approved fields: %w", err)
	}

	grant.CreatedAt = createdAt.Format(time.RFC3339)
	grant.UpdatedAt = updatedAt.Format(time.RFC3339)

	return grant, nil
}

// CreateConsumerGrant creates a new consumer grant
func (s *GrantsService) CreateConsumerGrant(req models.CreateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	now := time.Now().UTC()

	// Serialize approved fields to JSON
	approvedFieldsJSON, err := json.Marshal(req.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approved fields: %w", err)
	}

	grant := models.ConsumerGrant{
		ConsumerID:     req.ConsumerID,
		ApprovedFields: req.ApprovedFields,
		CreatedAt:      now.Format(time.RFC3339),
		UpdatedAt:      now.Format(time.RFC3339),
	}

	query := `INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4)`

	slog.Debug("Executing consumer grant insert", "consumerId", grant.ConsumerID)
	_, err = s.db.Exec(query, grant.ConsumerID, approvedFieldsJSON, now, now)
	if err != nil {
		slog.Error("Failed to insert consumer grant", "error", err, "consumerId", grant.ConsumerID, "query", query)
		return nil, errors.HandleDatabaseError(err, "create consumer grant")
	}

	slog.Info("Created consumer grant", "consumerId", req.ConsumerID, "fields", req.ApprovedFields)
	return &grant, nil
}

// UpdateConsumerGrant updates an existing consumer grant
func (s *GrantsService) UpdateConsumerGrant(consumerID string, req models.UpdateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	// First get the existing grant
	grant, err := s.GetConsumerGrant(consumerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	grant.ApprovedFields = req.ApprovedFields
	grant.UpdatedAt = now.Format(time.RFC3339)

	// Serialize approved fields to JSON
	approvedFieldsJSON, err := json.Marshal(grant.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approved fields: %w", err)
	}

	query := `UPDATE consumer_grants SET approved_fields = $1, updated_at = $2 WHERE consumer_id = $3`

	slog.Debug("Executing consumer grant update", "consumerId", consumerID, "query", query)
	_, err = s.db.Exec(query, approvedFieldsJSON, now, consumerID)
	if err != nil {
		slog.Error("Failed to update consumer grant", "error", err, "consumerId", consumerID, "query", query)
		return nil, errors.HandleDatabaseError(err, "update consumer grant")
	}

	slog.Info("Updated consumer grant", "consumerId", consumerID, "fields", req.ApprovedFields)
	return grant, nil
}

// DeleteConsumerGrant deletes a consumer grant
func (s *GrantsService) DeleteConsumerGrant(consumerID string) error {
	query := `DELETE FROM consumer_grants WHERE consumer_id = $1`

	result, err := s.db.Exec(query, consumerID)
	if err != nil {
		return fmt.Errorf("failed to delete consumer grant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("consumer grant not found")
	}

	slog.Info("Deleted consumer grant", "consumerId", consumerID)
	return nil
}

// Provider Metadata Management

// GetAllProviderFields retrieves all provider fields
func (s *GrantsService) GetAllProviderFields() (*models.ProviderMetadataData, error) {
	query := `SELECT field_name, owner, provider, consent_required, access_control_type, allow_list, description, expiry_time, metadata, created_at, updated_at 
			  FROM provider_metadata ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider fields: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]models.ProviderField)
	for rows.Next() {
		field := models.ProviderField{}
		var fieldName string
		var allowListJSON sql.NullString
		var metadataJSON sql.NullString
		var description sql.NullString
		var expiryTime sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(&fieldName, &field.Owner, &field.Provider, &field.ConsentRequired, &field.AccessControlType, &allowListJSON, &description, &expiryTime, &metadataJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider field: %w", err)
		}

		field.FieldName = fieldName
		field.CreatedAt = createdAt.Format(time.RFC3339)
		field.UpdatedAt = updatedAt.Format(time.RFC3339)

		// Parse JSON fields
		if allowListJSON.Valid {
			err = json.Unmarshal([]byte(allowListJSON.String), &field.AllowList)
			if err != nil {
				return nil, fmt.Errorf("failed to parse allow list: %w", err)
			}
		}
		if metadataJSON.Valid {
			err = json.Unmarshal([]byte(metadataJSON.String), &field.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to parse metadata: %w", err)
			}
		}
		if description.Valid {
			field.Description = &description.String
		}
		if expiryTime.Valid {
			field.ExpiryTime = &expiryTime.String
		}

		fields[fieldName] = field
	}

	return &models.ProviderMetadataData{
		Fields: fields,
	}, nil
}

// GetProviderField retrieves a specific provider field
func (s *GrantsService) GetProviderField(fieldName string) (*models.ProviderField, error) {
	query := `SELECT field_name, owner, provider, consent_required, access_control_type, allow_list, description, expiry_time, metadata, created_at, updated_at 
			  FROM provider_metadata WHERE field_name = $1`

	row := s.db.QueryRow(query, fieldName)

	field := &models.ProviderField{}
	var allowListJSON sql.NullString
	var metadataJSON sql.NullString
	var description sql.NullString
	var expiryTime sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(&field.FieldName, &field.Owner, &field.Provider, &field.ConsentRequired, &field.AccessControlType, &allowListJSON, &description, &expiryTime, &metadataJSON, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider field not found")
		}
		return nil, fmt.Errorf("failed to get provider field: %w", err)
	}

	field.CreatedAt = createdAt.Format(time.RFC3339)
	field.UpdatedAt = updatedAt.Format(time.RFC3339)

	// Parse JSON fields
	if allowListJSON.Valid {
		err = json.Unmarshal([]byte(allowListJSON.String), &field.AllowList)
		if err != nil {
			return nil, fmt.Errorf("failed to parse allow list: %w", err)
		}
	}
	if metadataJSON.Valid {
		err = json.Unmarshal([]byte(metadataJSON.String), &field.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}
	if description.Valid {
		field.Description = &description.String
	}
	if expiryTime.Valid {
		field.ExpiryTime = &expiryTime.String
	}

	return field, nil
}

// CreateProviderField creates a new provider field
func (s *GrantsService) CreateProviderField(req models.CreateProviderFieldRequest) (*models.ProviderField, error) {
	now := time.Now().UTC()

	// Serialize JSON fields
	allowListJSON, err := json.Marshal(req.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allow list: %w", err)
	}

	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	field := models.ProviderField{
		FieldName:         req.FieldName,
		Owner:             req.Owner,
		Provider:          req.Provider,
		ConsentRequired:   req.ConsentRequired,
		AccessControlType: req.AccessControlType,
		AllowList:         req.AllowList,
		Description:       req.Description,
		ExpiryTime:        req.ExpiryTime,
		Metadata:          req.Metadata,
		CreatedAt:         now.Format(time.RFC3339),
		UpdatedAt:         now.Format(time.RFC3339),
	}

	query := `INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, allow_list, description, expiry_time, metadata, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	slog.Debug("Executing provider field insert", "fieldName", field.FieldName)
	_, err = s.db.Exec(query, field.FieldName, field.Owner, field.Provider, field.ConsentRequired, field.AccessControlType, allowListJSON, field.Description, field.ExpiryTime, metadataJSON, now, now)
	if err != nil {
		slog.Error("Failed to insert provider field", "error", err, "fieldName", field.FieldName, "query", query)
		return nil, errors.HandleDatabaseError(err, "create provider field")
	}

	slog.Info("Created provider field", "fieldName", req.FieldName, "owner", req.Owner, "provider", req.Provider)
	return &field, nil
}

// UpdateProviderField updates an existing provider field
func (s *GrantsService) UpdateProviderField(fieldName string, req models.UpdateProviderFieldRequest) (*models.ProviderField, error) {
	// First get the existing field
	field, err := s.GetProviderField(fieldName)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Owner != nil {
		field.Owner = *req.Owner
	}
	if req.Provider != nil {
		field.Provider = *req.Provider
	}
	if req.ConsentRequired != nil {
		field.ConsentRequired = *req.ConsentRequired
	}
	if req.AccessControlType != nil {
		field.AccessControlType = *req.AccessControlType
	}
	if req.AllowList != nil {
		field.AllowList = req.AllowList
	}
	if req.Description != nil {
		field.Description = req.Description
	}
	if req.ExpiryTime != nil {
		field.ExpiryTime = req.ExpiryTime
	}
	if req.Metadata != nil {
		field.Metadata = req.Metadata
	}

	now := time.Now().UTC()
	field.UpdatedAt = now.Format(time.RFC3339)

	// Serialize JSON fields
	allowListJSON, err := json.Marshal(field.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allow list: %w", err)
	}

	metadataJSON, err := json.Marshal(field.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `UPDATE provider_metadata SET owner = $1, provider = $2, consent_required = $3, access_control_type = $4, allow_list = $5, description = $6, expiry_time = $7, metadata = $8, updated_at = $9 WHERE field_name = $10`

	slog.Debug("Executing provider field update", "fieldName", fieldName, "query", query)
	_, err = s.db.Exec(query, field.Owner, field.Provider, field.ConsentRequired, field.AccessControlType, allowListJSON, field.Description, field.ExpiryTime, metadataJSON, now, fieldName)
	if err != nil {
		slog.Error("Failed to update provider field", "error", err, "fieldName", fieldName, "query", query)
		return nil, errors.HandleDatabaseError(err, "update provider field")
	}

	slog.Info("Updated provider field", "fieldName", fieldName)
	return field, nil
}

// DeleteProviderField deletes a provider field
func (s *GrantsService) DeleteProviderField(fieldName string) error {
	query := `DELETE FROM provider_metadata WHERE field_name = $1`

	result, err := s.db.Exec(query, fieldName)
	if err != nil {
		return fmt.Errorf("failed to delete provider field: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("provider field not found")
	}

	slog.Info("Deleted provider field", "fieldName", fieldName)
	return nil
}

// Schema Conversion

// ConvertSchemaToProviderMetadata converts GraphQL SDL to provider metadata
func (s *GrantsService) ConvertSchemaToProviderMetadata(req models.SchemaConversionRequest) (*models.SchemaConversionResponse, error) {
	// This would integrate with the schema converter from the policy-decision-point
	// For now, we'll return a mock response
	// In a real implementation, this would call the schema converter service

	now := time.Now().UTC()

	// Mock conversion - in reality this would parse the SDL
	fields := map[string]models.ProviderField{
		"user.id": {
			Owner:             req.ProviderID,
			Provider:          req.ProviderID,
			ConsentRequired:   false,
			AccessControlType: "public",
			AllowList:         []models.AllowListEntry{},
		},
		"user.email": {
			Owner:             req.ProviderID,
			Provider:          req.ProviderID,
			ConsentRequired:   true,
			AccessControlType: "restricted",
			AllowList:         []models.AllowListEntry{},
		},
	}

	response := &models.SchemaConversionResponse{
		ProviderID:  req.ProviderID,
		Fields:      fields,
		ConvertedAt: now.Format(time.RFC3339),
	}

	slog.Info("Converted schema to provider metadata", "providerId", req.ProviderID, "fields", len(fields))
	return response, nil
}

// Export and Import

// ExportConsumerGrants exports consumer grants as JSON
func (s *GrantsService) ExportConsumerGrants() ([]byte, error) {
	grants, err := s.GetAllConsumerGrants()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(grants, "", "  ")
}

// ExportProviderMetadata exports provider metadata as JSON
func (s *GrantsService) ExportProviderMetadata() ([]byte, error) {
	fields, err := s.GetAllProviderFields()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(fields, "", "  ")
}

// ImportConsumerGrants imports consumer grants from JSON
func (s *GrantsService) ImportConsumerGrants(data []byte) error {
	var grants models.ConsumerGrantsData
	if err := json.Unmarshal(data, &grants); err != nil {
		return fmt.Errorf("failed to parse consumer grants JSON: %w", err)
	}

	// Import each grant
	for _, grant := range grants.ConsumerGrants {
		req := models.CreateConsumerGrantRequest{
			ConsumerID:     grant.ConsumerID,
			ApprovedFields: grant.ApprovedFields,
		}
		_, err := s.CreateConsumerGrant(req)
		if err != nil {
			return fmt.Errorf("failed to import consumer grant for %s: %w", grant.ConsumerID, err)
		}
	}

	slog.Info("Imported consumer grants", "count", len(grants.ConsumerGrants))
	return nil
}

// ImportProviderMetadata imports provider metadata from JSON
func (s *GrantsService) ImportProviderMetadata(data []byte) error {
	var metadata models.ProviderMetadataData
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse provider metadata JSON: %w", err)
	}

	// Import each field
	for fieldName, field := range metadata.Fields {
		req := models.CreateProviderFieldRequest{
			FieldName:         fieldName,
			Owner:             field.Owner,
			Provider:          field.Provider,
			ConsentRequired:   field.ConsentRequired,
			AccessControlType: field.AccessControlType,
			AllowList:         field.AllowList,
			Description:       field.Description,
			ExpiryTime:        field.ExpiryTime,
			Metadata:          field.Metadata,
		}
		_, err := s.CreateProviderField(req)
		if err != nil {
			return fmt.Errorf("failed to import provider field %s: %w", fieldName, err)
		}
	}

	slog.Info("Imported provider metadata", "fields", len(metadata.Fields))
	return nil
}

// Allow List Management Methods

// AddConsumerToAllowList adds a consumer to the allow_list for a specific field
func (s *GrantsService) AddConsumerToAllowList(fieldName string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	// Get current provider field
	field, err := s.GetProviderField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field '%s' not found: %w", fieldName, err)
	}

	// Check if consumer already exists in allow_list
	found := false
	for i, entry := range field.AllowList {
		if entry.ConsumerID == req.ConsumerID {
			// Update existing entry
			field.AllowList[i] = models.AllowListEntry{
				ConsumerID: req.ConsumerID,
				ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
				CreatedAt:  entry.CreatedAt, // Keep original creation time
			}
			found = true
			break
		}
	}

	// If consumer not found, add new entry
	if !found {
		now := time.Now().UTC()
		newEntry := models.AllowListEntry{
			ConsumerID: req.ConsumerID,
			ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
			CreatedAt:  now.Format(time.RFC3339),
		}
		field.AllowList = append(field.AllowList, newEntry)
	}

	// Update the field
	updateReq := models.UpdateProviderFieldRequest{
		AllowList: field.AllowList,
	}
	_, err = s.UpdateProviderField(fieldName, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}

	slog.Info("Added consumer to allow_list", "fieldName", fieldName, "consumerId", req.ConsumerID)

	// Find the entry for response
	var responseEntry models.AllowListEntry
	for _, entry := range field.AllowList {
		if entry.ConsumerID == req.ConsumerID {
			responseEntry = entry
			break
		}
	}

	response := &models.AllowListManagementResponse{
		Success:    true,
		Operation:  "consumer_added_to_allow_list",
		FieldName:  fieldName,
		ConsumerID: req.ConsumerID,
		Data:       responseEntry,
		Metadata: map[string]interface{}{
			"expires_at":     req.ExpiresAt,
			"grant_duration": req.GrantDuration,
			"reason":         req.Reason,
			"updated_by":     req.UpdatedBy,
			"updated_at":     time.Now().UTC().Format(time.RFC3339),
		},
	}

	return response, nil
}

// RemoveConsumerFromAllowList removes a consumer from the allow_list for a specific field
func (s *GrantsService) RemoveConsumerFromAllowList(fieldName, consumerID string) (*models.AllowListManagementResponse, error) {
	// Get current provider field
	field, err := s.GetProviderField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field '%s' not found: %w", fieldName, err)
	}

	// Find and remove consumer from allow_list
	var removedEntry models.AllowListEntry
	var newAllowList []models.AllowListEntry
	found := false

	for _, entry := range field.AllowList {
		if entry.ConsumerID == consumerID {
			removedEntry = entry
			found = true
		} else {
			newAllowList = append(newAllowList, entry)
		}
	}

	if !found {
		return nil, fmt.Errorf("consumer '%s' not found in allow_list for field '%s'", consumerID, fieldName)
	}

	// Update the field
	updateReq := models.UpdateProviderFieldRequest{
		AllowList: newAllowList,
	}
	_, err = s.UpdateProviderField(fieldName, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}

	slog.Info("Removed consumer from allow_list", "fieldName", fieldName, "consumerId", consumerID)

	response := &models.AllowListManagementResponse{
		Success:    true,
		Operation:  "consumer_removed_from_allow_list",
		FieldName:  fieldName,
		ConsumerID: consumerID,
		Data:       removedEntry,
		Metadata: map[string]interface{}{
			"removed_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	return response, nil
}

// GetAllowListForField retrieves the allow_list for a specific field
func (s *GrantsService) GetAllowListForField(fieldName string) (*models.AllowListListResponse, error) {
	// Get current provider field
	field, err := s.GetProviderField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field '%s' not found: %w", fieldName, err)
	}

	response := &models.AllowListListResponse{
		Success:   true,
		FieldName: fieldName,
		AllowList: field.AllowList,
		Count:     len(field.AllowList),
		Metadata: map[string]interface{}{
			"retrieved_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	return response, nil
}

// UpdateConsumerInAllowList updates an existing consumer in the allow_list
func (s *GrantsService) UpdateConsumerInAllowList(fieldName, consumerID string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	// Get current provider field
	field, err := s.GetProviderField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field '%s' not found: %w", fieldName, err)
	}

	// Find and update consumer in allow_list
	found := false
	for i, entry := range field.AllowList {
		if entry.ConsumerID == consumerID {
			field.AllowList[i] = models.AllowListEntry{
				ConsumerID: consumerID,
				ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
				CreatedAt:  entry.CreatedAt, // Keep original creation time
			}
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("consumer '%s' not found in allow_list for field '%s'", consumerID, fieldName)
	}

	// Update the field
	updateReq := models.UpdateProviderFieldRequest{
		AllowList: field.AllowList,
	}
	_, err = s.UpdateProviderField(fieldName, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update field: %w", err)
	}

	slog.Info("Updated consumer in allow_list", "fieldName", fieldName, "consumerId", consumerID)

	// Find the updated entry for response
	var responseEntry models.AllowListEntry
	for _, entry := range field.AllowList {
		if entry.ConsumerID == consumerID {
			responseEntry = entry
			break
		}
	}

	response := &models.AllowListManagementResponse{
		Success:    true,
		Operation:  "consumer_updated_in_allow_list",
		FieldName:  fieldName,
		ConsumerID: consumerID,
		Data:       responseEntry,
		Metadata: map[string]interface{}{
			"expires_at":     req.ExpiresAt,
			"grant_duration": req.GrantDuration,
			"reason":         req.Reason,
			"updated_by":     req.UpdatedBy,
			"updated_at":     time.Now().UTC().Format(time.RFC3339),
		},
	}

	return response, nil
}
