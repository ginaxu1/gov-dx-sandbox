package constants

// Common constants
const (
	StatusMethodNotAllowed  = "Method not allowed"
	StatusIDRequired        = "ID is required"
	StatusConsentIDRequired = "consent_id is required"
)

// Common error messages
const (
	ErrConsentNotFound     = "consent record not found"
	ErrConsentCreateFailed = "failed to create consent record"
	ErrConsentUpdateFailed = "failed to update consent record"
	ErrConsentRevokeFailed = "failed to revoke consent record"
	ErrConsentGetFailed    = "failed to get consent records"
	ErrConsentExpiryFailed = "failed to check consent expiry"
	ErrPortalRequestFailed = "failed to process consent portal request"
)

// Common log operations
const (
	OpCreateConsent         = "create consent"
	OpUpdateConsent         = "update consent"
	OpRevokeConsent         = "revoke consent"
	OpGetConsentStatus      = "get consent status"
	OpGetConsentsByOwner    = "get consents by data owner"
	OpGetConsentsByConsumer = "get consents by consumer"
	OpCheckConsentExpiry    = "check consent expiry"
	OpProcessPortalRequest  = "process consent portal"
)

// Policy Decision Point specific constants
const (
	StatusInvalidJSON  = "Invalid JSON input"
	StatusPolicyFailed = "Failed to evaluate policy"
	StatusDebugFailed  = "Failed to check debug data"
)

// Policy Decision Point specific error messages
const (
	ErrConsumerIDRequired     = "consumer ID is required"
	ErrResourceRequired       = "request resource is required"
	ErrActionRequired         = "request action is required"
	ErrDataFieldsRequired     = "data fields are required"
	ErrNoPolicyRulesMatched   = "No policy rules matched the request"
	ErrPolicyEvaluationFailed = "policy evaluation failed"
	ErrInvalidInput           = "Invalid input"
)

// Policy Decision Point specific log operations
const (
	OpPolicyEvaluation = "policy evaluation"
	OpDebugData        = "debug data check"
	OpDecisionSent     = "decision sent"
)
