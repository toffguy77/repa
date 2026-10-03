package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seasonQuestionTones returns the tones of the questions assigned to a group's first season, read
// straight from the database — the API never exposes tone, and the thing under test is which questions
// the selection query chose.
func seasonQuestionTones(t *testing.T, groupID string) []string {
	t.Helper()
	rows, err := suite.sqlDB.QueryContext(context.Background(), `
		SELECT q.tone
		FROM season_questions sq
		JOIN seasons s ON s.id = sq.season_id
		JOIN questions q ON q.id = sq.question_id
		WHERE s.group_id = $1`, groupID)
	require.NoError(t, err)
	defer rows.Close()

	var tones []string
	for rows.Next() {
		var tone string
		require.NoError(t, rows.Scan(&tone))
		tones = append(tones, tone)
	}
	require.NoError(t, rows.Err())
	return tones
}

func TestQuestionTone_KindOnlyGroupGetsNoEdgyQuestion(t *testing.T) {
	_, token := createTestUser(t, "tone_kind")

	// HOT is edgy end to end and FUNNY is not, so the group's bank contains both kinds — which is
	// what makes the assertion mean something rather than passing because nothing was edgy.
	group := createTestGroupWithKindOnly(t, token, "Kind Group", []string{"HOT", "FUNNY"}, true)
	tones := seasonQuestionTones(t, group["id"].(string))

	require.NotEmpty(t, tones, "a kind-only group must still get a season's worth of questions")
	for _, tone := range tones {
		assert.NotEqual(t, "EDGY", tone, "a kind-only group received an edgy question")
	}
}

func TestQuestionTone_OrdinaryGroupBankIsUnrestricted(t *testing.T) {
	_, token := createTestUser(t, "tone_any")

	// The counterpart to the test above, with the same categories: an ordinary group must still be
	// able to draw the edgy ones. Without this, a selection query that dropped every edgy question
	// unconditionally would pass the kind-only test.
	group := createTestGroupWithKindOnly(t, token, "Any Group", []string{"HOT"}, false)
	tones := seasonQuestionTones(t, group["id"].(string))

	require.NotEmpty(t, tones)
	assert.Contains(t, tones, "EDGY", "an ordinary HOT group should draw edgy questions")
}

func TestQuestionTone_UnderageCreatorGetsKindOnlyByDefault(t *testing.T) {
	_, token := createTestUserWithBirthYear(t, "tone_minor", time.Now().Year()-15)

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       shortName("Minor Group"),
		"categories": []string{"HOT", "FUNNY"},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(resp.RawBody))

	group := resp.Body["data"].(map[string]any)["group"].(map[string]any)
	assert.Equal(t, true, group["kind_only"], "a 15-year-old creator should get the setting on")
	for _, tone := range seasonQuestionTones(t, group["id"].(string)) {
		assert.NotEqual(t, "EDGY", tone)
	}
}

func TestQuestionTone_AdultCreatorDoesNot(t *testing.T) {
	_, token := createTestUserWithBirthYear(t, "tone_adult", time.Now().Year()-30)

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       shortName("Adult Group"),
		"categories": []string{"HOT"},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(resp.RawBody))

	group := resp.Body["data"].(map[string]any)["group"].(map[string]any)
	assert.Equal(t, false, group["kind_only"], "a 30-year-old creator should get the setting off")
}

func TestQuestionTone_ExplicitValueWinsOverTheAgeDefault(t *testing.T) {
	// A minor turning it off, and an adult turning it on: the default is a starting point, not a
	// restriction the member cannot see or change.
	_, minorToken := createTestUserWithBirthYear(t, "tone_minor_off", time.Now().Year()-15)
	minorGroup := createTestGroupWithKindOnly(t, minorToken, "Minor Off", []string{"HOT"}, false)
	assert.Equal(t, false, minorGroup["kind_only"])
	assert.Contains(t, seasonQuestionTones(t, minorGroup["id"].(string)), "EDGY")

	_, adultToken := createTestUserWithBirthYear(t, "tone_adult_on", time.Now().Year()-30)
	adultGroup := createTestGroupWithKindOnly(t, adultToken, "Adult On", []string{"FUNNY"}, true)
	assert.Equal(t, true, adultGroup["kind_only"])
}

func TestQuestionTone_AllEdgyCategoriesAreRefusedForAKindOnlyGroup(t *testing.T) {
	_, token := createTestUser(t, "tone_empty")

	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       shortName("Empty Bank"),
		"categories": []string{"HOT", "SECRETS"},
		"kind_only":  true,
	}, token)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, string(resp.RawBody))
	assert.Equal(t, "NO_KIND_CATEGORIES", getError(t, resp)["code"])
}

func TestQuestionTone_AdminCanChangeTheSettingAfterwards(t *testing.T) {
	_, token := createTestUser(t, "tone_switch")
	group := createTestGroupWithKindOnly(t, token, "Switch Me", []string{"HOT", "FUNNY"}, false)
	groupID := group["id"].(string)

	before := seasonQuestionTones(t, groupID)
	require.NotEmpty(t, before)

	resp := doRequest(t, "PATCH", "/api/v1/groups/"+groupID, map[string]any{"kind_only": true}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))
	updated := resp.Body["data"].(map[string]any)["group"].(map[string]any)
	assert.Equal(t, true, updated["kind_only"])

	// The open season is left exactly as members found it: rewriting it would change the questions
	// out from under anyone who already answered some, and their votes would point at questions no
	// longer in the season.
	assert.ElementsMatch(t, before, seasonQuestionTones(t, groupID),
		"changing the setting must not rewrite the open season")
}

func TestQuestionTone_TurningTheSettingOnIsRefusedWhenItWouldEmptyTheBank(t *testing.T) {
	_, token := createTestUser(t, "tone_switch_bad")
	group := createTestGroupWithKindOnly(t, token, "Edgy Only", []string{"HOT", "SECRETS"}, false)
	groupID := group["id"].(string)

	resp := doRequest(t, "PATCH", "/api/v1/groups/"+groupID, map[string]any{"kind_only": true}, token)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, string(resp.RawBody))
	assert.Equal(t, "NO_KIND_CATEGORIES", getError(t, resp)["code"])

	// And the setting really did stay off, rather than being applied before the refusal.
	get := doRequest(t, "GET", "/api/v1/groups/"+groupID, nil, token)
	require.Equal(t, http.StatusOK, get.StatusCode)
	assert.Equal(t, false, get.Body["data"].(map[string]any)["group"].(map[string]any)["kind_only"])
}
