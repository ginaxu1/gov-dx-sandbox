package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConsentStatus defines the possible states of a consent record
type ConsentStatus string

const (
	// StatusPending indicates that the consent request has been created but not yet actioned
	StatusPending ConsentStatus = "pending"
	// StatusApproved indicates that the data owner has approved the consent request
	StatusApproved ConsentStatus = "approved"
	// StatusDenied indicates that the data owner has denied the consent request
	StatusDenied ConsentStatus = "denied"
)

// ConsentRecord stores the details and state of a consent request
// It represents a formal agreement from the data owner to the data consumer
type ConsentRecord struct {
	// ID is the unique identifier for the consent record
	ID string `json:"id"`
	// Status indicates the current state of the consent
	Status ConsentStatus `json:"status"`
	// CreatedAt is the timestamp when the consent record was initially created
	CreatedAt time.Time `json:"created_at"`
	// DataConsumer identifies the entity requesting access to the data
	// It's currently a string but could be a URI or another structured identifier
	DataConsumer string `json:"data_consumer"`
	// DataOwner identifies the user or entity granting consent
	// It's currently a string but could be a URI or another structured identifier
	DataOwner string `json:"data_owner"`
	// Fields is a list of specific data fields the consent applies to
	// These are string names that correspond to fields in the data source
	Fields []string `json:"fields"`
}

// CreateConsentRequest defines the structure for creating a consent record
type CreateConsentRequest struct {
	// DataConsumer identifies the entity requesting access to the data
	DataConsumer string `json:"data_consumer"`
	// DataOwner identifies the user or entity that owns the data and can grant consent
	DataOwner string `json:"data_owner"`
	// Fields is the list of specific data fields for which access is being requested
	Fields []string `json:"fields"`
}

// ConsentEngine defines the interface for managing consent records
type ConsentEngine interface {
	// CreateConsent creates a new consent record and stores it
	CreateConsent(req CreateConsentRequest) (*ConsentRecord, error)
	// GetConsentStatus retrieves a consent record by its ID
	GetConsentStatus(id string) (*ConsentRecord, error)
}

// consentEngineImpl is the private, in-memory implementation of the ConsentEngine interface
type consentEngineImpl struct {
	consentRecords map[string]*ConsentRecord
	lock           sync.RWMutex
}

// NewConsentEngine creates a new in-memory instance of the ConsentEngine
func NewConsentEngine() ConsentEngine {
	return &consentEngineImpl{
		consentRecords: make(map[string]*ConsentRecord),
	}
}

// CreateConsent creates a new consent record and saves it in the in-memory store
func (ce *consentEngineImpl) CreateConsent(req CreateConsentRequest) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	record := &ConsentRecord{
		ID:           uuid.New().String(),
		Status:       StatusPending,
		CreatedAt:    time.Now(),
		DataConsumer: req.DataConsumer,
		DataOwner:    req.DataOwner,
		Fields:       req.Fields,
	}
	ce.consentRecords[record.ID] = record

	return record, nil
}

// GetConsentStatus retrieves a specific consent record from the in-memory store
func (ce *consentEngineImpl) GetConsentStatus(id string) (*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	record, ok := ce.consentRecords[id]
	if !ok {
		return nil, fmt.Errorf("consent record with ID '%s' not found", id)
	}
	return record, nil
}
