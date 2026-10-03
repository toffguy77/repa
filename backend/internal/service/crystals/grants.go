package crystals

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

// Grant amounts.
//
// The welcome grant is deliberately exactly one detector rather than two: the goal is to create
// the want, not to satisfy it. Two referrals buy a detector, which makes inviting a real path to
// the hook for someone with no card.
const (
	WelcomeGrant   = 10
	ReferralGrant  = 5
	StreakGrant    = 5
	RecruiterGrant = 10
)

// reasonWelcome and reasonReferral are shown in the user's crystal history, so a balance that
// grew without a purchase is not mysterious.
const (
	reasonWelcome  = "Подарок за регистрацию"
	reasonReferral = "Друг присоединился и проголосовал"
)

// GrantWelcome gives a new user enough for exactly one detector, so the product's central hook
// is experienced before it is sold.
func (s *Service) GrantWelcome(ctx context.Context, userID string) {
	s.grant(ctx, userID, WelcomeGrant, "welcome:"+userID, reasonWelcome)
}

// GrantReferral pays the inviter once the invited member has completed a voting session.
//
// Keyed on the *invitee*, which is what makes "once per invited member" true by construction
// rather than by a count query that races.
func (s *Service) GrantReferral(ctx context.Context, inviterID, inviteeID string) {
	if inviterID == "" || inviteeID == "" || inviterID == inviteeID {
		return
	}
	s.grant(ctx, inviterID, ReferralGrant, "referral:"+inviteeID, reasonReferral)
}

// GrantForAchievement pays out the achievements that mark sustained play or successful
// recruiting. Keyed on the achievement id, so re-running the achievement job cannot double-pay.
//
// Only these two pay: paying for every achievement would make the currency meaningless, and
// would reward BLIND (accuracy under 20%), which is a joke badge.
func (s *Service) GrantForAchievement(ctx context.Context, userID, achievementID string, achievementType db.AchievementType) {
	amount, reason := achievementGrant(achievementType)
	if amount == 0 {
		return
	}
	s.grant(ctx, userID, amount, "achievement:"+achievementID, reason)
}

// achievementGrant maps an achievement type to its payout. Amounts live beside the rule that
// earns them, so changing one is a code review rather than a data migration.
func achievementGrant(t db.AchievementType) (int32, string) {
	switch t {
	case db.AchievementTypeSTREAKVOTER:
		return StreakGrant, "Серия голосований"
	case db.AchievementTypeRECRUITER:
		return RecruiterGrant, "Привёл друзей в группу"
	default:
		return 0, ""
	}
}

// grant writes a BONUS crystal log, treating a duplicate as success.
//
// externalID is derived rather than random: a random id would make every retry a new payment.
// crystal_logs.external_id is UNIQUE, so a replayed job hits the constraint and the grant is
// already present — which is what the caller wanted.
//
// Failures are logged, never returned: every trigger for a grant is a path whose primary job
// (registering, recording a vote, awarding an achievement) must not be rolled back by a payout.
func (s *Service) grant(ctx context.Context, userID string, amount int32, externalID, reason string) {
	balance, err := s.queries.GetUserBalance(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("grant", externalID).Msg("grant: failed to read balance")
		return
	}

	_, err = s.queries.CreateCrystalLog(ctx, db.CreateCrystalLogParams{
		ID:          uuid.New().String(),
		UserID:      userID,
		Delta:       amount,
		Balance:     balance + amount,
		Type:        db.CrystalLogTypeBONUS,
		Description: sql.NullString{String: reason, Valid: reason != ""},
		ExternalID:  sql.NullString{String: externalID, Valid: externalID != ""},
	})
	if err != nil {
		if isDuplicateGrant(err) {
			// Already paid. Not an error for the caller: the grant exists.
			log.Debug().Str("grant", externalID).Msg("grant already paid, skipping")
			return
		}
		log.Error().Err(err).Str("user_id", userID).Str("grant", externalID).Msg("grant: failed to write crystal log")
		return
	}

	log.Info().
		Str("user_id", userID).
		Int32("amount", amount).
		Str("grant", externalID).
		Msg("crystals granted")
}

// isDuplicateGrant reports whether err is the unique-violation that means this grant was already
// paid. Matched on the message rather than a driver type so the grants service does not depend on
// which Postgres driver is in use.
func isDuplicateGrant(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "externalid") && strings.Contains(msg, "exists")
}
