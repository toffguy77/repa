package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crystalssvc "github.com/repa-app/repa/internal/service/crystals"
)

// registerViaOTP walks the real registration flow, which is what triggers the welcome grant —
// createTestUser inserts straight into the database and deliberately bypasses it.
func registerViaOTP(t *testing.T) (userID, token string) {
	t.Helper()
	phone := fmt.Sprintf("+7901%07d", uuid.New().ID()%10000000)

	send := doRequest(t, "POST", "/api/v1/auth/otp/send", map[string]any{"phone": phone}, "")
	require.Equal(t, http.StatusOK, send.StatusCode, string(send.RawBody))
	code := getData(t, send)["code"]
	require.NotEmpty(t, code, "dev mode should return the OTP code")

	verify := doRequest(t, "POST", "/api/v1/auth/otp/verify", map[string]any{
		"phone": phone,
		"code":  code,
	}, "")
	require.Equal(t, http.StatusOK, verify.StatusCode, string(verify.RawBody))

	data := getData(t, verify)
	token = data["token"].(string)
	user := data["user"].(map[string]any)
	return user["id"].(string), token
}

func balanceOf(t *testing.T, userID string) int {
	t.Helper()
	var balance int
	err := suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COALESCE(SUM(delta), 0) FROM crystal_logs WHERE user_id = $1`, userID).Scan(&balance)
	require.NoError(t, err)
	return balance
}

func grantCountOf(t *testing.T, userID, externalID string) int {
	t.Helper()
	var count int
	err := suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM crystal_logs WHERE user_id = $1 AND external_id = $2`,
		userID, externalID).Scan(&count)
	require.NoError(t, err)
	return count
}

// --- 6.1: the welcome grant ---

func TestCrystalGrants_NewUserCanAffordOneDetector(t *testing.T) {
	userID, token := registerViaOTP(t)

	assert.Equal(t, crystalssvc.WelcomeGrant, balanceOf(t, userID),
		"a new user should start with exactly one detector's worth")
	assert.Equal(t, 10, crystalssvc.WelcomeGrant,
		"the welcome grant is meant to equal one detector")

	// The grant is visible to the user, with a reason.
	resp := doRequest(t, "GET", "/api/v1/crystals/history", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	entries := getData(t, resp)["entries"].([]any)
	require.Len(t, entries, 1)

	entry := entries[0].(map[string]any)
	assert.Equal(t, true, entry["is_grant"])
	assert.NotEmpty(t, entry["reason"])
}

func TestCrystalGrants_WelcomeIsPaidOnce(t *testing.T) {
	userID, _ := registerViaOTP(t)

	// Logging in again must not top the balance up.
	assert.Equal(t, 1, grantCountOf(t, userID, "welcome:"+userID))
	assert.Equal(t, crystalssvc.WelcomeGrant, balanceOf(t, userID))
}

// --- 6.2 / 6.3: referral ---

// completeVotingFor answers every question of a season for one member.
func completeVotingFor(t *testing.T, token, seasonID string) {
	t.Helper()
	completeVoting(t, token, seasonID)
}

func TestCrystalGrants_ReferralPaysAfterTheInviteeVotes(t *testing.T) {
	inviterID, inviterToken := registerViaOTP(t)
	group := createTestGroup(t, inviterToken, shortName("Referral"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	seasonID := activeSeason(t, inviterToken, group["id"].(string))["id"].(string)

	// Two more members, so the group can actually produce a season worth voting in.
	_, cToken := registerViaOTP(t)
	joinGroup(t, cToken, code)

	inviteeID, inviteeToken := registerViaOTP(t)
	resp := doRequest(t, "POST",
		"/api/v1/groups/join/"+code+"?source=CARD&ref="+inviterID, nil, inviteeToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	before := balanceOf(t, inviterID)

	// Joining alone pays nothing.
	assert.Equal(t, before, balanceOf(t, inviterID))
	assert.Equal(t, 0, grantCountOf(t, inviterID, "referral:"+inviteeID))

	// Completing the first session pays.
	completeVotingFor(t, inviteeToken, seasonID)

	assert.Equal(t, 1, grantCountOf(t, inviterID, "referral:"+inviteeID),
		"the inviter should be paid once the invitee plays")
	assert.Equal(t, before+crystalssvc.ReferralGrant, balanceOf(t, inviterID))

	// A further session must not pay again — the grant is keyed on the invitee.
	completeVotingFor(t, inviteeToken, seasonID)
	assert.Equal(t, 1, grantCountOf(t, inviterID, "referral:"+inviteeID))
	assert.Equal(t, before+crystalssvc.ReferralGrant, balanceOf(t, inviterID))
}

func TestCrystalGrants_InvalidReferrerIsDroppedAndNobodyIsPaid(t *testing.T) {
	ownerID, ownerToken := registerViaOTP(t)
	group := createTestGroup(t, ownerToken, shortName("Referral Bad"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	seasonID := activeSeason(t, ownerToken, group["id"].(string))["id"].(string)

	_, cToken := registerViaOTP(t)
	joinGroup(t, cToken, code)

	// A referrer who is not a member of this group.
	strangerID, _ := registerViaOTP(t)
	strangerBefore := balanceOf(t, strangerID)

	inviteeID, inviteeToken := registerViaOTP(t)
	resp := doRequest(t, "POST",
		"/api/v1/groups/join/"+code+"?source=CARD&ref="+strangerID, nil, inviteeToken)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"an unverifiable referrer must not cost the join: %s", string(resp.RawBody))

	var invitedBy *string
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT invited_by FROM group_members WHERE user_id = $1 AND group_id = $2`,
		inviteeID, group["id"]).Scan(&invitedBy))
	assert.Nil(t, invitedBy, "an invalid referrer should be dropped, not stored")

	completeVotingFor(t, inviteeToken, seasonID)

	assert.Equal(t, strangerBefore, balanceOf(t, strangerID), "a stranger must not be paid")
	assert.Equal(t, 0, grantCountOf(t, ownerID, "referral:"+inviteeID))
}

func TestCrystalGrants_SelfReferralPaysNothing(t *testing.T) {
	ownerID, ownerToken := registerViaOTP(t)
	group := createTestGroup(t, ownerToken, shortName("Referral Self"), []string{"FUNNY"})
	code := group["invite_code"].(string)

	inviteeID, inviteeToken := registerViaOTP(t)
	resp := doRequest(t, "POST",
		"/api/v1/groups/join/"+code+"?ref="+inviteeID, nil, inviteeToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	var invitedBy *string
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT invited_by FROM group_members WHERE user_id = $1 AND group_id = $2`,
		inviteeID, group["id"]).Scan(&invitedBy))
	assert.Nil(t, invitedBy, "a self-referral is not a referral")
	assert.Equal(t, 0, grantCountOf(t, ownerID, "referral:"+inviteeID))
}

func TestCrystalGrants_NoReferrerPaysNobody(t *testing.T) {
	_, ownerToken := registerViaOTP(t)
	group := createTestGroup(t, ownerToken, shortName("Referral None"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	seasonID := activeSeason(t, ownerToken, group["id"].(string))["id"].(string)

	_, cToken := registerViaOTP(t)
	joinGroup(t, cToken, code)

	inviteeID, inviteeToken := registerViaOTP(t)
	joinGroup(t, inviteeToken, code)
	completeVotingFor(t, inviteeToken, seasonID)

	var referralGrants int
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM crystal_logs WHERE external_id = $1`,
		"referral:"+inviteeID).Scan(&referralGrants))
	assert.Zero(t, referralGrants, "an unattributed join pays nobody")
}
