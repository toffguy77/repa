package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/repa-app/repa/internal/service/cards"
)

func joinSourceOf(t *testing.T, userID, groupID string) string {
	t.Helper()
	var source string
	err := suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT join_source FROM group_members WHERE user_id = $1 AND group_id = $2`,
		userID, groupID).Scan(&source)
	require.NoError(t, err)
	return source
}

// --- 6.2: join attribution ---

func TestAttribution_JoinRecordsEachSource(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("att_owner"))
	group := createTestGroup(t, ownerToken, shortName("Attribution"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	groupID := group["id"].(string)

	cases := map[string]string{
		"LINK":     "LINK",
		"CODE":     "CODE",
		"CARD":     "CARD",
		"TELEGRAM": "TELEGRAM",
	}

	i := 0
	for source, want := range cases {
		i++
		t.Run(source, func(t *testing.T) {
			userID, token := createTestUser(t, shortUsername("att_")+string(rune('a'+i)))

			resp := doRequest(t, "POST",
				"/api/v1/groups/join/"+code+"?source="+source, nil, token)
			require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

			assert.Equal(t, want, joinSourceOf(t, userID, groupID))
		})
	}
}

func TestAttribution_MissingSourceIsUnknownAndStillJoins(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("att_miss_o"))
	group := createTestGroup(t, ownerToken, shortName("Attribution Missing"), []string{"FUNNY"})

	userID, token := createTestUser(t, shortUsername("att_miss_u"))
	resp := doRequest(t, "POST", "/api/v1/groups/join/"+group["invite_code"].(string), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	assert.Equal(t, "UNKNOWN", joinSourceOf(t, userID, group["id"].(string)))
}

func TestAttribution_UnrecognisedSourceDoesNotCostTheJoin(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("att_bad_o"))
	group := createTestGroup(t, ownerToken, shortName("Attribution Bad"), []string{"FUNNY"})

	userID, token := createTestUser(t, shortUsername("att_bad_u"))
	resp := doRequest(t, "POST",
		"/api/v1/groups/join/"+group["invite_code"].(string)+"?source=instagram-reels", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"a telemetry field must never cost an acquisition: %s", string(resp.RawBody))

	assert.Equal(t, "UNKNOWN", joinSourceOf(t, userID, group["id"].(string)))
}

func TestAttribution_FounderIsRecorded(t *testing.T) {
	userID, token := createTestUser(t, shortUsername("att_founder"))
	group := createTestGroup(t, token, shortName("Attribution Founder"), []string{"FUNNY"})

	// The founder did not arrive through an invite; UNKNOWN is the honest value.
	assert.Equal(t, "UNKNOWN", joinSourceOf(t, userID, group["id"].(string)))
}

// --- 6.3: share events and the admin funnel ---

func TestAttribution_SharesAreRecordedAndReported(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("sh_owner"))
	_, bToken := createTestUser(t, shortUsername("sh_b"))
	_, cToken := createTestUser(t, shortUsername("sh_c"))
	_, dToken := createTestUser(t, shortUsername("sh_d"))
	_, eToken := createTestUser(t, shortUsername("sh_e"))

	group := createTestGroup(t, ownerToken, shortName("Shares"), []string{"FUNNY"})
	code := group["invite_code"].(string)
	for _, tok := range []string{bToken, cToken, dToken, eToken} {
		joinGroup(t, tok, code)
	}

	seasonID := createRevealedSeason(t, group["id"].(string))

	// Two card shares and one Telegram share.
	for _, channel := range []string{"card", "card", "telegram"} {
		resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/shares",
			map[string]any{"channel": channel}, ownerToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	}

	var cardCount, tgCount int
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM share_events WHERE season_id = $1 AND channel = 'card'`,
		seasonID).Scan(&cardCount))
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM share_events WHERE season_id = $1 AND channel = 'telegram'`,
		seasonID).Scan(&tgCount))

	assert.Equal(t, 2, cardCount, "each share action should be recorded")
	assert.Equal(t, 1, tgCount)

	// Admin stats report the funnel.
	stats := doRequestBasicAuth(t, "GET", "/api/v1/admin/stats", nil,
		testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, stats.StatusCode, string(stats.RawBody))
	data := getData(t, stats)

	shares := data["shares_by_channel"].(map[string]any)
	assert.GreaterOrEqual(t, shares["card"], float64(2))

	joins := data["joins_by_source"].(map[string]any)
	assert.Contains(t, joins, "UNKNOWN")
}

func TestAttribution_UnknownShareChannelIsRecordedAsOther(t *testing.T) {
	_, token := createTestUser(t, shortUsername("sh_other"))
	group := createTestGroup(t, token, shortName("Shares Other"), []string{"FUNNY"})
	seasonID := createRevealedSeason(t, group["id"].(string))

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/shares",
		map[string]any{"channel": "instagram"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	var count int
	require.NoError(t, suite.sqlDB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM share_events WHERE season_id = $1 AND channel = 'other'`,
		seasonID).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestAttribution_ShareRequiresMembershipAndARevealedSeason(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("sh_guard_o"))
	group := createTestGroup(t, ownerToken, shortName("Shares Guard"), []string{"FUNNY"})

	votingSeason := createSeasonDirectly(t, group["id"].(string))
	resp := doRequest(t, "POST", "/api/v1/seasons/"+votingSeason+"/shares",
		map[string]any{"channel": "card"}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
		"there is no card to share before the reveal")

	revealed := createRevealedSeason(t, group["id"].(string))
	_, strangerToken := createTestUser(t, shortUsername("sh_guard_s"))
	resp = doRequest(t, "POST", "/api/v1/seasons/"+revealed+"/shares",
		map[string]any{"channel": "card"}, strangerToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"a non-member must not record against someone else's season")
}

// --- 6.1: the card carries the invitation ---

func TestAttribution_GeneratedCardCarriesTheInvitation(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("card_inv"))
	group := createTestGroup(t, ownerToken, shortName("Card Invite"), []string{"FUNNY"})
	code := group["invite_code"].(string)

	// Render the card HTML the way the card worker does, without needing a browser.
	html := cards.BuildCardHTML(cards.CardData{
		Username:        "alice",
		GroupName:       group["name"].(string),
		SeasonNumber:    1,
		InviteCode:      code,
		InviteQRDataURI: cards.InviteQRDataURI("https://repa.app/join/" + code + "?s=card"),
		TopAttributes: []cards.CardAttribute{
			{QuestionText: "Кто первым побежит при пожаре?", Percentage: 67},
		},
	})

	assert.Contains(t, html, code[:3]+" "+code[3:],
		"the card must show this group's code")
	assert.Contains(t, html, "data:image/png;base64,",
		"the card must embed a scannable code")
	assert.Contains(t, html, "Скачай Репу и введи код",
		"the card must tell a viewer what to do")
}
