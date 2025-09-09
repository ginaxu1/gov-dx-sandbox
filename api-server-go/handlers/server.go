package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/exchange/utils"
)

// APIServer manages all API routes and handlers
type APIServer struct {
	consumerService *services.ConsumerService
	providerService *services.ProviderService
	adminService    *services.AdminService
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
	consumerService := services.NewConsumerService()
	providerService := services.NewProviderService()
	return &APIServer{
		consumerService: consumerService,
		providerService: providerService,
		adminService:    services.NewAdminServiceWithServices(consumerService, providerService),
	}
}

// SetupRoutes configures all API routes
func (s *APIServer) SetupRoutes(mux *http.ServeMux) {
	// Consumer routes
	mux.Handle("/consumers", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumers)))
	mux.Handle("/consumers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerByID)))

	// Provider routes
	mux.Handle("/provider-submissions", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSubmissions)))
	mux.Handle("/provider-submissions/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSubmissionByID)))
	mux.Handle("/provider-profiles", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderProfiles)))
	mux.Handle("/provider-profiles/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderProfileByID)))
	mux.Handle("/provider-schemas", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSchemas)))
	mux.Handle("/provider-schemas/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviderSchemaByID)))

	// Admin routes
	mux.Handle("/admin/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleAdmin)))
}

// Generic handler for collection endpoints (GET all, POST create)
func (s *APIServer) handleCollection(w http.ResponseWriter, r *http.Request, getter func() (interface{}, error), creator func(interface{}) (interface{}, error)) {
	switch r.Method {
	case http.MethodGet:
		items, err := getter()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve items")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, items)
	case http.MethodPost:
		var req interface{}
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		item, err := creator(req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusCreated, item)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Generic handler for item endpoints (GET, PUT, DELETE by ID)
func (s *APIServer) handleItem(w http.ResponseWriter, r *http.Request, getter func(string) (interface{}, error), updater func(string, interface{}) (interface{}, error), deleter func(string) error) {
	id := utils.ExtractIDFromPath(r.URL.Path)
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := getter(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Item not found")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, item)
	case http.MethodPut:
		var req interface{}
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		item, err := updater(id, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := deleter(id); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete item")
			return
		}
		utils.RespondWithSuccess(w, http.StatusNoContent, nil)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper functions for parsing requests
func (s *APIServer) parseAndCreateApplication(req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var createReq models.CreateApplicationRequest
	if err := json.Unmarshal(reqBytes, &createReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.consumerService.CreateApplication(createReq)
}

func (s *APIServer) parseAndUpdateApplication(id string, req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var updateReq models.UpdateApplicationRequest
	if err := json.Unmarshal(reqBytes, &updateReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.consumerService.UpdateApplication(id, updateReq)
}

func (s *APIServer) parseAndCreateProviderSubmission(req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var createReq models.CreateProviderSubmissionRequest
	if err := json.Unmarshal(reqBytes, &createReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.providerService.CreateProviderSubmission(createReq)
}

func (s *APIServer) parseAndUpdateProviderSubmission(id string, req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var updateReq models.UpdateProviderSubmissionRequest
	if err := json.Unmarshal(reqBytes, &updateReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.providerService.UpdateProviderSubmission(id, updateReq)
}

func (s *APIServer) parseAndCreateProviderSchema(req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var createReq models.CreateProviderSchemaRequest
	if err := json.Unmarshal(reqBytes, &createReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.providerService.CreateProviderSchema(createReq)
}

func (s *APIServer) parseAndUpdateProviderSchema(id string, req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var updateReq models.UpdateProviderSchemaRequest
	if err := json.Unmarshal(reqBytes, &updateReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.providerService.UpdateProviderSchema(id, updateReq)
}

// Consumer handlers
func (s *APIServer) handleConsumers(w http.ResponseWriter, r *http.Request) {
	s.handleCollection(w, r,
		func() (interface{}, error) {
			apps, err := s.consumerService.GetAllApplications()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(apps, len(apps)), nil
		},
		func(req interface{}) (interface{}, error) {
			return s.parseAndCreateApplication(req)
		},
	)
}

func (s *APIServer) handleConsumerByID(w http.ResponseWriter, r *http.Request) {
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.consumerService.GetApplication(id) },
		s.parseAndUpdateApplication,
		func(id string) error { return s.consumerService.DeleteApplication(id) },
	)
}

// Provider submission handlers
func (s *APIServer) handleProviderSubmissions(w http.ResponseWriter, r *http.Request) {
	s.handleCollection(w, r,
		func() (interface{}, error) {
			subs, err := s.providerService.GetAllProviderSubmissions()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(subs, len(subs)), nil
		},
		s.parseAndCreateProviderSubmission,
	)
}

func (s *APIServer) handleProviderSubmissionByID(w http.ResponseWriter, r *http.Request) {
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.providerService.GetProviderSubmission(id) },
		s.parseAndUpdateProviderSubmission,
		nil, // No delete for submissions
	)
}

// Provider profile handlers
func (s *APIServer) handleProviderProfiles(w http.ResponseWriter, r *http.Request) {
	s.handleCollection(w, r,
		func() (interface{}, error) {
			profiles, err := s.providerService.GetAllProviderProfiles()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(profiles, len(profiles)), nil
		},
		nil, // No create for profiles (they're created via submission approval)
	)
}

func (s *APIServer) handleProviderProfileByID(w http.ResponseWriter, r *http.Request) {
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.providerService.GetProviderProfile(id) },
		nil, // No update for profiles
		nil, // No delete for profiles
	)
}

// Provider schema handlers
func (s *APIServer) handleProviderSchemas(w http.ResponseWriter, r *http.Request) {
	s.handleCollection(w, r,
		func() (interface{}, error) {
			schemas, err := s.providerService.GetAllProviderSchemas()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(schemas, len(schemas)), nil
		},
		s.parseAndCreateProviderSchema,
	)
}

func (s *APIServer) handleProviderSchemaByID(w http.ResponseWriter, r *http.Request) {
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.providerService.GetProviderSchema(id) },
		s.parseAndUpdateProviderSchema,
		nil, // No delete for schemas
	)
}

// Admin handler
func (s *APIServer) handleAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")

	switch {
	case path == "/dashboard" && r.Method == http.MethodGet:
		dashboard, err := s.adminService.GetDashboard()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get dashboard")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, dashboard)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
