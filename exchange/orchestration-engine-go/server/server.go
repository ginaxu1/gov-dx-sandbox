package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/logger"
)

type Response struct {
	Message string `json:"message"`
}

const DefaultPort = ":8000"

// RunServer starts a simple HTTP server with a health check endpoint.
func RunServer() {
	// /health route
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
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

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	logger.Log.Info(fmt.Sprintf("Listening on port %s", port))

	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Log.Error("Failed to start server: ", err.Error())
	}
}
