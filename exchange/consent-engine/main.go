package main

import (
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

// apiServer holds dependencies for the HTTP handlers
type apiServer struct {
	engine ConsentEngine
}

// Consent handlers - organized for better readability
func (s *apiServer) processConsentRequest(w http.ResponseWriter, r *http.Request) {
	var req ConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.ProcessConsentRequest(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to process consent request: %w", err)
		}
		return record, http.StatusCreated, nil
	})
}

func (s *apiServer) createConsent(w http.ResponseWriter, r *http.Request) {
	var req CreateConsentRequest
	utils.JSONHandler(w, r, &req, func() (interface{}, int, error) {
		record, err := s.engine.CreateConsent(req)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf(constants.ErrConsentCreateFailed+": %w", err)
		}
		utils.HandleSuccess(w, record, http.StatusCreated, constants.OpCreateConsent, map[string]interface{}{
			"id": record.ID, "owner": record.DataOwner,
		})
		return record, http.StatusCreated, nil
	})
}

func (s *apiServer) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	utils.PathHandler(w, r, "/consent/", func(id string) (interface{}, int, error) {
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
		id, err := utils.ExtractIDFromPath(r, "/consent/")
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
		id, err := utils.ExtractIDFromPath(r, "/consent/")
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
			"id": record.ID, "action": req.Action, "status": record.Status,
		})
		return record, http.StatusOK, nil
	})
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
		consentID, err := utils.ExtractIDFromPath(r, "/consent/")
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

// Route handlers - organized for better readability
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/consent")
	switch {
	case path == "" && r.Method == http.MethodPost:
		s.processConsentRequest(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodGet:
		s.getConsentStatus(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodPut:
		s.updateConsent(w, r)
	case strings.HasPrefix(path, "/") && r.Method == http.MethodDelete:
		s.revokeConsent(w, r)
	case strings.HasSuffix(path, "/otp") && r.Method == http.MethodPost:
		s.sendConsentOTP(w, r)
	default:
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

	// Create server using utils
	serverConfig := &utils.ServerConfig{
		Port:         cfg.Service.Port,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  60 * time.Second,
	}
	httpServer := utils.CreateServer(serverConfig, mux)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(httpServer, "consent-engine"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
