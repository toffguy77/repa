package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/repa-app/repa/internal/schedule"
)

// activeSeason returns the active_season object from GET /groups/:id.
func activeSeason(t *testing.T, token, groupID string) map[string]any {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "get group failed: %s", string(resp.RawBody))
	data := getData(t, resp)
	season, ok := data["active_season"].(map[string]any)
	require.True(t, ok, "group has no active_season: %s", string(resp.RawBody))
	return season
}

// completeVoting answers every question of a season for one member.
func completeVoting(t *testing.T, token, seasonID string) {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "voting session failed: %s", string(resp.RawBody))
	data := getData(t, resp)

	questions := data["questions"].([]any)
	targets := data["targets"].([]any)
	require.NotEmpty(t, targets, "no targets to vote for")

	// Rotate targets so every member ends up with votes; always picking targets[0] can
	// leave a member with an empty card and make the assertion lie about the cause.
	for i, q := range questions {
		question := q.(map[string]any)
		if answered, _ := question["answered"].(bool); answered {
			continue
		}
		target := targets[i%len(targets)].(map[string]any)
		voteResp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
			"question_id": question["question_id"],
			"target_id":   target["user_id"],
		}, token)
		require.Equal(t, http.StatusCreated, voteResp.StatusCode,
			"cast vote failed: %s", string(voteResp.RawBody))
	}
}

// forceRevealTimePassed moves a season's reveal time into the past so the reveal pipeline
// will pick it up, the way the worker would after Friday 20:00 MSK.
func forceRevealTimePassed(t *testing.T, seasonID string) {
	t.Helper()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`UPDATE seasons SET reveal_at = NOW() - INTERVAL '1 minute' WHERE id = $1`, seasonID)
	require.NoError(t, err)
}

func seasonRow(t *testing.T, seasonID string) (status string, kind string, revealAt time.Time, postponeCount int) {
	t.Helper()
	err := suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT status, kind, reveal_at, postpone_count FROM seasons WHERE id = $1`, seasonID,
	).Scan(&status, &kind, &revealAt, &postponeCount)
	require.NoError(t, err)
	return
}

// --- 7.1: the kickoff path ---

func TestSeasonColdStart_KickoffRevealsAfterThreeVoters(t *testing.T) {
	_, founderToken := createTestUser(t, shortUsername("kick_a"))
	_, secondToken := createTestUser(t, shortUsername("kick_b"))
	_, thirdToken := createTestUser(t, shortUsername("kick_c"))

	group := createTestGroup(t, founderToken, shortName("Kickoff Group"), []string{"FUNNY", "SKILLS"})
	groupID := group["id"].(string)
	inviteCode := group["invite_code"].(string)

	// Season #1 is a kickoff season and is already open for voting.
	season := activeSeason(t, founderToken, groupID)
	seasonID := season["id"].(string)
	assert.Equal(t, "KICKOFF", season["kind"])
	assert.Equal(t, "VOTING", season["status"])

	startsAt, err := time.Parse(time.RFC3339, season["starts_at"].(string))
	require.NoError(t, err)
	assert.False(t, startsAt.After(time.Now()),
		"kickoff voting must be open immediately, starts_at was %s", startsAt)

	// Before anyone joins, the Reveal is blocked on members.
	assert.Equal(t, "WAITING_FOR_MEMBERS", season["reveal_state"])
	assert.Equal(t, float64(2), season["members_needed"])

	joinGroup(t, secondToken, inviteCode)
	joinGroup(t, thirdToken, inviteCode)

	// The fallback reveal time is the weekly Friday, days away.
	_, _, fallbackRevealAt, _ := seasonRow(t, seasonID)
	assert.True(t, fallbackRevealAt.After(time.Now().Add(24*time.Hour)),
		"fallback reveal should be days out, was %s", fallbackRevealAt)

	completeVoting(t, founderToken, seasonID)
	completeVoting(t, secondToken, seasonID)

	// Two voters is not enough: the reveal time must not have moved.
	_, _, afterTwo, _ := seasonRow(t, seasonID)
	assert.True(t, afterTwo.Equal(fallbackRevealAt),
		"reveal time moved after only 2 voters: %s -> %s", fallbackRevealAt, afterTwo)

	completeVoting(t, thirdToken, seasonID)

	// The third completing voter brings the Reveal forward to ~KickoffRevealDelay.
	_, _, afterThree, _ := seasonRow(t, seasonID)
	assert.True(t, afterThree.Before(time.Now().Add(schedule.KickoffRevealDelay+2*time.Minute)),
		"reveal should be within the hour, was %s", afterThree)
	assert.True(t, afterThree.After(time.Now()),
		"reveal should still be in the future, was %s", afterThree)

	// Drive the reveal the way the worker would once that time passes.
	forceRevealTimePassed(t, seasonID)
	result, err := suiteRevealService.ProcessReveal(context.Background(), seasonID, 1)
	require.NoError(t, err)
	require.True(t, result.Revealed, "expected the kickoff season to reveal: %+v", result)

	status, _, _, _ := seasonRow(t, seasonID)
	assert.Equal(t, "REVEALED", status)

	// Every member now has a card with results.
	for _, token := range []string{founderToken, secondToken, thirdToken} {
		resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode, "reveal failed: %s", string(resp.RawBody))
		data := getData(t, resp)
		card := data["my_card"].(map[string]any)
		assert.NotEmpty(t, card["top_attributes"], "member should have attributes on their card")
	}
}

func TestSeasonColdStart_KickoffRevealIsNotPushedBackByLaterVoter(t *testing.T) {
	_, aToken := createTestUser(t, shortUsername("kickp_a"))
	_, bToken := createTestUser(t, shortUsername("kickp_b"))
	_, cToken := createTestUser(t, shortUsername("kickp_c"))
	_, dToken := createTestUser(t, shortUsername("kickp_d"))

	group := createTestGroup(t, aToken, shortName("Kickoff Push Back"), []string{"FUNNY", "SKILLS"})
	inviteCode := group["invite_code"].(string)
	seasonID := activeSeason(t, aToken, group["id"].(string))["id"].(string)

	joinGroup(t, bToken, inviteCode)
	joinGroup(t, cToken, inviteCode)
	joinGroup(t, dToken, inviteCode)

	completeVoting(t, aToken, seasonID)
	completeVoting(t, bToken, seasonID)
	completeVoting(t, cToken, seasonID)

	_, _, scheduled, _ := seasonRow(t, seasonID)

	// A fourth member finishing later must not move the Reveal.
	completeVoting(t, dToken, seasonID)

	_, _, afterFourth, _ := seasonRow(t, seasonID)
	assert.True(t, afterFourth.Equal(scheduled),
		"a later voter moved the reveal: %s -> %s", scheduled, afterFourth)
}

// --- 7.2: the participation floor path ---

func TestSeasonColdStart_TwoMemberGroupNeverReveals(t *testing.T) {
	_, aToken := createTestUser(t, shortUsername("floor_a"))
	_, bToken := createTestUser(t, shortUsername("floor_b"))

	group := createTestGroup(t, aToken, shortName("Floor Group"), []string{"FUNNY", "SKILLS"})
	groupID := group["id"].(string)
	joinGroup(t, bToken, group["invite_code"].(string))

	seasonID := activeSeason(t, aToken, groupID)["id"].(string)

	completeVoting(t, aToken, seasonID)
	completeVoting(t, bToken, seasonID)

	// Both members voted, so the 40% quorum is satisfied — but the floor is not.
	forceRevealTimePassed(t, seasonID)
	for attempt := 1; attempt <= 3; attempt++ {
		result, err := suiteRevealService.ProcessReveal(context.Background(), seasonID, attempt)
		require.NoError(t, err)
		assert.False(t, result.Revealed, "attempt %d must not reveal a 2-member group", attempt)
	}

	status, kind, revealAt, postponeCount := seasonRow(t, seasonID)
	assert.Equal(t, "VOTING", status, "season must stay open")
	assert.Equal(t, "KICKOFF", kind, "a postponed kickoff stays a kickoff")
	assert.Equal(t, 1, postponeCount, "the final attempt should postpone exactly once")
	assert.True(t, revealAt.After(time.Now()), "postponed reveal must be in the future")

	// No results were aggregated, so nobody can see a card.
	var resultCount int
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM season_results WHERE season_id = $1`, seasonID).Scan(&resultCount))
	assert.Zero(t, resultCount, "a postponed season must not aggregate results")

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, aToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"an unrevealed season must not serve a card")

	// The group reports what it is waiting for.
	season := activeSeason(t, aToken, groupID)
	assert.Equal(t, "WAITING_FOR_MEMBERS", season["reveal_state"])
	assert.Equal(t, float64(1), season["members_needed"])
}

func TestSeasonColdStart_PostponedSeasonRevealsOnceGroupGrows(t *testing.T) {
	_, aToken := createTestUser(t, shortUsername("grow_a"))
	_, bToken := createTestUser(t, shortUsername("grow_b"))
	_, cToken := createTestUser(t, shortUsername("grow_c"))

	group := createTestGroup(t, aToken, shortName("Growing Group"), []string{"FUNNY", "SKILLS"})
	groupID := group["id"].(string)
	inviteCode := group["invite_code"].(string)
	joinGroup(t, bToken, inviteCode)

	seasonID := activeSeason(t, aToken, groupID)["id"].(string)
	completeVoting(t, aToken, seasonID)
	completeVoting(t, bToken, seasonID)

	forceRevealTimePassed(t, seasonID)
	result, err := suiteRevealService.ProcessReveal(context.Background(), seasonID, 3)
	require.NoError(t, err)
	require.True(t, result.Postponed, "expected a postponement, got %+v", result)

	// The third member joins and votes; the earlier votes are still counted.
	joinGroup(t, cToken, inviteCode)
	completeVoting(t, cToken, seasonID)

	forceRevealTimePassed(t, seasonID)
	result, err = suiteRevealService.ProcessReveal(context.Background(), seasonID, 1)
	require.NoError(t, err)
	assert.True(t, result.Revealed, "expected the postponed season to reveal, got %+v", result)

	status, _, _, _ := seasonRow(t, seasonID)
	assert.Equal(t, "REVEALED", status)
}

// --- Detector gating in a small group ---

func TestSeasonColdStart_DetectorRefusedInSmallGroup(t *testing.T) {
	aUserID, aToken := createTestUser(t, shortUsername("det_a"))
	_, bToken := createTestUser(t, shortUsername("det_b"))
	_, cToken := createTestUser(t, shortUsername("det_c"))

	group := createTestGroup(t, aToken, shortName("Detector Small"), []string{"FUNNY", "SKILLS"})
	groupID := group["id"].(string)
	joinGroup(t, bToken, group["invite_code"].(string))
	joinGroup(t, cToken, group["invite_code"].(string))

	seasonID := createRevealedSeason(t, groupID)
	addCrystals(t, aUserID, 100)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector", nil, aToken)
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "body: %s", string(resp.RawBody))
	errObj := resp.Body["error"].(map[string]any)
	assert.Equal(t, "GROUP_TOO_SMALL", errObj["code"])

	// Status reports unavailability instead of merely "not purchased".
	statusResp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/detector", nil, aToken)
	require.Equal(t, http.StatusOK, statusResp.StatusCode)
	data := getData(t, statusResp)
	assert.Equal(t, false, data["available"])
	assert.Equal(t, false, data["purchased"])
}
