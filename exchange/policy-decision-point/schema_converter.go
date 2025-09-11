package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GraphQLField represents a field in the GraphQL schema
type GraphQLField struct {
	Name          string
	Type          string
	AccessControl string
	Source        string
	IsOwner       bool
	Owner         string // NEW: explicit owner from @owner directive
	Description   string
	ParentType    string
}

// AuthorizationConfig represents the authorization configuration for fields
type AuthorizationConfig struct {
	FieldOwners     map[string]string            `json:"field_owners,omitempty"`
	Authorization   map[string]FieldAuthorization `json:"authorization,omitempty"`
}

// FieldAuthorization represents authorization data for a specific field
type FieldAuthorization struct {
	AllowedConsumers []AllowListEntry `json:"allowed_consumers"`
}

// AllowListEntry represents an entry in the allow list
type AllowListEntry struct {
	ConsumerID    string `json:"consumerId"`
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
		"consumerId":     consumerID,
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

		// Determine the actual owner
		actualOwner := sc.determineOwner(field, providerID, authConfig)

		// Determine consent requirement based on owner vs provider
		consentRequired := actualOwner != providerID && field.AccessControl == "restricted"

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

		// Create field metadata
		fieldMetadata := map[string]interface{}{
			"consent_required":    consentRequired,
			"owner":               actualOwner,
			"provider":            providerID,
			"access_control_type": accessControlType,
			"allow_list":          allowList,
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
	owner := sc.extractDirective(line, "@owner")
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
		Owner:         owner,
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

// determineOwner determines the actual owner of a field
func (sc *SchemaConverter) determineOwner(field GraphQLField, providerID string, authConfig *AuthorizationConfig) string {
	// Priority 1: @owner directive in SDL
	if field.Owner != "" {
		return field.Owner
	}
	
	// Priority 2: field_owners in authorization config
	if authConfig != nil && authConfig.FieldOwners != nil {
		fieldPath := field.Name
		if field.ParentType != "" && field.ParentType != "Query" {
			fieldPath = strings.ToLower(field.ParentType) + "." + field.Name
		}
		if owner, exists := authConfig.FieldOwners[fieldPath]; exists {
			return owner
		}
	}
	
	// Priority 3: isOwner directive
	if field.IsOwner {
		return providerID // Provider owns the data
	}
	
	// Default fallback
	return "unknown"
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

// ConvertSDLToProviderMetadataLegacy converts GraphQL SDL to provider metadata format (legacy method for backward compatibility)
func (sc *SchemaConverter) ConvertSDLToProviderMetadataLegacy(sdl string, providerID string) (map[string]interface{}, error) {
	return sc.ConvertSDLToProviderMetadata(sdl, providerID, nil)
}
