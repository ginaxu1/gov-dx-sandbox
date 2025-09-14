package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
)

type Response struct {
	Message string `json:"message"`
}

type PolicyDecisionResponse struct {
	Allow                 bool     `json:"allow"`
	ConsentRequired       bool     `json:"consent_required"`
	ConsentRequiredFields []string `json:"consent_required_fields"`
}

const DefaultPort = "4000"

// RunServer starts a simple HTTP server with a health check endpoint.
func RunServer(f *federator.Federator) {
	mux := http.NewServeMux()
	// /health route
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := Response{Message: "OpenDIF Server is Healthy!"}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return
		}
	})

	// Handle GET /getData endpoint
	mux.HandleFunc("/getData", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters for GET request
		query := r.URL.Query().Get("query")
		variables := r.URL.Query().Get("variables")

		if query == "" {
			http.Error(w, "query parameter is required", http.StatusBadRequest)
			return
		}

		// Parse variables if provided
		var variablesMap map[string]interface{}
		if variables != "" {
			if err := json.Unmarshal([]byte(variables), &variablesMap); err != nil {
				http.Error(w, "invalid variables parameter: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		req := graphql.Request{
			Query:     query,
			Variables: variablesMap,
		}

		// Call policy decision point first
		policyDecision, err := callPolicyDecisionPoint(req)
		if err != nil {
			logger.Log.Error("Failed to call policy decision point", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If consent is required, initiate consent workflow
		if policyDecision.ConsentRequired {
			logger.Log.Info("Consent required, initiating consent workflow", "fields", policyDecision.ConsentRequiredFields)
			// Create consent request
			consentReq := map[string]interface{}{
				"app_id": "passport-app",
				"data_fields": []map[string]interface{}{
					{
						"owner_type": "citizen",
						"owner_id":   "199512345678", // This should come from the request
						"fields":     policyDecision.ConsentRequiredFields,
					},
				},
				"purpose":      "passport_application",
				"session_id":   fmt.Sprintf("session_%d", time.Now().Unix()),
				"redirect_url": "http://localhost:3000/apply", // Redirect back to passport app
			}

			// Call consent engine to create consent records
			consentResp, err := callConsentEngine(consentReq)
			if err != nil {
				logger.Log.Error("Failed to call consent engine", "error", err)
				http.Error(w, "Failed to initiate consent workflow", http.StatusInternalServerError)
				return
			}

			// Debug: log the consent response
			logger.Log.Info("Consent engine response", "response", consentResp)

			// Return consent portal URL for redirect
			response := map[string]interface{}{
				"allow":                 policyDecision.Allow,
				"consentRequired":       true,
				"consentRequiredFields": policyDecision.ConsentRequiredFields,
				"message":               "Consent is required to access this data",
				"consentPortalUrl":      consentResp["redirect_url"],
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// If no consent required, proceed with federation
		response := f.FederateQuery(req)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.Log.Error("Failed to write response", "error", err)
			return
		}
	})

	// Handle consent completion and data fetching
	mux.HandleFunc("/consent-complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req struct {
			ConsentID string `json:"consent_id"`
			Query     string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Check consent status
		consentStatus, err := callConsentEngineStatus(req.ConsentID)
		if err != nil {
			logger.Log.Error("Failed to get consent status", "error", err)
			http.Error(w, "Failed to check consent status", http.StatusInternalServerError)
			return
		}

		status, ok := consentStatus["status"].(string)
		if !ok {
			http.Error(w, "Invalid consent status", http.StatusInternalServerError)
			return
		}

		// If consent is not approved, return rejection
		if status != "approved" {
			response := map[string]interface{}{
				"success": false,
				"message": "Consent was rejected or not approved",
				"status":  status,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// If approved, proceed with data federation
		graphqlReq := graphql.Request{
			Query: req.Query,
		}
		response := f.FederateQuery(graphqlReq)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Handle other GraphQL requests
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		var req graphql.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		response := f.FederateQuery(req)

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.Log.Error("Failed to write response", "error", err)
			return
		}
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Convert port to string with colon prefix
	// e.g., "8000" -> ":8000"
	// This is needed for http.ListenAndServe
	// which expects the port in the format ":port"
	// If the port already has a colon, we don't add another one
	if port[0] != ':' {
		port = ":" + port
	}

	logger.Log.Info("Server is Listening", "port", port)

	if err := http.ListenAndServe(port, corsMiddleware(mux)); err != nil {
		logger.Log.Error("Failed to start server", "error", err)
	} else {
		logger.Log.Info("Server stopped")
	}
}

// corsMiddleware sets CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow specific methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// callPolicyDecisionPoint calls the policy decision point to determine if consent is required
func callPolicyDecisionPoint(req graphql.Request) (*PolicyDecisionResponse, error) {
	// Extract required fields from the GraphQL query
	requiredFields := extractRequiredFieldsFromQuery(req.Query)

	// Create policy decision request in the format expected by the policy decision point
	policyRequest := map[string]interface{}{
		"consumer_id":     "passport-app",
		"app_id":          "passport-app",
		"request_id":      fmt.Sprintf("req_%d", time.Now().Unix()),
		"required_fields": requiredFields,
		"timestamp":       time.Now(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(policyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal policy request: %v", err)
	}

	// Call policy decision point
	resp, err := http.Post("http://localhost:8082/decide", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call policy decision point: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("policy decision point returned status %d", resp.StatusCode)
	}

	// Parse response
	var policyResponse PolicyDecisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&policyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode policy response: %v", err)
	}

	return &policyResponse, nil
}

// extractRequiredFieldsFromQuery extracts the required fields from a GraphQL query
func extractRequiredFieldsFromQuery(query string) []string {
	// This is a simplified extraction - in a real implementation, you'd use a GraphQL parser
	// For now, we'll return common fields that might require consent
	fields := []string{}

	// Check for specific fields that might require consent
	if strings.Contains(query, "permanentAddress") {
		fields = append(fields, "person.permanentAddress")
	}
	if strings.Contains(query, "birthDate") {
		fields = append(fields, "person.birthDate")
	}
	if strings.Contains(query, "photo") {
		fields = append(fields, "person.photo")
	}

	// If no specific fields found, return a default set
	if len(fields) == 0 {
		fields = []string{"person.permanentAddress", "person.birthDate"}
	}

	return fields
}

// callConsentEngine calls the consent engine to create consent records
func callConsentEngine(req map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consent request: %v", err)
	}

	resp, err := http.Post("http://localhost:8081/consents", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call consent engine: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consent engine returned status %d", resp.StatusCode)
	}

	var consentResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&consentResp); err != nil {
		return nil, fmt.Errorf("failed to decode consent response: %v", err)
	}

	return consentResp, nil
}

// callConsentEngineStatus checks the status of a consent record
func callConsentEngineStatus(consentId string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8081/consents/%s", consentId))
	if err != nil {
		return nil, fmt.Errorf("failed to call consent engine: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consent engine returned status %d", resp.StatusCode)
	}

	var consentResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&consentResp); err != nil {
		return nil, fmt.Errorf("failed to decode consent response: %v", err)
	}

	return consentResp, nil
}
