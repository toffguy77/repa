package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func anticipationOf(t *testing.T, token, seasonID string) jsonResponse {
	t.Helper()
	return doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/anticipation", nil, token)
}

func TestAnticipation_CountsVotersAboutMeNotMyOwnVotes(t *testing.T) {
	suffix := uuid.New().String()[:6]
	_, aToken := createTestUser(t, "ant_a_"+suffix)
	group := createTestGroup(t, aToken, shortName("Anticipation"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	seasonID := activeSeason(t, aToken, group["id"].(string))["id"].(string)

	var others []string
	for i := 0; i < 2; i++ {
		_, tok := createTestUser(t, fmt.Sprintf("ant_%d_%s", i, suffix))
		joinGroup(t, tok, code)
		others = append(others, tok)
	}

	// Nobody has voted yet.
	resp := anticipationOf(t, aToken, seasonID)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	data := getData(t, resp)
	assert.Equal(t, float64(0), data["voters_about_me"])
	assert.NotEmpty(t, data["reveal_at"], "the countdown needs the reveal time")

	// A votes about others — that must not raise A's own count.
	completeVoting(t, aToken, seasonID)
	data = getData(t, anticipationOf(t, aToken, seasonID))
	assert.Equal(t, float64(0), data["voters_about_me"],
		"answering about others must not inflate your own count")

	// The others vote, which does raise it.
	for _, tok := range others {
		completeVoting(t, tok, seasonID)
	}
	data = getData(t, anticipationOf(t, aToken, seasonID))
	assert.Greater(t, data["voters_about_me"], float64(0))
}

func TestAnticipation_PayloadCarriesNoIdentityOrAttribute(t *testing.T) {
	suffix := uuid.New().String()[:6]
	_, aToken := createTestUser(t, "ant_p_"+suffix)
	group := createTestGroup(t, aToken, shortName("Anticipation Shape"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	seasonID := activeSeason(t, aToken, group["id"].(string))["id"].(string)

	for i := 0; i < 2; i++ {
		_, tok := createTestUser(t, fmt.Sprintf("ant_q%d_%s", i, suffix))
		joinGroup(t, tok, code)
		completeVoting(t, tok, seasonID)
	}

	data := getData(t, anticipationOf(t, aToken, seasonID))

	for key := range data {
		switch key {
		case "voters_about_me", "teaser_emoji", "reveal_at":
		default:
			t.Errorf("unexpected key %q: the payload must stay incapable of carrying identities", key)
		}
	}
}

func TestAnticipation_RefusedAfterTheReveal(t *testing.T) {
	suffix := uuid.New().String()[:6]
	_, aToken := createTestUser(t, "ant_r_"+suffix)
	group := createTestGroup(t, aToken, shortName("Anticipation Revealed"), []string{"FUNNY"})

	seasonID := createRevealedSeason(t, group["id"].(string))

	resp := anticipationOf(t, aToken, seasonID)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"a member with results does not need a partial signal: %s", string(resp.RawBody))
}

func TestAnticipation_RefusedForANonMember(t *testing.T) {
	suffix := uuid.New().String()[:6]
	_, ownerToken := createTestUser(t, "ant_o_"+suffix)
	group := createTestGroup(t, ownerToken, shortName("Anticipation Stranger"), []string{"FUNNY"})
	seasonID := activeSeason(t, ownerToken, group["id"].(string))["id"].(string)

	_, strangerToken := createTestUser(t, "ant_s_"+suffix)
	resp := anticipationOf(t, strangerToken, seasonID)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode, string(resp.RawBody))
}

// --- The signal debounce, against a real Redis ---

func TestAnticipation_SignalIsClaimedOncePerDay(t *testing.T) {
	userID := uuid.New().String()

	// The first claim of the day succeeds; the second does not. Without this bound a 20-person group
	// produces 19 notifications in an evening.
	if !suitePushService.ClaimDailySignal(ctxBackground(), userID) {
		t.Fatal("the first claim of the day should succeed")
	}
	if suitePushService.ClaimDailySignal(ctxBackground(), userID) {
		t.Error("the second claim on the same day must be suppressed")
	}

	// A different recipient is unaffected — the key is per user.
	if !suitePushService.ClaimDailySignal(ctxBackground(), uuid.New().String()) {
		t.Error("another member's signal should not be blocked by someone else's")
	}
}
