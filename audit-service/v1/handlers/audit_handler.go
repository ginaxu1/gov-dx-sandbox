package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"github.com/gov-dx-sandbox/audit-service/v1/services"
	"github.com/gov-dx-sandbox/audit-service/v1/utils"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	service *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// CreateAuditLog handles POST /api/audit-logs
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validation is handled by the service layer (auditLog.Validate())
	auditLog, err := h.service.CreateAuditLog(r.Context(), &req)
	if err != nil {
		// Return 400 Bad Request for validation errors, 500 for other errors
		if services.IsValidationError(err) {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create audit log", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, auditLog)
}

// GetAuditLogs handles GET /api/audit-logs
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	traceID := r.URL.Query().Get("traceId")
	eventType := r.URL.Query().Get("eventType")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100 // default
	offset := 0  // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var traceIDPtr *string
	if traceID != "" {
		traceIDPtr = &traceID
	}

	var eventTypePtr *string
	if eventType != "" {
		eventTypePtr = &eventType
	}

	logs, total, err := h.service.GetAuditLogs(r.Context(), traceIDPtr, eventTypePtr, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit logs", err)
		return
	}

	response := models.GetAuditLogsResponse{
		Logs:   make([]models.AuditLogResponse, len(logs)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	for i, log := range logs {
		response.Logs[i] = models.ToAuditLogResponse(log)
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
