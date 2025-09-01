package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const serverPort = ":8081"

// ConsentRecord stores the details and state of a consent request
type ConsentRecord struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"` // "pending", "approved", "denied"
	CreatedAt    time.Time `json:"created_at"`
	DataConsumer string    `json:"data_consumer"`
	DataOwner    string    `json:"data_owner"`
	Fields       []string  `json:"fields"`
}

// CreateConsentRequest defines the structure for the incoming request body to create a consent record
type CreateConsentRequest struct {
	DataConsumer string   `json:"data_consumer"`
	DataOwner    string   `json:"data_owner"`
	Fields       []string `json:"fields"`
}

// In-memory store for consent records.
var (
	consentRecords = make(map[string]*ConsentRecord)
	lock           = sync.RWMutex{}
)

// consentHandler manages creating and retrieving consent records
// It routes requests based on the HTTP method to the appropriate logic
func consentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createConsent(w, r)
	case http.MethodGet:
		getConsentStatus(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// respondWithJSON is a utility function to write a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR: Failed to encode JSON response: %v", err)
	}
}

// createConsent handles the creation of a new consent record
func createConsent(w http.ResponseWriter, r *http.Request) {
	var req CreateConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	record := &ConsentRecord{
		ID:           uuid.New().String(),
		Status:       "pending",
		CreatedAt:    time.Now(),
		DataConsumer: req.DataConsumer,
		DataOwner:    req.DataOwner,
		Fields:       req.Fields,
	}
	consentRecords[record.ID] = record

	log.Printf("Created new consent record: %s for owner %s", record.ID, record.DataOwner)
	respondWithJSON(w, http.StatusCreated, record)
}

// getConsentStatus handles retrieving the status of a specific consent record
func getConsentStatus(w http.ResponseWriter, r *http.Request) {
	// Expecting a URL like /consent/{id}
	id := strings.TrimPrefix(r.URL.Path, "/consent/")
	if id == "" {
		http.Error(w, "Consent ID is required", http.StatusBadRequest)
		return
	}

	lock.RLock()
	defer lock.RUnlock()

	record, ok := consentRecords[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	respondWithJSON(w, http.StatusOK, record)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// A single handler for the /consent/ resource, which routes by HTTP method
	http.HandleFunc("/consent/", consentHandler)

	log.Printf("CME server starting on port %s", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatalf("FATAL: Could not start CME server: %v", err)
	}
}
