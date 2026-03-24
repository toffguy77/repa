package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfile_GetMemberProfile(t *testing.T) {
	uid1, tok1 := createTestUser(t, "prof_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "prof_u2_"+t.Name())

	group := createTestGroup(t, tok1, "Profile Group "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))

	// Get user2's profile as user1
	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/members/"+uid2+"/profile", nil, tok1)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = uid1
}

func TestProfile_NotMember(t *testing.T) {
	_, tok1 := createTestUser(t, "prof_nm_u1_"+t.Name())
	uid2, _ := createTestUser(t, "prof_nm_u2_"+t.Name())
	_, outsiderToken := createTestUser(t, "prof_nm_outsider_"+t.Name())

	group := createTestGroup(t, tok1, "Profile NM "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)

	// outsider tries to get profile in a group they're not in
	resp := doRequest(t, "GET", "/api/v1/groups/"+groupID+"/members/"+uid2+"/profile", nil, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
