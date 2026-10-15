package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
)

// fakeEnqueuer records the task types a reveal produces.
type fakeEnqueuer struct {
	enqueued []*asynq.Task
}

func (f *fakeEnqueuer) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	f.enqueued = append(f.enqueued, task)
	return &asynq.TaskInfo{}, nil
}

func (f *fakeEnqueuer) types() []string {
	out := make([]string, 0, len(f.enqueued))
	for _, t := range f.enqueued {
		out = append(out, t.Type())
	}
	return out
}

func (f *fakeEnqueuer) has(taskType string) bool {
	for _, t := range f.enqueued {
		if t.Type() == taskType {
			return true
		}
	}
	return false
}

func (f *fakeEnqueuer) payloadOf(taskType string) []byte {
	for _, t := range f.enqueued {
		if t.Type() == taskType {
			return t.Payload()
		}
	}
	return nil
}

// revealQuerier is the minimum surface ProcessReveal touches for these tests.
type revealQuerier struct {
	db.Querier
	season       db.Season
	memberCount  int64
	votedCount   int64
	questions    int64
	statuses     []db.UpdateSeasonStatusParams
	postponed    []db.PostponeSeasonParams
	createdCount int
}

func (q *revealQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	if q.season.ID != id {
		return db.Season{}, sql.ErrNoRows
	}
	return q.season, nil
}
func (q *revealQuerier) CountGroupMembers(_ context.Context, _ string) (int64, error) {
	return q.memberCount, nil
}
func (q *revealQuerier) CountSeasonQuestions(_ context.Context, _ string) (int64, error) {
	return q.questions, nil
}
func (q *revealQuerier) CountCompletedVoters(_ context.Context, _ db.CountCompletedVotersParams) (int64, error) {
	return q.votedCount, nil
}
func (q *revealQuerier) CountUniqueVoters(_ context.Context, _ string) (int64, error) {
	return q.votedCount, nil
}
func (q *revealQuerier) AggregateVotesByTarget(_ context.Context, _ string) ([]db.AggregateVotesByTargetRow, error) {
	return []db.AggregateVotesByTargetRow{{TargetID: "u1", QuestionID: "q1", VoteCount: 3}}, nil
}
func (q *revealQuerier) DeleteSeasonResultsBySeason(_ context.Context, _ string) error { return nil }
func (q *revealQuerier) CreateSeasonResult(_ context.Context, _ db.CreateSeasonResultParams) (db.SeasonResult, error) {
	q.createdCount++
	return db.SeasonResult{}, nil
}
func (q *revealQuerier) UpdateSeasonStatus(_ context.Context, arg db.UpdateSeasonStatusParams) error {
	q.statuses = append(q.statuses, arg)
	return nil
}
func (q *revealQuerier) PostponeSeason(_ context.Context, arg db.PostponeSeasonParams) (db.Season, error) {
	q.postponed = append(q.postponed, arg)
	return q.season, nil
}

func newRevealProcessor(t *testing.T, kind db.SeasonKind, members, voted int64) (*RevealProcessor, *fakeEnqueuer, *revealQuerier) {
	t.Helper()
	q := &revealQuerier{
		season:      db.Season{ID: "s1", GroupID: "g1", Status: db.SeasonStatusVOTING, Kind: kind},
		memberCount: members,
		votedCount:  voted,
		questions:   5,
	}
	enq := &fakeEnqueuer{}
	return NewRevealProcessor(revealsvc.NewService(q, nil), enq), enq, q
}

func processSeason(t *testing.T, p *RevealProcessor, attempt int) {
	t.Helper()
	payload, _ := json.Marshal(RevealPayload{SeasonID: "s1", Attempt: attempt})
	if err := p.HandleRevealProcess(context.Background(), asynq.NewTask("reveal:process", payload)); err != nil {
		t.Fatalf("HandleRevealProcess: %v", err)
	}
}

// --- 3.3: a kickoff reveal starts the group's weekly cycle immediately ---

func TestRevealProcess_KickoffReveal_EnqueuesFollowUpSeason(t *testing.T) {
	p, enq, _ := newRevealProcessor(t, db.SeasonKindKICKOFF, 5, 5)

	processSeason(t, p, 1)

	if !enq.has("season:create-for-group") {
		t.Fatalf("expected a follow-up weekly season task, got %v", enq.types())
	}
	var payload SeasonForGroupPayload
	if err := json.Unmarshal(enq.payloadOf("season:create-for-group"), &payload); err != nil {
		t.Fatalf("unmarshal follow-up payload: %v", err)
	}
	if payload.GroupID != "g1" {
		t.Errorf("follow-up GroupID = %q, want g1", payload.GroupID)
	}
}

func TestRevealProcess_WeeklyReveal_NoFollowUpSeason(t *testing.T) {
	p, enq, _ := newRevealProcessor(t, db.SeasonKindWEEKLY, 5, 5)

	processSeason(t, p, 1)

	if enq.has("season:create-for-group") {
		t.Errorf("a weekly reveal must leave Fri–Sun free for discussion, got %v", enq.types())
	}
	if !enq.has("cards:generate") {
		t.Errorf("expected card generation for a successful reveal, got %v", enq.types())
	}
}

// --- 3.4: a postponed season produces no cards, achievements or Telegram post ---

func TestRevealProcess_Postponed_EnqueuesNoDownstreamJobs(t *testing.T) {
	p, enq, q := newRevealProcessor(t, db.SeasonKindWEEKLY, 4, 2)

	processSeason(t, p, 3)

	for _, forbidden := range []string{"cards:generate", "achievements:calculate", "telegram:reveal-post", "push:reveal-notification"} {
		if enq.has(forbidden) {
			t.Errorf("postponed season must not enqueue %s, got %v", forbidden, enq.types())
		}
	}
	if q.createdCount != 0 {
		t.Errorf("postponed season must not aggregate results, got %d", q.createdCount)
	}
	if len(q.statuses) != 0 {
		t.Errorf("postponed season must stay VOTING, got %v", q.statuses)
	}
	if len(q.postponed) != 1 {
		t.Errorf("expected exactly one postponement, got %d", len(q.postponed))
	}
}

// --- 3.5: the postponement push carries the missing-voter count ---

func TestRevealProcess_Postponed_EnqueuesPushWithVotersNeeded(t *testing.T) {
	p, enq, _ := newRevealProcessor(t, db.SeasonKindWEEKLY, 4, 2)

	processSeason(t, p, 3)

	if !enq.has("push:reveal-postponed") {
		t.Fatalf("expected a postponement push, got %v", enq.types())
	}
	var payload PostponePushPayload
	if err := json.Unmarshal(enq.payloadOf("push:reveal-postponed"), &payload); err != nil {
		t.Fatalf("unmarshal postponement payload: %v", err)
	}
	if payload.SeasonID != "s1" {
		t.Errorf("SeasonID = %q, want s1", payload.SeasonID)
	}
	if payload.VotersNeeded != 1 {
		t.Errorf("VotersNeeded = %d, want 1", payload.VotersNeeded)
	}
}

func TestPostponedPushBody(t *testing.T) {
	tests := []struct {
		needed int64
		want   string
	}{
		{1, "Не хватает 1 человека — позови своих, и откроем репу."},
		{2, "Не хватает 2 человека — позови своих, и откроем репу."},
		{5, "Не хватает 5 человек — позови своих, и откроем репу."},
		{11, "Не хватает 11 человек — позови своих, и откроем репу."},
		{21, "Не хватает 21 человека — позови своих, и откроем репу."},
		{0, "Ждём ещё голосов — Reveal будет в следующую пятницу."},
	}
	for _, tt := range tests {
		if got := postponedPushBody(tt.needed); got != tt.want {
			t.Errorf("postponedPushBody(%d) = %q, want %q", tt.needed, got, tt.want)
		}
	}
}

// --- a season short of quorum but able to reach the floor keeps retrying ---

func TestRevealProcess_QuorumShortEarlyAttempt_RetriesWithoutDownstream(t *testing.T) {
	p, enq, q := newRevealProcessor(t, db.SeasonKindWEEKLY, 10, 3)

	processSeason(t, p, 1)

	if !enq.has("reveal:process") {
		t.Errorf("expected a retry task, got %v", enq.types())
	}
	if enq.has("cards:generate") || enq.has("push:reveal-postponed") {
		t.Errorf("a retry must neither reveal nor postpone, got %v", enq.types())
	}
	if len(q.postponed) != 0 {
		t.Errorf("expected no postponement on an early attempt, got %d", len(q.postponed))
	}
}
