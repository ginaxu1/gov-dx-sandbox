package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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

// GetAuditEvents handles GET /audit/events (for Admin Portal)
func (h *AuditHandler) GetAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseFilterParams(r)

	// Get audit events
	response, err := h.auditService.GetAuditEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve audit events: %v", err), http.StatusInternalServerError)
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

// GetProviderAuditEvents handles GET /audit/providers (for Provider Portal)
func (h *AuditHandler) GetProviderAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseFilterParams(r)

	// Get provider_id from query parameter if provided
	providerID := r.URL.Query().Get("provider_id")
	if providerID != "" {
		filter.ProviderID = providerID
	}

	// Get provider audit events
	response, err := h.auditService.GetAuditEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve provider audit events: %v", err), http.StatusInternalServerError)
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

// GetConsumerAuditEvents handles GET /audit/consumers (for Consumer Portal)
func (h *AuditHandler) GetConsumerAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse filter parameters
	filter := h.parseFilterParams(r)

	// Get consumer_id from query parameter if provided
	consumerID := r.URL.Query().Get("consumer_id")
	if consumerID != "" {
		filter.ConsumerID = consumerID
	}

	// Get consumer audit events
	response, err := h.auditService.GetAuditEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve consumer audit events: %v", err), http.StatusInternalServerError)
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

// parseFilterParams parses filter parameters from query string
func (h *AuditHandler) parseFilterParams(r *http.Request) *models.AuditFilter {
	filter := &models.AuditFilter{}

	// Parse query parameters
	if consumerID := r.URL.Query().Get("consumer_id"); consumerID != "" {
		filter.ConsumerID = consumerID
	}

	if providerID := r.URL.Query().Get("provider_id"); providerID != "" {
		filter.ProviderID = providerID
	}

	if status := r.URL.Query().Get("transaction_status"); status != "" {
		filter.TransactionStatus = status
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
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

// CreateAuditLog handles POST /audit/logs (for api-server-go to send audit logs)
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ConsumerID == "" || req.ProviderID == "" || req.TransactionStatus == "" || len(req.RequestedData) == 0 || req.RequestPath == "" || req.RequestMethod == "" {
		http.Error(w, "Missing required fields: consumer_id, provider_id, transaction_status, requested_data, request_path, request_method", http.StatusBadRequest)
		return
	}

	// Validate transaction status
	if req.TransactionStatus != "SUCCESS" && req.TransactionStatus != "FAILURE" {
		http.Error(w, "Invalid transaction status. Must be SUCCESS or FAILURE", http.StatusBadRequest)
		return
	}

	// Generate event ID if not provided
	if req.EventID.String() == "00000000-0000-0000-0000-000000000000" {
		req.EventID = uuid.New()
	}

	// Auto-generate citizen hash from request/response data
	citizenHash := h.generateCitizenHash(req.RequestedData, req.ResponseData)

	// Create audit log directly in database
	query := `
		INSERT INTO audit_logs (
			event_id, timestamp, consumer_id, provider_id,
			requested_data, response_data, transaction_status, citizen_hash,
			request_path, request_method, user_agent, ip_address
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := h.auditService.DB().Exec(
		query,
		req.EventID,
		time.Now(),
		req.ConsumerID,
		req.ProviderID,
		req.RequestedData,
		req.ResponseData,
		req.TransactionStatus,
		citizenHash,
		req.RequestPath,
		req.RequestMethod,
		req.UserAgent,
		req.IPAddress,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create audit log: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the created audit log
	response := map[string]interface{}{
		"event_id": req.EventID,
		"status":   "created",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateAuditLogManual handles POST /audit/create (for manual testing)
func (h *AuditHandler) CreateAuditLogManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ConsumerID == "" || req.ProviderID == "" || req.TransactionStatus == "" || len(req.RequestedData) == 0 || req.RequestPath == "" || req.RequestMethod == "" {
		http.Error(w, "Missing required fields: consumer_id, provider_id, transaction_status, requested_data, request_path, request_method", http.StatusBadRequest)
		return
	}

	// Validate transaction status
	if req.TransactionStatus != "SUCCESS" && req.TransactionStatus != "FAILURE" {
		http.Error(w, "Invalid transaction status. Must be SUCCESS or FAILURE", http.StatusBadRequest)
		return
	}

	// Generate event ID if not provided
	if req.EventID.String() == "00000000-0000-0000-0000-000000000000" {
		req.EventID = uuid.New()
	}

	// Auto-generate citizen hash from request/response data
	citizenHash := h.generateCitizenHash(req.RequestedData, req.ResponseData)

	// Create audit log directly in database
	query := `
		INSERT INTO audit_logs (
			event_id, timestamp, consumer_id, provider_id,
			requested_data, response_data, transaction_status, citizen_hash,
			request_path, request_method, user_agent, ip_address
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := h.auditService.DB().Exec(
		query,
		req.EventID,
		time.Now(),
		req.ConsumerID,
		req.ProviderID,
		req.RequestedData,
		req.ResponseData,
		req.TransactionStatus,
		citizenHash,
		req.RequestPath,
		req.RequestMethod,
		req.UserAgent,
		req.IPAddress,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create audit log: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Return the created audit log with more details
	response := map[string]interface{}{
		"event_id":           req.EventID,
		"timestamp":          time.Now(),
		"consumer_id":        req.ConsumerID,
		"provider_id":        req.ProviderID,
		"transaction_status": req.TransactionStatus,
		"citizen_hash":       citizenHash,
		"request_path":       req.RequestPath,
		"request_method":     req.RequestMethod,
		"user_agent":         req.UserAgent,
		"ip_address":         req.IPAddress,
		"status":             "created",
		"message":            "Audit log created successfully for testing",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// generateCitizenHash extracts citizen IDs from request/response data and creates a hash
func (h *AuditHandler) generateCitizenHash(requestedData, responseData json.RawMessage) string {
	// Create a PII redaction service to extract citizen IDs
	piiService := services.NewPIIRedactionService()

	// Combine request and response data for citizen ID extraction
	var combinedData interface{}

	// Try to extract from response data first (more likely to contain citizen info)
	if len(responseData) > 0 {
		if err := json.Unmarshal(responseData, &combinedData); err == nil {
			if citizenID := piiService.ExtractCitizenID(combinedData); citizenID != "" {
				return piiService.HashCitizenID(citizenID)
			}
		}
	}

	// Fallback to request data
	if len(requestedData) > 0 {
		if err := json.Unmarshal(requestedData, &combinedData); err == nil {
			if citizenID := piiService.ExtractCitizenID(combinedData); citizenID != "" {
				return piiService.HashCitizenID(citizenID)
			}
		}
	}

	// If no citizen ID found, return a default hash
	return "no_citizen_id_found"
}
