package models

// DataExchangeEvent represents a data exchange audit event
// This is the event structure sent by the Orchestration Engine
type DataExchangeEvent struct {
	EventID           string   `json:"eventId"`           // UUID
	Timestamp         string   `json:"timestamp"`         // ISO 8601 timestamp
	ActorUserID       string   `json:"actorUserId"`       // User who is requesting (consumer subscriber)
	ConsumerAppID     string   `json:"consumerAppId"`     // Consumer application ID (e.g., "passport-app")
	OnBehalfOfOwnerID string   `json:"onBehalfOfOwnerId"` // Citizen ID (data owner)
	ProviderSchemaID  string   `json:"providerSchemaId"`  // Provider schema ID
	RequestedFields   []string `json:"requestedFields"`   // List of requested fields
	Status            string   `json:"status"`            // "SUCCESS" or "FAILURE"
}
