package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsentRecord_TableName(t *testing.T) {
	cr := &ConsentRecord{}
	assert.Equal(t, "consent_records", cr.TableName())
}

