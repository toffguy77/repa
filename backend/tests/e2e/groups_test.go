package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroups_Create(t *testing.T) {
	_, token := createTestUser(t, "grp_creator")

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "Test Group",
		"categories": []string{"FUNNY", "SKILLS"},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	data := getData(t, resp)
	group := data["group"].(map[string]any)
	assert.Equal(t, "Test Group", group["name"])
	assert.NotEmpty(t, group["id"])
	assert.NotEmpty(t, group["invite_code"])
	assert.NotEmpty(t, data["invite_url"])
}

func TestGroups_Create_InvalidName(t *testing.T) {
	_, token := createTestUser(t, "grp_bad_name")

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "ab", // too short
		"categories": []string{"FUNNY"},
	}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGroups_Create_NoCategories(t *testing.T) {
	_, token := createTestUser(t, "grp_no_cat")

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "Valid Name",
		"categories": []string{},
	}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGroups_ListGroups(t *testing.T) {
	_, token := createTestUser(t, "grp_lister")

	createTestGroup(t, token, "List Group 1", []string{"FUNNY"})
	createTestGroup(t, token, "List Group 2", []string{"HOT"})

	resp := doRequest(t, "GET", "/api/v1/groups", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	groups := data["groups"].([]any)
	assert.GreaterOrEqual(t, len(groups), 2)
}

func TestGroups_GetGroup(t *testing.T) {
	_, token := createTestUser(t, "grp_getter")

	group := createTestGroup(t, token, "Get Group", []string{"SKILLS"})
	groupID := group["id"].(string)

	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	grp := data["group"].(map[string]any)
	assert.Equal(t, "Get Group", grp["name"])
	members := data["members"].([]any)
	assert.Len(t, members, 1) // creator is the only member
}

func TestGroups_JoinPreview(t *testing.T) {
	_, creatorToken := createTestUser(t, "grp_preview_creator")
	group := createTestGroup(t, creatorToken, "Preview Group", []string{"FUNNY"})

	resp := doRequest(t, "GET", "/api/v1/groups/join/"+group["invite_code"].(string)+"/preview", nil, creatorToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, "Preview Group", data["name"])
	assert.Equal(t, float64(1), data["member_count"])
}

func TestGroups_JoinGroup(t *testing.T) {
	_, creatorToken := createTestUser(t, "grp_join_creator")
	_, joinerToken := createTestUser(t, "grp_joiner")

	group := createTestGroup(t, creatorToken, "Join Group", []string{"FUNNY"})
	inviteCode := group["invite_code"].(string)

	resp := doRequest(t, "POST", "/api/v1/groups/join/"+inviteCode, nil, joinerToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	grp := data["group"].(map[string]any)
	assert.Equal(t, "Join Group", grp["name"])
}

func TestGroups_JoinGroup_AlreadyMember(t *testing.T) {
	_, token := createTestUser(t, "grp_already")
	group := createTestGroup(t, token, "Already Group", []string{"FUNNY"})

	resp := doRequest(t, "POST", "/api/v1/groups/join/"+group["invite_code"].(string), nil, token)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "ALREADY_MEMBER", errObj["code"])
}

func TestGroups_LeaveGroup(t *testing.T) {
	_, creatorToken := createTestUser(t, "grp_leave_creator")
	_, leaverToken := createTestUser(t, "grp_leaver")

	group := createTestGroup(t, creatorToken, "Leave Group", []string{"FUNNY"})
	joinGroup(t, leaverToken, group["invite_code"].(string))

	groupID := group["id"].(string)
	resp := doRequest(t, "DELETE", "/api/v1/groups/"+groupID+"/leave", nil, leaverToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, true, data["left"])
}

func TestGroups_UpdateGroup_Admin(t *testing.T) {
	_, token := createTestUser(t, "grp_updater")
	group := createTestGroup(t, token, "Update Group", []string{"FUNNY"})

	resp := doRequest(t, "PATCH", "/api/v1/groups/"+group["id"].(string), map[string]any{
		"name": "Updated Group Name",
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	grp := data["group"].(map[string]any)
	assert.Equal(t, "Updated Group Name", grp["name"])
}

func TestGroups_UpdateGroup_NotAdmin(t *testing.T) {
	_, creatorToken := createTestUser(t, "grp_upd_creator")
	_, memberToken := createTestUser(t, "grp_upd_member")

	group := createTestGroup(t, creatorToken, "Admin Only", []string{"FUNNY"})
	joinGroup(t, memberToken, group["invite_code"].(string))

	resp := doRequest(t, "PATCH", "/api/v1/groups/"+group["id"].(string), map[string]any{
		"name": "Hacked Name",
	}, memberToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "NOT_ADMIN", errObj["code"])
}

func TestGroups_RegenerateInviteLink(t *testing.T) {
	_, token := createTestUser(t, "grp_regen")
	group := createTestGroup(t, token, "Regen Group", []string{"FUNNY"})
	oldCode := group["invite_code"].(string)

	resp := doRequest(t, "POST", "/api/v1/groups/"+group["id"].(string)+"/invite-link", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.NotEmpty(t, data["invite_url"])
	assert.NotEqual(t, oldCode, data["invite_url"]) // URL should contain a new code
}

func TestGroups_MaxGroupsPerUser(t *testing.T) {
	_, token := createTestUser(t, "grp_max_user")

	// Create 10 groups (max allowed)
	for i := 0; i < 10; i++ {
		createTestGroup(t, token, fmt.Sprintf("Max Group %d", i), []string{"FUNNY"})
	}

	// 11th should fail
	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "One Too Many",
		"categories": []string{"FUNNY"},
	}, token)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "GROUP_LIMIT", errObj["code"])
}

func TestGroups_RomanceBlocked_Under18(t *testing.T) {
	// User born in 2012 is under 18 in 2026
	_, token := createTestUserWithBirthYear(t, "grp_underage", 2012)

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       "Romance Group",
		"categories": []string{"ROMANCE"},
	}, token)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "ROMANCE_BLOCKED", errObj["code"])
}

func TestGroups_NotMember_GetGroup(t *testing.T) {
	_, creatorToken := createTestUser(t, "grp_notmem_creator")
	_, outsiderToken := createTestUser(t, "grp_outsider")

	group := createTestGroup(t, creatorToken, "Private Group", []string{"FUNNY"})

	resp := doRequest(t, "GET", "/api/v1/groups/"+group["id"].(string), nil, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
