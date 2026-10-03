package voting

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
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
	memberships    map[string]db.GroupMember
	membershipErr  error
	blocked        map[string]bool // "a:b" -> a blocked b
}

func (m *mockQuerier) GetVotingTargets(_ context.Context, arg db.GetVotingTargetsParams) ([]db.GetVotingTargetsRow, error) {
	var out []db.GetVotingTargetsRow
	for _, r := range m.memberRows[arg.GroupID] {
		if r.ID == arg.ID || m.blocked[arg.ID+":"+r.ID] || m.blocked[r.ID+":"+arg.ID] {
			continue
		}
		out = append(out, db.GetVotingTargetsRow{
			ID:          r.ID,
			Username:    r.Username,
			AvatarEmoji: r.AvatarEmoji,
			AvatarUrl:   r.AvatarUrl,
		})
	}
	return out, nil
}

func (m *mockQuerier) IsBlockedEitherWay(_ context.Context, arg db.IsBlockedEitherWayParams) (bool, error) {
	return m.blocked[arg.BlockerID+":"+arg.BlockedID] || m.blocked[arg.BlockedID+":"+arg.BlockerID], nil
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
		votes:         map[string][]db.Vote{},
		voteForQ:      map[string]int64{},
		createdVotes:  nil,
		memberCounts:  map[string]int64{"g1": 3},
		blocked:       map[string]bool{},
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

// --- Kickoff reveal scheduling on session completion ---

type fakeKickoffScheduler struct {
	calls     []string
	scheduled bool
	err       error
}

func (f *fakeKickoffScheduler) ScheduleKickoffIfEligible(_ context.Context, seasonID string) (bool, error) {
	f.calls = append(f.calls, seasonID)
	return f.scheduled, f.err
}

type fakeEnqueuer struct {
	tasks []*asynq.Task
}

func (f *fakeEnqueuer) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	f.tasks = append(f.tasks, task)
	return &asynq.TaskInfo{}, nil
}

// ofType filters the recorded tasks, so a test about one kind of push is not perturbed by another.
func (f *fakeEnqueuer) ofType(taskType string) []*asynq.Task {
	var out []*asynq.Task
	for _, t := range f.tasks {
		if t.Type() == taskType {
			out = append(out, t)
		}
	}
	return out
}

// kickoffSetup returns a service whose season is a kickoff with 2 questions, so the
// second vote completes the session.
func kickoffSetup(t *testing.T, kind db.SeasonKind, scheduled bool) (*Service, *fakeKickoffScheduler, *fakeEnqueuer) {
	t.Helper()
	m := newMock()
	s := m.seasons["s1"]
	s.Kind = kind
	m.seasons["s1"] = s

	sched := &fakeKickoffScheduler{scheduled: scheduled}
	enq := &fakeEnqueuer{}
	return NewService(m).WithKickoff(sched, enq), sched, enq
}

func TestCastVote_CompletingSessionSchedulesKickoffReveal(t *testing.T) {
	svc, sched, enq := kickoffSetup(t, db.SeasonKindKICKOFF, true)

	// First vote: session not yet complete (2 questions).
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if len(sched.calls) != 0 {
		t.Errorf("scheduling must only be attempted on a completed session, got %v", sched.calls)
	}

	// Second vote completes the session.
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatal(err)
	}
	if len(sched.calls) != 1 || sched.calls[0] != "s1" {
		t.Errorf("expected one scheduling attempt for s1, got %v", sched.calls)
	}
	kickoffPushes := enq.ofType(lib.TypePushKickoff)
	if len(kickoffPushes) != 1 {
		t.Fatalf("expected exactly one kickoff push, got %d", len(kickoffPushes))
	}
}

func TestCastVote_NoPushWhenAlreadyScheduled(t *testing.T) {
	svc, sched, enq := kickoffSetup(t, db.SeasonKindKICKOFF, false)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatal(err)
	}

	if len(sched.calls) != 1 {
		t.Errorf("expected one scheduling attempt, got %v", sched.calls)
	}
	if got := len(enq.ofType(lib.TypePushKickoff)); got != 0 {
		t.Errorf("a voter who did not trigger scheduling must not announce it, got %d pushes", got)
	}
}

func TestCastVote_WeeklySeasonNeverSchedulesKickoff(t *testing.T) {
	svc, sched, enq := kickoffSetup(t, db.SeasonKindWEEKLY, true)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatal(err)
	}

	if len(sched.calls) != 0 {
		t.Errorf("weekly seasons keep their Friday reveal, got %v", sched.calls)
	}
	if got := len(enq.ofType(lib.TypePushKickoff)); got != 0 {
		t.Errorf("expected no kickoff push, got %d", got)
	}
}

func TestCastVote_SchedulerErrorDoesNotFailTheVote(t *testing.T) {
	m := newMock()
	s := m.seasons["s1"]
	s.Kind = db.SeasonKindKICKOFF
	m.seasons["s1"] = s
	sched := &fakeKickoffScheduler{err: errors.New("db down")}
	enq := &fakeEnqueuer{}
	svc := NewService(m).WithKickoff(sched, enq)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	result, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3")
	if err != nil {
		t.Fatalf("a scheduling failure must not roll back the recorded vote: %v", err)
	}
	if result.Answered != 2 {
		t.Errorf("Answered = %d, want 2", result.Answered)
	}
	if got := len(enq.ofType(lib.TypePushKickoff)); got != 0 {
		t.Errorf("expected no kickoff push after a scheduling error, got %d", got)
	}
}

func TestCastVote_NilKickoffCollaboratorsAreSafe(t *testing.T) {
	m := newMock()
	s := m.seasons["s1"]
	s.Kind = db.SeasonKindKICKOFF
	m.seasons["s1"] = s
	svc := NewService(m) // no WithKickoff

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatalf("voting must work without the kickoff collaborators: %v", err)
	}
}

// --- Referral payout on first completed session ---

type fakeReferralGranter struct {
	calls [][2]string // {inviterID, inviteeID}
}

func (f *fakeReferralGranter) GrantReferral(_ context.Context, inviterID, inviteeID string) {
	f.calls = append(f.calls, [2]string{inviterID, inviteeID})
}

func (m *mockQuerier) GetMembershipByUserAndGroup(_ context.Context, arg db.GetMembershipByUserAndGroupParams) (db.GroupMember, error) {
	if m.membershipErr != nil {
		return db.GroupMember{}, m.membershipErr
	}
	gm, ok := m.memberships[arg.UserID]
	if !ok {
		return db.GroupMember{}, sql.ErrNoRows
	}
	return gm, nil
}

// referralSetup returns a service whose voter u1 was invited by [invitedBy] (empty for none).
func referralSetup(t *testing.T, invitedBy string) (*Service, *fakeReferralGranter) {
	t.Helper()
	m := newMock()
	m.memberships = map[string]db.GroupMember{
		"u1": {
			ID:        "mem-1",
			UserID:    "u1",
			GroupID:   "g1",
			InvitedBy: sql.NullString{String: invitedBy, Valid: invitedBy != ""},
		},
	}
	granter := &fakeReferralGranter{}
	return NewService(m).WithReferralGrants(granter), granter
}

func TestCastVote_PaysTheInviterOnFirstCompletedSession(t *testing.T) {
	svc, granter := referralSetup(t, "inviter")

	// First vote: the session is not complete, so nothing is payable.
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if len(granter.calls) != 0 {
		t.Errorf("a partial session must not pay: %v", granter.calls)
	}

	// Second vote completes it.
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatal(err)
	}
	if len(granter.calls) != 1 {
		t.Fatalf("expected one payout, got %v", granter.calls)
	}
	if granter.calls[0] != [2]string{"inviter", "u1"} {
		t.Errorf("payout = %v, want {inviter u1}", granter.calls[0])
	}
}

func TestCastVote_PaysNobodyWhenThereIsNoReferrer(t *testing.T) {
	svc, granter := referralSetup(t, "")

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatal(err)
	}

	if len(granter.calls) != 0 {
		t.Errorf("an unattributed join pays nobody, got %v", granter.calls)
	}
}

func TestCastVote_MembershipReadFailureDoesNotFailTheVote(t *testing.T) {
	m := newMock()
	m.membershipErr = errors.New("db down")
	granter := &fakeReferralGranter{}
	svc := NewService(m).WithReferralGrants(granter)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	result, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3")
	if err != nil {
		t.Fatalf("a payout problem must not roll back a recorded vote: %v", err)
	}
	if result.Answered != 2 {
		t.Errorf("Answered = %d, want 2", result.Answered)
	}
	if len(granter.calls) != 0 {
		t.Errorf("expected no payout after a read failure, got %v", granter.calls)
	}
}

func TestCastVote_NoReferralGranterIsSafe(t *testing.T) {
	m := newMock()
	svc := NewService(m) // no WithReferralGrants

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q2", "u3"); err != nil {
		t.Fatalf("voting must work without the grants collaborator: %v", err)
	}
}

// --- The immediate, non-attributed signal ---

func TestCastVote_SignalsTheTargetNotTheVoter(t *testing.T) {
	m := newMock()
	enq := &fakeEnqueuer{}
	svc := NewService(m).WithKickoff(nil, enq)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}

	signals := enq.ofType(lib.TypePushVoteSignal)
	if len(signals) != 1 {
		t.Fatalf("expected one signal, got %d", len(signals))
	}

	var payload map[string]string
	if err := json.Unmarshal(signals[0].Payload(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["target_id"] != "u2" {
		t.Errorf("target_id = %q, want u2 — the person voted about", payload["target_id"])
	}
	if payload["target_id"] == "u1" {
		t.Error("answering about others must never notify the voter about themselves")
	}
	// The payload carries no answer: the *what* stays sealed until Friday.
	if _, ok := payload["question_id"]; ok {
		t.Error("the signal must not carry the question")
	}
}

func TestCastVote_SignalFailureDoesNotFailTheVote(t *testing.T) {
	m := newMock()
	svc := NewService(m).WithKickoff(nil, &failingEnqueuer{})

	result, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if err != nil {
		t.Fatalf("a notification must never cost a recorded vote: %v", err)
	}
	if result.Answered != 1 {
		t.Errorf("Answered = %d, want 1", result.Answered)
	}
	if len(m.createdVotes) != 1 {
		t.Error("the vote should still be recorded")
	}
}

func TestCastVote_NoEnqueuerIsSafe(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatalf("voting must work without the push collaborator: %v", err)
	}
}

// failingEnqueuer stands in for a broken queue.
type failingEnqueuer struct{}

func (failingEnqueuer) Enqueue(*asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error) {
	return nil, errors.New("queue down")
}

// --- Blocks and voting ---

func TestGetVotingSession_ExcludesBlockedMembersBothWays(t *testing.T) {
	m := newMock()
	// u1 blocked u2; u3 blocked u1. Neither should be offered to u1.
	m.blocked["u1:u2"] = true
	m.blocked["u3:u1"] = true
	svc := NewService(m)

	session, err := svc.GetVotingSession(context.Background(), "s1", "u1")
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range session.Targets {
		if target.UserID == "u2" {
			t.Error("a member I blocked must not be offered as a target")
		}
		if target.UserID == "u3" {
			t.Error("a member who blocked me must not be offered either")
		}
	}
}

func TestCastVote_RefusesABlockedTarget(t *testing.T) {
	m := newMock()
	m.blocked["u1:u2"] = true
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if !errors.Is(err, ErrTargetBlocked) {
		t.Errorf("expected ErrTargetBlocked, got %v", err)
	}
	if len(m.createdVotes) != 0 {
		t.Error("no vote should be recorded between blocked members")
	}
}

func TestCastVote_RefusesWhenTheOtherPersonBlockedMe(t *testing.T) {
	m := newMock()
	// Only u2 created the block, but the effect is symmetric.
	m.blocked["u2:u1"] = true
	svc := NewService(m)

	_, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2")
	if !errors.Is(err, ErrTargetBlocked) {
		t.Errorf("expected ErrTargetBlocked, got %v", err)
	}
}

func TestCastVote_VotesCastBeforeABlockAreKept(t *testing.T) {
	m := newMock()
	svc := NewService(m)

	// Vote first, then block.
	if _, err := svc.CastVote(context.Background(), "s1", "u1", "q1", "u2"); err != nil {
		t.Fatal(err)
	}
	m.blocked["u1:u2"] = true

	// Removing the earlier vote would shift u2's card at the moment of the block — a worse result,
	// and a signal that a block happened.
	if len(m.createdVotes) != 1 {
		t.Errorf("the earlier vote must survive the block, got %d votes", len(m.createdVotes))
	}
}
