package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"github.com/gov-dx-sandbox/audit-service/v1/services"
	"github.com/gov-dx-sandbox/audit-service/v1/types"
)

const (
	// maxRequestSize limits the request body size to 1MB
	maxRequestSize = 1 << 20 // 1MB
	// maxStringLength limits string field lengths
	maxStringLength = 100
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

	var req types.CreateAuditLogRequest
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

	// Convert request to model
	auditLog, err := convertRequestToModel(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdLog, err := h.service.CreateAuditLog(r.Context(), auditLog)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to create audit log",
			"error", err,
			"traceId", req.TraceID,
			"eventName", req.EventName,
			"status", req.Status)

		writeError(w, http.StatusInternalServerError, "Failed to create audit log")
		return
	}

	// Log successful creation for observability
	slog.Info("Audit log created",
		"traceId", createdLog.TraceID,
		"eventName", createdLog.EventName,
		"status", createdLog.Status,
		"id", createdLog.ID)

	writeJSON(w, http.StatusCreated, createdLog)
}

// GetAuditLogs handles the GET request to retrieve logs with flexible filtering
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	filter := &types.GetAuditLogsRequest{}

	// Trace ID (optional)
	if traceIDStr := r.URL.Query().Get("traceId"); traceIDStr != "" {
		traceUUID, err := uuid.Parse(traceIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid traceId format (expected UUID)")
			return
		}
		filter.TraceID = &traceUUID
	}

	// Event name (optional)
	if eventName := r.URL.Query().Get("eventName"); eventName != "" {
		filter.EventName = &eventName
	}

	// Status (optional)
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	// Actor service name (optional)
	if actorServiceName := r.URL.Query().Get("actorServiceName"); actorServiceName != "" {
		filter.ActorServiceName = &actorServiceName
	}

	// Actor user ID (optional)
	if actorUserIDStr := r.URL.Query().Get("actorUserId"); actorUserIDStr != "" {
		actorUserUUID, err := uuid.Parse(actorUserIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid actorUserId format (expected UUID)")
			return
		}
		filter.ActorUserID = &actorUserUUID
	}

	// Target service name (optional)
	if targetServiceName := r.URL.Query().Get("targetServiceName"); targetServiceName != "" {
		filter.TargetServiceName = &targetServiceName
	}

	// Target resource (optional)
	if targetResource := r.URL.Query().Get("targetResource"); targetResource != "" {
		filter.TargetResource = &targetResource
	}

	// Start time (optional)
	if startTime := r.URL.Query().Get("startTime"); startTime != "" {
		filter.StartTime = &startTime
	}

	// End time (optional)
	if endTime := r.URL.Query().Get("endTime"); endTime != "" {
		filter.EndTime = &endTime
	}

	// Limit (optional)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			writeError(w, http.StatusBadRequest, "Invalid limit (must be positive integer)")
			return
		}
		filter.Limit = &limit
	}

	// Offset (optional)
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			writeError(w, http.StatusBadRequest, "Invalid offset (must be non-negative integer)")
			return
		}
		filter.Offset = &offset
	}

	response, err := h.service.GetAuditLogs(r.Context(), filter)
	if err != nil {
		// Log the error with details for debugging
		slog.Error("Failed to retrieve audit logs",
			"error", err,
			"filter", filter)

		writeError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Log successful retrieval for observability
	slog.Info("Audit logs retrieved",
		"total", response.Total,
		"count", len(response.Events))

	writeJSON(w, http.StatusOK, response)
}

// validateCreateRequest validates the create audit log request
func validateCreateRequest(req *types.CreateAuditLogRequest) error {
	if req.EventName == "" {
		return errors.New("eventName is required")
	}
	if req.Status == "" {
		return errors.New("status is required")
	}
	if req.Status != models.StatusSuccess && req.Status != models.StatusFailure {
		return errors.New("status must be either 'SUCCESS' or 'FAILURE'")
	}
	if req.ActorType == "" {
		return errors.New("actorType is required")
	}
	if req.ActorType != models.ActorTypeUser && req.ActorType != models.ActorTypeService {
		return errors.New("actorType must be either 'USER' or 'SERVICE'")
	}
	if req.TargetType == "" {
		return errors.New("targetType is required")
	}
	if req.TargetType != models.TargetTypeResource && req.TargetType != models.TargetTypeService {
		return errors.New("targetType must be either 'RESOURCE' or 'SERVICE'")
	}

	// Validate actor constraints
	if req.ActorType == models.ActorTypeService {
		if req.ActorServiceName == nil || *req.ActorServiceName == "" {
			return errors.New("actorServiceName is required when actorType is SERVICE")
		}
		if req.ActorUserID != nil {
			return errors.New("actorUserId must be null when actorType is SERVICE")
		}
	} else if req.ActorType == models.ActorTypeUser {
		if req.ActorUserID == nil || *req.ActorUserID == "" {
			return errors.New("actorUserId is required when actorType is USER")
		}
		if req.ActorServiceName != nil && *req.ActorServiceName != "" {
			return errors.New("actorServiceName must be null when actorType is USER")
		}
	}

	// Validate target constraints
	if req.TargetType == models.TargetTypeService {
		if req.TargetServiceName == nil || *req.TargetServiceName == "" {
			return errors.New("targetServiceName is required when targetType is SERVICE")
		}
		if req.TargetResource != nil && *req.TargetResource != "" {
			return errors.New("targetResource must be null when targetType is SERVICE")
		}
	} else if req.TargetType == models.TargetTypeResource {
		if req.TargetResource == nil || *req.TargetResource == "" {
			return errors.New("targetResource is required when targetType is RESOURCE")
		}
		if req.TargetServiceName != nil && *req.TargetServiceName != "" {
			return errors.New("targetServiceName must be null when targetType is RESOURCE")
		}
	}

	// Validate string lengths
	if len(req.EventName) > maxStringLength {
		return errors.New("eventName exceeds maximum length")
	}
	if req.EventType != nil && len(*req.EventType) > 20 {
		return errors.New("eventType exceeds maximum length")
	}

	return nil
}

// convertRequestToModel converts a CreateAuditLogRequest to an AuditLog model
func convertRequestToModel(req *types.CreateAuditLogRequest) (*models.AuditLog, error) {
	log := &models.AuditLog{
		EventName:        strings.TrimSpace(req.EventName),
		Status:           req.Status,
		ActorType:        req.ActorType,
		TargetType:       req.TargetType,
		RequestedData:    req.RequestedData,
		ResponseMetadata: req.ResponseMetadata,
		EventMetadata:    req.EventMetadata,
		ActorMetadata:    req.ActorMetadata,
		TargetMetadata:   req.TargetMetadata,
	}

	// Parse timestamp or use current time
	timestamp := time.Now().UTC()
	if req.Timestamp != nil && *req.Timestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format (expected RFC3339): %w", err)
		}
		timestamp = parsedTime
	}
	log.Timestamp = timestamp

	// Parse trace ID (nullable)
	if req.TraceID != nil && *req.TraceID != "" {
		traceUUID, err := uuid.Parse(*req.TraceID)
		if err != nil {
			return nil, fmt.Errorf("invalid traceId format (expected UUID): %w", err)
		}
		log.TraceID = &traceUUID
	}

	// Parse event type (nullable)
	if req.EventType != nil && *req.EventType != "" {
		eventType := strings.TrimSpace(*req.EventType)
		log.EventType = &eventType
	}

	// Parse actor fields
	if req.ActorServiceName != nil {
		serviceName := strings.TrimSpace(*req.ActorServiceName)
		log.ActorServiceName = &serviceName
	}
	if req.ActorUserID != nil && *req.ActorUserID != "" {
		userUUID, err := uuid.Parse(*req.ActorUserID)
		if err != nil {
			return nil, fmt.Errorf("invalid actorUserId format (expected UUID): %w", err)
		}
		log.ActorUserID = &userUUID
	}
	if req.ActorUserType != nil {
		userType := strings.TrimSpace(*req.ActorUserType)
		log.ActorUserType = &userType
	}

	// Parse target fields
	if req.TargetServiceName != nil {
		serviceName := strings.TrimSpace(*req.TargetServiceName)
		log.TargetServiceName = &serviceName
	}
	if req.TargetResource != nil {
		resource := strings.TrimSpace(*req.TargetResource)
		log.TargetResource = &resource
	}
	if req.TargetResourceID != nil && *req.TargetResourceID != "" {
		resourceUUID, err := uuid.Parse(*req.TargetResourceID)
		if err != nil {
			return nil, fmt.Errorf("invalid targetResourceId format (expected UUID): %w", err)
		}
		log.TargetResourceID = &resourceUUID
	}

	return log, nil
}
