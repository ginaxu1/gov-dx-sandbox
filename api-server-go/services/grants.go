package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type GrantsService struct {
	consumerGrants   *models.ConsumerGrantsData
	providerMetadata *models.ProviderMetadataData
	mutex            sync.RWMutex
}

func NewGrantsService() *GrantsService {
	return &GrantsService{
		consumerGrants: &models.ConsumerGrantsData{
			LegacyConsumerGrants: make(map[string]models.ConsumerGrant),
		},
		providerMetadata: &models.ProviderMetadataData{
			Fields: make(map[string]models.ProviderField),
		},
	}
}

// Consumer Grants Management

// GetAllConsumerGrants retrieves all consumer grants
func (s *GrantsService) GetAllConsumerGrants() (*models.ConsumerGrantsData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.consumerGrants, nil
}

// GetConsumerGrant retrieves a specific consumer grant
func (s *GrantsService) GetConsumerGrant(consumerID string) (*models.ConsumerGrant, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	grant, exists := s.consumerGrants.LegacyConsumerGrants[consumerID]
	if !exists {
		return nil, fmt.Errorf("consumer grant not found")
	}

	return &grant, nil
}

// CreateConsumerGrant creates a new consumer grant
func (s *GrantsService) CreateConsumerGrant(req models.CreateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)

	grant := models.ConsumerGrant{
		ConsumerID:     req.ConsumerID,
		ApprovedFields: req.ApprovedFields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.consumerGrants.LegacyConsumerGrants[req.ConsumerID] = grant

	slog.Info("Created consumer grant", "consumerId", req.ConsumerID, "fields", req.ApprovedFields)
	return &grant, nil
}

// UpdateConsumerGrant updates an existing consumer grant
func (s *GrantsService) UpdateConsumerGrant(consumerID string, req models.UpdateConsumerGrantRequest) (*models.ConsumerGrant, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	grant, exists := s.consumerGrants.LegacyConsumerGrants[consumerID]
	if !exists {
		return nil, fmt.Errorf("consumer grant not found")
	}

	grant.ApprovedFields = req.ApprovedFields
	grant.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	s.consumerGrants.LegacyConsumerGrants[consumerID] = grant

	slog.Info("Updated consumer grant", "consumerId", consumerID, "fields", req.ApprovedFields)
	return &grant, nil
}

// DeleteConsumerGrant deletes a consumer grant
func (s *GrantsService) DeleteConsumerGrant(consumerID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.consumerGrants.LegacyConsumerGrants[consumerID]
	if !exists {
		return fmt.Errorf("consumer grant not found")
	}

	delete(s.consumerGrants.LegacyConsumerGrants, consumerID)

	slog.Info("Deleted consumer grant", "consumerId", consumerID)
	return nil
}

// Provider Metadata Management

// GetAllProviderFields retrieves all provider fields
func (s *GrantsService) GetAllProviderFields() (*models.ProviderMetadataData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.providerMetadata, nil
}

// GetProviderField retrieves a specific provider field
func (s *GrantsService) GetProviderField(fieldName string) (*models.ProviderField, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	field, exists := s.providerMetadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("provider field not found")
	}

	return &field, nil
}

// CreateProviderField creates a new provider field
func (s *GrantsService) CreateProviderField(req models.CreateProviderFieldRequest) (*models.ProviderField, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

	s.providerMetadata.Fields[req.FieldName] = field

	slog.Info("Created provider field", "fieldName", req.FieldName, "owner", req.Owner, "provider", req.Provider)
	return &field, nil
}

// UpdateProviderField updates an existing provider field
func (s *GrantsService) UpdateProviderField(fieldName string, req models.UpdateProviderFieldRequest) (*models.ProviderField, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	field, exists := s.providerMetadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("provider field not found")
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

	s.providerMetadata.Fields[fieldName] = field

	slog.Info("Updated provider field", "fieldName", fieldName)
	return &field, nil
}

// DeleteProviderField deletes a provider field
func (s *GrantsService) DeleteProviderField(fieldName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.providerMetadata.Fields[fieldName]
	if !exists {
		return fmt.Errorf("provider field not found")
	}

	delete(s.providerMetadata.Fields, fieldName)

	slog.Info("Deleted provider field", "fieldName", fieldName)
	return nil
}

// Schema Conversion

// ConvertSchemaToProviderMetadata converts GraphQL SDL to provider metadata
func (s *GrantsService) ConvertSchemaToProviderMetadata(req models.SchemaConversionRequest) (*models.SchemaConversionResponse, error) {
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
func (s *GrantsService) ExportConsumerGrants() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return json.MarshalIndent(s.consumerGrants, "", "  ")
}

// ExportProviderMetadata exports provider metadata as JSON
func (s *GrantsService) ExportProviderMetadata() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return json.MarshalIndent(s.providerMetadata, "", "  ")
}

// ImportConsumerGrants imports consumer grants from JSON
func (s *GrantsService) ImportConsumerGrants(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var grants models.ConsumerGrantsData
	if err := json.Unmarshal(data, &grants); err != nil {
		return fmt.Errorf("failed to parse consumer grants JSON: %w", err)
	}

	s.consumerGrants = &grants
	slog.Info("Imported consumer grants", "count", len(grants.LegacyConsumerGrants))
	return nil
}

// ImportProviderMetadata imports provider metadata from JSON
func (s *GrantsService) ImportProviderMetadata(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var metadata models.ProviderMetadataData
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse provider metadata JSON: %w", err)
	}

	s.providerMetadata = &metadata
	slog.Info("Imported provider metadata", "fields", len(metadata.Fields))
	return nil
}

// Allow List Management Methods

// AddConsumerToAllowList adds a consumer to the allow_list for a specific field
func (s *GrantsService) AddConsumerToAllowList(fieldName string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current provider metadata
	metadata, err := s.GetAllProviderFields()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}

	// Check if field exists
	field, exists := metadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found", fieldName)
	}

	// Check if consumer already exists in allow_list
	for i, entry := range field.AllowList {
		if entry.ConsumerID == req.ConsumerID {
			// Update existing entry
			field.AllowList[i] = models.AllowListEntry{
				ConsumerID: req.ConsumerID,
				ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
				CreatedAt:  entry.CreatedAt, // Keep original creation time
			}
			slog.Info("Updated consumer in allow_list", "fieldName", fieldName, "consumerId", req.ConsumerID)
			break
		}
	}

	// If consumer not found, add new entry
	if !s.consumerExistsInAllowList(field.AllowList, req.ConsumerID) {
		newEntry := models.AllowListEntry{
			ConsumerID: req.ConsumerID,
			ExpiryTime: fmt.Sprintf("%d", req.ExpiresAt),
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
		field.AllowList = append(field.AllowList, newEntry)
		slog.Info("Added consumer to allow_list", "fieldName", fieldName, "consumerId", req.ConsumerID)
	}

	// Update the field in metadata
	metadata.Fields[fieldName] = field

	// Save updated metadata by exporting and importing
	exportData, err := s.ExportProviderMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to export provider metadata: %w", err)
	}
	if err := s.ImportProviderMetadata(exportData); err != nil {
		return nil, fmt.Errorf("failed to save provider metadata: %w", err)
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
func (s *GrantsService) RemoveConsumerFromAllowList(fieldName, consumerID string) (*models.AllowListManagementResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load current provider metadata
	metadata, err := s.loadProviderMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}

	// Check if field exists
	field, exists := metadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found", fieldName)
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
	field.AllowList = newAllowList
	metadata.Fields[fieldName] = field

	// Save updated metadata
	if err := s.saveProviderMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to save provider metadata: %w", err)
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
func (s *GrantsService) GetAllowListForField(fieldName string) (*models.AllowListListResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Load current provider metadata
	metadata, err := s.loadProviderMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}

	// Check if field exists
	field, exists := metadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found", fieldName)
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
func (s *GrantsService) UpdateConsumerInAllowList(fieldName, consumerID string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Load current provider metadata
	metadata, err := s.loadProviderMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}

	// Check if field exists
	field, exists := metadata.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found", fieldName)
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

	// Update the field in metadata
	metadata.Fields[fieldName] = field

	// Save updated metadata
	if err := s.saveProviderMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to save provider metadata: %w", err)
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

// Helper function to check if consumer exists in allow_list
func (s *GrantsService) consumerExistsInAllowList(allowList []models.AllowListEntry, consumerID string) bool {
	for _, entry := range allowList {
		if entry.ConsumerID == consumerID {
			return true
		}
	}
	return false
}
