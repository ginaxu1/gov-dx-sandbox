package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// SchemaConverter handles conversion from GraphQL SDL to provider metadata
type SchemaConverter struct {
	metadataPath string
}

// NewSchemaConverter creates a new schema converter
func NewSchemaConverter() *SchemaConverter {
	return &SchemaConverter{
		metadataPath: "../../exchange/policy-decision-point/data/provider-metadata.json",
	}
}

// ProviderMetadata represents the structure of provider-metadata.json
type ProviderMetadata struct {
	Fields map[string]ProviderField `json:"fields"`
}

// ProviderField represents a single field in provider metadata
type ProviderField struct {
	Owner             string                 `json:"owner"`
	Provider          string                 `json:"provider"`
	ConsentRequired   bool                   `json:"consent_required"`
	AccessControlType string                 `json:"access_control_type"`
	AllowList         []AllowListEntry       `json:"allow_list"`
	Description       string                 `json:"description,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// AllowListEntry represents an entry in the allow list
type AllowListEntry struct {
	ConsumerID string `json:"consumer_id"`
	ExpiryTime string `json:"expiry_time"`
}

// ConvertSDLToProviderMetadata converts GraphQL SDL to provider metadata format
func (sc *SchemaConverter) ConvertSDLToProviderMetadata(providerID, sdl string) (*ProviderMetadata, error) {
	// Simple SDL parsing - in a real implementation, you'd use a proper GraphQL parser
	fields := make(map[string]ProviderField)

	// Extract type definitions and their fields
	lines := strings.Split(sdl, "\n")
	var currentType string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for type definitions
		if strings.HasPrefix(line, "type ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentType = parts[1]
			}
			continue
		}

		// Look for field definitions
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") && currentType != "" {
			fieldName := sc.extractFieldName(line)
			if fieldName != "" {
				accessControl := sc.extractAccessControl(line)
				consentRequired := sc.extractConsentRequired(line)
				isOwner := sc.extractIsOwner(line)

				// Create field path
				fullFieldName := fmt.Sprintf("%s.%s", strings.ToLower(currentType), fieldName)

				// Determine owner based on isOwner directive
				owner := providerID
				if !isOwner {
					owner = "external" // In a real system, this would be determined by business logic
				}

				fields[fullFieldName] = ProviderField{
					Owner:             owner,
					Provider:          providerID,
					ConsentRequired:   consentRequired,
					AccessControlType: accessControl,
					AllowList:         []AllowListEntry{},
					Description:       sc.extractDescription(line),
				}
			}
		}
	}

	return &ProviderMetadata{Fields: fields}, nil
}

// UpdateProviderMetadataFile updates the provider-metadata.json file
func (sc *SchemaConverter) UpdateProviderMetadataFile(providerID, sdl string) error {
	// Convert SDL to provider metadata
	metadata, err := sc.ConvertSDLToProviderMetadata(providerID, sdl)
	if err != nil {
		return fmt.Errorf("failed to convert SDL: %w", err)
	}

	// Read existing metadata
	existingMetadata, err := sc.loadExistingMetadata()
	if err != nil {
		slog.Warn("Could not load existing metadata, creating new", "error", err)
		existingMetadata = &ProviderMetadata{Fields: make(map[string]ProviderField)}
	}

	// Merge new fields with existing ones
	for fieldName, field := range metadata.Fields {
		existingMetadata.Fields[fieldName] = field
	}

	// Write updated metadata back to file
	return sc.saveMetadata(existingMetadata)
}

// loadExistingMetadata loads existing provider metadata from file
func (sc *SchemaConverter) loadExistingMetadata() (*ProviderMetadata, error) {
	// Try to find the metadata file
	possiblePaths := []string{
		sc.metadataPath,
		"../../exchange/policy-decision-point/data/provider-metadata.json",
		"../exchange/policy-decision-point/data/provider-metadata.json",
		"./exchange/policy-decision-point/data/provider-metadata.json",
	}

	var data []byte
	var err error

	for _, path := range possiblePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			data, err = ioutil.ReadFile(path)
			if err == nil {
				break
			}
		}
	}

	if data == nil {
		return nil, fmt.Errorf("could not find provider-metadata.json file")
	}

	var metadata ProviderMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse existing metadata: %w", err)
	}

	return &metadata, nil
}

// saveMetadata saves provider metadata to file
func (sc *SchemaConverter) saveMetadata(metadata *ProviderMetadata) error {
	// Try to find the metadata file
	possiblePaths := []string{
		sc.metadataPath,
		"../../exchange/policy-decision-point/data/provider-metadata.json",
		"../exchange/policy-decision-point/data/provider-metadata.json",
		"./exchange/policy-decision-point/data/provider-metadata.json",
	}

	var targetPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(filepath.Dir(path)); err == nil {
			targetPath = path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("could not find target directory for provider-metadata.json")
	}

	// Write metadata to file
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := ioutil.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	slog.Info("Updated provider-metadata.json", "path", targetPath, "fields", len(metadata.Fields))
	return nil
}

// Helper functions for parsing SDL
func (sc *SchemaConverter) extractFieldName(line string) string {
	// Extract field name from line like "  fieldName: Type! @directive"
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	fieldPart := parts[0]
	if strings.Contains(fieldPart, ":") {
		return strings.Split(fieldPart, ":")[0]
	}

	return fieldPart
}

func (sc *SchemaConverter) extractAccessControl(line string) string {
	// Look for @accessControl(type: "public"|"restricted")
	if strings.Contains(line, "@accessControl") {
		if strings.Contains(line, "type: \"public\"") {
			return "public"
		}
		if strings.Contains(line, "type: \"restricted\"") {
			return "restricted"
		}
	}
	return "public" // default
}

func (sc *SchemaConverter) extractConsentRequired(line string) bool {
	// Look for @consentRequired or similar directives
	// For now, assume restricted fields require consent
	return strings.Contains(line, "type: \"restricted\"")
}

func (sc *SchemaConverter) extractIsOwner(line string) bool {
	// Look for @isOwner(value: true|false)
	if strings.Contains(line, "@isOwner") {
		return strings.Contains(line, "value: true")
	}
	return true // default to true
}

func (sc *SchemaConverter) extractDescription(line string) string {
	// Look for @description(value: "...")
	if strings.Contains(line, "@description") {
		// Simple extraction - in real implementation, use proper parsing
		start := strings.Index(line, "value: \"")
		if start != -1 {
			start += 8 // len("value: \"")
			end := strings.Index(line[start:], "\"")
			if end != -1 {
				return line[start : start+end]
			}
		}
	}
	return ""
}
