package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/repa-app/repa/internal/lib"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
	"github.com/rs/zerolog/log"
)

type RevealPayload struct {
	SeasonID string `json:"season_id"`
	Attempt  int    `json:"attempt"`
}

type RevealChecker struct {
	svc    *revealsvc.Service
	client *asynq.Client
}

func NewRevealChecker(svc *revealsvc.Service, client *asynq.Client) *RevealChecker {
	return &RevealChecker{svc: svc, client: client}
}

func (h *RevealChecker) HandleRevealChecker(ctx context.Context, t *asynq.Task) error {
	seasons, err := h.svc.GetSeasonsForReveal(ctx)
	if err != nil {
		return fmt.Errorf("get seasons for reveal: %w", err)
	}

	for _, season := range seasons {
		payload, _ := json.Marshal(RevealPayload{
			SeasonID: season.ID,
			Attempt:  1,
		})
		task := asynq.NewTask(lib.TypeRevealProcess, payload)
		_, err := h.client.Enqueue(task,
			asynq.Queue("critical"),
			asynq.TaskID("reveal:"+season.ID),
		)
		if err != nil {
			log.Error().Err(err).Str("season_id", season.ID).Msg("failed to enqueue reveal-process")
		} else {
			log.Info().Str("season_id", season.ID).Msg("enqueued reveal-process")
		}
	}

	return nil
}

type RevealProcessor struct {
	svc    *revealsvc.Service
	client *asynq.Client
}

func NewRevealProcessor(svc *revealsvc.Service, client *asynq.Client) *RevealProcessor {
	return &RevealProcessor{svc: svc, client: client}
}

func (h *RevealProcessor) HandleRevealProcess(ctx context.Context, t *asynq.Task) error {
	var p RevealPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal reveal payload: %w", err)
	}

	result, err := h.svc.ProcessReveal(ctx, p.SeasonID, p.Attempt)
	if err != nil {
		return fmt.Errorf("process reveal for season %s: %w", p.SeasonID, err)
	}

	if result.Retry {
		// Re-enqueue with incremented attempt and 2-hour delay
		payload, _ := json.Marshal(RevealPayload{
			SeasonID: p.SeasonID,
			Attempt:  p.Attempt + 1,
		})
		task := asynq.NewTask(lib.TypeRevealProcess, payload)
		_, err := h.client.Enqueue(task,
			asynq.Queue("critical"),
			asynq.ProcessIn(2*time.Hour),
			asynq.TaskID(fmt.Sprintf("reveal:%s:attempt:%d", p.SeasonID, p.Attempt+1)),
		)
		if err != nil {
			log.Error().Err(err).Str("season_id", p.SeasonID).Int("attempt", p.Attempt).Msg("failed to re-enqueue reveal")
		} else {
			log.Info().Str("season_id", p.SeasonID).Int("next_attempt", p.Attempt+1).Msg("quorum not met, retrying in 2h")
		}
		return nil
	}

	if result.Revealed {
		// Enqueue downstream jobs (T11, T17, T19 — stubs for now)
		achievePayload, _ := json.Marshal(map[string]string{"season_id": p.SeasonID})
		achieveTask := asynq.NewTask(lib.TypeAchievements, achievePayload)
		if _, err := h.client.Enqueue(achieveTask, asynq.Queue("default")); err != nil {
			log.Error().Err(err).Str("season_id", p.SeasonID).Msg("failed to enqueue achievements")
		}

		pushPayload, _ := json.Marshal(map[string]string{"season_id": p.SeasonID})
		pushTask := asynq.NewTask(lib.TypePushReveal, pushPayload)
		if _, err := h.client.Enqueue(pushTask, asynq.Queue("default")); err != nil {
			log.Error().Err(err).Str("season_id", p.SeasonID).Msg("failed to enqueue reveal push")
		}

		// Card generation
		cardsPayload, _ := json.Marshal(CardsPayload{SeasonID: p.SeasonID})
		cardsTask := asynq.NewTask(lib.TypeCardsGenerate, cardsPayload)
		if _, err := h.client.Enqueue(cardsTask, asynq.Queue("default")); err != nil {
			log.Error().Err(err).Str("season_id", p.SeasonID).Msg("failed to enqueue card generation")
		}

		// Telegram post will be enqueued here when T19 is implemented
	}

	return nil
}
