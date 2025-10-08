package services

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Schema represents a GraphQL schema with basic versioning
type Schema struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	SDL       string    `json:"sdl"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	Checksum  string    `json:"checksum"`
}

// SchemaService handles schema management operations
type SchemaService struct {
	db *database.SchemaDB
}

// NewSchemaService creates a new schema service
func NewSchemaService(db *database.SchemaDB) *SchemaService {
	return &SchemaService{
		db: db,
	}
}

// CreateSchema creates a new schema version
func (s *SchemaService) CreateSchema(version, sdl, createdBy string) (*Schema, error) {
	// Validate SDL
	if !s.isValidSDL(sdl) {
		return nil, fmt.Errorf("invalid SDL syntax")
	}

	// Generate checksum
	checksum := s.generateChecksum(sdl)

	// Create schema
	schema := &database.Schema{
		ID:        s.generateID(),
		Version:   version,
		SDL:       sdl,
		Status:    "inactive",
		IsActive:  false,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
		Checksum:  checksum,
	}

	// Save to database
	if err := s.db.CreateSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to save schema to database: %w", err)
	}

	// Convert to service schema
	serviceSchema := &Schema{
		ID:        schema.ID,
		Version:   schema.Version,
		SDL:       schema.SDL,
		IsActive:  schema.IsActive,
		CreatedAt: schema.CreatedAt,
		CreatedBy: schema.CreatedBy,
		Checksum:  schema.Checksum,
	}

	return serviceSchema, nil
}

// GetActiveSchema returns the currently active schema
func (s *SchemaService) GetActiveSchema() (*Schema, error) {
	dbSchema, err := s.db.GetActiveSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to get active schema: %w", err)
	}

	if dbSchema == nil {
		return nil, nil // No active schema
	}

	// Convert to service schema
	serviceSchema := &Schema{
		ID:        dbSchema.ID,
		Version:   dbSchema.Version,
		SDL:       dbSchema.SDL,
		IsActive:  dbSchema.IsActive,
		CreatedAt: dbSchema.CreatedAt,
		CreatedBy: dbSchema.CreatedBy,
		Checksum:  dbSchema.Checksum,
	}

	return serviceSchema, nil
}

// ActivateSchema activates a specific schema version
func (s *SchemaService) ActivateSchema(version string) error {
	return s.db.ActivateSchema(version)
}

// GetAllSchemas returns all schemas
func (s *SchemaService) GetAllSchemas() ([]Schema, error) {
	dbSchemas, err := s.db.GetAllSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}

	// Convert to service schemas
	schemas := make([]Schema, len(dbSchemas))
	for i, dbSchema := range dbSchemas {
		schemas[i] = Schema{
			ID:        dbSchema.ID,
			Version:   dbSchema.Version,
			SDL:       dbSchema.SDL,
			IsActive:  dbSchema.IsActive,
			CreatedAt: dbSchema.CreatedAt,
			CreatedBy: dbSchema.CreatedBy,
			Checksum:  dbSchema.Checksum,
		}
	}

	return schemas, nil
}

// ValidateSDL validates GraphQL SDL syntax
func (s *SchemaService) ValidateSDL(sdl string) bool {
	return s.isValidSDL(sdl)
}

// CheckCompatibility checks if a new SDL is backward compatible with the active schema
func (s *SchemaService) CheckCompatibility(newSDL string) (bool, string) {
	activeSchema, err := s.GetActiveSchema()
	if err != nil {
		return false, "failed to get active schema: " + err.Error()
	}

	if activeSchema == nil {
		return true, "no active schema to compare against"
	}

	// Use the new comprehensive compatibility checker
	checker := NewSchemaCompatibilityChecker()
	result := checker.CheckCompatibility(activeSchema.SDL, newSDL)
	return result.IsCompatible, result.Reason
}

// Helper methods
func (s *SchemaService) isValidSDL(sdl string) bool {
	// Simple validation - check for basic GraphQL structure
	return len(sdl) > 0 && strings.Contains(sdl, "type")
}

func (s *SchemaService) isBackwardCompatible(oldSDL, newSDL string) (bool, string) {
	// Use the new comprehensive compatibility checker
	checker := NewSchemaCompatibilityChecker()
	result := checker.CheckCompatibility(oldSDL, newSDL)
	return result.IsCompatible, result.Reason
}

// analyzeCompatibility performs detailed compatibility analysis using the new checker
func (s *SchemaService) analyzeCompatibility(oldSDL, newSDL string) (bool, string, map[string]interface{}) {
	checker := NewSchemaCompatibilityChecker()
	result := checker.CheckCompatibility(oldSDL, newSDL)
	return result.IsCompatible, result.Reason, result.Changes
}

// hasRemovedFields checks if any fields were removed using AST comparison
func (s *SchemaService) hasRemovedFields(oldSDL, newSDL string) bool {
	// Parse both SDL strings into AST documents
	oldDoc, err := s.parseSDL(oldSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse old SDL for field removal detection", "Error", err)
		return false
	}
	newDoc, err := s.parseSDL(newSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse new SDL for field removal detection", "Error", err)
		return false
	}

	// Extract field definitions from both schemas
	oldFields := s.extractFieldDefinitions(oldDoc)
	newFields := s.extractFieldDefinitions(newDoc)

	// Check if any field from the old schema is missing in the new schema
	for fieldKey, oldField := range oldFields {
		if newField, exists := newFields[fieldKey]; !exists {
			// Field was removed
			logger.Log.Debug("Field removed", "type", oldField.TypeName, "field", oldField.FieldName)
			return true
		} else {
			// Field exists, check if type changed
			if oldField.TypeDefinition != newField.TypeDefinition {
				logger.Log.Debug("Field type changed", "type", oldField.TypeName, "field", oldField.FieldName,
					"oldType", oldField.TypeDefinition, "newType", newField.TypeDefinition)
				return true
			}
		}
	}
	return false
}

// hasAddedFields checks if new fields were added using AST comparison
func (s *SchemaService) hasAddedFields(oldSDL, newSDL string) bool {
	// Parse both SDL strings into AST documents
	oldDoc, err := s.parseSDL(oldSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse old SDL for field addition detection", "Error", err)
		return false
	}
	newDoc, err := s.parseSDL(newSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse new SDL for field addition detection", "Error", err)
		return false
	}

	// Extract field definitions from both schemas
	oldFields := s.extractFieldDefinitions(oldDoc)
	newFields := s.extractFieldDefinitions(newDoc)

	// Check if any field in the new schema is not in the old schema
	for fieldKey, newField := range newFields {
		if _, exists := oldFields[fieldKey]; !exists {
			// New field was added
			logger.Log.Debug("Field added", "type", newField.TypeName, "field", newField.FieldName)
			return true
		}
	}

	return false
}

// hasTypeChanges checks if field types were changed by parsing SDL and comparing type definitions
func (s *SchemaService) hasTypeChanges(oldSDL, newSDL string) bool {
	// Parse both SDL strings into AST documents
	oldDoc, err := s.parseSDL(oldSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse old SDL for type change detection", "Error", err)
		return false
	}

	newDoc, err := s.parseSDL(newSDL)
	if err != nil {
		logger.Log.Warn("Failed to parse new SDL for type change detection", "Error", err)
		return false
	}

	// Extract field definitions from both schemas
	oldFields := s.extractFieldDefinitions(oldDoc)
	newFields := s.extractFieldDefinitions(newDoc)

	// Compare field types systematically
	return s.compareFieldTypes(oldFields, newFields)
}

// parseSDL parses a GraphQL SDL string into an AST document
func (s *SchemaService) parseSDL(sdl string) (*ast.Document, error) {
	src := source.NewSource(&source.Source{
		Body: []byte(sdl),
		Name: "SchemaSDL",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	return doc, nil
}

// FieldTypeInfo represents a field with its type information
type FieldTypeInfo struct {
	TypeName       string
	FieldName      string
	TypeDefinition string
}

// extractFieldDefinitions extracts all field definitions from a schema document
func (s *SchemaService) extractFieldDefinitions(doc *ast.Document) map[string]FieldTypeInfo {
	fields := make(map[string]FieldTypeInfo)

	for _, def := range doc.Definitions {
		if objectType, ok := def.(*ast.ObjectDefinition); ok {
			for _, field := range objectType.Fields {
				if field != nil && field.Name != nil {
					fieldKey := fmt.Sprintf("%s.%s", objectType.Name.Value, field.Name.Value)
					typeDef := s.getTypeDefinition(field.Type)
					fields[fieldKey] = FieldTypeInfo{
						TypeName:       objectType.Name.Value,
						FieldName:      field.Name.Value,
						TypeDefinition: typeDef,
					}
				}
			}
		}
	}

	return fields
}

// getTypeDefinition converts a GraphQL type to its string representation
func (s *SchemaService) getTypeDefinition(t ast.Type) string {
	switch typeNode := t.(type) {
	case *ast.NonNull:
		return s.getTypeDefinition(typeNode.Type) + "!"
	case *ast.List:
		return "[" + s.getTypeDefinition(typeNode.Type) + "]"
	case *ast.Named:
		if typeNode.Name != nil {
			return typeNode.Name.Value
		}
	}
	return "Unknown"
}

// compareFieldTypes compares field types between old and new schemas
func (s *SchemaService) compareFieldTypes(oldFields, newFields map[string]FieldTypeInfo) bool {
	// Check for type changes in existing fields
	for fieldKey, oldField := range oldFields {
		if newField, exists := newFields[fieldKey]; exists {
			// Field exists in both schemas, check if type changed
			if oldField.TypeDefinition != newField.TypeDefinition {
				logger.Log.Info("Type change detected",
					"field", fieldKey,
					"oldType", oldField.TypeDefinition,
					"newType", newField.TypeDefinition)
				return true
			}
		}
	}

	// Check for removed fields (breaking change)
	for fieldKey := range oldFields {
		if _, exists := newFields[fieldKey]; !exists {
			logger.Log.Info("Field removed (breaking change)", "field", fieldKey)
			return true
		}
	}

	// Note: Adding new fields is not a breaking change, so we don't check for that

	return false
}

// hasDeprecatedFields checks if any fields are marked as deprecated
func (s *SchemaService) hasDeprecatedFields(sdl string) bool {
	return strings.Contains(sdl, "@deprecated")
}

func (s *SchemaService) generateID() string {
	return fmt.Sprintf("schema-%d", time.Now().UnixNano())
}

func (s *SchemaService) generateChecksum(sdl string) string {
	hash := sha256.Sum256([]byte(sdl))
	return fmt.Sprintf("%x", hash)
}
