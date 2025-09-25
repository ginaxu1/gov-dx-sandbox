package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/auth"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
)

type Response struct {
	Message string `json:"message"`
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

	mux.HandleFunc("/public/sdl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if configs.AppConfig == nil || configs.AppConfig.Sdl == nil {
			http.Error(w, "SDL not configured", http.StatusInternalServerError)
			return
		}

		sdl := configs.AppConfig.Sdl
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"sdl": string(sdl)}

		err := json.NewEncoder(w).Encode(response)

		if err != nil {
			logger.Log.Error("Failed to write SDL response", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

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

		// decode the token
		consumerAssertion, err := auth.GetConsumerJwtFromToken(r)

		if err != nil {
			logger.Log.Error("Failed to get consumer JWT from token", "error", err)
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		response, statusCode := f.FederateQuery(req, consumerAssertion)

		w.WriteHeader(statusCode)
		// Set content type to application/json

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(response)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight (OPTIONS) requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
