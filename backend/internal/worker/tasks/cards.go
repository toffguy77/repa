package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	cardssvc "github.com/repa-app/repa/internal/service/cards"
	"github.com/rs/zerolog/log"
)

type CardsPayload struct {
	SeasonID string `json:"season_id"`
}

type CardsProcessor struct {
	svc *cardssvc.Service
}

func NewCardsProcessor(svc *cardssvc.Service) *CardsProcessor {
	return &CardsProcessor{svc: svc}
}

func (h *CardsProcessor) HandleCardsGenerate(ctx context.Context, t *asynq.Task) error {
	var p CardsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal cards payload: %w", err)
	}

	log.Info().Str("season_id", p.SeasonID).Msg("generating cards for season")

	if err := h.svc.GenerateCardsForSeason(ctx, p.SeasonID); err != nil {
		return fmt.Errorf("generate cards for season %s: %w", p.SeasonID, err)
	}

	log.Info().Str("season_id", p.SeasonID).Msg("cards generation complete")
	return nil
}
