package models

// ConsentEngineOperation represents the operation
type ConsentEngineOperation string

// ConsentEngineOperation constants
const (
	OpCreateConsent         ConsentEngineOperation = "create consent"
	OpUpdateConsent         ConsentEngineOperation = "update consent"
	OpRevokeConsent         ConsentEngineOperation = "revoke consent"
	OpGetConsentStatus      ConsentEngineOperation = "get consent status"
	OpGetConsentsByOwner    ConsentEngineOperation = "get consents by data owner"
	OpGetConsentsByConsumer ConsentEngineOperation = "get consents by consumer"
	OpCheckConsentExpiry    ConsentEngineOperation = "check consent expiry"
	OpProcessPortalRequest  ConsentEngineOperation = "process consent portal"
)
