package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

// Constants for configuration
const (
	serverPort        = ":8080"
	httpClientTimeout = 10 * time.Second
)

var opaURL = "http://localhost:8181/v1/data/opendif/authz/decision"

// OPAInput uses json.RawMessage to pass the input through without needing to know its structure
type OPAInput struct {
	Input json.RawMessage `json:"input"`
}

// OPAOutput uses interface{} as we only care about forwarding the result, not its contents
type OPAOutput struct {
	Result interface{} `json:"result"`
}

// panicRecoveryMiddleware is an HTTP middleware that recovers from panics
// It logs the error and stack trace, then returns a 500 Internal Server Error
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

// policyDecisionHandler forwards a request to OPA for a decision
func policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", "method", r.Method, "path", r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Use json.RawMessage directly, avoiding an unnecessary unmarshal/marshal cycle
	opaInputBytes, err := json.Marshal(OPAInput{Input: body})
	if err != nil {
		slog.Error("Failed to marshal OPA input", "error", err)
		http.Error(w, "Failed to create OPA input", http.StatusInternalServerError)
		return
	}

	var opaResponse *http.Response
	maxRetries := 5
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", opaURL, bytes.NewBuffer(opaInputBytes))
		if err != nil {
			slog.Error("Failed to create OPA request", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Use the httpClientTimeout constant
		client := &http.Client{Timeout: httpClientTimeout}
		opaResponse, err = client.Do(req)
		if err == nil && opaResponse.StatusCode == http.StatusOK {
			break // Success
		}

		if i < maxRetries-1 {
			status := "no response"
			if opaResponse != nil {
				status = opaResponse.Status
			}
			slog.Warn("OPA request failed, retrying...",
				"attempt", i+1,
				"max_attempts", maxRetries,
				"status", status,
				"error", err,
				"retry_delay", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if opaResponse == nil || opaResponse.StatusCode != http.StatusOK {
		slog.Error("Failed to get successful response from OPA after retries", "retries", maxRetries, "error", err)
		http.Error(w, "Policy decision point is unavailable", http.StatusServiceUnavailable)
		return
	}
	defer opaResponse.Body.Close()

	responseBody, err := io.ReadAll(opaResponse.Body)
	if err != nil {
		slog.Error("Failed to read OPA response body", "error", err)
		http.Error(w, "Failed to read OPA response", http.StatusInternalServerError)
		return
	}

	var opaOutput OPAOutput
	if err := json.Unmarshal(responseBody, &opaOutput); err != nil {
		slog.Error("Failed to unmarshal OPA output", "error", err, "raw_response", string(responseBody))
		http.Error(w, "Failed to parse OPA response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(opaOutput.Result); err != nil {
		slog.Error("Failed to encode response for client", "error", err)
	}
	slog.Info("Decision sent", "method", r.Method, "path", r.URL.Path)
}

func main() {
	// Setup structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if envURL := os.Getenv("OPA_URL"); envURL != "" {
		opaURL = envURL
	}
	slog.Info("Using OPA URL", "url", opaURL)

	// Wrap the handler with the panic recovery middleware
	http.Handle("/decide", panicRecoveryMiddleware(http.HandlerFunc(policyDecisionHandler)))

	slog.Info("PCE server starting", "port", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
