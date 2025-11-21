package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/models"
	"github.com/gov-dx-sandbox/exchange/consent-engine/service"
	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// ConsentHandler holds dependencies for the HTTP handlers
type ConsentHandler struct {
	engine service.ConsentEngine
}

// NewConsentHandler creates a new ConsentHandler
func NewConsentHandler(engine service.ConsentEngine) *ConsentHandler {
	return &ConsentHandler{
		engine: engine,
	}
}

// Consent handlers - organized for better readability
func (h *ConsentHandler) HandleConsentPost(w http.ResponseWriter, r *http.Request) {
	// POST /consents should only create new consent records
	// The engine will handle reuse logic internally
	h.createConsent(w, r)
}

func (h *ConsentHandler) createConsent(w http.ResponseWriter, r *http.Request) {
	var req models.ConsentRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	// Log the raw request body for debugging
	slog.Info("POST /consents request body", "body", string(body))

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate that all required fields are present and not empty
	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if len(req.ConsentRequirements) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_requirements is required and cannot be empty"})
		return
	}

	// Validate each consent requirement
	for i, requirement := range req.ConsentRequirements {
		if requirement.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if len(requirement.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].fields is required and cannot be empty", i)})
			return
		}
		// Validate each field has fieldName and schemaId
		for j, field := range requirement.Fields {
			if field.FieldName == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].fields[%d].fieldName is required and cannot be empty", i, j)})
				return
			}
			if field.SchemaID == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].fields[%d].schemaId is required and cannot be empty", i, j)})
				return
			}
		}
	}

	// Process consent request using the engine
	record, err := h.engine.ProcessConsentRequest(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "create consent", "id", record.ConsentID, "owner", record.OwnerID)

	// Return simplified response format
	response := record.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (h *ConsentHandler) UpdateConsent(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := h.engine.UpdateConsent(id, req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(models.ErrConsentUpdateFailed+": %w", err)
		}

		// Return simplified response format
		response := record.ToConsentResponse()
		return response, http.StatusOK, nil
	})
}

func (h *ConsentHandler) RevokeConsent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Reason string }
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := h.engine.RevokeConsent(id, req.Reason)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(models.ErrConsentRevokeFailed+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

// revokeConsentByID handles DELETE /consents/{id} - revoke consent by ID
func (h *ConsentHandler) revokeConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct{ Reason string }

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	record, err := h.engine.RevokeConsent(consentID, req.Reason)
	if err != nil {
		handleConsentError(w, err, "revoke consent")
		return
	}

	// Return simplified response format
	response := record.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// patchConsentByID handles PATCH /consents/{id} - partial update of consent resource
func (h *ConsentHandler) patchConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status        string   `json:"status,omitempty"`
		UpdatedBy     string   `json:"updated_by,omitempty"`
		Reason        string   `json:"reason,omitempty"`
		GrantDuration string   `json:"grant_duration,omitempty"`
		Fields        []string `json:"fields,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Get the existing record first
	existingRecord, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		handleConsentError(w, err, "get consent record")
		return
	}

	// Apply partial updates
	updateReq := models.UpdateConsentRequest{
		Status:    models.ConsentStatus(existingRecord.Status), // Keep existing status by default
		UpdatedBy: existingRecord.OwnerID,                      // Keep existing updated_by by default
		Reason:    "",                                          // Will be set if provided
	}

	// Update only provided fields
	if req.Status != "" {
		updateReq.Status = models.ConsentStatus(req.Status)
	}
	if req.UpdatedBy != "" {
		updateReq.UpdatedBy = req.UpdatedBy
	}
	if req.Reason != "" {
		updateReq.Reason = req.Reason
	}
	if req.GrantDuration != "" {
		updateReq.GrantDuration = req.GrantDuration
	}
	if len(req.Fields) > 0 {
		updateReq.Fields = req.Fields
	}

	// Update the record
	updatedRecord, err := h.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		handleConsentError(w, err, "update consent record")
		return
	}
	slog.Info("Consent record updated", "consent_id", updatedRecord.ConsentID, "owner_id", updatedRecord.OwnerID, "owner_email", updatedRecord.OwnerEmail, "app_id", updatedRecord.AppID, "status", updatedRecord.Status, "type", updatedRecord.Type, "created_at", updatedRecord.CreatedAt, "updated_at", updatedRecord.UpdatedAt, "expires_at", updatedRecord.ExpiresAt, "grant_duration", updatedRecord.GrantDuration, "fields", updatedRecord.Fields, "session_id", updatedRecord.SessionID, "consent_portal_url", updatedRecord.ConsentPortalURL)

	// Return simplified response format
	response := updatedRecord.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *ConsentHandler) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := h.engine.GetConsentsByDataOwner(dataOwner)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(models.ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"owner_id": dataOwner,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (h *ConsentHandler) getConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consumer/", func(consumer string) (interface{}, int, error) {
		records, err := h.engine.GetConsentsByConsumer(consumer)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(models.ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"consumer": consumer,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (h *ConsentHandler) checkConsentExpiry(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		expiredRecords, err := h.engine.CheckConsentExpiry()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(models.ErrConsentExpiryFailed+": %w", err)
		}

		// Log the operation
		slog.Info("Operation successful",
			"operation", models.OpCheckConsentExpiry,
			"expired_count", len(expiredRecords),
		)

		// Ensure expired_records is always an array, never null
		expiredRecordsList := make([]*models.ConsentRecord, 0)
		if expiredRecords != nil {
			expiredRecordsList = expiredRecords
		}

		return map[string]interface{}{
			"expired_records": expiredRecordsList,
			"count":           len(expiredRecordsList),
			"checked_at":      time.Now(),
		}, http.StatusOK, nil
	})
}

// Route handlers - organized for better readability
func (h *ConsentHandler) ConsentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	switch {
	case path == "" && r.Method == http.MethodPost:
		// POST /consents - create new consent record
		h.HandleConsentPost(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// ConsentHandlerWithID handles operations on /consents/{id} with different auth requirements
func (h *ConsentHandler) ConsentHandlerWithID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	switch {
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		// GET /consents/{id} - get consent by ID (requires auth)
		consentID := strings.TrimPrefix(path, "/")
		h.getConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		// PUT /consents/{id} - replace entire consent resource (requires auth)
		consentID := strings.TrimPrefix(path, "/")
		h.updateConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPatch:
		// PATCH /consents/{id} - partial update of consent resource (no auth required)
		consentID := strings.TrimPrefix(path, "/")
		h.patchConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		// DELETE /consents/{id} - revoke consent (no auth required)
		consentID := strings.TrimPrefix(path, "/")
		h.revokeConsentByID(w, r, consentID)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) DataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getConsentsByDataOwner(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) ConsumerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getConsentsByConsumer(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) getConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Convert to the user-facing ConsentPortalView
	consentView := record.ToConsentPortalView()

	// Return only the UI-necessary fields
	utils.RespondWithJSON(w, http.StatusOK, consentView)
}

func (h *ConsentHandler) getDataInfo(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Return only owner_id and owner_email
	dataInfo := map[string]interface{}{
		"owner_id":    record.OwnerID,
		"owner_email": record.OwnerEmail,
	}

	utils.RespondWithJSON(w, http.StatusOK, dataInfo)
}

func (h *ConsentHandler) updateConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status        string `json:"status"`
		UpdatedBy     string `json:"updated_by,omitempty"`
		GrantDuration string `json:"grant_duration,omitempty"`
		Reason        string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Get the existing consent record to extract owner information
	existingRecord, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		handleConsentError(w, err, "get consent record")
		return
	}

	// Validate status if provided
	var newStatus models.ConsentStatus
	if req.Status != "" {
		if !validateConsentStatus(w, req.Status) {
			return
		}
		newStatus = models.ConsentStatus(req.Status)
	} else {
		// Keep existing status if not provided
		newStatus = models.ConsentStatus(existingRecord.Status)
	}

	// Set default reason if not provided
	reason := req.Reason
	if reason == "" {
		reason = getDefaultReason(newStatus)
	}

	// Update the record
	updateReq := models.UpdateConsentRequest{
		Status:        newStatus,
		UpdatedBy:     existingRecord.OwnerID, // Use existing owner ID
		Reason:        reason,
		GrantDuration: req.GrantDuration, // Will be empty string if not provided
	}

	updatedRecord, err := h.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		handleConsentError(w, err, "update consent record")
		return
	}

	// Log the operation
	slog.Info("Consent updated via PUT", "consentId", consentID, "status", string(newStatus), "ownerId", existingRecord.OwnerID, "grantDuration", req.GrantDuration)

	// Return simplified response format
	response := updatedRecord.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *ConsentHandler) DataInfoHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/data-info/")
	if path == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	if r.Method == http.MethodGet {
		h.getDataInfo(w, r, path)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		h.checkConsentExpiry(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// HandlePortalAction handles POST /consents/portal/actions
func (h *ConsentHandler) HandlePortalAction(w http.ResponseWriter, r *http.Request) {
	var req models.ConsentPortalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	record, err := h.engine.ProcessConsentPortalRequest(req)
	if err != nil {
		handleConsentError(w, err, "process portal request")
		return
	}

	// Return simplified response format
	response := record.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}
