package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDataExchangeEvent_ToResponse(t *testing.T) {
	now := time.Now()
	event := &DataExchangeEvent{
		ID:                "id-1",
		Timestamp:         now,
		Status:            "success",
		ApplicationID:     "app-1",
		SchemaID:          "schema-1",
		RequestedData:     json.RawMessage(`"data"`),
		OnBehalfOfOwnerID: strPtr("owner-1"),
		ConsumerID:        strPtr("consumer-1"),
		ProviderID:        strPtr("provider-1"),
		AdditionalInfo:    json.RawMessage(`"info"`),
		CreatedAt:         now,
	}

	resp := event.ToResponse()

	assert.Equal(t, event.ID, resp.ID)
	assert.Equal(t, event.Timestamp.Format(time.RFC3339), resp.Timestamp)
	assert.Equal(t, event.Status, resp.Status)
	assert.Equal(t, event.ApplicationID, resp.ApplicationID)
	assert.Equal(t, event.SchemaID, resp.SchemaID)
	assert.Equal(t, event.RequestedData, resp.RequestedData)
	assert.Equal(t, event.OnBehalfOfOwnerID, resp.OnBehalfOfOwnerID)
	assert.Equal(t, event.ConsumerID, resp.ConsumerID)
	assert.Equal(t, event.ProviderID, resp.ProviderID)
	assert.Equal(t, event.AdditionalInfo, resp.AdditionalInfo)
}

func strPtr(s string) *string {
	return &s
}


func TestMetadata_Value_Scan(t *testing.T) {
	m := Metadata{"key": "value", "number": 123.0}
	
	// Test Value
	val, err := m.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)
	
	// Test Scan
	var m2 Metadata
	err = m2.Scan(val)
	assert.NoError(t, err)
	assert.Equal(t, m["key"], m2["key"])
	assert.Equal(t, m["number"], m2["number"])
	
	// Test Scan with string
	jsonStr := `{"key": "value", "number": 123.0}`
	var m3 Metadata
	err = m3.Scan(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, m["key"], m3["key"])
	
	// Test Scan with nil
	var m4 Metadata
	err = m4.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, m4)
}

