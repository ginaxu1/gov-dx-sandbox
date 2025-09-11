package main

import (
	"fmt"
	"regexp"
	"strings"
)

// GraphQLField represents a field in the GraphQL schema
type GraphQLField struct {
	Name          string
	Type          string
	AccessControl string
	Source        string
	IsOwner       bool
	Description   string
	ParentType    string
}

// SchemaConverter converts GraphQL SDL to provider metadata
type SchemaConverter struct{}

// NewSchemaConverter creates a new schema converter
func NewSchemaConverter() *SchemaConverter {
	return &SchemaConverter{}
}

// ConvertSDLToProviderMetadata converts GraphQL SDL to provider metadata format
func (sc *SchemaConverter) ConvertSDLToProviderMetadata(sdl string, providerID string) (map[string]interface{}, error) {
	fields, err := sc.parseSDL(sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	metadata := map[string]interface{}{
		"fields": make(map[string]interface{}),
	}

	fieldsMap := metadata["fields"].(map[string]interface{})

	for _, field := range fields {
		// Skip Query type fields as they are not data fields
		if field.ParentType == "Query" {
			continue
		}

		// Create field path (e.g., "user.name", "birthInfo.birthDate")
		fieldPath := field.Name
		if field.ParentType != "" && field.ParentType != "Query" {
			fieldPath = strings.ToLower(field.ParentType) + "." + field.Name
		}

		// Determine consent requirement based on access control and ownership
		consentRequired := sc.determineConsentRequired(field)

		// Determine access control type
		accessControlType := field.AccessControl
		if accessControlType == "" {
			accessControlType = "restricted" // Default to restricted if not specified
		}

		// Create field metadata
		fieldMetadata := map[string]interface{}{
			"consent_required":    consentRequired,
			"owner":               providerID, // Provider is the owner by default
			"provider":            providerID,
			"access_control_type": accessControlType,
			"allow_list":          []interface{}{}, // Empty by default, to be populated by admin
		}

		// Add description if available
		if field.Description != "" {
			fieldMetadata["description"] = field.Description
		}

		fieldsMap[fieldPath] = fieldMetadata
	}

	return metadata, nil
}

// parseSDL parses GraphQL SDL and extracts field information
func (sc *SchemaConverter) parseSDL(sdl string) ([]GraphQLField, error) {
	var fields []GraphQLField

	// Remove comments and normalize whitespace
	sdl = sc.cleanSDL(sdl)

	// Find all type definitions
	typeRegex := regexp.MustCompile(`type\s+(\w+)\s*\{([^}]+)\}`)
	typeMatches := typeRegex.FindAllStringSubmatch(sdl, -1)

	for _, typeMatch := range typeMatches {
		typeName := typeMatch[1]
		typeBody := typeMatch[2]

		// Skip Query type for now as it's not a data type
		if typeName == "Query" {
			continue
		}

		// Parse fields within the type
		fieldLines := strings.Split(typeBody, "\n")
		for _, line := range fieldLines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			field := sc.parseFieldLine(line, typeName)
			if field != nil {
				fields = append(fields, *field)
			}
		}
	}

	return fields, nil
}

// parseFieldLine parses a single field line and extracts directives
func (sc *SchemaConverter) parseFieldLine(line, parentType string) *GraphQLField {
	// Extract field name and type (including ! and array notation)
	fieldRegex := regexp.MustCompile(`(\w+):\s*([^@]+)`)
	fieldMatch := fieldRegex.FindStringSubmatch(line)
	if len(fieldMatch) < 3 {
		return nil
	}

	fieldName := fieldMatch[1]
	fieldType := strings.TrimSpace(fieldMatch[2])

	// Extract directives
	accessControl := sc.extractDirective(line, "@accessControl")
	source := sc.extractDirective(line, "@source")
	isOwner := sc.extractDirective(line, "@isOwner")
	description := sc.extractDirective(line, "@description")

	// Parse isOwner boolean
	isOwnerBool := false
	if isOwner == "true" {
		isOwnerBool = true
	}

	return &GraphQLField{
		Name:          fieldName,
		Type:          fieldType,
		AccessControl: accessControl,
		Source:        source,
		IsOwner:       isOwnerBool,
		Description:   description,
		ParentType:    parentType,
	}
}

// extractDirective extracts a directive value from a field line
func (sc *SchemaConverter) extractDirective(line, directive string) string {
	pattern := directive + `\([^)]*\)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllString(line, -1)

	if len(matches) == 0 {
		return ""
	}

	// Extract value from directive (for @source, @description)
	valueRegex := regexp.MustCompile(`\([^)]*value:\s*"([^"]*)"`)
	valueMatch := valueRegex.FindStringSubmatch(matches[0])
	if len(valueMatch) >= 2 {
		return valueMatch[1]
	}

	// Try to extract type value for @accessControl
	typeRegex := regexp.MustCompile(`\([^)]*type:\s*"([^"]*)"`)
	typeMatch := typeRegex.FindStringSubmatch(matches[0])
	if len(typeMatch) >= 2 {
		return typeMatch[1]
	}

	// Try to extract boolean value for @isOwner
	boolRegex := regexp.MustCompile(`\([^)]*value:\s*(true|false)`)
	boolMatch := boolRegex.FindStringSubmatch(matches[0])
	if len(boolMatch) >= 2 {
		return boolMatch[1]
	}

	return ""
}

// determineConsentRequired determines if a field requires consent
func (sc *SchemaConverter) determineConsentRequired(field GraphQLField) bool {
	// Consent is required if:
	// 1. Field is not owned by the provider (isOwner: false)
	// 2. AND field has restricted access control
	return !field.IsOwner && field.AccessControl == "restricted"
}

// cleanSDL removes comments and normalizes whitespace
func (sc *SchemaConverter) cleanSDL(sdl string) string {
	// Remove single-line comments
	lines := strings.Split(sdl, "\n")
	var cleanedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") && line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}
