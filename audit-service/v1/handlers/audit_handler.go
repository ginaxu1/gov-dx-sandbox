package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"github.com/gov-dx-sandbox/audit-service/v1/services"
)

const (
	// maxRequestSize limits the request body size to 1MB
	maxRequestSize = 1 << 20 // 1MB
	// maxServiceNameLength limits service name length
	maxServiceNameLength = 50
	// maxEventTypeLength limits event type length
	maxEventTypeLength = 50
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	service services.AuditService
}

// NewAuditHandler creates a new instance of AuditHandler
func NewAuditHandler(service services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// errorResponse represents a structured error response
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// writeError writes a structured error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}

// writeJSON writes a JSON response with proper error handling
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

// CreateAuditLog handles the POST request to create a new audit log
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	var req models.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			writeError(w, http.StatusBadRequest, "Invalid JSON syntax")
		case errors.As(err, &unmarshalTypeError):
			writeError(w, http.StatusBadRequest, "Invalid JSON type")
		case errors.Is(err, io.EOF):
			writeError(w, http.StatusBadRequest, "Request body is empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			writeError(w, http.StatusBadRequest, "Request body contains invalid JSON")
		default:
			writeError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Validate required fields
	if err := validateCreateRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse TraceID to UUID
	traceUUID, err := uuid.Parse(req.TraceID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid traceId format (expected UUID)")
		return
	}

	// Parse timestamp or use current time
	timestamp := time.Now()
	if req.Timestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid timestamp format (expected RFC3339)")
			return
		}
		timestamp = parsedTime
	}

	auditLog := &models.AuditLog{
		TraceID:       traceUUID,
		Timestamp:     timestamp,
		SourceService: strings.TrimSpace(req.SourceService),
		TargetService: strings.TrimSpace(req.TargetService),
		EventType:     strings.TrimSpace(req.EventType),
		Status:        req.Status,
		ActorID:       req.ActorID,
		Resources:     req.Resources,
		Metadata:      req.Metadata,
	}

	createdLog, err := h.service.CreateAuditLog(r.Context(), auditLog)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to create audit log",
			"error", err,
			"traceId", req.TraceID,
			"sourceService", req.SourceService,
			"targetService", req.TargetService,
			"eventType", req.EventType,
			"status", req.Status)

		writeError(w, http.StatusInternalServerError, "Failed to create audit log")
		return
	}

	// Log successful creation for observability
	slog.Info("Audit log created",
		"traceId", createdLog.TraceID,
		"sourceService", createdLog.SourceService,
		"targetService", createdLog.TargetService,
		"eventType", createdLog.EventType,
		"status", createdLog.Status,
		"id", createdLog.ID)

	writeJSON(w, http.StatusCreated, createdLog)
}

// validateCreateRequest validates the create audit log request
func validateCreateRequest(req *models.CreateAuditLogRequest) error {
	if req.TraceID == "" {
		return errors.New("traceId is required")
	}
	if req.SourceService == "" {
		return errors.New("sourceService is required")
	}
	if req.EventType == "" {
		return errors.New("eventType is required")
	}
	if req.Status == "" {
		return errors.New("status is required")
	}

	// Validate Status value
	if req.Status != models.StatusSuccess && req.Status != models.StatusFailure {
		return errors.New("status must be either 'SUCCESS' or 'FAILURE'")
	}

	// Validate string lengths
	if len(req.SourceService) > maxServiceNameLength {
		return errors.New("sourceService exceeds maximum length")
	}
	if req.TargetService != "" && len(req.TargetService) > maxServiceNameLength {
		return errors.New("targetService exceeds maximum length")
	}
	if len(req.EventType) > maxEventTypeLength {
		return errors.New("eventType exceeds maximum length")
	}

	return nil
}

// GetAuditLogs handles the GET request to retrieve logs for a trace
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	traceIDStr := r.URL.Query().Get("traceId")
	if traceIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing traceId query parameter")
		return
	}

	traceUUID, err := uuid.Parse(traceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid traceId format (expected UUID)")
		return
	}

	logs, err := h.service.GetAuditLogs(r.Context(), traceUUID)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to retrieve audit logs",
			"error", err,
			"traceId", traceIDStr)

		writeError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Return empty array instead of null if no logs found
	if logs == nil {
		logs = []models.AuditLog{}
	}

	// Log successful retrieval for observability
	slog.Info("Audit logs retrieved",
		"traceId", traceUUID,
		"count", len(logs))

	writeJSON(w, http.StatusOK, logs)
}
