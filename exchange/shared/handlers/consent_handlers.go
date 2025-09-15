package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// ConsentEngine interface for dependency injection - using interface{} to avoid type conflicts
type ConsentEngine interface {
	CreateConsent(req interface{}) (interface{}, error)
	GetConsentStatus(id string) (interface{}, error)
	UpdateConsent(id string, req interface{}) (interface{}, error)
	RevokeConsent(id string, reason string) (interface{}, error)
	ProcessConsentPortalRequest(req interface{}) (interface{}, error)
	GetConsentsByDataOwner(dataOwner string) ([]interface{}, error)
	GetConsentsByConsumer(consumer string) ([]interface{}, error)
	CheckConsentExpiry() ([]interface{}, error)
}

// ConsentHandler holds dependencies for consent-related handlers
type ConsentHandler struct {
	engine ConsentEngine
}

// NewConsentHandler creates a new consent handler
func NewConsentHandler(engine ConsentEngine) *ConsentHandler {
	return &ConsentHandler{engine: engine}
}

// Create consent handler
func (h *ConsentHandler) CreateConsent(w http.ResponseWriter, r *http.Request) {
	var req CreateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := h.engine.CreateConsent(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentCreateFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusCreated, constants.OpCreateConsent, map[string]interface{}{
			"id": record.ID, "owner": record.DataOwner,
		})
		return record, http.StatusCreated, nil
	})
}

func (h *ConsentHandler) GetConsentStatus(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consents/", func(id string) (interface{}, int, error) {
		record, err := h.engine.GetConsentStatus(id)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(constants.ErrConsentNotFound+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

func (h *ConsentHandler) UpdateConsent(w http.ResponseWriter, r *http.Request) {
	var req UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := h.engine.UpdateConsent(id, req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentUpdateFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusOK, constants.OpUpdateConsent, map[string]interface{}{
			"id": record.ID, "status": record.Status,
		})
		return record, http.StatusOK, nil
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
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentRevokeFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusOK, constants.OpRevokeConsent, map[string]interface{}{
			"id": record.ID, "reason": req.Reason,
		})
		return record, http.StatusOK, nil
	})
}

// Portal and admin handlers
func (h *ConsentHandler) ProcessConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentPortalRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := h.engine.ProcessConsentPortalRequest(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrPortalRequestFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusOK, constants.OpProcessPortalRequest, map[string]interface{}{
			"id": record.ID, "action": req.Action, "status": record.Status,
		})
		return record, http.StatusOK, nil
	})
}

func (h *ConsentHandler) GetConsentPortalInfo(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		consentID, err := utils.ExtractQueryParam(r, "consent_id")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := h.engine.GetConsentStatus(consentID)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(constants.ErrConsentNotFound+": %w", err)
		}

		return map[string]interface{}{
			"consent_id":         record.ID,
			"status":             record.Status,
			"data_consumer":      record.DataConsumer,
			"data_owner":         record.DataOwner,
			"fields":             record.Fields,
			"consent_portal_url": record.ConsentPortalURL,
			"expires_at":         record.ExpiresAt,
			"created_at":         record.CreatedAt,
		}, http.StatusOK, nil
	})
}

func (h *ConsentHandler) GetConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := h.engine.GetConsentsByDataOwner(dataOwner)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"data_owner": dataOwner,
			"consents":   records,
			"count":      len(records),
		}, http.StatusOK, nil
	})
}

func (h *ConsentHandler) GetConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consumer/", func(consumer string) (interface{}, int, error) {
		records, err := h.engine.GetConsentsByConsumer(consumer)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"consumer": consumer,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (h *ConsentHandler) CheckConsentExpiry(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		expiredRecords, err := h.engine.CheckConsentExpiry()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentExpiryFailed+": %w", err)
		}
		utils.HandleSuccess(w, map[string]interface{}{
			"expired_records": expiredRecords,
			"count":           len(expiredRecords),
			"checked_at":      time.Now(),
		}, http.StatusOK, constants.OpCheckConsentExpiry, map[string]interface{}{
			"expired_count": len(expiredRecords),
		})
		return map[string]interface{}{
			"expired_records": expiredRecords,
			"count":           len(expiredRecords),
			"checked_at":      time.Now(),
		}, http.StatusOK, nil
	})
}

// Route handlers
func (h *ConsentHandler) ConsentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	switch {
	case path == "" && r.Method == http.MethodPost:
		h.CreateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		h.GetConsentStatus(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		h.UpdateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		h.RevokeConsent(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) ConsentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.ProcessConsentPortalRequest(w, r)
	case http.MethodGet:
		h.GetConsentPortalInfo(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) DataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.GetConsentsByDataOwner(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) ConsumerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.GetConsentsByConsumer(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (h *ConsentHandler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		h.CheckConsentExpiry(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}
