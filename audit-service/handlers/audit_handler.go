package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"github.com/gov-dx-sandbox/audit-service/services"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	service services.AuditService
}

// NewAuditHandler creates a new instance of AuditHandler
func NewAuditHandler(service services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// CreateAuditLog handles the POST request to create a new audit log
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.TraceID == "" || req.SourceService == "" || req.EventType == "" || req.Status == "" {
		http.Error(w, "Missing required fields (traceId, sourceService, eventType, status)", http.StatusBadRequest)
		return
	}

	// Validate Status value
	if req.Status != "SUCCESS" && req.Status != "FAILURE" {
		http.Error(w, "Status must be either 'SUCCESS' or 'FAILURE'", http.StatusBadRequest)
		return
	}

	// Parse timestamp or use current time
	timestamp := time.Now()
	if req.Timestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			timestamp = parsedTime
		}
	}

	auditLog := &models.AuditLog{
		TraceID:       req.TraceID,
		Timestamp:     timestamp,
		SourceService: req.SourceService,
		TargetService: req.TargetService,
		EventType:     req.EventType,
		Status:        req.Status,
		ActorID:       req.ActorID,
		Resources:     req.Resources,
		Metadata:      req.Metadata,
	}

	createdLog, err := h.service.CreateAuditLog(auditLog)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to create audit log",
			"error", err,
			"traceId", req.TraceID,
			"sourceService", req.SourceService,
			"eventType", req.EventType)

		http.Error(w, "Failed to create audit log", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdLog); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// GetAuditLogs handles the GET request to retrieve logs for a trace
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	traceID := r.URL.Query().Get("traceId")
	if traceID == "" {
		http.Error(w, "Missing traceId query parameter", http.StatusBadRequest)
		return
	}

	logs, err := h.service.GetAuditLogs(traceID)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to retrieve audit logs",
			"error", err,
			"traceId", traceID)

		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
