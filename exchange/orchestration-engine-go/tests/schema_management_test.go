package tests

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
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

	// Test the problematic case: 2.0.0 vs 10.0.0
	if isVersionGreater("2.0.0", "10.0.0") {
		t.Error("2.0.0 should NOT be greater than 10.0.0 (semantic versioning)")
	}

	// Test edge cases
	if !isVersionGreater("10.0.0", "2.0.0") {
		t.Error("10.0.0 should be greater than 2.0.0")
	}

	if !isVersionGreater("1.0.1", "1.0.0") {
		t.Error("1.0.1 should be greater than 1.0.0")
	}

	if !isVersionGreater("1.1.0", "1.0.9") {
		t.Error("1.1.0 should be greater than 1.0.9")
	}
}

func isVersionGreater(version1, version2 string) bool {
	// Proper semantic version comparison
	// Parse version strings into major.minor.patch components
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	// Ensure both versions have at least 3 parts (major.minor.patch)
	if len(v1Parts) < 3 || len(v2Parts) < 3 {
		// Fallback to string comparison if format is invalid
		return version1 > version2
	}

	// Compare major version
	v1Major, err1 := strconv.Atoi(v1Parts[0])
	v2Major, err2 := strconv.Atoi(v2Parts[0])
	if err1 != nil || err2 != nil {
		return version1 > version2 // Fallback to string comparison
	}
	if v1Major != v2Major {
		return v1Major > v2Major
	}

	// Compare minor version
	v1Minor, err1 := strconv.Atoi(v1Parts[1])
	v2Minor, err2 := strconv.Atoi(v2Parts[1])
	if err1 != nil || err2 != nil {
		return version1 > version2 // Fallback to string comparison
	}
	if v1Minor != v2Minor {
		return v1Minor > v2Minor
	}

	// Compare patch version
	v1Patch, err1 := strconv.Atoi(v1Parts[2])
	v2Patch, err2 := strconv.Atoi(v2Parts[2])
	if err1 != nil || err2 != nil {
		return version1 > version2 // Fallback to string comparison
	}

	return v1Patch > v2Patch
}

// Test backward compatibility checking
func TestBackwardCompatible(t *testing.T) {
	// Import the compatibility checker
	checker := services.NewSchemaCompatibilityChecker()

	// Test 1: Adding a field should be backward compatible
	oldSDL := "type Query { hello: String }"
	newSDL := "type Query { hello: String world: String }"

	result := checker.CheckCompatibility(oldSDL, newSDL)
	if !result.IsCompatible {
		t.Errorf("Adding a field should be backward compatible, but got: %s", result.Reason)
	}

	// Verify it's detected as a non-breaking change
	nonBreaking := result.Changes["non_breaking"].([]services.NonBreakingChange)
	if len(nonBreaking) == 0 {
		t.Error("Should detect field addition as non-breaking change")
	}

	// Test 2: Removing a field should not be backward compatible
	removingFieldSDL := "type Query { }"
	result = checker.CheckCompatibility(oldSDL, removingFieldSDL)
	if result.IsCompatible {
		t.Error("Removing a field should not be backward compatible")
	}

	// Verify it's detected as a breaking change
	breaking := result.Changes["breaking"].([]services.BreakingChange)
	if len(breaking) == 0 {
		t.Error("Should detect field removal as breaking change")
	}

	// Test 3: Type change should be breaking
	typeChangeSDL := "type Query { hello: Int }"
	result = checker.CheckCompatibility(oldSDL, typeChangeSDL)
	if result.IsCompatible {
		t.Error("Changing field type should not be backward compatible")
	}

	// Test 4: Adding a new type should be compatible
	newTypeSDL := "type Query { hello: String } type Mutation { update: String }"
	result = checker.CheckCompatibility(oldSDL, newTypeSDL)
	if !result.IsCompatible {
		t.Errorf("Adding a new type should be backward compatible, but got: %s", result.Reason)
	}

	// Test 5: Removing a type should be breaking
	removingTypeSDL := "type Mutation { update: String }"
	result = checker.CheckCompatibility(newTypeSDL, removingTypeSDL)
	if result.IsCompatible {
		t.Error("Removing a type should not be backward compatible")
	}

	// Test 6: Adding optional argument should be compatible
	oldWithArgs := "type Query { hello(name: String): String }"
	newWithOptionalArg := "type Query { hello(name: String, age: Int): String }"
	result = checker.CheckCompatibility(oldWithArgs, newWithOptionalArg)
	if !result.IsCompatible {
		t.Errorf("Adding optional argument should be backward compatible, but got: %s", result.Reason)
	}

	// Test 7: Adding required argument should be breaking
	newWithRequiredArg := "type Query { hello(name: String!, age: Int!): String }"
	result = checker.CheckCompatibility(oldWithArgs, newWithRequiredArg)
	if result.IsCompatible {
		t.Error("Adding required argument should not be backward compatible")
	}

	// Test 8: Removing argument should be breaking
	removingArgSDL := "type Query { hello: String }"
	result = checker.CheckCompatibility(oldWithArgs, removingArgSDL)
	if result.IsCompatible {
		t.Error("Removing argument should not be backward compatible")
	}
}

// Test comprehensive compatibility scenarios
func TestComprehensiveCompatibility(t *testing.T) {
	checker := services.NewSchemaCompatibilityChecker()

	testCases := []struct {
		name                string
		oldSDL              string
		newSDL              string
		shouldBeCompatible  bool
		expectedBreaking    int
		expectedNonBreaking int
	}{
		{
			name:                "Add field",
			oldSDL:              "type Query { name: String }",
			newSDL:              "type Query { name: String age: Int }",
			shouldBeCompatible:  true,
			expectedBreaking:    0,
			expectedNonBreaking: 1,
		},
		{
			name:                "Remove field",
			oldSDL:              "type Query { name: String age: Int }",
			newSDL:              "type Query { name: String }",
			shouldBeCompatible:  false,
			expectedBreaking:    1,
			expectedNonBreaking: 0,
		},
		{
			name:                "Change field type",
			oldSDL:              "type Query { age: String }",
			newSDL:              "type Query { age: Int }",
			shouldBeCompatible:  false,
			expectedBreaking:    1,
			expectedNonBreaking: 0,
		},
		{
			name:                "Add new type",
			oldSDL:              "type Query { name: String }",
			newSDL:              "type Query { name: String } type User { id: ID }",
			shouldBeCompatible:  true,
			expectedBreaking:    0,
			expectedNonBreaking: 1,
		},
		{
			name:                "Remove type",
			oldSDL:              "type Query { name: String } type User { id: ID }",
			newSDL:              "type Query { name: String }",
			shouldBeCompatible:  false,
			expectedBreaking:    1,
			expectedNonBreaking: 0,
		},
		{
			name:                "Change to non-null",
			oldSDL:              "type Query { name: String }",
			newSDL:              "type Query { name: String! }",
			shouldBeCompatible:  false,
			expectedBreaking:    1,
			expectedNonBreaking: 0,
		},
		{
			name:                "Change to nullable",
			oldSDL:              "type Query { name: String! }",
			newSDL:              "type Query { name: String }",
			shouldBeCompatible:  false,
			expectedBreaking:    1,
			expectedNonBreaking: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checker.CheckCompatibility(tc.oldSDL, tc.newSDL)

			if result.IsCompatible != tc.shouldBeCompatible {
				t.Errorf("Expected compatible=%v, got %v. Reason: %s",
					tc.shouldBeCompatible, result.IsCompatible, result.Reason)
			}

			breaking := result.Changes["breaking"].([]services.BreakingChange)
			nonBreaking := result.Changes["non_breaking"].([]services.NonBreakingChange)

			if len(breaking) != tc.expectedBreaking {
				t.Errorf("Expected %d breaking changes, got %d", tc.expectedBreaking, len(breaking))
			}

			if len(nonBreaking) != tc.expectedNonBreaking {
				t.Errorf("Expected %d non-breaking changes, got %d", tc.expectedNonBreaking, len(nonBreaking))
			}
		})
	}
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
