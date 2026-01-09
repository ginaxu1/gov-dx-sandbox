package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/pkg/graphql"
	"github.com/google/uuid"
	auditpkg "github.com/gov-dx-sandbox/shared/audit"
)

// Context key for audit metadata
type contextKey string

const auditMetadataKey contextKey = "auditMetadata"

// Context key for trace ID
type traceIDKey struct{}

// Metadata holds metadata needed for audit logging in orchestration-engine
type Metadata struct {
	ConsumerAppID    string
	ProviderFieldMap *[]ProviderLevelFieldRecord
}

// ProviderLevelFieldRecord represents a field record for provider-level operations
// This is imported from federator package context
type ProviderLevelFieldRecord struct {
	SchemaId   string
	ServiceKey string
	FieldPath  string
}

// NewContextWithMetadata creates a new context with audit metadata
func NewContextWithMetadata(ctx context.Context, metadata *Metadata) context.Context {
	return context.WithValue(ctx, auditMetadataKey, metadata)
}

// MetadataFromContext retrieves audit metadata from context
func MetadataFromContext(ctx context.Context) *Metadata {
	metadata, ok := ctx.Value(auditMetadataKey).(*Metadata)
	if !ok {
		return nil
	}
	return metadata
}

// GetTraceIDFromContext retrieves the trace ID from the context
// Returns empty string if trace ID is not found in context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// FederationServiceRequest represents a service request for audit logging
type FederationServiceRequest struct {
	ServiceKey     string
	SchemaID       string
	GraphQLRequest graphql.Request
}

// LogProviderFetch logs a provider fetch event to the audit service asynchronously
func LogProviderFetch(ctx context.Context, providerSchemaID string, req *FederationServiceRequest, response *graphql.Response, err error) {
	// Retrieve metadata from context
	metadata := MetadataFromContext(ctx)
	if metadata == nil {
		logger.Log.Warn("Audit metadata missing from context, skipping audit log")
		return
	}

	// Extract requested fields for this provider
	requestedFields := make([]string, 0)
	if metadata.ProviderFieldMap != nil {
		for _, field := range *metadata.ProviderFieldMap {
			if field.SchemaId == req.SchemaID && field.ServiceKey == req.ServiceKey {
				requestedFields = append(requestedFields, field.FieldPath)
			}
		}
	}

	// Create audit request using the new v1 API structure
	traceID := GetTraceIDFromContext(ctx)
	if traceID == "" {
		// If no trace ID in context, generate one (fallback)
		traceID = uuid.New().String()
	}
	eventType := "PROVIDER_FETCH"
	actorType := "SERVICE"
	actorID := "orchestration-engine"
	targetType := "SERVICE"
	targetID := req.ServiceKey

	// Combine requested data and additional info into response metadata
	// (since we're logging after receiving the response)
	responseMetadata := map[string]interface{}{
		"applicationId":   metadata.ConsumerAppID,
		"schemaId":        providerSchemaID,
		"requestedFields": requestedFields,
		"query":           req.GraphQLRequest.Query,
		"serviceKey":      req.ServiceKey,
	}
	if err != nil {
		responseMetadata["error"] = err.Error()
	}
	if response != nil {
		responseMetadata["hasErrors"] = len(response.Errors) > 0
		if len(response.Errors) > 0 {
			responseMetadata["errorCount"] = len(response.Errors)
			// Include first few errors (limit to avoid large payloads)
			errorDetails := make([]interface{}, 0)
			for i, err := range response.Errors {
				if i >= 3 { // Limit to first 3 errors
					break
				}
				errorDetails = append(errorDetails, err)
			}
			responseMetadata["errors"] = errorDetails
		}
		if response.Data != nil {
			// Include data keys for reference (not full data to avoid large payloads)
			dataKeys := make([]string, 0, len(response.Data))
			for key := range response.Data {
				dataKeys = append(dataKeys, key)
			}
			responseMetadata["dataKeys"] = dataKeys
		}
	}
	responseMetadataJSON, jsonErr := json.Marshal(responseMetadata)
	if jsonErr != nil {
		logger.Log.Error("Failed to marshal response metadata for audit", "error", jsonErr)
		responseMetadataJSON = []byte("{}")
	}

	auditStatus := auditpkg.StatusSuccess
	if err != nil || (response != nil && len(response.Errors) > 0) {
		auditStatus = auditpkg.StatusFailure
	}

	auditRequest := &auditpkg.AuditLogRequest{
		TraceID:          &traceID,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		EventType:        &eventType,
		Status:           auditStatus,
		ActorType:        actorType,
		ActorID:          actorID,
		TargetType:       targetType,
		TargetID:         &targetID,
		ResponseMetadata: json.RawMessage(responseMetadataJSON),
	}

	// Log the audit event asynchronously using the global audit package
	auditpkg.LogAuditEvent(ctx, auditRequest)
}
