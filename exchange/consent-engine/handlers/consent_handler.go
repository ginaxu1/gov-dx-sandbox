package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	service "github.com/gov-dx-sandbox/exchange/consent-engine/v1/services"
	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// ConsentHandler groups HTTP handlers that operate on consent engine resources.
type ConsentHandler struct {
	engine service.ConsentEngine
}

// NewConsentHandler returns a handler set wired with the provided consent engine.
func NewConsentHandler(engine service.ConsentEngine) *ConsentHandler {
	return &ConsentHandler{
		engine: engine,
	}
}

// HandleConsents exposes operations for /v1/consents without an ID.
func (h *ConsentHandler) HandleConsents(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/consents")
	switch {
	case path == "" && r.Method == http.MethodPost:
		h.handleConsentPost(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// HandleConsentWithID exposes operations for /v1/consents/{id}.
func (h *ConsentHandler) HandleConsentWithID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/consents")
	switch {
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		consentID := strings.TrimPrefix(path, "/")
		h.getConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		consentID := strings.TrimPrefix(path, "/")
		h.updateConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPatch:
		consentID := strings.TrimPrefix(path, "/")
		h.patchConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		consentID := strings.TrimPrefix(path, "/")
		h.revokeConsentByID(w, r, consentID)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// HandleDataOwner exposes /v1/data-owner/{ownerId}.
func (h *ConsentHandler) HandleDataOwner(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getConsentsByDataOwner(w, r)
		return
	}
	utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
}

// HandleConsumer exposes /v1/consumer/{consumerId}.
func (h *ConsentHandler) HandleConsumer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.getConsentsByConsumer(w, r)
		return
	}
	utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
}

// HandleDataInfo exposes /v1/data-info/{consentId}.
func (h *ConsentHandler) HandleDataInfo(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/data-info/")
	if path == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	if r.Method == http.MethodGet {
		h.getDataInfo(w, r, path)
		return
	}

	utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
}

// HandleAdmin exposes /v1/admin endpoints.
func (h *ConsentHandler) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		h.checkConsentExpiry(w, r)
		return
	}
	utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
}

func (h *ConsentHandler) handleConsentPost(w http.ResponseWriter, r *http.Request) {
	h.createConsent(w, r)
}

func (h *ConsentHandler) createConsent(w http.ResponseWriter, r *http.Request) {
	var req models.ConsentRequest

	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	slog.Info("POST /consents request body", "body", string(body))

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if len(req.ConsentRequirements) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_requirements is required and cannot be empty"})
		return
	}

	for i, requirement := range req.ConsentRequirements {
		if requirement.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if len(requirement.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("consent_requirements[%d].fields is required and cannot be empty", i)})
			return
		}
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

	record, err := h.engine.ProcessConsentRequest(req)
	if err != nil {
		handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to process consent request: "+err.Error())
		return
	}

	slog.Info("Operation successful", "operation", "create consent", "id", record.ConsentID, "owner", record.OwnerID)

	response := record.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (h *ConsentHandler) updateConsent(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/v1/consents/")
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

		response := record.ToConsentResponse()
		return response, http.StatusOK, nil
	})
}

func (h *ConsentHandler) revokeConsent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Reason string }
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/v1/consents/")
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

func (h *ConsentHandler) revokeConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct{ Reason string }

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	record, err := h.engine.RevokeConsent(consentID, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to revoke consent: "+err.Error())
		}
		return
	}

	response := record.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}

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

	existingRecord, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to get consent record: "+err.Error())
		}
		return
	}

	updateReq := models.UpdateConsentRequest{
		Status:    models.ConsentStatus(existingRecord.Status),
		UpdatedBy: existingRecord.OwnerID,
		Reason:    "",
	}

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

	updatedRecord, err := h.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to update consent record: "+err.Error())
		}
		return
	}
	slog.Info("Consent record updated", "consent_id", updatedRecord.ConsentID, "owner_id", updatedRecord.OwnerID, "owner_email", updatedRecord.OwnerEmail, "app_id", updatedRecord.AppID, "status", updatedRecord.Status, "type", updatedRecord.Type, "created_at", updatedRecord.CreatedAt, "updated_at", updatedRecord.UpdatedAt, "expires_at", updatedRecord.ExpiresAt, "grant_duration", updatedRecord.GrantDuration, "fields", updatedRecord.Fields, "session_id", updatedRecord.SessionID, "consent_portal_url", updatedRecord.ConsentPortalURL)

	response := updatedRecord.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *ConsentHandler) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/v1/data-owner/", func(dataOwner string) (interface{}, int, error) {
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
	utils.PathHandler(w, r, "/v1/consumer/", func(consumer string) (interface{}, int, error) {
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

		slog.Info("Operation successful",
			"operation", models.OpCheckConsentExpiry,
			"expired_count", len(expiredRecords),
		)

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

func (h *ConsentHandler) getConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	consentView := record.ToConsentPortalView()
	utils.RespondWithJSON(w, http.StatusOK, consentView)
}

func (h *ConsentHandler) getDataInfo(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

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

	existingRecord, err := h.engine.GetConsentStatus(consentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to get consent record: "+err.Error())
		}
		return
	}

	newStatus, err := validateConsentStatus(req.Status)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}
	if newStatus == "" {
		newStatus = models.ConsentStatus(existingRecord.Status)
	}

	reason := req.Reason
	if reason == "" {
		reason = getDefaultReason(newStatus)
	}

	updateReq := models.UpdateConsentRequest{
		Status:        newStatus,
		UpdatedBy:     existingRecord.OwnerID,
		Reason:        reason,
		GrantDuration: req.GrantDuration,
	}

	updatedRecord, err := h.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			handleConsentError(w, http.StatusInternalServerError, models.ErrorCodeInternalError, "Failed to update consent record: "+err.Error())
		}
		return
	}

	slog.Info("Consent updated via PUT", "consentId", consentID, "status", string(newStatus), "ownerId", existingRecord.OwnerID, "grantDuration", req.GrantDuration)

	response := updatedRecord.ToConsentResponse()
	utils.RespondWithJSON(w, http.StatusOK, response)
}
