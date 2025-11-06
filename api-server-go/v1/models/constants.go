package models

// UserGroup represents different user groups in the system
type UserGroup string

const (
	UserGroupAdmin  UserGroup = "OpenDIF_Admin"
	UserGroupMember UserGroup = "OpenDIF_Member"
)

// Status represents the status of submissions and applications
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Version represents application versioning states
type Version string

const (
	ActiveVersion     Version = "active"
	DeprecatedVersion Version = "deprecated"
)

// Field length constraints remain as regular constants
const (
	MaxNameLength        = 255
	MaxDescriptionLength = 1000
	MaxEmailLength       = 320 // RFC 3696 specification
	MaxPhoneLength       = 15  // E.164 format
	MaxEndpointLength    = 2048
)
