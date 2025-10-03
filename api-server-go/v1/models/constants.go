package models

// Status constants for submissions and applications
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Entity type constants
const (
	EntityTypeProvider = "provider"
	EntityTypeConsumer = "consumer"
	EntityTypeAdmin    = "admin"
)

// Common field constraints
const (
	MaxNameLength        = 255
	MaxDescriptionLength = 1000
	MaxEmailLength       = 320 // RFC 3696 specification
	MaxPhoneLength       = 15  // E.164 format
	MaxEndpointLength    = 2048
)
