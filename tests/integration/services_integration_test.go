package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	portalBackendURL = "http://127.0.0.1:3000"
)

func TestPortalBackend_Health(t *testing.T) {
	resp, err := http.Get(portalBackendURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}




