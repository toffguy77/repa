package tasks

import (
	"context"

	"github.com/hibiken/asynq"
	groupssvc "github.com/repa-app/repa/internal/service/groups"
	"github.com/rs/zerolog/log"
)

type SeasonCreator struct {
	svc *groupssvc.Service
}

func NewSeasonCreator(svc *groupssvc.Service) *SeasonCreator {
	return &SeasonCreator{svc: svc}
}

func (s *SeasonCreator) HandleSeasonCreator(ctx context.Context, t *asynq.Task) error {
	log.Info().Msg("season creator: creating new seasons for active groups")

	if err := s.svc.CreateNewSeasons(ctx); err != nil {
		log.Error().Err(err).Msg("season creator: failed")
		return err
	}

	log.Info().Msg("season creator: completed")
	return nil
}
