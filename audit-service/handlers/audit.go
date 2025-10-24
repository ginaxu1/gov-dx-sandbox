package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// AuditLogResponse represents the response structure for audit logs
type AuditLogResponse struct {
	Logs   []AuditLogEntry `json:"logs"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	RequestedData string    `json:"requestedData"`
	ApplicationID string    `json:"applicationId"`
	SchemaID      string    `json:"schemaId"`
	ConsumerID    string    `json:"consumerId,omitempty"`
	ProviderID    string    `json:"providerId,omitempty"`
}

// HandleAuditLogs handles GET requests for audit logs
func (h *AuditHandler) HandleAuditLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.handleGetAuditLogs(w, r)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetAuditLogs handles GET /api/logs requests
func (h *AuditHandler) handleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()

	// Note: Filter parameters are not used in the simplified GORM implementation

	// Parse pagination parameters
	defaultLimit := parseIntOrDefault("AUDIT_DEFAULT_LIMIT", 50)
	maxLimit := parseIntOrDefault("AUDIT_MAX_LIMIT", 1000)
	limit := defaultLimit
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= maxLimit {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Note: Date filters are not used in the simplified GORM implementation

	// Get audit logs from service
	logs, total, err := h.auditService.GetAuditLogs(r.Context(), limit, offset)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve audit logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Create response
	response := models.LogResponse{
		Logs:   logs,
		Total:  total, // Keep as int64
		Limit:  limit,
		Offset: offset,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// parseIntOrDefault gets environment variable as int or returns default value
func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
