package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// TestDatabaseSchemaOperations tests database operations
func TestDatabaseSchemaOperations(t *testing.T) {
	// Skip if no database connection
	if !hasDatabaseConnection() {
		t.Skip("Skipping database tests - no database connection")
	}

	// Create test database connection
	connectionString := "host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable"

	// Create schema mapping database
	schemaMappingDB, err := database.NewSchemaMappingDB(connectionString)
	if err != nil {
		t.Fatalf("Failed to connect to schema mapping database: %v", err)
	}
	defer schemaMappingDB.Close()

	// Create schema service
	schemaService := services.NewSchemaService(schemaMappingDB)

	// Test 1: Create schema
	t.Run("CreateSchema", func(t *testing.T) {
		schema, err := schemaService.CreateSchema("1.0.0", "type Query { hello: String }", "test-user")
		if err != nil {
			t.Errorf("Failed to create schema: %v", err)
		}
		if schema == nil {
			t.Error("Schema should not be nil")
		}
		if schema.Version != "1.0.0" {
			t.Errorf("Expected version 1.0.0, got %s", schema.Version)
		}
	})

	// Test 2: Get all schemas
	t.Run("GetAllSchemas", func(t *testing.T) {
		schemas, err := schemaService.GetAllSchemas()
		if err != nil {
			t.Errorf("Failed to get schemas: %v", err)
		}
		if len(schemas) == 0 {
			t.Error("Expected at least one schema")
		}
	})

	// Test 3: Activate schema
	t.Run("ActivateSchema", func(t *testing.T) {
		err := schemaService.ActivateSchema("1.0.0")
		if err != nil {
			t.Errorf("Failed to activate schema: %v", err)
		}
	})

	// Test 4: Get active schema
	t.Run("GetActiveSchema", func(t *testing.T) {
		schema, err := schemaService.GetActiveSchema()
		if err != nil {
			t.Errorf("Failed to get active schema: %v", err)
		}
		if schema == nil {
			t.Error("Active schema should not be nil")
		}
		if schema.Version != "1.0.0" {
			t.Errorf("Expected active schema version 1.0.0, got %s", schema.Version)
		}
	})

	// Test 5: Create another schema and activate it
	t.Run("CreateAndActivateNewSchema", func(t *testing.T) {
		// Create new schema
		_, err := schemaService.CreateSchema("1.1.0", "type Query { hello: String world: String }", "test-user")
		if err != nil {
			t.Errorf("Failed to create new schema: %v", err)
		}

		// Activate new schema
		err = schemaService.ActivateSchema("1.1.0")
		if err != nil {
			t.Errorf("Failed to activate new schema: %v", err)
		}

		// Verify only new schema is active
		schema, err := schemaService.GetActiveSchema()
		if err != nil {
			t.Errorf("Failed to get active schema: %v", err)
		}
		if schema.Version != "1.1.0" {
			t.Errorf("Expected active schema version 1.1.0, got %s", schema.Version)
		}
	})

	// Test 6: Compatibility checking
	t.Run("CompatibilityChecking", func(t *testing.T) {
		// Test compatible change
		compatible, reason := schemaService.CheckCompatibility("type Query { hello: String world: String newField: String }")
		if !compatible {
			t.Errorf("Expected compatible change, got incompatible: %s", reason)
		}

		// Test incompatible change (removing fields)
		compatible, reason = schemaService.CheckCompatibility("type Query { }")
		if compatible {
			t.Errorf("Expected incompatible change, got compatible: %s", reason)
		}
	})

}

// TestSchemaValidation tests SDL validation
func TestSchemaValidation(t *testing.T) {
	// Skip if no database connection
	if !hasDatabaseConnection() {
		t.Skip("Skipping database tests - no database connection")
	}

	connectionString := "host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable"

	schemaMappingDB, err := database.NewSchemaMappingDB(connectionString)
	if err != nil {
		t.Fatalf("Failed to connect to schema mapping database: %v", err)
	}
	defer schemaMappingDB.Close()

	schemaService := services.NewSchemaService(schemaMappingDB)

	// Test valid SDL
	valid := schemaService.ValidateSDL("type Query { hello: String }")
	if !valid {
		t.Error("Valid SDL should pass validation")
	}

	// Test invalid SDL
	valid = schemaService.ValidateSDL("invalid graphql syntax")
	if valid {
		t.Error("Invalid SDL should fail validation")
	}
}

// TestDatabaseErrorHandling tests error handling
func TestDatabaseErrorHandling(t *testing.T) {
	// Test with invalid connection string
	_, err := database.NewSchemaDB("invalid connection string")
	if err == nil {
		t.Error("Expected error with invalid connection string")
	}

	// Test with valid connection but invalid operations
	db, err := database.NewSchemaDB("host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable")
	if err != nil {
		t.Skip("Skipping test - no database connection")
	}
	defer db.Close()

	// Test getting non-existent schema
	_, err = db.GetSchemaByVersion("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent schema")
	}

	// Test activating non-existent schema
	err = db.ActivateSchema("non-existent")
	if err == nil {
		t.Error("Expected error when activating non-existent schema")
	}
}

// Helper function to check if database connection is available
func hasDatabaseConnection() bool {
	// Try to connect to database
	db, err := database.NewSchemaDB("host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable")
	if err != nil {
		return false
	}
	defer db.Close()
	return true
}

// TestSchemaVersioning tests schema versioning functionality
func TestSchemaVersioning(t *testing.T) {
	// Skip if no database connection
	if !hasDatabaseConnection() {
		t.Skip("Skipping database tests - no database connection")
	}

	connectionString := "host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable"

	schemaMappingDB, err := database.NewSchemaMappingDB(connectionString)
	if err != nil {
		t.Fatalf("Failed to connect to schema mapping database: %v", err)
	}
	defer schemaMappingDB.Close()

	schemaService := services.NewSchemaService(schemaMappingDB)

	// Create multiple schema versions
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}
	for i, version := range versions {
		sdl := "type Query { hello: String }"
		if i > 0 {
			sdl += " world: String"
		}
		if i == 2 {
			sdl += " newField: String"
		}

		_, err := schemaService.CreateSchema(version, sdl, "test-user")
		if err != nil {
			t.Errorf("Failed to create schema version %s: %v", version, err)
		}
	}

	// Test getting all schemas
	schemas, err := schemaService.GetAllSchemas()
	if err != nil {
		t.Errorf("Failed to get all schemas: %v", err)
	}
	if len(schemas) < len(versions) {
		t.Errorf("Expected at least %d schemas, got %d", len(versions), len(schemas))
	}

	// Test activating different versions
	for _, version := range versions {
		err := schemaService.ActivateSchema(version)
		if err != nil {
			t.Errorf("Failed to activate schema version %s: %v", version, err)
		}

		activeSchema, err := schemaService.GetActiveSchema()
		if err != nil {
			t.Errorf("Failed to get active schema: %v", err)
		}
		if activeSchema.Version != version {
			t.Errorf("Expected active schema version %s, got %s", version, activeSchema.Version)
		}
	}
}
