package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"net/url"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	groupssvc "github.com/repa-app/repa/internal/service/groups"
)

// A pasted link is reduced to a bare code by the client (extractInviteCode) because the code
// travels as a URL path segment; the server-side handling of the URL form is covered by
// TestNormalizeInviteCode.

func inviteCodeOf(t *testing.T, group map[string]any) string {
	t.Helper()
	code, ok := group["invite_code"].(string)
	require.True(t, ok, "group has no invite_code: %v", group)
	return code
}

// --- 5.1: the short-code path ---

func TestInviteCodes_NewGroupGetsShortTranscribableCode(t *testing.T) {
	_, token := createTestUser(t, shortUsername("inv_shape"))

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "Invite Shape",
		"categories": []string{"FUNNY"},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(resp.RawBody))

	data := getData(t, resp)
	group := data["group"].(map[string]any)
	code := inviteCodeOf(t, group)

	assert.Len(t, code, groupssvc.InviteCodeLength)
	assert.NotContains(t, code, "0")
	assert.NotContains(t, code, "1")
	assert.NotContains(t, code, "O")
	assert.NotContains(t, code, "I")
	assert.NotContains(t, code, "L")
	assert.Equal(t, strings.ToUpper(code), code, "codes are stored upper-case")

	assert.Equal(t, "https://repa.app/join/"+code, data["invite_url"])
}

func TestInviteCodes_JoinAcceptsEveryInputForm(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("inv_forms"))
	group := createTestGroup(t, ownerToken, shortName("Invite Forms"), []string{"FUNNY"})
	code := inviteCodeOf(t, group)
	groupID := group["id"].(string)

	forms := map[string]string{
		"exact":      code,
		"lowercase":  strings.ToLower(code),
		"grouped":    code[:3] + " " + code[3:],
		"hyphenated": code[:3] + "-" + code[3:],
	}

	i := 0
	for name, form := range forms {
		i++
		t.Run(name, func(t *testing.T) {
			_, token := createTestUser(t, shortUsername("inv_f")+string(rune('a'+i)))

			// The preview resolves the same group...
			preview := doRequest(t, "GET",
				"/api/v1/groups/join/"+urlEscape(form)+"/preview", nil, token)
			require.Equal(t, http.StatusOK, preview.StatusCode,
				"preview for %q failed: %s", form, string(preview.RawBody))

			// ...and so does the join.
			join := doRequest(t, "POST", "/api/v1/groups/join/"+urlEscape(form), nil, token)
			require.Equal(t, http.StatusOK, join.StatusCode,
				"join with %q failed: %s", form, string(join.RawBody))
			joined := getData(t, join)["group"].(map[string]any)
			assert.Equal(t, groupID, joined["id"])
		})
	}
}

func TestInviteCodes_RegenerationRevokesTheOldCode(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("inv_regen"))
	group := createTestGroup(t, ownerToken, shortName("Invite Regen"), []string{"FUNNY"})
	oldCode := inviteCodeOf(t, group)
	groupID := group["id"].(string)

	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/invite-link", nil, ownerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	newURL := getData(t, resp)["invite_url"].(string)
	newCode := strings.TrimPrefix(newURL, "https://repa.app/join/")

	assert.Len(t, newCode, groupssvc.InviteCodeLength)
	assert.NotEqual(t, oldCode, newCode)

	// The old code is gone.
	_, strangerToken := createTestUser(t, shortUsername("inv_stranger"))
	stale := doRequest(t, "POST", "/api/v1/groups/join/"+oldCode, nil, strangerToken)
	assert.Equal(t, http.StatusNotFound, stale.StatusCode,
		"the revoked code must not join: %s", string(stale.RawBody))

	// The new one works.
	fresh := doRequest(t, "POST", "/api/v1/groups/join/"+newCode, nil, strangerToken)
	assert.Equal(t, http.StatusOK, fresh.StatusCode, string(fresh.RawBody))
}

func TestInviteCodes_UnknownCodeIsNotFound(t *testing.T) {
	_, token := createTestUser(t, shortUsername("inv_unknown"))

	resp := doRequest(t, "POST", "/api/v1/groups/join/ZZZZZZ", nil, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, string(resp.RawBody))
}

// --- 5.2: legacy UUID codes ---

func TestInviteCodes_LegacyUUIDCodeStillJoins(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("inv_legacy"))
	group := createTestGroup(t, ownerToken, shortName("Invite Legacy"), []string{"FUNNY"})
	groupID := group["id"].(string)

	// Put the group back on a pre-migration code, the way an existing row looks.
	legacy := uuid.New().String()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`UPDATE groups SET invite_code = $1 WHERE id = $2`, legacy, groupID)
	require.NoError(t, err)

	_, token := createTestUser(t, shortUsername("inv_legacy_j"))

	preview := doRequest(t, "GET", "/api/v1/groups/join/"+legacy+"/preview", nil, token)
	require.Equal(t, http.StatusOK, preview.StatusCode,
		"a legacy code must still resolve: %s", string(preview.RawBody))

	join := doRequest(t, "POST", "/api/v1/groups/join/"+legacy, nil, token)
	require.Equal(t, http.StatusOK, join.StatusCode, string(join.RawBody))

	// Regenerating migrates the group onto the short format.
	resp := doRequest(t, "POST", "/api/v1/groups/"+groupID+"/invite-link", nil, ownerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	newCode := strings.TrimPrefix(getData(t, resp)["invite_url"].(string), "https://repa.app/join/")
	assert.Len(t, newCode, groupssvc.InviteCodeLength)
}

// urlEscape makes a code safe to put in a path segment, so forms containing spaces, slashes
// or query strings still exercise the server's normalisation rather than the router.
func urlEscape(s string) string {
	return url.PathEscape(s)
}
