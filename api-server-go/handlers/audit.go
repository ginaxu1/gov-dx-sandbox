package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
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

// CreateAuditLog handles POST /audit/logs
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	var req models.AuditLogRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if req.ConsumerID == "" || req.ProviderID == "" || req.TransactionStatus == "" || req.CitizenHash == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Validate transaction status
	if req.TransactionStatus != models.TransactionStatusSuccess && req.TransactionStatus != models.TransactionStatusFailure {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid transaction status")
		return
	}

	auditLog, err := h.auditService.CreateAuditLog(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create audit log")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, auditLog)
}

// GetAuditLogsForProvider handles GET /audit/provider/{providerId}
func (h *AuditHandler) GetAuditLogsForProvider(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider_id")
	if providerID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Provider ID is required")
		return
	}

	limit, offset := h.parsePaginationParams(r)

	logs, err := h.auditService.GetAuditLogsSummaryForProvider(providerID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Return the simplified array format as requested
	utils.RespondWithSuccess(w, http.StatusOK, logs)
}

// GetAuditLogsForAdmin handles GET /audit/admin
func (h *AuditHandler) GetAuditLogsForAdmin(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.parsePaginationParams(r)

	// Parse filters from query parameters
	filter := h.parseFilterParams(r)

	var logs []*models.AuditLogSummaryResponse
	var err error

	if filter != nil && h.hasFilters(filter) {
		// For filtered queries, we need to get detailed logs first, then convert
		detailedLogs, err := h.auditService.GetAuditLogsWithFilter(filter)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
			return
		}
		logs = h.auditService.ConvertToSummaryResponse(detailedLogs)
	} else {
		logs, err = h.auditService.GetAuditLogsSummaryForAdmin(limit, offset)
	}

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Return the simplified array format as requested
	utils.RespondWithSuccess(w, http.StatusOK, logs)
}

// GetAuditLogsForCitizen handles GET /audit/citizen
func (h *AuditHandler) GetAuditLogsForCitizen(w http.ResponseWriter, r *http.Request) {
	citizenID := r.URL.Query().Get("citizen_id")
	if citizenID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Citizen ID is required")
		return
	}

	// Hash the citizen ID for lookup
	piiService := services.NewPIIRedactionService()
	citizenHash := piiService.HashCitizenID(citizenID)

	limit, offset := h.parsePaginationParams(r)

	logs, err := h.auditService.GetAuditLogsSummaryForCitizen(citizenHash, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Return the simplified array format as requested
	utils.RespondWithSuccess(w, http.StatusOK, logs)
}

// parsePaginationParams parses limit and offset from query parameters
func (h *AuditHandler) parsePaginationParams(r *http.Request) (int, int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

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

	return limit, offset
}

// parseFilterParams parses filter parameters from query string
func (h *AuditHandler) parseFilterParams(r *http.Request) *models.AuditLogFilter {
	filter := &models.AuditLogFilter{}

	if consumerID := r.URL.Query().Get("consumer_id"); consumerID != "" {
		filter.ConsumerID = consumerID
	}

	if providerID := r.URL.Query().Get("provider_id"); providerID != "" {
		filter.ProviderID = providerID
	}

	if citizenHash := r.URL.Query().Get("citizen_hash"); citizenHash != "" {
		filter.CitizenHash = citizenHash
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

	limit, offset := h.parsePaginationParams(r)
	filter.Limit = limit
	filter.Offset = offset

	return filter
}

// hasFilters checks if any filters are set
func (h *AuditHandler) hasFilters(filter *models.AuditLogFilter) bool {
	return filter.ConsumerID != "" ||
		filter.ProviderID != "" ||
		filter.CitizenHash != "" ||
		filter.TransactionStatus != "" ||
		!filter.StartDate.IsZero() ||
		!filter.EndDate.IsZero()
}
