package profile

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// profileStats drives GET /groups/:id/members/:userId/profile and returns the decoded `stats` object,
// so the assertion is about the JSON a client actually receives rather than about a Go struct.
func profileStats(t *testing.T, viewerID, viewedID string) map[string]any {
	t.Helper()
	mq := newProfileMock()
	mq.stats[viewedID+":g1"] = db.UserGroupStat{
		UserID:             viewedID,
		GroupID:            "g1",
		SeasonsPlayed:      4,
		VotingStreak:       3,
		MaxVotingStreak:    4,
		GuessAccuracy:      87.5,
		TotalVotesCast:     40,
		TotalVotesReceived: 22,
	}

	h := buildProfileHandler(mq)
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "userId")
	c.SetParamValues("g1", viewedID)
	setUser(c, viewerID, viewerID)

	if err := h.GetProfile(c); err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	return body["data"].(map[string]any)["stats"].(map[string]any)
}

func TestGetProfile_AccuracyKeyAbsentForAnotherMember(t *testing.T) {
	stats := profileStats(t, "u2", "u1")

	// Absent, not zero. A `"guess_accuracy": 0` is something a client renders as "never matched
	// anything", which is a different and false claim.
	if _, present := stats["guess_accuracy"]; present {
		t.Errorf("another member's profile carries guess_accuracy: %v", stats["guess_accuracy"])
	}
}

func TestGetProfile_AccuracyKeyPresentForOneself(t *testing.T) {
	stats := profileStats(t, "u1", "u1")

	value, present := stats["guess_accuracy"]
	if !present {
		t.Fatal("one's own profile should carry guess_accuracy")
	}
	if value.(float64) != 87.5 {
		t.Errorf("guess_accuracy = %v, want 87.5", value)
	}
}

func TestGetProfile_ParticipationStatsStayVisible(t *testing.T) {
	// None of these counts *matches*, so none can be combined with the published per-question winners
	// to recover a vote. Stripping a profile of everything has its own cost; the rule is specific.
	stats := profileStats(t, "u2", "u1")

	for key, want := range map[string]float64{
		"seasons_played":       4,
		"voting_streak":        3,
		"max_voting_streak":    4,
		"total_votes_cast":     40,
		"total_votes_received": 22,
	} {
		got, present := stats[key]
		if !present {
			t.Errorf("%s is missing from another member's profile", key)
			continue
		}
		if got.(float64) != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
}
