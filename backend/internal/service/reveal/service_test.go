package reveal

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// mockQuerier implements only the methods used by the reveal service.
type mockQuerier struct {
	db.Querier
	seasons          map[string]db.Season
	members          map[string]map[string]bool
	memberRows       map[string][]db.GetGroupMembersRow
	memberCounts     map[string]int64
	uniqueVoters     map[string]int64
	aggregated       map[string][]db.AggregateVotesByTargetRow
	resultsByUser    map[string][]db.GetSeasonResultsByUserRow
	topPerQuestion   map[string][]db.GetTopResultPerQuestionRow
	prevSeason       map[string]db.Season // groupID -> prev season
	balance          map[string]int32     // userID -> balance
	createdResults   []db.CreateSeasonResultParams
	createdCrystals  []db.CreateCrystalLogParams
	deletedSeasons   []string
	updatedStatuses  []db.UpdateSeasonStatusParams
	seasonsForReveal []db.Season
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

// --- Fixtures ---

func newMock() *mockQuerier {
	return &mockQuerier{
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
	m.uniqueVoters["s1"] = 1
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
	// Add results for all members
	m.resultsByUser["s2:u2"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Who is funniest?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 50},
	}
	m.resultsByUser["s2:u3"] = []db.GetSeasonResultsByUserRow{}
	svc := NewService(m, nil)

	cards, err := svc.GetMembersCards(context.Background(), "s2", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if len(cards) != 3 {
		t.Errorf("expected 3 member cards, got %d", len(cards))
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
