package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetLogs handles GET /api/logs (for Admin Portal and Entity Portals)
func (h *AuditHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseLogFilterParams(r)

	// Get logs
	response, err := h.auditService.GetLogs(r.Context(), filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseLogFilterParams parses filter parameters from query string for logs
func (h *AuditHandler) parseLogFilterParams(r *http.Request) *models.LogFilter {
	filter := &models.LogFilter{}

	// Parse query parameters
	if consumerID := r.URL.Query().Get("consumerId"); consumerID != "" {
		filter.ConsumerID = consumerID
	}

	if providerID := r.URL.Query().Get("providerId"); providerID != "" {
		filter.ProviderID = providerID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}

	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = endDate
		}
	}

	// Parse pagination parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

// CreateDataExchangeEvent handles POST /v1/audit/exchange (for data exchange logging from OE)
func (h *AuditHandler) CreateDataExchangeEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.DataExchangeEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ConsumerAppID == "" {
		http.Error(w, "Missing required field: consumerAppId", http.StatusBadRequest)
		return
	}

	if req.ProviderSchemaID == "" {
		http.Error(w, "Missing required field: providerSchemaId", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Missing required field: status", http.StatusBadRequest)
		return
	}

	// Validate status
	if req.Status != "SUCCESS" && req.Status != "FAILURE" {
		http.Error(w, "Invalid status. Must be 'SUCCESS' or 'FAILURE'", http.StatusBadRequest)
		return
	}

	// Convert DataExchangeEvent to LogRequest format for storage
	// Map requestedFields to requestedData (JSON string)
	requestedDataJSON, err := json.Marshal(req.RequestedFields)
	if err != nil {
		http.Error(w, "Failed to marshal requestedFields", http.StatusBadRequest)
		return
	}

	logReq := models.LogRequest{
		Status:        convertStatusToLower(req.Status), // Convert "SUCCESS" -> "success"
		RequestedData: string(requestedDataJSON),
		ApplicationID: req.ConsumerAppID,    // Map consumerAppId to applicationId
		SchemaID:      req.ProviderSchemaID, // Map providerSchemaId to schemaId
	}

	// Create log
	log, err := h.auditService.CreateLog(r.Context(), &logReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create log: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the created log
	if err := json.NewEncoder(w).Encode(log); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// convertStatusToLower converts "SUCCESS"/"FAILURE" to "success"/"failure"
func convertStatusToLower(status string) string {
	switch status {
	case "SUCCESS":
		return "success"
	case "FAILURE":
		return "failure"
	default:
		return status
	}
}
