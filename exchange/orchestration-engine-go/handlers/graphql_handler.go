package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQLHandler handles GraphQL requests
type GraphQLHandler struct {
	GraphQLService services.GraphQLService
}

// NewGraphQLHandler creates a new GraphQL handler
func NewGraphQLHandler(graphQLService services.GraphQLService) *GraphQLHandler {
	return &GraphQLHandler{
		GraphQLService: graphQLService,
	}
}

// HandleGraphQL handles GraphQL queries with version support
func (h *GraphQLHandler) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	var req models.GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract version from X-Schema-Version header (primary method)
	version := r.Header.Get("X-Schema-Version")

	// Fallback to query parameter if header not present
	if version == "" {
		version = r.URL.Query().Get("version")
	}

	// Route to appropriate schema
	schema, err := h.GraphQLService.RouteQuery(req.Query, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process GraphQL query with selected schema
	result, err := h.ProcessGraphQLQuery(req.Query, schema)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add schema version to response headers for debugging
	h.SetResponseHeaders(w, version)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ExtractVersionFromRequest extracts the schema version from request headers or query params
func (h *GraphQLHandler) ExtractVersionFromRequest(r *http.Request) string {
	// Check X-Schema-Version header first
	version := r.Header.Get("X-Schema-Version")
	if version != "" {
		return version
	}

	// Fallback to query parameter
	return r.URL.Query().Get("version")
}

// ValidateGraphQLRequest validates a GraphQL request
func (h *GraphQLHandler) ValidateGraphQLRequest(req *models.GraphQLRequest) error {
	if req.Query == "" {
		return &models.ValidationError{Field: "query", Message: "query is required"}
	}

	// Basic query validation
	if len(req.Query) < 10 {
		return &models.ValidationError{Field: "query", Message: "query is too short"}
	}

	return nil
}

// ProcessGraphQLQuery processes a GraphQL query with the given schema
func (h *GraphQLHandler) ProcessGraphQLQuery(query string, schema *ast.QueryDocument) (interface{}, error) {
	return h.GraphQLService.ProcessQuery(query, schema)
}

// SetResponseHeaders sets appropriate response headers
func (h *GraphQLHandler) SetResponseHeaders(w http.ResponseWriter, version string) {
	if version != "" {
		w.Header().Set("X-Schema-Version-Used", version)
	}
	w.Header().Set("Content-Type", "application/json")
}

// CheckVersionCompatibility checks if a version string is valid
func (h *GraphQLHandler) CheckVersionCompatibility(version string) error {
	if version == "" {
		return nil // Empty version is allowed (uses default)
	}

	// Basic semantic version validation
	versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !versionRegex.MatchString(version) {
		return &models.ValidationError{Field: "version", Message: "invalid version format"}
	}

	return nil
}

// HandleGraphQLIntrospection handles GraphQL introspection queries
func (h *GraphQLHandler) HandleGraphQLIntrospection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract version from headers or query parameters
	version := r.Header.Get("X-Schema-Version")
	if version == "" {
		version = r.URL.Query().Get("version")
	}

	// Route to appropriate schema
	schema, err := h.GraphQLService.RouteQuery(req.Query, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process introspection query
	result, err := h.processIntrospectionQuery(req, schema)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.SetResponseHeaders(w, version)
	json.NewEncoder(w).Encode(result)
}

// GetSchemaInfo returns information about the current schema
func (h *GraphQLHandler) GetSchemaInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active schema version - simplified for now
	version := "1.0.0"

	// Get schema versions - simplified for now
	versions := map[string]interface{}{
		"1.0.0": "active",
	}

	info := map[string]interface{}{
		"active_version":  version,
		"loaded_versions": versions,
		"total_versions":  len(versions),
	}

	h.SetResponseHeaders(w, version)
	json.NewEncoder(w).Encode(info)
}

// processIntrospectionQuery processes GraphQL introspection queries
func (h *GraphQLHandler) processIntrospectionQuery(req models.GraphQLRequest, schema interface{}) (interface{}, error) {
	// This is a placeholder implementation
	// In a real implementation, you would process the introspection query
	return map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"name": "Query",
				},
			},
		},
	}, nil
}
