package chronicle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/eligibility"
)

// --- Mock querier ---

type mockQuerier struct {
	db.Querier

	isMember int64
	blocked  []string
	rows     []db.GetGroupChronicleRow
	total    int64
	standing []db.GetGroupGuessStandingRow

	memberErr   error
	rowsErr     error
	totalErr    error
	standingErr error

	// standingCalls records whether the standing query was asked for at all, so a test can assert it
	// was skipped rather than fetched and discarded.
	standingCalls int
}

func (m *mockQuerier) IsGroupMember(_ context.Context, _ db.IsGroupMemberParams) (int64, error) {
	return m.isMember, m.memberErr
}

func (m *mockQuerier) ListBlockedUserIDs(_ context.Context, _ string) ([]string, error) {
	return m.blocked, nil
}

func (m *mockQuerier) GetGroupChronicle(_ context.Context, _ db.GetGroupChronicleParams) ([]db.GetGroupChronicleRow, error) {
	return m.rows, m.rowsErr
}

func (m *mockQuerier) CountRevealedSeasons(_ context.Context, _ string) (int64, error) {
	return m.total, m.totalErr
}

func (m *mockQuerier) GetGroupGuessStanding(_ context.Context, _ string) ([]db.GetGroupGuessStandingRow, error) {
	m.standingCalls++
	return m.standing, m.standingErr
}

// --- Fixtures ---

func chronicleRow(seasonID string, number int32, questionID, targetID, username string, pct float64, voters int32) db.GetGroupChronicleRow {
	return db.GetGroupChronicleRow{
		SeasonID:         seasonID,
		SeasonNumber:     number,
		RevealAt:         time.Date(2026, 9, int(number)+1, 17, 0, 0, 0, time.UTC),
		QuestionID:       questionID,
		QuestionText:     "Вопрос " + questionID,
		QuestionCategory: db.QuestionCategoryFUNNY,
		TargetID:         targetID,
		Username:         username,
		Percentage:       pct,
		VoteCount:        int32(pct) / 10,
		TotalVoters:      voters,
	}
}

func standingRow(id, username string, accuracy float64, seasons int32, ranked bool) db.GetGroupGuessStandingRow {
	r := db.GetGroupGuessStandingRow{ID: id, Username: username}
	if ranked {
		r.GuessAccuracy = sql.NullFloat64{Float64: accuracy, Valid: true}
		r.SeasonsPlayed = sql.NullInt32{Int32: seasons, Valid: true}
	}
	return r
}

func newService(q *mockQuerier) *Service { return NewService(q) }

// --- Membership ---

func TestGet_NonMemberIsRefused(t *testing.T) {
	svc := newService(&mockQuerier{isMember: 0})
	if _, err := svc.Get(context.Background(), "outsider", "g1"); !errors.Is(err, ErrNotMember) {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestGet_MembershipErrorSurfaces(t *testing.T) {
	boom := fmt.Errorf("db down")
	svc := newService(&mockQuerier{memberErr: boom})
	if _, err := svc.Get(context.Background(), "u1", "g1"); !errors.Is(err, boom) {
		t.Fatalf("expected the db error, got %v", err)
	}
}

// --- Grouping ---

func TestGet_GroupsRowsIntoSeasonsNewestFirst(t *testing.T) {
	// The query orders by season id, not number, because DISTINCT ON forces its own leading ORDER BY.
	// These rows arrive in the order Postgres would produce them for ids s1 < s2 < s3.
	q := &mockQuerier{
		isMember: 1,
		total:    3,
		rows: []db.GetGroupChronicleRow{
			chronicleRow("s1", 1, "q1", "u-a", "alice", 60, 10),
			chronicleRow("s1", 1, "q2", "u-b", "bob", 70, 10),
			chronicleRow("s2", 2, "q1", "u-b", "bob", 50, 10),
			chronicleRow("s3", 3, "q1", "u-a", "alice", 80, 10),
		},
		standing: []db.GetGroupGuessStandingRow{standingRow("u-a", "alice", 0.8, 3, true)},
	}

	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(chron.Seasons) != 3 {
		t.Fatalf("got %d seasons, want 3", len(chron.Seasons))
	}
	for i, want := range []int32{3, 2, 1} {
		if chron.Seasons[i].Number != want {
			t.Errorf("season[%d] number = %d, want %d — newest first", i, chron.Seasons[i].Number, want)
		}
	}
	if n := len(chron.Seasons[2].Entries); n != 2 {
		t.Errorf("season 1 has %d entries, want 2", n)
	}
	if chron.SeasonsShown != 3 || chron.SeasonsTotal != 3 {
		t.Errorf("shown=%d total=%d, want 3 and 3", chron.SeasonsShown, chron.SeasonsTotal)
	}
}

func TestGet_EmptyChronicleIsEmptyNotNil(t *testing.T) {
	q := &mockQuerier{isMember: 1, total: 0}
	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if chron.Seasons == nil {
		t.Error("seasons must marshal as [] rather than null")
	}
	if len(chron.Seasons) != 0 {
		t.Errorf("got %d seasons, want 0", len(chron.Seasons))
	}
}

// --- Blocks ---

func TestGet_BlockedWinnersEntryIsWithheld(t *testing.T) {
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		blocked:  []string{"u-b"},
		rows: []db.GetGroupChronicleRow{
			chronicleRow("s1", 1, "q1", "u-a", "alice", 60, 10),
			chronicleRow("s1", 1, "q2", "u-b", "bob", 70, 10),
		},
		standing: []db.GetGroupGuessStandingRow{
			standingRow("u-a", "alice", 0.8, 2, true),
			standingRow("u-b", "bob", 0.9, 2, true),
		},
	}

	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(chron.Seasons) != 1 {
		t.Fatalf("the season itself must stay: dropping it would signal that a block happened")
	}
	for _, e := range chron.Seasons[0].Entries {
		if e.UserID == "u-b" {
			t.Error("a blocked member was named in the chronicle")
		}
	}
	if len(chron.Seasons[0].Entries) != 1 {
		t.Errorf("got %d entries, want 1", len(chron.Seasons[0].Entries))
	}
	for _, m := range chron.Standing.Members {
		if m.UserID == "u-b" {
			t.Error("a blocked member was listed in the standing")
		}
	}
}

func TestGet_DepartedWinnerStaysInTheRecord(t *testing.T) {
	// The winner is absent from the standing — they left the group — but the entry records something
	// that happened, so it remains.
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		rows:     []db.GetGroupChronicleRow{chronicleRow("s1", 1, "q1", "u-gone", "ушёл", 90, 10)},
		standing: []db.GetGroupGuessStandingRow{standingRow("u-a", "alice", 0.5, 2, true)},
	}

	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(chron.Seasons[0].Entries) != 1 || chron.Seasons[0].Entries[0].Username != "ушёл" {
		t.Error("a departed member's result must stay in the record")
	}
}

// --- Small-sample marking ---

func TestGet_MarksEntriesComputedOverTooFewVoters(t *testing.T) {
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		rows: []db.GetGroupChronicleRow{
			chronicleRow("s1", 1, "q1", "u-a", "alice", 66, int32(eligibility.MinDetectorMembers)-1),
			chronicleRow("s1", 1, "q2", "u-a", "alice", 66, int32(eligibility.MinDetectorMembers)),
			chronicleRow("s1", 1, "q3", "u-a", "alice", 75, 12),
		},
		standing: []db.GetGroupGuessStandingRow{standingRow("u-a", "alice", 0.5, 2, true)},
	}

	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"q1": true, "q2": false, "q3": false}
	for _, e := range chron.Seasons[0].Entries {
		if e.SmallSample != want[e.QuestionID] {
			t.Errorf("%s: small_sample = %v, want %v (%d voters, threshold %d)",
				e.QuestionID, e.SmallSample, want[e.QuestionID], e.TotalVoters, eligibility.MinDetectorMembers)
		}
	}
}

func TestGet_MarkingUsesTheDetectorsThreshold(t *testing.T) {
	// Not a second number: the chronicle marks exactly what the detector is gated on. Expressed as a
	// boundary test against the constant so changing the constant moves this with it.
	for voters, wantMarked := range map[int32]bool{
		int32(eligibility.MinDetectorMembers) - 1: true,
		int32(eligibility.MinDetectorMembers):     false,
	} {
		q := &mockQuerier{
			isMember: 1,
			total:    2,
			rows:     []db.GetGroupChronicleRow{chronicleRow("s1", 1, "q1", "u-a", "alice", 50, voters)},
			standing: []db.GetGroupGuessStandingRow{standingRow("u-a", "alice", 0.5, 2, true)},
		}
		chron, err := newService(q).Get(context.Background(), "u-a", "g1")
		if err != nil {
			t.Fatal(err)
		}
		if got := chron.Seasons[0].Entries[0].SmallSample; got != wantMarked {
			t.Errorf("%d voters: small_sample = %v, want %v", voters, got, wantMarked)
		}
	}
}

// --- Standing ---

func TestGet_StandingWithheldBelowTwoSeasons(t *testing.T) {
	for _, total := range []int64{0, 1} {
		q := &mockQuerier{isMember: 1, total: total}
		chron, err := newService(q).Get(context.Background(), "u-a", "g1")
		if err != nil {
			t.Fatal(err)
		}
		if chron.Standing.Available {
			t.Errorf("%d revealed seasons: the standing must be withheld", total)
		}
		if chron.Standing.Reason == "" {
			t.Errorf("%d revealed seasons: the app needs a reason to show", total)
		}
		if chron.Standing.Members == nil {
			t.Errorf("%d revealed seasons: members must marshal as [] rather than null", total)
		}
		// Not fetched at all, rather than fetched and dropped.
		if q.standingCalls != 0 {
			t.Errorf("%d revealed seasons: queried the standing %d times, want 0", total, q.standingCalls)
		}
	}
}

func TestGet_StandingShownFromTwoSeasons(t *testing.T) {
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		standing: []db.GetGroupGuessStandingRow{
			standingRow("u-a", "alice", 0.9, 2, true),
			standingRow("u-b", "bob", 0.4, 2, true),
		},
	}
	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if !chron.Standing.Available {
		t.Fatal("two revealed seasons should be enough")
	}
	if len(chron.Standing.Members) != 2 {
		t.Fatalf("got %d members, want 2", len(chron.Standing.Members))
	}
	if chron.Standing.Members[0].UserID != "u-a" || chron.Standing.Members[0].Rank != 1 {
		t.Errorf("expected alice ranked 1, got %+v", chron.Standing.Members[0])
	}
	if chron.Standing.Members[1].Rank != 2 {
		t.Errorf("expected bob ranked 2, got %d", chron.Standing.Members[1].Rank)
	}
	for _, m := range chron.Standing.Members {
		if m.SeasonsPlayed != 2 {
			t.Errorf("%s: seasons_played = %d, want 2 — the standing must say what it is based on", m.Username, m.SeasonsPlayed)
		}
	}
}

func TestGet_MemberWithNoRevealYetIsUnranked(t *testing.T) {
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		standing: []db.GetGroupGuessStandingRow{
			standingRow("u-a", "alice", 0.9, 2, true),
			standingRow("u-new", "новенький", 0, 0, false),
		},
	}
	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	newcomer := chron.Standing.Members[1]
	if newcomer.Ranked {
		t.Error("a member with no revealed season must not be ranked")
	}
	if newcomer.Rank != 0 {
		t.Errorf("rank = %d, want 0 (unranked, not last)", newcomer.Rank)
	}
}

func TestGet_RanksAreComputedBeforeBlocksAreApplied(t *testing.T) {
	// Blocking the leader must not promote everyone else: two members comparing screens would then
	// see different places for the same person.
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		blocked:  []string{"u-a"},
		standing: []db.GetGroupGuessStandingRow{
			standingRow("u-a", "alice", 0.9, 2, true),
			standingRow("u-b", "bob", 0.5, 2, true),
			standingRow("u-c", "carol", 0.1, 2, true),
		},
	}
	chron, err := newService(q).Get(context.Background(), "u-b", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(chron.Standing.Members) != 2 {
		t.Fatalf("got %d members, want 2", len(chron.Standing.Members))
	}
	if chron.Standing.Members[0].Rank != 2 || chron.Standing.Members[1].Rank != 3 {
		t.Errorf("ranks = %d, %d; want 2, 3 — the blocked leader still occupies rank 1",
			chron.Standing.Members[0].Rank, chron.Standing.Members[1].Rank)
	}
}

func TestGet_StandingCarriesNoVote(t *testing.T) {
	// The standing is derived from accuracy figures alone. Asserted structurally: a question, target
	// or vote field appearing here would be a way to infer who voted how.
	q := &mockQuerier{
		isMember: 1,
		total:    2,
		standing: []db.GetGroupGuessStandingRow{standingRow("u-a", "alice", 1.0, 2, true)},
	}
	chron, err := newService(q).Get(context.Background(), "u-a", "g1")
	if err != nil {
		t.Fatal(err)
	}
	entry := chron.Standing.Members[0]
	if !entry.Ranked || entry.Rank != 1 {
		t.Fatalf("expected the member ranked first, got %+v", entry)
	}
	// The accuracy figure must not be here. With the chronicle publishing who won each question, a
	// member at 100% has their vote on every one of those questions recovered exactly — and because
	// the figure is a rolling average weighted by seasons_played, two consecutive weeks can be diffed
	// into that season's match count. A rank cannot.
	blob := fmt.Sprintf("%#v", entry)
	for _, forbidden := range []string{"Accuracy", "GuessAccuracy", "QuestionID", "QuestionText", "TargetID", "VoteCount"} {
		if strings.Contains(blob, forbidden) {
			t.Errorf("the standing entry exposes %s", forbidden)
		}
	}
	// And the same in the wire form, which is what actually reaches a client.
	wire, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(wire), "accuracy") {
		t.Errorf("the serialised standing entry carries an accuracy: %s", wire)
	}
}

// --- Caps ---

func TestGet_RequestsAtMostMaxSeasons(t *testing.T) {
	// The cap is what keeps a year-old group's response bounded; a limit of zero would silently empty
	// every chronicle.
	if MaxSeasons < 1 {
		t.Fatal("MaxSeasons must be at least 1")
	}
	if MinSeasonsForStanding < 2 {
		t.Error("the standing must need more than one season: a rolling average over one season is luck")
	}
}
