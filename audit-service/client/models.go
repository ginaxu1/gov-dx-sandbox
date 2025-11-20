package client

import "github.com/gov-dx-sandbox/audit-service/models"

// Re-export types from models package for convenience
// This avoids duplication while keeping the client package API clean

// DataExchangeEvent is re-exported from models package
type DataExchangeEvent = models.DataExchangeEvent
