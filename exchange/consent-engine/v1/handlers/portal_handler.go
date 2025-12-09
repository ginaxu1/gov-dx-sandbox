package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/services"
)

// PortalHandler handles external API requests (authentication required)
type PortalHandler struct {
	consentService *services.ConsentService
}

// NewPortalHandler creates a new portal handler
func NewPortalHandler(consentService *services.ConsentService) *PortalHandler {
	return &PortalHandler{
		consentService: consentService,
	}
}

// HealthCheck handles GET /api/v1/health
func (h *PortalHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, models.ErrorCodeMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]string{
		"status": "healthy",
	}
	respondWithJSON(w, http.StatusOK, response)
}

// GetConsent handles GET /api/v1/consents/:consentId
// Authorization: Bearer Token
// Verifies that consent.owner_email matches the email from the decoded token
func (h *PortalHandler) GetConsent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, models.ErrorCodeMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract consentId from URL path parameter
	consentID := r.PathValue("consentId")
	if consentID == "" {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "consentId is required")
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(consentID); err != nil {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "invalid consentId format")
		return
	}

	// TODO: Extract email from bearer token (to be implemented later)
	// For now, we'll get the email from context (will be set by auth middleware)
	userEmail := r.Context().Value("userEmail")
	if userEmail == nil {
		respondWithError(w, http.StatusUnauthorized, models.ErrorCodeUnauthorized, "User email not found in token")
		return
	}

	// Get consent from service (context with timeout is propagated)
	consent, err := h.consentService.GetConsentPortalView(r.Context(), consentID)
	if err != nil {
		// Check if error is due to context cancellation or timeout
		if r.Context().Err() != nil {
			slog.Warn("Request context cancelled during service call", "error", r.Context().Err())
			respondWithError(w, http.StatusRequestTimeout, models.ErrorCodeInternalError, "Request timeout or cancelled")
			return
		}
		if containsError(err, string(models.ErrConsentNotFound)) {
			respondWithError(w, http.StatusNotFound, models.ErrorCodeConsentNotFound, "Consent not found")
			return
		}
		slog.Error("Failed to get consent", "error", err)
		respondWithError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "An unexpected error occurred")
		return
	}

	// Verify that the consent owner email matches the authenticated user email
	if consent.OwnerEmail != userEmail.(string) {
		respondWithError(w, http.StatusForbidden, models.ErrorCodeForbidden, "Access denied: consent belongs to a different user")
		return
	}

	respondWithJSON(w, http.StatusOK, consent)
}

// UpdateConsent handles PUT /api/v1/consents/:consentId
// Authorization: Bearer Token
// Verifies that consent.owner_email matches the email from the decoded token
// Body: { "action": "approve" | "reject" }
func (h *PortalHandler) UpdateConsent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, models.ErrorCodeMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract consentId from URL path parameter
	consentID := r.PathValue("consentId")
	if consentID == "" {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "consentId is required")
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(consentID); err != nil {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "invalid consentId format")
		return
	}

	// TODO: Extract email from bearer token (to be implemented later)
	// For now, we'll get the email from context (will be set by auth middleware)
	userEmail := r.Context().Value("userEmail")
	if userEmail == nil {
		respondWithError(w, http.StatusUnauthorized, models.ErrorCodeUnauthorized, "User email not found in token")
		return
	}

	// Parse request body
	var actionReq struct {
		Action string `json:"action"`
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&actionReq); err != nil {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate action
	if actionReq.Action != string(models.ActionApprove) && actionReq.Action != string(models.ActionReject) {
		respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, fmt.Sprintf("Invalid action: %s. Must be 'approve' or 'reject'", actionReq.Action))
		return
	}

	// First, get the consent to verify ownership (context with timeout is propagated)
	consent, err := h.consentService.GetConsentPortalView(r.Context(), consentID)
	if err != nil {
		// Check if error is due to context cancellation or timeout
		if r.Context().Err() != nil {
			slog.Warn("Request context cancelled during service call", "error", r.Context().Err())
			respondWithError(w, http.StatusRequestTimeout, models.ErrorCodeInternalError, "Request timeout or cancelled")
			return
		}
		if containsError(err, string(models.ErrConsentNotFound)) {
			respondWithError(w, http.StatusNotFound, models.ErrorCodeConsentNotFound, "Consent not found")
			return
		}
		slog.Error("Failed to get consent", "error", err)
		respondWithError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "An unexpected error occurred")
		return
	}

	// Verify that the consent owner email matches the authenticated user email
	if consent.OwnerEmail != userEmail.(string) {
		respondWithError(w, http.StatusForbidden, models.ErrorCodeForbidden, "Access denied: consent belongs to a different user")
		return
	}

	// Update consent status
	updateReq := models.ConsentPortalActionRequest{
		ConsentID: consentID,
		Action:    models.ConsentPortalAction(actionReq.Action),
		UpdatedBy: userEmail.(string),
	}

	if err := h.consentService.UpdateConsentStatusByPortalAction(r.Context(), updateReq); err != nil {
		// Check if error is due to context cancellation or timeout
		if r.Context().Err() != nil {
			slog.Warn("Request context cancelled during update operation", "error", r.Context().Err())
			respondWithError(w, http.StatusRequestTimeout, models.ErrorCodeInternalError, "Request timeout or cancelled")
			return
		}
		if containsError(err, string(models.ErrConsentNotFound)) {
			respondWithError(w, http.StatusNotFound, models.ErrorCodeConsentNotFound, "Consent not found")
			return
		}
		if containsError(err, string(models.ErrPortalRequestFailed)) {
			respondWithError(w, http.StatusBadRequest, models.ErrorCodeBadRequest, "Invalid consent update request")
			return
		}
		slog.Error("Failed to update consent", "error", err)
		respondWithError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "An unexpected error occurred")
		return
	}

	// Return success response with actual consent status
	statusMap := map[string]string{
		string(models.ActionApprove): string(models.StatusApproved),
		string(models.ActionReject):  string(models.StatusRejected),
	}
	response := map[string]string{
		"message": "Consent updated successfully",
		"status":  statusMap[actionReq.Action],
	}
	respondWithJSON(w, http.StatusOK, response)
}
