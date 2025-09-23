package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type GrantsServiceDB struct {
	db *sql.DB
}

// Consumer Grants Management

// GetAllConsumerGrants retrieves all consumer grants
func (s *GrantsServiceDB) GetAllConsumerGrants() (*models.ConsumerGrantsData, error) {
	query := `SELECT consumer_id, approved_fields, created_at, updated_at FROM consumer_grants ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer grants: %w", err)
	}
	defer rows.Close()

	grants := make(map[string]models.ConsumerGrant)
	for rows.Next() {
		var grant models.ConsumerGrant
		var approvedFieldsJSON string

		err := rows.Scan(&grant.ConsumerID, &approvedFieldsJSON, &grant.CreatedAt, &grant.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumer grant: %w", err)
		}

		// Parse approved fields JSON
		err = json.Unmarshal([]byte(approvedFieldsJSON), &grant.ApprovedFields)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal approved fields: %w", err)
		}

		grants[grant.ConsumerID] = grant
	}

	return &models.ConsumerGrantsData{
		LegacyConsumerGrants: grants,
	}, nil
}

// GetConsumerGrant retrieves a specific consumer grant
func (s *GrantsServiceDB) GetConsumerGrant(consumerID string) (*models.ConsumerGrant, error) {
	query := `SELECT consumer_id, approved_fields, created_at, updated_at FROM consumer_grants WHERE consumer_id = $1`

	row := s.db.QueryRow(query, consumerID)

	var grant models.ConsumerGrant
	var approvedFieldsJSON string

	err := row.Scan(&grant.ConsumerID, &approvedFieldsJSON, &grant.CreatedAt, &grant.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consumer grant not found")
		}
		return nil, fmt.Errorf("failed to get consumer grant: %w", err)
	}

	// Parse approved fields JSON
	err = json.Unmarshal([]byte(approvedFieldsJSON), &grant.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal approved fields: %w", err)
	}

	return &grant, nil
}

// CreateConsumerGrant creates a new consumer grant
func (s *GrantsServiceDB) CreateConsumerGrant(req models.CreateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	grant := models.ConsumerGrant{
		ConsumerID:     req.ConsumerID,
		ApprovedFields: req.ApprovedFields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Convert approved fields to JSON
	approvedFieldsJSON, err := json.Marshal(grant.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approved fields: %w", err)
	}

	query := `INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) ON CONFLICT (consumer_id) DO UPDATE SET 
			  approved_fields = EXCLUDED.approved_fields, updated_at = EXCLUDED.updated_at`

	_, err = s.db.Exec(query, grant.ConsumerID, approvedFieldsJSON, grant.CreatedAt, grant.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer grant: %w", err)
	}

	slog.Info("Created consumer grant", "consumerId", req.ConsumerID, "fields", req.ApprovedFields)
	return &grant, nil
}

// UpdateConsumerGrant updates an existing consumer grant
func (s *GrantsServiceDB) UpdateConsumerGrant(consumerID string, req models.UpdateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	// First get the existing grant
	grant, err := s.GetConsumerGrant(consumerID)
	if err != nil {
		return nil, err
	}

	grant.ApprovedFields = req.ApprovedFields
	grant.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// Convert approved fields to JSON
	approvedFieldsJSON, err := json.Marshal(grant.ApprovedFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approved fields: %w", err)
	}

	query := `UPDATE consumer_grants SET approved_fields = $1, updated_at = $2 WHERE consumer_id = $3`

	_, err = s.db.Exec(query, approvedFieldsJSON, grant.UpdatedAt, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update consumer grant: %w", err)
	}

	slog.Info("Updated consumer grant", "consumerId", consumerID, "fields", req.ApprovedFields)
	return grant, nil
}

// DeleteConsumerGrant deletes a consumer grant
func (s *GrantsServiceDB) DeleteConsumerGrant(consumerID string) error {
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
func (s *GrantsServiceDB) GetAllProviderFields() (*models.ProviderMetadataData, error) {
	query := `SELECT field_name, owner, provider, consent_required, access_control_type, allow_list, 
			  description, expiry_time, metadata, created_at, updated_at FROM provider_metadata ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider fields: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]models.ProviderField)
	for rows.Next() {
		var field models.ProviderField
		var fieldName string
		var allowListJSON, metadataJSON sql.NullString
		var expiryTime sql.NullString

		err := rows.Scan(&fieldName, &field.Owner, &field.Provider, &field.ConsentRequired,
			&field.AccessControlType, &allowListJSON, &field.Description, &expiryTime,
			&metadataJSON, &field.CreatedAt, &field.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider field: %w", err)
		}

		// Parse allow list JSON
		if allowListJSON.Valid {
			err = json.Unmarshal([]byte(allowListJSON.String), &field.AllowList)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal allow list: %w", err)
			}
		}

		// Parse expiry time
		if expiryTime.Valid {
			field.ExpiryTime = expiryTime.String
		}

		// Parse metadata JSON
		if metadataJSON.Valid {
			err = json.Unmarshal([]byte(metadataJSON.String), &field.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		fields[fieldName] = field
	}

	return &models.ProviderMetadataData{
		Fields: fields,
	}, nil
}

// GetProviderField retrieves a specific provider field
func (s *GrantsServiceDB) GetProviderField(fieldName string) (*models.ProviderField, error) {
	query := `SELECT owner, provider, consent_required, access_control_type, allow_list, 
			  description, expiry_time, metadata, created_at, updated_at FROM provider_metadata WHERE field_name = $1`

	row := s.db.QueryRow(query, fieldName)

	var field models.ProviderField
	var allowListJSON, metadataJSON sql.NullString
	var expiryTime sql.NullString

	err := row.Scan(&field.Owner, &field.Provider, &field.ConsentRequired,
		&field.AccessControlType, &allowListJSON, &field.Description, &expiryTime,
		&metadataJSON, &field.CreatedAt, &field.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider field not found")
		}
		return nil, fmt.Errorf("failed to get provider field: %w", err)
	}

	// Parse allow list JSON
	if allowListJSON.Valid {
		err = json.Unmarshal([]byte(allowListJSON.String), &field.AllowList)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal allow list: %w", err)
		}
	}

	// Parse expiry time
	if expiryTime.Valid {
		field.ExpiryTime = expiryTime.String
	}

	// Parse metadata JSON
	if metadataJSON.Valid {
		err = json.Unmarshal([]byte(metadataJSON.String), &field.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &field, nil
}

// CreateProviderField creates a new provider field
func (s *GrantsServiceDB) CreateProviderField(req models.CreateProviderFieldRequest) (*models.ProviderField, error) {
	field := models.ProviderField{
		Owner:             req.Owner,
		Provider:          req.Provider,
		ConsentRequired:   req.ConsentRequired,
		AccessControlType: req.AccessControlType,
		AllowList:         req.AllowList,
		Description:       req.Description,
		ExpiryTime:        req.ExpiryTime,
		Metadata:          req.Metadata,
	}

	// Convert fields to JSON
	allowListJSON, err := json.Marshal(field.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allow list: %w", err)
	}

	var metadataJSON sql.NullString
	if field.Metadata != nil {
		metadataBytes, err := json.Marshal(field.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
	}

	var expiryTimeJSON sql.NullString
	if field.ExpiryTime != "" {
		expiryTimeJSON = sql.NullString{String: field.ExpiryTime, Valid: true}
	}

	query := `INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, 
			  allow_list, description, expiry_time, metadata, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now()
	_, err = s.db.Exec(query, req.FieldName, field.Owner, field.Provider, field.ConsentRequired,
		field.AccessControlType, allowListJSON, field.Description, expiryTimeJSON, metadataJSON, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider field: %w", err)
	}

	slog.Info("Created provider field", "fieldName", req.FieldName, "owner", req.Owner, "provider", req.Provider)
	return &field, nil
}

// UpdateProviderField updates an existing provider field
func (s *GrantsServiceDB) UpdateProviderField(fieldName string, req models.UpdateProviderFieldRequest) (*models.ProviderField, error) {
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
		field.Description = *req.Description
	}
	if req.ExpiryTime != nil {
		field.ExpiryTime = *req.ExpiryTime
	}
	if req.Metadata != nil {
		field.Metadata = req.Metadata
	}

	field.UpdatedAt = time.Now()

	// Convert fields to JSON
	allowListJSON, err := json.Marshal(field.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allow list: %w", err)
	}

	var metadataJSON sql.NullString
	if field.Metadata != nil {
		metadataBytes, err := json.Marshal(field.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
	}

	var expiryTimeJSON sql.NullString
	if field.ExpiryTime != "" {
		expiryTimeJSON = sql.NullString{String: field.ExpiryTime, Valid: true}
	}

	query := `UPDATE provider_metadata SET owner = $1, provider = $2, consent_required = $3, access_control_type = $4, 
			  allow_list = $5, description = $6, expiry_time = $7, metadata = $8, updated_at = $9 WHERE field_name = $10`

	_, err = s.db.Exec(query, field.Owner, field.Provider, field.ConsentRequired,
		field.AccessControlType, allowListJSON, field.Description, expiryTimeJSON, metadataJSON,
		field.UpdatedAt, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider field: %w", err)
	}

	slog.Info("Updated provider field", "fieldName", fieldName)
	return field, nil
}

// DeleteProviderField deletes a provider field
func (s *GrantsServiceDB) DeleteProviderField(fieldName string) error {
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
func (s *GrantsServiceDB) ConvertSchemaToProviderMetadata(req models.SchemaConversionRequest) (*models.SchemaConversionResponse, error) {
	// This would integrate with the schema converter from the policy-decision-point
	// For now, we'll return a mock response
	// In a real implementation, this would call the schema converter service

	now := time.Now().UTC().Format(time.RFC3339)

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
		ConvertedAt: now,
	}

	slog.Info("Converted schema to provider metadata", "providerId", req.ProviderID, "fields", len(fields))
	return response, nil
}

// Export and Import

// ExportConsumerGrants exports consumer grants as JSON
func (s *GrantsServiceDB) ExportConsumerGrants() ([]byte, error) {
	grants, err := s.GetAllConsumerGrants()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(grants, "", "  ")
}

// ExportProviderMetadata exports provider metadata as JSON
func (s *GrantsServiceDB) ExportProviderMetadata() ([]byte, error) {
	metadata, err := s.GetAllProviderFields()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metadata, "", "  ")
}

// ImportConsumerGrants imports consumer grants from JSON
func (s *GrantsServiceDB) ImportConsumerGrants(data []byte) error {
	var grants models.ConsumerGrantsData
	if err := json.Unmarshal(data, &grants); err != nil {
		return fmt.Errorf("failed to parse consumer grants JSON: %w", err)
	}

	// Insert or update each grant
	for consumerID, grant := range grants.LegacyConsumerGrants {
		approvedFieldsJSON, err := json.Marshal(grant.ApprovedFields)
		if err != nil {
			return fmt.Errorf("failed to marshal approved fields for consumer %s: %w", consumerID, err)
		}

		query := `INSERT INTO consumer_grants (consumer_id, approved_fields, created_at, updated_at) 
				  VALUES ($1, $2, $3, $4) ON CONFLICT (consumer_id) DO UPDATE SET 
				  approved_fields = EXCLUDED.approved_fields, updated_at = EXCLUDED.updated_at`

		_, err = s.db.Exec(query, grant.ConsumerID, approvedFieldsJSON, grant.CreatedAt, grant.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to import consumer grant for %s: %w", consumerID, err)
		}
	}

	slog.Info("Imported consumer grants", "count", len(grants.LegacyConsumerGrants))
	return nil
}

// ImportProviderMetadata imports provider metadata from JSON
func (s *GrantsServiceDB) ImportProviderMetadata(data []byte) error {
	var metadata models.ProviderMetadataData
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse provider metadata JSON: %w", err)
	}

	// Insert or update each field
	for fieldName, field := range metadata.Fields {
		allowListJSON, err := json.Marshal(field.AllowList)
		if err != nil {
			return fmt.Errorf("failed to marshal allow list for field %s: %w", fieldName, err)
		}

		var metadataJSON sql.NullString
		if field.Metadata != nil {
			metadataBytes, err := json.Marshal(field.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for field %s: %w", fieldName, err)
			}
			metadataJSON = sql.NullString{String: string(metadataBytes), Valid: true}
		}

		var expiryTimeJSON sql.NullString
		if field.ExpiryTime != "" {
			expiryTimeJSON = sql.NullString{String: field.ExpiryTime, Valid: true}
		}

		query := `INSERT INTO provider_metadata (field_name, owner, provider, consent_required, access_control_type, 
				  allow_list, description, expiry_time, metadata, created_at, updated_at) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT (field_name) DO UPDATE SET 
				  owner = EXCLUDED.owner, provider = EXCLUDED.provider, consent_required = EXCLUDED.consent_required, 
				  access_control_type = EXCLUDED.access_control_type, allow_list = EXCLUDED.allow_list, 
				  description = EXCLUDED.description, expiry_time = EXCLUDED.expiry_time, metadata = EXCLUDED.metadata, 
				  updated_at = EXCLUDED.updated_at`

		now := time.Now()
		_, err = s.db.Exec(query, fieldName, field.Owner, field.Provider, field.ConsentRequired,
			field.AccessControlType, allowListJSON, field.Description, expiryTimeJSON, metadataJSON, now, now)
		if err != nil {
			return fmt.Errorf("failed to import provider field %s: %w", fieldName, err)
		}
	}

	slog.Info("Imported provider metadata", "fields", len(metadata.Fields))
	return nil
}

// Allow List Management Methods

// AddConsumerToAllowList adds a consumer to the allow_list for a specific field
func (s *GrantsServiceDB) AddConsumerToAllowList(fieldName string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	// Get current field
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
			slog.Info("Updated consumer in allow_list", "fieldName", fieldName, "consumerId", req.ConsumerID)
			break
		}
	}

	// If consumer not found, add new entry
	if !found {
		newEntry := models.AllowListEntry{
			ConsumerID: req.ConsumerID,
			ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
		field.AllowList = append(field.AllowList, newEntry)
		slog.Info("Added consumer to allow_list", "fieldName", fieldName, "consumerId", req.ConsumerID)
	}

	// Update the field in database
	err = s.updateFieldAllowList(fieldName, field.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to update allow list: %w", err)
	}

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
			"updated_at":     time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// RemoveConsumerFromAllowList removes a consumer from the allow_list for a specific field
func (s *GrantsServiceDB) RemoveConsumerFromAllowList(fieldName, consumerID string) (*models.AllowListManagementResponse, error) {
	// Get current field
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

	// Update the field in database
	err = s.updateFieldAllowList(fieldName, newAllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to update allow list: %w", err)
	}

	slog.Info("Removed consumer from allow_list", "fieldName", fieldName, "consumerId", consumerID)

	response := &models.AllowListManagementResponse{
		Success:    true,
		Operation:  "consumer_removed_from_allow_list",
		FieldName:  fieldName,
		ConsumerID: consumerID,
		Data:       removedEntry,
		Metadata: map[string]interface{}{
			"removed_at": time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// GetAllowListForField retrieves the allow_list for a specific field
func (s *GrantsServiceDB) GetAllowListForField(fieldName string) (*models.AllowListListResponse, error) {
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
			"retrieved_at": time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// UpdateConsumerInAllowList updates an existing consumer in the allow_list
func (s *GrantsServiceDB) UpdateConsumerInAllowList(fieldName, consumerID string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	// Get current field
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

	// Update the field in database
	err = s.updateFieldAllowList(fieldName, field.AllowList)
	if err != nil {
		return nil, fmt.Errorf("failed to update allow list: %w", err)
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
			"updated_at":     time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// Helper function to update allow list for a field
func (s *GrantsServiceDB) updateFieldAllowList(fieldName string, allowList []models.AllowListEntry) error {
	allowListJSON, err := json.Marshal(allowList)
	if err != nil {
		return fmt.Errorf("failed to marshal allow list: %w", err)
	}

	query := `UPDATE provider_metadata SET allow_list = $1, updated_at = $2 WHERE field_name = $3`

	_, err = s.db.Exec(query, allowListJSON, time.Now(), fieldName)
	if err != nil {
		return fmt.Errorf("failed to update allow list: %w", err)
	}

	return nil
}
