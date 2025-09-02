package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/utils"
)

// Constants for configuration
const (
	defaultPort       = "8080"
	httpClientTimeout = 10 * time.Second
	maxHTTPRetries    = 5
)

var opaURL string

// OPAInput uses json.RawMessage to pass the input through without needing to know its structure
type OPAInput struct {
	Input json.RawMessage `json:"input"`
}

// OPAOutput uses interface{} as we only care about forwarding the result, not its contents
type OPAOutput struct {
	Result interface{} `json:"result"`
}

// policyDecisionHandler forwards a request to OPA for a decision
func policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	// Enforce that only POST is allowed
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		utils.RespondWithJSON(w, http.StatusMethodNotAllowed, utils.ErrorResponse{Error: "Method not allowed"})
		return
	}

	slog.Info("Request received", "method", r.Method, "path", r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to read request body"})
		return
	}
	defer r.Body.Close()

	opaInputBytes, err := json.Marshal(OPAInput{Input: body})
	if err != nil {
		slog.Error("Failed to marshal OPA input", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to create OPA input"})
		return
	}

	var opaResponse *http.Response
	retryDelay := 1 * time.Second

	for i := 0; i < maxHTTPRetries; i++ {
		req, err := http.NewRequest("POST", opaURL, bytes.NewBuffer(opaInputBytes))
		if err != nil {
			slog.Error("Failed to create OPA request", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Internal server error"})
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: httpClientTimeout}
		opaResponse, err = client.Do(req)
		if err == nil && opaResponse.StatusCode == http.StatusOK {
			break // Success
		}

		if i < maxHTTPRetries-1 {
			status := "no response"
			if opaResponse != nil {
				status = opaResponse.Status
			}
			slog.Warn("OPA request failed, retrying...",
				"attempt", i+1, "max_attempts", maxHTTPRetries, "status", status,
				"error", err, "retry_delay", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if opaResponse == nil || opaResponse.StatusCode != http.StatusOK {
		slog.Error("Failed to get successful response from OPA after retries", "retries", maxHTTPRetries, "error", err)
		utils.RespondWithJSON(w, http.StatusServiceUnavailable, utils.ErrorResponse{Error: "Policy decision point is unavailable"})
		return
	}
	defer opaResponse.Body.Close()

	responseBody, err := io.ReadAll(opaResponse.Body)
	if err != nil {
		slog.Error("Failed to read OPA response body", "error", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to read OPA response"})
		return
	}

	var opaOutput OPAOutput
	if err := json.Unmarshal(responseBody, &opaOutput); err != nil {
		slog.Error("Failed to unmarshal OPA output", "error", err, "raw_response", string(responseBody))
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{Error: "Failed to parse OPA response"})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, opaOutput.Result)
	slog.Info("Decision sent", "method", r.Method, "path", r.URL.Path)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Fail fast if the OPA_URL is not set
	opaURL = os.Getenv("OPA_URL")
	if opaURL == "" {
		slog.Error("CRITICAL: OPA_URL environment variable must be set.")
		os.Exit(1) // Exit with a non-zero code to indicate failure
	}
	slog.Info("Using OPA URL", "url", opaURL)

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
