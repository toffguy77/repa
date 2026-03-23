package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	telegramsvc "github.com/repa-app/repa/internal/service/telegram"
	"github.com/rs/zerolog/log"
)

type TelegramPayload struct {
	GroupID  string `json:"group_id,omitempty"`
	SeasonID string `json:"season_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

type TelegramProcessor struct {
	svc *telegramsvc.Service
}

func NewTelegramProcessor(svc *telegramsvc.Service) *TelegramProcessor {
	return &TelegramProcessor{svc: svc}
}

func (p *TelegramProcessor) HandleSeasonStart(ctx context.Context, t *asynq.Task) error {
	// Cron mode (no payload): post to all groups with active voting seasons
	if t.Payload() == nil || len(t.Payload()) == 0 {
		return p.svc.PostSeasonStartAll(ctx)
	}

	var payload TelegramPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal telegram season-start payload: %w", err)
	}

	if err := p.svc.PostSeasonStart(ctx, payload.GroupID); err != nil {
		log.Warn().Err(err).Str("group_id", payload.GroupID).Msg("telegram season-start post failed")
	}

	return nil
}

func (p *TelegramProcessor) HandleRevealPost(ctx context.Context, t *asynq.Task) error {
	var payload TelegramPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal telegram reveal payload: %w", err)
	}

	if err := p.svc.PostReveal(ctx, payload.SeasonID); err != nil {
		log.Warn().Err(err).Str("season_id", payload.SeasonID).Msg("telegram reveal post failed")
	}

	return nil
}

func (p *TelegramProcessor) HandleShareCard(ctx context.Context, t *asynq.Task) error {
	var payload TelegramPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal telegram share-card payload: %w", err)
	}

	if err := p.svc.ShareCard(ctx, payload.UserID, payload.SeasonID); err != nil {
		log.Warn().Err(err).Str("user_id", payload.UserID).Str("season_id", payload.SeasonID).Msg("telegram share-card failed")
	}

	return nil
}
