package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReactionScenario(t *testing.T) (seasonID string, userIDs []string, tokens []string) {
	t.Helper()

	uid1, tok1 := createTestUser(t, "react_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "react_u2_"+t.Name())
	uid3, tok3 := createTestUser(t, "react_u3_"+t.Name())

	group := createTestGroup(t, tok1, "React Group "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID = createRevealedSeason(t, groupID)
	questionIDs := addSeasonQuestions(t, seasonID, 3)
	createSeasonResults(t, seasonID, uid1, questionIDs)
	createSeasonResults(t, seasonID, uid2, questionIDs)

	return seasonID, []string{uid1, uid2, uid3}, []string{tok1, tok2, tok3}
}

func TestReactions_CreateReaction(t *testing.T) {
	seasonID, userIDs, tokens := setupReactionScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/members/"+userIDs[1]+"/reactions", map[string]any{
		"emoji": "🔥",
	}, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReactions_GetReactions(t *testing.T) {
	seasonID, userIDs, tokens := setupReactionScenario(t)

	// Create a reaction first
	doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/members/"+userIDs[1]+"/reactions", map[string]any{
		"emoji": "😂",
	}, tokens[0])

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/members/"+userIDs[1]+"/reactions", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReactions_InvalidEmoji(t *testing.T) {
	seasonID, userIDs, tokens := setupReactionScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/members/"+userIDs[1]+"/reactions", map[string]any{
		"emoji": "❤️", // not in allowed list
	}, tokens[0])
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "INVALID_EMOJI", errObj["code"])
}

func TestReactions_SelfReaction(t *testing.T) {
	seasonID, userIDs, tokens := setupReactionScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/members/"+userIDs[0]+"/reactions", map[string]any{
		"emoji": "🔥",
	}, tokens[0]) // reacting to self
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "SELF_REACTION", errObj["code"])
}

func TestReactions_SeasonNotRevealed(t *testing.T) {
	uid1, tok1 := createTestUser(t, "react_norev_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "react_norev_u2_"+t.Name())
	_, tok3 := createTestUser(t, "react_norev_u3_"+t.Name())

	group := createTestGroup(t, tok1, "Not Revealed React "+t.Name(), []string{"FUNNY"})
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID := createSeasonDirectly(t, group["id"].(string)) // VOTING status
	_ = uid1

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/members/"+uid2+"/reactions", map[string]any{
		"emoji": "🔥",
	}, tok1)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "SEASON_NOT_REVEALED", errObj["code"])
}
