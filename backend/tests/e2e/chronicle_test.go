package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/repa-app/repa/internal/eligibility"
	chroniclesvc "github.com/repa-app/repa/internal/service/chronicle"
)

// pastSeason inserts a season that already revealed, with one result per question.
//
// status is a parameter because CLOSED is what a REVEALED season becomes when the next one opens — a
// group's older history is CLOSED, and the chronicle has to count both.
func pastSeason(t *testing.T, groupID, status string, number int, results []pastResult) string {
	t.Helper()
	seasonID := uuid.New().String()
	now := time.Now()
	weeksAgo := time.Duration(number) * -7 * 24 * time.Hour
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO seasons (id, group_id, number, status, starts_at, reveal_at, ends_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		seasonID, groupID, number, status,
		now.Add(weeksAgo-7*24*time.Hour), now.Add(weeksAgo), now.Add(weeksAgo))
	require.NoError(t, err)

	for _, r := range results {
		questionID := uuid.New().String()
		_, err := suite.sqlDB.ExecContext(context.Background(),
			`INSERT INTO questions (id, text, category, source, status, tone)
			 VALUES ($1, $2, 'FUNNY', 'SYSTEM', 'ACTIVE', 'NEUTRAL')`,
			questionID, r.questionText)
		require.NoError(t, err)

		_, err = suite.sqlDB.ExecContext(context.Background(),
			`INSERT INTO season_results (id, season_id, target_id, question_id, vote_count, total_voters, percentage)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New().String(), seasonID, r.targetID, questionID,
			r.voteCount, r.totalVoters, r.percentage)
		require.NoError(t, err)
	}
	return seasonID
}

type pastResult struct {
	questionText string
	targetID     string
	voteCount    int
	totalVoters  int
	percentage   float64
}

func setGuessAccuracy(t *testing.T, userID, groupID string, accuracy float64, seasons int) {
	t.Helper()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO user_group_stats (id, user_id, group_id, seasons_played, voting_streak,
		   max_voting_streak, guess_accuracy, total_votes_cast, total_votes_received)
		 VALUES ($1, $2, $3, $4, 1, 1, $5, 0, 0)
		 ON CONFLICT (user_id, group_id) DO UPDATE SET guess_accuracy = EXCLUDED.guess_accuracy,
		   seasons_played = EXCLUDED.seasons_played`,
		uuid.New().String(), userID, groupID, seasons, accuracy)
	require.NoError(t, err)
}

func getChronicle(t *testing.T, token, groupID string) map[string]any {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/chronicle", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	return resp.Body["data"].(map[string]any)
}

func TestChronicle_ShowsPastSeasonsAndHidesTheOpenOne(t *testing.T) {
	userID, token := createTestUser(t, "chron_owner")
	group := createTestGroup(t, token, "Chronicle", []string{"FUNNY"})
	groupID := group["id"].(string)

	// Season 1 has aged into CLOSED, season 2 is the most recent reveal. Both must appear: filtering
	// on REVEALED alone would show only season 2 while looking perfectly correct.
	pastSeason(t, groupID, "CLOSED", 1, []pastResult{
		{questionText: "Кто чаще всех смеётся?", targetID: userID, voteCount: 8, totalVoters: 10, percentage: 80},
	})
	pastSeason(t, groupID, "REVEALED", 2, []pastResult{
		{questionText: "Кто всегда поможет?", targetID: userID, voteCount: 7, totalVoters: 10, percentage: 70},
	})

	// An open season carrying results, which is the state that makes this assertion mean something:
	// without a status filter it would appear, and the group's own kickoff season has no results so
	// its absence would prove nothing.
	pastSeason(t, groupID, "VOTING", 3, []pastResult{
		{questionText: "Ещё голосуем", targetID: userID, voteCount: 1, totalVoters: 2, percentage: 50},
	})

	data := getChronicle(t, token, groupID)
	seasons := data["seasons"].([]any)
	require.Len(t, seasons, 2, "both the CLOSED and the REVEALED season belong in the record")

	// Newest first.
	assert.Equal(t, float64(2), seasons[0].(map[string]any)["number"])
	assert.Equal(t, float64(1), seasons[1].(map[string]any)["number"])
	assert.Equal(t, float64(2), data["seasons_total"])

	for _, s := range seasons {
		season := s.(map[string]any)
		entries := season["entries"].([]any)
		assert.NotEmpty(t, entries, "a revealed season in the record should carry its results")
		for _, e := range entries {
			assert.NotEqual(t, "Ещё голосуем", e.(map[string]any)["question_text"],
				"a season still being voted in has no place in the record")
		}
	}
}

func TestChronicle_NonMemberIsRefused(t *testing.T) {
	_, ownerToken := createTestUser(t, "chron_in")
	_, outsiderToken := createTestUser(t, "chron_out")
	group := createTestGroup(t, ownerToken, "Private", []string{"FUNNY"})

	resp := doRequest(t, "GET", "/api/v1/groups/"+group["id"].(string)+"/chronicle", nil, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "NOT_MEMBER", getError(t, resp)["code"])
}

func TestChronicle_EmptyForANewGroup(t *testing.T) {
	_, token := createTestUser(t, "chron_new")
	group := createTestGroup(t, token, "Brand New", []string{"FUNNY"})

	data := getChronicle(t, token, group["id"].(string))
	assert.Empty(t, data["seasons"].([]any))
	assert.Equal(t, float64(0), data["seasons_total"])
	// Empty, not an error: a new group is the normal case, not a failure.
	standing := data["standing"].(map[string]any)
	assert.Equal(t, false, standing["available"])
	assert.NotEmpty(t, standing["reason"], "the app needs something to show instead of the ranking")
}

func TestChronicle_MarksResultsFromTooFewVoters(t *testing.T) {
	userID, token := createTestUser(t, "chron_small")
	group := createTestGroup(t, token, "Tiny", []string{"FUNNY"})
	groupID := group["id"].(string)

	below := eligibility.MinDetectorMembers - 1
	pastSeason(t, groupID, "REVEALED", 1, []pastResult{
		{questionText: "Мало голосов", targetID: userID, voteCount: 2, totalVoters: below, percentage: 66.7},
		{questionText: "Достаточно голосов", targetID: userID, voteCount: 9, totalVoters: 12, percentage: 75},
	})

	data := getChronicle(t, token, groupID)
	entries := data["seasons"].([]any)[0].(map[string]any)["entries"].([]any)
	require.Len(t, entries, 2)

	marked := map[string]bool{}
	for _, e := range entries {
		entry := e.(map[string]any)
		marked[entry["question_text"].(string)] = entry["small_sample"].(bool)
	}
	assert.True(t, marked["Мало голосов"], "a result over %d voters must carry the warning", below)
	assert.False(t, marked["Достаточно голосов"], "12 voters is not a small sample")
}

func TestChronicle_StandingWithheldUntilTwoSeasons(t *testing.T) {
	userID, token := createTestUser(t, "chron_standing")
	group := createTestGroup(t, token, "Standing", []string{"FUNNY"})
	groupID := group["id"].(string)
	setGuessAccuracy(t, userID, groupID, 0.8, 2)

	pastSeason(t, groupID, "REVEALED", 1, []pastResult{
		{questionText: "Первая репа", targetID: userID, voteCount: 5, totalVoters: 10, percentage: 50},
	})

	// One revealed season: the rolling average has averaged nothing, so a ranking would report luck.
	standing := getChronicle(t, token, groupID)["standing"].(map[string]any)
	require.Equal(t, false, standing["available"],
		"a ranking on one season ranks luck; expected it withheld with %d required", chroniclesvc.MinSeasonsForStanding)

	pastSeason(t, groupID, "CLOSED", 2, []pastResult{
		{questionText: "Вторая репа", targetID: userID, voteCount: 6, totalVoters: 10, percentage: 60},
	})

	standing = getChronicle(t, token, groupID)["standing"].(map[string]any)
	assert.Equal(t, true, standing["available"])
	assert.NotEmpty(t, standing["members"].([]any))
}

func TestChronicle_StandingIsOrderedByAccuracy(t *testing.T) {
	ownerID, ownerToken := createTestUser(t, "chron_rank_a")
	group := createTestGroup(t, ownerToken, "Ranked", []string{"FUNNY"})
	groupID := group["id"].(string)

	secondID, secondToken := createTestUser(t, "chron_rank_b")
	joinGroup(t, secondToken, group["invite_code"].(string))
	thirdID, thirdToken := createTestUser(t, "chron_rank_c")
	joinGroup(t, thirdToken, group["invite_code"].(string))

	setGuessAccuracy(t, ownerID, groupID, 0.40, 3)
	setGuessAccuracy(t, secondID, groupID, 0.90, 3)
	// thirdID deliberately has no stats row: joined after the last reveal.
	_ = thirdID

	for i := 1; i <= 2; i++ {
		status := "CLOSED"
		if i == 2 {
			status = "REVEALED"
		}
		pastSeason(t, groupID, status, i, []pastResult{
			{questionText: fmt.Sprintf("Репа %d", i), targetID: ownerID, voteCount: 5, totalVoters: 10, percentage: 50},
		})
	}

	standing := getChronicle(t, ownerToken, groupID)["standing"].(map[string]any)
	require.Equal(t, true, standing["available"])
	members := standing["members"].([]any)
	require.Len(t, members, 3)

	first := members[0].(map[string]any)
	assert.Equal(t, secondID, first["user_id"], "the best predictor leads")
	assert.Equal(t, float64(1), first["rank"])
	assert.Equal(t, true, first["ranked"])
	assert.Equal(t, float64(3), first["seasons_played"])

	second := members[1].(map[string]any)
	assert.Equal(t, ownerID, second["user_id"])
	assert.Equal(t, float64(2), second["rank"])

	// The member with no revealed season is unranked, not ranked at zero.
	third := members[2].(map[string]any)
	assert.Equal(t, false, third["ranked"], "a member who has not played must not be ranked")
	assert.Equal(t, float64(0), third["rank"])

	// The standing exposes an order, not a measurement. The accuracy figure is the one that matters
	// here: the chronicle above publishes each question's winner, so an extreme figure would recover a
	// member's individual votes, and the rolling average can be diffed across two weeks into a match
	// count. A rank cannot.
	for _, m := range members {
		entry := m.(map[string]any)
		for _, forbidden := range []string{"accuracy", "guess_accuracy", "question_id", "question_text", "target_id", "vote_count"} {
			assert.NotContains(t, entry, forbidden,
				"the standing must not be a way to infer who voted how")
		}
	}
}

func TestChronicle_GrowthThresholdMatchesEligibility(t *testing.T) {
	_, token := createTestUser(t, "growth_owner")
	group := createTestGroup(t, token, "Growing", []string{"FUNNY"})
	groupID := group["id"].(string)
	inviteCode := group["invite_code"].(string)

	// One member: below the reveal floor.
	threshold := groupThreshold(t, token, groupID)
	require.NotNil(t, threshold, "a group of one has a threshold to cross")
	assert.Equal(t, eligibility.UnlockReveal, threshold["unlocks"])
	assert.Equal(t, float64(eligibility.MinMembers-1), threshold["needed"])

	for i := 2; i <= eligibility.MinMembers; i++ {
		_, memberToken := createTestUser(t, fmt.Sprintf("growth_m%d", i))
		joinGroup(t, memberToken, inviteCode)
	}

	// At the reveal floor: the next threshold is the detector.
	threshold = groupThreshold(t, token, groupID)
	require.NotNil(t, threshold)
	assert.Equal(t, eligibility.UnlockDetector, threshold["unlocks"])
	assert.Equal(t, float64(eligibility.MinDetectorMembers-eligibility.MinMembers), threshold["needed"])

	for i := eligibility.MinMembers + 1; i <= eligibility.MinDetectorMembers; i++ {
		_, memberToken := createTestUser(t, fmt.Sprintf("growth_m%d", i))
		joinGroup(t, memberToken, inviteCode)
	}

	// Past every threshold: no target, rather than an invented one.
	assert.Nil(t, groupThreshold(t, token, groupID),
		"a group past every threshold must not be given something to chase")
}

func groupThreshold(t *testing.T, token, groupID string) map[string]any {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	data := resp.Body["data"].(map[string]any)
	raw, ok := data["next_threshold"]
	if !ok || raw == nil {
		return nil
	}
	return raw.(map[string]any)
}

// TestReveal_TrendUsesAClosedPreviousSeason is the regression for the status defect at the heart of
// this change. At reveal time the current season is the only REVEALED one — every earlier season was
// closed when this one opened — so a GetPreviousRevealedSeason matching only REVEALED found nothing
// and the card's trend line never rendered. The failure looked like "no trend yet", which is also what
// a genuinely first season looks like, so nothing ever flagged it.
func TestReveal_TrendUsesAClosedPreviousSeason(t *testing.T) {
	userID, token := createTestUser(t, "trend_owner")
	group := createTestGroup(t, token, "Trend", []string{"FUNNY"})
	groupID := group["id"].(string)

	// The same question in both seasons, so the trend has something to compare.
	questionID := uuid.New().String()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO questions (id, text, category, source, status, tone)
		 VALUES ($1, 'Кто чаще всех смеётся?', 'FUNNY', 'SYSTEM', 'ACTIVE', 'NEUTRAL')`, questionID)
	require.NoError(t, err)

	addResult := func(seasonID string, percentage float64) {
		_, err := suite.sqlDB.ExecContext(context.Background(),
			`INSERT INTO season_results (id, season_id, target_id, question_id, vote_count, total_voters, percentage)
			 VALUES ($1, $2, $3, $4, 5, 10, $5)`,
			uuid.New().String(), seasonID, userID, questionID, percentage)
		require.NoError(t, err)
	}

	// Previous season: CLOSED, which is the state every past season is actually in.
	prev := pastSeason(t, groupID, "CLOSED", 1, nil)
	addResult(prev, 40)

	current := pastSeason(t, groupID, "REVEALED", 2, nil)
	addResult(current, 55)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+current+"/reveal", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	card := resp.Body["data"].(map[string]any)["my_card"].(map[string]any)
	trend := card["trend"]
	require.NotNil(t, trend, "the trend must be computed against the CLOSED previous season")

	trendMap := trend.(map[string]any)
	assert.Equal(t, "up", trendMap["change"])
	assert.Equal(t, float64(15), trendMap["delta"])
	assert.Equal(t, "Кто чаще всех смеётся?", trendMap["attribute"])
}

// --- Profile accuracy privacy ---

// TestProfile_AccuracyIsOnlyVisibleToItsOwner is the end-to-end form of the anonymity rule. A member's
// match rate, read next to the per-question winners this very suite checks are public, narrows or pins
// how that member voted; and because the figure is a rolling average weighted by seasons_played, two
// consecutive readings solve for one week's match count.
func TestProfile_AccuracyIsOnlyVisibleToItsOwner(t *testing.T) {
	ownerID, ownerToken := createTestUser(t, "acc_owner")
	group := createTestGroup(t, ownerToken, "Accuracy", []string{"FUNNY"})
	groupID := group["id"].(string)

	otherID, otherToken := createTestUser(t, "acc_other")
	joinGroup(t, otherToken, group["invite_code"].(string))

	setGuessAccuracy(t, ownerID, groupID, 87.5, 4)
	setGuessAccuracy(t, otherID, groupID, 42.0, 4)

	profileStats := func(token, viewedID string) map[string]any {
		resp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/members/"+viewedID+"/profile", nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
		return resp.Body["data"].(map[string]any)["stats"].(map[string]any)
	}

	// Another member's figure is absent — not zero, which a client would render as "never matched".
	seen := profileStats(ownerToken, otherID)
	assert.NotContains(t, seen, "guess_accuracy",
		"another member's match rate must not be readable")

	// One's own is present: it describes votes they cast themselves.
	own := profileStats(ownerToken, ownerID)
	require.Contains(t, own, "guess_accuracy")
	assert.Equal(t, 87.5, own["guess_accuracy"])

	// And it is symmetric — not a privilege of the admin.
	assert.NotContains(t, profileStats(otherToken, ownerID), "guess_accuracy")
	assert.Contains(t, profileStats(otherToken, otherID), "guess_accuracy")

	// Participation stats survive: none of them counts matches.
	for _, key := range []string{"seasons_played", "voting_streak", "max_voting_streak",
		"total_votes_cast", "total_votes_received"} {
		assert.Contains(t, seen, key, "%s should still be visible", key)
	}
}

// TestProfile_AndStandingAgree checks the two surfaces that could leak the same figure. Neither may
// hand out another member's measurement; the standing publishes an order instead.
func TestProfile_AndStandingAgree(t *testing.T) {
	ownerID, ownerToken := createTestUser(t, "acc_both_a")
	group := createTestGroup(t, ownerToken, "Both", []string{"FUNNY"})
	groupID := group["id"].(string)

	otherID, otherToken := createTestUser(t, "acc_both_b")
	joinGroup(t, otherToken, group["invite_code"].(string))

	setGuessAccuracy(t, ownerID, groupID, 95.0, 3)
	setGuessAccuracy(t, otherID, groupID, 10.0, 3)

	// Two revealed seasons, so the standing is shown rather than withheld.
	for i, status := range []string{"CLOSED", "REVEALED"} {
		pastSeason(t, groupID, status, i+1, []pastResult{
			{questionText: fmt.Sprintf("Репа %d", i+1), targetID: ownerID,
				voteCount: 5, totalVoters: 10, percentage: 50},
		})
	}

	standing := getChronicle(t, ownerToken, groupID)["standing"].(map[string]any)
	require.Equal(t, true, standing["available"])
	for _, m := range standing["members"].([]any) {
		assert.NotContains(t, m.(map[string]any), "accuracy",
			"the standing publishes an order, not a measurement")
	}

	profile := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/members/"+otherID+"/profile", nil, ownerToken)
	require.Equal(t, http.StatusOK, profile.StatusCode)
	assert.NotContains(t, profile.Body["data"].(map[string]any)["stats"].(map[string]any), "guess_accuracy")
}
