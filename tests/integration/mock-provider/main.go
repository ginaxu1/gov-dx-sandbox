package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"person": map[string]interface{}{
					"email":      "test@example.com",
					"fullName":   "John Doe",
					"address":    "123 Test St",
					"profession": "Tester",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	})

	log.Println("Mock Provider listening on :8083")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal(err)
	}
}
