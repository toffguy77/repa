package lib

import "github.com/hibiken/asynq"

// Task type constants for asynq workers.
const (
	TypeRevealChecker    = "reveal:checker"
	TypeRevealProcess    = "reveal:process"
	TypeSeasonCreator    = "season:creator"
	TypeAchievements     = "achievements:calculate"
	TypePushWeekly       = "push:weekly-scheduler"
	TypePushTuesday      = "push:tuesday-signal"
	TypePushWednesday    = "push:wednesday-quorum"
	TypePushThursday     = "push:thursday-teaser"
	TypePushFriPreReveal = "push:friday-pre-reveal"
	TypePushReveal       = "push:reveal-notification"
	TypePushSundayPrev   = "push:sunday-preview"
	TypePushSundayStreak = "push:sunday-streak"
	TypeTelegramStart    = "telegram:season-start"
	TypeTelegramReveal   = "telegram:reveal-post"
	TypeTelegramShare    = "telegram:share-card"
	TypeReactionPush     = "push:reaction"
)

func NewAsynqClient(redisURL string) (*asynq.Client, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}
	return asynq.NewClient(opts), nil
}

func NewAsynqServer(redisURL string) (*asynq.Server, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}
	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})
	return srv, nil
}
