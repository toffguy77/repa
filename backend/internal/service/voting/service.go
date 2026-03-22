package voting

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
)

var (
	ErrSeasonNotFound    = errors.New("season not found")
	ErrSeasonNotVoting   = errors.New("season is not in VOTING status")
	ErrNotMember         = errors.New("user is not a member of this group")
	ErrAlreadyVoted      = errors.New("already voted for this question")
	ErrSelfVote          = errors.New("cannot vote for yourself")
	ErrTargetNotMember   = errors.New("target is not a member of this group")
	ErrInvalidQuestion   = errors.New("question is not part of this season")
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
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

	members, err := s.queries.GetGroupMembers(ctx, season.GroupID)
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0, len(members)-1)
	for _, m := range members {
		if m.ID == userID {
			continue
		}
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

	return &CastVoteResult{
		QuestionID: questionID,
		TargetID:   targetID,
		Answered:   len(votes),
		Total:      len(seasonQuestions),
	}, nil
}

type VotingProgress struct {
	VotedCount     int64
	TotalCount     int64
	QuorumReached  bool
	QuorumThreshold float64
	UserVoted      bool
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
