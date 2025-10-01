package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestParseSDL(t *testing.T) {
	// Create a mock schema service for testing
	schemaService := &services.SchemaServiceImpl{
		SchemaDocs: make(map[string]*ast.SchemaDocument),
	}

	t.Run("ValidSDL", func(t *testing.T) {
		validSDL := `
			type Query {
				hello: String
				world: String
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})

	t.Run("ValidSDLWithTypes", func(t *testing.T) {
		validSDL := `
			type Query {
				personInfo(nic: String!): PersonInfo
				vehicle: VehicleInfo
			}

			type PersonInfo {
				fullName: String
				name: String
				address: String
			}

			type VehicleInfo {
				regNo: String
				make: String
				model: String
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})

	t.Run("ValidSDLWithDirectives", func(t *testing.T) {
		validSDL := `
			directive @deprecated(
				reason: String = "No longer supported"
			) on FIELD_DEFINITION | ENUM_VALUE

			directive @sourceInfo(
				providerKey: String!
				providerField: String!
			) on FIELD_DEFINITION

			type Query {
				personInfo(nic: String!): PersonInfo
			}

			type PersonInfo {
				fullName: String @sourceInfo(providerKey: "drp", providerField: "person.fullName")
				name: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.name")
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})

	t.Run("InvalidSDL", func(t *testing.T) {
		invalidSDL := `
			type Query {
				hello: String
				// Missing closing brace
		`

		queryDoc, err := schemaService.ParseSDL(invalidSDL)
		assert.Error(t, err)
		assert.Nil(t, queryDoc)
		assert.Contains(t, err.Error(), "failed to parse SDL")
	})

	t.Run("EmptySDL", func(t *testing.T) {
		emptySDL := ""

		queryDoc, err := schemaService.ParseSDL(emptySDL)
		assert.Error(t, err)
		assert.Nil(t, queryDoc)
		assert.Contains(t, err.Error(), "SDL string cannot be empty")
	})

	t.Run("SDLWithoutQueryType", func(t *testing.T) {
		sdlWithoutQuery := `
			type Person {
				name: String
			}
		`

		queryDoc, err := schemaService.ParseSDL(sdlWithoutQuery)
		assert.Error(t, err)
		assert.Nil(t, queryDoc)
		assert.Contains(t, err.Error(), "schema must contain a Query type")
	})

	t.Run("SDLWithInputTypes", func(t *testing.T) {
		validSDL := `
			type Query {
				search(input: SearchInput!): [SearchResult!]!
			}

			input SearchInput {
				query: String!
				filters: [String!]
			}

			type SearchResult {
				id: ID!
				title: String!
				description: String
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})

	t.Run("SDLWithEnums", func(t *testing.T) {
		validSDL := `
			type Query {
				users(status: UserStatus): [User!]!
			}

			type User {
				id: ID!
				name: String!
				status: UserStatus!
			}

			enum UserStatus {
				ACTIVE
				INACTIVE
				PENDING
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})

	t.Run("SDLWithUnions", func(t *testing.T) {
		validSDL := `
			type Query {
				search(query: String!): [SearchResult!]!
			}

			union SearchResult = User | Post | Comment

			type User {
				id: ID!
				name: String!
			}

			type Post {
				id: ID!
				title: String!
				content: String!
			}

			type Comment {
				id: ID!
				text: String!
			}
		`

		queryDoc, err := schemaService.ParseSDL(validSDL)
		assert.NoError(t, err)
		assert.NotNil(t, queryDoc)
	})
}

func TestValidateSDL(t *testing.T) {
	schemaService := &services.SchemaServiceImpl{
		SchemaDocs: make(map[string]*ast.SchemaDocument),
	}

	t.Run("ValidSDL", func(t *testing.T) {
		validSDL := `
			type Query {
				hello: String
			}
		`

		err := schemaService.ValidateSDL(validSDL)
		assert.NoError(t, err)
	})

	t.Run("InvalidSDL", func(t *testing.T) {
		invalidSDL := `
			type Query {
				hello: String
				// Missing closing brace
		`

		err := schemaService.ValidateSDL(invalidSDL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse SDL")
	})

	t.Run("EmptySDL", func(t *testing.T) {
		err := schemaService.ValidateSDL("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SDL string cannot be empty")
	})
}
