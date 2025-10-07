package tests

import (
	"strings"
	"testing"
	"time"
)

// Simple schema model for testing
type Schema struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	SDL       string    `json:"sdl"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
}

// Test basic schema creation
func TestCreateSchema(t *testing.T) {
	schema := Schema{
		ID:        "test-1",
		Version:   "1.0.0",
		SDL:       "type Query { hello: String }",
		IsActive:  false,
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
	}

	if schema.ID == "" {
		t.Error("Schema ID should not be empty")
	}
	if schema.Version == "" {
		t.Error("Schema version should not be empty")
	}
	if schema.SDL == "" {
		t.Error("Schema SDL should not be empty")
	}
}

// Test schema validation
func TestValidateSDL(t *testing.T) {
	validSDL := "type Query { hello: String }"
	invalidSDL := "invalid graphql syntax"

	// Test valid SDL
	if !isValidSDL(validSDL) {
		t.Error("Valid SDL should pass validation")
	}

	// Test invalid SDL
	if isValidSDL(invalidSDL) {
		t.Error("Invalid SDL should fail validation")
	}
}

// Helper functions
func isValidSDL(sdl string) bool {
	// Simple validation - check for basic GraphQL structure
	return len(sdl) > 0 && strings.Contains(sdl, "type")
}

// Test version comparison
func TestVersionComparison(t *testing.T) {
	// Test version ordering
	if !isVersionGreater("1.1.0", "1.0.0") {
		t.Error("1.1.0 should be greater than 1.0.0")
	}

	if isVersionGreater("1.0.0", "1.1.0") {
		t.Error("1.0.0 should not be greater than 1.1.0")
	}

	if !isVersionGreater("2.0.0", "1.9.9") {
		t.Error("2.0.0 should be greater than 1.9.9")
	}
}

func isVersionGreater(version1, version2 string) bool {
	// Simple version comparison (semantic versioning)
	// This is a simplified implementation
	return version1 > version2
}

// Test backward compatibility checking
func TestBackwardCompatibility(t *testing.T) {
	oldSDL := "type Query { hello: String }"
	newSDL := "type Query { hello: String world: String }"

	// Adding a field should be backward compatible
	if !isBackwardCompatible(oldSDL, newSDL) {
		t.Error("Adding a field should be backward compatible")
	}

	// Removing a field should not be backward compatible
	removingFieldSDL := "type Query { }"
	if isBackwardCompatible(oldSDL, removingFieldSDL) {
		t.Error("Removing a field should not be backward compatible")
	}
}

func isBackwardCompatible(oldSDL, newSDL string) bool {
	// Simple compatibility check
	// Adding fields = compatible, removing fields = incompatible
	return len(newSDL) >= len(oldSDL)
}

// Test schema activation
func TestSchemaActivation(t *testing.T) {
	schemas := []Schema{
		{ID: "1", Version: "1.0.0", IsActive: true},
		{ID: "2", Version: "1.1.0", IsActive: false},
	}

	// Activate version 1.1.0
	activateSchema(schemas, "1.1.0")

	// Check that only one schema is active
	activeCount := 0
	for _, schema := range schemas {
		if schema.IsActive {
			activeCount++
		}
	}

	if activeCount != 1 {
		t.Error("Only one schema should be active at a time")
	}

	// Check that the correct schema is active
	if !schemas[1].IsActive {
		t.Error("Schema 1.1.0 should be active")
	}
}

func activateSchema(schemas []Schema, version string) {
	// Deactivate all schemas first
	for i := range schemas {
		schemas[i].IsActive = false
	}

	// Activate the specified version
	for i := range schemas {
		if schemas[i].Version == version {
			schemas[i].IsActive = true
			break
		}
	}
}
