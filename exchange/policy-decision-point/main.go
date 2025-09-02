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

var preparedQuery rego.PreparedEvalQuery

// policyDecisionHandler evaluates an input against the loaded policies
func policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
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
	defer r.Body.Close()

	// Evaluate the prepared query with the input from the request
	results, err := preparedQuery.Eval(context.Background(), rego.EvalInput(input))
	if err != nil {
		slog.Error("Failed to evaluate policy", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to evaluate policy"})
		return
	}

	if len(results) == 0 {
		// This can happen if the policy doesn't produce any result for the query
		// You might want to return a default deny or a specific error
		slog.Warn("Policy returned no results")
		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{}) // Return empty object
		return
	}

	// Send the first result's expressions back to the client.
	utils.RespondWithJSON(w, http.StatusOK, results[0].Expressions[0].Value)
	slog.Info("Decision sent", "method", r.Method, "path", r.URL.Path)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// This query tells OPA which rule to evaluate from your policy.=
	// We are querying for the result of the 'decision' rule within the 'opendif.authz' package
	query := "data.opendif.authz.decision"

	// Create a new Rego object that can be evaluated
	r := rego.New(
		rego.Query(query),
		rego.Load([]string{"./policies", "./data"}, nil), // Load all .rego and .json files from policies/ and data/
	)

	// Prepare the query for evaluation
	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		slog.Error("Failed to prepare OPA query", "error", err)
		os.Exit(1)
	}
	preparedQuery = pq

	slog.Info("OPA policies and data loaded successfully")

	http.Handle("/decide", utils.PanicRecoveryMiddleware(http.HandlerFunc(policyDecisionHandler)))

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
