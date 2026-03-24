package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnonymity_RevealNeverExposesVoterID ensures that reveal responses
// never contain voter_id linked to specific votes.
func TestAnonymity_RevealNeverExposesVoterID(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Serialize the full response and check no "voter_id" key exists
	raw, err := json.Marshal(resp.Body)
	require.NoError(t, err)

	bodyStr := string(raw)
	assert.NotContains(t, bodyStr, `"voter_id"`,
		"Reveal response must NEVER contain voter_id — anonymity violation!")
}

// TestAnonymity_DetectorDoesNotExposeQuestionBinding ensures that the detector
// returns only voter IDs without binding them to specific questions/answers.
func TestAnonymity_DetectorDoesNotExposeQuestionBinding(t *testing.T) {
	seasonID, userIDs, tokens, questionIDs := setupRevealScenario(t)

	// Cast some votes first (need a voting season)
	// Create a separate voting scenario for this test
	uid1, tok1 := createTestUser(t, "anon_det_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "anon_det_u2_"+t.Name())
	uid3, tok3 := createTestUser(t, "anon_det_u3_"+t.Name())

	group := createTestGroup(t, tok1, "Anon Det "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	votingSeasonID := createSeasonDirectly(t, groupID)
	votingQIDs := addSeasonQuestions(t, votingSeasonID, 3)

	// User1 votes for User2
	doRequest(t, "POST", "/api/v1/seasons/"+votingSeasonID+"/votes", map[string]any{
		"question_id": votingQIDs[0],
		"target_id":   uid2,
	}, tok1)

	// User2 votes for User3
	doRequest(t, "POST", "/api/v1/seasons/"+votingSeasonID+"/votes", map[string]any{
		"question_id": votingQIDs[0],
		"target_id":   uid3,
	}, tok2)

	// Now test the revealed season's detector
	_ = userIDs
	_ = tokens
	_ = questionIDs
	_ = uid1
	_ = tok3

	addCrystals(t, userIDs[0], 20)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, err := json.Marshal(resp.Body)
	require.NoError(t, err)
	bodyStr := string(raw)

	// Detector result should never contain question_id linked to voter_id
	// Look for patterns like "voter_id" next to "question_id"
	assert.False(t,
		strings.Contains(bodyStr, `"voter_id"`) && strings.Contains(bodyStr, `"question_id"`),
		"Detector response must NOT link voter_id to question_id")
}

// TestAnonymity_VotingSessionExcludesVoterInfo ensures that voting session
// responses don't leak who voted for whom.
func TestAnonymity_VotingSessionExcludesVoterInfo(t *testing.T) {
	_, seasonID, questionIDs, userIDs, tokens := setupVotingScenario(t)

	// Cast a vote
	doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionIDs[0],
		"target_id":   userIDs[1],
	}, tokens[0])

	// Get voting session as another user
	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, tokens[1])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, err := json.Marshal(resp.Body)
	require.NoError(t, err)
	bodyStr := string(raw)

	// The response should not contain voter_id
	assert.NotContains(t, bodyStr, `"voter_id"`,
		"Voting session response must not contain voter_id")
}

// TestAnonymity_MembersCardsNoVoterLink checks that members cards
// don't expose voter identity.
func TestAnonymity_MembersCardsNoVoterLink(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/members-cards", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, err := json.Marshal(resp.Body)
	require.NoError(t, err)
	bodyStr := string(raw)

	assert.NotContains(t, bodyStr, `"voter_id"`,
		"Members cards response must not contain voter_id")
}
