package voting

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// mockQuerier implements only the methods used by the voting service.
// All other Querier methods panic if called.
type mockQuerier struct {
	db.Querier
	seasons        map[string]db.Season
	members        map[string]map[string]bool // groupID -> userID -> true
	memberRows     map[string][]db.GetGroupMembersRow
	questions      map[string][]db.Question // seasonID -> questions
	votes          map[string][]db.Vote     // seasonID:voterID -> votes
	voteForQ       map[string]int64         // seasonID:voterID:questionID -> count
	createdVotes   []db.CreateVoteParams
	memberCounts   map[string]int64 // groupID -> count
	completedCount int64
	userVoteCount  int64
	questionCount  int64
}

func (m *mockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	s, ok := m.seasons[id]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if g, ok := m.members[arg.GroupID]; ok {
		if g[arg.UserID] {
			return 1, nil
		}
	}
	return 0, nil
}

func (m *mockQuerier) GetSeasonQuestions(_ context.Context, seasonID string) ([]db.Question, error) {
	return m.questions[seasonID], nil
}

func (m *mockQuerier) GetVotesBySeasonAndVoter(_ context.Context, arg db.GetVotesBySeasonAndVoterParams) ([]db.Vote, error) {
	key := arg.SeasonID + ":" + arg.VoterID
	return m.votes[key], nil
}

func (m *mockQuerier) GetGroupMembers(_ context.Context, groupID string) ([]db.GetGroupMembersRow, error) {
	return m.memberRows[groupID], nil
}

func (m *mockQuerier) HasVoteForQuestion(_ context.Context, arg db.HasVoteForQuestionParams) (int64, error) {
	key := arg.SeasonID + ":" + arg.VoterID + ":" + arg.QuestionID
	return m.voteForQ[key], nil
}

func (m *mockQuerier) CreateVote(_ context.Context, arg db.CreateVoteParams) (db.Vote, error) {
	m.createdVotes = append(m.createdVotes, arg)
	// Add to votes map so GetVotesBySeasonAndVoter returns updated count
	key := arg.SeasonID + ":" + arg.VoterID
	m.votes[key] = append(m.votes[key], db.Vote{
		ID:         arg.ID,
		SeasonID:   arg.SeasonID,
		VoterID:    arg.VoterID,
		TargetID:   arg.TargetID,
		QuestionID: arg.QuestionID,
		CreatedAt:  time.Now(),
	})
	return db.Vote{}, nil
}

func (m *mockQuerier) CountGroupMembers(_ context.Context, groupID string) (int64, error) {
	return m.memberCounts[groupID], nil
}

func (m *mockQuerier) CountSeasonQuestions(_ context.Context, _ string) (int64, error) {
	return m.questionCount, nil
}

func (m *mockQuerier) CountCompletedVoters(_ context.Context, _ db.CountCompletedVotersParams) (int64, error) {
	return m.completedCount, nil
}

func (m *mockQuerier) HasUserVotedInSeason(_ context.Context, _ db.HasUserVotedInSeasonParams) (int64, error) {
	return m.userVoteCount, nil
}

// --- Test fixtures ---

func newMock() *mockQuerier {
	return &mockQuerier{
		seasons: map[string]db.Season{
			"s1": {ID: "s1", GroupID: "g1", Status: db.SeasonStatusVOTING},
		},
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true, "u3": true},
		},
		memberRows: map[string][]db.GetGroupMembersRow{
			"g1": {
				{ID: "u1", Username: "user1"},
				{ID: "u2", Username: "user2"},
				{ID: "u3", Username: "user3"},
			},
		},
		questions: map[string][]db.Question{
			"s1": {
				{ID: "q1", Text: "Question 1", Category: db.QuestionCategoryFUNNY},
				{ID: "q2", Text: "Question 2", Category: db.QuestionCategoryHOT},
			},
		},
		votes:        map[string][]db.Vote{},
		voteForQ:     map[string]int64{},
		createdVotes: nil,
		memberCounts: map[string]int64{"g1": 3},
		questionCount: 2,
	}
}

// --- GetVotingSession tests ---

func TestGetVotingSession_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	session, err := svc.GetVotingSession(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if session.SeasonID != "s1" {
		t.Errorf("expected season s1, got %s", session.SeasonID)
	}
	if len(session.Questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(session.Questions))
	}
	if len(session.Targets) != 2 {
		t.Errorf("expected 2 targets (excluding self), got %d", len(session.Targets))
	}
	if session.Answered != 0 || session.Total != 2 {
		t.Errorf("expected progress 0/2, got %d/%d", session.Answered, session.Total)
	}
}

func TestGetVotingSession_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.GetVotingSession(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestGetVotingSession_SeasonNotVoting(t *testing.T) {
	m := newMock()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1", Status: db.SeasonStatusREVEALED}
	svc := NewService(m)

	_, err := svc.GetVotingSession(context.Background(), "s1", "u1")
	if !errors.Is(err, ErrSeasonNotVoting) {
		t.Errorf("expected ErrSeasonNotVoting, got %v", err)
	}
}

func TestGetVotingSession_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.GetVotingSession(context.Background(), "s1", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetVotingSession_PartialProgress(t *testing.T) {
	m := newMock()
	m.votes["s1:u1"] = []db.Vote{
		{SeasonID: "s1", VoterID: "u1", QuestionID: "q1"},
	}
	svc := NewService(m)

	session, err := svc.GetVotingSession(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if session.Answered != 1 || session.Total != 2 {
		t.Errorf("expected progress 1/2, got %d/%d", session.Answered, session.Total)
	}
	if !session.Questions[0].Answered {
		t.Error("expected q1 to be marked as answered")
	}
	if session.Questions[1].Answered {
		t.Error("expected q2 to not be answered")
	}
}

// --- CastVote tests ---

func TestCastVote_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	result, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if err != nil {
		t.Fatal(err)
	}

	if result.QuestionID != "q1" || result.TargetID != "u2" {
		t.Errorf("unexpected result: %+v", result)
	}
	if result.Answered != 1 || result.Total != 2 {
		t.Errorf("expected progress 1/2, got %d/%d", result.Answered, result.Total)
	}
	if len(m.createdVotes) != 1 {
		t.Errorf("expected 1 created vote, got %d", len(m.createdVotes))
	}
}

func TestCastVote_SelfVote(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u1")
	if !errors.Is(err, ErrSelfVote) {
		t.Errorf("expected ErrSelfVote, got %v", err)
	}
}

func TestCastVote_AlreadyVoted(t *testing.T) {
	m := newMock()
	m.voteForQ["s1:u1:q1"] = 1
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if !errors.Is(err, ErrAlreadyVoted) {
		t.Errorf("expected ErrAlreadyVoted, got %v", err)
	}
}

func TestCastVote_SeasonNotVoting(t *testing.T) {
	m := newMock()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1", Status: db.SeasonStatusCLOSED}
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if !errors.Is(err, ErrSeasonNotVoting) {
		t.Errorf("expected ErrSeasonNotVoting, got %v", err)
	}
}

func TestCastVote_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "outsider", "q1", "u2")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestCastVote_TargetNotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "outsider")
	if !errors.Is(err, ErrTargetNotMember) {
		t.Errorf("expected ErrTargetNotMember, got %v", err)
	}
}

func TestCastVote_InvalidQuestion(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "nonexistent", "u2")
	if !errors.Is(err, ErrInvalidQuestion) {
		t.Errorf("expected ErrInvalidQuestion, got %v", err)
	}
}

func TestCastVote_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "nonexistent", "u1", "q1", "u2")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestCastVote_FullSession(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	// Vote on q1
	r1, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if r1.Answered != 1 || r1.Total != 2 {
		t.Errorf("expected 1/2 after first vote, got %d/%d", r1.Answered, r1.Total)
	}

	// Vote on q2
	r2, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3")
	if err != nil {
		t.Fatal(err)
	}
	if r2.Answered != 2 || r2.Total != 2 {
		t.Errorf("expected 2/2 after second vote, got %d/%d", r2.Answered, r2.Total)
	}
}

// --- GetProgress tests ---

func TestGetProgress_Success(t *testing.T) {
	m := newMock()
	m.completedCount = 1
	m.userVoteCount = 2
	svc := NewService(m)

	p, err := svc.GetProgress(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if p.VotedCount != 1 {
		t.Errorf("expected voted_count 1, got %d", p.VotedCount)
	}
	if p.TotalCount != 3 {
		t.Errorf("expected total_count 3, got %d", p.TotalCount)
	}
	if p.QuorumThreshold != 0.4 {
		t.Errorf("expected threshold 0.4 for <8 members, got %f", p.QuorumThreshold)
	}
	if !p.UserVoted {
		t.Error("expected user_voted true (2 votes == 2 questions)")
	}
}

func TestGetProgress_QuorumLargeGroup(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 10
	m.completedCount = 5
	m.userVoteCount = 0
	svc := NewService(m)

	p, err := svc.GetProgress(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if p.QuorumThreshold != 0.5 {
		t.Errorf("expected threshold 0.5 for >=8 members, got %f", p.QuorumThreshold)
	}
	if !p.QuorumReached {
		t.Error("expected quorum reached (5/10 >= 50%)")
	}
	if p.UserVoted {
		t.Error("expected user_voted false")
	}
}

func TestGetProgress_QuorumNotReached(t *testing.T) {
	m := newMock()
	m.memberCounts["g1"] = 10
	m.completedCount = 4
	svc := NewService(m)

	p, err := svc.GetProgress(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	if p.QuorumReached {
		t.Error("expected quorum not reached (4/10 < 50%)")
	}
}

func TestGetProgress_NotMember(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.GetProgress(context.Background(), "s1", "outsider")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestGetProgress_SeasonNotFound(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	_, err := svc.GetProgress(context.Background(), "nonexistent", "u1")
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}
