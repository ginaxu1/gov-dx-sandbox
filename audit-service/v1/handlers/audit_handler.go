package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gov-dx-sandbox/audit-service/v1/services"
	v1types "github.com/gov-dx-sandbox/audit-service/v1/types"
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

	var req v1types.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EventName == "" {
		http.Error(w, "eventName is required", http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}
	if req.ActorType == "" {
		http.Error(w, "actorType is required", http.StatusBadRequest)
		return
	}
	if req.TargetType == "" {
		http.Error(w, "targetType is required", http.StatusBadRequest)
		return
	}

	auditLog, err := h.service.CreateAuditLog(&req)
	if err != nil {
		http.Error(w, "Failed to create audit log: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(auditLog)
}

// GetAuditLogs handles GET /api/audit-logs
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	traceID := r.URL.Query().Get("traceId")
	eventName := r.URL.Query().Get("eventName")
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

	var eventNamePtr *string
	if eventName != "" {
		eventNamePtr = &eventName
	}

	logs, total, err := h.service.GetAuditLogs(traceIDPtr, eventNamePtr, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := v1types.GetAuditLogsResponse{
		Logs:  make([]v1types.AuditLogResponse, len(logs)),
		Total: int(total),
	}

	for i, log := range logs {
		response.Logs[i] = v1types.AuditLogResponse{
			ID:                log.ID,
			Timestamp:         log.Timestamp,
			TraceID:           log.TraceID,
			EventName:         log.EventName,
			EventType:         log.EventType,
			Status:            log.Status,
			ActorType:         log.ActorType,
			ActorServiceName:  log.ActorServiceName,
			ActorUserID:       log.ActorUserID,
			ActorUserType:     log.ActorUserType,
			TargetType:        log.TargetType,
			TargetServiceName: log.TargetServiceName,
			TargetResource:    log.TargetResource,
			TargetResourceID:  log.TargetResourceID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
