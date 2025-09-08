package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/utils"
)

const defaultPort = "8081"

// apiServer holds dependencies for the HTTP handlers, like the consent engine
type apiServer struct {
	engine ConsentEngine
}

// consentHandler manages creating and retrieving consent records
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consent")

	switch {
	case path == "" && r.Method == http.MethodPost:
		s.createConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		s.getConsentStatus(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		s.updateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		s.revokeConsent(w, r)
	default:
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

// consentPortalHandler manages consent portal interactions
func (s *apiServer) consentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.processConsentPortalRequest(w, r)
	case http.MethodGet:
		s.getConsentPortalInfo(w, r)
	default:
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

// dataOwnerHandler manages data owner consent operations
func (s *apiServer) dataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/data-owner")

	switch {
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		s.getConsentsByDataOwner(w, r)
	default:
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

// consumerHandler manages consumer consent operations
func (s *apiServer) consumerHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consumer")

	switch {
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		s.getConsentsByConsumer(w, r)
	default:
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

// adminHandler manages administrative operations
func (s *apiServer) adminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")

	switch {
	case path == "/expiry-check" && r.Method == http.MethodPost:
		s.checkConsentExpiry(w, r)
	default:
		slog.Warn("Method not allowed", "method", r.Method, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

// createConsent handles the HTTP request for creating a new consent record
func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req CreateConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request body for create consent", "error", err, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
		return
	}

	record, err := s.engine.CreateConsent(req)
	if err != nil {
		slog.Error("Failed to create consent record", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to create consent record"})
		return
	}

	slog.Info("Created new consent record", "id", record.ID, "owner", record.DataOwner)
	utils.RespondWithJSON(w, http.StatusCreated, record)
}

// getConsentStatus handles the HTTP request for retrieving a consent record
func (s *apiServer) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/consent/")
	if id == "" {
		slog.Warn("Consent ID is missing in request path", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	record, err := s.engine.GetConsentStatus(id)
	if err != nil {
		slog.Warn("Consent record not found", "id", id, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, record)
}

// updateConsent handles the HTTP request for updating a consent record
func (s *apiServer) updateConsent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := strings.TrimPrefix(r.URL.Path, "/consent/")
	if id == "" {
		slog.Warn("Consent ID is missing in request path", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	var req UpdateConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request body for update consent", "error", err, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
		return
	}

	record, err := s.engine.UpdateConsent(id, req)
	if err != nil {
		slog.Error("Failed to update consent record", "error", err, "id", id)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to update consent record"})
		return
	}

	slog.Info("Updated consent record", "id", record.ID, "status", record.Status)
	utils.RespondWithJSON(w, http.StatusOK, record)
}

// revokeConsent handles the HTTP request for revoking a consent record
func (s *apiServer) revokeConsent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := strings.TrimPrefix(r.URL.Path, "/consent/")
	if id == "" {
		slog.Warn("Consent ID is missing in request path", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request body for revoke consent", "error", err, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
		return
	}

	record, err := s.engine.RevokeConsent(id, req.Reason)
	if err != nil {
		slog.Error("Failed to revoke consent record", "error", err, "id", id)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to revoke consent record"})
		return
	}

	slog.Info("Revoked consent record", "id", record.ID, "reason", req.Reason)
	utils.RespondWithJSON(w, http.StatusOK, record)
}

// processConsentPortalRequest handles consent portal interactions
func (s *apiServer) processConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req ConsentPortalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid request body for consent portal", "error", err, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
		return
	}

	record, err := s.engine.ProcessConsentPortalRequest(req)
	if err != nil {
		slog.Error("Failed to process consent portal request", "error", err, "consent_id", req.ConsentID)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to process consent portal request"})
		return
	}

	slog.Info("Processed consent portal request", "id", record.ID, "action", req.Action, "status", record.Status)
	utils.RespondWithJSON(w, http.StatusOK, record)
}

// getConsentPortalInfo handles getting consent portal information
func (s *apiServer) getConsentPortalInfo(w http.ResponseWriter, r *http.Request) {
	// Extract consent ID from query parameters or path
	consentID := r.URL.Query().Get("consent_id")
	if consentID == "" {
		slog.Warn("Consent ID is missing in request", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
		return
	}

	record, err := s.engine.GetConsentStatus(consentID)
	if err != nil {
		slog.Warn("Consent record not found", "id", consentID, "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
		return
	}

	// Return portal information
	portalInfo := map[string]interface{}{
		"consent_id":         record.ID,
		"status":             record.Status,
		"data_consumer":      record.DataConsumer,
		"data_owner":         record.DataOwner,
		"fields":             record.Fields,
		"consent_portal_url": record.ConsentPortalURL,
		"expires_at":         record.ExpiresAt,
		"created_at":         record.CreatedAt,
	}

	utils.RespondWithJSON(w, http.StatusOK, portalInfo)
}

// getConsentsByDataOwner handles getting all consent records for a data owner
func (s *apiServer) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	dataOwner := strings.TrimPrefix(r.URL.Path, "/data-owner/")
	if dataOwner == "" {
		slog.Warn("Data owner ID is missing in request path", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Data owner ID is required"})
		return
	}

	records, err := s.engine.GetConsentsByDataOwner(dataOwner)
	if err != nil {
		slog.Error("Failed to get consent records for data owner", "error", err, "data_owner", dataOwner)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to get consent records"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data_owner": dataOwner,
		"consents":   records,
		"count":      len(records),
	})
}

// getConsentsByConsumer handles getting all consent records for a consumer
func (s *apiServer) getConsentsByConsumer(w http.ResponseWriter, r *http.Request) {
	consumer := strings.TrimPrefix(r.URL.Path, "/consumer/")
	if consumer == "" {
		slog.Warn("Consumer ID is missing in request path", "path", r.URL.Path)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consumer ID is required"})
		return
	}

	records, err := s.engine.GetConsentsByConsumer(consumer)
	if err != nil {
		slog.Error("Failed to get consent records for consumer", "error", err, "consumer", consumer)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to get consent records"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"consumer": consumer,
		"consents": records,
		"count":    len(records),
	})
}

// checkConsentExpiry handles checking and updating expired consent records
func (s *apiServer) checkConsentExpiry(w http.ResponseWriter, r *http.Request) {
	expiredRecords, err := s.engine.CheckConsentExpiry()
	if err != nil {
		slog.Error("Failed to check consent expiry", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to check consent expiry"})
		return
	}

	slog.Info("Checked consent expiry", "expired_count", len(expiredRecords))
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"expired_records": expiredRecords,
		"count":           len(expiredRecords),
		"checked_at":      time.Now(),
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Apply panic recovery middleware to all handlers
	http.Handle("/consent", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	http.Handle("/consent/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	http.Handle("/consent-portal/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentPortalHandler)))
	http.Handle("/data-owner/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.dataOwnerHandler)))
	http.Handle("/consumer/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consumerHandler)))
	http.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.adminHandler)))

	// Get port from environment variable, falling back to the default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	listenAddr := fmt.Sprintf(":%s", port)

	slog.Info("Consent Engine server starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		slog.Error("could not start Consent Engine server", "error", err)
		os.Exit(1)
	}
}
