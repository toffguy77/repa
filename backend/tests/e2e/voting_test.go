package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupVotingScenario creates a group with 3 members and a VOTING season with questions.
func setupVotingScenario(t *testing.T) (groupID, seasonID string, questionIDs []string, userIDs []string, tokens []string) {
	t.Helper()

	uid1, tok1 := createTestUser(t, "vote_user1_"+t.Name())
	uid2, tok2 := createTestUser(t, "vote_user2_"+t.Name())
	uid3, tok3 := createTestUser(t, "vote_user3_"+t.Name())

	group := createTestGroup(t, tok1, "Voting Group "+t.Name(), []string{"FUNNY", "SKILLS"})
	groupID = group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID = createSeasonDirectly(t, groupID)
	questionIDs = addSeasonQuestions(t, seasonID, 3)

	return groupID, seasonID, questionIDs, []string{uid1, uid2, uid3}, []string{tok1, tok2, tok3}
}

func TestVoting_GetVotingSession(t *testing.T) {
	_, seasonID, _, _, tokens := setupVotingScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, seasonID, data["season_id"])

	questions := data["questions"].([]any)
	assert.GreaterOrEqual(t, len(questions), 3)

	targets := data["targets"].([]any)
	// User should see other members but not themselves
	assert.Equal(t, 2, len(targets), "should see 2 other members as targets")

	progress := data["progress"].(map[string]any)
	assert.Equal(t, float64(0), progress["answered"])
}

func TestVoting_CastVote(t *testing.T) {
	_, seasonID, questionIDs, userIDs, tokens := setupVotingScenario(t)

	// User1 votes for User2 on question1
	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[1],
	}, tokens[0])
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	data := getData(t, resp)
	vote := data["vote"].(map[string]any)
	assert.Equal(t, questionIDs[0], vote["question_id"])
	assert.Equal(t, userIDs[1], vote["target_id"])

	progress := data["progress"].(map[string]any)
	assert.Equal(t, float64(1), progress["answered"])
}

func TestVoting_SelfVote(t *testing.T) {
	_, seasonID, questionIDs, userIDs, tokens := setupVotingScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[0], // voting for self
	}, tokens[0])
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "SELF_VOTE", errObj["code"])
}

func TestVoting_AlreadyVoted(t *testing.T) {
	_, seasonID, questionIDs, userIDs, tokens := setupVotingScenario(t)

	// First vote succeeds
	doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[1],
	}, tokens[0])

	// Second vote on same question should fail
	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[2],
	}, tokens[0])
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "ALREADY_VOTED", errObj["code"])
}

func TestVoting_NotMember(t *testing.T) {
	_, seasonID, questionIDs, userIDs, _ := setupVotingScenario(t)
	_, outsiderToken := createTestUser(t, "vote_outsider_"+t.Name())

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[1],
	}, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "NOT_MEMBER", errObj["code"])
}

func TestVoting_InvalidQuestion(t *testing.T) {
	_, seasonID, _, userIDs, tokens := setupVotingScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": "nonexistent-question-id",
		"target_id":   userIDs[1],
	}, tokens[0])
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "INVALID_QUESTION", errObj["code"])
}

func TestVoting_SeasonNotVoting(t *testing.T) {
	uid1, tok1 := createTestUser(t, "vote_revealed_u1_"+t.Name())
	_, tok2 := createTestUser(t, "vote_revealed_u2_"+t.Name())
	_, tok3 := createTestUser(t, "vote_revealed_u3_"+t.Name())

	group := createTestGroup(t, tok1, "Revealed Voting "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID := createRevealedSeason(t, groupID)
	_ = uid1

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, tok1)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "SEASON_NOT_VOTING", errObj["code"])
}

func TestVoting_GetProgress(t *testing.T) {
	_, seasonID, _, _, tokens := setupVotingScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/progress", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.NotNil(t, data["voted_count"])
	assert.NotNil(t, data["total_count"])
	assert.NotNil(t, data["quorum_reached"])
	assert.NotNil(t, data["user_voted"])
}
