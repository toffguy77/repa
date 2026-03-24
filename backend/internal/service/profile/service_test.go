package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/sqlc-dev/pqtype"
)

// mockQuerier implements only the methods used by the profile service.
type mockQuerier struct {
	db.Querier

	members      map[string]map[string]bool // groupID -> userID -> true
	users        map[string]db.GetUserProfileInfoRow
	stats        map[string]db.UserGroupStat // "userID:groupID" -> stats
	statsErr     map[string]error
	topAttr      map[string]db.GetTopAttributeAllTimeRow // "userID:groupID"
	topAttrErr   map[string]error
	achievements map[string][]db.Achievement // "userID:groupID"
	achieveErr   error
	history      map[string][]db.GetUserSeasonHistoryRow // "userID:groupID"
	historyErr   error
}

func newMock() *mockQuerier {
	return &mockQuerier{
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true},
		},
		users: map[string]db.GetUserProfileInfoRow{
			"u1": {
				ID:       "u1",
				Username: "alice",
				AvatarEmoji: sql.NullString{String: "🦊", Valid: true},
				AvatarUrl:   sql.NullString{String: "https://example.com/avatar.png", Valid: true},
			},
			"u2": {
				ID:       "u2",
				Username: "bob",
				AvatarEmoji: sql.NullString{Valid: false},
				AvatarUrl:   sql.NullString{Valid: false},
			},
		},
		stats: map[string]db.UserGroupStat{
			"u1:g1": {
				SeasonsPlayed:      3,
				VotingStreak:       2,
				MaxVotingStreak:    5,
				GuessAccuracy:      72.35,
				TotalVotesCast:     20,
				TotalVotesReceived: 15,
			},
		},
		statsErr: map[string]error{},
		topAttr: map[string]db.GetTopAttributeAllTimeRow{
			"u1:g1": {QuestionText: "Most creative?", Percentage: 85.67},
		},
		topAttrErr:   map[string]error{},
		achievements: map[string][]db.Achievement{},
		history:      map[string][]db.GetUserSeasonHistoryRow{},
	}
}

func (m *mockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if g, ok := m.members[arg.GroupID]; ok && g[arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

func (m *mockQuerier) GetUserProfileInfo(_ context.Context, id string) (db.GetUserProfileInfoRow, error) {
	u, ok := m.users[id]
	if !ok {
		return db.GetUserProfileInfoRow{}, sql.ErrNoRows
	}
	return u, nil
}

func (m *mockQuerier) GetUserGroupStats(_ context.Context, arg db.GetUserGroupStatsParams) (db.UserGroupStat, error) {
	key := arg.UserID + ":" + arg.GroupID
	if err, ok := m.statsErr[key]; ok {
		return db.UserGroupStat{}, err
	}
	s, ok := m.stats[key]
	if !ok {
		return db.UserGroupStat{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) GetTopAttributeAllTime(_ context.Context, arg db.GetTopAttributeAllTimeParams) (db.GetTopAttributeAllTimeRow, error) {
	key := arg.TargetID + ":" + arg.GroupID
	if err, ok := m.topAttrErr[key]; ok {
		return db.GetTopAttributeAllTimeRow{}, err
	}
	row, ok := m.topAttr[key]
	if !ok {
		return db.GetTopAttributeAllTimeRow{}, sql.ErrNoRows
	}
	return row, nil
}

func (m *mockQuerier) GetUserAchievements(_ context.Context, arg db.GetUserAchievementsParams) ([]db.Achievement, error) {
	if m.achieveErr != nil {
		return nil, m.achieveErr
	}
	key := arg.UserID + ":" + arg.GroupID
	return m.achievements[key], nil
}

func (m *mockQuerier) GetUserSeasonHistory(_ context.Context, arg db.GetUserSeasonHistoryParams) ([]db.GetUserSeasonHistoryRow, error) {
	if m.historyErr != nil {
		return nil, m.historyErr
	}
	key := arg.TargetID + ":" + arg.GroupID
	return m.history[key], nil
}

// --- GetProfile tests ---

func TestGetProfile_Success(t *testing.T) {
	m := newMock()
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	m.achievements["u1:g1"] = []db.Achievement{
		{
			ID:              "a1",
			UserID:          "u1",
			GroupID:         "g1",
			AchievementType: db.AchievementTypeLEGEND,
			EarnedAt:        now,
			Metadata:        pqtype.NullRawMessage{Valid: false},
		},
	}
	m.history["u1:g1"] = []db.GetUserSeasonHistoryRow{
		{
			SeasonID:         "s1",
			SeasonNumber:     1,
			QuestionText:     "Funniest?",
			QuestionCategory: db.QuestionCategoryFUNNY,
			Percentage:       60.0,
		},
	}

	svc := NewService(m)
	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// User
	if resp.User.Username != "alice" {
		t.Errorf("expected username alice, got %s", resp.User.Username)
	}
	if resp.User.AvatarEmoji == nil || *resp.User.AvatarEmoji != "🦊" {
		t.Errorf("expected avatar emoji 🦊, got %v", resp.User.AvatarEmoji)
	}
	if resp.User.AvatarUrl == nil || *resp.User.AvatarUrl != "https://example.com/avatar.png" {
		t.Errorf("expected avatar url, got %v", resp.User.AvatarUrl)
	}

	// Stats
	if resp.Stats.SeasonsPlayed != 3 {
		t.Errorf("expected seasons_played 3, got %d", resp.Stats.SeasonsPlayed)
	}
	if resp.Stats.VotingStreak != 2 {
		t.Errorf("expected voting_streak 2, got %d", resp.Stats.VotingStreak)
	}
	if resp.Stats.MaxVotingStreak != 5 {
		t.Errorf("expected max_voting_streak 5, got %d", resp.Stats.MaxVotingStreak)
	}
	if resp.Stats.TotalVotesCast != 20 {
		t.Errorf("expected total_votes_cast 20, got %d", resp.Stats.TotalVotesCast)
	}
	if resp.Stats.TotalVotesReceived != 15 {
		t.Errorf("expected total_votes_received 15, got %d", resp.Stats.TotalVotesReceived)
	}

	// Top attribute
	if resp.Stats.TopAttributeAllTime == nil {
		t.Fatal("expected top attribute to be set")
	}
	if resp.Stats.TopAttributeAllTime.QuestionText != "Most creative?" {
		t.Errorf("expected question text 'Most creative?', got %s", resp.Stats.TopAttributeAllTime.QuestionText)
	}

	// Achievements
	if len(resp.Achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(resp.Achievements))
	}
	if resp.Achievements[0].Type != "LEGEND" {
		t.Errorf("expected LEGEND achievement, got %s", resp.Achievements[0].Type)
	}
	if resp.Achievements[0].EarnedAt != "2025-01-15" {
		t.Errorf("expected earned_at 2025-01-15, got %s", resp.Achievements[0].EarnedAt)
	}

	// Season history
	if len(resp.SeasonHistory) != 1 {
		t.Fatalf("expected 1 season card, got %d", len(resp.SeasonHistory))
	}
	if resp.SeasonHistory[0].SeasonID != "s1" {
		t.Errorf("expected season s1, got %s", resp.SeasonHistory[0].SeasonID)
	}
	if resp.SeasonHistory[0].TopAttribute != "Funniest?" {
		t.Errorf("expected top attribute 'Funniest?', got %s", resp.SeasonHistory[0].TopAttribute)
	}
	if resp.SeasonHistory[0].Category != "FUNNY" {
		t.Errorf("expected category FUNNY, got %s", resp.SeasonHistory[0].Category)
	}

	// Legend should be non-empty
	if resp.Legend == "" {
		t.Error("expected non-empty legend")
	}
}

func TestGetProfile_RequesterNotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.GetProfile(context.Background(), "g1", "u1", "outsider")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetProfile_TargetUserNotFound(t *testing.T) {
	m := newMock()
	// requester is a member, target does not exist
	svc := NewService(m)

	_, err := svc.GetProfile(context.Background(), "g1", "nonexistent", "u1")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetProfile_TargetNotMember(t *testing.T) {
	m := newMock()
	// u2 exists in users but is not a member of g2
	m.members["g2"] = map[string]bool{"u1": true}
	svc := NewService(m)

	_, err := svc.GetProfile(context.Background(), "g2", "u2", "u1")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetProfile_StatsNotFound_ZeroStats(t *testing.T) {
	m := newMock()
	// No stats for u2 in g1 — should return zero stats without error
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u2", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Stats.SeasonsPlayed != 0 {
		t.Errorf("expected seasons_played 0, got %d", resp.Stats.SeasonsPlayed)
	}
	if resp.Stats.VotingStreak != 0 {
		t.Errorf("expected voting_streak 0, got %d", resp.Stats.VotingStreak)
	}
	if resp.Stats.TotalVotesCast != 0 {
		t.Errorf("expected total_votes_cast 0, got %d", resp.Stats.TotalVotesCast)
	}
	if resp.Stats.GuessAccuracy != 0 {
		t.Errorf("expected guess_accuracy 0, got %f", resp.Stats.GuessAccuracy)
	}
}

func TestGetProfile_TopAttributeNotFound(t *testing.T) {
	m := newMock()
	// Remove top attribute for u1:g1
	delete(m.topAttr, "u1:g1")
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Stats.TopAttributeAllTime != nil {
		t.Errorf("expected top attribute to be nil, got %+v", resp.Stats.TopAttributeAllTime)
	}
}

func TestGetProfile_AchievementsWithMetadata(t *testing.T) {
	m := newMock()
	meta := map[string]interface{}{"question": "Best dancer?", "count": float64(3)}
	metaJSON, _ := json.Marshal(meta)

	m.achievements["u1:g1"] = []db.Achievement{
		{
			ID:              "a1",
			UserID:          "u1",
			GroupID:         "g1",
			AchievementType: db.AchievementTypeMONOPOLIST,
			EarnedAt:        time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC),
			Metadata: pqtype.NullRawMessage{
				RawMessage: metaJSON,
				Valid:      true,
			},
		},
	}
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(resp.Achievements))
	}
	a := resp.Achievements[0]
	if a.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}
	if a.Metadata["question"] != "Best dancer?" {
		t.Errorf("expected question 'Best dancer?', got %v", a.Metadata["question"])
	}
	if a.Metadata["count"] != float64(3) {
		t.Errorf("expected count 3, got %v", a.Metadata["count"])
	}
}

func TestGetProfile_AchievementsWithoutMetadata(t *testing.T) {
	m := newMock()
	m.achievements["u1:g1"] = []db.Achievement{
		{
			ID:              "a1",
			UserID:          "u1",
			GroupID:         "g1",
			AchievementType: db.AchievementTypeSNIPER,
			EarnedAt:        time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC),
			Metadata:        pqtype.NullRawMessage{Valid: false},
		},
	}
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Achievements) != 1 {
		t.Fatalf("expected 1 achievement, got %d", len(resp.Achievements))
	}
	if resp.Achievements[0].Metadata != nil {
		t.Errorf("expected metadata to be nil, got %v", resp.Achievements[0].Metadata)
	}
	if resp.Achievements[0].EarnedAt != "2025-02-20" {
		t.Errorf("expected earned_at 2025-02-20, got %s", resp.Achievements[0].EarnedAt)
	}
}

func TestGetProfile_SeasonHistory(t *testing.T) {
	m := newMock()
	m.history["u1:g1"] = []db.GetUserSeasonHistoryRow{
		{
			SeasonID:         "s3",
			SeasonNumber:     3,
			QuestionText:     "Smartest?",
			QuestionCategory: db.QuestionCategorySKILLS,
			Percentage:       91.5,
		},
		{
			SeasonID:         "s2",
			SeasonNumber:     2,
			QuestionText:     "Funniest?",
			QuestionCategory: db.QuestionCategoryFUNNY,
			Percentage:       44.44,
		},
	}
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.SeasonHistory) != 2 {
		t.Fatalf("expected 2 season cards, got %d", len(resp.SeasonHistory))
	}

	// First card
	if resp.SeasonHistory[0].SeasonNumber != 3 {
		t.Errorf("expected season_number 3, got %d", resp.SeasonHistory[0].SeasonNumber)
	}
	if resp.SeasonHistory[0].Category != "SKILLS" {
		t.Errorf("expected category SKILLS, got %s", resp.SeasonHistory[0].Category)
	}
	// 91.5 rounds to 91.5
	if resp.SeasonHistory[0].Percentage != 91.5 {
		t.Errorf("expected percentage 91.5, got %f", resp.SeasonHistory[0].Percentage)
	}

	// Second card: 44.44 rounds to 44.4
	if resp.SeasonHistory[1].Percentage != 44.4 {
		t.Errorf("expected percentage 44.4, got %f", resp.SeasonHistory[1].Percentage)
	}
}

func TestGetProfile_PercentageRounding(t *testing.T) {
	m := newMock()
	// GuessAccuracy 72.35 should round to 72.4
	// TopAttribute percentage 85.67 should round to 85.7
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := math.Round(72.35*10) / 10
	if resp.Stats.GuessAccuracy != expected {
		t.Errorf("expected guess_accuracy %f, got %f", expected, resp.Stats.GuessAccuracy)
	}

	if resp.Stats.TopAttributeAllTime == nil {
		t.Fatal("expected top attribute to be set")
	}
	expectedTop := math.Round(85.67*10) / 10
	if resp.Stats.TopAttributeAllTime.Percentage != expectedTop {
		t.Errorf("expected top attr percentage %f, got %f", expectedTop, resp.Stats.TopAttributeAllTime.Percentage)
	}
}

func TestGetProfile_UserWithNilAvatarFields(t *testing.T) {
	m := newMock()
	// u2 has no avatar emoji or url (Valid = false)
	svc := NewService(m)

	resp, err := svc.GetProfile(context.Background(), "g1", "u2", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.User.AvatarEmoji != nil {
		t.Errorf("expected avatar_emoji nil, got %v", resp.User.AvatarEmoji)
	}
	if resp.User.AvatarUrl != nil {
		t.Errorf("expected avatar_url nil, got %v", resp.User.AvatarUrl)
	}
}

// --- generateLegend tests ---

func TestGenerateLegend_LegendAndSniper(t *testing.T) {
	achievements := []AchievementDto{
		{Type: "LEGEND"},
		{Type: "SNIPER"},
	}
	stats := StatsDto{SeasonsPlayed: 10}
	legend := generateLegend("alice", stats, achievements)

	if !strings.Contains(legend, "Снайпер") || !strings.Contains(legend, "Легенда") {
		t.Errorf("expected legend to mention Снайпер and Легенда, got: %s", legend)
	}
	if !strings.HasPrefix(legend, "alice") {
		t.Errorf("expected legend to start with username, got: %s", legend)
	}
}

func TestGenerateLegend_LegendOnly(t *testing.T) {
	achievements := []AchievementDto{{Type: "LEGEND"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "Легенда группы") {
		t.Errorf("expected legend to mention 'Легенда группы', got: %s", legend)
	}
}

func TestGenerateLegend_Telepath(t *testing.T) {
	achievements := []AchievementDto{{Type: "TELEPATH"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "читает мысли") {
		t.Errorf("expected legend to mention 'читает мысли', got: %s", legend)
	}
}

func TestGenerateLegend_Oracle(t *testing.T) {
	achievements := []AchievementDto{{Type: "ORACLE"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "Оракул") {
		t.Errorf("expected legend to mention 'Оракул', got: %s", legend)
	}
}

func TestGenerateLegend_Sniper(t *testing.T) {
	achievements := []AchievementDto{{Type: "SNIPER"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "Снайпер") {
		t.Errorf("expected legend to mention 'Снайпер', got: %s", legend)
	}
}

func TestGenerateLegend_Monopolist(t *testing.T) {
	achievements := []AchievementDto{{Type: "MONOPOLIST"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "Монополист") {
		t.Errorf("expected legend to mention 'Монополист', got: %s", legend)
	}
}

func TestGenerateLegend_StreakVoter(t *testing.T) {
	achievements := []AchievementDto{{Type: "STREAK_VOTER"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "не пропускает") {
		t.Errorf("expected legend to mention 'не пропускает', got: %s", legend)
	}
}

func TestGenerateLegend_NightOwl(t *testing.T) {
	achievements := []AchievementDto{{Type: "NIGHT_OWL"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "Ночная сова") {
		t.Errorf("expected legend to mention 'Ночная сова', got: %s", legend)
	}
}

func TestGenerateLegend_Recruiter(t *testing.T) {
	achievements := []AchievementDto{{Type: "RECRUITER"}}
	stats := StatsDto{}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "душа компании") {
		t.Errorf("expected legend to mention 'душа компании', got: %s", legend)
	}
}

func TestGenerateLegend_SeasonsGte5_NoAchievements(t *testing.T) {
	achievements := []AchievementDto{}
	stats := StatsDto{SeasonsPlayed: 7}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "опытный участник") {
		t.Errorf("expected legend to mention 'опытный участник', got: %s", legend)
	}
	if !strings.Contains(legend, "7 сезонов") {
		t.Errorf("expected legend to mention '7 сезонов', got: %s", legend)
	}
}

func TestGenerateLegend_SeasonsGt0Lt5_NoAchievements(t *testing.T) {
	achievements := []AchievementDto{}
	stats := StatsDto{SeasonsPlayed: 2}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "начинает свой путь") {
		t.Errorf("expected legend to mention 'начинает свой путь', got: %s", legend)
	}
}

func TestGenerateLegend_ZeroSeasons_NoAchievements(t *testing.T) {
	achievements := []AchievementDto{}
	stats := StatsDto{SeasonsPlayed: 0}
	legend := generateLegend("bob", stats, achievements)

	if !strings.Contains(legend, "загадочная личность") {
		t.Errorf("expected legend to mention 'загадочная личность', got: %s", legend)
	}
}

func TestGenerateLegend_TruncationAt150Runes(t *testing.T) {
	// Use a very long username to force truncation
	longName := strings.Repeat("А", 200)
	achievements := []AchievementDto{}
	stats := StatsDto{SeasonsPlayed: 0}
	legend := generateLegend(longName, stats, achievements)

	runes := []rune(legend)
	if len(runes) != 150 {
		t.Errorf("expected legend to be exactly 150 runes, got %d", len(runes))
	}
	if !strings.HasSuffix(legend, "...") {
		t.Errorf("expected legend to end with '...', got: %s", legend[len(legend)-6:])
	}
}
