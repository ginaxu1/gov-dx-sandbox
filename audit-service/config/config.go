package config

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// AuditEnums represents the enum configuration for audit service
// All enum values are configurable via YAML for maximum reusability
type AuditEnums struct {
	EventTypes   []string `yaml:"eventTypes"`
	EventActions []string `yaml:"eventActions"`
	ActorTypes   []string `yaml:"actorTypes"`
	TargetTypes  []string `yaml:"targetTypes"`

	// Maps for O(1) validation lookups (initialized from slices)
	eventTypesMap   map[string]struct{}
	eventActionsMap map[string]struct{}
	actorTypesMap   map[string]struct{}
	targetTypesMap  map[string]struct{}

	// initOnce ensures thread-safe lazy initialization of maps
	initOnce sync.Once
}

// Config holds the audit service configuration
type Config struct {
	Enums AuditEnums `yaml:"enums"`
}

var (
	// DefaultEnums provides default enum values if config file is not found
	// Note: OpenDIF-specific event types (ORCHESTRATION_REQUEST_RECEIVED, POLICY_CHECK, CONSENT_CHECK, PROVIDER_FETCH)
	// should be added to config/enums.yaml for project-specific configurations
	DefaultEnums = AuditEnums{
		EventTypes: []string{
			"MANAGEMENT_EVENT",
			"USER_MANAGEMENT",
			"DATA_FETCH",
		},
		EventActions: []string{
			"CREATE",
			"READ",
			"UPDATE",
			"DELETE",
		},
		ActorTypes: []string{
			"SERVICE",
			"ADMIN",
			"MEMBER",
			"SYSTEM",
		},
		TargetTypes: []string{
			"SERVICE",
			"RESOURCE",
		},
	}
)

// LoadEnums loads enum configuration from YAML file
// If the file is not found or cannot be read, returns default enums
func LoadEnums(configPath string) (*AuditEnums, error) {
	// If no path provided, try default location
	if configPath == "" {
		configPath = "config/enums.yaml"
	}

	// Try to read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If file doesn't exist, return defaults
		if os.IsNotExist(err) {
			return GetDefaultEnums(), nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		slog.Warn("Failed to parse config file, using defaults", "path", configPath, "error", err)
		return GetDefaultEnums(), nil
	}

	// Use defaults for any missing enum arrays
	enums := &config.Enums
	if len(enums.EventTypes) == 0 {
		enums.EventTypes = DefaultEnums.EventTypes
	}
	if len(enums.EventActions) == 0 {
		enums.EventActions = DefaultEnums.EventActions
	}
	if len(enums.ActorTypes) == 0 {
		enums.ActorTypes = DefaultEnums.ActorTypes
	}
	if len(enums.TargetTypes) == 0 {
		enums.TargetTypes = DefaultEnums.TargetTypes
	}

	// Initialize maps for O(1) validation lookups
	enums.InitializeMaps()

	return enums, nil
}

// GetDefaultEnums creates a new AuditEnums instance with default values
// Slices are copied to avoid sharing references with the global DefaultEnums
func GetDefaultEnums() *AuditEnums {
	enums := &AuditEnums{
		EventTypes:   append([]string(nil), DefaultEnums.EventTypes...),
		EventActions: append([]string(nil), DefaultEnums.EventActions...),
		ActorTypes:   append([]string(nil), DefaultEnums.ActorTypes...),
		TargetTypes:  append([]string(nil), DefaultEnums.TargetTypes...),
	}
	enums.InitializeMaps()
	return enums
}

// InitializeMaps converts slices to maps for O(1) validation lookups
// Uses sync.Once to ensure thread-safe initialization that happens only once
// This is called automatically by LoadEnums, but can be called manually for testing
func (e *AuditEnums) InitializeMaps() {
	e.initOnce.Do(func() {
		e.eventTypesMap = make(map[string]struct{}, len(e.EventTypes))
		for _, et := range e.EventTypes {
			e.eventTypesMap[et] = struct{}{}
		}

		e.eventActionsMap = make(map[string]struct{}, len(e.EventActions))
		for _, ea := range e.EventActions {
			e.eventActionsMap[ea] = struct{}{}
		}

		e.actorTypesMap = make(map[string]struct{}, len(e.ActorTypes))
		for _, at := range e.ActorTypes {
			e.actorTypesMap[at] = struct{}{}
		}

		e.targetTypesMap = make(map[string]struct{}, len(e.TargetTypes))
		for _, tt := range e.TargetTypes {
			e.targetTypesMap[tt] = struct{}{}
		}
	})
}

// IsValidEventType checks if the given event type is valid
func (e *AuditEnums) IsValidEventType(eventType string) bool {
	if eventType == "" {
		return true // Empty is allowed (nullable field)
	}
	_, exists := e.eventTypesMap[eventType]
	return exists
}

// IsValidEventAction checks if the given event action is valid
func (e *AuditEnums) IsValidEventAction(action string) bool {
	if action == "" {
		return true // Empty is allowed (nullable field)
	}
	_, exists := e.eventActionsMap[action]
	return exists
}

// IsValidActorType checks if the given actor type is valid
func (e *AuditEnums) IsValidActorType(actorType string) bool {
	_, exists := e.actorTypesMap[actorType]
	return exists
}

// IsValidTargetType checks if the given target type is valid
func (e *AuditEnums) IsValidTargetType(targetType string) bool {
	_, exists := e.targetTypesMap[targetType]
	return exists
}

// GetEnvOrDefault returns the environment variable value or a default
// This is a utility function for reading environment variables with defaults
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
