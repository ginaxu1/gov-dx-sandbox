package policy

// OwnerType represents the owner enum (matches PolicyDecisionPoint Owner type)
type OwnerType string

const (
	OwnerCitizen OwnerType = "citizen"
)

// Endpoint paths
const (
	policyDecisionEndpointPath = "/api/v1/policy/decide"
)
