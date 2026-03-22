package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	achievesvc "github.com/repa-app/repa/internal/service/achievements"
	"github.com/rs/zerolog/log"
)

type AchievementsPayload struct {
	SeasonID string `json:"season_id"`
}

type AchievementsProcessor struct {
	svc *achievesvc.Service
}

func NewAchievementsProcessor(svc *achievesvc.Service) *AchievementsProcessor {
	return &AchievementsProcessor{svc: svc}
}

func (h *AchievementsProcessor) HandleAchievements(ctx context.Context, t *asynq.Task) error {
	var p AchievementsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal achievements payload: %w", err)
	}

	log.Info().Str("season_id", p.SeasonID).Msg("calculating achievements")

	if err := h.svc.CalculateAchievements(ctx, p.SeasonID); err != nil {
		return fmt.Errorf("calculate achievements for season %s: %w", p.SeasonID, err)
	}

	log.Info().Str("season_id", p.SeasonID).Msg("achievements calculated successfully")
	return nil
}
