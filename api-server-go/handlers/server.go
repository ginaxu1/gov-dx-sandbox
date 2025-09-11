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

	// RESTful provider routes
	mux.Handle("/providers/", utils.PanicRecoveryMiddleware(http.HandlerFunc(s.handleProviders)))

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

// handleProviders handles RESTful provider routes
func (s *APIServer) handleProviders(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract provider ID from path like /providers/{provider-id}/...
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "providers" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid provider path")
		return
	}

	providerID := pathParts[1]
	if providerID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Provider ID is required")
		return
	}

	// Check if this is a schemas sub-resource (approved schemas only)
	if strings.HasSuffix(path, "/schemas") {
		s.handleProviderSchemas(w, r, providerID)
		return
	}

	// Check if this is a schema-submissions sub-resource
	if strings.HasSuffix(path, "/schema-submissions") {
		s.handleProviderSchemaSubmissions(w, r, providerID)
		return
	}

	// Check if this is a specific schema submission (e.g., /providers/:id/schema-submissions/:schemaId)
	if strings.Contains(path, "/schema-submissions/") {
		schemaID := utils.ExtractIDFromPathString(strings.TrimPrefix(path, "/providers/"+providerID+"/schema-submissions/"))
		s.handleProviderSchemaSubmissionByID(w, r, providerID, schemaID)
		return
	}

	// Check if this is a schema submission action (e.g., /providers/:id/schema-submissions/:schemaId/submit)
	if strings.Contains(path, "/schema-submissions/") && strings.HasSuffix(path, "/submit") {
		schemaID := utils.ExtractIDFromPathString(strings.TrimPrefix(strings.TrimSuffix(path, "/submit"), "/providers/"+providerID+"/schema-submissions/"))
		s.handleSubmitSchemaForReview(w, r, providerID, schemaID)
		return
	}

	// Handle other provider sub-resources here in the future
	utils.RespondWithError(w, http.StatusNotFound, "Resource not found")
}

// handleProviderSchemaSubmissions handles /providers/:provider-id/schema-submissions
func (s *APIServer) handleProviderSchemaSubmissions(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodPost:
		// POST /providers/:provider-id/schema-submissions - Create new schema submission or modify existing
		var req models.CreateProviderSchemaSubmissionRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		schema, err := s.providerService.CreateProviderSchemaSubmission(providerID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, schema)
	case http.MethodGet:
		// GET /providers/:provider-id/schema-submissions - List all schema submissions for provider
		schemas, err := s.providerService.GetProviderSchemasByProviderID(providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(schemas, len(schemas)))
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleProviderSchemaSubmissionByID handles /providers/:provider-id/schema-submissions/:schemaId
func (s *APIServer) handleProviderSchemaSubmissionByID(w http.ResponseWriter, r *http.Request, providerID, schemaID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers/:provider-id/schema-submissions/:schemaId - Get specific schema
		schema, err := s.providerService.GetProviderSchema(schemaID)

		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Verify the schema belongs to the provider
		if schema.ProviderID != providerID {
			utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, schema)
	case http.MethodPut:
		// PUT /providers/:provider-id/schema-submissions/:schemaId - Update schema (admin approval)
		var req models.UpdateProviderSchemaRequest
		if err := utils.ParseJSONRequest(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		schema, err := s.providerService.UpdateProviderSchema(schemaID, req)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Verify the schema belongs to the provider
		if schema.ProviderID != providerID {
			utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, schema)
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
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

// handleSubmitSchemaForReview handles POST /providers/:provider-id/schema-submissions/:schemaId/submit
func (s *APIServer) handleSubmitSchemaForReview(w http.ResponseWriter, r *http.Request, providerID, schemaID string) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Verify the schema belongs to the provider
	schema, err := s.providerService.GetProviderSchema(schemaID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Schema not found")
		return
	}

	if schema.ProviderID != providerID {
		utils.RespondWithError(w, http.StatusNotFound, "Schema not found for this provider")
		return
	}

	// Submit schema for review
	updatedSchema, err := s.providerService.SubmitSchemaForReview(schemaID)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, updatedSchema)
}

// handleProviderSchemas handles /providers/:provider-id/schemas
func (s *APIServer) handleProviderSchemas(w http.ResponseWriter, r *http.Request, providerID string) {
	switch r.Method {
	case http.MethodGet:
		// GET /providers/:provider-id/schemas - List approved schemas for provider
		schemas, err := s.providerService.GetApprovedSchemasByProviderID(providerID)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.RespondWithSuccess(w, http.StatusOK, utils.CreateCollectionResponse(schemas, len(schemas)))
	default:
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
