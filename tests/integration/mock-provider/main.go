package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Determine response based on query?
		// For now, just return a flat success response matching the schema.
		// Schema: Person { fullName, email, address, profession }
		// The orchestrator expects data in the key "mock-provider" if that's the service key,
		// OR it expects the structure "person { ... }" if the query sent to it is such.
		
		// The query sent to provider (from QueryBuilder in mapper.go) matches the source structure.
		// If Source info is "person.email", it sends "person { email }".
		
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
