package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/middleware"
	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
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
	var req models.PolicyMetadataCreateRequest

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
	var req models.AllowListUpdateRequest
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
	ctx := r.Context()
	traceID := middleware.GetTraceIDFromContext(ctx)

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		errorMetadata, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		timestamp := time.Now().UTC().Format(time.RFC3339)
		actorServiceName := "policy-decision-point"
		middleware.LogGeneralizedAuditEvent(ctx, &middleware.CreateAuditLogRequest{
			TraceID:          &traceID,
			Timestamp:        &timestamp,
			EventName:        "POLICY_CHECK",
			EventType:        stringPtr("READ"),
			Status:           "FAILURE",
			ActorType:        "SERVICE",
			ActorServiceName: &actorServiceName,
			TargetType:       "SERVICE",
			ResponseMetadata: errorMetadata,
		})
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Log policy decision request
	requestData, _ := json.Marshal(map[string]interface{}{
		"requestBody": json.RawMessage(bodyBytes),
	})
	timestamp := time.Now().UTC().Format(time.RFC3339)
	actorServiceName := "policy-decision-point"
	middleware.LogGeneralizedAuditEvent(ctx, &middleware.CreateAuditLogRequest{
		TraceID:          &traceID,
		Timestamp:        &timestamp,
		EventName:        "POLICY_CHECK",
		EventType:        stringPtr("READ"),
		Status:           "SUCCESS",
		ActorType:        "SERVICE",
		ActorServiceName: &actorServiceName,
		TargetType:       "SERVICE",
		RequestedData:    requestData,
	})

	var req models.PolicyDecisionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		// Log failure
		errorMetadata, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		timestamp := time.Now().UTC().Format(time.RFC3339)
		actorServiceName := "policy-decision-point"
		middleware.LogGeneralizedAuditEvent(ctx, &middleware.CreateAuditLogRequest{
			TraceID:          &traceID,
			Timestamp:        &timestamp,
			EventName:        "POLICY_CHECK",
			EventType:        stringPtr("READ"),
			Status:           "FAILURE",
			ActorType:        "SERVICE",
			ActorServiceName: &actorServiceName,
			TargetType:       "SERVICE",
			ResponseMetadata: errorMetadata,
		})
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.policyService.GetPolicyDecision(&req)
	if err != nil {
		// Log failure
		errorMetadata, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
		})
		timestamp := time.Now().UTC().Format(time.RFC3339)
		actorServiceName := "policy-decision-point"
		middleware.LogGeneralizedAuditEvent(ctx, &middleware.CreateAuditLogRequest{
			TraceID:          &traceID,
			Timestamp:        &timestamp,
			EventName:        "POLICY_CHECK",
			EventType:        stringPtr("READ"),
			Status:           "FAILURE",
			ActorType:        "SERVICE",
			ActorServiceName: &actorServiceName,
			TargetType:       "SERVICE",
			ResponseMetadata: errorMetadata,
		})
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Log successful policy decision response
	responseMetadata, _ := json.Marshal(map[string]interface{}{
		"applicationId":              req.ApplicationID,
		"appAuthorized":              resp.AppAuthorized,
		"appAccessExpired":           resp.AppAccessExpired,
		"appRequiresOwnerConsent":    resp.AppRequiresOwnerConsent,
		"unauthorizedFieldsCount":    len(resp.UnauthorizedFields),
		"expiredFieldsCount":         len(resp.ExpiredFields),
		"consentRequiredFieldsCount": len(resp.ConsentRequiredFields),
	})
	timestamp = time.Now().UTC().Format(time.RFC3339)
	actorServiceName = "policy-decision-point"
	middleware.LogGeneralizedAuditEvent(ctx, &middleware.CreateAuditLogRequest{
		TraceID:          &traceID,
		Timestamp:        &timestamp,
		EventName:        "POLICY_CHECK",
		EventType:        stringPtr("READ"),
		Status:           "SUCCESS",
		ActorType:        "SERVICE",
		ActorServiceName: &actorServiceName,
		TargetType:       "SERVICE",
		ResponseMetadata: responseMetadata,
	})

	utils.RespondWithSuccess(w, http.StatusOK, resp)
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
