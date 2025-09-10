package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
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

// ProviderService returns the provider service instance
func (s *APIServer) ProviderService() *services.ProviderService {
	return s.providerService
}

// SetupRoutes configures all API routes
func (s *APIServer) SetupRoutes(mux *http.ServeMux) {
	// Consumer routes
	mux.Handle("/consumers", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumers)))
	mux.Handle("/consumers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerByID)))

	// Consumer application routes (RESTful)
	mux.Handle("/consumer-applications", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerApplications)))
	mux.Handle("/consumer-applications/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleConsumerApplicationByID)))

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
	id := utils.ExtractIDFromPathString(r.URL.Path)
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

// Service method wrappers for use with generic handlers
func (s *APIServer) createConsumerServiceMethod(req interface{}) (interface{}, error) {
	createReq := req.(models.CreateConsumerRequest)
	return s.consumerService.CreateConsumer(createReq)
}

func (s *APIServer) parseAndCreateConsumer(req interface{}) (interface{}, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var createReq models.CreateConsumerRequest
	if err := json.Unmarshal(reqBytes, &createReq); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return s.consumerService.CreateConsumer(createReq)
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
			consumers, err := s.consumerService.GetAllConsumers()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(consumers, len(consumers)), nil
		},
		func(req interface{}) (interface{}, error) {
			return s.parseAndCreateConsumer(req)
		},
	)
}

func (s *APIServer) handleConsumerByID(w http.ResponseWriter, r *http.Request) {
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.consumerService.GetConsumer(id) },
		nil, // No update for consumers
		nil, // No delete for consumers
	)
}

// Consumer application handlers (RESTful)
func (s *APIServer) handleConsumerApplications(w http.ResponseWriter, r *http.Request) {
	// GET /consumer-applications - Get all applications (admin view)
	s.handleCollection(w, r,
		func() (interface{}, error) {
			apps, err := s.consumerService.GetAllConsumerApps()
			if err != nil {
				return nil, err
			}
			return utils.CreateCollectionResponse(apps, len(apps)), nil
		},
		nil, // No POST at this level - use /consumer-applications/:consumerId
	)
}

func (s *APIServer) handleConsumerApplicationByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := utils.ExtractIDFromPathString(path)

	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "ID is required")
		return
	}

	// Determine if this is a consumer ID or submission ID based on the ID format and HTTP method
	if strings.HasPrefix(id, "consumer_") {
		// This is a consumer ID - handle consumer-specific operations
		s.handleConsumerApplicationsForConsumer(w, r, id)
	} else if strings.HasPrefix(id, "sub_") {
		// This is a submission ID - handle individual application operations
		s.handleIndividualConsumerApplication(w, r, id)
	} else {
		// Unknown ID format - try consumer first, then submission
		if r.Method == "POST" || r.Method == "GET" {
			// For POST/GET, assume it's a consumer ID
			s.handleConsumerApplicationsForConsumer(w, r, id)
		} else {
			// For other methods, assume it's a submission ID
			s.handleIndividualConsumerApplication(w, r, id)
		}
	}
}

// Handle operations for a specific consumer's applications
func (s *APIServer) handleConsumerApplicationsForConsumer(w http.ResponseWriter, r *http.Request, consumerID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /consumer-applications/:consumerId - Get applications for specific consumer
		apps, err := s.consumerService.GetConsumerAppsByConsumerID(consumerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(apps, len(apps)))
	case http.MethodPost:
		// POST /consumer-applications/:consumerId - Create application for specific consumer
		var req models.CreateConsumerAppRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Set the consumer ID from the URL
		req.ConsumerID = consumerID

		app, err := s.consumerService.CreateConsumerApp(req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, app)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handle operations for individual consumer applications
func (s *APIServer) handleIndividualConsumerApplication(w http.ResponseWriter, r *http.Request, submissionID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /consumer-applications/:submissionId - Get specific application
		app, err := s.consumerService.GetConsumerApp(submissionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, app)
	case http.MethodPut:
		// PUT /consumer-applications/:submissionId - Update application (admin approval)
		var req models.UpdateConsumerAppRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		response, err := s.consumerService.UpdateConsumerApp(submissionID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, response)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
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
		nil, // No POST at this level - use /provider-schemas/:providerId
	)
}

func (s *APIServer) handleProviderSchemaByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	providerID := utils.ExtractIDFromPathString(path)

	if providerID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Provider ID is required")
		return
	}

	// Handle provider-specific schema creation
	if r.Method == http.MethodPost {
		s.handleCreateProviderSchemaSDL(w, r, providerID)
		return
	}

	// Handle other operations (GET, PUT) for individual schemas
	s.handleItem(w, r,
		func(id string) (interface{}, error) { return s.providerService.GetProviderSchema(id) },
		s.parseAndUpdateProviderSchema,
		nil, // No delete for schemas
	)
}

// handleCreateProviderSchemaSDL handles POST /provider-schemas/:providerId
func (s *APIServer) handleCreateProviderSchemaSDL(w http.ResponseWriter, r *http.Request, providerID string) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateProviderSchemaSDLRequest
	if err := utils.ParseJSONRequest(r, &req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schema, err := s.providerService.CreateProviderSchemaSDL(providerID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, schema)
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
