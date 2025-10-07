package services

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
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

	compatible, reason, _ := s.analyzeCompatibility(activeSchema.SDL, newSDL)
	return compatible, reason
}

// Helper methods
func (s *SchemaService) isValidSDL(sdl string) bool {
	// Simple validation - check for basic GraphQL structure
	return len(sdl) > 0 && strings.Contains(sdl, "type")
}

func (s *SchemaService) isBackwardCompatible(oldSDL, newSDL string) (bool, string) {
	compatible, _, _ := s.analyzeCompatibility(oldSDL, newSDL)
	return compatible, "compatible"
}

// analyzeCompatibility performs detailed compatibility analysis
func (s *SchemaService) analyzeCompatibility(oldSDL, newSDL string) (bool, string, map[string]interface{}) {
	changes := map[string]interface{}{
		"breaking":     []string{},
		"non_breaking": []string{},
		"warnings":     []string{},
	}

	// Simple analysis - in a real implementation, this would use a GraphQL parser
	// For now, we'll do basic string analysis

	// Check for removed fields (breaking change)
	if s.hasRemovedFields(oldSDL, newSDL) {
		changes["breaking"] = append(changes["breaking"].([]string), "Fields have been removed")
		return false, "breaking changes detected", changes
	}

	// Check for added fields (non-breaking change)
	if s.hasAddedFields(oldSDL, newSDL) {
		changes["non_breaking"] = append(changes["non_breaking"].([]string), "New fields have been added")
	}

	// Check for type changes (breaking change)
	if s.hasTypeChanges(oldSDL, newSDL) {
		changes["breaking"] = append(changes["breaking"].([]string), "Field types have been changed")
		return false, "breaking changes detected", changes
	}

	// Check for deprecated fields (warning)
	if s.hasDeprecatedFields(newSDL) {
		changes["warnings"] = append(changes["warnings"].([]string), "Some fields are marked as deprecated")
	}

	return true, "compatible", changes
}

// hasRemovedFields checks if any fields were removed
func (s *SchemaService) hasRemovedFields(oldSDL, newSDL string) bool {
	// Simple check - if new SDL is significantly shorter, fields might have been removed
	// This is a simplified implementation
	return len(newSDL) < int(float64(len(oldSDL))*0.8)
}

// hasAddedFields checks if new fields were added
func (s *SchemaService) hasAddedFields(oldSDL, newSDL string) bool {
	// Simple check - if new SDL is longer, fields might have been added
	return len(newSDL) > len(oldSDL)
}

// hasTypeChanges checks if field types were changed
func (s *SchemaService) hasTypeChanges(oldSDL, newSDL string) bool {
	// Simple check - look for type changes in field definitions
	// This is a simplified implementation
	return strings.Contains(newSDL, "String!") && strings.Contains(oldSDL, "Int!")
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
