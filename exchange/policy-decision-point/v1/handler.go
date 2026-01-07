package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/middleware"
	auditclient "github.com/gov-dx-sandbox/shared/audit"
	pdpmodels "github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/services"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
	"gorm.io/gorm"
)

// Handler handles all API requests
type Handler struct {
	policyService *services.PolicyMetadataService
}

// NewHandler creates a new API handler
func NewHandler(db *gorm.DB) *Handler {
	policyService := services.NewPolicyMetadataService(db)
	return &Handler{
		policyService: policyService,
	}
}

// SetupRoutes configures all API routes
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/policy/", utils.PanicRecoveryMiddleware(http.HandlerFunc(h.handlePolicyService)))
}

// handlePolicyService handles policy metadata service requests
func (h *Handler) handlePolicyService(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/policy")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 1 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	switch parts[0] {
	case "metadata":
		switch r.Method {
		case http.MethodPost:
			h.CreatePolicyMetadata(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	case "update-allowlist":
		switch r.Method {
		case http.MethodPost:
			h.UpdateAllowList(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	case "decide":
		switch r.Method {
		case http.MethodPost:
			h.GetPolicyDecision(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// CreatePolicyMetadata handles creating policy metadata
func (h *Handler) CreatePolicyMetadata(w http.ResponseWriter, r *http.Request) {
	var req pdpmodels.PolicyMetadataCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.policyService.CreatePolicyMetadata(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, resp)
}

// UpdateAllowList handles updating the allow list for a policy
func (h *Handler) UpdateAllowList(w http.ResponseWriter, r *http.Request) {
	var req pdpmodels.AllowListUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.policyService.UpdateAllowList(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, resp)
}

// GetPolicyDecision handles getting a policy decision
func (h *Handler) GetPolicyDecision(w http.ResponseWriter, r *http.Request) {
	// Extract trace ID from request and add to context
	ctx := middleware.ExtractTraceIDFromRequest(r)

	var req pdpmodels.PolicyDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.policyService.GetPolicyDecision(&req)

	// Log policy check result
	h.logPolicyCheck(ctx, &req, resp, err)

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, resp)
}

// logPolicyCheck logs a POLICY_CHECK event from policy-decision-point's perspective
func (h *Handler) logPolicyCheck(ctx context.Context, req *pdpmodels.PolicyDecisionRequest, resp *pdpmodels.PolicyDecisionResponse, err error) {
	traceID := middleware.GetTraceIDFromContext(ctx)
	if traceID == "" {
		return
	}

	eventType := "POLICY_CHECK"
	actorType := "SERVICE"
	actorID := "policy-decision-point"
	targetType := "SERVICE"
	targetID := "orchestration-engine"

	status := auditclient.StatusSuccess
	responseMetadata := make(map[string]interface{})

	if err != nil {
		status = auditclient.StatusFailure
		responseMetadata["error"] = err.Error()
	} else if resp != nil {
		responseMetadata["authorized"] = resp.AppAuthorized
		responseMetadata["consentRequired"] = resp.AppRequiresOwnerConsent
		responseMetadata["accessExpired"] = resp.AppAccessExpired
		if !resp.AppAuthorized {
			status = auditclient.StatusFailure
			responseMetadata["unauthorizedFields"] = resp.UnauthorizedFields
		}
		if resp.AppAccessExpired {
			status = auditclient.StatusFailure
			responseMetadata["expiredFields"] = resp.ExpiredFields
		}
		if resp.AppRequiresOwnerConsent {
			responseMetadata["consentRequiredFields"] = resp.ConsentRequiredFields
		}
	}

	var responseMetadataJSON []byte
	responseMetadataBytes, jsonErr := json.Marshal(responseMetadata)
	if jsonErr != nil {
		responseMetadataJSON = []byte("{}")
	} else {
		responseMetadataJSON = responseMetadataBytes
	}

	auditRequest := &auditclient.AuditLogRequest{
		TraceID:          &traceID,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		EventType:        &eventType,
		Status:           status,
		ActorType:        actorType,
		ActorID:          actorID,
		TargetType:       targetType,
		TargetID:         &targetID,
		ResponseMetadata: json.RawMessage(responseMetadataJSON),
	}

	middleware.LogGeneralizedAuditEvent(ctx, auditRequest)
}
