package utils

import (
	"log/slog"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
)

// Example usage of GraphQLHandler
func ExampleGraphQLHandler() {
	handler := NewGraphQLHandler()

	sdl := `
	directive @accessControl(type: String) on FIELD_DEFINITION
	directive @source(value: String) on FIELD_DEFINITION
	directive @isOwner(value: Boolean) on FIELD_DEFINITION
	directive @owner(value: String) on FIELD_DEFINITION
	directive @displayName(value: String) on FIELD_DEFINITION
	directive @description(value: String) on FIELD_DEFINITION

	type BirthInfo {
	  birthCertificateID: ID! @accessControl(type: "public") @source(value: "fallback")
	  birthPlace: String! @accessControl(type: "public") @source(value: "fallback")
	  birthDate: String! @accessControl(type: "public") @source(value: "fallback")
	}

	type User {
	  id: ID! @accessControl(type: "public") @source(value: "fallback")
	  name: String! @accessControl(type: "public") @source(value: "fallback")
	  email: String! @accessControl(type: "public") @source(value: "fallback")
	  birthInfo: BirthInfo @description(value: "Default Description")
	}

	type Query {
	  getUser(id: ID!): User @description(value: "Default Description")
	  listUsers: [User!]! @description(value: "Default Description")
	  getBirthInfo(userId: ID!): BirthInfo @description(value: "Default Description")
	  listUsersByBirthPlace(birthPlace: String!): [User!]! @description(value: "Default Description")
	  searchUsersByName(name: String!): [User!]! @description(value: "Default Description")
	}
	`

	request, err := handler.ParseSDLToPolicyRequest("schema123", sdl)
	slog.Info("Parsed Policy Metadata Create Request", "request", request)
	if err != nil {
		panic(err)
	}

	// request now contains PolicyMetadataCreateRequest with field paths like:
	// - user.id
	// - user.name
	// - user.email
	// - user.birthInfo (if it has directives)
	// - user.birthInfo.birthCertificateID
	// - user.birthInfo.birthPlace
	// - user.birthInfo.birthDate
	// - birthinfo.birthCertificateID
	// - birthinfo.birthPlace
	// - birthinfo.birthDate

	_ = request // Use the request as needed
}

func TestGraphQLHandler_ParseSDLToPolicyRequest(t *testing.T) {
	handler := NewGraphQLHandler()

	sdl := `
	directive @accessControl(type: String) on FIELD_DEFINITION
	directive @source(value: String) on FIELD_DEFINITION

	type User {
	  id: ID! @accessControl(type: "public") @source(value: "fallback")
	  name: String! @accessControl(type: "public") @source(value: "fallback")
	}
	`

	request, err := handler.ParseSDLToPolicyRequest("test-schema", sdl)
	slog.Info("Parsed Policy Metadata Create Request", "request", request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if request.SchemaID != "test-schema" {
		t.Errorf("Expected schema ID 'test-schema', got '%s'", request.SchemaID)
	}

	if len(request.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(request.Records))
	}

	// Check field names
	expectedFields := map[string]bool{
		"user.id":   true,
		"user.name": true,
	}

	for _, record := range request.Records {
		if !expectedFields[record.FieldName] {
			t.Errorf("Unexpected field name: %s", record.FieldName)
		}

		if record.Source != models.SourceFallback {
			t.Errorf("Expected source 'fallback', got '%s'", record.Source)
		}

		if record.AccessControlType != models.AccessControlTypePublic {
			t.Errorf("Expected access control 'public', got '%s'", record.AccessControlType)
		}
	}
}
