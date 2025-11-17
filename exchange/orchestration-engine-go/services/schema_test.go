package services

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.Init()
}

func TestSchemaService_ValidateSDL(t *testing.T) {
	service := &SchemaService{}

	// Valid SDL
	valid := service.ValidateSDL("type Query { test: String }")
	assert.True(t, valid)

	// Invalid SDL (empty)
	valid = service.ValidateSDL("")
	assert.False(t, valid)

	// Invalid SDL (no type)
	valid = service.ValidateSDL("invalid")
	assert.False(t, valid)

	// Valid SDL with multiple types
	valid = service.ValidateSDL("type Query { test: String } type Mutation { create: String }")
	assert.True(t, valid)
}

func TestSchemaService_CheckCompatibility_NoActiveSchema(t *testing.T) {
	// This test requires database setup
	// When db is nil, GetActiveSchema will panic (nil pointer dereference)
	// Full integration tests with proper database setup are in tests/schema_database_test.go
	t.Skip("Requires database setup - tested in integration tests")
}

func TestSchemaService_isValidSDL(t *testing.T) {
	service := &SchemaService{}

	// Test the isValidSDL helper through ValidateSDL
	// Valid cases
	assert.True(t, service.ValidateSDL("type Query { test: String }"))
	assert.True(t, service.ValidateSDL("type Mutation { create: String }"))
	assert.True(t, service.ValidateSDL("type Query { test: String } type User { id: ID }"))

	// Invalid cases
	assert.False(t, service.ValidateSDL(""))
	assert.False(t, service.ValidateSDL("invalid"))
	assert.False(t, service.ValidateSDL("query { test }"))
}

func TestSchemaService_hasDeprecatedFields(t *testing.T) {
	// Test hasDeprecatedFields through CheckCompatibility
	// This is tested indirectly through compatibility checks
	// Direct testing would require database setup
	// Full integration tests are in tests/schema_database_test.go
	t.Log("hasDeprecatedFields is tested through CheckCompatibility in integration tests")
}
