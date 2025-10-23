package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	// Extract filter parameters
	consumerID := queryParams.Get("consumerId")
	providerID := queryParams.Get("providerId")
	status := queryParams.Get("status")
	startDateStr := queryParams.Get("startDate")
	endDateStr := queryParams.Get("endDate")

	// Parse pagination parameters
	limit := 50 // default
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := queryParams.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Parse date filters
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Get audit logs from service
	logs, total, err := h.auditService.GetAuditLogs(services.AuditLogFilter{
		ConsumerID: consumerID,
		ProviderID: providerID,
		Status:     status,
		StartDate:  startDate,
		EndDate:    endDate,
		Limit:      limit,
		Offset:     offset,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve audit logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := AuditLogResponse{
		Logs:   make([]AuditLogEntry, len(logs)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	for i, log := range logs {
		response.Logs[i] = AuditLogEntry{
			ID:            log.ID,
			Timestamp:     log.CreatedAt,
			Status:        log.TransactionStatus,
			RequestedData: log.RequestedData,
			ApplicationID: log.ApplicationID,
			SchemaID:      log.SchemaID,
			ConsumerID:    log.ConsumerID,
			ProviderID:    log.ProviderID,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
