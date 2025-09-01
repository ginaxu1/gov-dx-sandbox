package main

import (
	"bytes"
	"context"
	"encoding/json"
	"gov-dx-sandbox/exchange/internal/httputil"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/open-policy-agent/opa/sdk"
)

const (
	serverPort = ":8080"
	policyFile = "main.rego" // Path to the local policy file
)

// opaInstance holds the prepared OPA engine, initialized once at startup
var opaInstance *sdk.OPA

func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("handler panic recovered", "error", err, "stack", string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// policyDecisionHandler now uses the in-process OPA engine
func policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", "method", r.Method, "path", r.URL.Path)

	// Unmarshal the request body directly into a generic map for OPA input
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Make the decision by calling the OPA SDK directly
	// This replaces all the previous http client and retry logic
	decision, err := opaInstance.Decision(r.Context(), sdk.DecisionOptions{
		Path:  "/opendif/authz/decision", // The path to the rule to evaluate
		Input: input,
	})

	if err != nil {
		slog.Error("Failed to get OPA decision", "error", err)
		http.Error(w, "Failed to evaluate policy", http.StatusInternalServerError)
		return
	}

	// Send the result back to the client
	httputil.RespondWithJSON(w, http.StatusOK, decision.Result)
	slog.Info("Decision sent", "method", r.Method, "path", r.URL.Path)
}

func main() {
	// Setup structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ctx := context.Background()

	// Read the policy file from disk
	policyBytes, err := os.ReadFile(policyFile)
	if err != nil {
		slog.Error("Failed to read policy file", "path", policyFile, "error", err)
		os.Exit(1)
	}

	// Create a new OPA instance with the policy
	// This loads, parses, and compiles the policy, preparing it for queries
	opaInstance, err = sdk.New(ctx, sdk.Options{
		ID:     "policy-decision-point-sdk",
		Config: bytes.NewReader(policyBytes),
	})
	if err != nil {
		slog.Error("Failed to initialize OPA SDK", "error", err)
		os.Exit(1)
	}
	// The defer will ensure the OPA instance is cleaned up on shutdown
	defer opaInstance.Stop(ctx)

	slog.Info("OPA engine initialized successfully with local policy", "policy_file", policyFile)

	http.Handle("/decide", panicRecoveryMiddleware(http.HandlerFunc(policyDecisionHandler)))

	slog.Info("PDP server starting", "port", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
