package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/exchange/utils"
)

// apiServer holds dependencies for the HTTP handlers
type apiServer struct {
	engine ConsentEngine
}

// Consent handlers using utils patterns
func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	var req CreateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.CreateConsent(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create consent record: %w", err)
		}
		slog.Info("Created consent record", "id", record.ID, "owner", record.DataOwner)
		return record, http.StatusCreated, nil
	})
}

func (s *apiServer) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consent/", func(id string) (interface{}, int, error) {
		record, err := s.engine.GetConsentStatus(id)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf("consent record not found: %w", err)
		}
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) updateConsent(w http.ResponseWriter, r *http.Request) {
	var req UpdateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id := strings.TrimPrefix(r.URL.Path, "/consent/")
		if id == "" {
			return nil, http.StatusBadRequest, fmt.Errorf("consent ID is required")
		}

		record, err := s.engine.UpdateConsent(id, req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to update consent record: %w", err)
		}
		slog.Info("Updated consent record", "id", record.ID, "status", record.Status)
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) revokeConsent(w http.ResponseWriter, r *http.Request) {
	var req struct{ Reason string }
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		id := strings.TrimPrefix(r.URL.Path, "/consent/")
		if id == "" {
			return nil, http.StatusBadRequest, fmt.Errorf("consent ID is required")
		}

		record, err := s.engine.RevokeConsent(id, req.Reason)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to revoke consent record: %w", err)
		}
		slog.Info("Revoked consent record", "id", record.ID, "reason", req.Reason)
		return record, http.StatusOK, nil
	})
}

// Portal and admin handlers
func (s *apiServer) processConsentPortalRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentPortalRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.ProcessConsentPortalRequest(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to process consent portal request: %w", err)
		}
		slog.Info("Processed consent portal request", "id", record.ID, "action", req.Action, "status", record.Status)
		return record, http.StatusOK, nil
	})
}

func (s *apiServer) getConsentPortalInfo(w http.ResponseWriter, r *http.Request) {
	utils.GenericHandler(w, r, func() (interface{}, int, error) {
		consentID := r.URL.Query().Get("consent_id")
		if consentID == "" {
			return nil, http.StatusBadRequest, fmt.Errorf("consent ID is required")
		}

		record, err := s.engine.GetConsentStatus(consentID)
		if err != nil {
			return nil, http.StatusNotFound, fmt.Errorf("consent record not found: %w", err)
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

func (s *apiServer) getConsentsByDataOwner(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/data-owner/", func(dataOwner string) (interface{}, int, error) {
		records, err := s.engine.GetConsentsByDataOwner(dataOwner)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to get consent records: %w", err)
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
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to get consent records: %w", err)
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
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to check consent expiry: %w", err)
		}
		slog.Info("Checked consent expiry", "expired_count", len(expiredRecords))
		return map[string]interface{}{
			"expired_records": expiredRecords,
			"count":           len(expiredRecords),
			"checked_at":      time.Now(),
		}, http.StatusOK, nil
	})
}

// Route handlers
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
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

func (s *apiServer) consentPortalHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.processConsentPortalRequest(w, r)
	case http.MethodGet:
		s.getConsentPortalInfo(w, r)
	default:
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

func (s *apiServer) dataOwnerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByDataOwner(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

func (s *apiServer) consumerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.getConsentsByConsumer(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

func (s *apiServer) adminHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "/expiry-check" && r.Method == http.MethodPost {
		s.checkConsentExpiry(w, r)
	} else {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
	}
}

func (s *apiServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "consent-engine"})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Setup routes using utils
	mux := http.NewServeMux()
	mux.Handle("/consent", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consent/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))
	mux.Handle("/consent-portal/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentPortalHandler)))
	mux.Handle("/data-owner/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.dataOwnerHandler)))
	mux.Handle("/consumer/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consumerHandler)))
	mux.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.adminHandler)))
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("consent-engine")))

	// Setup server with default configuration
	config := utils.DefaultServerConfig()
	config.Port = utils.GetEnvOrDefault("PORT", "8081")
	serverInstance := utils.CreateServer(config, mux)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(serverInstance, "consent-engine"); err != nil {
		os.Exit(1)
	}
}
