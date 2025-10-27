package consumer

import (
	"context"
	"fmt"
	"log"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// DatabaseEventProcessor implements the AuditEventProcessor interface.
// It's the "glue" that takes a message from Redis and saves it to the DB.
type DatabaseEventProcessor struct {
	auditService *services.AuditService
}

// NewDatabaseEventProcessor creates a new processor.
func NewDatabaseEventProcessor(auditService *services.AuditService) *DatabaseEventProcessor {
	return &DatabaseEventProcessor{auditService: auditService}
}

// ProcessAuditEvent parses the Redis message and saves it to the database.
func (p *DatabaseEventProcessor) ProcessAuditEvent(ctx context.Context, event map[string]string) error {
	// Parse the map[string]string from Redis into our Go struct
	// Handle both 'status' and 'transaction_status' field names for backward compatibility
	status := event["transaction_status"]
	if status == "" {
		status = event["status"]
	}

	logEntry := &models.LogRequest{
		Status:        status,
		RequestedData: event["requested_data"],
		ApplicationID: event["consumer_id"], // Map consumer_id to application_id
		SchemaID:      event["provider_id"], // Map provider_id to schema_id
		// New fields for M2M vs User differentiation
		RequestType: event["request_type"],
		AuthMethod:  event["auth_method"],
		UserID:      event["user_id"],
		SessionID:   event["session_id"],
	}

	// Validate required fields
	if logEntry.Status == "" {
		return fmt.Errorf("missing required field: status or transaction_status")
	}
	if logEntry.RequestedData == "" {
		return fmt.Errorf("missing required field: requested_data")
	}
	if logEntry.ApplicationID == "" {
		return fmt.Errorf("missing required field: consumer_id")
	}
	if logEntry.SchemaID == "" {
		return fmt.Errorf("missing required field: provider_id")
	}

	// Validate status
	if logEntry.Status != "success" && logEntry.Status != "failure" {
		return fmt.Errorf("invalid status: %s (must be 'success' or 'failure')", logEntry.Status)
	}

	log.Printf("Saving event %s to database...", event["event_id"])

	// Call the existing database logic
	_, err := p.auditService.CreateLog(ctx, logEntry)
	if err != nil {
		return fmt.Errorf("failed to save log %s to database: %w", event["event_id"], err)
	}

	log.Printf("Successfully saved event %s", event["event_id"])
	return nil
}
