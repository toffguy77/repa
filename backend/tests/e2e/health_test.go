package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_OK(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/health", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, "ok", data["status"])
	assert.Equal(t, "ok", data["db"])
	assert.Equal(t, "ok", data["redis"])
}

func TestHealth_ReturnsDBAndRedisStatus(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/health", nil, "")
	data := getData(t, resp)
	_, hasDB := data["db"]
	_, hasRedis := data["redis"]
	assert.True(t, hasDB, "should have db status")
	assert.True(t, hasRedis, "should have redis status")
}
