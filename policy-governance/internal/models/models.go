// internal/models/models.go
package models

// Classification defines the access classification levels.
type Classification string

const (
	ALLOW                  Classification = "ALLOW"
	ALLOW_PROVIDER_CONSENT Classification = "ALLOW_PROVIDER_CONSENT"
	ALLOW_CITIZEN_CONSENT  Classification = "ALLOW_CITIZEN_CONSENT"
	ALLOW_CONSENT          Classification = "ALLOW_CONSENT" // Generic consent, could be either provider or citizen
	DENIED                 Classification = "DENIED"
)

// Context can hold additional information for policy evaluation.
type Context map[string]interface{}

// RequestedField represents a single field requested by the consumer.
type RequestedField struct {
	SubgraphName   string         `json:"subgraphName"`
	TypeName       string         `json:"typeName"`
	FieldName      string         `json:"fieldName"`
	Classification Classification `json:"classification"` // This will be the policy decision (initial hint)
	Context        Context        `json:"context"`
}

// PolicyRequest is the format of the request coming from the GraphQL Router.
type PolicyRequest struct {
	ConsumerID      string           `json:"consumerId"`
	RequestedFields []RequestedField `json:"requestedFields"`
}

// AccessScope represents the determined access scope for a field.
type AccessScope struct {
	SubgraphName string `json:"subgraphName"`
	TypeName     string `json:"typeName"`
	FieldName    string `json:"fieldName"`
	// ResolvedClassification is the classification determined by the Policy Governance service.
	ResolvedClassification Classification `json:"resolvedClassification"`
	// ConsentRequired indicates if consent is needed for this specific field.
	ConsentRequired bool `json:"consentRequired"`
	// ConsentType specifies the type of consent if required (e.g., "provider" or "citizen").
	ConsentType []string `json:"consentType,omitempty"`
}

// PolicyResponse is the format of the response sent back to the GraphQL Router.
type PolicyResponse struct {
	ConsumerID   string        `json:"consumerId"`
	AccessScopes []AccessScope `json:"accessScopes"`
	// OverallConsentRequired indicates if any of the requested fields require consent.
	OverallConsentRequired bool `json:"overallConsentRequired"`
}

// PolicyRecord represents a simplified policy stored in the database.
// In a real application, this schema would be much more complex to support
// dynamic rules, roles, conditions, etc.
type PolicyRecord struct {
	ID             int            `json:"id"`
	SubgraphName   string         `json:"subgraph_name"`
	TypeName       string         `json:"type_name"`
	FieldName      string         `json:"field_name"`
	Classification Classification `json:"classification"`
	// Add other fields here for more complex policy rules, e.g., consumer_roles, conditions_json
}
