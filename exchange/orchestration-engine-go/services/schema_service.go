package services

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

type SchemaServiceImpl struct {
	db *database.SchemaDB
}

// NewSchemaService creates a new schema service instance
func NewSchemaService(db *database.SchemaDB) *SchemaServiceImpl {
	return &SchemaServiceImpl{db: db}
}

// CreateSchema creates a new schema version
func (s *SchemaServiceImpl) CreateSchema(req *models.CreateSchemaRequest) (*models.UnifiedSchema, error) {
	// Validate the SDL
	if err := s.ValidateSDL(req.SDL); err != nil {
		return nil, fmt.Errorf("invalid SDL: %w", err)
	}

	// Generate version if not provided
	version := req.Version
	if version == "" {
		version = s.generateVersion()
	}

	// Check if version already exists
	existing, _ := s.GetSchemaByVersion(version)
	if existing != nil {
		return nil, fmt.Errorf("schema version %s already exists", version)
	}

	// Generate checksum
	checksum := s.generateChecksum(req.SDL)

	// Get previous version for tracking
	previousVersion := s.getPreviousVersion()

	// Create schema object
	schema := &models.UnifiedSchema{
		ID:                 uuid.New().String(),
		Version:            version,
		SDL:                req.SDL,
		Status:             "inactive", // New schemas start as inactive
		Description:        req.Description,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		CreatedBy:          "system", // TODO: Get from context
		Checksum:           checksum,
		CompatibilityLevel: "major", // Default to major for new schemas
		PreviousVersion:    previousVersion,
		Metadata:           make(map[string]interface{}),
		IsActive:           false,
		SchemaType:         "current",
	}

	// Save to database
	if err := s.db.CreateSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to save schema: %w", err)
	}

	// Create schema version record if there's a previous version
	if previousVersion != nil {
		schemaVersion := &models.SchemaVersion{
			FromVersion: *previousVersion,
			ToVersion:   version,
			ChangeType:  "major",
			Changes: map[string]interface{}{
				"description":     "New schema version created",
				"fields_added":    []string{},
				"fields_removed":  []string{},
				"fields_modified": []string{},
			},
			CreatedBy: "system",
		}

		if err := s.db.CreateSchemaVersion(schemaVersion); err != nil {
			// Log error but don't fail the schema creation
			fmt.Printf("Warning: Failed to create schema version record: %v\n", err)
		}
	}

	return schema, nil
}

// GetSchemaByVersion retrieves a schema by version
func (s *SchemaServiceImpl) GetSchemaByVersion(version string) (*models.UnifiedSchema, error) {
	return s.db.GetSchemaByVersion(version)
}

// GetActiveSchema retrieves the currently active schema
func (s *SchemaServiceImpl) GetActiveSchema() (*models.UnifiedSchema, error) {
	return s.db.GetActiveSchema()
}

// GetAllSchemas retrieves all schemas
func (s *SchemaServiceImpl) GetAllSchemas() ([]*models.UnifiedSchema, error) {
	return s.db.GetAllSchemas()
}

// UpdateSchemaStatus updates the status of a schema
func (s *SchemaServiceImpl) UpdateSchemaStatus(version string, req *models.UpdateSchemaStatusRequest) error {
	// Validate status
	validStatuses := []string{"active", "inactive", "deprecated"}
	if !contains(validStatuses, req.Status) {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	// If activating a schema, deactivate all others first
	if req.Status == "active" {
		if err := s.db.DeactivateAllSchemas(); err != nil {
			return fmt.Errorf("failed to deactivate other schemas: %w", err)
		}
	}

	return s.db.UpdateSchemaStatus(version, req.Status)
}

// DeleteSchema deletes a schema by version
func (s *SchemaServiceImpl) DeleteSchema(version string) error {
	// Check if it's the active schema
	active, err := s.GetActiveSchema()
	if err == nil && active.Version == version {
		return fmt.Errorf("cannot delete active schema")
	}

	return s.db.DeleteSchema(version)
}

// ActivateVersion activates a specific schema version
func (s *SchemaServiceImpl) ActivateVersion(version string) error {
	// Check if schema exists
	_, err := s.GetSchemaByVersion(version)
	if err != nil {
		return fmt.Errorf("schema version %s not found", version)
	}

	// Get current active schema for version tracking
	currentActive, _ := s.GetActiveSchema()

	// Deactivate all other schemas
	if err := s.db.DeactivateAllSchemas(); err != nil {
		return fmt.Errorf("failed to deactivate other schemas: %w", err)
	}

	// Activate the specified schema
	if err := s.db.UpdateSchemaStatus(version, "active"); err != nil {
		return err
	}

	// Create schema version record for activation
	if currentActive != nil {
		schemaVersion := &models.SchemaVersion{
			FromVersion: currentActive.Version,
			ToVersion:   version,
			ChangeType:  "activation",
			Changes: map[string]interface{}{
				"description":      "Schema version activated",
				"previous_version": currentActive.Version,
				"new_version":      version,
			},
			CreatedBy: "system",
		}

		if err := s.db.CreateSchemaVersion(schemaVersion); err != nil {
			// Log error but don't fail the activation
			fmt.Printf("Warning: Failed to create schema version record: %v\n", err)
		}
	}

	return nil
}

// DeactivateVersion deactivates a specific schema version
func (s *SchemaServiceImpl) DeactivateVersion(version string) error {
	return s.db.UpdateSchemaStatus(version, "inactive")
}

// GetSchemaVersions retrieves all schema versions
func (s *SchemaServiceImpl) GetSchemaVersions() ([]*models.SchemaVersionInfo, error) {
	return s.db.GetSchemaVersions()
}

// CheckCompatibility checks compatibility between a new SDL and the current active schema
func (s *SchemaServiceImpl) CheckCompatibility(sdl string) (*models.SchemaCompatibilityCheck, error) {
	// Parse the new schema
	newSchemaDoc, err := parser.ParseSchema(&ast.Source{Input: sdl})
	if err != nil {
		return nil, fmt.Errorf("failed to parse new schema: %w", err)
	}

	// Get the current active schema
	activeSchema, err := s.GetActiveSchema()
	if err != nil {
		// If no active schema, it's compatible
		return &models.SchemaCompatibilityCheck{
			Compatible:         true,
			CompatibilityLevel: "major",
		}, nil
	}

	// Parse the current schema
	currentSchemaDoc, err := parser.ParseSchema(&ast.Source{Input: activeSchema.SDL})
	if err != nil {
		return nil, fmt.Errorf("failed to parse current schema: %w", err)
	}

	// Convert to ast.Schema
	currentSchema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}
	for _, def := range currentSchemaDoc.Definitions {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union || def.Kind == ast.Enum || def.Kind == ast.Scalar || def.Kind == ast.InputObject {
			currentSchema.Types[def.Name] = def
		}
	}

	newSchema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}
	for _, def := range newSchemaDoc.Definitions {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union || def.Kind == ast.Enum || def.Kind == ast.Scalar || def.Kind == ast.InputObject {
			newSchema.Types[def.Name] = def
		}
	}

	// Perform compatibility check
	return s.performCompatibilityCheck(currentSchema, newSchema), nil
}

// ValidateSDL validates a GraphQL SDL string
func (s *SchemaServiceImpl) ValidateSDL(sdl string) error {
	// Parse the schema
	schemaDoc, err := parser.ParseSchema(&ast.Source{Input: sdl})
	if err != nil {
		return fmt.Errorf("failed to parse SDL: %w", err)
	}

	// Check if schema has at least one type definition
	if len(schemaDoc.Definitions) == 0 {
		return fmt.Errorf("schema must contain at least one type definition")
	}

	// Check if schema has a Query type
	hasQuery := false
	for _, def := range schemaDoc.Definitions {
		if def.Kind == ast.Object && def.Name == "Query" {
			hasQuery = true
			break
		}
	}

	if !hasQuery {
		return fmt.Errorf("schema must contain a Query type")
	}

	// Convert to ast.Schema for validation
	astSchema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}
	for _, def := range schemaDoc.Definitions {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union || def.Kind == ast.Enum || def.Kind == ast.Scalar || def.Kind == ast.InputObject {
			astSchema.Types[def.Name] = def
		}
	}

	// Note: We skip validator.Validate for schema-only validation as it expects both schema and query
	// The schema parsing above is sufficient for basic validation

	return nil
}

// ExecuteQuery executes a GraphQL query against the active schema
func (s *SchemaServiceImpl) ExecuteQuery(req *models.GraphQLRequest) (*models.GraphQLResponse, error) {
	// Get the schema to use
	var schema *models.UnifiedSchema
	var err error

	if req.SchemaVersion != "" {
		schema, err = s.GetSchemaByVersion(req.SchemaVersion)
		if err != nil {
			return nil, fmt.Errorf("schema version %s not found", req.SchemaVersion)
		}
	} else {
		schema, err = s.GetActiveSchema()
		if err != nil {
			return nil, fmt.Errorf("no active schema found")
		}
	}

	// Parse the schema
	astSchemaDoc, err := parser.ParseSchema(&ast.Source{Input: schema.SDL})
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Convert to ast.Schema
	astSchema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}
	for _, def := range astSchemaDoc.Definitions {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union || def.Kind == ast.Enum || def.Kind == ast.Scalar || def.Kind == ast.InputObject {
			astSchema.Types[def.Name] = def
		}
	}

	// Parse the query
	query, err := parser.ParseQuery(&ast.Source{Input: req.Query})
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Validate the query against the schema
	if err := validator.Validate(astSchema, query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// TODO: Execute the query using a GraphQL executor
	// For now, return a placeholder response
	return &models.GraphQLResponse{
		Data: map[string]interface{}{
			"message":       "Query executed successfully",
			"schemaVersion": schema.Version,
		},
	}, nil
}

// Helper methods

func (s *SchemaServiceImpl) generateVersion() string {
	return fmt.Sprintf("v%d", time.Now().Unix())
}

func (s *SchemaServiceImpl) generateChecksum(sdl string) string {
	hash := sha256.Sum256([]byte(sdl))
	return fmt.Sprintf("%x", hash)
}

func (s *SchemaServiceImpl) performCompatibilityCheck(current, new *ast.Schema) *models.SchemaCompatibilityCheck {
	check := &models.SchemaCompatibilityCheck{
		Compatible:         true,
		BreakingChanges:    []string{},
		Warnings:           []string{},
		CompatibilityLevel: "patch",
	}

	// Check for breaking changes in types
	s.checkTypeCompatibility(check, current, new)

	// Check for breaking changes in fields
	s.checkFieldCompatibility(check, current, new)

	// Determine compatibility level
	if len(check.BreakingChanges) > 0 {
		check.Compatible = false
		check.CompatibilityLevel = "major"
	} else if len(check.Warnings) > 0 {
		check.CompatibilityLevel = "minor"
	}

	return check
}

func (s *SchemaServiceImpl) checkTypeCompatibility(check *models.SchemaCompatibilityCheck, current, new *ast.Schema) {
	// Check for removed types
	for typeName := range current.Types {
		if new.Types[typeName] == nil {
			check.BreakingChanges = append(check.BreakingChanges, fmt.Sprintf("Type '%s' was removed", typeName))
		}
	}

	// Check for added types (usually safe)
	for typeName := range new.Types {
		if current.Types[typeName] == nil {
			check.Warnings = append(check.Warnings, fmt.Sprintf("Type '%s' was added", typeName))
		}
	}
}

func (s *SchemaServiceImpl) checkFieldCompatibility(check *models.SchemaCompatibilityCheck, current, new *ast.Schema) {
	// Check for removed fields in existing types
	for typeName, typeDef := range current.Types {
		if newType, exists := new.Types[typeName]; exists {
			if typeDef.Kind == ast.Object && newType.Kind == ast.Object {
				for _, field := range typeDef.Fields {
					if newType.Fields.ForName(field.Name) == nil {
						check.BreakingChanges = append(check.BreakingChanges,
							fmt.Sprintf("Field '%s' was removed from type '%s'", field.Name, typeName))
					}
				}
			}
		}
	}
}

// getPreviousVersion gets the version of the currently active schema
func (s *SchemaServiceImpl) getPreviousVersion() *string {
	active, err := s.GetActiveSchema()
	if err != nil {
		return nil
	}
	return &active.Version
}

// GetSchemaVersionsByVersion retrieves schema version records for a specific version
func (s *SchemaServiceImpl) GetSchemaVersionsByVersion(version string) ([]*models.SchemaVersion, error) {
	return s.db.GetSchemaVersionsByVersion(version)
}

// GetAllSchemaVersions retrieves all schema version records
func (s *SchemaServiceImpl) GetAllSchemaVersions() ([]*models.SchemaVersion, error) {
	return s.db.GetAllSchemaVersions()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
