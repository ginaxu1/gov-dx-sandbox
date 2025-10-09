package services

import (
	"fmt"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
)

// SchemaMappingService handles schema mapping operations
type SchemaMappingService struct {
	db *database.SchemaMappingDB
}

// NewSchemaMappingService creates a new schema mapping service
func NewSchemaMappingService(db *database.SchemaMappingDB) *SchemaMappingService {
	return &SchemaMappingService{
		db: db,
	}
}

// Unified Schema Operations

// CreateUnifiedSchema creates a new unified schema
func (s *SchemaMappingService) CreateUnifiedSchema(req *models.CreateUnifiedSchemaRequest) (*models.CreateUnifiedSchemaResponse, error) {
	// Validate request
	if err := s.validateCreateUnifiedSchemaRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if version already exists
	existingSchema, err := s.db.GetUnifiedSchemaByVersion(req.Version)
	if err == nil && existingSchema != nil {
		return nil, fmt.Errorf("unified schema version %s already exists", req.Version)
	}

	// Create new schema
	schema := &models.UnifiedSchema{
		ID:        database.GenerateID(),
		Version:   req.Version,
		SDL:       req.SDL,
		IsActive:  false, // New schemas are not active by default
		Notes:     req.Notes,
		CreatedBy: req.CreatedBy,
		Status:    "draft",
	}

	// Save to database
	if err := s.db.CreateUnifiedSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to save unified schema: %w", err)
	}

	// Create response
	response := &models.CreateUnifiedSchemaResponse{
		ID:        schema.ID,
		Version:   schema.Version,
		SDL:       schema.SDL,
		IsActive:  schema.IsActive,
		Notes:     schema.Notes,
		CreatedAt: schema.CreatedAt.Format(time.RFC3339),
		CreatedBy: schema.CreatedBy,
	}

	logger.Log.Info("Created unified schema", "version", schema.Version, "created_by", schema.CreatedBy)
	return response, nil
}

// GetUnifiedSchemas retrieves all unified schemas
func (s *SchemaMappingService) GetUnifiedSchemas() ([]*models.UnifiedSchema, error) {
	schemas, err := s.db.GetAllUnifiedSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to get unified schemas: %w", err)
	}

	return schemas, nil
}

// GetActiveUnifiedSchema retrieves the currently active unified schema
func (s *SchemaMappingService) GetActiveUnifiedSchema() (*models.UnifiedSchema, error) {
	schema, err := s.db.GetActiveUnifiedSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to get active unified schema: %w", err)
	}

	return schema, nil
}

// GetUnifiedSchemaByVersion retrieves a unified schema by version
func (s *SchemaMappingService) GetUnifiedSchemaByVersion(version string) (*models.UnifiedSchema, error) {
	schema, err := s.db.GetUnifiedSchemaByVersion(version)
	if err != nil {
		return nil, fmt.Errorf("failed to get unified schema by version: %w", err)
	}

	return schema, nil
}

// ActivateUnifiedSchema activates a specific unified schema version
func (s *SchemaMappingService) ActivateUnifiedSchema(version string) (*models.ActivateSchemaResponse, error) {
	// Check if schema exists
	_, err := s.db.GetUnifiedSchemaByVersion(version)
	if err != nil {
		return nil, fmt.Errorf("unified schema version %s not found: %w", version, err)
	}

	// Activate schema
	if err := s.db.ActivateUnifiedSchema(version); err != nil {
		return nil, fmt.Errorf("failed to activate schema: %w", err)
	}

	logger.Log.Info("Activated unified schema", "version", version)

	response := &models.ActivateSchemaResponse{
		Message: "Schema activated successfully",
	}

	return response, nil
}

// Provider Schema Operations

// GetProviderSchemas retrieves all active provider schemas
func (s *SchemaMappingService) GetProviderSchemas() (map[string]*models.ProviderSchema, error) {
	schemas, err := s.db.GetAllProviderSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider schemas: %w", err)
	}

	return schemas, nil
}

// CreateProviderSchema creates a new provider schema
func (s *SchemaMappingService) CreateProviderSchema(providerID, schemaName, sdl string) (*models.ProviderSchema, error) {
	schema := &models.ProviderSchema{
		ID:         database.GenerateID(),
		ProviderID: providerID,
		SchemaName: schemaName,
		SDL:        sdl,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := s.db.CreateProviderSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to create provider schema: %w", err)
	}

	logger.Log.Info("Created provider schema", "provider_id", providerID, "schema_name", schemaName)
	return schema, nil
}

// Field Mapping Operations

// CreateFieldMapping creates a new field mapping
func (s *SchemaMappingService) CreateFieldMapping(unifiedSchemaID string, req *models.FieldMappingRequest) (*models.FieldMappingResponse, error) {
	// Validate request
	if err := s.validateFieldMappingRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create field mapping
	mapping := &models.FieldMapping{
		ID:                database.GenerateID(),
		UnifiedSchemaID:   unifiedSchemaID,
		UnifiedFieldPath:  req.UnifiedFieldPath,
		ProviderID:        req.ProviderID,
		ProviderFieldPath: req.ProviderFieldPath,
		FieldType:         req.FieldType,
		IsRequired:        req.IsRequired,
		Directives:        req.Directives,
		CreatedAt:         time.Now(),
	}

	// Save to database
	if err := s.db.CreateFieldMapping(mapping); err != nil {
		return nil, fmt.Errorf("failed to save field mapping: %w", err)
	}

	// Create response
	response := &models.FieldMappingResponse{
		ID:                mapping.ID,
		UnifiedFieldPath:  mapping.UnifiedFieldPath,
		ProviderID:        mapping.ProviderID,
		ProviderFieldPath: mapping.ProviderFieldPath,
		FieldType:         mapping.FieldType,
		IsRequired:        mapping.IsRequired,
		Directives:        mapping.Directives,
		CreatedAt:         mapping.CreatedAt.Format(time.RFC3339),
	}

	logger.Log.Info("Created field mapping", "unified_field", mapping.UnifiedFieldPath, "provider_field", mapping.ProviderFieldPath)
	return response, nil
}

// GetFieldMappings retrieves all field mappings for a unified schema
func (s *SchemaMappingService) GetFieldMappings(unifiedSchemaID string) ([]*models.FieldMapping, error) {
	mappings, err := s.db.GetFieldMappingsBySchemaID(unifiedSchemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get field mappings: %w", err)
	}

	return mappings, nil
}

// UpdateFieldMapping updates an existing field mapping
func (s *SchemaMappingService) UpdateFieldMapping(mappingID string, req *models.FieldMappingRequest) (*models.FieldMappingResponse, error) {
	// Validate request
	if err := s.validateFieldMappingRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing mapping to update
	existingMappings, err := s.db.GetFieldMappingsBySchemaID("") // We need to find by mapping ID
	if err != nil {
		return nil, fmt.Errorf("failed to get existing mapping: %w", err)
	}

	var existingMapping *models.FieldMapping
	for _, mapping := range existingMappings {
		if mapping.ID == mappingID {
			existingMapping = mapping
			break
		}
	}

	if existingMapping == nil {
		return nil, fmt.Errorf("field mapping with id %s not found", mappingID)
	}

	// Update mapping
	existingMapping.ProviderID = req.ProviderID
	existingMapping.ProviderFieldPath = req.ProviderFieldPath
	existingMapping.FieldType = req.FieldType
	existingMapping.IsRequired = req.IsRequired
	existingMapping.Directives = req.Directives

	if err := s.db.UpdateFieldMapping(existingMapping); err != nil {
		return nil, fmt.Errorf("failed to update field mapping: %w", err)
	}

	// Create response
	response := &models.FieldMappingResponse{
		ID:                existingMapping.ID,
		UnifiedFieldPath:  existingMapping.UnifiedFieldPath,
		ProviderID:        existingMapping.ProviderID,
		ProviderFieldPath: existingMapping.ProviderFieldPath,
		FieldType:         existingMapping.FieldType,
		IsRequired:        existingMapping.IsRequired,
		Directives:        existingMapping.Directives,
		CreatedAt:         existingMapping.CreatedAt.Format(time.RFC3339),
	}

	logger.Log.Info("Updated field mapping", "mapping_id", mappingID)
	return response, nil
}

// DeleteFieldMapping deletes a field mapping
func (s *SchemaMappingService) DeleteFieldMapping(mappingID string) error {
	if err := s.db.DeleteFieldMapping(mappingID); err != nil {
		return fmt.Errorf("failed to delete field mapping: %w", err)
	}

	logger.Log.Info("Deleted field mapping", "mapping_id", mappingID)
	return nil
}

// Validation functions

func (s *SchemaMappingService) validateCreateUnifiedSchemaRequest(req *models.CreateUnifiedSchemaRequest) error {
	if req.Version == "" {
		return fmt.Errorf("version is required")
	}
	if req.SDL == "" {
		return fmt.Errorf("SDL is required")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}
	// Basic version format validation
	if len(req.Version) < 3 {
		return fmt.Errorf("invalid version format")
	}
	return nil
}

func (s *SchemaMappingService) validateFieldMappingRequest(req *models.FieldMappingRequest) error {
	if req.UnifiedFieldPath == "" {
		return fmt.Errorf("unified_field_path is required")
	}
	if req.ProviderID == "" {
		return fmt.Errorf("provider_id is required")
	}
	if req.ProviderFieldPath == "" {
		return fmt.Errorf("provider_field_path is required")
	}
	if req.FieldType == "" {
		return fmt.Errorf("field_type is required")
	}
	return nil
}
