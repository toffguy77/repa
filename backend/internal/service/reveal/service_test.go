package reveal

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
	"github.com/repa-app/repa/internal/schedule"
	"github.com/sqlc-dev/pqtype"
)

// mockQuerier implements only the methods used by the reveal service.
type mockQuerier struct {
	db.Querier
	seasons             map[string]db.Season
	members             map[string]map[string]bool
	memberRows          map[string][]db.GetGroupMembersRow
	memberCounts        map[string]int64
	uniqueVoters        map[string]int64
	aggregated          map[string][]db.AggregateVotesByTargetRow
	resultsByUser       map[string][]db.GetSeasonResultsByUserRow
	topPerQuestion      map[string][]db.GetTopResultPerQuestionRow
	prevSeason          map[string]db.Season // groupID -> prev season
	allResultsWithUsers map[string][]db.GetAllSeasonResultsWithUsersRow
	balance             map[string]int32 // userID -> balance
	createdResults      []db.CreateSeasonResultParams
	createdCrystals     []db.CreateCrystalLogParams
	deletedSeasons      []string
	updatedStatuses     []db.UpdateSeasonStatusParams
	seasonsForReveal    []db.Season
	seasonQuestions     map[string]int64 // seasonID -> question count (default defaultSeasonQuestions)
	postponed           []db.PostponeSeasonParams
	kickoffScheduled    []db.ScheduleKickoffRevealParams
	blockedIDs          map[string][]string // userID -> people blocked in either direction
	shareEvents         []db.CreateShareEventParams
	shareEventErr       error
}

// defaultSeasonQuestions keeps fixtures terse: tests that care about completion counts
// set uniqueVoters, which the mock reports as the number of *completed* voters.
const defaultSeasonQuestions = 5

func (m *mockQuerier) CountSeasonQuestions(_ context.Context, seasonID string) (int64, error) {
	if n, ok := m.seasonQuestions[seasonID]; ok {
		return n, nil
	}
	return defaultSeasonQuestions, nil
}

func (m *mockQuerier) CountCompletedVoters(_ context.Context, arg db.CountCompletedVotersParams) (int64, error) {
	return m.uniqueVoters[arg.SeasonID], nil
}

func (m *mockQuerier) PostponeSeason(_ context.Context, arg db.PostponeSeasonParams) (db.Season, error) {
	m.postponed = append(m.postponed, arg)
	s, ok := m.seasons[arg.ID]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	s.RevealAt = arg.RevealAt
	s.EndsAt = arg.EndsAt
	s.PostponeCount++
	m.seasons[arg.ID] = s
	return s, nil
}

// blockedIDs lets a test say "this viewer has blocked these people".
func (m *mockQuerier) ListBlockedUserIDs(_ context.Context, userID string) ([]string, error) {
	return m.blockedIDs[userID], nil
}

// CountBlockedGroupMembers is the group-scoped form the detector and the anonymity count use. The
// fixture's block list is keyed by user rather than by group, so this counts all of it — which is what
// these tests already assumed when the production code used the global list.
func (m *mockQuerier) CountBlockedGroupMembers(_ context.Context, arg db.CountBlockedGroupMembersParams) (int64, error) {
	return int64(len(m.blockedIDs[arg.BlockerID])), nil
}

func (m *mockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	s, ok := m.seasons[id]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if g, ok := m.members[arg.GroupID]; ok && g[arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

func (m *mockQuerier) CountGroupMembers(_ context.Context, groupID string) (int64, error) {
	return m.memberCounts[groupID], nil
}

func (m *mockQuerier) CountUniqueVoters(_ context.Context, seasonID string) (int64, error) {
	return m.uniqueVoters[seasonID], nil
}

func (m *mockQuerier) AggregateVotesByTarget(_ context.Context, seasonID string) ([]db.AggregateVotesByTargetRow, error) {
	return m.aggregated[seasonID], nil
}

func (m *mockQuerier) DeleteSeasonResultsBySeason(_ context.Context, seasonID string) error {
	m.deletedSeasons = append(m.deletedSeasons, seasonID)
	return nil
}

func (m *mockQuerier) CreateSeasonResult(_ context.Context, arg db.CreateSeasonResultParams) (db.SeasonResult, error) {
	m.createdResults = append(m.createdResults, arg)
	return db.SeasonResult{}, nil
}

func (m *mockQuerier) UpdateSeasonStatus(_ context.Context, arg db.UpdateSeasonStatusParams) error {
	m.updatedStatuses = append(m.updatedStatuses, arg)
	if s, ok := m.seasons[arg.ID]; ok {
		s.Status = arg.Status
		m.seasons[arg.ID] = s
	}
	return nil
}

func (m *mockQuerier) GetSeasonsForReveal(_ context.Context) ([]db.Season, error) {
	return m.seasonsForReveal, nil
}

func (m *mockQuerier) GetSeasonResultsByUser(_ context.Context, arg db.GetSeasonResultsByUserParams) ([]db.GetSeasonResultsByUserRow, error) {
	key := arg.SeasonID + ":" + arg.TargetID
	return m.resultsByUser[key], nil
}

func (m *mockQuerier) GetTopResultPerQuestion(_ context.Context, seasonID string) ([]db.GetTopResultPerQuestionRow, error) {
	return m.topPerQuestion[seasonID], nil
}

func (m *mockQuerier) GetGroupMembers(_ context.Context, groupID string) ([]db.GetGroupMembersRow, error) {
	return m.memberRows[groupID], nil
}

func (m *mockQuerier) GetPreviousRevealedSeason(_ context.Context, arg db.GetPreviousRevealedSeasonParams) (db.Season, error) {
	s, ok := m.prevSeason[arg.GroupID]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) GetUserBalance(_ context.Context, userID string) (int32, error) {
	b, ok := m.balance[userID]
	if !ok {
		return 0, nil
	}
	return b, nil
}

func (m *mockQuerier) CreateCrystalLog(_ context.Context, arg db.CreateCrystalLogParams) (db.CrystalLog, error) {
	m.createdCrystals = append(m.createdCrystals, arg)
	return db.CrystalLog{}, nil
}

func (m *mockQuerier) GetAllSeasonResultsWithUsers(_ context.Context, seasonID string) ([]db.GetAllSeasonResultsWithUsersRow, error) {
	return m.allResultsWithUsers[seasonID], nil
}

func (m *mockQuerier) GetSeasonAchievements(_ context.Context, _ sql.NullString) ([]db.Achievement, error) {
	return []db.Achievement{}, nil
}

func (m *mockQuerier) GetCardCache(_ context.Context, arg db.GetCardCacheParams) (db.CardCache, error) {
	return db.CardCache{}, sql.ErrNoRows
}

// --- Fixtures ---

func newMock() *mockQuerier {
	return &mockQuerier{
		blockedIDs: map[string][]string{},
		seasons: map[string]db.Season{
			"s1": {ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusVOTING},
			"s2": {ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		},
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true, "u3": true, "u4": true, "u5": true},
		},
		memberRows: map[string][]db.GetGroupMembersRow{
			"g1": {
				{ID: "u1", Username: "alice"},
				{ID: "u2", Username: "bob"},
				{ID: "u3", Username: "charlie"},
			},
		},
		memberCounts: map[string]int64{"g1": 5},
		uniqueVoters: map[string]int64{"s1": 3, "s2": 4},
		aggregated: map[string][]db.AggregateVotesByTargetRow{
			"s1": {
				{TargetID: "u1", QuestionID: "q1", VoteCount: 2},
				{TargetID: "u2", QuestionID: "q1", VoteCount: 1},
				{TargetID: "u1", QuestionID: "q2", VoteCount: 3},
				{TargetID: "u3", QuestionID: "q2", VoteCount: 1},
			},
		},
		resultsByUser: map[string][]db.GetSeasonResultsByUserRow{
			"s2:u1": {
				{QuestionID: "q1", QuestionText: "Who is funniest?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 80, VoteCount: 4, TotalVoters: 5},
				{QuestionID: "q2", QuestionText: "Who is hottest?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 60, VoteCount: 3, TotalVoters: 5},
				{QuestionID: "q3", QuestionText: "Who knows secrets?", QuestionCategory: db.QuestionCategorySECRETS, Percentage: 40, VoteCount: 2, TotalVoters: 5},
				{QuestionID: "q4", QuestionText: "Who studies most?", QuestionCategory: db.QuestionCategorySTUDY, Percentage: 20, VoteCount: 1, TotalVoters: 5},
			},
		},
		topPerQuestion: map[string][]db.GetTopResultPerQuestionRow{
			"s2": {
				{QuestionID: "q1", TargetID: "u1", QuestionText: "Who is funniest?", Username: "alice", Percentage: 80},
				{QuestionID: "q2", TargetID: "u1", QuestionText: "Who is hottest?", Username: "alice", Percentage: 60},
			},
		},
		allResultsWithUsers: map[string][]db.GetAllSeasonResultsWithUsersRow{
			"s2": {
				{TargetID: "u1", QuestionID: "q1", QuestionText: "Who is funniest?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 80, Username: "alice"},
				{TargetID: "u1", QuestionID: "q2", QuestionText: "Who is hottest?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 60, Username: "alice"},
				{TargetID: "u1", QuestionID: "q3", QuestionText: "Who knows secrets?", QuestionCategory: db.QuestionCategorySECRETS, Percentage: 40, Username: "alice"},
				{TargetID: "u1", QuestionID: "q4", QuestionText: "Who studies most?", QuestionCategory: db.QuestionCategorySTUDY, Percentage: 20, Username: "alice"},
				{TargetID: "u2", QuestionID: "q1", QuestionText: "Who is funniest?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 50, Username: "bob"},
			},
		},
		balance:         map[string]int32{"u1": 20},
		createdResults:  nil,
		createdCrystals: nil,
		deletedSeasons:  nil,
		updatedStatuses: nil,
	}
}

// --- ProcessReveal tests ---

func TestProcessReveal_QuorumMet(t *testing.T) {
	m := newMock()
	// 3 voters out of 5 members, threshold 40% for <8 → 3 >= 2 = quorum met
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Revealed {
		t.Error("expected season to be revealed")
	}
	if result.Retry {
		t.Error("expected no retry")
	}
	if len(m.createdResults) != 4 {
		t.Errorf("expected 4 season results, got %d", len(m.createdResults))
	}
	if len(m.updatedStatuses) != 1 || m.updatedStatuses[0].Status != db.SeasonStatusREVEALED {
		t.Error("expected status updated to REVEALED")
	}
}

func TestProcessReveal_QuorumNotMet_Retry(t *testing.T) {
	m := newMock()
	m.uniqueVoters["s1"] = 1 // 1 out of 5, neither 40% nor 50% threshold
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Revealed {
		t.Error("expected not revealed")
	}
	if !result.Retry {
		t.Error("expected retry")
	}
	if len(m.createdResults) != 0 {
		t.Errorf("expected no results created, got %d", len(m.createdResults))
	}
}

func TestProcessReveal_QuorumNotMet_ForcedAfterMaxAttempts(t *testing.T) {
	m := newMock()
	// Floor met (>=3 members, >=3 completed voters) but 3/10 is under the 50% quorum:
	// the forced path bypasses the percentage, so this still reveals.
	m.memberCounts["g1"] = 10
	m.uniqueVoters["s1"] = 3
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 3)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Revealed {
		t.Error("expected forced reveal after max attempts")
	}
	if result.Retry {
		t.Error("expected no retry")
	}
	if result.Postponed {
		t.Error("expected no postponement when the participation floor is met")
	}
}

func TestProcessReveal_AlreadyRevealed(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s2", 1)
	if err != nil {
		t.Fatal(err)
	}

	if result.Revealed {
		t.Error("expected not revealed (already REVEALED)")
	}
}

func TestProcessReveal_AggregationPercentages(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}

	// 3 unique voters: u1/q1 got 2 votes → 66.7%, u2/q1 got 1 → 33.3%, u1/q2 got 3 → 100%, u3/q2 got 1 → 33.3%
	for _, r := range m.createdResults {
		if r.TargetID == "u1" && r.QuestionID == "q1" {
			if r.Percentage != 66.7 {
				t.Errorf("expected 66.7%% for u1/q1, got %.1f%%", r.Percentage)
			}
		}
		if r.TargetID == "u1" && r.QuestionID == "q2" {
			if r.Percentage != 100.0 {
				t.Errorf("expected 100%% for u1/q2, got %.1f%%", r.Percentage)
			}
		}
	}
}

// --- GetReveal tests ---

func TestGetReveal_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	data, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(data.MyCard.TopAttributes) != 3 {
		t.Errorf("expected 3 top attributes, got %d", len(data.MyCard.TopAttributes))
	}
	if len(data.MyCard.HiddenAttributes) != 1 {
		t.Errorf("expected 1 hidden attribute, got %d", len(data.MyCard.HiddenAttributes))
	}
	if data.MyCard.ReputationTitle == "" {
		t.Error("expected non-empty reputation title")
	}
	if data.MyCard.TopAttributes[0].Rank != 1 {
		t.Errorf("expected rank 1 for top attribute, got %d", data.MyCard.TopAttributes[0].Rank)
	}
	if data.GroupSummary.VoterCount != 4 {
		t.Errorf("expected voter count 4, got %d", data.GroupSummary.VoterCount)
	}
}

func TestGetReveal_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestGetReveal_NotRevealed(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

func TestGetReveal_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetReveal_TitleFromCategory(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	data, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	// Top category is FUNNY → "Душа компании"
	if data.MyCard.ReputationTitle != "Душа компании" {
		t.Errorf("expected title 'Душа компании', got '%s'", data.MyCard.ReputationTitle)
	}
}

// --- GetMembersCards tests ---

func TestGetMembersCards_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	// 2 users have results in the allResultsWithUsers fixture (u1, u2)
	if len(cards) != 2 {
		t.Errorf("expected 2 member cards, got %d", len(cards))
	}

	// u1 should have 3 top attributes (out of 4 total)
	for _, c := range cards {
		if c.UserID == "u1" && len(c.TopAttributes) != 3 {
			t.Errorf("expected 3 top attributes for u1, got %d", len(c.TopAttributes))
		}
	}
}

func TestGetMembersCards_NotRevealed(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetMembersCards(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

// --- OpenHidden tests ---

func TestOpenHidden_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	result, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.AllAttributes) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(result.AllAttributes))
	}
	if result.CrystalBalance != 15 { // 20 - 5
		t.Errorf("expected balance 15, got %d", result.CrystalBalance)
	}
	if len(m.createdCrystals) != 1 {
		t.Errorf("expected 1 crystal log, got %d", len(m.createdCrystals))
	}
	if m.createdCrystals[0].Delta != -5 {
		t.Errorf("expected delta -5, got %d", m.createdCrystals[0].Delta)
	}
}

func TestOpenHidden_InsufficientFunds(t *testing.T) {
	m := newMock()
	m.balance["u1"] = 3
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestOpenHidden_ZeroBalance(t *testing.T) {
	m := newMock()
	m.balance["u1"] = 0
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestOpenHidden_NotRevealed(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

// --- Helper tests ---

func TestSplitAttributes_LessThan3(t *testing.T) {
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 80},
		{QuestionID: "q2", QuestionText: "Q2", Percentage: 60},
	}
	top, hidden := splitAttributes(results)
	if len(top) != 2 {
		t.Errorf("expected 2 top, got %d", len(top))
	}
	if len(hidden) != 0 {
		t.Errorf("expected 0 hidden, got %d", len(hidden))
	}
}

func TestSplitAttributes_Empty(t *testing.T) {
	top, hidden := splitAttributes(nil)
	if len(top) != 0 {
		t.Errorf("expected 0 top, got %d", len(top))
	}
	if len(hidden) != 0 {
		t.Errorf("expected 0 hidden, got %d", len(hidden))
	}
}

func TestGenerateTitle_AllCategories(t *testing.T) {
	tests := []struct {
		category string
		title    string
	}{
		{"HOT", "Горячая штучка"},
		{"FUNNY", "Душа компании"},
		{"SECRETS", "Хранитель тайн"},
		{"SKILLS", "Мастер на все руки"},
		{"ROMANCE", "Сердцеед"},
		{"STUDY", "Ботан года"},
	}

	for _, tt := range tests {
		attrs := []AttributeDto{{Category: tt.category}}
		got := generateTitle(attrs)
		if got != tt.title {
			t.Errorf("category %s: expected '%s', got '%s'", tt.category, tt.title, got)
		}
	}
}

func TestGenerateTitle_Empty(t *testing.T) {
	got := generateTitle(nil)
	if got != "Загадка века" {
		t.Errorf("expected 'Загадка века', got '%s'", got)
	}
}

func TestGenerateTitle_UnknownCategory(t *testing.T) {
	got := generateTitle([]AttributeDto{{Category: "UNKNOWN"}})
	if got != "Загадка века" {
		t.Errorf("expected 'Загадка века', got '%s'", got)
	}
}

// --- ValidateRevealAccess tests ---

func TestValidateRevealAccess_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	season, err := svc.ValidateRevealAccess(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if season.ID != "s2" {
		t.Errorf("expected season s2, got %s", season.ID)
	}
}

func TestValidateRevealAccess_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ValidateRevealAccess(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestValidateRevealAccess_NotRevealed(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ValidateRevealAccess(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

func TestValidateRevealAccess_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ValidateRevealAccess(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- GetSeasonsForReveal ---

func TestGetSeasonsForReveal_Empty(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	seasons, err := svc.GetSeasonsForReveal(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(seasons) != 0 {
		t.Errorf("expected 0 seasons, got %d", len(seasons))
	}
}

func TestGetSeasonsForReveal_ReturnsList(t *testing.T) {
	m := newMock()
	m.seasonsForReveal = []db.Season{
		{ID: "s1", GroupID: "g1", Status: db.SeasonStatusVOTING},
	}
	svc := NewService(m, nil)

	seasons, err := svc.GetSeasonsForReveal(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(seasons) != 1 {
		t.Errorf("expected 1 season, got %d", len(seasons))
	}
}

// --- GetDetector tests ---

// detectorMock extends mockQuerier to add HasDetector and GetVoterProfilesBySeason
type detectorMock struct {
	mockQuerier
	hasDetector   map[string]bool // key = userID:seasonID
	voterProfiles map[string][]db.GetVoterProfilesBySeasonRow
	detectors     []db.CreateDetectorParams

	// Ladder state.
	seasonVoters map[string]int64                       // seasonID -> voter count
	hints        map[string][]db.GetDetectorHintsRow    // "userID:seasonID" -> revealed
	unrevealed   map[string][]db.PickUnrevealedVoterRow // "userID:seasonID" -> pickable
	createdHints []db.CreateDetectorHintParams
}

func (m *detectorMock) CountSeasonVoters(_ context.Context, seasonID string) (int64, error) {
	return m.seasonVoters[seasonID], nil
}

func (m *detectorMock) GetDetectorHints(_ context.Context, arg db.GetDetectorHintsParams) ([]db.GetDetectorHintsRow, error) {
	return m.hints[arg.UserID+":"+arg.SeasonID], nil
}

func (m *detectorMock) PickUnrevealedVoter(_ context.Context, arg db.PickUnrevealedVoterParams) (db.PickUnrevealedVoterRow, error) {
	key := arg.VoterID + ":" + arg.SeasonID
	pool := m.unrevealed[key]
	if len(pool) == 0 {
		return db.PickUnrevealedVoterRow{}, sql.ErrNoRows
	}
	return pool[0], nil
}

func (m *detectorMock) CreateDetectorHint(_ context.Context, arg db.CreateDetectorHintParams) (db.DetectorHint, error) {
	m.createdHints = append(m.createdHints, arg)

	key := arg.UserID + ":" + arg.SeasonID
	// Move the revealed voter out of the pool and into the revealed list, mirroring the
	// without-replacement rule the SQL enforces.
	var kept []db.PickUnrevealedVoterRow
	for _, v := range m.unrevealed[key] {
		if v.ID == arg.RevealedUserID {
			if m.hints == nil {
				m.hints = map[string][]db.GetDetectorHintsRow{}
			}
			m.hints[key] = append(m.hints[key], db.GetDetectorHintsRow{
				RevealedUserID: v.ID,
				Username:       v.Username,
				AvatarEmoji:    v.AvatarEmoji,
				AvatarUrl:      v.AvatarUrl,
			})
			continue
		}
		kept = append(kept, v)
	}
	m.unrevealed[key] = kept

	return db.DetectorHint{ID: arg.ID}, nil
}

func (m *detectorMock) HasDetector(_ context.Context, arg db.HasDetectorParams) (bool, error) {
	return m.hasDetector[arg.UserID+":"+arg.SeasonID], nil
}

func (m *detectorMock) GetVoterProfilesBySeason(_ context.Context, seasonID string) ([]db.GetVoterProfilesBySeasonRow, error) {
	return m.voterProfiles[seasonID], nil
}

func (m *detectorMock) CreateDetector(_ context.Context, arg db.CreateDetectorParams) (db.Detector, error) {
	m.detectors = append(m.detectors, arg)
	return db.Detector{}, nil
}

func (m *detectorMock) LockUserForUpdate(_ context.Context, id string) (string, error) {
	return id, nil
}

func newDetectorMock() *detectorMock {
	base := newMock()
	return &detectorMock{
		mockQuerier:   *base,
		hasDetector:   map[string]bool{},
		voterProfiles: map[string][]db.GetVoterProfilesBySeasonRow{},
		seasonVoters:  map[string]int64{},
		hints:         map[string][]db.GetDetectorHintsRow{},
		unrevealed:    map[string][]db.PickUnrevealedVoterRow{},
	}
}

func TestGetDetector_NotPurchased(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Purchased {
		t.Error("expected purchased=false")
	}
	if len(result.Voters) != 0 {
		t.Errorf("expected 0 voters, got %d", len(result.Voters))
	}
	if result.CrystalBalance != 20 {
		t.Errorf("expected balance 20, got %d", result.CrystalBalance)
	}
}

func TestGetDetector_Purchased(t *testing.T) {
	m := newDetectorMock()
	m.hasDetector["u1:s2"] = true
	m.voterProfiles["s2"] = []db.GetVoterProfilesBySeasonRow{
		{ID: "u2", Username: "bob", AvatarEmoji: sql.NullString{String: "X", Valid: true}},
		{ID: "u3", Username: "charlie"},
	}
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Purchased {
		t.Error("expected purchased=true")
	}
	if len(result.Voters) != 2 {
		t.Errorf("expected 2 voters, got %d", len(result.Voters))
	}
	if result.Voters[0].AvatarEmoji == nil || *result.Voters[0].AvatarEmoji != "X" {
		t.Error("expected avatar emoji for first voter")
	}
	if result.Voters[1].AvatarEmoji != nil {
		t.Error("expected nil avatar emoji for second voter")
	}
}

func TestGetDetector_NotMember(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	_, err := svc.GetDetector(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- BuyDetector tests ---

func TestBuyDetector_Success(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	result, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Purchased {
		t.Error("expected purchased=true")
	}
	if result.CrystalBalance != 10 { // 20 - 10
		t.Errorf("expected balance 10, got %d", result.CrystalBalance)
	}
	if len(m.createdCrystals) != 1 {
		t.Errorf("expected 1 crystal log, got %d", len(m.createdCrystals))
	}
	if m.createdCrystals[0].Delta != -10 {
		t.Errorf("expected delta -10, got %d", m.createdCrystals[0].Delta)
	}
}

func TestBuyDetector_InsufficientFunds(t *testing.T) {
	m := newDetectorMock()
	m.balance["u1"] = 5 // cost is 10
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestBuyDetector_AlreadyPurchased(t *testing.T) {
	m := newDetectorMock()
	m.hasDetector["u1:s2"] = true
	svc := NewService(m, nil)

	result, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	// Should return existing detector without deducting crystals
	if !result.Purchased {
		t.Error("expected purchased=true")
	}
	if len(m.createdCrystals) != 0 {
		t.Errorf("expected no crystal log for already purchased, got %d", len(m.createdCrystals))
	}
}

func TestBuyDetector_NotMember(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestBuyDetector_SeasonNotFound(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

// --- computeTrend tests ---

func TestComputeTrend_NoPreviousSeason(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend != nil {
		t.Error("expected nil trend when no previous season")
	}
}

func TestComputeTrend_WithPreviousSeason_Up(t *testing.T) {
	m := newMock()
	m.prevSeason = map[string]db.Season{
		"g1": {ID: "s-prev", GroupID: "g1", Status: db.SeasonStatusREVEALED},
	}
	m.resultsByUser["s-prev:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 60},
	}
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend == nil {
		t.Fatal("expected non-nil trend")
	}
	if trend.Change != "up" {
		t.Errorf("expected change 'up', got '%s'", trend.Change)
	}
	if trend.Delta != 20.0 {
		t.Errorf("expected delta 20.0, got %f", trend.Delta)
	}
}

func TestComputeTrend_WithPreviousSeason_Down(t *testing.T) {
	m := newMock()
	m.prevSeason = map[string]db.Season{
		"g1": {ID: "s-prev", GroupID: "g1", Status: db.SeasonStatusREVEALED},
	}
	m.resultsByUser["s-prev:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 90},
	}
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend == nil {
		t.Fatal("expected non-nil trend")
	}
	if trend.Change != "down" {
		t.Errorf("expected change 'down', got '%s'", trend.Change)
	}
}

func TestComputeTrend_WithPreviousSeason_Same(t *testing.T) {
	m := newMock()
	m.prevSeason = map[string]db.Season{
		"g1": {ID: "s-prev", GroupID: "g1", Status: db.SeasonStatusREVEALED},
	}
	m.resultsByUser["s-prev:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80.5},
	}
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend == nil {
		t.Fatal("expected non-nil trend")
	}
	if trend.Change != "same" {
		t.Errorf("expected change 'same', got '%s'", trend.Change)
	}
}

func TestComputeTrend_EmptyResults(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	trend := svc.computeTrend(context.Background(), season, "u1", nil)
	if trend != nil {
		t.Error("expected nil trend for empty results")
	}
}

func TestComputeTrend_QuestionNotInPrevious(t *testing.T) {
	m := newMock()
	m.prevSeason = map[string]db.Season{
		"g1": {ID: "s-prev", GroupID: "g1", Status: db.SeasonStatusREVEALED},
	}
	m.resultsByUser["s-prev:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q-other", QuestionText: "Different?", Percentage: 50},
	}
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend != nil {
		t.Error("expected nil trend when question not found in previous season")
	}
}

// --- ProcessReveal edge cases ---

func TestProcessReveal_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ProcessReveal(context.Background(), "nonexistent", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent season")
	}
}

func TestProcessReveal_LargeGroup_50PercentThreshold(t *testing.T) {
	m := newMock()
	// 10 members → threshold = 50%, need 5 voters
	m.memberCounts["g1"] = 10
	m.uniqueVoters["s1"] = 4 // below 50%
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revealed {
		t.Error("expected not revealed (4/10 < 50%)")
	}
	if !result.Retry {
		t.Error("expected retry")
	}
}

func TestProcessReveal_LargeGroup_QuorumMet(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 10
	m.uniqueVoters["s1"] = 5 // exactly 50%
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Revealed {
		t.Error("expected revealed (5/10 = 50%)")
	}
}

// --- splitAttributes extra tests ---

func TestSplitAttributes_ExactlyThree(t *testing.T) {
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 80},
		{QuestionID: "q2", QuestionText: "Q2", Percentage: 60},
		{QuestionID: "q3", QuestionText: "Q3", Percentage: 40},
	}
	top, hidden := splitAttributes(results)
	if len(top) != 3 {
		t.Errorf("expected 3 top, got %d", len(top))
	}
	if len(hidden) != 0 {
		t.Errorf("expected 0 hidden, got %d", len(hidden))
	}
}

func TestSplitAttributes_RankAssignment(t *testing.T) {
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 80},
		{QuestionID: "q2", QuestionText: "Q2", Percentage: 60},
		{QuestionID: "q3", QuestionText: "Q3", Percentage: 40},
		{QuestionID: "q4", QuestionText: "Q4", Percentage: 20},
	}
	top, hidden := splitAttributes(results)
	if top[0].Rank != 1 || top[1].Rank != 2 || top[2].Rank != 3 {
		t.Errorf("unexpected top ranks: %d, %d, %d", top[0].Rank, top[1].Rank, top[2].Rank)
	}
	if hidden[0].Rank != 4 {
		t.Errorf("expected hidden rank 4, got %d", hidden[0].Rank)
	}
}

// =============================================================================
// Additional tests to boost coverage
// =============================================================================

// --- ProcessReveal: small group (<8) uses 40% threshold ---

func TestProcessReveal_SmallGroup_40PercentThreshold_Boundary(t *testing.T) {
	m := newMock()
	// 7 members → threshold = 40%, need 2.8 → 3 voters minimum
	m.memberCounts["g1"] = 7
	m.uniqueVoters["s1"] = 2 // below 40% of 7 (2.8)
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revealed {
		t.Error("expected not revealed (2/7 < 40%)")
	}
	if !result.Retry {
		t.Error("expected retry")
	}
}

func TestProcessReveal_SmallGroup_40PercentThreshold_Met(t *testing.T) {
	m := newMock()
	// 7 members → threshold = 40%, need 2.8 → 3 voters needed
	m.memberCounts["g1"] = 7
	m.uniqueVoters["s1"] = 3 // 3 >= 2.8 → quorum met
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Revealed {
		t.Error("expected revealed (3/7 >= 40%)")
	}
}

// --- ProcessReveal: exactly 8 members uses 50% threshold ---

func TestProcessReveal_ExactlyEightMembers_50Threshold(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 8
	m.uniqueVoters["s1"] = 3 // 3/8 = 37.5% < 50%
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revealed {
		t.Error("expected not revealed (3/8 < 50%)")
	}
	if !result.Retry {
		t.Error("expected retry")
	}
}

// --- ProcessReveal: attempt boundary (attempt=2 still retries, attempt=3 forces) ---

func TestProcessReveal_QuorumNotMet_Attempt2_StillRetries(t *testing.T) {
	m := newMock()
	m.uniqueVoters["s1"] = 0
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revealed {
		t.Error("expected not revealed at attempt 2")
	}
	if !result.Retry {
		t.Error("expected retry at attempt 2")
	}
}

// --- ProcessReveal: zero voters, forced after max attempts ---

func TestProcessReveal_ZeroVoters_PostponedNotForced(t *testing.T) {
	m := newMock()
	m.uniqueVoters["s1"] = 0
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 3)
	if err != nil {
		t.Fatal(err)
	}
	if result.Revealed {
		t.Error("a season with 0 completed voters must never reveal, not even on the forced path")
	}
	if !result.Postponed {
		t.Error("expected the season to be postponed")
	}
	if len(m.updatedStatuses) != 0 {
		t.Errorf("expected no status change, got %v", m.updatedStatuses)
	}
	if len(m.createdResults) != 0 {
		t.Errorf("expected no aggregated results, got %d", len(m.createdResults))
	}
}

// --- ProcessReveal: zero total voters → percentage should be 0 ---

func TestProcessReveal_ZeroVoters_ZeroPercentage(t *testing.T) {
	m := newMock()
	m.uniqueVoters["s1"] = 0
	m.aggregated["s1"] = []db.AggregateVotesByTargetRow{
		{TargetID: "u1", QuestionID: "q1", VoteCount: 0},
	}
	svc := NewService(m, nil)

	// ProcessReveal would postpone here (floor unmet), so exercise aggregation directly
	// to keep the divide-by-zero guard covered.
	if err := svc.aggregateAndReveal(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}

	if len(m.createdResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(m.createdResults))
	}
	if m.createdResults[0].Percentage != 0 {
		t.Errorf("expected 0%% for zero total voters, got %.1f%%", m.createdResults[0].Percentage)
	}
}

// --- ProcessReveal: empty aggregated results ---

func TestProcessReveal_EmptyAggregatedResults(t *testing.T) {
	m := newMock()
	m.aggregated["s1"] = []db.AggregateVotesByTargetRow{}
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Revealed {
		t.Error("expected revealed even with no aggregated results")
	}
	if len(m.createdResults) != 0 {
		t.Errorf("expected 0 results, got %d", len(m.createdResults))
	}
	// Status should still be updated
	if len(m.updatedStatuses) != 1 {
		t.Errorf("expected 1 status update, got %d", len(m.updatedStatuses))
	}
}

// --- ProcessReveal: idempotency - DeleteSeasonResultsBySeason is called ---

func TestProcessReveal_DeletesExistingResults(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.deletedSeasons) != 1 || m.deletedSeasons[0] != "s1" {
		t.Errorf("expected existing results deleted for s1, got %v", m.deletedSeasons)
	}
}

// --- GetReveal: avatar emoji in top per question ---

func TestGetReveal_TopPerQuestionAvatarEmoji(t *testing.T) {
	m := newMock()
	m.topPerQuestion["s2"] = []db.GetTopResultPerQuestionRow{
		{QuestionID: "q1", TargetID: "u1", QuestionText: "Who is funniest?", Username: "alice", Percentage: 80, AvatarEmoji: sql.NullString{String: "X", Valid: true}},
		{QuestionID: "q2", TargetID: "u2", QuestionText: "Who is hottest?", Username: "bob", Percentage: 60, AvatarEmoji: sql.NullString{Valid: false}},
	}
	svc := NewService(m, nil)

	data, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if data.GroupSummary.TopPerQuestion[0].AvatarEmoji == nil || *data.GroupSummary.TopPerQuestion[0].AvatarEmoji != "X" {
		t.Error("expected avatar emoji 'X' for first top question")
	}
	if data.GroupSummary.TopPerQuestion[1].AvatarEmoji != nil {
		t.Error("expected nil avatar emoji for second top question")
	}
}

// --- GetReveal: user with no results ---

func TestGetReveal_NoResults(t *testing.T) {
	m := newMock()
	// u2 is a member but has no results in s2
	m.resultsByUser["s2:u2"] = []db.GetSeasonResultsByUserRow{}
	svc := NewService(m, nil)

	data, err := svc.GetReveal(context.Background(), "s2", "u2")
	if err != nil {
		t.Fatal(err)
	}

	if len(data.MyCard.TopAttributes) != 0 {
		t.Errorf("expected 0 top attributes, got %d", len(data.MyCard.TopAttributes))
	}
	if len(data.MyCard.HiddenAttributes) != 0 {
		t.Errorf("expected 0 hidden attributes, got %d", len(data.MyCard.HiddenAttributes))
	}
	if data.MyCard.ReputationTitle != "Загадка века" {
		t.Errorf("expected default title for empty results, got '%s'", data.MyCard.ReputationTitle)
	}
	if data.MyCard.Trend != nil {
		t.Error("expected nil trend for empty results")
	}
}

// --- GetReveal: card cache found ---

type mockQuerierWithCardCache struct {
	mockQuerier
	cardCaches map[string]db.CardCache // key = userID:seasonID
}

func (m *mockQuerierWithCardCache) GetCardCache(_ context.Context, arg db.GetCardCacheParams) (db.CardCache, error) {
	key := arg.UserID + ":" + arg.SeasonID
	cc, ok := m.cardCaches[key]
	if !ok {
		return db.CardCache{}, sql.ErrNoRows
	}
	return cc, nil
}

func TestGetReveal_CardCacheFound(t *testing.T) {
	base := newMock()
	m := &mockQuerierWithCardCache{
		mockQuerier: *base,
		cardCaches: map[string]db.CardCache{
			"u1:s2": {ImageUrl: "https://cdn.repa.app/cards/u1-s2.png"},
		},
	}
	svc := NewService(m, nil)

	data, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if data.MyCard.CardImageURL != "https://cdn.repa.app/cards/u1-s2.png" {
		t.Errorf("expected card image URL, got '%s'", data.MyCard.CardImageURL)
	}
}

// --- GetMembersCards: empty results ---

func TestGetMembersCards_EmptyResults(t *testing.T) {
	m := newMock()
	m.allResultsWithUsers["s2"] = []db.GetAllSeasonResultsWithUsersRow{}
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

// --- GetMembersCards: not member ---

func TestGetMembersCards_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetMembersCards(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- GetMembersCards: season not found ---

func TestGetMembersCards_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.GetMembersCards(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

// --- GetMembersCards: avatar fields ---

func TestGetMembersCards_AvatarFields(t *testing.T) {
	m := newMock()
	m.allResultsWithUsers["s2"] = []db.GetAllSeasonResultsWithUsersRow{
		{
			TargetID: "u1", QuestionID: "q1", QuestionText: "Q?", QuestionCategory: db.QuestionCategoryHOT,
			Percentage: 80, Username: "alice",
			AvatarEmoji: sql.NullString{String: "Y", Valid: true},
			AvatarUrl:   sql.NullString{String: "https://img.example.com/u1.png", Valid: true},
		},
		{
			TargetID: "u2", QuestionID: "q1", QuestionText: "Q?", QuestionCategory: db.QuestionCategoryFUNNY,
			Percentage: 50, Username: "bob",
			AvatarEmoji: sql.NullString{Valid: false},
			AvatarUrl:   sql.NullString{Valid: false},
		},
	}
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}

	// u1 has avatar fields
	if cards[0].AvatarEmoji == nil || *cards[0].AvatarEmoji != "Y" {
		t.Error("expected avatar emoji 'Y' for u1")
	}
	if cards[0].AvatarURL == nil || *cards[0].AvatarURL != "https://img.example.com/u1.png" {
		t.Error("expected avatar URL for u1")
	}

	// u2 has nil avatar fields
	if cards[1].AvatarEmoji != nil {
		t.Error("expected nil avatar emoji for u2")
	}
	if cards[1].AvatarURL != nil {
		t.Error("expected nil avatar URL for u2")
	}
}

// --- GetMembersCards: title generated per member ---

func TestGetMembersCards_TitlePerMember(t *testing.T) {
	m := newMock()
	m.allResultsWithUsers["s2"] = []db.GetAllSeasonResultsWithUsersRow{
		{TargetID: "u1", QuestionID: "q1", QuestionText: "Q?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 80, Username: "alice"},
		{TargetID: "u2", QuestionID: "q1", QuestionText: "Q?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 50, Username: "bob"},
	}
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if cards[0].ReputationTitle != "Горячая штучка" {
		t.Errorf("expected 'Горячая штучка' for HOT, got '%s'", cards[0].ReputationTitle)
	}
	if cards[1].ReputationTitle != "Душа компании" {
		t.Errorf("expected 'Душа компании' for FUNNY, got '%s'", cards[1].ReputationTitle)
	}
}

// --- OpenHidden: exact balance (5) should succeed ---

func TestOpenHidden_ExactBalance(t *testing.T) {
	m := newMock()
	m.balance["u1"] = 5 // exactly the cost
	svc := NewService(m, nil)

	result, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if result.CrystalBalance != 0 {
		t.Errorf("expected balance 0, got %d", result.CrystalBalance)
	}
}

// --- OpenHidden: not member ---

func TestOpenHidden_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- OpenHidden: season not found ---

func TestOpenHidden_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

// --- OpenHidden: attribute ranks are correct ---

func TestOpenHidden_AttributeRanks(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	result, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	for i, attr := range result.AllAttributes {
		if attr.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, attr.Rank)
		}
	}
}

// --- OpenHidden: crystal log type ---

func TestOpenHidden_CrystalLogType(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(m.createdCrystals) != 1 {
		t.Fatalf("expected 1 crystal log, got %d", len(m.createdCrystals))
	}
	if m.createdCrystals[0].Type != db.CrystalLogTypeSPENDATTRIBUTES {
		t.Errorf("expected type SPEND_ATTRIBUTES, got %s", m.createdCrystals[0].Type)
	}
	if !m.createdCrystals[0].Description.Valid || m.createdCrystals[0].Description.String == "" {
		t.Error("expected non-empty description in crystal log")
	}
}

// --- getNewAchievements tests ---

// achievementMock extends mockQuerier with custom GetSeasonAchievements
type achievementMock struct {
	mockQuerier
	achievements []db.Achievement
}

func (m *achievementMock) GetSeasonAchievements(_ context.Context, _ sql.NullString) ([]db.Achievement, error) {
	return m.achievements, nil
}

func TestGetNewAchievements_Empty(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 0 {
		t.Errorf("expected 0 achievements, got %d", len(achievements))
	}
}

func TestGetNewAchievements_FiltersForUser(t *testing.T) {
	base := newMock()
	m := &achievementMock{
		mockQuerier: *base,
		achievements: []db.Achievement{
			{ID: "a1", UserID: "u1", AchievementType: db.AchievementTypeSNIPER},
			{ID: "a2", UserID: "u2", AchievementType: db.AchievementTypeORACLE},
			{ID: "a3", UserID: "u1", AchievementType: db.AchievementTypeTELEPATH},
		},
	}
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 2 {
		t.Fatalf("expected 2 achievements for u1, got %d", len(achievements))
	}
	if achievements[0].Type != "SNIPER" {
		t.Errorf("expected SNIPER, got %s", achievements[0].Type)
	}
	if achievements[1].Type != "TELEPATH" {
		t.Errorf("expected TELEPATH, got %s", achievements[1].Type)
	}
}

func TestGetNewAchievements_WithMetadata(t *testing.T) {
	base := newMock()
	m := &achievementMock{
		mockQuerier: *base,
		achievements: []db.Achievement{
			{
				ID:              "a1",
				UserID:          "u1",
				AchievementType: db.AchievementTypeEXPERTOF,
				Metadata:        newRawMessage(`{"category":"HOT"}`),
			},
		},
	}
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(achievements))
	}
	if achievements[0].Metadata == nil {
		t.Fatal("expected non-nil metadata")
	}
	meta, ok := achievements[0].Metadata.(map[string]any)
	if !ok {
		t.Fatal("expected metadata to be map[string]any")
	}
	if meta["category"] != "HOT" {
		t.Errorf("expected category HOT, got %v", meta["category"])
	}
}

func TestGetNewAchievements_WithInvalidMetadata(t *testing.T) {
	base := newMock()
	m := &achievementMock{
		mockQuerier: *base,
		achievements: []db.Achievement{
			{
				ID:              "a1",
				UserID:          "u1",
				AchievementType: db.AchievementTypeSNIPER,
				Metadata:        newRawMessage(`not valid json`),
			},
		},
	}
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(achievements))
	}
	// Invalid JSON metadata should result in nil metadata
	if achievements[0].Metadata != nil {
		t.Errorf("expected nil metadata for invalid JSON, got %v", achievements[0].Metadata)
	}
}

func TestGetNewAchievements_NullMetadata(t *testing.T) {
	base := newMock()
	m := &achievementMock{
		mockQuerier: *base,
		achievements: []db.Achievement{
			{
				ID:              "a1",
				UserID:          "u1",
				AchievementType: db.AchievementTypeSNIPER,
				// Metadata is zero-value (Valid=false)
			},
		},
	}
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(achievements))
	}
	if achievements[0].Metadata != nil {
		t.Errorf("expected nil metadata for null Metadata, got %v", achievements[0].Metadata)
	}
}

// achievementErrorMock returns an error from GetSeasonAchievements
type achievementErrorMock struct {
	mockQuerier
}

func (m *achievementErrorMock) GetSeasonAchievements(_ context.Context, _ sql.NullString) ([]db.Achievement, error) {
	return nil, errors.New("db error")
}

func TestGetNewAchievements_DBError_ReturnsEmpty(t *testing.T) {
	base := newMock()
	m := &achievementErrorMock{mockQuerier: *base}
	svc := NewService(m, nil)

	achievements := svc.getNewAchievements(context.Background(), "s2", "u1")
	if len(achievements) != 0 {
		t.Errorf("expected 0 achievements on DB error, got %d", len(achievements))
	}
}

// --- GetDetector: with avatar URL ---

func TestGetDetector_VoterWithAvatarURL(t *testing.T) {
	m := newDetectorMock()
	m.hasDetector["u1:s2"] = true
	m.voterProfiles["s2"] = []db.GetVoterProfilesBySeasonRow{
		{ID: "u2", Username: "bob", AvatarUrl: sql.NullString{String: "https://img.example.com/u2.png", Valid: true}},
	}
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Voters) != 1 {
		t.Fatalf("expected 1 voter, got %d", len(result.Voters))
	}
	if result.Voters[0].AvatarURL == nil || *result.Voters[0].AvatarURL != "https://img.example.com/u2.png" {
		t.Error("expected avatar URL for voter")
	}
}

// --- BuyDetector: voter profiles with avatar data ---

func TestBuyDetector_VoterProfiles(t *testing.T) {
	m := newDetectorMock()
	m.voterProfiles["s2"] = []db.GetVoterProfilesBySeasonRow{
		{ID: "u2", Username: "bob", AvatarEmoji: sql.NullString{String: "Z", Valid: true}, AvatarUrl: sql.NullString{String: "https://img.example.com/u2.png", Valid: true}},
		{ID: "u3", Username: "charlie"},
	}
	svc := NewService(m, nil)

	result, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Voters) != 2 {
		t.Fatalf("expected 2 voters, got %d", len(result.Voters))
	}
	if result.Voters[0].AvatarEmoji == nil || *result.Voters[0].AvatarEmoji != "Z" {
		t.Error("expected avatar emoji 'Z' for first voter")
	}
	if result.Voters[0].AvatarURL == nil || *result.Voters[0].AvatarURL != "https://img.example.com/u2.png" {
		t.Error("expected avatar URL for first voter")
	}
	if result.Voters[1].AvatarEmoji != nil {
		t.Error("expected nil avatar emoji for second voter")
	}
	if result.Voters[1].AvatarURL != nil {
		t.Error("expected nil avatar URL for second voter")
	}
}

// --- BuyDetector: exact balance (10) should succeed ---

func TestBuyDetector_ExactBalance(t *testing.T) {
	m := newDetectorMock()
	m.balance["u1"] = 10
	svc := NewService(m, nil)

	result, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if result.CrystalBalance != 0 {
		t.Errorf("expected balance 0, got %d", result.CrystalBalance)
	}
}

// --- BuyDetector: not revealed ---

func TestBuyDetector_NotRevealed(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

// --- splitAttributes: many hidden (>3 total) ---

func TestSplitAttributes_ManyResults(t *testing.T) {
	results := make([]db.GetSeasonResultsByUserRow, 10)
	for i := range results {
		results[i] = db.GetSeasonResultsByUserRow{
			QuestionID:   "q" + string(rune('0'+i)),
			QuestionText: "Q?",
			Percentage:   float64(100 - i*10),
		}
	}
	top, hidden := splitAttributes(results)
	if len(top) != 3 {
		t.Errorf("expected 3 top, got %d", len(top))
	}
	if len(hidden) != 7 {
		t.Errorf("expected 7 hidden, got %d", len(hidden))
	}
}

// --- splitAttributes: single result ---

func TestSplitAttributes_SingleResult(t *testing.T) {
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 100},
	}
	top, hidden := splitAttributes(results)
	if len(top) != 1 {
		t.Errorf("expected 1 top, got %d", len(top))
	}
	if len(hidden) != 0 {
		t.Errorf("expected 0 hidden, got %d", len(hidden))
	}
	if top[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", top[0].Rank)
	}
}

// helper: create pqtype.NullRawMessage from JSON string
func newRawMessage(s string) pqtype.NullRawMessage {
	return pqtype.NullRawMessage{
		RawMessage: json.RawMessage(s),
		Valid:      true,
	}
}

// --- ValidateRevealAccess: db error (non-ErrNoRows) ---

type errorQuerier struct {
	mockQuerier
	getSeasonErr    error
	isMemberErr     error
	countMembersErr error
	countVotersErr  error
}

func (m *errorQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	if m.getSeasonErr != nil {
		return db.Season{}, m.getSeasonErr
	}
	return m.mockQuerier.GetSeasonByID(nil, id)
}

func (m *errorQuerier) CountCompletedVoters(_ context.Context, arg db.CountCompletedVotersParams) (int64, error) {
	if m.countVotersErr != nil {
		return 0, m.countVotersErr
	}
	return m.mockQuerier.CountCompletedVoters(nil, arg)
}

func (m *errorQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if m.isMemberErr != nil {
		return 0, m.isMemberErr
	}
	return m.mockQuerier.IsGroupMember(nil, arg)
}

func (m *errorQuerier) CountGroupMembers(_ context.Context, groupID string) (int64, error) {
	if m.countMembersErr != nil {
		return 0, m.countMembersErr
	}
	return m.mockQuerier.CountGroupMembers(nil, groupID)
}

func (m *errorQuerier) CountUniqueVoters(_ context.Context, seasonID string) (int64, error) {
	if m.countVotersErr != nil {
		return 0, m.countVotersErr
	}
	return m.mockQuerier.CountUniqueVoters(nil, seasonID)
}

func TestValidateRevealAccess_DBError(t *testing.T) {
	base := newMock()
	m := &errorQuerier{
		mockQuerier:  *base,
		getSeasonErr: errors.New("connection refused"),
	}
	svc := NewService(m, nil)

	_, err := svc.ValidateRevealAccess(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "connection refused" {
		t.Errorf("expected connection refused error, got %v", err)
	}
}

func TestValidateRevealAccess_IsGroupMemberError(t *testing.T) {
	base := newMock()
	m := &errorQuerier{
		mockQuerier: *base,
		isMemberErr: errors.New("query timeout"),
	}
	svc := NewService(m, nil)

	_, err := svc.ValidateRevealAccess(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "query timeout" {
		t.Errorf("expected query timeout error, got %v", err)
	}
}

// --- ProcessReveal: CountGroupMembers error ---

func TestProcessReveal_CountMembersError(t *testing.T) {
	base := newMock()
	m := &errorQuerier{
		mockQuerier:     *base,
		countMembersErr: errors.New("db error"),
	}
	svc := NewService(m, nil)

	_, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

// --- ProcessReveal: CountUniqueVoters error ---

func TestProcessReveal_CountVotersError(t *testing.T) {
	base := newMock()
	m := &errorQuerier{
		mockQuerier:    *base,
		countVotersErr: errors.New("voters error"),
	}
	svc := NewService(m, nil)

	_, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err == nil || err.Error() != "voters error" {
		t.Errorf("expected voters error, got %v", err)
	}
}

// --- GetReveal: GetSeasonResultsByUser error ---

type resultsErrorQuerier struct {
	mockQuerier
	getResultsErr error
}

func (m *resultsErrorQuerier) GetSeasonResultsByUser(_ context.Context, _ db.GetSeasonResultsByUserParams) ([]db.GetSeasonResultsByUserRow, error) {
	if m.getResultsErr != nil {
		return nil, m.getResultsErr
	}
	return nil, nil
}

func TestGetReveal_GetSeasonResultsByUser_Error(t *testing.T) {
	base := newMock()
	m := &resultsErrorQuerier{
		mockQuerier:   *base,
		getResultsErr: errors.New("results db error"),
	}
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "results db error" {
		t.Errorf("expected results db error, got %v", err)
	}
}

// --- GetReveal: GetTopResultPerQuestion error ---

type topPerQErrorQuerier struct {
	mockQuerier
}

func (m *topPerQErrorQuerier) GetTopResultPerQuestion(_ context.Context, _ string) ([]db.GetTopResultPerQuestionRow, error) {
	return nil, errors.New("top per q error")
}

func TestGetReveal_GetTopResultPerQuestion_Error(t *testing.T) {
	base := newMock()
	m := &topPerQErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "top per q error" {
		t.Errorf("expected top per q error, got %v", err)
	}
}

// --- GetMembersCards: GetAllSeasonResultsWithUsers error ---

type allResultsErrorQuerier struct {
	mockQuerier
}

func (m *allResultsErrorQuerier) GetAllSeasonResultsWithUsers(_ context.Context, _ string) ([]db.GetAllSeasonResultsWithUsersRow, error) {
	return nil, errors.New("all results error")
}

func TestGetMembersCards_GetAllResults_Error(t *testing.T) {
	base := newMock()
	m := &allResultsErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "all results error" {
		t.Errorf("expected all results error, got %v", err)
	}
}

// --- OpenHidden: GetUserBalance error ---

type balanceErrorQuerier struct {
	mockQuerier
}

func (m *balanceErrorQuerier) GetUserBalance(_ context.Context, _ string) (int32, error) {
	return 0, errors.New("balance error")
}

func TestOpenHidden_GetBalanceError(t *testing.T) {
	base := newMock()
	m := &balanceErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "balance error" {
		t.Errorf("expected balance error, got %v", err)
	}
}

// --- OpenHidden: CreateCrystalLog error ---

type crystalLogErrorQuerier struct {
	mockQuerier
}

func (m *crystalLogErrorQuerier) CreateCrystalLog(_ context.Context, _ db.CreateCrystalLogParams) (db.CrystalLog, error) {
	return db.CrystalLog{}, errors.New("crystal log error")
}

func TestOpenHidden_CreateCrystalLogError(t *testing.T) {
	base := newMock()
	m := &crystalLogErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "crystal log error" {
		t.Errorf("expected crystal log error, got %v", err)
	}
}

// --- GetDetector: HasDetector error ---

type detectorErrorMock struct {
	mockQuerier
	hasDetectorErr error
}

func (m *detectorErrorMock) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, m.hasDetectorErr
}

func TestGetDetector_HasDetectorError(t *testing.T) {
	base := newMock()
	m := &detectorErrorMock{
		mockQuerier:    *base,
		hasDetectorErr: errors.New("has detector error"),
	}
	svc := NewService(m, nil)

	_, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "has detector error" {
		t.Errorf("expected has detector error, got %v", err)
	}
}

// --- GetDetector: season not found ---

func TestGetDetector_SeasonNotFound(t *testing.T) {
	m := newDetectorMock()
	svc := NewService(m, nil)

	_, err := svc.GetDetector(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

// --- GetReveal: CountUniqueVoters error ---

type voterCountErrorQuerier struct {
	mockQuerier
}

func (m *voterCountErrorQuerier) CountUniqueVoters(_ context.Context, _ string) (int64, error) {
	return 0, errors.New("voter count error")
}

func TestGetReveal_CountUniqueVotersError(t *testing.T) {
	base := newMock()
	m := &voterCountErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.GetReveal(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "voter count error" {
		t.Errorf("expected voter count error, got %v", err)
	}
}

// --- OpenHidden: GetSeasonResultsByUser error after deduction ---

type openHiddenResultsErrorQuerier struct {
	mockQuerier
}

func (m *openHiddenResultsErrorQuerier) GetSeasonResultsByUser(_ context.Context, arg db.GetSeasonResultsByUserParams) ([]db.GetSeasonResultsByUserRow, error) {
	return nil, errors.New("results error after deduction")
}

func TestOpenHidden_GetResultsAfterDeduction_Error(t *testing.T) {
	base := newMock()
	m := &openHiddenResultsErrorQuerier{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.OpenHidden(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "results error after deduction" {
		t.Errorf("expected results error after deduction, got %v", err)
	}
}

// --- BuyDetector: HasDetector error ---

type buyDetectorHasErr struct {
	mockQuerier
}

func (m *buyDetectorHasErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, errors.New("has detector error")
}

func TestBuyDetector_HasDetectorError(t *testing.T) {
	base := newMock()
	m := &buyDetectorHasErr{mockQuerier: *base}
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "has detector error" {
		t.Errorf("expected has detector error, got %v", err)
	}
}

// --- BuyDetector: GetUserBalance error ---

type buyDetectorBalanceErr struct {
	detectorMock
}

func (m *buyDetectorBalanceErr) GetUserBalance(_ context.Context, _ string) (int32, error) {
	return 0, errors.New("balance error")
}

func (m *buyDetectorBalanceErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, nil
}

func TestBuyDetector_GetBalanceError(t *testing.T) {
	base := newDetectorMock()
	m := &buyDetectorBalanceErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "balance error" {
		t.Errorf("expected balance error, got %v", err)
	}
}

// --- BuyDetector: CreateCrystalLog error ---

type buyDetectorCrystalErr struct {
	detectorMock
}

func (m *buyDetectorCrystalErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, nil
}

func (m *buyDetectorCrystalErr) CreateCrystalLog(_ context.Context, _ db.CreateCrystalLogParams) (db.CrystalLog, error) {
	return db.CrystalLog{}, errors.New("crystal log error")
}

func TestBuyDetector_CreateCrystalLogError(t *testing.T) {
	base := newDetectorMock()
	m := &buyDetectorCrystalErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "crystal log error" {
		t.Errorf("expected crystal log error, got %v", err)
	}
}

// --- BuyDetector: CreateDetector error ---

type buyDetectorCreateErr struct {
	detectorMock
}

func (m *buyDetectorCreateErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, nil
}

func (m *buyDetectorCreateErr) CreateDetector(_ context.Context, _ db.CreateDetectorParams) (db.Detector, error) {
	return db.Detector{}, errors.New("create detector error")
}

func TestBuyDetector_CreateDetectorError(t *testing.T) {
	base := newDetectorMock()
	m := &buyDetectorCreateErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "create detector error" {
		t.Errorf("expected create detector error, got %v", err)
	}
}

// --- BuyDetector: GetVoterProfilesBySeason error ---

type buyDetectorVotersErr struct {
	detectorMock
}

func (m *buyDetectorVotersErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, nil
}

func (m *buyDetectorVotersErr) GetVoterProfilesBySeason(_ context.Context, _ string) ([]db.GetVoterProfilesBySeasonRow, error) {
	return nil, errors.New("voters error")
}

func TestBuyDetector_GetVoterProfilesError(t *testing.T) {
	base := newDetectorMock()
	m := &buyDetectorVotersErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "voters error" {
		t.Errorf("expected voters error, got %v", err)
	}
}

// --- GetDetector: GetVoterProfilesBySeason error when purchased ---

type getDetectorVotersErr struct {
	detectorMock
}

func (m *getDetectorVotersErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return true, nil
}

func (m *getDetectorVotersErr) GetVoterProfilesBySeason(_ context.Context, _ string) ([]db.GetVoterProfilesBySeasonRow, error) {
	return nil, errors.New("get voters error")
}

func TestGetDetector_GetVoterProfilesError(t *testing.T) {
	base := newDetectorMock()
	m := &getDetectorVotersErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err == nil || err.Error() != "get voters error" {
		t.Errorf("expected get voters error, got %v", err)
	}
}

// --- GetDetector: GetUserBalance error ---

type getDetectorBalanceErr struct {
	detectorMock
}

func (m *getDetectorBalanceErr) HasDetector(_ context.Context, _ db.HasDetectorParams) (bool, error) {
	return false, nil
}

func (m *getDetectorBalanceErr) GetUserBalance(_ context.Context, _ string) (int32, error) {
	return 0, errors.New("balance db error")
}

func TestGetDetector_GetBalanceError(t *testing.T) {
	base := newDetectorMock()
	m := &getDetectorBalanceErr{detectorMock: *base}
	svc := NewService(m, nil)

	_, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err == nil || !strings.Contains(err.Error(), "balance db error") {
		t.Errorf("expected balance db error, got %v", err)
	}
}

// --- computeTrend: previous season has empty results ---

func TestComputeTrend_PrevSeasonEmptyResults(t *testing.T) {
	m := newMock()
	m.prevSeason = map[string]db.Season{
		"g1": {ID: "s-prev", GroupID: "g1", Status: db.SeasonStatusREVEALED},
	}
	m.resultsByUser["s-prev:u1"] = []db.GetSeasonResultsByUserRow{}
	svc := NewService(m, nil)

	season := db.Season{ID: "s2", GroupID: "g1"}
	results := []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Funniest?", Percentage: 80},
	}

	trend := svc.computeTrend(context.Background(), season, "u1", results)
	if trend != nil {
		t.Error("expected nil trend when previous results are empty")
	}
}

// --- Eligibility: participation floors ---

func TestEligibility_Table(t *testing.T) {
	tests := []struct {
		name              string
		members           int64
		voters            int64
		wantMembersNeeded int64
		wantVotersNeeded  int64
		wantFloorMet      bool
		wantQuorumMet     bool
		wantEligible      bool
	}{
		{"solo group", 1, 0, 2, 3, false, false, false},
		{"two members both voted", 2, 2, 1, 1, false, true, false},
		{"three members two voted", 3, 2, 0, 1, false, true, false},
		{"three members all voted", 3, 3, 0, 0, true, true, true},
		{"four members two voted meets 40pct but not floor", 4, 2, 0, 1, false, true, false},
		{"six members four voted", 6, 4, 0, 0, true, true, true},
		{"ten members three voted floor met quorum short", 10, 3, 0, 0, true, false, false},
		{"eight members four voted 50pct boundary", 8, 4, 0, 0, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMock()
			m.memberCounts["g1"] = tt.members
			m.uniqueVoters["s1"] = tt.voters
			svc := NewService(m, nil)

			e, err := svc.Eligibility(context.Background(), "s1")
			if err != nil {
				t.Fatal(err)
			}

			if e.MemberCount != tt.members {
				t.Errorf("MemberCount = %d, want %d", e.MemberCount, tt.members)
			}
			if e.VotedCount != tt.voters {
				t.Errorf("VotedCount = %d, want %d", e.VotedCount, tt.voters)
			}
			if e.MembersNeeded != tt.wantMembersNeeded {
				t.Errorf("MembersNeeded = %d, want %d", e.MembersNeeded, tt.wantMembersNeeded)
			}
			if e.VotersNeeded != tt.wantVotersNeeded {
				t.Errorf("VotersNeeded = %d, want %d", e.VotersNeeded, tt.wantVotersNeeded)
			}
			if e.FloorMet != tt.wantFloorMet {
				t.Errorf("FloorMet = %v, want %v", e.FloorMet, tt.wantFloorMet)
			}
			if e.QuorumMet != tt.wantQuorumMet {
				t.Errorf("QuorumMet = %v, want %v", e.QuorumMet, tt.wantQuorumMet)
			}
			if e.Eligible != tt.wantEligible {
				t.Errorf("Eligible = %v, want %v", e.Eligible, tt.wantEligible)
			}
		})
	}
}

// --- ProcessReveal: a solo group is never revealed, at any attempt ---

func TestProcessReveal_SoloGroup_NeverRevealedAtAnyAttempt(t *testing.T) {
	for attempt := 1; attempt <= 4; attempt++ {
		m := newMock()
		m.memberCounts["g1"] = 1
		m.uniqueVoters["s1"] = 0
		svc := NewService(m, nil)

		result, err := svc.ProcessReveal(context.Background(), "s1", attempt)
		if err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
		if result.Revealed {
			t.Errorf("attempt %d: solo group must never reveal", attempt)
		}
		if len(m.createdResults) != 0 {
			t.Errorf("attempt %d: no card data should be produced, got %d results", attempt, len(m.createdResults))
		}
	}
}

// --- ProcessReveal: 4 members / 2 voters satisfies 40% quorum but not the voter floor ---

func TestProcessReveal_QuorumMetButFloorUnmet_Postponed(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 4
	m.uniqueVoters["s1"] = 2
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 3)
	if err != nil {
		t.Fatal(err)
	}

	if result.Revealed {
		t.Error("expected no reveal: only 2 members completed voting")
	}
	if !result.Postponed {
		t.Error("expected the season to be postponed")
	}
	if result.VotersNeeded != 1 {
		t.Errorf("VotersNeeded = %d, want 1", result.VotersNeeded)
	}
	if len(m.postponed) != 1 {
		t.Fatalf("expected 1 postponement, got %d", len(m.postponed))
	}
	if !m.postponed[0].RevealAt.After(time.Now()) {
		t.Errorf("postponed reveal time %s should be in the future", m.postponed[0].RevealAt)
	}
	if m.seasons["s1"].Status != db.SeasonStatusVOTING {
		t.Errorf("status should stay VOTING, got %s", m.seasons["s1"].Status)
	}
	if m.seasons["s1"].PostponeCount != 1 {
		t.Errorf("PostponeCount = %d, want 1", m.seasons["s1"].PostponeCount)
	}
}

// --- ProcessReveal: postponed season reveals once the floor is reached ---

func TestProcessReveal_PostponedThenEligible_Reveals(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 2
	m.uniqueVoters["s1"] = 2
	svc := NewService(m, nil)

	first, err := svc.ProcessReveal(context.Background(), "s1", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Postponed {
		t.Fatalf("expected first pass to postpone, got %+v", first)
	}

	// The group grew and a third member finished voting.
	m.memberCounts["g1"] = 3
	m.uniqueVoters["s1"] = 3

	second, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Revealed {
		t.Errorf("expected reveal once the floor was met, got %+v", second)
	}
}

// --- ProcessReveal: result carries the group and kind the worker needs ---

func TestProcessReveal_ResultCarriesGroupAndKind(t *testing.T) {
	m := newMock()
	s := m.seasons["s1"]
	s.Kind = db.SeasonKindKICKOFF
	m.seasons["s1"] = s
	svc := NewService(m, nil)

	result, err := svc.ProcessReveal(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Revealed {
		t.Fatalf("expected reveal, got %+v", result)
	}
	if result.GroupID != "g1" {
		t.Errorf("GroupID = %q, want g1", result.GroupID)
	}
	if result.Kind != db.SeasonKindKICKOFF {
		t.Errorf("Kind = %q, want KICKOFF", result.Kind)
	}
}

// --- ScheduleKickoffIfEligible ---

func (m *mockQuerier) ScheduleKickoffReveal(_ context.Context, arg db.ScheduleKickoffRevealParams) (db.Season, error) {
	s, ok := m.seasons[arg.ID]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	// Mirror the SQL guards: kickoff, still voting, and only ever moves earlier.
	if s.Kind != db.SeasonKindKICKOFF || s.Status != db.SeasonStatusVOTING || !s.RevealAt.After(arg.RevealAt) {
		return db.Season{}, sql.ErrNoRows
	}
	s.RevealAt = arg.RevealAt
	s.EndsAt = arg.EndsAt
	m.seasons[arg.ID] = s
	m.kickoffScheduled = append(m.kickoffScheduled, arg)
	return s, nil
}

func newKickoffMock(members, voters int64) *mockQuerier {
	m := newMock()
	s := m.seasons["s1"]
	s.Kind = db.SeasonKindKICKOFF
	s.RevealAt = time.Now().Add(5 * 24 * time.Hour) // weekly fallback Friday
	m.seasons["s1"] = s
	m.memberCounts["g1"] = members
	m.uniqueVoters["s1"] = voters
	return m
}

func TestScheduleKickoffIfEligible_SchedulesOnThirdVoter(t *testing.T) {
	m := newKickoffMock(3, 2)
	svc := NewService(m, nil)

	// Two voters: not eligible yet.
	scheduled, err := svc.ScheduleKickoffIfEligible(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if scheduled {
		t.Error("must not schedule with only 2 completed voters")
	}

	// Third voter finishes.
	m.uniqueVoters["s1"] = 3
	before := time.Now()
	scheduled, err = svc.ScheduleKickoffIfEligible(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if !scheduled {
		t.Fatal("expected the third completing voter to schedule the reveal")
	}
	if len(m.kickoffScheduled) != 1 {
		t.Fatalf("expected 1 scheduling update, got %d", len(m.kickoffScheduled))
	}

	got := m.kickoffScheduled[0].RevealAt
	wantMin := before.Add(schedule.KickoffRevealDelay)
	if got.Before(wantMin) || got.After(wantMin.Add(time.Minute)) {
		t.Errorf("reveal scheduled at %s, want ~%s (KickoffRevealDelay from now)", got, wantMin)
	}
}

func TestScheduleKickoffIfEligible_LaterVoterDoesNotPushItBack(t *testing.T) {
	m := newKickoffMock(5, 3)
	svc := NewService(m, nil)

	if _, err := svc.ScheduleKickoffIfEligible(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	firstRevealAt := m.seasons["s1"].RevealAt

	// A fourth member finishes later; the reveal time must not move.
	m.uniqueVoters["s1"] = 4
	scheduled, err := svc.ScheduleKickoffIfEligible(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if scheduled {
		t.Error("a later voter must not reschedule the reveal")
	}
	if !m.seasons["s1"].RevealAt.Equal(firstRevealAt) {
		t.Errorf("reveal time moved from %s to %s", firstRevealAt, m.seasons["s1"].RevealAt)
	}
	if len(m.kickoffScheduled) != 1 {
		t.Errorf("expected exactly 1 scheduling update, got %d", len(m.kickoffScheduled))
	}
}

func TestScheduleKickoffIfEligible_IgnoresWeeklySeasons(t *testing.T) {
	m := newKickoffMock(5, 5)
	s := m.seasons["s1"]
	s.Kind = db.SeasonKindWEEKLY
	m.seasons["s1"] = s
	svc := NewService(m, nil)

	scheduled, err := svc.ScheduleKickoffIfEligible(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if scheduled {
		t.Error("weekly seasons keep their Friday reveal")
	}
	if len(m.kickoffScheduled) != 0 {
		t.Errorf("expected no scheduling update, got %d", len(m.kickoffScheduled))
	}
}

func TestScheduleKickoffIfEligible_IgnoresRevealedSeasons(t *testing.T) {
	m := newKickoffMock(5, 5)
	s := m.seasons["s1"]
	s.Status = db.SeasonStatusREVEALED
	m.seasons["s1"] = s
	svc := NewService(m, nil)

	scheduled, err := svc.ScheduleKickoffIfEligible(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if scheduled {
		t.Error("an already revealed season must not be rescheduled")
	}
}

// --- Share recording ---

func (m *mockQuerier) CreateShareEvent(_ context.Context, arg db.CreateShareEventParams) (db.ShareEvent, error) {
	if m.shareEventErr != nil {
		return db.ShareEvent{}, m.shareEventErr
	}
	m.shareEvents = append(m.shareEvents, arg)
	return db.ShareEvent{
		ID:       arg.ID,
		UserID:   arg.UserID,
		SeasonID: arg.SeasonID,
		Channel:  arg.Channel,
	}, nil
}

func TestNormalizeShareChannel(t *testing.T) {
	tests := map[string]string{
		"card":       "card",
		"telegram":   "telegram",
		"link":       "link",
		"code":       "code",
		"CARD":       "card",
		" Telegram ": "telegram",
		// An unknown channel is recorded, not refused: the share already happened.
		"":          "other",
		"instagram": "other",
		"whatever":  "other",
	}
	for in, want := range tests {
		if got := NormalizeShareChannel(in); got != want {
			t.Errorf("NormalizeShareChannel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRecordShare_StoresTheEvent(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	if err := svc.RecordShare(context.Background(), "s2", "u1", "card"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.shareEvents) != 1 {
		t.Fatalf("expected 1 share event, got %d", len(m.shareEvents))
	}
	ev := m.shareEvents[0]
	if ev.SeasonID != "s2" || ev.UserID != "u1" || ev.Channel != "card" {
		t.Errorf("unexpected event %+v", ev)
	}
}

func TestRecordShare_RecordsEveryAction(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	for i := 0; i < 2; i++ {
		if err := svc.RecordShare(context.Background(), "s2", "u1", "card"); err != nil {
			t.Fatalf("share %d: %v", i, err)
		}
	}

	if len(m.shareEvents) != 2 {
		t.Errorf("sharing twice should record twice, got %d", len(m.shareEvents))
	}
}

func TestRecordShare_SwallowsStorageFailure(t *testing.T) {
	m := newMock()
	m.shareEventErr = errors.New("db down")
	svc := NewService(m, nil)

	// The user's share already happened in the OS share sheet; reporting an error here would
	// describe a failure that did not occur.
	if err := svc.RecordShare(context.Background(), "s2", "u1", "card"); err != nil {
		t.Errorf("expected the failure to be swallowed, got %v", err)
	}
}

func TestRecordShare_StillValidatesAccess(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)

	// s1 is VOTING, not revealed — there is no card to share yet.
	if err := svc.RecordShare(context.Background(), "s1", "u1", "card"); !errors.Is(err, ErrSeasonNotRevealed) {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}

	// A non-member must not be able to record against someone else's season.
	if err := svc.RecordShare(context.Background(), "s2", "stranger", "card"); !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}

	if len(m.shareEvents) != 0 {
		t.Errorf("no event should be recorded for a refused share, got %d", len(m.shareEvents))
	}
}

// --- Detector ladder ---

func TestDetectorPrices_HintIsTheCheapRung(t *testing.T) {
	if DetectorHintCost >= DetectorFullCost {
		t.Errorf("a hint (%d) must cost less than the full list (%d)", DetectorHintCost, DetectorFullCost)
	}
	// Below the referral grant on purpose: one invited friend should buy a rung.
	const referralGrant = 5
	if DetectorHintCost >= referralGrant {
		t.Errorf("a hint (%d) should be affordable from one referral grant (%d)", DetectorHintCost, referralGrant)
	}
}

func TestFirstRune(t *testing.T) {
	tests := map[string]string{
		"alice": "a",
		"Маша":  "М", // a multi-byte name must not be cut in half
		"7even": "7",
		"":      "",
		"🍆bob":  "🍆",
	}
	for in, want := range tests {
		if got := firstRune(in); got != want {
			t.Errorf("firstRune(%q) = %q, want %q", in, got, want)
		}
	}
}

// ladderMock returns a detector mock where userID has `voters` people who voted about them.
func ladderMock(userID, seasonID string, voters int) *detectorMock {
	m := newDetectorMock()
	m.seasonVoters[seasonID] = int64(voters)
	key := userID + ":" + seasonID
	for i := 0; i < voters; i++ {
		m.unrevealed[key] = append(m.unrevealed[key], db.PickUnrevealedVoterRow{
			ID:          fmt.Sprintf("voter-%d", i),
			Username:    fmt.Sprintf("Маша%d", i),
			AvatarEmoji: sql.NullString{String: "\U0001F346", Valid: true},
		})
	}
	return m
}

func TestGetDetector_UntouchedLadderReportsItsPrices(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if result.VoterCount != 3 {
		t.Errorf("VoterCount = %d, want 3 — the free rung must be visible without a purchase", result.VoterCount)
	}
	if result.HintCost != DetectorHintCost || result.FullCost != DetectorFullCost {
		t.Errorf("prices = %d/%d, want %d/%d", result.HintCost, result.FullCost, DetectorHintCost, DetectorFullCost)
	}
	if !result.HintAvailable {
		t.Error("a hint should be available when voters remain unrevealed")
	}
	if len(result.Hints) != 0 {
		t.Errorf("expected no revealed hints yet, got %d", len(result.Hints))
	}
	if result.Purchased {
		t.Error("nothing has been bought")
	}
}

func TestBuyDetectorHint_RevealsOneVoterPartially(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.balance["u1"] = 10
	svc := NewService(m, nil)

	result, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Hints) != 1 {
		t.Fatalf("expected 1 revealed hint, got %d", len(result.Hints))
	}
	hint := result.Hints[0]
	if hint.FirstLetter != "М" {
		t.Errorf("FirstLetter = %q, want М", hint.FirstLetter)
	}
	if hint.AvatarEmoji == nil {
		t.Error("a hint should carry the avatar")
	}
	// The full voter list must not come with a hint.
	if len(result.Voters) != 0 {
		t.Errorf("a hint must not return the full list, got %d voters", len(result.Voters))
	}
	if result.Purchased {
		t.Error("a hint is not the full purchase")
	}
}

func TestBuyDetectorHint_DoesNotRepeat(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.balance["u1"] = 20
	svc := NewService(m, nil)

	if _, err := svc.BuyDetectorHint(context.Background(), "s2", "u1"); err != nil {
		t.Fatal(err)
	}
	result, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Hints) != 2 {
		t.Fatalf("expected 2 revealed hints, got %d", len(result.Hints))
	}
	if m.createdHints[0].RevealedUserID == m.createdHints[1].RevealedUserID {
		t.Error("two hints revealed the same voter; hints are drawn without replacement")
	}
}

func TestBuyDetectorHint_ExhaustsAndThenRefuses(t *testing.T) {
	m := ladderMock("u1", "s2", 2)
	m.balance["u1"] = 50
	svc := NewService(m, nil)

	for i := 0; i < 2; i++ {
		if _, err := svc.BuyDetectorHint(context.Background(), "s2", "u1"); err != nil {
			t.Fatalf("hint %d: %v", i, err)
		}
	}

	balanceBefore := m.balance["u1"]
	_, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrNothingToReveal) {
		t.Fatalf("expected ErrNothingToReveal, got %v", err)
	}
	if m.balance["u1"] != balanceBefore {
		t.Error("a refused hint must not spend crystals")
	}
	if len(m.createdHints) != 2 {
		t.Errorf("expected exactly 2 hints recorded, got %d", len(m.createdHints))
	}
}

func TestBuyDetectorHint_NoVotersIsRefusedWithoutSpending(t *testing.T) {
	m := ladderMock("u1", "s2", 0)
	m.balance["u1"] = 50
	svc := NewService(m, nil)

	_, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrNothingToReveal) {
		t.Fatalf("expected ErrNothingToReveal, got %v", err)
	}
	if m.balance["u1"] != 50 {
		t.Error("nothing to reveal must cost nothing")
	}
}

func TestBuyDetectorHint_InsufficientFundsSpendsNothing(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.balance["u1"] = DetectorHintCost - 1
	svc := NewService(m, nil)

	_, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	if len(m.createdHints) != 0 {
		t.Error("no voter should be revealed when the balance is short")
	}
}

func TestBuyDetectorHint_RefusedInSmallGroup(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.balance["u1"] = 50
	m.memberCounts["g1"] = MinDetectorMembers - 1
	svc := NewService(m, nil)

	_, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrGroupTooSmall) {
		t.Fatalf("expected ErrGroupTooSmall, got %v", err)
	}
	if len(m.createdHints) != 0 {
		t.Error("a partial reveal in a tiny group identifies someone as surely as a list")
	}
}

func TestGetDetector_SmallGroupStillShowsTheFreeCount(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.memberCounts["g1"] = MinDetectorMembers - 1
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if result.VoterCount != 3 {
		t.Error("a count names nobody, so it stays visible at any group size")
	}
	if result.Available {
		t.Error("paid rungs must be unavailable in a small group")
	}
	if result.HintAvailable {
		t.Error("the hint rung must be unavailable in a small group")
	}
}

func TestBuyDetectorHint_RefusedOnceTheFullListIsOwned(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.balance["u1"] = 50
	m.hasDetector["u1:s2"] = true
	svc := NewService(m, nil)

	_, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrNothingToReveal) {
		t.Fatalf("expected ErrNothingToReveal, got %v", err)
	}
}

func TestGetDetector_CompletedLadderHasNothingLeft(t *testing.T) {
	m := ladderMock("u1", "s2", 3)
	m.hasDetector["u1:s2"] = true
	svc := NewService(m, nil)

	result, err := svc.GetDetector(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if !result.Purchased {
		t.Error("the full list is owned")
	}
	if result.HintAvailable {
		t.Error("nothing is left to hint at once the list is owned")
	}
}

func TestDetectorHint_CarriesNoAnswerOrFullName(t *testing.T) {
	m := ladderMock("u1", "s2", 1)
	m.balance["u1"] = 10
	svc := NewService(m, nil)

	result, err := svc.BuyDetectorHint(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(result.Hints[0])
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)

	// The hint type has no username field at all, so a full name cannot leak through it.
	if strings.Contains(body, "Маша0") {
		t.Errorf("a hint must not carry the full name: %s", body)
	}
	for _, forbidden := range []string{"question", "answer", "percentage", "attribute"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("a hint must never bind a voter to an answer, found %q in %s", forbidden, body)
		}
	}
}

// --- Blocks hide results and shrink effective size ---

func TestGetMembersCards_HidesBlockedMembersButNotMyOwnCard(t *testing.T) {
	m := newMock()
	m.blockedIDs["u1"] = []string{"u2"}
	m.allResultsWithUsers["s2"] = []db.GetAllSeasonResultsWithUsersRow{
		{TargetID: "u1", Username: "alice", QuestionText: "q", Percentage: 50},
		{TargetID: "u2", Username: "bob", QuestionText: "q", Percentage: 50},
		{TargetID: "u3", Username: "carol", QuestionText: "q", Percentage: 50},
	}
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	for _, c := range cards {
		seen[c.UserID] = true
	}
	if seen["u2"] {
		t.Error("a blocked member's card must not be shown")
	}
	if !seen["u1"] {
		t.Error("my own card is never hidden")
	}
	if !seen["u3"] {
		t.Error("an unblocked member's card should still appear")
	}
}

func TestDetectorFloor_CountsEffectiveSizeNotRawMembership(t *testing.T) {
	m := newMock()
	// Five members, but this viewer has blocked three: effectively two people can rate them.
	m.memberCounts["g1"] = 5
	m.blockedIDs["u1"] = []string{"u2", "u3", "u4"}
	m.balance["u1"] = 50
	svc := NewService(m, nil)

	_, err := svc.BuyDetector(context.Background(), "s2", "u1")
	if !errors.Is(err, ErrGroupTooSmall) {
		t.Errorf("a group of five with three blocks behaves like a group of two: got %v", err)
	}
}

func TestDetectorFloor_UnblockedGroupIsStillBigEnough(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 5
	m.balance["u1"] = 50
	svc := NewService(m, nil)

	// No blocks: the raw and effective sizes agree.
	tooSmall, err := svc.detectorTooSmall(context.Background(), "g1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if tooSmall {
		t.Error("five members with no blocks clears the floor")
	}
}
