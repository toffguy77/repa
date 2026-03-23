package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	pushsvc "github.com/repa-app/repa/internal/service/push"
	"github.com/rs/zerolog/log"
)

type PushProcessor struct {
	svc *pushsvc.Service
}

func NewPushProcessor(svc *pushsvc.Service) *PushProcessor {
	return &PushProcessor{svc: svc}
}

// HandleWeeklyScheduler — Monday 17:00 MSK: "New season started" to all VOTING season members.
func (p *PushProcessor) HandleWeeklyScheduler(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	seasons, err := queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}

	for _, season := range seasons {
		if err := p.svc.SendToGroupMembers(ctx, season.GroupID, db.PushCategorySEASONSTART,
			"Новый сезон начался!",
			"Время голосовать — кто ответит за тебя? 🍆",
			map[string]string{"screen": "voting", "groupId": season.GroupID, "seasonId": season.ID},
		); err != nil {
			log.Warn().Err(err).Str("group_id", season.GroupID).Msg("weekly push failed")
		}
	}

	return nil
}

// HandleTuesdaySignal — Tuesday 19:00 MSK: tease voters that someone already voted about them.
func (p *PushProcessor) HandleTuesdaySignal(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	seasons, err := queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}

	for _, season := range seasons {
		votedUsers, err := queries.GetVotedUsersBySeason(ctx, season.ID)
		if err != nil {
			log.Warn().Err(err).Str("season_id", season.ID).Msg("failed to get voted users")
			continue
		}

		// Send to members who already voted (they know someone answered about them)
		p.svc.SendToUsers(ctx, votedUsers, db.PushCategoryREMINDER,
			"Кто-то уже ответил на вопросы про тебя 👀",
			"Зайди и посмотри, кто активен",
			map[string]string{"screen": "reveal-waiting", "groupId": season.GroupID},
		)
	}

	return nil
}

// HandleWednesdayQuorum — Wednesday 15:00 UTC: quorum status push.
func (p *PushProcessor) HandleWednesdayQuorum(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	seasons, err := queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}

	for _, season := range seasons {
		memberCount, err := queries.CountGroupMembers(ctx, season.GroupID)
		if err != nil {
			log.Warn().Err(err).Str("group_id", season.GroupID).Msg("failed to count members")
			continue
		}

		voterCount, err := queries.CountUniqueVoters(ctx, season.ID)
		if err != nil {
			log.Warn().Err(err).Str("season_id", season.ID).Msg("failed to count voters")
			continue
		}

		if memberCount == 0 {
			continue
		}

		quorumPct := float64(voterCount) / float64(memberCount) * 100

		if quorumPct >= 80 {
			// Almost done — push to non-voters
			nonVoters, err := queries.GetNonVotersByseason(ctx, db.GetNonVotersByseasonParams{
				GroupID:  season.GroupID,
				SeasonID: season.ID,
			})
			if err != nil {
				log.Warn().Err(err).Msg("failed to get non-voters")
				continue
			}
			p.svc.SendToUsers(ctx, nonVoters, db.PushCategoryREMINDER,
				"Все уже ответили, остался только ты",
				"Не пропусти Reveal — голосуй сейчас!",
				map[string]string{"screen": "voting", "groupId": season.GroupID, "seasonId": season.ID},
			)
		} else if quorumPct < 50 {
			// Quorum at risk — push to everyone
			remaining := memberCount/2 - voterCount + 1
			if remaining < 1 {
				remaining = 1
			}
			if err := p.svc.SendToGroupMembers(ctx, season.GroupID, db.PushCategoryREMINDER,
				"Reveal под угрозой!",
				fmt.Sprintf("Не хватает %d голосов — помоги группе", remaining),
				map[string]string{"screen": "voting", "groupId": season.GroupID, "seasonId": season.ID},
			); err != nil {
				log.Warn().Err(err).Str("group_id", season.GroupID).Msg("quorum push failed")
			}
		}
	}

	return nil
}

// HandleThursdayTeaser — Thursday 20:00 MSK: tease leading category to voters.
func (p *PushProcessor) HandleThursdayTeaser(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	seasons, err := queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}

	categoryEmoji := map[db.QuestionCategory]string{
		db.QuestionCategoryHOT:     "🔥",
		db.QuestionCategoryFUNNY:   "😂",
		db.QuestionCategorySECRETS: "🤫",
		db.QuestionCategorySKILLS:  "💪",
		db.QuestionCategoryROMANCE: "💕",
		db.QuestionCategorySTUDY:   "📚",
	}

	for _, season := range seasons {
		topCategory, err := queries.GetTopVotedCategoryBySeason(ctx, season.ID)
		if err != nil {
			log.Warn().Err(err).Str("season_id", season.ID).Msg("failed to get top category")
			continue
		}

		emoji := categoryEmoji[topCategory]
		if emoji == "" {
			emoji = "🍆"
		}

		votedUsers, err := queries.GetVotedUsersBySeason(ctx, season.ID)
		if err != nil {
			log.Warn().Err(err).Str("season_id", season.ID).Msg("failed to get voted users")
			continue
		}

		p.svc.SendToUsers(ctx, votedUsers, db.PushCategoryREMINDER,
			fmt.Sprintf("Один твой атрибут уже почти определился… %s", emoji),
			"Скоро Reveal — узнай, что думают о тебе",
			map[string]string{"screen": "reveal-waiting", "groupId": season.GroupID},
		)
	}

	return nil
}

// HandleFridayPreReveal — Friday 16:00 UTC (1h before reveal): reminder to everyone.
func (p *PushProcessor) HandleFridayPreReveal(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	seasons, err := queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}

	for _, season := range seasons {
		if err := p.svc.SendToGroupMembers(ctx, season.GroupID, db.PushCategoryREVEAL,
			"Через час — репа. Готов?",
			"Reveal уже скоро — последний шанс проголосовать!",
			map[string]string{"screen": "voting", "groupId": season.GroupID, "seasonId": season.ID},
		); err != nil {
			log.Warn().Err(err).Str("group_id", season.GroupID).Msg("pre-reveal push failed")
		}
	}

	return nil
}

type RevealPushPayload struct {
	SeasonID string `json:"season_id"`
}

// HandleRevealNotification — after successful reveal: "Your repa is ready".
func (p *PushProcessor) HandleRevealNotification(ctx context.Context, t *asynq.Task) error {
	var payload RevealPushPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal reveal push payload: %w", err)
	}

	queries := p.svc.Queries()
	season, err := queries.GetSeasonByID(ctx, payload.SeasonID)
	if err != nil {
		return fmt.Errorf("get season: %w", err)
	}

	if err := p.svc.SendToGroupMembers(ctx, season.GroupID, db.PushCategoryREVEAL,
		"Твоя репа готова 🍆",
		"Зайди и посмотри свою карточку репутации!",
		map[string]string{"screen": "reveal", "groupId": season.GroupID, "seasonId": season.ID},
	); err != nil {
		return fmt.Errorf("send reveal push: %w", err)
	}

	return nil
}

// HandleSundayPreview — Sunday 12:00 MSK: vote for next week's questions.
func (p *PushProcessor) HandleSundayPreview(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	groups, err := queries.GetAllGroupsWithMembers(ctx)
	if err != nil {
		return fmt.Errorf("get groups: %w", err)
	}

	for _, group := range groups {
		if err := p.svc.SendToGroupMembers(ctx, group.ID, db.PushCategoryNEXTSEASON,
			"Голосуй за вопросы следующей недели",
			"Выбери, о чём будут спрашивать в новом сезоне",
			map[string]string{"screen": "question-vote", "groupId": group.ID},
		); err != nil {
			log.Warn().Err(err).Str("group_id", group.ID).Msg("sunday preview push failed")
		}
	}

	return nil
}

// HandleSundayStreak — Sunday 18:00 MSK: remind streak holders.
func (p *PushProcessor) HandleSundayStreak(ctx context.Context, t *asynq.Task) error {
	queries := p.svc.Queries()

	groups, err := queries.GetAllGroupsWithMembers(ctx)
	if err != nil {
		return fmt.Errorf("get groups: %w", err)
	}

	for _, group := range groups {
		streakUsers, err := queries.GetUsersWithStreakInGroup(ctx, db.GetUsersWithStreakInGroupParams{
			GroupID:      group.ID,
			VotingStreak: 3,
		})
		if err != nil {
			log.Warn().Err(err).Str("group_id", group.ID).Msg("failed to get streak users")
			continue
		}

		p.svc.SendToUsers(ctx, streakUsers, db.PushCategoryREMINDER,
			"Не прерывай серию — скоро новый сезон 🔥",
			"Твоя серия голосований на кону!",
			nil,
		)
	}

	return nil
}

// HandleReactionPush — when someone reacts to a user's card.
func (p *PushProcessor) HandleReactionPush(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		TargetID  string `json:"target_id"`
		ReactorID string `json:"reactor_id"`
		Emoji     string `json:"emoji"`
		GroupID   string `json:"group_id"`
		SeasonID  string `json:"season_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal reaction payload: %w", err)
	}

	return p.svc.SendToUser(ctx, payload.TargetID, db.PushCategoryREACTION,
		fmt.Sprintf("Кто-то отреагировал на твою карточку %s", payload.Emoji),
		"Зайди и посмотри!",
		map[string]string{"screen": "reveal", "groupId": payload.GroupID, "seasonId": payload.SeasonID},
	)
}

