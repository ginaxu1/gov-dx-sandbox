package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/config"
	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

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

	// Otherwise, treat as new consent request
	var req ConsentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Process new consent request
	s.processConsentRequest(w, r, req)
}

func (s *apiServer) processConsentRequest(w http.ResponseWriter, r *http.Request, req ConsentRequest) {
	response, err := s.engine.ProcessConsentRequest(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent request: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Operation successful", "operation", constants.OpCreateConsent, "consentId", response.ConsentID, "status", response.Status)

	// Return the ConsentResponse directly as it already has the correct format
	utils.RespondWithJSON(w, http.StatusCreated, response)
}

func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	var req CreateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.CreateConsent(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentCreateFailed+": %w", err)
		}

		// Format timestamps as ISO strings
		var expiresAtStr string
		if record.ExpiresAt != nil {
			expiresAtStr = record.ExpiresAt.Format(time.RFC3339)
		}
		createdAtStr := record.CreatedAt.Format(time.RFC3339)

		response := map[string]interface{}{
			"consentId":        record.ConsentID,
			"status":           record.Status,
			"dataConsumer":     record.DataConsumer,
			"dataOwner":        record.OwnerID,
			"fields":           record.Fields,
			"consentPortalUrl": fmt.Sprintf("/consent-portal/%s", record.ConsentID),
			"expiresAt":        expiresAtStr,
			"createdAt":        createdAtStr,
		}

		utils.HandleSuccess(w, response, http.StatusCreated, constants.OpCreateConsent, map[string]interface{}{
			"id": record.ConsentID, "owner": record.OwnerID,
		})
		return response, http.StatusCreated, nil
	})
}

func (s *apiServer) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consents/", func(id string) (interface{}, int, error) {
		record, err := s.engine.GetConsentStatus(id)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf(constants.ErrConsentNotFound+": %w", err)
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
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentUpdateFailed+": %w", err)
		}
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
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentRevokeFailed+": %w", err)
		}
		return record, http.StatusOK, nil
	})
}

// Portal and admin handlers
func (s *apiServer) processConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentPortalRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.ProcessConsentPortalRequest(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrPortalRequestFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusOK, constants.OpProcessPortalRequest, map[string]interface{}{
			"id": record.ConsentID, "action": req.Action, "status": record.Status,
		})
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
	var newStatus ConsentStatus
	if req.Status == "approved" {
		newStatus = StatusApproved
	} else {
		newStatus = StatusRejected
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
			return nil, http.StatusNotFound, fmt.Errorf(constants.ErrConsentNotFound+": %w", err)
		}

		// Format timestamps as ISO strings
		var expiresAtStr string
		if record.ExpiresAt != nil {
			expiresAtStr = record.ExpiresAt.Format(time.RFC3339)
		}
		createdAtStr := record.CreatedAt.Format(time.RFC3339)

		return map[string]interface{}{
			"consentId":        record.ConsentID,
			"status":           record.Status,
			"dataConsumer":     record.DataConsumer,
			"dataOwner":        record.OwnerID,
			"fields":           record.Fields,
			"consentPortalUrl": fmt.Sprintf("/consent-portal/%s", record.ConsentID),
			"expiresAt":        expiresAtStr,
			"createdAt":        createdAtStr,
		}, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByDataOwner(dataOwner)
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

func (s *apiServer) getConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consumer/", func(consumer string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByConsumer(consumer)
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

func (s *apiServer) checkConsentExpiry(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		expiredRecords, err := s.engine.CheckConsentExpiry()
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

// sendConsentOTP sends an OTP for consent verification
func (s *apiServer) sendConsentOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		consentID, err := utils.ExtractIDFromPath(r, "/consents/")
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		// Remove the /otp suffix from the path
		consentID = strings.TrimSuffix(consentID, "/otp")

		response, err := s.engine.SendConsentOTP(consentID, req.PhoneNumber)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to send OTP: %w", err)
		}

		return response, http.StatusOK, nil
	})
}

// updateConsentWithOTP handles POST /consent/update with OTP verification
func (s *apiServer) updateConsentWithOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConsentID string `json:"consent_id"`
		Status    string `json:"status"` // "approved" or "rejected"
		OTP       string `json:"otp"`
		OwnerID   string `json:"owner_id"`
	}

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
	if req.ConsentID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_id is required"})
		return
	}

	if req.Status == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status is required"})
		return
	}

	if req.OTP == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "otp is required"})
		return
	}

	if req.OwnerID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "owner_id is required"})
		return
	}

	// Validate status
	if req.Status != "approved" && req.Status != "rejected" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status must be 'approved' or 'rejected'"})
		return
	}

	// Verify OTP (simplified for testing - always accept "000000")
	if req.OTP != "000000" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid OTP"})
		return
	}

	// Get the consent record
	record, err := s.engine.GetConsentStatus(req.ConsentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "consent record not found"})
		return
	}

	// Verify owner ID matches
	if record.OwnerID != req.OwnerID {
		utils.RespondWithJSON(w, http.StatusForbidden, utils.ErrorResponse{Error: "owner_id does not match consent record"})
		return
	}

	// Update the status
	var newStatus ConsentStatus
	if req.Status == "approved" {
		newStatus = StatusApproved
	} else {
		newStatus = StatusRejected
	}

	record.Status = newStatus
	record.UpdatedAt = time.Now()

	// Store the updated record

	updateReq := UpdateConsentRequest{
		Status: newStatus,
	}
	_, err = s.engine.UpdateConsent(record.ConsentID, updateReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record"})
		return
	}

	// Return success response
	response := map[string]interface{}{
		"consent_id": record.ConsentID,
		"status":     string(record.Status),
		"updated_at": record.UpdatedAt.Format(time.RFC3339),
		"message":    "Consent status updated successfully",
	}

	utils.HandleSuccess(w, response, http.StatusOK, constants.OpUpdateConsent, map[string]interface{}{
		"consentId": record.ConsentID, "status": string(record.Status),
	})
}

// Route handlers - organized for better readability
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consents")
	fmt.Printf("DEBUG: Path=%s, Method=%s\n", path, r.Method)
	switch {
	case path == "" && r.Method == http.MethodPost:
		// Check if this is a consent update (has consent_id) or new consent request
		s.handleConsentPost(w, r)
	case path == "/update" && r.Method == http.MethodPost:
		s.updateConsentWithOTP(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		// Handle GET /consent/{id} - get consent by ID
		consentID := strings.TrimPrefix(path, "/")
		s.getConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPost:
		// Handle POST /consent/{id} - update consent by ID with OTP
		consentID := strings.TrimPrefix(path, "/")
		s.updateConsentByID(w, r, consentID)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		s.updateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		s.revokeConsent(w, r)
	case strings.HasSuffix(path, "/otp") && r.Method == http.MethodPost:
		s.sendConsentOTP(w, r)
	default:
		fmt.Printf("DEBUG: No match found for path=%s, method=%s\n", path, r.Method)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
}

func (s *apiServer) consentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.processConsentPortalRequest(w, r)
	case http.MethodGet:
		s.getConsentPortalInfo(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: constants.StatusMethodNotAllowed})
	}
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

func (s *apiServer) consentUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req ConsentRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate required fields
	if req.ConsentID == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "consent_id is required"})
		return
	}

	if req.Status == "" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status is required"})
		return
	}

	// Validate status values
	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
	}
	if !validStatuses[string(req.Status)] {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "status must be 'pending', 'approved', or 'rejected'"})
		return
	}

	// Validate type values
	validTypes := map[string]bool{
		"realtime": true,
		"offline":  true,
	}
	if !validTypes[string(req.Type)] {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "type must be 'realtime' or 'offline'"})
		return
	}

	// Create or update the consent record
	record, err := s.engine.CreateOrUpdateConsentRecord(req)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to create/update consent record: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Consent record created/updated", "consentId", record.ConsentID, "status", record.Status)

	// Return the created/updated record
	utils.RespondWithJSON(w, http.StatusOK, record)
}

func (s *apiServer) getConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	record, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Convert to the expected response format with consent_uuid field
	response := map[string]interface{}{
		"consent_uuid":  record.ConsentID,
		"owner_id":      record.OwnerID,
		"data_consumer": record.DataConsumer,
		"status":        string(record.Status),
		"type":          string(record.Type),
		"created_at":    record.CreatedAt.Format(time.RFC3339),
		"updated_at":    record.UpdatedAt.Format(time.RFC3339),
		"fields":        record.Fields,
		"session_id":    record.SessionID,
		"redirect_url":  record.RedirectURL,
	}

	if record.ExpiresAt != nil {
		response["expires_at"] = record.ExpiresAt.Format(time.RFC3339)
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (s *apiServer) updateConsentByID(w http.ResponseWriter, r *http.Request, consentID string) {
	var req struct {
		Status string `json:"status"`
		OTP    string `json:"otp"`
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

	// Validate OTP (hardcoded to 000000 for testing)
	if req.OTP != "000000" {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid OTP"})
		return
	}

	// Get existing record
	_, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Convert status
	var newStatus ConsentStatus
	if req.Status == "approved" {
		newStatus = StatusApproved
	} else {
		newStatus = StatusRejected
	}

	// Update the record
	updateReq := UpdateConsentRequest{
		Status:    newStatus,
		UpdatedBy: "data_owner",
		Reason:    "User decision with OTP verification",
		OTP:       req.OTP,
	}

	updatedRecord, err := s.engine.UpdateConsent(consentID, updateReq)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record: " + err.Error()})
		return
	}

	// Log the operation
	slog.Info("Consent status updated with OTP", "consentId", consentID, "status", req.Status)

	// Return success response
	response := map[string]interface{}{
		"consent_uuid": updatedRecord.ConsentID,
		"status":       string(updatedRecord.Status),
		"updated_at":   updatedRecord.UpdatedAt.Format(time.RFC3339),
		"message":      "Consent status updated successfully",
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
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
	engine := NewConsentEngine(cfg.ConsentPortalUrl)
	server := &apiServer{engine: engine}

	// Setup routes using utils
	mux := http.NewServeMux()
	mux.Handle("/consents", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consents/update", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentUpdateHandler)))
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
