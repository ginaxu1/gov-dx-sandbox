package client

import "github.com/gov-dx-sandbox/audit-service/models"

// Re-export types from models package for convenience
// This avoids duplication while keeping the client package API clean

// DataExchangeEvent is re-exported from models package
type DataExchangeEvent = models.DataExchangeEvent

// ManagementEventRequest is re-exported from models package
// Note: We use ManagementEventRequest (the API request type) instead of duplicating
type ManagementEventRequest = models.ManagementEventRequest

// Actor is re-exported from models package
type Actor = models.Actor

// Target is re-exported from models package
type Target = models.Target
