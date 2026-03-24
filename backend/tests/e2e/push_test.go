package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPush_RegisterToken(t *testing.T) {
	_, token := createTestUser(t, "push_reg_"+t.Name())

	resp := doRequest(t, "POST", "/api/v1/push/register", map[string]any{
		"token":    "fcm-test-token-12345",
		"platform": "ios",
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, true, data["registered"])
}

func TestPush_RegisterToken_InvalidPlatform(t *testing.T) {
	_, token := createTestUser(t, "push_bad_plat_"+t.Name())

	resp := doRequest(t, "POST", "/api/v1/push/register", map[string]any{
		"token":    "fcm-token",
		"platform": "windows", // invalid
	}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPush_GetQuestionCandidates(t *testing.T) {
	_, tok1 := createTestUser(t, "push_cand_u1_"+t.Name())
	_, tok2 := createTestUser(t, "push_cand_u2_"+t.Name())

	group := createTestGroup(t, tok1, "Push Candidates "+t.Name(), []string{"FUNNY"})
	joinGroup(t, tok2, group["invite_code"].(string))

	resp := doRequest(t, "GET", "/api/v1/groups/"+group["id"].(string)+"/next-season/question-candidates", nil, tok1)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	candidates := data["candidates"].([]any)
	assert.Len(t, candidates, 3)

	cand := candidates[0].(map[string]any)
	assert.NotEmpty(t, cand["id"])
	assert.NotEmpty(t, cand["text"])
	assert.NotEmpty(t, cand["category"])
}

func TestPush_VoteQuestion(t *testing.T) {
	_, tok1 := createTestUser(t, "push_vote_u1_"+t.Name())
	_, tok2 := createTestUser(t, "push_vote_u2_"+t.Name())

	group := createTestGroup(t, tok1, "Push Vote "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))

	// Get candidates first
	candResp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/next-season/question-candidates", nil, tok1)
	require.Equal(t, http.StatusOK, candResp.StatusCode)
	candidates := getData(t, candResp)["candidates"].([]any)
	questionID := candidates[0].(map[string]any)["id"].(string)

	// Vote for a question
	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/next-season/vote-question", map[string]any{
		"questionId": questionID,
	}, tok1)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := getData(t, resp)
	assert.Equal(t, true, data["voted"])
}
