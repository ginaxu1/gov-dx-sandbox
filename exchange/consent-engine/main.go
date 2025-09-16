package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/config"
	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify your frontend domain
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// apiServer holds dependencies for the HTTP handlers
type apiServer struct {
	engine ConsentEngine
}

// ConsentPortalCreateRequest represents the request format for creating consent via portal
type ConsentPortalCreateRequest struct {
	AppID            string      `json:"app_id"`
	DataFields       []DataField `json:"data_fields"`
	Purpose          string      `json:"purpose"`
	SessionID        string      `json:"session_id"`
	ConsentPortalURL string      `json:"consent_portal_url"`
}

// ConsentPortalUpdateRequest represents the request format for updating consent via portal
type ConsentPortalUpdateRequest struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updated_by"`
	Reason    string `json:"reason,omitempty"`
}

// Consent handlers - organized for better readability
func (s *apiServer) handleConsentPost(w http.ResponseWriter, r *http.Request) {
	// First, try to parse as a consent update (has consent_id field)
	var updateReq struct {
		ConsentID string `json:"consent_id"`
		Status    string `json:"status"`
	}

	// Read the request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	// Try to parse as update request
	if err := json.Unmarshal(body, &updateReq); err == nil && updateReq.ConsentID != "" {
		// This is a consent update request
		s.updateConsentStatus(w, r, updateReq)
		return
	}

	// Otherwise, treat as new consent request - reset the body for createConsent to read
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	s.createConsent(w, r)
}

func (s *apiServer) processConsentRequest(w http.ResponseWriter, r *http.Request, req ConsentRequest) {
	response, err := s.engine.ProcessConsentRequest(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", OpCreateConsent, "consentId", response.ConsentID, "status", response.Status)

	// Return the ConsentRecord directly as it already has the correct format
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	var req ConsentRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate that all required fields are present and not empty
	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if req.Purpose == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "purpose is required and cannot be empty"})
		return
	}
	if req.SessionID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "session_id is required and cannot be empty"})
		return
	}
	if req.ConsentPortalURL == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_portal_url is required and cannot be empty"})
		return
	}
	if len(req.DataFields) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "data_fields is required and cannot be empty"})
		return
	}

	// Validate each data field
	for i, dataField := range req.DataFields {
		if dataField.OwnerType == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_type is required and cannot be empty", i)})
			return
		}
		if dataField.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if len(dataField.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields is required and cannot be empty", i)})
			return
		}
		// Validate that no field is empty
		for j, field := range dataField.Fields {
			if field == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields[%d] cannot be empty", i, j)})
				return
			}
		}
	}

	// Process consent request using the engine
	response, err := s.engine.ProcessConsentRequest(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "create consent", "id", response.ConsentID, "owner", response.OwnerID, "existing", false)

	// Return the ConsentRecord
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (s *apiServer) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consents/", func(id string) (interface{}, int, error) {
		record, err := s.engine.GetConsentStatus(id)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(ErrConsentNotFound+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) updateConsent(w http.ResponseWriter, r *http.Request) {
	var req UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.UpdateConsent(id, req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentUpdateFailed+": %w", err)
		}

		// Return the ConsentRecord directly
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) revokeConsent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Reason string }
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.RevokeConsent(id, req.Reason)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil, http.StatusNotFound, err
			}
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentRevokeFailed+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

// Simple endpoint for consent website to approve/reject consent
func (s *apiServer) updateConsentStatus(w http.ResponseWriter, r *http.Request, req struct {
	ConsentID string `json:"consent_id"`
	Status    string `json:"status"` // "approved" or "rejected"
}) {
	if req.ConsentID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_id is required"})
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status must be 'approved' or 'rejected'"})
		return
	}

	// Get the consent record
	record, err := s.engine.GetConsentStatus(req.ConsentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "consent record not found"})
		return
	}

	// Update the status
	var newStatus string
	if req.Status == "approved" {
		newStatus = string(StatusApproved)
	} else {
		newStatus = string(StatusRejected)
	}

	record.Status = newStatus
	record.UpdatedAt = time.Now()

	// Store the updated record
	s.engine.(*consentEngineImpl).consentRecords[req.ConsentID] = record

	response := map[string]interface{}{
		"id":                      record.ConsentID,
		"status":                  string(record.Status),
		"updated_at":              record.UpdatedAt.Format(time.RFC3339),
		"approved_at":             record.UpdatedAt.Format(time.RFC3339),
		"data_owner_confirmation": true,
	}

	// If approved, redirect to orchestration engine's redirect endpoint
	if req.Status == "approved" {
		redirectURL := fmt.Sprintf("http://localhost:4000/consent-redirect?consent_id=%s", req.ConsentID)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *apiServer) getConsentPortalInfo(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		consentID, err := utils.ExtractQueryParam(r, "consent_id")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		record, err := s.engine.GetConsentStatus(consentID)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(ErrConsentNotFound+": %w", err)
		}

		// Convert to the user-facing ConsentPortalView
		consentView := record.ToConsentPortalView()

		return consentView, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByDataOwner(dataOwner)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"data_owner": dataOwner,
			"consents":   records,
			"count":      len(records),
		}, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consumer/", func(consumer string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByConsumer(consumer)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentGetFailed+": %w", err)
		}
		return map[string]interface{}{
			"consumer": consumer,
			"consents": records,
			"count":    len(records),
		}, http.StatusOK, nil
	})
}

func (s *apiServer) checkConsentExpiry(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		expiredRecords, err := s.engine.CheckConsentExpiry()
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(ErrConsentExpiryFailed+": %w", err)
		}
		utils.HandleSuccess(w, map[string]interface{}{
			"expired_records": expiredRecords,
			"count":           len(expiredRecords),
			"checked_at":      time.Now(),
		}, http.StatusOK, OpCheckConsentExpiry, map[string]interface{}{
			"expired_count": len(expiredRecords),
		})
		return map[string]interface{}{
			"expired_records": expiredRecords,
			"count":           len(expiredRecords),
			"checked_at":      time.Now(),
		}, http.StatusOK, nil
	})
}

// Route handlers - organized for better readability
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	switch {
	case path == "" && r.Method == http.MethodPost:
		// Check if this is a consent update (has consent_id) or new consent request
		s.handleConsentPost(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		// Handle GET /consents/{id} - get consent by ID
		consentID := strings.TrimPrefix(path, "/")
		s.getConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPost:
		// Handle POST /consents/{id} - update consent by ID
		consentID := strings.TrimPrefix(path, "/")
		s.updateConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		s.updateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		s.revokeConsent(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) consentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.processConsentPortalRequest(w, r)
	case http.MethodPut:
		s.processConsentPortalUpdate(w, r)
	case http.MethodGet:
		s.getConsentPortalInfo(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

// processConsentPortalRequest handles POST requests to the consent portal
func (s *apiServer) processConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentPortalCreateRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate required fields
	if req.AppID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "app_id is required and cannot be empty"})
		return
	}
	if req.Purpose == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "purpose is required and cannot be empty"})
		return
	}
	if req.SessionID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "session_id is required and cannot be empty"})
		return
	}
	if req.ConsentPortalURL == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_portal_url is required and cannot be empty"})
		return
	}
	if len(req.DataFields) == 0 {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "data_fields is required and cannot be empty"})
		return
	}

	// Validate each data field
	for i, dataField := range req.DataFields {
		if dataField.OwnerType == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_type is required and cannot be empty", i)})
			return
		}
		if dataField.OwnerID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].owner_id is required and cannot be empty", i)})
			return
		}
		if len(dataField.Fields) == 0 {
			utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields is required and cannot be empty", i)})
			return
		}
		// Validate that no field is empty
		for j, field := range dataField.Fields {
			if field == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: fmt.Sprintf("data_fields[%d].fields[%d] cannot be empty", i, j)})
				return
			}
		}
	}

	// Convert to ConsentRequest format
	consentReq := ConsentRequest{
		AppID:            req.AppID,
		DataFields:       req.DataFields,
		Purpose:          req.Purpose,
		SessionID:        req.SessionID,
		ConsentPortalURL: req.ConsentPortalURL,
	}

	// Process consent request using the engine
	response, err := s.engine.ProcessConsentRequest(consentReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "create consent via portal", "id", response.ConsentID, "owner", response.OwnerID, "existing", false)

	// Return the ConsentRecord
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

// processConsentPortalUpdate handles PUT requests to the consent portal
func (s *apiServer) processConsentPortalUpdate(w http.ResponseWriter, r *http.Request) {
	// Extract consent ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/consent-portal/")
	if path == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	var req ConsentPortalUpdateRequest

	// Parse request body
	body, err := utils.ReadRequestBody(r)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate required fields for update
	if req.Status == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status is required and cannot be empty"})
		return
	}
	if req.UpdatedBy == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "updated_by is required and cannot be empty"})
		return
	}

	// Convert to UpdateConsentRequest format
	updateReq := UpdateConsentRequest{
		Status:    ConsentStatus(req.Status),
		UpdatedBy: req.UpdatedBy,
		Reason:    req.Reason,
	}

	// Process consent update using the engine
	response, err := s.engine.UpdateConsent(path, updateReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", "update consent via portal", "id", response.ConsentID, "status", response.Status, "updated_by", req.UpdatedBy)

	// Return the updated ConsentRecord
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *apiServer) dataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByDataOwner(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) consumerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByConsumer(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) getConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Convert to the user-facing ConsentPortalView
	consentView := record.ToConsentPortalView()

	// Return only the UI-necessary fields
	utils.RespondWithJSON(w, http.StatusOK, consentView)
}

func (s *apiServer) updateConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status  string `json:"status"`
		OwnerID string `json:"owner_id"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate status
	if req.Status != "approved" && req.Status != "rejected" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status must be 'approved' or 'rejected'"})
		return
	}

	// Validate required fields
	if req.OwnerID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "owner_id is required"})
		return
	}

	// Note: We don't need to check if the record exists here since UpdateConsent will handle it

	// Convert status and set message based on status
	var newStatus string
	var message string
	if req.Status == "approved" {
		// For approved consents, set to approved
		newStatus = string(StatusApproved)
		message = "User approved consent via portal"
		slog.Info("DEBUG: Setting status to approved for consent", "consentId", consentID, "requestStatus", req.Status, "newStatus", newStatus)
	} else {
		newStatus = string(StatusRejected)
		message = "User rejected consent via portal"
		slog.Info("DEBUG: Setting status to rejected", "consentId", consentID, "requestStatus", req.Status, "newStatus", newStatus)
	}

	// Update the record
	updateReq := UpdateConsentRequest{
		Status:    ConsentStatus(newStatus),
		UpdatedBy: req.OwnerID,
		Reason:    message,
	}

	updatedRecord, err := s.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		} else {
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record: " + err.Error()})
		}
		return
	}

	// Log the operation
	slog.Info("Consent status updated", "consentId", consentID, "status", req.Status, "ownerId", req.OwnerID)

	// Build redirect URL with consent_id for pending status
	if updatedRecord.Status == string(StatusPending) {
		updatedRecord.ConsentPortalURL = fmt.Sprintf("http://localhost:5173/?consent_id=%s", updatedRecord.ConsentID)
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedRecord)
}

func (s *apiServer) consentWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the consent website HTML file
	http.ServeFile(w, r, "consent-website.html")
}

func (s *apiServer) adminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		s.checkConsentExpiry(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func main() {
	// Load configuration using flags
	cfg := config.LoadConfig("consent-engine")

	// Setup logging
	utils.SetupLogging(cfg.Logging.Format, cfg.Logging.Level)

	slog.Info("Starting consent engine",
		"environment", cfg.Environment,
		"port", cfg.Service.Port,
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	// Initialize consent engine
	consentPortalUrl := getEnvOrDefault("CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := NewConsentEngine(consentPortalUrl)
	server := &apiServer{engine: engine}

	// Setup routes using utils
	mux := http.NewServeMux()
	mux.Handle("/consents", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consents/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consent-portal/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentPortalHandler)))
	mux.Handle("/consent-website", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentWebsiteHandler)))
	mux.Handle("/data-owner/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.dataOwnerHandler)))
	mux.Handle("/consumer/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consumerHandler)))
	mux.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.adminHandler)))
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("consent-engine")))

	// Create server using utils
	serverConfig := &utils.ServerConfig{
		Port:         cfg.Service.Port,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  60 * time.Second,
	}

	// Wrap the mux with CORS middleware
	handler := corsMiddleware(mux)
	httpServer := utils.CreateServer(serverConfig, handler)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(httpServer, "consent-engine"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
