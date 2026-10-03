package reveal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/schedule"
)

// antMock adds the anticipation queries to the shared detector mock.
type antMock struct {
	detectorMock
	votersAbout map[string]int64               // "seasonID:targetID" -> distinct voters
	leading     map[string]db.QuestionCategory // "seasonID:targetID" -> leading category
	leadingErr  error
}

func (m *antMock) CountVotersAboutTarget(_ context.Context, arg db.CountVotersAboutTargetParams) (int64, error) {
	return m.votersAbout[arg.SeasonID+":"+arg.TargetID], nil
}

func (m *antMock) GetLeadingCategoryForTarget(_ context.Context, arg db.GetLeadingCategoryForTargetParams) (db.GetLeadingCategoryForTargetRow, error) {
	if m.leadingErr != nil {
		return db.GetLeadingCategoryForTargetRow{}, m.leadingErr
	}
	c, ok := m.leading[arg.SeasonID+":"+arg.TargetID]
	if !ok {
		return db.GetLeadingCategoryForTargetRow{}, errors.New("no rows")
	}
	return db.GetLeadingCategoryForTargetRow{Category: c, Votes: 3}, nil
}

func newAntMock() *antMock {
	return &antMock{
		detectorMock: *newDetectorMock(),
		votersAbout:  map[string]int64{},
		leading:      map[string]db.QuestionCategory{},
	}
}

func TestGetAnticipation_CountsVotersNotVotes(t *testing.T) {
	m := newAntMock()
	// One member answered three questions about u1 — that is one voter.
	m.votersAbout["s1:u1"] = 1
	svc := NewService(m, nil)

	state, err := svc.GetAnticipation(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if state.VotersAboutMe != 1 {
		t.Errorf("VotersAboutMe = %d, want 1", state.VotersAboutMe)
	}
	if state.RevealAt == "" {
		t.Error("the state should carry the reveal time for the countdown")
	}
}

func TestGetAnticipation_ZeroWhenNobodyVotedAboutMe(t *testing.T) {
	m := newAntMock()
	svc := NewService(m, nil)

	state, err := svc.GetAnticipation(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if state.VotersAboutMe != 0 {
		t.Errorf("VotersAboutMe = %d, want 0", state.VotersAboutMe)
	}
	if state.TeaserEmoji != "" {
		t.Error("no votes means no teaser, whatever day it is")
	}
}

func TestGetAnticipation_RefusedForARevealedSeason(t *testing.T) {
	m := newAntMock()
	svc := NewService(m, nil)

	// s2 is REVEALED in the shared fixture.
	_, err := svc.GetAnticipation(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrSeasonAlreadyRevealed) {
		t.Errorf("expected ErrSeasonAlreadyRevealed, got %v", err)
	}
}

func TestGetAnticipation_RefusedForANonMember(t *testing.T) {
	m := newAntMock()
	svc := NewService(m, nil)

	_, err := svc.GetAnticipation(context.Background(), "s1", "stranger")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetAnticipation_AvailableBeforeIHaveVoted(t *testing.T) {
	m := newAntMock()
	m.votersAbout["s1:u1"] = 2
	svc := NewService(m, nil)

	// The fixture's u1 has cast no votes; the count must still be served.
	state, err := svc.GetAnticipation(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatalf("the count must not require having voted: %v", err)
	}
	if state.VotersAboutMe != 2 {
		t.Errorf("VotersAboutMe = %d, want 2", state.VotersAboutMe)
	}
}

func TestGetAnticipation_PayloadCannotCarryIdentitiesOrAttributes(t *testing.T) {
	m := newAntMock()
	m.votersAbout["s1:u1"] = 2
	m.leading["s1:u1"] = db.QuestionCategoryHOT
	svc := NewService(m, nil)

	state, err := svc.GetAnticipation(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the exact key set: this is the shape being guaranteed, and it fails loudly if anyone
	// adds a field rather than relying on a substring blocklist that could miss one.
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"voters_about_me": true, "teaser_emoji": true, "reveal_at": true}
	for key := range decoded {
		if !want[key] {
			t.Errorf("unexpected key %q in the anticipation payload: the shape must stay incapable of carrying identities or attributes", key)
		}
	}
	if len(decoded) != len(want) {
		t.Errorf("payload has %d keys, want %d: %s", len(decoded), len(want), raw)
	}

	// And the teaser is the emoji itself, not a category code the client could map to a name.
	body := strings.ToLower(string(raw))
	for _, forbidden := range []string{"hot", "funny", "secrets", "skills", "romance", "study"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("the teaser must be an emoji, not the category name %q: %s", forbidden, body)
		}
	}
}

func TestTeaserVisible_OnlyFromThursday(t *testing.T) {
	msk := schedule.MSK()
	// Week of Mon 2026-10-05.
	days := map[string]struct {
		day  int
		want bool
	}{
		"monday":    {5, false},
		"tuesday":   {6, false},
		"wednesday": {7, false},
		"thursday":  {8, true},
		"friday":    {9, true},
		"saturday":  {10, true},
		"sunday":    {11, true},
	}
	for name, c := range days {
		at := time.Date(2026, 10, c.day, 12, 0, 0, 0, msk)
		if got := teaserVisible(at); got != c.want {
			t.Errorf("%s: teaserVisible = %v, want %v", name, got, c.want)
		}
	}
}

func TestCategoryEmoji_EveryCategoryHasOne(t *testing.T) {
	for _, c := range []db.QuestionCategory{
		db.QuestionCategoryHOT,
		db.QuestionCategoryFUNNY,
		db.QuestionCategorySECRETS,
		db.QuestionCategorySKILLS,
		db.QuestionCategoryROMANCE,
		db.QuestionCategorySTUDY,
	} {
		if CategoryEmoji(c) == "" {
			t.Errorf("category %s has no teaser emoji", c)
		}
	}
	if CategoryEmoji(db.QuestionCategory("NOPE")) != "" {
		t.Error("an unknown category should yield no emoji rather than a placeholder")
	}
}

func TestGetAnticipation_MissingLeadingCategoryIsNotAnError(t *testing.T) {
	m := newAntMock()
	m.votersAbout["s1:u1"] = 2
	m.leadingErr = errors.New("db down")
	svc := NewService(m, nil)

	state, err := svc.GetAnticipation(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatalf("a missing teaser must not fail the screen: %v", err)
	}
	if state.TeaserEmoji != "" {
		t.Error("expected no teaser")
	}
	if state.VotersAboutMe != 2 {
		t.Error("the count should still be served")
	}
}
