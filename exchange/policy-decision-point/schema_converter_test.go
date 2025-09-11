package main

import (
	"testing"
)

func TestSchemaConverter_ConvertSDLToProviderMetadata(t *testing.T) {
	converter := NewSchemaConverter()

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
  birthInfo: BirthInfo @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
}`,
			providerID:   "rgd",
			expectedKeys: []string{"birthinfo.birthDate", "birthinfo.birthPlace", "user.id", "user.birthInfo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := converter.ConvertSDLToProviderMetadata(tt.sdl, tt.providerID)
			if err != nil {
				t.Fatalf("ConvertSDLToProviderMetadata() error = %v", err)
			}

			fields, ok := metadata["fields"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected 'fields' key in metadata")
			}

			// Check that all expected keys are present
			for _, expectedKey := range tt.expectedKeys {
				if _, exists := fields[expectedKey]; !exists {
					t.Errorf("Expected field %s not found in metadata", expectedKey)
				}
			}

			// Check that provider ID is set correctly
			for fieldName, fieldData := range fields {
				fieldMap, ok := fieldData.(map[string]interface{})
				if !ok {
					t.Errorf("Field %s data is not a map", fieldName)
					continue
				}

				if owner, exists := fieldMap["owner"]; !exists || owner != tt.providerID {
					t.Errorf("Field %s owner = %v, want %s", fieldName, owner, tt.providerID)
				}

				if provider, exists := fieldMap["provider"]; !exists || provider != tt.providerID {
					t.Errorf("Field %s provider = %v, want %s", fieldName, provider, tt.providerID)
				}
			}
		})
	}
}

func TestSchemaConverter_DetermineConsentRequired(t *testing.T) {
	converter := NewSchemaConverter()

	tests := []struct {
		name            string
		field           GraphQLField
		expectedConsent bool
	}{
		{
			name: "Owner field with public access - no consent",
			field: GraphQLField{
				Name:          "id",
				AccessControl: "public",
				IsOwner:       true,
			},
			expectedConsent: false,
		},
		{
			name: "Non-owner field with public access - no consent",
			field: GraphQLField{
				Name:          "name",
				AccessControl: "public",
				IsOwner:       false,
			},
			expectedConsent: false,
		},
		{
			name: "Owner field with restricted access - no consent",
			field: GraphQLField{
				Name:          "email",
				AccessControl: "restricted",
				IsOwner:       true,
			},
			expectedConsent: false,
		},
		{
			name: "Non-owner field with restricted access - consent required",
			field: GraphQLField{
				Name:          "ssn",
				AccessControl: "restricted",
				IsOwner:       false,
			},
			expectedConsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.determineConsentRequired(tt.field)
			if result != tt.expectedConsent {
				t.Errorf("determineConsentRequired() = %v, want %v", result, tt.expectedConsent)
			}
		})
	}
}

func TestSchemaConverter_ParseFieldLine(t *testing.T) {
	converter := NewSchemaConverter()

	tests := []struct {
		name       string
		line       string
		parentType string
		expected   *GraphQLField
	}{
		{
			name:       "Field with all directives",
			line:       `id: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true) @description(value: "User ID")`,
			parentType: "User",
			expected: &GraphQLField{
				Name:          "id",
				Type:          "ID!",
				AccessControl: "public",
				Source:        "authoritative",
				IsOwner:       true,
				Description:   "User ID",
				ParentType:    "User",
			},
		},
		{
			name:       "Field with minimal directives",
			line:       `name: String! @accessControl(type: "restricted")`,
			parentType: "User",
			expected: &GraphQLField{
				Name:          "name",
				Type:          "String!",
				AccessControl: "restricted",
				Source:        "",
				IsOwner:       false,
				Description:   "",
				ParentType:    "User",
			},
		},
		{
			name:       "Field without directives",
			line:       `email: String!`,
			parentType: "User",
			expected: &GraphQLField{
				Name:          "email",
				Type:          "String!",
				AccessControl: "",
				Source:        "",
				IsOwner:       false,
				Description:   "",
				ParentType:    "User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.parseFieldLine(tt.line, tt.parentType)
			if result == nil && tt.expected != nil {
				t.Fatal("Expected field, got nil")
			}
			if result != nil && tt.expected == nil {
				t.Fatal("Expected nil, got field")
			}
			if result != nil && tt.expected != nil {
				if *result != *tt.expected {
					t.Errorf("parseFieldLine() = %+v, want %+v", *result, *tt.expected)
				}
			}
		})
	}
}
