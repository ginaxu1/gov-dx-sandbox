package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConsentRequest stores the details and status of a consent request
type ConsentRequest struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"` // "pending", "approved", "denied"
	CreatedAt    time.Time `json:"created_at"`
	DataConsumer string    `json:"data_consumer"`
	DataOwner    string    `json:"data_owner"`
	Fields       []string  `json:"fields"`
}

// InitiateConsentPayload defines the structure for the incoming request body
type InitiateConsentPayload struct {
	DataConsumer string   `json:"data_consumer"`
	DataOwner    string   `json:"data_owner"`
	Fields       []string `json:"fields"`
}

// In-memory store for pending requests. TODO: Update to a database for production
var (
	requests = make(map[string]*ConsentRequest)
	lock     = sync.RWMutex{}
)

// initiateConsentHandler creates a new consent request
func initiateConsentHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the request body to get request details
	var payload InitiateConsentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	req := &ConsentRequest{
		ID:           uuid.New().String(),
		Status:       "pending",
		CreatedAt:    time.Now(),
		DataConsumer: payload.DataConsumer,
		DataOwner:    payload.DataOwner,
		Fields:       payload.Fields,
	}
	requests[req.ID] = req

	log.Printf("Initiated new consent request: %s for owner %s", req.ID, req.DataOwner)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(req)
}

// consentStatusHandler returns the status of a specific consent request
func consentStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/consent-status/"):]

	lock.RLock()
	defer lock.RUnlock()

	req, ok := requests[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/initiate-consent", initiateConsentHandler)
	http.HandleFunc("/consent-status/", consentStatusHandler)

	port := ":8081"
	log.Printf("CME server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("FATAL: Could not start CME server: %v", err)
	}
}
