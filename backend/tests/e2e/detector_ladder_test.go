package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ladderScenario builds a five-member group with a revealed season, real votes, and a funded
// member — five because every paid rung of the detector needs the group to clear the anonymity
// floor.
func ladderScenario(t *testing.T) (seasonID, token, userID string) {
	t.Helper()

	suffix := uuid.New().String()[:6]
	userID, token = createTestUser(t, "lad_a_"+suffix)
	group := createTestGroup(t, token, shortName("Ladder"), []string{"FUNNY"})
	groupID := group["id"].(string)
	code := group["invite_code"].(string)

	var tokens []string
	for i := 0; i < 4; i++ {
		_, tok := createTestUser(t, fmt.Sprintf("lad_%d_%s", i, suffix))
		joinGroup(t, tok, code)
		tokens = append(tokens, tok)
	}

	// Vote in the open season so there are real voters, then reveal it.
	openSeason := activeSeason(t, token, groupID)["id"].(string)
	for _, tok := range tokens {
		completeVoting(t, tok, openSeason)
	}
	forceRevealTimePassed(t, openSeason)
	result, err := suiteRevealService.ProcessReveal(ctxBackground(), openSeason, 1)
	require.NoError(t, err)
	require.True(t, result.Revealed, "the scenario needs a revealed season")

	addCrystals(t, userID, 100)
	return openSeason, token, userID
}

func detectorState(t *testing.T, token, seasonID string) map[string]any {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/detector", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	return getData(t, resp)
}

// --- 5.1 ---

func TestDetectorLadder_FreeCountIsVisibleWithoutBuying(t *testing.T) {
	seasonID, token, _ := ladderScenario(t)

	data := detectorState(t, token, seasonID)

	assert.Equal(t, false, data["purchased"])
	assert.Greater(t, data["voter_count"], float64(0),
		"the free rung must be visible without a purchase")
	assert.Empty(t, data["voters"], "no names before paying")
	assert.Equal(t, true, data["hint_available"])
	assert.Equal(t, float64(3), data["hint_cost"])
	assert.Equal(t, float64(10), data["full_cost"])
}

func TestDetectorLadder_HintRevealsOneVoterPartially(t *testing.T) {
	seasonID, token, _ := ladderScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector/hint", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "hint failed: %s", string(resp.RawBody))

	data := getData(t, resp)
	hints := data["hints"].([]any)
	require.Len(t, hints, 1)

	hint := hints[0].(map[string]any)
	assert.NotEmpty(t, hint["first_letter"])
	assert.Len(t, []rune(hint["first_letter"].(string)), 1, "a hint reveals one character")
	assert.NotContains(t, hint, "username", "a hint has no field for a full name")
	assert.Empty(t, data["voters"], "a hint is not the full list")
	assert.Equal(t, false, data["purchased"])
}

func TestDetectorLadder_TwoHintsRevealTwoDifferentVoters(t *testing.T) {
	seasonID, token, _ := ladderScenario(t)

	for i := 0; i < 2; i++ {
		resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector/hint", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode, "hint %d: %s", i, string(resp.RawBody))
	}

	data := detectorState(t, token, seasonID)
	hints := data["hints"].([]any)
	require.Len(t, hints, 2, "hints are drawn without replacement")
}

// --- 5.2 ---

func TestDetectorLadder_HintsExhaustThenRefuse(t *testing.T) {
	seasonID, token, _ := ladderScenario(t)

	// Four other members voted, so four hints are available.
	for i := 0; i < 4; i++ {
		resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector/hint", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode, "hint %d: %s", i, string(resp.RawBody))
	}

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector/hint", nil, token)
	require.Equal(t, http.StatusConflict, resp.StatusCode, string(resp.RawBody))
	assert.Equal(t, "NOTHING_TO_REVEAL", resp.Body["error"].(map[string]any)["code"])
}

func TestDetectorLadder_FullListStillWorksAfterHints(t *testing.T) {
	seasonID, token, _ := ladderScenario(t)

	require.Equal(t, http.StatusOK,
		doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector/hint", nil, token).StatusCode)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	data := getData(t, resp)
	assert.Equal(t, true, data["purchased"])
	assert.NotEmpty(t, data["voters"], "the top rung names everyone")
	assert.Equal(t, false, data["hint_available"], "nothing left to hint at")
}

func TestDetectorLadder_SmallGroupRefusesPaidRungsButShowsTheCount(t *testing.T) {
	suffix := uuid.New().String()[:6]
	userID, token := createTestUser(t, "lad_s_"+suffix)
	group := createTestGroup(t, token, shortName("Ladder Small"), []string{"FUNNY"})
	_, bToken := createTestUser(t, "lad_sb_"+suffix)
	joinGroup(t, bToken, group["invite_code"].(string))

	seasonID := createRevealedSeason(t, group["id"].(string))
	addCrystals(t, userID, 100)

	data := detectorState(t, token, seasonID)
	assert.Equal(t, false, data["available"])
	assert.Equal(t, false, data["hint_available"])
	assert.Contains(t, data, "voter_count", "a count names nobody, so it is still reported")

	for _, path := range []string{"/detector/hint", "/detector"} {
		resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+path, nil, token)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode,
			"%s should be refused in a small group: %s", path, string(resp.RawBody))
	}
}

// ctxBackground keeps the reveal-service call sites short.
func ctxBackground() context.Context { return context.Background() }
