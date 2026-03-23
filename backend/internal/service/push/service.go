package push

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var mskLocation *time.Location

func init() {
	var err error
	mskLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		mskLocation = time.FixedZone("MSK", 3*60*60)
	}
}

type Service struct {
	queries *db.Queries
	rdb     *redis.Client
	fcm     *lib.FCMClient
}

func NewService(queries *db.Queries, rdb *redis.Client, fcm *lib.FCMClient) *Service {
	return &Service{queries: queries, rdb: rdb, fcm: fcm}
}

func (s *Service) canSendPush(ctx context.Context, userID string) bool {
	now := time.Now().In(mskLocation)

	// Quiet hours: 23:00-09:00 MSK
	hour := now.Hour()
	if hour >= 23 || hour < 9 {
		return false
	}

	// Rate limit: max 3 pushes per day per user
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", userID, dateKey)

	count, err := s.rdb.Get(ctx, redisKey).Int()
	if err != nil && err != redis.Nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("failed to check push rate limit")
		return true // fail open
	}

	return count < 3
}

func (s *Service) incrementPushCount(ctx context.Context, userID string) {
	now := time.Now().In(mskLocation)
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", userID, dateKey)

	pipe := s.rdb.Pipeline()
	pipe.Incr(ctx, redisKey)

	// TTL until end of day MSK
	endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, mskLocation)
	pipe.ExpireAt(ctx, redisKey, endOfDay)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("failed to increment push count")
	}
}

func (s *Service) SendToUser(ctx context.Context, userID string, category db.PushCategory, title, body string, data map[string]string) error {
	if s.fcm == nil {
		log.Debug().Str("user_id", userID).Msg("FCM not configured, skipping push")
		return nil
	}

	if !s.canSendPush(ctx, userID) {
		log.Debug().Str("user_id", userID).Msg("push rate limited or quiet hours")
		return nil
	}

	// Check user preference
	enabled, err := s.queries.IsPushEnabled(ctx, db.IsPushEnabledParams{
		UserID:   userID,
		Category: category,
	})
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("failed to check push preference")
	} else if val, ok := enabled.(bool); ok && !val {
		return nil
	}

	if err := s.fcm.SendPushToUser(ctx, userID, title, body, data); err != nil {
		return err
	}

	s.incrementPushCount(ctx, userID)
	return nil
}

func (s *Service) SendToUsers(ctx context.Context, userIDs []string, category db.PushCategory, title, body string, data map[string]string) {
	for _, userID := range userIDs {
		if err := s.SendToUser(ctx, userID, category, title, body, data); err != nil {
			log.Warn().Err(err).Str("user_id", userID).Msg("failed to send push")
		}
	}
}

func (s *Service) SendToGroupMembers(ctx context.Context, groupID string, category db.PushCategory, title, body string, data map[string]string) error {
	memberIDs, err := s.queries.GetGroupMemberIDs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group members: %w", err)
	}

	s.SendToUsers(ctx, memberIDs, category, title, body, data)
	return nil
}

func (s *Service) Queries() *db.Queries {
	return s.queries
}
