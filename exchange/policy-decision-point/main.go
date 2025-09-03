package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gov-dx-sandbox/exchange/utils"
	"github.com/open-policy-agent/opa/rego"
)

// Constants for configuration
const (
	defaultPort = "8080"
)

// PolicyEvaluator holds the prepared OPA query, ready for evaluation
type PolicyEvaluator struct {
	preparedQuery rego.PreparedEvalQuery
}

// NewPolicyEvaluator creates and initializes a new evaluator by loading policies from disk
func NewPolicyEvaluator(ctx context.Context) (*PolicyEvaluator, error) {
	slog.Info("Loading OPA policies and data...")

	query := "data.opendif.authz.decision"

	r := rego.New(
		rego.Query(query),
		rego.Load([]string{"./policies", "./data"}, nil), // Load all .rego and .json files
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		// Wrapping the error provides more context
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	slog.Info("OPA policies and data loaded successfully")
	return &PolicyEvaluator{preparedQuery: pq}, nil
}

// Authorize evaluates the given input against the loaded policy
func (p *PolicyEvaluator) Authorize(ctx context.Context, input interface{}) (interface{}, error) {
	results, err := p.preparedQuery.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	if len(results) == 0 {
		slog.Warn("Policy returned no results for the input")
		// Return a predictable empty object for an undefined result
		return map[string]interface{}{}, nil
	}

	return results[0].Expressions[0].Value, nil
}

// policyDecisionHandler is an HTTP handler that uses the Authorize method to make decisions
func (p *PolicyEvaluator) policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
		return
	}

	var input interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid JSON input"})
		return
	}

	// Delegate the core logic to the Authorize method
	decision, err := p.Authorize(r.Context(), input)
	if err != nil {
		slog.Error("Policy authorization failed", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to evaluate policy"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, decision)
	slog.Info("Decision sent", "method", r.Method, "path", r.URL.Path, "decision", decision)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// Create an instance of our evaluator
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		slog.Error("Could not initialize policy evaluator", "error", err)
		os.Exit(1)
	}

	// Register the handler method from our evaluator instance
	http.Handle("/decide", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.policyDecisionHandler)))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	listenAddr := fmt.Sprintf(":%s", port)

	slog.Info("PCE server starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
