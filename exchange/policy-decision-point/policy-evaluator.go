package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/constants"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
	"github.com/open-policy-agent/opa/rego"
)

// PolicyEvaluator holds the prepared OPA query, ready for evaluation.
type PolicyEvaluator struct {
	preparedQuery rego.PreparedEvalQuery
}

// Use shared constants instead of local ones

// Helper functions
// Use shared utility functions instead of custom ones

// AuthorizationRequest represents the input structure for policy evaluation
type AuthorizationRequest struct {
	Consumer  ConsumerInfo `json:"consumer"`
	Request   RequestInfo  `json:"request"`
	Timestamp time.Time    `json:"timestamp"`
}

// ConsumerInfo contains information about the requesting consumer
type ConsumerInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name,omitempty"`
	Type       string            `json:"type,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// RequestInfo contains details about the data access request
type RequestInfo struct {
	Resource   string   `json:"resource"`
	Action     string   `json:"action"`
	DataFields []string `json:"data_fields"`
	DataOwner  string   `json:"data_owner,omitempty"`
}

// AuthorizationDecision represents the output of policy evaluation
type AuthorizationDecision struct {
	Allow                 bool                   `json:"allow"`
	DenyReason            string                 `json:"deny_reason,omitempty"`
	ConsentRequired       bool                   `json:"consent_required"`
	ConsentRequiredFields []string               `json:"consent_required_fields,omitempty"`
	DataOwner             string                 `json:"data_owner,omitempty"`
	ExpiryTime            string                 `json:"expiry_time,omitempty"`
	Conditions            map[string]interface{} `json:"conditions,omitempty"`
}

// NewPolicyEvaluator creates and initializes a new evaluator by loading policies from disk.
func NewPolicyEvaluator(ctx context.Context) (*PolicyEvaluator, error) {
	slog.Info("Loading OPA policies and data...")

	query := "data.opendif.authz.decision"

	// Load data files explicitly
	consumerGrantsData, err := loadJSONFile("./data/consumer-grants.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer grants: %w", err)
	}
	slog.Info("Consumer grants data loaded", "data", consumerGrantsData)

	providerMetadataData, err := loadJSONFile("./data/provider-metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}
	slog.Info("Provider metadata data loaded", "data", providerMetadataData)

	// Convert data to JSON strings for embedding in policy
	consumerGrantsJSON, _ := json.Marshal(consumerGrantsData)
	providerMetadataJSON, _ := json.Marshal(providerMetadataData)

	// Create a module with the data embedded as JSON values
	dataModule := fmt.Sprintf(`
		package opendif.authz

		consumer_grants = %s
		provider_metadata = %s
		`, string(consumerGrantsJSON), string(providerMetadataJSON))

	r := rego.New(
		rego.Query(query),
		rego.Load([]string{"./policies"}, nil), // Load policy files
		rego.Module("data.rego", dataModule),   // Add data as module
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		// Wrapping the error provides more context
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	slog.Info("OPA policies and data loaded successfully")
	return &PolicyEvaluator{preparedQuery: pq}, nil
}

// loadJSONFile loads and parses a JSON file
func loadJSONFile(filepath string) (interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", filepath, err)
	}

	return result, nil
}

// Authorize evaluates the given input against the loaded policy using ABAC model
// It returns a structured AuthorizationDecision with detailed access control information
func (p *PolicyEvaluator) Authorize(ctx context.Context, input interface{}) (*AuthorizationDecision, error) {
	// Validate input structure
	authReq, err := p.validateInput(input)
	if err != nil {
		return &AuthorizationDecision{
			Allow:      false,
			DenyReason: fmt.Sprintf("Invalid input: %v", err),
		}, nil
	}

	// Add timestamp if not provided
	if authReq.Timestamp.IsZero() {
		authReq.Timestamp = time.Now()
	}

	results, err := p.preparedQuery.Eval(ctx, rego.EvalInput(authReq))
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	if len(results) == 0 {
		slog.Warn("Policy returned no results for the input")
		return &AuthorizationDecision{
			Allow:      false,
			DenyReason: "No policy rules matched the request",
		}, nil
	}

	// Convert OPA result to structured decision
	decision, err := p.convertToDecision(results[0].Expressions[0].Value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert policy result: %w", err)
	}

	slog.Info("Policy evaluation completed",
		"consumer", authReq.Consumer.ID,
		"resource", authReq.Request.Resource,
		"allow", decision.Allow,
		"consent_required", decision.ConsentRequired)

	return decision, nil
}

// validateInput validates and converts the input to AuthorizationRequest
func (p *PolicyEvaluator) validateInput(input interface{}) (*AuthorizationRequest, error) {
	// Convert to JSON and back to ensure proper structure
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	var authReq AuthorizationRequest
	if err := json.Unmarshal(jsonData, &authReq); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// Validate required fields
	if authReq.Consumer.ID == "" {
		return nil, fmt.Errorf(constants.ErrConsumerIDRequired)
	}
	if authReq.Request.Resource == "" {
		return nil, fmt.Errorf(constants.ErrResourceRequired)
	}
	if authReq.Request.Action == "" {
		return nil, fmt.Errorf(constants.ErrActionRequired)
	}
	if len(authReq.Request.DataFields) == 0 {
		return nil, fmt.Errorf(constants.ErrDataFieldsRequired)
	}

	return &authReq, nil
}

// convertToDecision converts OPA result to AuthorizationDecision
func (p *PolicyEvaluator) convertToDecision(result interface{}) (*AuthorizationDecision, error) {
	// Convert to JSON and back to ensure proper structure
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var decision AuthorizationDecision
	if err := json.Unmarshal(jsonData, &decision); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &decision, nil
}

// DebugData checks if data is loaded correctly
func (p *PolicyEvaluator) DebugData(ctx context.Context) (interface{}, error) {
	query := "data.opendif.authz.debug_data"

	// Load data files explicitly
	consumerGrantsData, err := loadJSONFile("./data/consumer-grants.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load consumer grants: %w", err)
	}

	providerMetadataData, err := loadJSONFile("./data/provider-metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load provider metadata: %w", err)
	}

	// Convert data to JSON strings for embedding in policy
	debugConsumerGrantsJSON, _ := json.Marshal(consumerGrantsData)
	debugProviderMetadataJSON, _ := json.Marshal(providerMetadataData)

	// Create a module with the data embedded as JSON values
	debugDataModule := fmt.Sprintf(`
package opendif.authz

consumer_grants = %s
provider_metadata = %s
`, string(debugConsumerGrantsJSON), string(debugProviderMetadataJSON))

	r := rego.New(
		rego.Query(query),
		rego.Load([]string{"./policies"}, nil),
		rego.Module("debug_data.rego", debugDataModule),
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare debug query: %w", err)
	}

	results, err := pq.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("debug query evaluation failed: %w", err)
	}

	if len(results) == 0 {
		return map[string]interface{}{"error": "no debug results"}, nil
	}

	return results[0].Expressions[0].Value, nil
}

// policyDecisionHandler is an HTTP handler that uses the Authorize method to make decisions.
func (p *PolicyEvaluator) policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	if !utils.ValidateMethod(w, r, http.MethodPost) {
		return
	}
	defer r.Body.Close()

	var input interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.HandleError(w, err, http.StatusBadRequest, "decode request body")
		return
	}

	// Delegate the core logic to the Authorize method
	decision, err := p.Authorize(r.Context(), input)
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError, constants.OpPolicyEvaluation)
		return
	}

	// Return appropriate HTTP status based on decision
	status := http.StatusOK
	if !decision.Allow {
		status = http.StatusForbidden
	}

	utils.HandleSuccess(w, decision, status, constants.OpDecisionSent, map[string]interface{}{
		"method":           r.Method,
		"path":             r.URL.Path,
		"allow":            decision.Allow,
		"consent_required": decision.ConsentRequired,
		"status":           status,
	})
}

// debugHandler is an HTTP handler for debugging data loading
func (p *PolicyEvaluator) debugHandler(w http.ResponseWriter, r *http.Request) {
	if !utils.ValidateMethod(w, r, http.MethodGet) {
		return
	}

	debugResult, err := p.DebugData(r.Context())
	if err != nil {
		utils.HandleError(w, err, http.StatusInternalServerError, constants.OpDebugData)
		return
	}

	utils.HandleSuccess(w, debugResult, http.StatusOK, constants.OpDebugData, map[string]interface{}{
		"result": debugResult,
	})
}
