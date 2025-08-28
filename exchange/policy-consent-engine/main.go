package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

var opaURL = "http://localhost:8181/v1/data/opendif/authz/decision"

// OPAInput represents the structure of the request body sent to OPA
type OPAInput struct {
	Input interface{} `json:"input"`
}

// OPAOutput represents the structure of the response body received from OPA
type OPAOutput struct {
	Result interface{} `json:"result"`
}

// policyDecisionHandler is the main HTTP handler. It receives a request,
// forwards it to OPA for a decision, and returns the decision to the client
func policyDecisionHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("!PANIC! recovered in handler: %v\nStack: %s", r, debug.Stack())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	opaInputBytes, err := json.Marshal(OPAInput{Input: json.RawMessage(body)})
	if err != nil {
		log.Printf("ERROR: Failed to marshal OPA input: %v", err)
		http.Error(w, "Failed to create OPA input", http.StatusInternalServerError)
		return
	}

	var opaResponse *http.Response
	maxRetries := 5
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", opaURL, bytes.NewBuffer(opaInputBytes))
		if err != nil {
			log.Printf("ERROR: Failed to create OPA request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		opaResponse, err = client.Do(req)
		if err == nil && opaResponse.StatusCode == http.StatusOK {
			break
		}

		if i < maxRetries-1 {
			status := "no response"
			if opaResponse != nil {
				status = opaResponse.Status
			}
			log.Printf("WARN: OPA request failed (attempt %d/%d), status: %s, error: %v. Retrying in %s...", i+1, maxRetries, status, err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if opaResponse == nil || opaResponse.StatusCode != http.StatusOK {
		log.Printf("ERROR: Failed to get successful response from OPA after %d retries. Last error: %v", maxRetries, err)
		http.Error(w, "Policy decision point is unavailable", http.StatusServiceUnavailable)
		return
	}
	defer opaResponse.Body.Close()

	responseBody, err := io.ReadAll(opaResponse.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read OPA response body: %v", err)
		http.Error(w, "Failed to read OPA response", http.StatusInternalServerError)
		return
	}

	var opaOutput OPAOutput
	if err := json.Unmarshal(responseBody, &opaOutput); err != nil {
		log.Printf("ERROR: Failed to unmarshal OPA output: %v. Raw response: %s", err, string(responseBody))
		http.Error(w, "Failed to parse OPA response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(opaOutput.Result); err != nil {
		log.Printf("ERROR: Failed to encode response for client: %v", err)
	}
	log.Printf("Decision sent for %s %s", r.Method, r.URL.Path)
}

// main initializes the server, sets up logging, configures the OPA URL
// and starts listening for HTTP requests
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if envURL := os.Getenv("OPA_URL"); envURL != "" {
		opaURL = envURL
	}
	log.Printf("Using OPA URL: %s", opaURL)
	http.HandleFunc("/decide", policyDecisionHandler)

	port := ":8080"
	log.Printf("PCE server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("FATAL: Could not start server: %v", err)
	}
}
