package voting

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var (
	ErrSeasonNotFound  = errors.New("season not found")
	ErrSeasonNotVoting = errors.New("season is not in VOTING status")
	ErrNotMember       = errors.New("user is not a member of this group")
	ErrAlreadyVoted    = errors.New("already voted for this question")
	ErrSelfVote        = errors.New("cannot vote for yourself")
	ErrTargetNotMember = errors.New("target is not a member of this group")
	ErrInvalidQuestion = errors.New("question is not part of this season")
	// ErrTargetBlocked means a block exists in one direction or the other. Votes cast *before* a block
	// are kept: removing them would shift a member's card at the moment of a block, which is both a
	// worse result and a signal that a block happened.
	ErrTargetBlocked = errors.New("cannot vote for a blocked member")
)

// TaskEnqueuer is the slice of *asynq.Client this service uses (same pattern as the
// reactions service).
type TaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// ReferralGranter pays the member who invited someone, once that someone has actually played.
// Implemented by the crystals service.
type ReferralGranter interface {
	GrantReferral(ctx context.Context, inviterID, inviteeID string)
}

// KickoffScheduler brings a kickoff season's Reveal forward once the group first becomes
// reveal-eligible. Implemented by the reveal service, which owns the eligibility
// predicate — duplicating it here would let the app's explanation drift from the
// worker's behaviour.
type KickoffScheduler interface {
	ScheduleKickoffIfEligible(ctx context.Context, seasonID string) (bool, error)
}

type Service struct {
	queries     db.Querier
	kickoff     KickoffScheduler
	asynqClient TaskEnqueuer
	referrals   ReferralGranter
}

// WithReferralGrants wires the referral payout. Optional: nil means no payout, which is what
// unit tests that only exercise vote validation want.
func (s *Service) WithReferralGrants(g ReferralGranter) *Service {
	s.referrals = g
	return s
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

// WithKickoff wires the optional kickoff-scheduling collaborators. Both are nil in unit
// tests that only exercise vote validation.
func (s *Service) WithKickoff(kickoff KickoffScheduler, asynqClient TaskEnqueuer) *Service {
	s.kickoff = kickoff
	s.asynqClient = asynqClient
	return s
}

type VotingQuestion struct {
	SeasonQuestionID string
	QuestionID       string
	Text             string
	Category         string
	Answered         bool
}

type Target struct {
	UserID      string
	Username    string
	AvatarEmoji sql.NullString
	AvatarURL   sql.NullString
}

type VotingSession struct {
	SeasonID  string
	Questions []VotingQuestion
	Targets   []Target
	Answered  int
	Total     int
}

func (s *Service) GetVotingSession(ctx context.Context, seasonID, userID string) (*VotingSession, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeasonNotFound
		}
		return nil, err
	}

	if season.Status != db.SeasonStatusVOTING {
		return nil, ErrSeasonNotVoting
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	questions, err := s.queries.GetSeasonQuestions(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	existingVotes, err := s.queries.GetVotesBySeasonAndVoter(ctx, db.GetVotesBySeasonAndVoterParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil {
		return nil, err
	}

	votedQuestions := make(map[string]bool, len(existingVotes))
	for _, v := range existingVotes {
		votedQuestions[v.QuestionID] = true
	}

	votingQuestions := make([]VotingQuestion, len(questions))
	for i, q := range questions {
		votingQuestions[i] = VotingQuestion{
			QuestionID: q.ID,
			Text:       q.Text,
			Category:   string(q.Category),
			Answered:   votedQuestions[q.ID],
		}
	}

	// Targets come from a query that already excludes the voter and anyone blocked in either
	// direction, so the session's list and the vote endpoint cannot disagree about who is rateable —
	// the bug that shape prevents is offering a target the vote then refuses.
	members, err := s.queries.GetVotingTargets(ctx, db.GetVotingTargetsParams{
		GroupID: season.GroupID,
		ID:      userID,
	})
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0, len(members))
	for _, m := range members {
		targets = append(targets, Target{
			UserID:      m.ID,
			Username:    m.Username,
			AvatarEmoji: m.AvatarEmoji,
			AvatarURL:   m.AvatarUrl,
		})
	}

	return &VotingSession{
		SeasonID:  seasonID,
		Questions: votingQuestions,
		Targets:   targets,
		Answered:  len(existingVotes),
		Total:     len(questions),
	}, nil
}

type CastVoteResult struct {
	QuestionID string
	TargetID   string
	Answered   int
	Total      int
}

func (s *Service) CastVote(ctx context.Context, seasonID, voterID, questionID, targetID string) (*CastVoteResult, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeasonNotFound
		}
		return nil, err
	}

	if season.Status != db.SeasonStatusVOTING {
		return nil, ErrSeasonNotVoting
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  voterID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	if voterID == targetID {
		return nil, ErrSelfVote
	}

	// The product must never record a rating between two people who cut each other off.
	blocked, err := s.queries.IsBlockedEitherWay(ctx, db.IsBlockedEitherWayParams{
		BlockerID: voterID,
		BlockedID: targetID,
	})
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrTargetBlocked
	}

	isTargetMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  targetID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isTargetMember == 0 {
		return nil, ErrTargetNotMember
	}

	// Verify question belongs to this season
	seasonQuestions, err := s.queries.GetSeasonQuestions(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	questionInSeason := false
	for _, q := range seasonQuestions {
		if q.ID == questionID {
			questionInSeason = true
			break
		}
	}
	if !questionInSeason {
		return nil, ErrInvalidQuestion
	}

	// Check if already voted for this question
	hasVote, err := s.queries.HasVoteForQuestion(ctx, db.HasVoteForQuestionParams{
		SeasonID:   seasonID,
		VoterID:    voterID,
		QuestionID: questionID,
	})
	if err != nil {
		return nil, err
	}
	if hasVote > 0 {
		return nil, ErrAlreadyVoted
	}

	_, err = s.queries.CreateVote(ctx, db.CreateVoteParams{
		ID:         uuid.New().String(),
		SeasonID:   seasonID,
		VoterID:    voterID,
		TargetID:   targetID,
		QuestionID: questionID,
	})
	if err != nil {
		return nil, err
	}

	// Get updated progress
	votes, err := s.queries.GetVotesBySeasonAndVoter(ctx, db.GetVotesBySeasonAndVoterParams{
		SeasonID: seasonID,
		VoterID:  voterID,
	})
	if err != nil {
		return nil, err
	}

	// Tell the target that the number of people who answered about them went up. Non-attributed by
	// design: the *what* stays sealed until Friday. Best-effort — a notification must never cost a
	// recorded vote.
	s.signalVoteToTarget(ctx, season, targetID)

	// A completed session can be the moment a new group first becomes reveal-eligible, and it
	// is also the moment a referral becomes payable.
	if len(votes) == len(seasonQuestions) {
		s.maybeScheduleKickoffReveal(ctx, season)
		s.maybePayReferral(ctx, season.GroupID, voterID)
	}

	return &CastVoteResult{
		QuestionID: questionID,
		TargetID:   targetID,
		Answered:   len(votes),
		Total:      len(seasonQuestions),
	}, nil
}

// maybeScheduleKickoffReveal brings a kickoff Reveal forward when this completed session
// made the group eligible for the first time. Failures are logged, never surfaced: the
// vote itself has already been recorded and must not be rolled back by a scheduling hiccup.
func (s *Service) maybeScheduleKickoffReveal(ctx context.Context, season db.Season) {
	if s.kickoff == nil || season.Kind != db.SeasonKindKICKOFF {
		return
	}

	scheduled, err := s.kickoff.ScheduleKickoffIfEligible(ctx, season.ID)
	if err != nil {
		log.Error().Err(err).Str("season_id", season.ID).Msg("failed to schedule kickoff reveal")
		return
	}
	if !scheduled || s.asynqClient == nil {
		return
	}

	payload, _ := json.Marshal(map[string]string{"season_id": season.ID})
	task := asynq.NewTask(lib.TypePushKickoff, payload)
	// TaskID makes the announcement idempotent: a concurrent completion that also saw
	// "scheduled" cannot produce a second push.
	if _, err := s.asynqClient.Enqueue(task, asynq.TaskID("kickoff-push:"+season.ID)); err != nil {
		log.Error().Err(err).Str("season_id", season.ID).Msg("failed to enqueue kickoff reveal push")
	}
}

type VotingProgress struct {
	VotedCount      int64
	TotalCount      int64
	QuorumReached   bool
	QuorumThreshold float64
	UserVoted       bool
}

func (s *Service) GetProgress(ctx context.Context, seasonID, userID string) (*VotingProgress, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeasonNotFound
		}
		return nil, err
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, season.GroupID)
	if err != nil {
		return nil, err
	}

	totalQuestions, err := s.queries.CountSeasonQuestions(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	// A voter has "completed" voting if they voted on all questions
	completedVoters, err := s.queries.CountCompletedVoters(ctx, db.CountCompletedVotersParams{
		SeasonID: seasonID,
		Column2:  totalQuestions,
	})
	if err != nil {
		return nil, err
	}

	// Check if current user voted (has any votes)
	userVoteCount, err := s.queries.HasUserVotedInSeason(ctx, db.HasUserVotedInSeasonParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil {
		return nil, err
	}

	threshold := 0.5
	if memberCount < 8 {
		threshold = 0.4
	}

	quorumReached := float64(completedVoters) >= float64(memberCount)*threshold

	return &VotingProgress{
		VotedCount:      completedVoters,
		TotalCount:      memberCount,
		QuorumReached:   quorumReached,
		QuorumThreshold: threshold,
		UserVoted:       userVoteCount >= totalQuestions,
	}, nil
}

// maybePayReferral pays the member who brought this voter into the group.
//
// The trigger is a *completed* session rather than the join: rewarding a join would pay for an
// account, and accounts are free. A completed session means the voter answered every question
// about real people in a real group, which is not worth faking for a handful of crystals.
//
// The grant itself is keyed on the invitee, so it pays once however many sessions they complete.
// Failures are logged inside the grants service — a payout must never roll back a recorded vote.
func (s *Service) maybePayReferral(ctx context.Context, groupID, voterID string) {
	if s.referrals == nil {
		return
	}

	membership, err := s.queries.GetMembershipByUserAndGroup(ctx, db.GetMembershipByUserAndGroupParams{
		UserID:  voterID,
		GroupID: groupID,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", voterID).Msg("could not read membership for referral payout")
		return
	}
	if !membership.InvitedBy.Valid || membership.InvitedBy.String == "" {
		return // nobody to pay
	}

	s.referrals.GrantReferral(ctx, membership.InvitedBy.String, voterID)
}

// signalVoteToTarget enqueues the "someone answered about you" push for the person voted about.
//
// Never for the voter: answering about others must not notify you about yourself. The debounce to one
// per recipient per day lives in the push worker, where the Redis client is.
func (s *Service) signalVoteToTarget(ctx context.Context, season db.Season, targetID string) {
	if s.asynqClient == nil || targetID == "" {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"target_id": targetID,
		"season_id": season.ID,
		"group_id":  season.GroupID,
	})
	task := asynq.NewTask(lib.TypePushVoteSignal, payload)
	if _, err := s.asynqClient.Enqueue(task); err != nil {
		log.Error().Err(err).Str("target_id", targetID).Msg("failed to enqueue vote signal")
	}
}
