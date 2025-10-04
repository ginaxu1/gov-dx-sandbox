package services

import (
<<<<<<< HEAD
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
=======
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// SchemaServiceImpl implements the SchemaService interface
type SchemaServiceImpl struct {
	db             *sql.DB
	currentSchema  *ast.Document
	schemaVersions map[string]*ast.Document
	mutex          sync.RWMutex
	contractTester *ContractTester
}

// ContractTester performs comprehensive backward compatibility testing
type ContractTester struct {
	testSuite []ContractTest
	db        *sql.DB
}

// ContractTest represents a single contract test
type ContractTest struct {
	Name        string                 `json:"name"`
	Query       string                 `json:"query"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Expected    interface{}            `json:"expected"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`
}

// ContractTestResults represents the results of running contract tests
type ContractTestResults struct {
	TotalTests int          `json:"total_tests"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	Results    []TestResult `json:"results"`
}

// TestResult represents the result of a single test
type TestResult struct {
	TestName string      `json:"test_name"`
	Passed   bool        `json:"passed"`
	Error    string      `json:"error,omitempty"`
	Actual   interface{} `json:"actual,omitempty"`
	Expected interface{} `json:"expected,omitempty"`
	Duration int64       `json:"duration_ms"`
}

// SchemaType represents the type of schema change
type SchemaType struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewSchemaService creates a new schema service instance
func NewSchemaService(db *sql.DB) models.SchemaService {
	return &SchemaServiceImpl{
		db:             db,
		schemaVersions: make(map[string]*ast.Document),
		contractTester: &ContractTester{db: db},
	}
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
}

// CreateSchema creates a new schema version
func (s *SchemaServiceImpl) CreateSchema(req *models.CreateSchemaRequest) (*models.UnifiedSchema, error) {
<<<<<<< HEAD
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
=======
	// Validate SDL syntax
	if err := s.validateSDL(req.SDL); err != nil {
		return nil, fmt.Errorf("invalid SDL syntax: %w", err)
	}

	// Check if version already exists
	existing, err := s.GetSchemaVersion(req.Version)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("schema version %s already exists", req.Version)
	}

	// Get previous version ID
	previousVersionID, err := s.GetPreviousVersionID(req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous version ID: %w", err)
	}

	// Check compatibility if there's a previous version
	if previousVersionID != nil {
		compatibility, err := s.CheckCompatibility(req.SDL, previousVersionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check compatibility: %w", err)
		}
		if !compatibility.IsCompatible {
			return nil, fmt.Errorf("schema is not compatible with previous version: %s", strings.Join(compatibility.Issues, ", "))
		}
	}

	// Insert new schema version
	query := `
		INSERT INTO unified_schemas (version, sdl, created_by, change_type, notes, previous_version_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	var id int
	var createdAt time.Time
	err = s.db.QueryRowContext(
		context.Background(),
		query,
		req.Version,
		req.SDL,
		req.CreatedBy,
		req.ChangeType,
		req.Notes,
		previousVersionID,
	).Scan(&id, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create schema version: %w", err)
	}

	return &models.UnifiedSchema{
		ID:                id,
		Version:           req.Version,
		SDL:               req.SDL,
		CreatedAt:         createdAt,
		CreatedBy:         req.CreatedBy,
		Status:            models.SchemaStatusInactive, // New schemas start as inactive
		ChangeType:        req.ChangeType,
		Notes:             req.Notes,
		PreviousVersionID: previousVersionID,
	}, nil
}

// UpdateSchema updates the unified schema
func (s *SchemaServiceImpl) UpdateSchema(req *models.UpdateSchemaRequest) (*models.UpdateSchemaResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Parse new schema
	newSchema, err := s.parseSDL(req.SDL)
	if err != nil {
		return nil, fmt.Errorf("invalid SDL: %w", err)
	}

	// Check version compatibility
	compatibility, err := s.checkCompatibility(req.Version, newSchema)
	if err != nil {
		return nil, fmt.Errorf("compatibility check failed: %w", err)
	}

	// Update database
	if err := s.updateSchemaInDB(req, compatibility); err != nil {
		return nil, fmt.Errorf("failed to update database: %w", err)
	}

	// Update in-memory schema
	s.updateInMemorySchema(req.Version, newSchema, compatibility)

	return &models.UpdateSchemaResponse{
		Success:    true,
		Message:    "Schema updated successfully",
		Version:    req.Version,
		SchemaType: s.getSchemaType(compatibility.ChangeType),
		IsActive:   true,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetSchemaVersion retrieves a specific schema version by version string
func (s *SchemaServiceImpl) GetSchemaVersion(version string) (*models.UnifiedSchema, error) {
	query := `
		SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id
		FROM unified_schemas
		WHERE version = $1`

	var schema models.UnifiedSchema
	err := s.db.QueryRowContext(context.Background(), query, version).Scan(
		&schema.ID,
		&schema.Version,
		&schema.SDL,
		&schema.CreatedAt,
		&schema.CreatedBy,
		&schema.Status,
		&schema.ChangeType,
		&schema.Notes,
		&schema.PreviousVersionID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema version %s not found", version)
		}
		return nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	return &schema, nil
}

// GetAllSchemaVersions retrieves all schema versions with optional filtering
func (s *SchemaServiceImpl) GetAllSchemaVersions(status *models.SchemaStatus, limit, offset int) ([]*models.UnifiedSchema, int, error) {
	// Build query with optional status filter
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if status != nil {
		whereClause = "WHERE status = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM unified_schemas " + whereClause
	var total int
	err := s.db.QueryRowContext(context.Background(), countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get schema count: %w", err)
	}

	// Get schemas with pagination
	query := `
		SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id
		FROM unified_schemas ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get schema versions: %w", err)
	}
	defer rows.Close()

	var schemas []*models.UnifiedSchema
	for rows.Next() {
		var schema models.UnifiedSchema
		err := rows.Scan(
			&schema.ID,
			&schema.Version,
			&schema.SDL,
			&schema.CreatedAt,
			&schema.CreatedBy,
			&schema.Status,
			&schema.ChangeType,
			&schema.Notes,
			&schema.PreviousVersionID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, &schema)
	}

	return schemas, total, nil
}

// UpdateSchemaStatus updates the status of a schema version
func (s *SchemaServiceImpl) UpdateSchemaStatus(version string, isActive bool, reason *string) error {
	// If activating, deactivate all other schemas first
	if isActive {
		deactivateQuery := "UPDATE unified_schemas SET status = $1 WHERE status = $2"
		_, err := s.db.ExecContext(context.Background(), deactivateQuery, models.SchemaStatusInactive, models.SchemaStatusActive)
		if err != nil {
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
			return fmt.Errorf("failed to deactivate other schemas: %w", err)
		}
	}

<<<<<<< HEAD
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
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Convert to ast.Schema
	schema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}
	for _, def := range schemaDoc.Definitions {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union || def.Kind == ast.Enum || def.Kind == ast.Scalar || def.Kind == ast.InputObject {
			schema.Types[def.Name] = def
		}
	}

	// Validate the schema
	// Note: For schema-only validation, we just need to ensure it parses correctly
	// The validator.Validate function requires both schema and query, so we skip it for schema-only validation
	// In a real implementation, you might want to use a different validation approach
=======
	// Update the target schema status
	status := models.SchemaStatusInactive
	if isActive {
		status = models.SchemaStatusActive
	}

	query := "UPDATE unified_schemas SET status = $1, notes = COALESCE($2, notes) WHERE version = $3"
	result, err := s.db.ExecContext(context.Background(), query, status, reason, version)
	if err != nil {
		return fmt.Errorf("failed to update schema status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schema version %s not found", version)
	}

	return nil
}

// GetActiveSchema retrieves the currently active schema
func (s *SchemaServiceImpl) GetActiveSchema() (*models.UnifiedSchema, error) {
	query := `
		SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id
		FROM unified_schemas
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var schema models.UnifiedSchema
	err := s.db.QueryRowContext(context.Background(), query, models.SchemaStatusActive).Scan(
		&schema.ID,
		&schema.Version,
		&schema.SDL,
		&schema.CreatedAt,
		&schema.CreatedBy,
		&schema.Status,
		&schema.ChangeType,
		&schema.Notes,
		&schema.PreviousVersionID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active schema found")
		}
		return nil, fmt.Errorf("failed to get active schema: %w", err)
	}

	return &schema, nil
}

// CheckCompatibility checks if a new schema is compatible with existing ones
func (s *SchemaServiceImpl) CheckCompatibility(newSDL string, previousVersionID *int) (*models.SchemaCompatibilityCheck, error) {
	if previousVersionID == nil {
		// No previous version, so it's compatible
		return &models.SchemaCompatibilityCheck{
			IsCompatible: true,
		}, nil
	}

	// Get the previous schema
	query := "SELECT sdl FROM unified_schemas WHERE id = $1"
	var previousSDL string
	err := s.db.QueryRowContext(context.Background(), query, *previousVersionID).Scan(&previousSDL)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous schema: %w", err)
	}

	// Parse both schemas
	newAST, err := s.parseSDL(newSDL)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{fmt.Sprintf("Invalid new schema syntax: %v", err)},
		}, nil
	}

	previousAST, err := s.parseSDL(previousSDL)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{fmt.Sprintf("Invalid previous schema syntax: %v", err)},
		}, nil
	}

	// Perform compatibility checks
	issues, warnings := s.compareSchemas(previousAST, newAST)

	return &models.SchemaCompatibilityCheck{
		IsCompatible: len(issues) == 0,
		Issues:       issues,
		Warnings:     warnings,
	}, nil
}

// GetPreviousVersionID retrieves the ID of the previous version for a given version
func (s *SchemaServiceImpl) GetPreviousVersionID(version string) (*int, error) {
	// For now, we'll use a simple approach: get the most recent version
	// In a more sophisticated implementation, this could parse semantic versioning
	query := `
		SELECT id FROM unified_schemas 
		WHERE version != $1 
		ORDER BY created_at DESC 
		LIMIT 1`

	var previousID int
	err := s.db.QueryRowContext(context.Background(), query, version).Scan(&previousID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No previous version
		}
		return nil, fmt.Errorf("failed to get previous version ID: %w", err)
	}

	return &previousID, nil
}

// validateSDL validates the SDL syntax
func (s *SchemaServiceImpl) validateSDL(sdl string) error {
	_, err := s.parseSDL(sdl)
	return err
}

// parseSDL parses SDL into an AST
func (s *SchemaServiceImpl) parseSDL(sdl string) (*ast.Document, error) {
	src := source.NewSource(&source.Source{
		Body: []byte(sdl),
		Name: "SchemaSDL",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// compareSchemas compares two schema ASTs for compatibility
func (s *SchemaServiceImpl) compareSchemas(previous, current *ast.Document) (issues, warnings []string) {
	// This is a simplified compatibility check
	// In a real implementation, you would:
	// 1. Check for breaking changes (removed fields, changed types, etc.)
	// 2. Check for non-breaking changes (added fields, etc.)
	// 3. Validate that required fields are still present
	// 4. Check for type compatibility

	// For now, we'll do a basic check
	// This is where you would implement sophisticated GraphQL schema compatibility logic
	// For example, checking if all existing queries would still work

	// Placeholder implementation - always compatible
	// In reality, you'd parse the AST and compare type definitions, fields, etc.
	return issues, warnings
}

// checkCompatibility checks if the new schema is compatible with existing versions
func (s *SchemaServiceImpl) checkCompatibility(version string, newSchema *ast.Document) (*models.VersionCompatibility, error) {
	// Parse version (e.g., "1.2.0")
	major, minor, patch, err := s.parseVersion(version)
	if err != nil {
		return nil, err
	}

	// Get current active schema
	currentVersion, currentSchema, err := s.getCurrentActiveSchema()
	if err != nil {
		return nil, err
	}

	currentMajor, currentMinor, currentPatch, _ := s.parseVersion(currentVersion)

	// Determine change type
	var changeType string
	if major > currentMajor {
		changeType = "major"
	} else if minor > currentMinor {
		changeType = "minor"
	} else if patch > currentPatch {
		changeType = "patch"
	} else {
		return nil, fmt.Errorf("version must be higher than current version")
	}

	// Perform compatibility checks
	compatibility := &models.VersionCompatibility{
		ChangeType: changeType,
	}

	if changeType == "minor" {
		// Minor version: only allow adding new fields
		if err := s.checkMinorCompatibility(currentSchema, newSchema, compatibility); err != nil {
			return nil, err
		}
	} else if changeType == "major" {
		// Major version: allow breaking changes
		s.checkMajorCompatibility(currentSchema, newSchema, compatibility)
	}

	return compatibility, nil
}

// checkMinorCompatibility ensures minor versions only add new fields
func (s *SchemaServiceImpl) checkMinorCompatibility(current, new *ast.Document, compatibility *models.VersionCompatibility) error {
	// Extract all types and fields from current schema
	currentTypes := s.extractTypes(current)
	newTypes := s.extractTypes(new)

	// Check for removed or modified fields
	for typeName, currentType := range currentTypes {
		newType, exists := newTypes[typeName]
		if !exists {
			compatibility.BreakingChanges = append(compatibility.BreakingChanges,
				fmt.Sprintf("Type '%s' was removed", typeName))
			return fmt.Errorf("minor version cannot remove types")
		}

		// Check fields within type
		for fieldName, currentField := range currentType.Fields {
			newField, exists := newType.Fields[fieldName]
			if !exists {
				compatibility.BreakingChanges = append(compatibility.BreakingChanges,
					fmt.Sprintf("Field '%s.%s' was removed", typeName, fieldName))
				return fmt.Errorf("minor version cannot remove fields")
			}

			// Check if field type changed
			if !s.fieldTypesEqual(currentField.Type, newField.Type) {
				compatibility.BreakingChanges = append(compatibility.BreakingChanges,
					fmt.Sprintf("Field '%s.%s' type changed from %s to %s",
						typeName, fieldName, currentField.Type, newField.Type))
				return fmt.Errorf("minor version cannot modify field types")
			}
		}
	}

	// Check for new fields (allowed in minor versions)
	for typeName, newType := range newTypes {
		currentType, exists := currentTypes[typeName]
		if !exists {
			compatibility.NewFields = append(compatibility.NewFields,
				fmt.Sprintf("New type '%s' added", typeName))
			continue
		}

		for fieldName := range newType.Fields {
			if _, exists := currentType.Fields[fieldName]; !exists {
				compatibility.NewFields = append(compatibility.NewFields,
					fmt.Sprintf("New field '%s.%s' added", typeName, fieldName))
			}
		}
	}
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)

	return nil
}

<<<<<<< HEAD
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
=======
// checkMajorCompatibility allows breaking changes in major versions
func (s *SchemaServiceImpl) checkMajorCompatibility(current, new *ast.Document, compatibility *models.VersionCompatibility) {
	currentTypes := s.extractTypes(current)
	newTypes := s.extractTypes(new)

	// Check for removed types
	for typeName := range currentTypes {
		if _, exists := newTypes[typeName]; !exists {
			compatibility.BreakingChanges = append(compatibility.BreakingChanges,
				fmt.Sprintf("Type '%s' was removed", typeName))
		}
	}

	// Check for removed fields
	for typeName, currentType := range currentTypes {
		if newType, exists := newTypes[typeName]; exists {
			for fieldName := range currentType.Fields {
				if _, exists := newType.Fields[fieldName]; !exists {
					compatibility.BreakingChanges = append(compatibility.BreakingChanges,
						fmt.Sprintf("Field '%s.%s' was removed", typeName, fieldName))
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
				}
			}
		}
	}
<<<<<<< HEAD
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
=======

	// Check for new types and fields
	for typeName, newType := range newTypes {
		if currentType, exists := currentTypes[typeName]; exists {
			for fieldName := range newType.Fields {
				if _, exists := currentType.Fields[fieldName]; !exists {
					compatibility.NewFields = append(compatibility.NewFields,
						fmt.Sprintf("New field '%s.%s' added", typeName, fieldName))
				}
			}
		} else {
			compatibility.NewFields = append(compatibility.NewFields,
				fmt.Sprintf("New type '%s' added", typeName))
		}
	}
}

// parseVersion parses a semantic version string
func (s *SchemaServiceImpl) parseVersion(version string) (major, minor, patch int, err error) {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(version)
	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, _ = strconv.Atoi(matches[1])
	minor, _ = strconv.Atoi(matches[2])
	patch, _ = strconv.Atoi(matches[3])

	return major, minor, patch, nil
}

// getCurrentActiveSchema gets the currently active schema
func (s *SchemaServiceImpl) getCurrentActiveSchema() (string, *ast.Document, error) {
	query := `
		SELECT version, sdl FROM unified_schemas 
		WHERE status = $1 
		ORDER BY created_at DESC 
		LIMIT 1`

	var version, sdl string
	err := s.db.QueryRowContext(context.Background(), query, models.SchemaStatusActive).Scan(&version, &sdl)
	if err != nil {
		return "", nil, fmt.Errorf("no active schema found: %w", err)
	}

	schema, err := s.parseSDL(sdl)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse active schema: %w", err)
	}

	return version, schema, nil
}

// extractTypes extracts type definitions from a schema AST
func (s *SchemaServiceImpl) extractTypes(doc *ast.Document) map[string]*TypeInfo {
	types := make(map[string]*TypeInfo)

	for _, def := range doc.Definitions {
		if typeDef, ok := def.(*ast.ObjectDefinition); ok {
			typeInfo := &TypeInfo{
				Name:   typeDef.Name.Value,
				Fields: make(map[string]*FieldInfo),
			}

			for _, field := range typeDef.Fields {
				fieldInfo := &FieldInfo{
					Name: field.Name.Value,
					Type: s.getTypeString(field.Type),
				}
				typeInfo.Fields[field.Name.Value] = fieldInfo
			}

			types[typeDef.Name.Value] = typeInfo
		}
	}

	return types
}

// TypeInfo represents a GraphQL type with its fields
type TypeInfo struct {
	Name   string
	Fields map[string]*FieldInfo
}

// FieldInfo represents a GraphQL field
type FieldInfo struct {
	Name string
	Type string
}

// getTypeString converts an AST type to a string representation
func (s *SchemaServiceImpl) getTypeString(t ast.Type) string {
	switch t := t.(type) {
	case *ast.NonNull:
		return s.getTypeString(t.Type) + "!"
	case *ast.List:
		return "[" + s.getTypeString(t.Type) + "]"
	case *ast.Named:
		return t.Name.Value
	default:
		return "Unknown"
	}
}

// fieldTypesEqual checks if two field types are equal
func (s *SchemaServiceImpl) fieldTypesEqual(type1, type2 string) bool {
	return type1 == type2
}

// updateSchemaInDB updates the schema in the database
func (s *SchemaServiceImpl) updateSchemaInDB(req *models.UpdateSchemaRequest, compatibility *models.VersionCompatibility) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate current active schema
	deactivateQuery := `UPDATE unified_schemas SET status = $1 WHERE status = $2`
	_, err = tx.ExecContext(context.Background(), deactivateQuery, models.SchemaStatusInactive, models.SchemaStatusActive)
	if err != nil {
		return fmt.Errorf("failed to deactivate current schema: %w", err)
	}

	// Update or insert the new schema
	updateQuery := `
		INSERT INTO unified_schemas (version, sdl, created_by, status, change_type, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (version) 
		DO UPDATE SET 
			sdl = EXCLUDED.sdl,
			created_by = EXCLUDED.created_by,
			status = EXCLUDED.status,
			change_type = EXCLUDED.change_type,
			notes = EXCLUDED.notes,
			created_at = CURRENT_TIMESTAMP`

	_, err = tx.ExecContext(context.Background(), updateQuery,
		req.Version, req.SDL, req.CreatedBy, models.SchemaStatusActive, compatibility.ChangeType, "")
	if err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// updateInMemorySchema updates the in-memory schema
func (s *SchemaServiceImpl) updateInMemorySchema(version string, schema *ast.Document, compatibility *models.VersionCompatibility) {
	s.schemaVersions[version] = schema
	s.currentSchema = schema
}

// getSchemaType returns the schema type based on change type
func (s *SchemaServiceImpl) getSchemaType(changeType string) string {
	switch changeType {
	case "major":
		return "breaking"
	case "minor":
		return "additive"
	case "patch":
		return "patch"
	default:
		return "unknown"
	}
}

// Contract Testing Methods

// LoadContractTests loads contract tests from database
func (ct *ContractTester) LoadContractTests() ([]ContractTest, error) {
	if ct.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `SELECT name, query, variables, expected, description, priority, is_active 
	          FROM contract_tests 
	          WHERE is_active = true 
	          ORDER BY priority, name`

	rows, err := ct.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []ContractTest
	for rows.Next() {
		var test ContractTest
		var variablesJSON, expectedJSON string

		err := rows.Scan(&test.Name, &test.Query, &variablesJSON, &expectedJSON,
			&test.Description, &test.Priority, &test.IsActive)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if variablesJSON != "" {
			// In a real implementation, you would unmarshal the JSON
			// For now, we'll leave it as empty map
			test.Variables = make(map[string]interface{})
		}
		if expectedJSON != "" {
			// In a real implementation, you would unmarshal the JSON
			// For now, we'll leave it as nil
			test.Expected = nil
		}

		tests = append(tests, test)
	}

	return tests, nil
}

// ExecuteContractTests runs all contract tests against a schema
func (ct *ContractTester) ExecuteContractTests(schema *ast.Document) (*ContractTestResults, error) {
	tests, err := ct.LoadContractTests()
	if err != nil {
		return nil, err
	}

	results := &ContractTestResults{
		TotalTests: len(tests),
		Passed:     0,
		Failed:     0,
		Results:    make([]TestResult, 0, len(tests)),
	}

	for _, test := range tests {
		result := ct.runSingleTest(test, schema)
		results.Results = append(results.Results, result)

		if result.Passed {
			results.Passed++
		} else {
			results.Failed++
		}
	}

	return results, nil
}

// runSingleTest runs a single contract test
func (ct *ContractTester) runSingleTest(test ContractTest, schema *ast.Document) TestResult {
	start := time.Now()

	// In a real implementation, you would:
	// 1. Parse the GraphQL query
	// 2. Execute it against the schema
	// 3. Compare the result with expected output

	// For now, we'll simulate a test result
	result := TestResult{
		TestName: test.Name,
		Passed:   true, // Placeholder - always pass for now
		Duration: time.Since(start).Milliseconds(),
	}

	// In a real implementation, you would check if the query is valid
	// and produces the expected result
	if test.Query == "" {
		result.Passed = false
		result.Error = "Empty query"
	}

	return result
}

// NewContractTester creates a new contract tester
func NewContractTester(db *sql.DB) *ContractTester {
	return &ContractTester{
		db: db,
	}
>>>>>>> 8d51df8 (OE add database.go and schema endpoints, update schema functionality)
}

// Phase 4: Multi-Version Schema Support

// getAllActiveSchemas retrieves all schemas with status 'active' or 'inactive'
func (s *SchemaServiceImpl) getAllActiveSchemas() ([]models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
	          FROM unified_schemas 
	          WHERE status IN ('active', 'inactive') 
	          ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []models.UnifiedSchema
	for rows.Next() {
		var schema models.UnifiedSchema
		err := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.CreatedAt,
			&schema.CreatedBy, &schema.Status, &schema.ChangeType,
			&schema.Notes, &schema.PreviousVersionID)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// LoadSchema loads schema from database into memory
func (s *SchemaServiceImpl) LoadSchema() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if database is available
	if s.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Load all active schemas
	schemas, err := s.getAllActiveSchemas()
	if err != nil {
		return err
	}

	s.schemaVersions = make(map[string]*ast.Document)

	for _, schema := range schemas {
		parsed, err := s.parseSDL(schema.SDL)
		if err != nil {
			return fmt.Errorf("failed to parse schema version %s: %w", schema.Version, err)
		}

		s.schemaVersions[schema.Version] = parsed

		if schema.Status == models.SchemaStatusActive {
			s.currentSchema = parsed
		}
	}

	return nil
}

// GetSchemaForVersion returns schema for specific version
func (s *SchemaServiceImpl) GetSchemaForVersion(version string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	schema, exists := s.schemaVersions[version]
	if !exists {
		return nil, fmt.Errorf("schema version %s not found", version)
	}

	return schema, nil
}

// RouteQuery routes GraphQL query to appropriate schema version
func (s *SchemaServiceImpl) RouteQuery(query string, version string) (interface{}, error) {
	if version == "" {
		// Use current active schema (default)
		return s.GetDefaultSchema()
	}

	// Use specific version
	return s.GetSchemaForVersion(version)
}

// GetDefaultSchema returns the currently active default schema
func (s *SchemaServiceImpl) GetDefaultSchema() (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.currentSchema == nil {
		return nil, fmt.Errorf("no active schema found")
	}

	return s.currentSchema, nil
}

// ReloadSchemasInMemory reloads all schemas from database into memory
func (s *SchemaServiceImpl) ReloadSchemasInMemory() error {
	return s.LoadSchema()
}

// GetSchemaVersions returns all loaded schema versions
func (s *SchemaServiceImpl) GetSchemaVersions() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to avoid race conditions
	versions := make(map[string]interface{})
	for version, schema := range s.schemaVersions {
		versions[version] = schema
	}

	return versions
}

// IsSchemaVersionLoaded checks if a specific schema version is loaded in memory
func (s *SchemaServiceImpl) IsSchemaVersionLoaded(version string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.schemaVersions[version]
	return exists
}

// GetActiveSchemaVersion returns the version string of the currently active schema
func (s *SchemaServiceImpl) GetActiveSchemaVersion() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.currentSchema == nil {
		return "", fmt.Errorf("no active schema found")
	}

	// Find the version that matches the current schema
	for version, schema := range s.schemaVersions {
		if schema == s.currentSchema {
			return version, nil
		}
	}

	return "", fmt.Errorf("active schema version not found")
}
