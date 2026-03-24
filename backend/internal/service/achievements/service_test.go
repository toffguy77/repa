package achievements

import (
	"context"
	"database/sql"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/sqlc-dev/pqtype"
)

// mockQuerier implements the Querier interface for testing achievements.
type mockQuerier struct {
	db.Querier

	seasons              map[string]db.Season
	memberIDs            map[string][]string
	winnerPerQuestion    map[string][]db.GetWinnerPerQuestionRow
	votesByVoter         map[string][]db.GetVotesByVoterInSeasonRow
	seasonResultsByUser  map[string][]db.GetSeasonResultsByUserRow
	topResultPerQuestion map[string][]db.GetTopResultPerQuestionRow
	revealedSeasons      map[string][]db.Season
	lastNSeasons         map[string][]db.Season
	topAttribute         map[string]db.GetTopAttributeForUserRow
	firstCompletedVoter  map[string]string
	firstVoteTime        map[string]time.Time
	memberCounts         map[string]int64
	membersJoinedAfter   map[string]int64
	seasonQuestionCount  map[string]int64
	achievements         map[string][]db.Achievement // key: userID:groupID
	existingAchievements map[string]int64            // key: userID:groupID:type
	groupStats           map[string]db.UserGroupStat // key: userID:groupID
	votesCast            map[string]int64            // key: seasonID:userID
	votesReceived        map[string]int64            // key: seasonID:userID

	createdAchievements []db.CreateAchievementParams
	upsertedStats       []db.UpsertUserGroupStatsParams
}

func (m *mockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	s, ok := m.seasons[id]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) GetGroupMemberIDs(_ context.Context, groupID string) ([]string, error) {
	return m.memberIDs[groupID], nil
}

func (m *mockQuerier) GetWinnerPerQuestion(_ context.Context, seasonID string) ([]db.GetWinnerPerQuestionRow, error) {
	return m.winnerPerQuestion[seasonID], nil
}

func (m *mockQuerier) CountSeasonQuestions(_ context.Context, seasonID string) (int64, error) {
	return m.seasonQuestionCount[seasonID], nil
}

func (m *mockQuerier) GetVotesByVoterInSeason(_ context.Context, arg db.GetVotesByVoterInSeasonParams) ([]db.GetVotesByVoterInSeasonRow, error) {
	key := arg.SeasonID + ":" + arg.VoterID
	return m.votesByVoter[key], nil
}

func (m *mockQuerier) HasAchievementInGroup(_ context.Context, arg db.HasAchievementInGroupParams) (int64, error) {
	key := arg.UserID + ":" + arg.GroupID + ":" + string(arg.AchievementType)
	return m.existingAchievements[key], nil
}

func (m *mockQuerier) CreateAchievement(_ context.Context, arg db.CreateAchievementParams) (db.Achievement, error) {
	m.createdAchievements = append(m.createdAchievements, arg)
	// Track so HasAchievementInGroup sees it for dedup
	key := arg.UserID + ":" + arg.GroupID + ":" + string(arg.AchievementType)
	m.existingAchievements[key]++
	return db.Achievement{}, nil
}

func (m *mockQuerier) GetSeasonResultsByUser(_ context.Context, arg db.GetSeasonResultsByUserParams) ([]db.GetSeasonResultsByUserRow, error) {
	key := arg.SeasonID + ":" + arg.TargetID
	return m.seasonResultsByUser[key], nil
}

func (m *mockQuerier) GetTopResultPerQuestion(_ context.Context, seasonID string) ([]db.GetTopResultPerQuestionRow, error) {
	return m.topResultPerQuestion[seasonID], nil
}

func (m *mockQuerier) GetRevealedSeasonsForGroup(_ context.Context, groupID string) ([]db.Season, error) {
	return m.revealedSeasons[groupID], nil
}

func (m *mockQuerier) HasQuestionBeenToppedInGroup(_ context.Context, arg db.HasQuestionBeenToppedInGroupParams) (int64, error) {
	// Check if any revealed season for this group has this question in top results
	seasons := m.revealedSeasons[arg.GroupID]
	for _, s := range seasons {
		if s.ID == arg.SeasonID {
			continue
		}
		tops := m.topResultPerQuestion[s.ID]
		for _, t := range tops {
			if t.QuestionID == arg.QuestionID {
				return 1, nil
			}
		}
	}
	return 0, nil
}

func (m *mockQuerier) GetLastNRevealedSeasons(_ context.Context, arg db.GetLastNRevealedSeasonsParams) ([]db.Season, error) {
	seasons := m.lastNSeasons[arg.GroupID]
	if int32(len(seasons)) > arg.Limit {
		return seasons[:arg.Limit], nil
	}
	return seasons, nil
}

func (m *mockQuerier) GetTopAttributeForUser(_ context.Context, arg db.GetTopAttributeForUserParams) (db.GetTopAttributeForUserRow, error) {
	key := arg.SeasonID + ":" + arg.TargetID
	row, ok := m.topAttribute[key]
	if !ok {
		return db.GetTopAttributeForUserRow{}, sql.ErrNoRows
	}
	return row, nil
}

func (m *mockQuerier) GetFirstCompletedVoter(_ context.Context, arg db.GetFirstCompletedVoterParams) (string, error) {
	voter, ok := m.firstCompletedVoter[arg.SeasonID]
	if !ok {
		return "", sql.ErrNoRows
	}
	return voter, nil
}

func (m *mockQuerier) GetFirstVoteTimeByUser(_ context.Context, arg db.GetFirstVoteTimeByUserParams) (time.Time, error) {
	key := arg.SeasonID + ":" + arg.VoterID
	t, ok := m.firstVoteTime[key]
	if !ok {
		return time.Time{}, sql.ErrNoRows
	}
	return t, nil
}

func (m *mockQuerier) CountMembersJoinedAfterUser(_ context.Context, arg db.CountMembersJoinedAfterUserParams) (int64, error) {
	key := arg.GroupID + ":" + arg.UserID
	return m.membersJoinedAfter[key], nil
}

func (m *mockQuerier) GetUserGroupStats(_ context.Context, arg db.GetUserGroupStatsParams) (db.UserGroupStat, error) {
	key := arg.UserID + ":" + arg.GroupID
	stats, ok := m.groupStats[key]
	if !ok {
		return db.UserGroupStat{}, sql.ErrNoRows
	}
	return stats, nil
}

func (m *mockQuerier) UpsertUserGroupStats(_ context.Context, arg db.UpsertUserGroupStatsParams) (db.UserGroupStat, error) {
	m.upsertedStats = append(m.upsertedStats, arg)
	return db.UserGroupStat{}, nil
}

func (m *mockQuerier) CountVotesCastByUser(_ context.Context, arg db.CountVotesCastByUserParams) (int64, error) {
	key := arg.SeasonID + ":" + arg.VoterID
	return m.votesCast[key], nil
}

func (m *mockQuerier) CountVotesReceivedByUser(_ context.Context, arg db.CountVotesReceivedByUserParams) (int64, error) {
	key := arg.SeasonID + ":" + arg.TargetID
	return m.votesReceived[key], nil
}

func (m *mockQuerier) GetSeasonAchievements(_ context.Context, seasonID sql.NullString) ([]db.Achievement, error) {
	return nil, nil
}

// --- Fixtures ---

func newMock() *mockQuerier {
	return &mockQuerier{
		seasons: map[string]db.Season{
			"s1": {ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
		},
		memberIDs: map[string][]string{
			"g1": {"u1", "u2", "u3", "u4", "u5"},
		},
		winnerPerQuestion: map[string][]db.GetWinnerPerQuestionRow{
			"s1": {
				{QuestionID: "q1", TargetID: "u1", VoteCount: 3},
				{QuestionID: "q2", TargetID: "u2", VoteCount: 2},
				{QuestionID: "q3", TargetID: "u1", VoteCount: 4},
				{QuestionID: "q4", TargetID: "u3", VoteCount: 2},
				{QuestionID: "q5", TargetID: "u1", VoteCount: 3},
				{QuestionID: "q6", TargetID: "u2", VoteCount: 2},
				{QuestionID: "q7", TargetID: "u1", VoteCount: 3},
				{QuestionID: "q8", TargetID: "u1", VoteCount: 4},
				{QuestionID: "q9", TargetID: "u3", VoteCount: 2},
				{QuestionID: "q10", TargetID: "u2", VoteCount: 3},
			},
		},
		seasonQuestionCount: map[string]int64{"s1": 10},
		votesByVoter:        map[string][]db.GetVotesByVoterInSeasonRow{},
		seasonResultsByUser: map[string][]db.GetSeasonResultsByUserRow{},
		topResultPerQuestion: map[string][]db.GetTopResultPerQuestionRow{
			"s1": {
				{QuestionID: "q1", TargetID: "u1", QuestionText: "Q1", Percentage: 60},
			},
		},
		revealedSeasons:      map[string][]db.Season{},
		lastNSeasons:         map[string][]db.Season{},
		topAttribute:         map[string]db.GetTopAttributeForUserRow{},
		firstCompletedVoter:  map[string]string{},
		firstVoteTime:        map[string]time.Time{},
		memberCounts:         map[string]int64{"g1": 5},
		membersJoinedAfter:   map[string]int64{},
		existingAchievements: map[string]int64{},
		groupStats:           map[string]db.UserGroupStat{},
		votesCast:            map[string]int64{},
		votesReceived:        map[string]int64{},
	}
}

// --- Tests ---

func TestSniper_HighAccuracy(t *testing.T) {
	m := newMock()
	// u1 votes for the winner in 8 out of 10 questions (80%)
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"}, // correct
		{QuestionID: "q2", TargetID: "u2"}, // correct
		{QuestionID: "q3", TargetID: "u1"}, // correct
		{QuestionID: "q4", TargetID: "u3"}, // correct
		{QuestionID: "q5", TargetID: "u1"}, // correct
		{QuestionID: "q6", TargetID: "u2"}, // correct
		{QuestionID: "q7", TargetID: "u1"}, // correct
		{QuestionID: "q8", TargetID: "u1"}, // correct
		{QuestionID: "q9", TargetID: "u1"}, // wrong (u3 won)
		{QuestionID: "q10", TargetID: "u1"}, // wrong (u2 won)
	}
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeSNIPER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SNIPER achievement for u1 with 80% accuracy")
	}
}

func TestTelepath_PerfectAccuracy(t *testing.T) {
	m := newMock()
	// u1 votes correctly for all 10 questions
	votes := make([]db.GetVotesByVoterInSeasonRow, 10)
	for i, w := range m.winnerPerQuestion["s1"] {
		votes[i] = db.GetVotesByVoterInSeasonRow{
			QuestionID: w.QuestionID,
			TargetID:   w.TargetID,
		}
	}
	m.votesByVoter["s1:u1"] = votes
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	foundTelepath := false
	foundSniper := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeTELEPATH {
			foundTelepath = true
		}
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeSNIPER {
			foundSniper = true
		}
	}
	if !foundTelepath {
		t.Error("expected TELEPATH achievement for 100% accuracy")
	}
	if !foundSniper {
		t.Error("expected SNIPER achievement too (100% >= 80%)")
	}
}

func TestBlind_LowAccuracy(t *testing.T) {
	m := newMock()
	// u1 gets only 1 out of 10 correct (10%)
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"}, // correct
		{QuestionID: "q2", TargetID: "u5"}, // wrong
		{QuestionID: "q3", TargetID: "u5"}, // wrong
		{QuestionID: "q4", TargetID: "u5"}, // wrong
		{QuestionID: "q5", TargetID: "u5"}, // wrong
		{QuestionID: "q6", TargetID: "u5"}, // wrong
		{QuestionID: "q7", TargetID: "u5"}, // wrong
		{QuestionID: "q8", TargetID: "u5"}, // wrong
		{QuestionID: "q9", TargetID: "u5"}, // wrong
		{QuestionID: "q10", TargetID: "u5"}, // wrong
	}
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeBLIND {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected BLIND achievement for <20% accuracy")
	}
}

func TestMonopolist_HighPercentage(t *testing.T) {
	m := newMock()
	m.seasonResultsByUser["s1:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 75},
	}
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeMONOPOLIST {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected MONOPOLIST achievement for 75% on a question")
	}
}

func TestNoDuplicateAchievements(t *testing.T) {
	m := newMock()
	// u1 already has MONOPOLIST
	m.existingAchievements["u1:g1:MONOPOLIST"] = 1
	m.seasonResultsByUser["s1:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 75},
	}
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeMONOPOLIST {
			count++
		}
	}
	if count != 0 {
		t.Errorf("expected no duplicate MONOPOLIST, got %d", count)
	}
}

func TestVotingStreak_UpdatesCorrectly(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesCast["s1:u1"] = 10
	m.groupStats["u1:g1"] = db.UserGroupStat{
		SeasonsPlayed:   3,
		VotingStreak:    3,
		MaxVotingStreak: 3,
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	if len(m.upsertedStats) != 1 {
		t.Fatalf("expected 1 stats upsert, got %d", len(m.upsertedStats))
	}

	stats := m.upsertedStats[0]
	if stats.VotingStreak != 4 {
		t.Errorf("expected voting streak 4, got %d", stats.VotingStreak)
	}
	if stats.MaxVotingStreak != 4 {
		t.Errorf("expected max voting streak 4, got %d", stats.MaxVotingStreak)
	}
	if stats.SeasonsPlayed != 4 {
		t.Errorf("expected seasons played 4, got %d", stats.SeasonsPlayed)
	}
}

func TestVotingStreak_ResetsOnMiss(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesCast["s1:u1"] = 0 // didn't vote
	m.groupStats["u1:g1"] = db.UserGroupStat{
		SeasonsPlayed:   5,
		VotingStreak:    5,
		MaxVotingStreak: 5,
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	stats := m.upsertedStats[0]
	if stats.VotingStreak != 0 {
		t.Errorf("expected voting streak reset to 0, got %d", stats.VotingStreak)
	}
	if stats.MaxVotingStreak != 5 {
		t.Errorf("expected max voting streak preserved at 5, got %d", stats.MaxVotingStreak)
	}
}

func TestStreakVoter_MilestoneAt5(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesCast["s1:u1"] = 10
	m.groupStats["u1:g1"] = db.UserGroupStat{
		SeasonsPlayed:   4,
		VotingStreak:    4,
		MaxVotingStreak: 4,
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeSTREAKVOTER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected STREAK_VOTER at milestone 5")
	}
}

func TestFirstVoter(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u2"},
	}
	m.firstCompletedVoter["s1"] = "u1"
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeFIRSTVOTER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected FIRST_VOTER achievement")
	}
}

func TestNightOwl(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u2"},
	}
	// 01:30 MSK = 22:30 UTC
	m.firstVoteTime["s1:u1"] = time.Date(2024, 1, 15, 22, 30, 0, 0, time.UTC)
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeNIGHTOWL {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected NIGHT_OWL achievement for voting at 01:30 MSK")
	}
}

func TestRecruiter(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.membersJoinedAfter["g1:u1"] = 3
	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeRECRUITER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected RECRUITER achievement for 3+ members joined after")
	}
}

func TestLegend_SameTopAttribute5Seasons(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}

	// 5 revealed seasons
	seasons := []db.Season{
		{ID: "s5", GroupID: "g1", Number: 5, Status: db.SeasonStatusREVEALED},
		{ID: "s4", GroupID: "g1", Number: 4, Status: db.SeasonStatusREVEALED},
		{ID: "s3", GroupID: "g1", Number: 3, Status: db.SeasonStatusREVEALED},
		{ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		{ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
	}
	m.lastNSeasons["g1"] = seasons
	for _, s := range seasons {
		m.seasons[s.ID] = s
	}

	// Same top attribute (q1) for all 5 seasons
	for _, s := range seasons {
		m.topAttribute[s.ID+":u1"] = db.GetTopAttributeForUserRow{
			QuestionID: "q1",
			Percentage: 80,
		}
	}

	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeLEGEND {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected LEGEND achievement for same top attribute 5 seasons in a row")
	}
}

// --- Oracle achievement (exercises computeAccuracy) ---

func TestOracle_ThreeConsecutiveHighAccuracy(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}

	// 3 revealed seasons
	seasons := []db.Season{
		{ID: "s3", GroupID: "g1", Number: 3, Status: db.SeasonStatusREVEALED},
		{ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		{ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
	}
	m.lastNSeasons["g1"] = seasons
	for _, s := range seasons {
		m.seasons[s.ID] = s
	}

	// Current season s1: u1 gets 8/10 correct = 80% (>70%)
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"},
		{QuestionID: "q2", TargetID: "u2"},
		{QuestionID: "q3", TargetID: "u1"},
		{QuestionID: "q4", TargetID: "u3"},
		{QuestionID: "q5", TargetID: "u1"},
		{QuestionID: "q6", TargetID: "u2"},
		{QuestionID: "q7", TargetID: "u1"},
		{QuestionID: "q8", TargetID: "u1"},
		{QuestionID: "q9", TargetID: "u5"},
		{QuestionID: "q10", TargetID: "u5"},
	}
	m.votesCast["s1:u1"] = 10

	// Previous seasons: set up winners and votes for computeAccuracy
	// s2 winners
	m.winnerPerQuestion["s2"] = []db.GetWinnerPerQuestionRow{
		{QuestionID: "q1", TargetID: "u1", VoteCount: 3},
		{QuestionID: "q2", TargetID: "u2", VoteCount: 2},
	}
	m.votesByVoter["s2:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"}, // correct
		{QuestionID: "q2", TargetID: "u2"}, // correct
	}

	// s3 winners
	m.winnerPerQuestion["s3"] = []db.GetWinnerPerQuestionRow{
		{QuestionID: "q1", TargetID: "u3", VoteCount: 3},
		{QuestionID: "q2", TargetID: "u3", VoteCount: 2},
	}
	m.votesByVoter["s3:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u3"}, // correct
		{QuestionID: "q2", TargetID: "u3"}, // correct
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeORACLE {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ORACLE achievement for 3 consecutive seasons with >70% accuracy")
	}
}

func TestOracle_NotEnoughSeasons(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}

	// Only 2 revealed seasons (need 3)
	m.lastNSeasons["g1"] = []db.Season{
		{ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		{ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
	}

	// u1 gets 100% accuracy this season
	votes := make([]db.GetVotesByVoterInSeasonRow, 10)
	for i, w := range m.winnerPerQuestion["s1"] {
		votes[i] = db.GetVotesByVoterInSeasonRow{QuestionID: w.QuestionID, TargetID: w.TargetID}
	}
	m.votesByVoter["s1:u1"] = votes
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeORACLE {
			t.Error("should not grant ORACLE with fewer than 3 seasons")
		}
	}
}

func TestOracle_AlreadyHas(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.existingAchievements["u1:g1:ORACLE"] = 1

	// Even with perfect accuracy and enough seasons, should not re-grant
	votes := make([]db.GetVotesByVoterInSeasonRow, 10)
	for i, w := range m.winnerPerQuestion["s1"] {
		votes[i] = db.GetVotesByVoterInSeasonRow{QuestionID: w.QuestionID, TargetID: w.TargetID}
	}
	m.votesByVoter["s1:u1"] = votes
	m.votesCast["s1:u1"] = 10

	m.lastNSeasons["g1"] = []db.Season{
		{ID: "s3", GroupID: "g1"}, {ID: "s2", GroupID: "g1"}, {ID: "s1", GroupID: "g1"},
	}

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.AchievementType == db.AchievementTypeORACLE {
			t.Error("should not grant duplicate ORACLE")
		}
	}
}

func TestOracle_StreakBroken(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}

	seasons := []db.Season{
		{ID: "s3", GroupID: "g1", Number: 3, Status: db.SeasonStatusREVEALED},
		{ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		{ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
	}
	m.lastNSeasons["g1"] = seasons
	for _, s := range seasons {
		m.seasons[s.ID] = s
	}

	// Current season: high accuracy
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"},
		{QuestionID: "q2", TargetID: "u2"},
		{QuestionID: "q3", TargetID: "u1"},
		{QuestionID: "q4", TargetID: "u3"},
		{QuestionID: "q5", TargetID: "u1"},
		{QuestionID: "q6", TargetID: "u2"},
		{QuestionID: "q7", TargetID: "u1"},
		{QuestionID: "q8", TargetID: "u1"},
		{QuestionID: "q9", TargetID: "u5"},
		{QuestionID: "q10", TargetID: "u5"},
	}
	m.votesCast["s1:u1"] = 10

	// s3: low accuracy (breaks streak)
	m.winnerPerQuestion["s3"] = []db.GetWinnerPerQuestionRow{
		{QuestionID: "q1", TargetID: "u1", VoteCount: 3},
		{QuestionID: "q2", TargetID: "u2", VoteCount: 2},
	}
	m.votesByVoter["s3:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u5"}, // wrong
		{QuestionID: "q2", TargetID: "u5"}, // wrong
	}

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.AchievementType == db.AchievementTypeORACLE {
			t.Error("should not grant ORACLE when accuracy streak is broken")
		}
	}
}

// --- computeAccuracy direct tests ---

func TestComputeAccuracy_NoWinners(t *testing.T) {
	m := newMock()
	// No winners for season s99
	svc := NewService(m)
	acc := svc.computeAccuracy(context.Background(), "u1", "s99")
	if acc != 0 {
		t.Errorf("expected 0 accuracy for no winners, got %f", acc)
	}
}

func TestComputeAccuracy_NoVotes(t *testing.T) {
	m := newMock()
	// Winners exist but no votes by user
	m.winnerPerQuestion["s99"] = []db.GetWinnerPerQuestionRow{
		{QuestionID: "q1", TargetID: "u2", VoteCount: 3},
	}
	svc := NewService(m)
	acc := svc.computeAccuracy(context.Background(), "u1", "s99")
	if acc != 0 {
		t.Errorf("expected 0 accuracy for no votes, got %f", acc)
	}
}

func TestComputeAccuracy_PartialCorrect(t *testing.T) {
	m := newMock()
	m.winnerPerQuestion["s99"] = []db.GetWinnerPerQuestionRow{
		{QuestionID: "q1", TargetID: "u2", VoteCount: 3},
		{QuestionID: "q2", TargetID: "u3", VoteCount: 2},
		{QuestionID: "q3", TargetID: "u4", VoteCount: 4},
		{QuestionID: "q4", TargetID: "u5", VoteCount: 2},
	}
	m.votesByVoter["s99:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u2"}, // correct
		{QuestionID: "q2", TargetID: "u1"}, // wrong
		{QuestionID: "q3", TargetID: "u4"}, // correct
		{QuestionID: "q4", TargetID: "u1"}, // wrong
	}

	svc := NewService(m)
	acc := svc.computeAccuracy(context.Background(), "u1", "s99")
	expected := 2.0 / 4.0
	if acc != expected {
		t.Errorf("expected accuracy %f, got %f", expected, acc)
	}
}

// --- CalculateAchievements error paths ---

func TestCalculateAchievements_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent season")
	}
}

// --- Pioneer: user is top and no previous season had the question topped ---

func TestPioneer_FirstTopAttribute(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.topResultPerQuestion["s1"] = []db.GetTopResultPerQuestionRow{
		{QuestionID: "q1", TargetID: "u1", QuestionText: "Q1", Percentage: 60},
	}
	// No previous seasons with this question topped
	m.revealedSeasons["g1"] = []db.Season{}
	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypePIONEER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected PIONEER achievement for first top attribute")
	}
}

func TestPioneer_QuestionAlreadyTopped(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.topResultPerQuestion["s1"] = []db.GetTopResultPerQuestionRow{
		{QuestionID: "q1", TargetID: "u1", QuestionText: "Q1", Percentage: 60},
	}
	// Previous season had q1 topped
	m.revealedSeasons["g1"] = []db.Season{
		{ID: "s0", GroupID: "g1", Number: 0, Status: db.SeasonStatusREVEALED},
	}
	m.topResultPerQuestion["s0"] = []db.GetTopResultPerQuestionRow{
		{QuestionID: "q1", TargetID: "u3", QuestionText: "Q1", Percentage: 50},
	}
	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypePIONEER {
			t.Error("should not grant PIONEER when question already topped in previous season")
		}
	}
}

// --- Recruiter: not enough members ---

func TestRecruiter_NotEnoughMembers(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.membersJoinedAfter["g1:u1"] = 2 // need >= 3
	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeRECRUITER {
			t.Error("should not grant RECRUITER with only 2 members joined after")
		}
	}
}

// --- Legend: different top attributes ---

func TestLegend_DifferentTopAttributes(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}

	seasons := []db.Season{
		{ID: "s5", GroupID: "g1", Number: 5, Status: db.SeasonStatusREVEALED},
		{ID: "s4", GroupID: "g1", Number: 4, Status: db.SeasonStatusREVEALED},
		{ID: "s3", GroupID: "g1", Number: 3, Status: db.SeasonStatusREVEALED},
		{ID: "s2", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
		{ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusREVEALED},
	}
	m.lastNSeasons["g1"] = seasons
	for _, s := range seasons {
		m.seasons[s.ID] = s
	}

	// Different top attribute in s3 (q2 instead of q1)
	for _, s := range seasons {
		qID := "q1"
		if s.ID == "s3" {
			qID = "q2"
		}
		m.topAttribute[s.ID+":u1"] = db.GetTopAttributeForUserRow{
			QuestionID: qID,
			Percentage: 80,
		}
	}
	m.votesCast["s1:u1"] = 0

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeLEGEND {
			t.Error("should not grant LEGEND with different top attributes")
		}
	}
}

// --- Night Owl: daytime vote should not trigger ---

func TestNightOwl_DaytimeVote(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u2"},
	}
	// 14:00 MSK = 11:00 UTC (daytime)
	m.firstVoteTime["s1:u1"] = time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeNIGHTOWL {
			t.Error("should not grant NIGHT_OWL for daytime vote")
		}
	}
}

// --- First Voter: not first ---

func TestFirstVoter_NotFirst(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u2"},
	}
	m.firstCompletedVoter["s1"] = "u2" // someone else was first
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeFIRSTVOTER {
			t.Error("should not grant FIRST_VOTER when someone else was first")
		}
	}
}

// --- Monopolist: below threshold ---

func TestMonopolist_BelowThreshold(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.seasonResultsByUser["s1:u1"] = []db.GetSeasonResultsByUserRow{
		{QuestionID: "q1", QuestionText: "Q1", Percentage: 60}, // below 70%
	}
	m.votesCast["s1:u1"] = 10

	svc := NewService(m)
	_ = svc.CalculateAchievements(context.Background(), "s1")

	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeMONOPOLIST {
			t.Error("should not grant MONOPOLIST with only 60% on a question")
		}
	}
}

// --- Streak voter milestone at 10 ---

func TestStreakVoter_MilestoneAt10(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesCast["s1:u1"] = 10
	m.groupStats["u1:g1"] = db.UserGroupStat{
		SeasonsPlayed:   9,
		VotingStreak:    9,
		MaxVotingStreak: 9,
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, a := range m.createdAchievements {
		if a.UserID == "u1" && a.AchievementType == db.AchievementTypeSTREAKVOTER {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected STREAK_VOTER at milestone 10")
	}
}

// --- updateGroupStats accuracy averaging ---

func TestUpdateGroupStats_AccuracyAveraging(t *testing.T) {
	m := newMock()
	m.memberIDs["g1"] = []string{"u1"}
	m.votesCast["s1:u1"] = 10
	m.groupStats["u1:g1"] = db.UserGroupStat{
		SeasonsPlayed:   2,
		VotingStreak:    2,
		MaxVotingStreak: 2,
		GuessAccuracy:   0.5,
	}

	// u1 gets 80% accuracy this season
	m.votesByVoter["s1:u1"] = []db.GetVotesByVoterInSeasonRow{
		{QuestionID: "q1", TargetID: "u1"},
		{QuestionID: "q2", TargetID: "u2"},
		{QuestionID: "q3", TargetID: "u1"},
		{QuestionID: "q4", TargetID: "u3"},
		{QuestionID: "q5", TargetID: "u1"},
		{QuestionID: "q6", TargetID: "u2"},
		{QuestionID: "q7", TargetID: "u1"},
		{QuestionID: "q8", TargetID: "u1"},
		{QuestionID: "q9", TargetID: "u5"},
		{QuestionID: "q10", TargetID: "u5"},
	}

	svc := NewService(m)
	err := svc.CalculateAchievements(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}

	if len(m.upsertedStats) != 1 {
		t.Fatalf("expected 1 stats upsert, got %d", len(m.upsertedStats))
	}

	stats := m.upsertedStats[0]
	// Rolling average: (0.5 * 2 + 0.8) / 3 = 0.6
	expectedAccuracy := (0.5*2 + 0.8) / 3
	if stats.GuessAccuracy < expectedAccuracy-0.01 || stats.GuessAccuracy > expectedAccuracy+0.01 {
		t.Errorf("expected accuracy ~%.3f, got %.3f", expectedAccuracy, stats.GuessAccuracy)
	}
}

// Ensure unused imports are referenced
var _ pqtype.NullRawMessage
