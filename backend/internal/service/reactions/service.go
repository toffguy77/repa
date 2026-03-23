package reactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var (
	ErrSeasonNotFound    = errors.New("season not found")
	ErrSeasonNotRevealed = errors.New("season is not revealed")
	ErrNotMember         = errors.New("user is not a member of this group")
	ErrSelfReaction      = errors.New("cannot react to your own card")
	ErrInvalidEmoji      = errors.New("invalid emoji")
)

var allowedEmojis = map[string]bool{
	"\U0001F602": true, // 😂
	"\U0001F525": true, // 🔥
	"\U0001F480": true, // 💀
	"\U0001F440": true, // 👀
	"\U0001FAE1": true, // 🫡
}

type ReactionCounts struct {
	Counts   map[string]int `json:"counts"`
	MyEmoji  *string        `json:"my_emoji"`
}

type Service struct {
	queries     *db.Queries
	asynqClient *asynq.Client
}

func NewService(queries *db.Queries, asynqClient *asynq.Client) *Service {
	return &Service{queries: queries, asynqClient: asynqClient}
}

func (s *Service) CreateReaction(ctx context.Context, seasonID, reactorID, targetID, emoji string) (*ReactionCounts, error) {
	if !allowedEmojis[emoji] {
		return nil, ErrInvalidEmoji
	}

	if reactorID == targetID {
		return nil, ErrSelfReaction
	}

	season, err := s.validateAccess(ctx, seasonID, reactorID)
	if err != nil {
		return nil, err
	}

	_, err = s.queries.CreateReaction(ctx, db.CreateReactionParams{
		ID:        uuid.New().String(),
		SeasonID:  seasonID,
		ReactorID: reactorID,
		TargetID:  targetID,
		Emoji:     emoji,
	})
	if err != nil {
		return nil, fmt.Errorf("create reaction: %w", err)
	}

	// Enqueue push notification
	if s.asynqClient != nil {
		payload, _ := json.Marshal(map[string]string{
			"target_id":  targetID,
			"reactor_id": reactorID,
			"emoji":      emoji,
			"group_id":   season.GroupID,
			"season_id":  seasonID,
		})
		task := asynq.NewTask(lib.TypeReactionPush, payload)
		if _, err := s.asynqClient.Enqueue(task); err != nil {
			log.Warn().Err(err).Str("target_id", targetID).Msg("failed to enqueue reaction push")
		}
	}

	return s.getReactionCounts(ctx, seasonID, targetID, reactorID)
}

func (s *Service) GetReactions(ctx context.Context, seasonID, targetID, currentUserID string) (*ReactionCounts, error) {
	if _, err := s.validateAccess(ctx, seasonID, currentUserID); err != nil {
		return nil, err
	}
	return s.getReactionCounts(ctx, seasonID, targetID, currentUserID)
}

func (s *Service) validateAccess(ctx context.Context, seasonID, userID string) (db.Season, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return season, ErrSeasonNotFound
	}
	if season.Status != db.SeasonStatusREVEALED {
		return season, ErrSeasonNotRevealed
	}
	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return season, err
	}
	if isMember == 0 {
		return season, ErrNotMember
	}
	return season, nil
}

func (s *Service) getReactionCounts(ctx context.Context, seasonID, targetID, currentUserID string) (*ReactionCounts, error) {
	rows, err := s.queries.GetReactionsForUser(ctx, db.GetReactionsForUserParams{
		SeasonID: seasonID,
		TargetID: targetID,
	})
	if err != nil {
		return nil, fmt.Errorf("get reactions: %w", err)
	}

	counts := make(map[string]int)
	var myEmoji *string
	for _, r := range rows {
		counts[r.Emoji]++
		if r.ReactorID == currentUserID {
			e := r.Emoji
			myEmoji = &e
		}
	}

	return &ReactionCounts{
		Counts:  counts,
		MyEmoji: myEmoji,
	}, nil
}
