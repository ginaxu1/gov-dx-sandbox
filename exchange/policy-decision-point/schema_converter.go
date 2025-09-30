package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GraphQLField represents a field in the GraphQL schema (PDP-focused)
type GraphQLField struct {
	Name          string
	Type          string
	AccessControl string
	Description   string
	ParentType    string
	Source        string
	IsOwner       bool
	Owner         string
}

// AuthorizationConfig represents the authorization configuration for fields (PDP-focused)
type AuthorizationConfig struct {
	Authorization map[string]FieldAuthorization `json:"authorization,omitempty"`
}

// FieldAuthorization represents authorization data for a specific field
type FieldAuthorization struct {
	AllowedConsumers []AllowListEntry `json:"allowed_consumers"`
}

// AllowListEntry represents an entry in the allow list
type AllowListEntry struct {
	ConsumerID    string `json:"consumer_id"`
	ExpiresAt     int64  `json:"expires_at"`
	GrantDuration string `json:"grant_duration,omitempty"`
}

// SchemaConverter converts GraphQL SDL to provider metadata
type SchemaConverter struct{}

// NewSchemaConverter creates a new schema converter
func NewSchemaConverter() *SchemaConverter {
	return &SchemaConverter{}
}

// createAllowListEntry creates an allow_list entry with expires_at and grant_duration
func (sc *SchemaConverter) createAllowListEntry(consumerID, grantDuration string) map[string]interface{} {
	// Calculate expires_at as epoch timestamp (30 days from now by default)
	expiresAt := time.Now().Add(30 * 24 * time.Hour).Unix()

	// Parse grant duration if provided
	if grantDuration != "" {
		if duration, err := sc.parseDuration(grantDuration); err == nil {
			expiresAt = time.Now().Add(duration).Unix()
		}
	}

	return map[string]interface{}{
		"consumer_id":    consumerID,
		"expires_at":     expiresAt,
		"grant_duration": grantDuration,
	}
}

// parseDuration parses duration strings like "30d", "7d", "1h"
func (sc *SchemaConverter) parseDuration(duration string) (time.Duration, error) {
	if len(duration) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	unit := duration[len(duration)-1:]
	value := duration[:len(duration)-1]

	var multiplier time.Duration
	switch unit {
	case "d":
		multiplier = 24 * time.Hour
	case "h":
		multiplier = time.Hour
	case "m":
		multiplier = time.Minute
	case "s":
		multiplier = time.Second
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}

	// Parse numeric value
	var num int
	if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", value)
	}

	return time.Duration(num) * multiplier, nil
}

// AddConsumerToAllowList adds a consumer to the allow_list with expires_at and grant_duration
func (sc *SchemaConverter) AddConsumerToAllowList(metadata map[string]interface{}, fieldPath, consumerID, grantDuration string) error {
	fields, ok := metadata["fields"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid metadata structure")
	}

	fieldMetadata, ok := fields[fieldPath].(map[string]interface{})
	if !ok {
		return fmt.Errorf("field %s not found", fieldPath)
	}

	allowList, ok := fieldMetadata["allow_list"].([]interface{})
	if !ok {
		allowList = []interface{}{}
	}

	// Create new allow_list entry
	entry := sc.createAllowListEntry(consumerID, grantDuration)
	allowList = append(allowList, entry)

	// Update the field metadata
	fieldMetadata["allow_list"] = allowList
	fields[fieldPath] = fieldMetadata

	return nil
}

// ConvertSDLToProviderMetadata converts GraphQL SDL to provider metadata format
// Supports both @owner directive and separate authorization configuration
func (sc *SchemaConverter) ConvertSDLToProviderMetadata(sdl string, providerID string, authConfig *AuthorizationConfig) (map[string]interface{}, error) {
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

		// Determine if consent is required based on @isOwner directive
		// If provider is not the owner (@isOwner: false), consent is required for restricted fields
		consentRequired := false
		if !field.IsOwner && field.AccessControl == "restricted" {
			consentRequired = true
		}

		// Determine access control type
		accessControlType := field.AccessControl
		if accessControlType == "" {
			accessControlType = "restricted" // Default to restricted if not specified
		}

		// Get authorization data for this field
		var allowList []interface{}
		if authConfig != nil && authConfig.Authorization != nil {
			if fieldAuth, exists := authConfig.Authorization[fieldPath]; exists {
				for _, consumer := range fieldAuth.AllowedConsumers {
					allowList = append(allowList, map[string]interface{}{
						"consumerId":     consumer.ConsumerID,
						"expires_at":     consumer.ExpiresAt,
						"grant_duration": consumer.GrantDuration,
					})
				}
			}
		}

		// Determine owner: use explicit @owner directive or default to "citizen" if @isOwner: true
		owner := field.Owner
		if owner == "" && field.IsOwner {
			owner = "citizen"
		}

		// Create field metadata (only what PDP needs)
		fieldMetadata := map[string]interface{}{
			"access_control_type": accessControlType,
			"allow_list":          allowList,
			"consent_required":    consentRequired,
			"owner":               owner,
			"provider":            providerID,
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
	description := sc.extractDirective(line, "@description")
	source := sc.extractDirective(line, "@source")
	isOwnerStr := sc.extractDirective(line, "@isOwner")
	owner := sc.extractDirective(line, "@owner")

	// Parse isOwner boolean
	isOwner := false
	if isOwnerStr == "true" {
		isOwner = true
	}

	return &GraphQLField{
		Name:          fieldName,
		Type:          fieldType,
		AccessControl: accessControl,
		Description:   description,
		ParentType:    parentType,
		Source:        source,
		IsOwner:       isOwner,
		Owner:         owner,
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

	// Try to extract boolean value for @isOwner (without quotes)
	boolRegex := regexp.MustCompile(`\([^)]*value:\s*(true|false)`)
	boolMatch := boolRegex.FindStringSubmatch(matches[0])
	if len(boolMatch) >= 2 {
		return boolMatch[1]
	}

	return ""
}

// determineOwner determines the actual owner of a field
// Note: Ownership and consent determination are handled by the Orchestration Engine
// The PDP only needs to know about access control types and consumer authorization

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

