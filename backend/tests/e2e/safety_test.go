package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safetyGroup builds a group with an admin and three other members, each with a unique username.
func safetyGroup(t *testing.T, label string) (groupID, code string, tokens map[string]string, ids map[string]string) {
	t.Helper()
	suffix := uuid.New().String()[:6]

	tokens = map[string]string{}
	ids = map[string]string{}

	adminID, adminToken := createTestUser(t, fmt.Sprintf("%s_ad_%s", label, suffix))
	tokens["admin"] = adminToken
	ids["admin"] = adminID

	group := createTestGroup(t, adminToken, shortName("Safety "+label), []string{"FUNNY"})
	groupID = group["id"].(string)
	code = group["invite_code"].(string)

	for _, name := range []string{"a", "b", "c"} {
		id, tok := createTestUser(t, fmt.Sprintf("%s_%s_%s", label, name, suffix))
		joinGroup(t, tok, code)
		tokens[name] = tok
		ids[name] = id
	}
	return groupID, code, tokens, ids
}

func votingTargets(t *testing.T, token, seasonID string) map[string]bool {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	out := map[string]bool{}
	for _, raw := range getData(t, resp)["targets"].([]any) {
		out[raw.(map[string]any)["user_id"].(string)] = true
	}
	return out
}

// --- 7.1: blocks and voting ---

func TestSafety_BlockRemovesTheMemberFromTargetsBothWays(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "blk")
	seasonID := activeSeason(t, tokens["admin"], groupID)["id"].(string)

	// a blocks b.
	resp := doRequest(t, "POST", "/api/v1/members/"+ids["b"]+"/block", nil, tokens["a"])
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	aTargets := votingTargets(t, tokens["a"], seasonID)
	assert.False(t, aTargets[ids["b"]], "a blocked b, so b must not be offered to a")
	assert.True(t, aTargets[ids["c"]], "an unblocked member should still be offered")

	// The effect is symmetric even though only a created it.
	bTargets := votingTargets(t, tokens["b"], seasonID)
	assert.False(t, bTargets[ids["a"]], "a member who blocked me must not be offered either")

	// And the vote endpoint agrees with the target list.
	session := getData(t, doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, tokens["a"]))
	questionID := session["questions"].([]any)[0].(map[string]any)["question_id"].(string)

	vote := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionID,
		"target_id":   ids["b"],
	}, tokens["a"])
	assert.Equal(t, http.StatusBadRequest, vote.StatusCode,
		"the vote endpoint must refuse what the target list excluded: %s", string(vote.RawBody))
}

func TestSafety_VotesCastBeforeABlockAreKept(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "keep")
	seasonID := activeSeason(t, tokens["admin"], groupID)["id"].(string)

	session := getData(t, doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/voting-session", nil, tokens["a"]))
	questionID := session["questions"].([]any)[0].(map[string]any)["question_id"].(string)

	vote := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/votes", map[string]any{
		"question_id": questionID,
		"target_id":   ids["b"],
	}, tokens["a"])
	require.Equal(t, http.StatusCreated, vote.StatusCode, string(vote.RawBody))

	// Now block, and confirm the earlier vote survived: removing it would shift b's card at the moment
	// of the block and signal that one happened.
	require.Equal(t, http.StatusOK,
		doRequest(t, "POST", "/api/v1/members/"+ids["b"]+"/block", nil, tokens["a"]).StatusCode)

	var count int
	require.NoError(t, suite.sqlDB.QueryRowContext(ctxBackground(),
		`SELECT COUNT(*) FROM votes WHERE season_id = $1 AND voter_id = $2 AND target_id = $3`,
		seasonID, ids["a"], ids["b"]).Scan(&count))
	assert.Equal(t, 1, count, "the earlier vote must survive the block")
}

func TestSafety_UnblockRestoresTheTarget(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "unblk")
	seasonID := activeSeason(t, tokens["admin"], groupID)["id"].(string)

	require.Equal(t, http.StatusOK,
		doRequest(t, "POST", "/api/v1/members/"+ids["b"]+"/block", nil, tokens["a"]).StatusCode)
	assert.False(t, votingTargets(t, tokens["a"], seasonID)[ids["b"]])

	require.Equal(t, http.StatusOK,
		doRequest(t, "DELETE", "/api/v1/members/"+ids["b"]+"/block", nil, tokens["a"]).StatusCode)
	assert.True(t, votingTargets(t, tokens["a"], seasonID)[ids["b"]],
		"unblocking should restore the previous state")
}

// --- 7.2: removal and re-joining ---

func TestSafety_RemovedMemberCannotRejoinEvenWithANewInvite(t *testing.T) {
	groupID, code, tokens, ids := safetyGroup(t, "rm")

	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/members/"+ids["a"], nil, tokens["admin"])
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	// The old invite is refused.
	rejoin := doRequest(t, "POST", "/api/v1/groups/join/"+code, nil, tokens["a"])
	assert.Equal(t, http.StatusForbidden, rejoin.StatusCode,
		"a removal must not be undone by a link: %s", string(rejoin.RawBody))

	// And so is a freshly regenerated one — the ban is on the person and the group, not the code.
	regen := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/invite-link", nil, tokens["admin"])
	require.Equal(t, http.StatusOK, regen.StatusCode)
	newCode := getData(t, regen)["invite_url"].(string)[len("https://repa.app/join/"):]

	rejoin = doRequest(t, "POST", "/api/v1/groups/join/"+newCode, nil, tokens["a"])
	assert.Equal(t, http.StatusForbidden, rejoin.StatusCode,
		"regenerating the invite must not undo a removal: %s", string(rejoin.RawBody))

	// Someone who merely left can come back.
	require.Equal(t, http.StatusOK,
		doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/leave", nil, tokens["b"]).StatusCode)
	back := doRequest(t, "POST", "/api/v1/groups/join/"+newCode, nil, tokens["b"])
	assert.Equal(t, http.StatusOK, back.StatusCode,
		"ordinary leaving stays reversible: %s", string(back.RawBody))
}

func TestSafety_LeavingPermanentlyIsIrreversible(t *testing.T) {
	groupID, code, tokens, _ := safetyGroup(t, "perm")

	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/leave?permanent=true", nil, tokens["a"])
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	assert.Equal(t, true, getData(t, resp)["permanent"])

	rejoin := doRequest(t, "POST", "/api/v1/groups/join/"+code, nil, tokens["a"])
	assert.Equal(t, http.StatusForbidden, rejoin.StatusCode, string(rejoin.RawBody))
}

func TestSafety_OnlyAdminCanRemove(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "auth")

	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/members/"+ids["b"], nil, tokens["a"])
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, string(resp.RawBody))

	// b is still a member.
	var count int
	require.NoError(t, suite.sqlDB.QueryRowContext(ctxBackground(),
		`SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2`,
		groupID, ids["b"]).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestSafety_BanIsPerGroup(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "scope")

	require.Equal(t, http.StatusOK,
		doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/members/"+ids["a"], nil, tokens["admin"]).StatusCode)

	// A different group with its own invite is unaffected.
	other := createTestGroup(t, tokens["admin"], shortName("Safety Other"), []string{"FUNNY"})
	join := doRequest(t, "POST", "/api/v1/groups/join/"+other["invite_code"].(string), nil, tokens["a"])
	assert.Equal(t, http.StatusOK, join.StatusCode,
		"a ban is per group: %s", string(join.RawBody))
}

// --- 7.3: reporting ---

func TestSafety_ReportReachesTheAdminQueueSilently(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "rep")

	resp := doRequest(t, "POST", "/api/v1/members/"+ids["b"]+"/report", map[string]any{
		"group_id": groupID,
		"reason":   "травит в чате",
	}, tokens["a"])
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	queue := doRequestBasicAuth(t, "GET", "/api/v1/admin/user-reports", nil,
		testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, queue.StatusCode, string(queue.RawBody))

	reports := getData(t, queue)["reports"].([]any)
	found := false
	for _, raw := range reports {
		r := raw.(map[string]any)
		if r["reported_id"] == ids["b"] && r["reporter_id"] == ids["a"] {
			found = true
			assert.Equal(t, "травит в чате", r["reason"])
		}
	}
	assert.True(t, found, "the report should reach the queue with both identities and the reason")

	// Reporting must not block.
	seasonID := activeSeason(t, tokens["admin"], groupID)["id"].(string)
	assert.True(t, votingTargets(t, tokens["a"], seasonID)[ids["b"]],
		"reporting is not blocking")
}

func TestSafety_ReportIsIdempotent(t *testing.T) {
	groupID, _, tokens, ids := safetyGroup(t, "repi")

	for i := 0; i < 2; i++ {
		resp := doRequest(t, "POST", "/api/v1/members/"+ids["b"]+"/report", map[string]any{
			"group_id": groupID,
			"reason":   fmt.Sprintf("раз %d", i),
		}, tokens["a"])
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	}

	var count int
	require.NoError(t, suite.sqlDB.QueryRowContext(ctxBackground(),
		`SELECT COUNT(*) FROM user_reports WHERE reported_id = $1 AND reporter_id = $2`,
		ids["b"], ids["a"]).Scan(&count))
	assert.Equal(t, 1, count, "one report per person per reporter")
}
