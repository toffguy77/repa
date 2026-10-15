package tasks

import (
	"context"
	"encoding/json"
	"fmt"

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

// SeasonForGroupPayload targets a single group, used by the post-kickoff follow-up so a
// group that reveals mid-week starts its weekly cycle immediately.
type SeasonForGroupPayload struct {
	GroupID string `json:"group_id"`
}

func (s *SeasonCreator) HandleSeasonMaintain(ctx context.Context, t *asynq.Task) error {
	if err := s.svc.MaintainSeasons(ctx); err != nil {
		log.Error().Err(err).Msg("season maintenance: failed")
		return err
	}
	return nil
}

func (s *SeasonCreator) HandleSeasonForGroup(ctx context.Context, t *asynq.Task) error {
	var p SeasonForGroupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal season-for-group payload: %w", err)
	}

	if err := s.svc.CreateWeeklySeasonForGroup(ctx, p.GroupID); err != nil {
		log.Error().Err(err).Str("group_id", p.GroupID).Msg("failed to create weekly season for group")
		return err
	}

	log.Info().Str("group_id", p.GroupID).Msg("weekly season ensured for group")
	return nil
}
