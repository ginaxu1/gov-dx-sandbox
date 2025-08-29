package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Message string `json:"message"`
}

func RunServer() {
	// /health route
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
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
		port = ":8080"
	}
	fmt.Printf("OpenDIF server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
