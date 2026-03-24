package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupQuestionsScenario(t *testing.T) (groupID string, userIDs []string, tokens []string) {
	t.Helper()

	uid1, tok1 := createTestUser(t, "q_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "q_u2_"+t.Name())

	group := createTestGroup(t, tok1, "Questions Group "+t.Name(), []string{"FUNNY", "SKILLS"})
	groupID = group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))

	return groupID, []string{uid1, uid2}, []string{tok1, tok2}
}

func TestQuestions_CreateQuestion(t *testing.T) {
	groupID, _, tokens := setupQuestionsScenario(t)

	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Who is the funniest person?",
		"category": "FUNNY",
	}, tokens[0])
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	data := getData(t, resp)
	question := data["question"].(map[string]any)
	assert.Equal(t, "Who is the funniest person?", question["text"])
	assert.Equal(t, "FUNNY", question["category"])
	assert.Equal(t, "USER", question["source"])

	// Moderation response
	moderation := data["moderation"].(map[string]any)
	assert.NotNil(t, moderation["approved"])
}

func TestQuestions_ListQuestions(t *testing.T) {
	groupID, _, tokens := setupQuestionsScenario(t)

	// Create a question
	doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Who codes the best?",
		"category": "SKILLS",
	}, tokens[0])

	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/questions", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	questions := data["questions"].([]any)
	assert.GreaterOrEqual(t, len(questions), 1)
}

func TestQuestions_DeleteOwnQuestion(t *testing.T) {
	groupID, _, tokens := setupQuestionsScenario(t)

	// Create a question
	createResp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Delete me please",
		"category": "FUNNY",
	}, tokens[0])
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	qID := getData(t, createResp)["question"].(map[string]any)["id"].(string)

	// Delete own question
	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/questions/"+qID, nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := getData(t, resp)
	assert.Equal(t, true, data["deleted"])
}

func TestQuestions_DeleteOthersQuestion_Forbidden(t *testing.T) {
	groupID, _, tokens := setupQuestionsScenario(t)

	// User1 creates a question
	createResp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Unauthorized delete",
		"category": "FUNNY",
	}, tokens[0])
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	qID := getData(t, createResp)["question"].(map[string]any)["id"].(string)

	// User2 tries to delete it (not admin, not author)
	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/questions/"+qID, nil, tokens[1])
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestQuestions_ReportQuestion(t *testing.T) {
	groupID, _, tokens := setupQuestionsScenario(t)

	// Create a question
	createResp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "Report me question",
		"category": "FUNNY",
	}, tokens[0])
	require.Equal(t, http.StatusCreated, createResp.StatusCode, "create question failed: %s", string(createResp.RawBody))
	qID := getData(t, createResp)["question"].(map[string]any)["id"].(string)

	// Report from user2
	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions/"+qID+"/report", map[string]any{
		"reason": "inappropriate",
	}, tokens[1])
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := getData(t, resp)
	assert.Equal(t, true, data["reported"])
}

func TestQuestions_NotMember(t *testing.T) {
	groupID, _, _ := setupQuestionsScenario(t)
	_, outsiderToken := createTestUser(t, "q_outsider_"+t.Name())

	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/questions", map[string]any{
		"text":     "I should not be here",
		"category": "FUNNY",
	}, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
