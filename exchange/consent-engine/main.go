package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gov-dx-sandbox/exchange/utils"
)

const defaultPort = "8081"

// apiServer holds dependencies for the HTTP handlers, like the consent engine
type apiServer struct {
	engine ConsentEngine
}

// consentHandler manages creating and retrieving consent records
func (s *apiServer) consentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createConsent(w, r)
	case http.MethodGet:
		s.getConsentStatus(w, r)
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	engine := NewConsentEngine()
	server := &apiServer{engine: engine}

	// Apply panic recovery middleware to the main handler
	http.Handle("/consent/", utils.PanicRecoveryMiddleware(http.HandlerFunc(server.consentHandler)))

	// Get port from environment variable, falling back to the default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	listenAddr := fmt.Sprintf(":%s", port)

	slog.Info("CME server starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		slog.Error("could not start CME server", "error", err)
		os.Exit(1)
	}
}
