package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRevealScenario(t *testing.T) (seasonID string, userIDs []string, tokens []string, questionIDs []string) {
	t.Helper()

	uid1, tok1 := createTestUser(t, "reveal_u1_"+t.Name())
	uid2, tok2 := createTestUser(t, "reveal_u2_"+t.Name())
	uid3, tok3 := createTestUser(t, "reveal_u3_"+t.Name())

	group := createTestGroup(t, tok1, "Reveal Group "+t.Name(), []string{"FUNNY"})
	groupID := group["id"].(string)
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID = createRevealedSeason(t, groupID)
	questionIDs = addSeasonQuestions(t, seasonID, 3)

	// Create season results
	createSeasonResults(t, seasonID, uid1, questionIDs)
	createSeasonResults(t, seasonID, uid2, questionIDs)
	createSeasonResults(t, seasonID, uid3, questionIDs)

	return seasonID, []string{uid1, uid2, uid3}, []string{tok1, tok2, tok3}, questionIDs
}

func TestReveal_GetReveal(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.Body["data"])
}

func TestReveal_GetMembersCards(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/members-cards", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	members := data["members"]
	assert.NotNil(t, members)
}

func TestReveal_SeasonNotRevealed(t *testing.T) {
	_, tok1 := createTestUser(t, "reveal_notr_u1_"+t.Name())
	_, tok2 := createTestUser(t, "reveal_notr_u2_"+t.Name())
	_, tok3 := createTestUser(t, "reveal_notr_u3_"+t.Name())

	group := createTestGroup(t, tok1, "Not Revealed "+t.Name(), []string{"FUNNY"})
	joinGroup(t, tok2, group["invite_code"].(string))
	joinGroup(t, tok3, group["invite_code"].(string))

	seasonID := createSeasonDirectly(t, group["id"].(string)) // VOTING status

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, tok1)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "SEASON_NOT_REVEALED", errObj["code"])
}

func TestReveal_NotMember(t *testing.T) {
	seasonID, _, _, _ := setupRevealScenario(t)
	_, outsiderToken := createTestUser(t, "reveal_outsider_"+t.Name())

	resp := doRequest(t, "GET", "/api/v1/seasons/"+seasonID+"/reveal", nil, outsiderToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestReveal_BuyDetector_Success(t *testing.T) {
	seasonID, userIDs, tokens, _ := setupRevealScenario(t)
	addCrystals(t, userIDs[0], 20)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReveal_BuyDetector_InsufficientFunds(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	// No crystals added — should fail
	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/detector", nil, tokens[0])
	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "INSUFFICIENT_FUNDS", errObj["code"])
}

func TestReveal_OpenHidden_Success(t *testing.T) {
	seasonID, userIDs, tokens, _ := setupRevealScenario(t)
	addCrystals(t, userIDs[0], 10)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/reveal/open-hidden", nil, tokens[0])
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReveal_OpenHidden_InsufficientFunds(t *testing.T) {
	seasonID, _, tokens, _ := setupRevealScenario(t)

	resp := doRequest(t, "POST", "/api/v1/seasons/"+seasonID+"/reveal/open-hidden", nil, tokens[0])
	assert.Equal(t, http.StatusPaymentRequired, resp.StatusCode)
}
