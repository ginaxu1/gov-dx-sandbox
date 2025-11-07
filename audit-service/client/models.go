package client

// DataExchangeEvent represents a data exchange audit event (Case 1)
// This is used by the Orchestration Engine
type DataExchangeEvent struct {
	EventID           string   `json:"eventId"`           // UUID
	Timestamp         string   `json:"timestamp"`         // ISO 8601 timestamp
	ActorUserID       string   `json:"actorUserId"`       // User who is requesting (consumer subscriber)
	ConsumerAppID     string   `json:"consumerAppId"`     // Consumer application ID (e.g., "passport-app")
	ConsumerID        string   `json:"consumerId"`        // Member ID who owns the consumer application (REQUIRED)
	OnBehalfOfOwnerID string   `json:"onBehalfOfOwnerId"` // Citizen ID (data owner)
	ProviderSchemaID  string   `json:"providerSchemaId"`  // Provider schema ID
	ProviderID        string   `json:"providerId"`        // Member ID who owns the provider schema (REQUIRED)
	RequestedFields   []string `json:"requestedFields"`   // List of requested fields
	Status            string   `json:"status"`            // "SUCCESS" or "FAILURE"
}

// ManagementEvent represents a management event (Case 2)
// This is used by the API Server
type ManagementEvent struct {
	EventID   string                  `json:"eventId"`             // UUID
	Timestamp *string                 `json:"timestamp,omitempty"` // ISO 8601 timestamp (optional)
	EventType string                  `json:"eventType"`           // "CREATE", "UPDATE", "DELETE", "READ"
	Actor     Actor                   `json:"actor"`
	Target    Target                  `json:"target"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"` // Optional additional context
}

// Actor represents the actor who performed the action
type Actor struct {
	Type string  `json:"type"` // "USER" or "SERVICE"
	ID   *string `json:"id"`   // User ID (null if SERVICE type)
	Role *string `json:"role"` // "MEMBER" or "ADMIN" (null if SERVICE type)
}

// Target represents the resource that was acted upon
type Target struct {
	Resource   string `json:"resource"`   // "MEMBERS", "SCHEMAS", "APPLICATIONS", etc.
	ResourceID string `json:"resourceId"` // The ID of the resource
}
