package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin_NoAuth(t *testing.T) {
	resp := doRequestBasicAuth(t, "GET", "/api/v1/admin/stats", nil, "", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdmin_WrongCredentials(t *testing.T) {
	resp := doRequestBasicAuth(t, "GET", "/api/v1/admin/stats", nil, "wrong", "creds")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAdmin_GetStats(t *testing.T) {
	resp := doRequestBasicAuth(t, "GET", "/api/v1/admin/stats", nil, testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.NotNil(t, data["dau_7d"])
	assert.NotNil(t, data["mau_30d"])
	assert.NotNil(t, data["groups_count"])
	assert.NotNil(t, data["revenue_7d_rub"])
}

func TestAdmin_ListAndResolveReports(t *testing.T) {
	// Create a question and report it
	uid1, tok1 := createTestUser(t, "adm_q_u1_"+t.Name())
	_, tok2 := createTestUser(t, "adm_q_u2_"+t.Name())

	group := createTestGroup(t, tok1, "Admin Reports "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	_ = uid1

	// Create a question
	createResp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Bad question for admin test",
		"category": "FUNNY",
	}, tok1)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	qID := getData(t, createResp)["question"].(map[string]any)["id"].(string)

	// Report it
	doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions/"+qID+"/report", map[string]any{
		"reason": "test report",
	}, tok2)

	// List reports
	listResp := doRequestBasicAuth(t, "GET", "/api/v1/admin/reports", nil, testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	reports := listResp.Body["data"].([]any)
	require.GreaterOrEqual(t, len(reports), 1)

	// Find our report
	var reportID string
	for _, r := range reports {
		rep := r.(map[string]any)
		if rep["question_id"] == qID {
			reportID = rep["id"].(string)
			break
		}
	}
	require.NotEmpty(t, reportID, "should find our report")

	// Resolve report
	resolveResp := doRequestBasicAuth(t, "PATCH", "/api/v1/admin/reports/"+reportID, map[string]any{
		"action": "reject",
	}, testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, resolveResp.StatusCode)

	resolveData := getData(t, resolveResp)
	assert.Equal(t, "reject", resolveData["action"])
	assert.Equal(t, "REJECTED", resolveData["new_status"])
}
