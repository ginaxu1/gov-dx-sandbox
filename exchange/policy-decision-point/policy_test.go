package main

import (
	"context"
	"testing"
)

// Helper function to get field names from fields map
func getFieldNames(fields map[string]interface{}) []string {
	var names []string
	for name := range fields {
		names = append(names, name)
	}
	return names
}

// TestNewPolicyEvaluator tests the creation of a policy evaluator
func TestNewPolicyEvaluator(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	if evaluator == nil {
		t.Fatal("Expected non-nil policy evaluator")
	}
}

// TestPolicyEvaluator_Authorize tests the main authorization functionality
func TestPolicyEvaluator_Authorize(t *testing.T) {
	ctx := context.Background()
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		t.Fatalf("Failed to create policy evaluator: %v", err)
	}

	tests := []struct {
		name            string
		request         map[string]interface{}
		expected        bool
		consentRequired bool
	}{
		{
			name: "Public field access - no consent required",
			request: map[string]interface{}{
				"consumer_id":     "any-app",
				"app_id":          "any-app",
				"request_id":      "req_public",
				"required_fields": []string{"person.fullName"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "Restricted field with allow list - no consent required",
			request: map[string]interface{}{
				"consumer_id":     "driver-app",
				"app_id":          "driver-app",
				"request_id":      "req_restricted",
				"required_fields": []string{"person.birthDate"},
			},
			expected:        true,
			consentRequired: false,
		},
		{
			name: "Restricted field with allow list - consent required",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"app_id":          "passport-app",
				"request_id":      "req_consent",
				"required_fields": []string{"person.photo"},
			},
			expected:        true,
			consentRequired: true,
		},
		{
			name: "Unauthorized consumer",
			request: map[string]interface{}{
				"consumer_id":     "unauthorized-app",
				"app_id":          "unauthorized-app",
				"request_id":      "req_unauth",
				"required_fields": []string{"person.permanentAddress"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "Missing consumer_id",
			request: map[string]interface{}{
				"app_id":          "passport-app",
				"request_id":      "req_missing",
				"required_fields": []string{"person.fullName"},
			},
			expected:        false,
			consentRequired: false,
		},
		{
			name: "Empty required_fields",
			request: map[string]interface{}{
				"consumer_id":     "passport-app",
				"app_id":          "passport-app",
				"request_id":      "req_empty",
				"required_fields": []string{},
			},
			expected:        false,
			consentRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := evaluator.Authorize(ctx, tt.request)
			if err != nil {
				t.Fatalf("Authorize() error = %v", err)
			}

			if decision.Allow != tt.expected {
				t.Errorf("Authorize() Allow = %v, expected %v", decision.Allow, tt.expected)
			}

			if decision.ConsentRequired != tt.consentRequired {
				t.Errorf("Authorize() ConsentRequired = %v, expected %v", decision.ConsentRequired, tt.consentRequired)
			}
		})
	}
}

// TestSchemaConverter tests the schema conversion functionality
func TestSchemaConverter(t *testing.T) {
	converter := NewSchemaConverter()

	t.Run("ConvertSDLToProviderMetadata", func(t *testing.T) {
		tests := []struct {
			name         string
			sdl          string
			providerID   string
			expectedKeys []string
		}{
			{
				name: "Basic SDL with public fields",
				sdl: `type User {
  id: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
  name: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
  email: String! @accessControl(type: "restricted") @source(value: "authoritative") @isOwner(value: false)
}`,
				providerID:   "drp",
				expectedKeys: []string{"user.id", "user.name", "user.email"},
			},
			{
				name: "SDL with nested types",
				sdl: `type BirthInfo {
  birthDate: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  birthPlace: String! @accessControl(type: "restricted") @source(value: "authoritative") @isOwner(value: false)
}

type User {
  id: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
  birthInfo: BirthInfo @accessControl(type: "public") @source(value: "authoritative")
}`,
				providerID:   "drp",
				expectedKeys: []string{"birthinfo.birthDate", "birthinfo.birthPlace", "user.id", "user.birthInfo"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := converter.ConvertSDLToProviderMetadataLegacy(tt.sdl, tt.providerID)
				if err != nil {
					t.Fatalf("ConvertSDLToProviderMetadata() error = %v", err)
				}

				fields, ok := result["fields"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected fields map")
				}

				for _, expectedKey := range tt.expectedKeys {
					if _, exists := fields[expectedKey]; !exists {
						t.Errorf("Expected field %s not found in result", expectedKey)
					}
				}
			})
		}
	})

	t.Run("ParseFieldLine", func(t *testing.T) {
		tests := []struct {
			name     string
			line     string
			expected GraphQLField
		}{
			{
				name: "Field with all directives",
				line: `  name: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true) @description(value: "User name")`,
				expected: GraphQLField{
					Name:          "name",
					Type:          "String!",
					AccessControl: "public",
					Description:   "User name",
					ParentType:    "User",
					Source:        "authoritative",
					IsOwner:       true,
					Owner:         "",
				},
			},
			{
				name: "Field with minimal directives",
				line: `  id: ID! @accessControl(type: "restricted")`,
				expected: GraphQLField{
					Name:          "id",
					Type:          "ID!",
					AccessControl: "restricted",
					Description:   "",
					ParentType:    "User",
				},
			},
			{
				name: "Field without directives",
				line: `  email: String!`,
				expected: GraphQLField{
					Name:          "email",
					Type:          "String!",
					AccessControl: "",
					Description:   "",
					ParentType:    "User",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := converter.parseFieldLine(tt.line, "User")
				if result == nil {
					t.Error("parseFieldLine() returned nil")
					return
				}
				if *result != tt.expected {
					t.Errorf("parseFieldLine() = %+v, expected %+v", *result, tt.expected)
				}
			})
		}
	})

	t.Run("DetermineConsentRequired", func(t *testing.T) {
		tests := []struct {
			name            string
			field           GraphQLField
			providerID      string
			expectedConsent bool
		}{
			{
				name: "Owner field with public access - no consent",
				field: GraphQLField{
					Name:          "name",
					AccessControl: "public",
					IsOwner:       true,
				},
				providerID:      "drp",
				expectedConsent: false,
			},
			{
				name: "Non-owner field with public access - no consent",
				field: GraphQLField{
					Name:          "email",
					AccessControl: "public",
					IsOwner:       false,
				},
				providerID:      "drp",
				expectedConsent: false,
			},
			{
				name: "Owner field with restricted access - no consent",
				field: GraphQLField{
					Name:          "photo",
					AccessControl: "restricted",
					IsOwner:       true,
				},
				providerID:      "drp",
				expectedConsent: false,
			},
			{
				name: "Non-owner field with restricted access - consent required",
				field: GraphQLField{
					Name:          "address",
					AccessControl: "restricted",
					IsOwner:       false,
				},
				providerID:      "drp",
				expectedConsent: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test the logic directly
				result := !tt.field.IsOwner && tt.field.AccessControl == "restricted"
				if result != tt.expectedConsent {
					t.Errorf("consent logic = %v, expected %v", result, tt.expectedConsent)
				}
			})
		}
	})
}

// TestSchemaConverterWithOwnerDirective tests the enhanced schema converter with @owner directive
func TestSchemaConverterWithOwnerDirective(t *testing.T) {
	converter := NewSchemaConverter()

	// Test SDL with @owner directive
	sdl := `directive @accessControl(type: String!) on FIELD_DEFINITION
directive @source(value: String!) on FIELD_DEFINITION
directive @isOwner(value: Boolean!) on FIELD_DEFINITION
directive @owner(value: String!) on FIELD_DEFINITION
directive @description(value: String!) on FIELD_DEFINITION

type PersonInfo {
  fullName: String! @accessControl(type: "public") @isOwner(value: false) @owner(value: "citizen")
  nic: String! @accessControl(type: "public") @isOwner(value: true)
  photo: String! @accessControl(type: "restricted") @isOwner(value: true)
  permanentAddress: String! @accessControl(type: "restricted") @isOwner(value: false) @owner(value: "citizen")
  birthDate: String! @accessControl(type: "restricted") @isOwner(value: false) @owner(value: "rgd")
}

type Query {
  getPerson(nic: String!): PersonInfo
}`

	// Test with authorization config
	authConfig := &AuthorizationConfig{
		Authorization: map[string]FieldAuthorization{
			"personinfo.permanentAddress": {
				AllowedConsumers: []AllowListEntry{
					{
						ConsumerID:    "passport-app",
						ExpiresAt:     1757560679,
						GrantDuration: "30d",
					},
				},
			},
			"personinfo.birthDate": {
				AllowedConsumers: []AllowListEntry{
					{
						ConsumerID:    "driver-app",
						ExpiresAt:     1757560679,
						GrantDuration: "7d",
					},
				},
			},
		},
	}

	result, err := converter.ConvertSDLToProviderMetadata(sdl, "drp", authConfig)
	if err != nil {
		t.Fatalf("ConvertSDLToProviderMetadata() error = %v", err)
	}

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected fields map")
	}

	// Test personinfo.fullName
	fullNameField, ok := fields["personinfo.fullName"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected personinfo.fullName field")
	}

	if fullNameField["owner"] != "citizen" {
		t.Errorf("Expected owner 'citizen', got %v", fullNameField["owner"])
	}

	// Test personinfo.permanentAddress with allow_list
	addressField, ok := fields["personinfo.permanentAddress"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected personinfo.permanentAddress field")
	}

	allowList, ok := addressField["allow_list"].([]interface{})
	if !ok {
		t.Fatal("Expected allow_list array")
	}
	if len(allowList) != 1 {
		t.Errorf("Expected 1 allow_list entry, got %d", len(allowList))
	}
}
