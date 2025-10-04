package services

import (
<<<<<<< HEAD
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
=======
>>>>>>> e62b19e (Clean up and unit tests)
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/vektah/gqlparser/v2/ast"
)

// SchemaService defines the interface for schema management
// This interface extends models.SchemaService with additional methods
type SchemaService interface {
	models.SchemaService

	// Additional methods for multi-version support
	LoadSchema() error
	GetSchemaForVersion(version string) (interface{}, error)
	RouteQuery(query string, version string) (interface{}, error)
	GetDefaultSchema() (interface{}, error)
	ReloadSchemasInMemory() error
	GetSchemaVersions() map[string]interface{}
	IsSchemaVersionLoaded(version string) bool
	GetActiveSchemaVersion() (string, error)

	// Configuration methods
	GetConfiguration() *configs.SchemaConfig
	ReloadConfiguration() error
}

// SchemaServiceImpl implements the SchemaService interface
type SchemaServiceImpl struct {
	db             *sql.DB
	currentSchema  *ast.QueryDocument
	schemaVersions map[string]*ast.QueryDocument
	mutex          sync.RWMutex
	contractTester *ContractTester
	config         *configs.SchemaConfig
}

// NewSchemaService creates a new schema service
func NewSchemaService(db *sql.DB) models.SchemaService {
	config := configs.LoadSchemaConfig()
	return &SchemaServiceImpl{
		db:             db,
		schemaVersions: make(map[string]*ast.QueryDocument),
		contractTester: &ContractTester{db: db},
		config:         config,
	}
}

// ContractTester performs comprehensive backward compatibility testing
type ContractTester struct {
	testSuite []ContractTest
	db        *sql.DB
}

// ContractTest represents a single contract test
type ContractTest struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Query       string                 `json:"query"`
	Variables   map[string]interface{} `json:"variables"`
	Expected    map[string]interface{} `json:"expected"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`
}

// NewContractTester creates a new contract tester
func NewContractTester(db *sql.DB) *ContractTester {
	return &ContractTester{
		db: db,
	}
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
<<<<<<< HEAD
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
=======
	TestName string                 `json:"test_name"`
	Passed   bool                   `json:"passed"`
	Error    string                 `json:"error,omitempty"`
	Actual   map[string]interface{} `json:"actual,omitempty"`
	Expected map[string]interface{} `json:"expected,omitempty"`
	Duration time.Duration          `json:"duration"`
>>>>>>> e62b19e (Clean up and unit tests)
}

// CreateSchema creates a new schema version
func (s *SchemaServiceImpl) CreateSchema(req *models.CreateSchemaRequest) (*models.UnifiedSchema, error) {
<<<<<<< HEAD
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
=======
>>>>>>> e62b19e (Clean up and unit tests)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check version limits
	if err := s.checkVersionLimits(); err != nil {
		return nil, err
	}

	// Parse the SDL
	parsedSchema, err := s.parseSDL(req.SDL)
	if err != nil {
		return nil, fmt.Errorf("invalid SDL: %w", err)
	}

	// Check compatibility if enabled
	if s.config.IsCompatibilityCheckEnabled() {
		compatibility, err := s.checkCompatibility(req.Version, parsedSchema)
		if err != nil {
			return nil, fmt.Errorf("compatibility check failed: %w", err)
		}
		req.ChangeType = models.VersionChangeType(compatibility.ChangeType)
	}

	// Create schema in database
	schema := &models.UnifiedSchema{
		Version:    req.Version,
		SDL:        req.SDL,
		CreatedAt:  time.Now(),
		CreatedBy:  req.CreatedBy,
		Status:     models.SchemaStatusInactive,
		ChangeType: req.ChangeType,
		Notes:      req.Notes,
	}

	// Insert into database
	query := `
		INSERT INTO unified_schemas (version, sdl, created_by, status, change_type, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err = s.db.QueryRow(query, schema.Version, schema.SDL, schema.CreatedBy, schema.Status, schema.ChangeType, schema.Notes).Scan(&schema.ID, &schema.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Auto-activate if enabled
	if s.config.IsAutoActivateEnabled() {
		if err := s.activateSchemaVersion(schema.Version); err != nil {
			return nil, fmt.Errorf("failed to auto-activate schema: %w", err)
		}
		schema.Status = models.SchemaStatusActive
	}

	// Load into memory
	s.schemaVersions[schema.Version] = parsedSchema
	if schema.Status == models.SchemaStatusActive {
		s.currentSchema = parsedSchema
	}

	// Cleanup old versions if needed
	if err := s.cleanupOldVersions(); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to cleanup old versions: %v\n", err)
	}

	return schema, nil
}

// UpdateSchemaStatus updates the status of a schema version
func (s *SchemaServiceImpl) UpdateSchemaStatus(version string, isActive bool, reason *string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var newStatus models.SchemaStatus
	if isActive {
<<<<<<< HEAD
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
=======
		newStatus = models.SchemaStatusActive
	} else {
		newStatus = models.SchemaStatusInactive
	}

	// Update database
	query := `UPDATE unified_schemas SET status = $1, updated_at = NOW() WHERE version = $2`
	result, err := s.db.Exec(query, newStatus, version)
>>>>>>> e62b19e (Clean up and unit tests)
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

	// If activating, deactivate all other versions
	if isActive {
		deactivateQuery := `UPDATE unified_schemas SET status = $1, updated_at = NOW() WHERE version != $2`
		_, err = s.db.Exec(deactivateQuery, models.SchemaStatusInactive, version)
		if err != nil {
			return fmt.Errorf("failed to deactivate other schemas: %w", err)
		}

		// Update current schema in memory
		if schema, exists := s.schemaVersions[version]; exists {
			s.currentSchema = schema
		}
	}

	return nil
}

// GetAllSchemaVersions retrieves all schema versions
func (s *SchemaServiceImpl) GetAllSchemaVersions(status *models.SchemaStatus, limit, offset int) ([]*models.UnifiedSchema, int, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
			FROM unified_schemas WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
			FROM unified_schemas ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*models.UnifiedSchema
	for rows.Next() {
		var schema models.UnifiedSchema
		err := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.CreatedAt, &schema.CreatedBy,
			&schema.Status, &schema.ChangeType, &schema.Notes, &schema.PreviousVersionID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, &schema)
	}

	// Get total count
	var countQuery string
	var countArgs []interface{}
	if status != nil {
		countQuery = `SELECT COUNT(*) FROM unified_schemas WHERE status = $1`
		countArgs = []interface{}{*status}
	} else {
		countQuery = `SELECT COUNT(*) FROM unified_schemas`
		countArgs = []interface{}{}
	}

	var total int
	err = s.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return schemas, total, nil
}

// GetSchemaVersion retrieves a specific schema version
func (s *SchemaServiceImpl) GetSchemaVersion(version string) (*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
		FROM unified_schemas WHERE version = $1`

	var schema models.UnifiedSchema
	err := s.db.QueryRow(query, version).Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.CreatedAt,
		&schema.CreatedBy, &schema.Status, &schema.ChangeType, &schema.Notes, &schema.PreviousVersionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema version %s not found", version)
		}
		return nil, fmt.Errorf("failed to get schema version: %w", err)
	}

	return &schema, nil
}

// GetActiveSchema returns the currently active schema
func (s *SchemaServiceImpl) GetActiveSchema() (*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
		FROM unified_schemas WHERE status = $1 ORDER BY created_at DESC LIMIT 1`

	var schema models.UnifiedSchema
	err := s.db.QueryRow(query, models.SchemaStatusActive).Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.CreatedAt,
		&schema.CreatedBy, &schema.Status, &schema.ChangeType, &schema.Notes, &schema.PreviousVersionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active schema found")
		}
		return nil, fmt.Errorf("failed to get active schema: %w", err)
	}

	return &schema, nil
}

// GetCurrentActiveSchema returns the currently active schema as QueryDocument
func (s *SchemaServiceImpl) GetCurrentActiveSchema() (*ast.QueryDocument, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.currentSchema == nil {
		return nil, fmt.Errorf("no active schema found")
	}

	return s.currentSchema, nil
}

// CheckCompatibility checks if a new schema is compatible with existing ones
func (s *SchemaServiceImpl) CheckCompatibility(newSDL string, previousVersionID *int) (*models.SchemaCompatibilityCheck, error) {
	// Parse the new SDL
	parsedSchema, err := s.parseSDL(newSDL)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{fmt.Sprintf("Invalid SDL: %v", err)},
		}, nil
	}

	// If no previous version specified, assume compatible
	if previousVersionID == nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: true,
		}, nil
	}

	// Get the previous schema
	query := `SELECT sdl FROM unified_schemas WHERE id = $1`
	var previousSDL string
	err = s.db.QueryRow(query, *previousVersionID).Scan(&previousSDL)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{fmt.Sprintf("Previous version not found: %v", err)},
		}, nil
	}

	// Parse previous schema
	_, err = s.parseSDL(previousSDL)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{fmt.Sprintf("Invalid previous SDL: %v", err)},
		}, nil
	}

	// Check compatibility
	compatibility, err := s.checkCompatibility("", parsedSchema)
	if err != nil {
		return &models.SchemaCompatibilityCheck{
			IsCompatible: false,
			Issues:       []string{err.Error()},
		}, nil
	}

	// Convert to expected format
	result := &models.SchemaCompatibilityCheck{
		IsCompatible: len(compatibility.BreakingChanges) == 0,
		Issues:       compatibility.BreakingChanges,
		Warnings:     compatibility.NewFields,
	}

	return result, nil
}

// GetPreviousVersionID retrieves the ID of the previous version for a given version
func (s *SchemaServiceImpl) GetPreviousVersionID(version string) (*int, error) {
	query := `SELECT previous_version_id FROM unified_schemas WHERE version = $1`
	var previousVersionID *int
	err := s.db.QueryRow(query, version).Scan(&previousVersionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema version %s not found", version)
		}
		return nil, fmt.Errorf("failed to get previous version ID: %w", err)
	}
	return previousVersionID, nil
}

// GetSchemaForVersion returns schema for specific version
func (s *SchemaServiceImpl) GetSchemaForVersion(version string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	schema, exists := s.schemaVersions[version]
	if !exists {
		return nil, fmt.Errorf("schema version %s not loaded in memory", version)
	}

	return schema, nil
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

	s.schemaVersions = make(map[string]*ast.QueryDocument)

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

// RouteQuery routes GraphQL query to appropriate schema version
func (s *SchemaServiceImpl) RouteQuery(query string, version string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// If no version specified, use current active schema
	if version == "" {
		if s.currentSchema == nil {
			return nil, fmt.Errorf("no active schema available")
		}
		return s.currentSchema, nil
	}

	// Get specific version
	schema, exists := s.schemaVersions[version]
	if !exists {
		return nil, fmt.Errorf("schema version %s not found", version)
	}

	return schema, nil
}

// GetDefaultSchema returns the currently active default schema
func (s *SchemaServiceImpl) GetDefaultSchema() (interface{}, error) {
	return s.GetCurrentActiveSchema()
}

// ReloadSchemasInMemory reloads all schemas from database into memory
func (s *SchemaServiceImpl) ReloadSchemasInMemory() error {
	return s.LoadSchema()
}

// GetSchemaVersions returns all loaded schema versions
func (s *SchemaServiceImpl) GetSchemaVersions() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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

	// Find the version of the current schema
	for version, schema := range s.schemaVersions {
		if schema == s.currentSchema {
			return version, nil
		}
	}

	return "", fmt.Errorf("active schema version not found in memory")
}

// GetConfiguration returns the current schema configuration
func (s *SchemaServiceImpl) GetConfiguration() *configs.SchemaConfig {
	return s.config
}

// ReloadConfiguration reloads the schema configuration
func (s *SchemaServiceImpl) ReloadConfiguration() error {
	s.config = configs.LoadSchemaConfig()
	return s.config.Validate()
}

// Helper methods

// parseSDL parses SDL string into AST document
func (s *SchemaServiceImpl) parseSDL(sdl string) (*ast.QueryDocument, error) {
	// For now, return a simple QueryDocument
	// In a real implementation, you would parse the SDL properly
	doc := &ast.QueryDocument{
		Operations: ast.OperationList{},
		Fragments:  ast.FragmentDefinitionList{},
	}

	return doc, nil
}

// checkCompatibility checks if the new schema is compatible with existing versions
func (s *SchemaServiceImpl) checkCompatibility(version string, newSchema *ast.QueryDocument) (*models.VersionCompatibility, error) {
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

// parseVersion parses version string into major, minor, patch
func (s *SchemaServiceImpl) parseVersion(version string) (int, int, int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

// getCurrentActiveSchema gets the current active schema from database
func (s *SchemaServiceImpl) getCurrentActiveSchema() (string, *ast.QueryDocument, error) {
	query := `SELECT version, sdl FROM unified_schemas WHERE status = $1 ORDER BY created_at DESC LIMIT 1`

	var version, sdl string
	err := s.db.QueryRow(query, models.SchemaStatusActive).Scan(&version, &sdl)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("no active schema found")
		}
		return "", nil, fmt.Errorf("failed to get active schema: %w", err)
	}

	parsed, err := s.parseSDL(sdl)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse active schema: %w", err)
	}

	return version, parsed, nil
}

// checkMinorCompatibility ensures minor versions only add new fields
func (s *SchemaServiceImpl) checkMinorCompatibility(current, new *ast.QueryDocument, compatibility *models.VersionCompatibility) error {
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
			if currentField.Type != newField.Type {
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
=======
// checkMajorCompatibility allows breaking changes for major versions
func (s *SchemaServiceImpl) checkMajorCompatibility(current, new *ast.QueryDocument, compatibility *models.VersionCompatibility) {
	// Major versions allow all changes
	compatibility.BreakingChanges = append(compatibility.BreakingChanges, "Major version allows breaking changes")
>>>>>>> e62b19e (Clean up and unit tests)
}

// extractTypes extracts type definitions from AST document
func (s *SchemaServiceImpl) extractTypes(doc *ast.QueryDocument) map[string]*TypeInfo {
	types := make(map[string]*TypeInfo)

	// QueryDocument doesn't have Definitions field, so we'll return empty for now
	// In a real implementation, you would need to parse the schema differently
	return types
}

// getTypeString converts AST type to string representation
func (s *SchemaServiceImpl) getTypeString(t ast.Type) string {
	// Simplified implementation for now
	return "String"
}

// TypeInfo represents type information
type TypeInfo struct {
	Name   string
	Fields map[string]*FieldInfo
}

// FieldInfo represents field information
type FieldInfo struct {
	Name string
	Type string
}

// VersionCompatibility represents compatibility information
type VersionCompatibility struct {
	ChangeType      string   `json:"change_type"`
	BreakingChanges []string `json:"breaking_changes"`
	NewFields       []string `json:"new_fields"`
}

// activateSchemaVersion activates a specific schema version
func (s *SchemaServiceImpl) activateSchemaVersion(version string) error {
	// Deactivate all other schemas
	deactivateQuery := `UPDATE unified_schemas SET status = $1, updated_at = NOW()`
	_, err := s.db.Exec(deactivateQuery, models.SchemaStatusInactive)
	if err != nil {
		return fmt.Errorf("failed to deactivate other schemas: %w", err)
	}

	// Activate the specified version
	activateQuery := `UPDATE unified_schemas SET status = $1, updated_at = NOW() WHERE version = $2`
	result, err := s.db.Exec(activateQuery, models.SchemaStatusActive, version)
	if err != nil {
		return fmt.Errorf("failed to activate schema version: %w", err)
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

// getAllActiveSchemas retrieves all active schemas from database
func (s *SchemaServiceImpl) getAllActiveSchemas() ([]*models.UnifiedSchema, error) {
	query := `SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id 
		FROM unified_schemas 
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4`

	rows, err := s.db.Query(query, models.SchemaStatusActive, models.SchemaStatusInactive, s.config.GetVersionHistoryLimit(), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to query active schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*models.UnifiedSchema
	for rows.Next() {
		var schema models.UnifiedSchema
		err := rows.Scan(&schema.ID, &schema.Version, &schema.SDL, &schema.CreatedAt, &schema.CreatedBy,
			&schema.Status, &schema.ChangeType, &schema.Notes, &schema.PreviousVersionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, &schema)
	}

	return schemas, nil
}

// checkVersionLimits checks if we're within version limits
func (s *SchemaServiceImpl) checkVersionLimits() error {
	if !s.config.IsVersioningEnabled() {
		return nil
	}

	// Count current versions
	query := `SELECT COUNT(*) FROM unified_schemas`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count versions: %w", err)
	}

	if count >= s.config.GetMaxVersions() {
		return fmt.Errorf("maximum number of versions (%d) reached", s.config.GetMaxVersions())
	}

	return nil
}

// cleanupOldVersions removes old versions if we exceed limits
func (s *SchemaServiceImpl) cleanupOldVersions() error {
	if !s.config.IsVersioningEnabled() {
		return nil
	}

	// Get versions to delete (oldest first)
	query := `
		SELECT id FROM unified_schemas 
		WHERE status = $1
		ORDER BY created_at ASC 
		LIMIT $2 OFFSET $3`

	limit := s.config.GetMaxVersions() - s.config.GetVersionHistoryLimit()
	if limit <= 0 {
		return nil
	}

	rows, err := s.db.Query(query, models.SchemaStatusInactive, limit, 0)
	if err != nil {
		return fmt.Errorf("failed to query old versions: %w", err)
	}
	defer rows.Close()

	var idsToDelete []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan version id: %w", err)
		}
		idsToDelete = append(idsToDelete, id)
	}

	// Delete old versions
	if len(idsToDelete) > 0 {
		deleteQuery := `DELETE FROM unified_schemas WHERE id = ANY($1)`
		_, err = s.db.Exec(deleteQuery, idsToDelete)
		if err != nil {
			return fmt.Errorf("failed to delete old versions: %w", err)
		}
	}

	return nil
}

// LoadContractTests loads contract tests from database
func (ct *ContractTester) LoadContractTests() ([]ContractTest, error) {
	if ct.db == nil {
		return []ContractTest{}, nil
	}

	query := `SELECT id, name, query, variables, expected, description, priority, is_active 
		FROM contract_tests WHERE is_active = true ORDER BY priority DESC, created_at ASC`

	rows, err := ct.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to load contract tests: %w", err)
	}
	defer rows.Close()

	var tests []ContractTest
	for rows.Next() {
		var test ContractTest
		var variablesJSON, expectedJSON string

		err := rows.Scan(&test.ID, &test.Name, &test.Query, &variablesJSON, &expectedJSON,
			&test.Description, &test.Priority, &test.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contract test: %w", err)
		}

		// Parse JSON fields
		if variablesJSON != "" {
			// In a real implementation, you'd parse the JSON
			test.Variables = make(map[string]interface{})
		}
		if expectedJSON != "" {
			// In a real implementation, you'd parse the JSON
			test.Expected = make(map[string]interface{})
		}

		tests = append(tests, test)
	}

	return tests, nil
}

// ExecuteContractTests runs all contract tests against a schema
func (ct *ContractTester) ExecuteContractTests(schema *ast.QueryDocument) (*ContractTestResults, error) {
	tests, err := ct.LoadContractTests()
	if err != nil {
		return nil, err
	}

	results := &ContractTestResults{
		TotalTests: len(tests),
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
func (ct *ContractTester) runSingleTest(test ContractTest, schema *ast.QueryDocument) TestResult {
	start := time.Now()

	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Parse the GraphQL query
	// 2. Execute it against the schema
	// 3. Compare the result with expected output

	result := TestResult{
		TestName: test.Name,
		Passed:   true, // Placeholder
		Duration: time.Since(start),
	}

	return result
}
<<<<<<< HEAD

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
=======
>>>>>>> e62b19e (Clean up and unit tests)
