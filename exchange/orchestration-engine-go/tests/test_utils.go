package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/require"
)

// getEnvOrDefault gets an environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getTestDatabaseConnectionString returns a database connection string from environment variables
// Requires TEST_DB_* environment variables - no hardcoded credentials
func getTestDatabaseConnectionString() (string, error) {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := os.Getenv("TEST_DB_USERNAME")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_DATABASE")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	// Require sensitive credentials from environment - no defaults
	if user == "" {
		return "", fmt.Errorf("TEST_DB_USERNAME environment variable not set")
	}
	if password == "" {
		return "", fmt.Errorf("TEST_DB_PASSWORD environment variable not set")
	}
	if dbname == "" {
		return "", fmt.Errorf("TEST_DB_DATABASE environment variable not set")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return dsn, nil
}

// hasDatabaseConnection checks if a database connection is available
// Uses environment variables - no hardcoded credentials
func hasDatabaseConnection() bool {
	dsn, err := getTestDatabaseConnectionString()
	if err != nil {
		return false
	}

	db, err := database.NewSchemaDB(dsn)
	if err != nil {
		return false
	}
	defer db.Close()
	return true
}

// setupTestDatabase creates a test database connection
// Requires TEST_DB_* environment variables - no hardcoded credentials
func setupTestDatabase() (*database.SchemaDB, error) {
	dsn, err := getTestDatabaseConnectionString()
	if err != nil {
		return nil, err
	}

	return database.NewSchemaDB(dsn)
}

// ParseTestQuery parses a GraphQL query string into an AST document
// Helper function for tests - similar to federator/test_helpers_test.go
func ParseTestQuery(t *testing.T, query string) *ast.Document {
	t.Helper()
	src := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "TestQuery",
	})
	doc, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err)
	return doc
}

// CreateTestSchema creates a test schema AST document
// Helper function for tests - similar to federator/test_helpers_test.go
func CreateTestSchema(t *testing.T) *ast.Document {
	t.Helper()
	schemaSDL := `
directive @sourceInfo(
	providerKey: String!
	schemaId: String
	providerField: String!
) on FIELD_DEFINITION

type Query {
	personInfo(nic: String!): PersonInfo
	personInfos(nics: [String!]!): [PersonInfo]
}

type PersonInfo {
	fullName: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.fullName")
	name: String @sourceInfo(providerKey: "rgd", schemaId: "rgd-schema-v1", providerField: "getPersonInfo.name")
	address: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.permanentAddress")
	ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data")
	class: [VehicleClass] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.classes")
}

type VehicleInfo {
	regNo: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.registrationNumber")
	make: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.make")
	model: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.model")
}

type VehicleClass {
	className: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.classes.className")
}
`
	src := source.NewSource(&source.Source{
		Body: []byte(schemaSDL),
		Name: "TestSchema",
	})
	doc, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err)
	return doc
}
