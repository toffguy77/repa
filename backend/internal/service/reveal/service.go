package reveal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/repa-app/repa/internal/cardorder"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/eligibility"
	"github.com/repa-app/repa/internal/schedule"
	"github.com/rs/zerolog/log"
)

var (
	ErrSeasonNotFound    = errors.New("season not found")
	ErrSeasonNotRevealed = errors.New("season is not revealed")
	ErrNotMember         = errors.New("user is not a member of this group")
	ErrInsufficientFunds = errors.New("insufficient crystal balance")
	ErrGroupTooSmall     = errors.New("group is too small for this action")
	// ErrNothingToReveal means every voter is already revealed, nobody voted, or the full list is
	// already owned — in each case there is no hint left to sell.
	ErrNothingToReveal = errors.New("nothing left to reveal")
	// ErrSeasonAlreadyRevealed means the member has their card, so a partial pre-Reveal signal would
	// be noise next to a full result.
	ErrSeasonAlreadyRevealed = errors.New("season has already revealed")
)

// The reveal rules live in internal/eligibility so the groups API can explain them with
// the same numbers this service enforces.
const (
	MinRevealMembers   = eligibility.MinMembers
	MinRevealVoters    = eligibility.MinVoters
	MinDetectorMembers = eligibility.MinDetectorMembers
)

// Detector ladder prices.
//
// The count is free because its job is to *create* the question — charging for the question is how a
// single-purchase detector ends up selling only the answer. The hint sits below the referral grant
// (5 crystals) on purpose: a player who invited one friend can afford a rung without a bank card.
const (
	DetectorHintCost = 3
	DetectorFullCost = 10
)

// DetectorHint is one partially revealed voter.
//
// There is deliberately no username field: with no place to put a full name, no code path can leak
// one through a hint, and no future change to a shared DTO can accidentally add it.
type DetectorHint struct {
	// FirstLetter is the first *rune* of the voter's name — usernames here can begin with a digit or
	// a Cyrillic letter, and "first letter" would need a definition that excludes them.
	FirstLetter string  `json:"first_letter"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarURL   *string `json:"avatar_url"`
}

// firstRune returns the first rune of s as a string, so a multi-byte name is not cut in half.
func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

type Service struct {
	queries db.Querier
	sqlDB   *sql.DB
}

func NewService(queries db.Querier, sqlDB *sql.DB) *Service {
	return &Service{queries: queries, sqlDB: sqlDB}
}

// --- Reveal processing (called by worker) ---

type RevealResult struct {
	Revealed bool
	Retry    bool
	// Postponed is set when the participation floor was unmet and the season's reveal
	// moved to the next weekly slot instead of producing an empty card.
	Postponed bool
	// GroupID and Kind let the worker decide what to enqueue next without a second read.
	GroupID      string
	Kind         db.SeasonKind
	VotersNeeded int64
}

// Eligibility reports where a season stands relative to the reveal rules. It is the one
// predicate consumed by the reveal worker (gate), the voting service (kickoff scheduling)
// and the groups API (the reveal_state the app renders).
func (s *Service) Eligibility(ctx context.Context, seasonID string) (*eligibility.Status, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	return s.eligibilityForSeason(ctx, season)
}

func (s *Service) eligibilityForSeason(ctx context.Context, season db.Season) (*eligibility.Status, error) {
	memberCount, err := s.queries.CountGroupMembers(ctx, season.GroupID)
	if err != nil {
		return nil, err
	}

	totalQuestions, err := s.queries.CountSeasonQuestions(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	// "Voted" means completed every question, matching the progress shown in the app.
	votedCount, err := s.queries.CountCompletedVoters(ctx, db.CountCompletedVotersParams{
		SeasonID: season.ID,
		Column2:  totalQuestions,
	})
	if err != nil {
		return nil, err
	}

	status := eligibility.Evaluate(memberCount, votedCount)
	return &status, nil
}

func (s *Service) ProcessReveal(ctx context.Context, seasonID string, attempt int) (*RevealResult, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	if season.Status != db.SeasonStatusVOTING {
		return &RevealResult{Revealed: false}, nil
	}

	e, err := s.eligibilityForSeason(ctx, season)
	if err != nil {
		return nil, err
	}

	// If quorum is short but the floors could still be met, keep retrying.
	maxAttempts := 3
	if !e.Eligible && attempt < maxAttempts {
		return &RevealResult{
			Revealed:     false,
			Retry:        true,
			GroupID:      season.GroupID,
			Kind:         season.Kind,
			VotersNeeded: e.VotersNeeded,
		}, nil
	}

	// The floors are absolute: the forced path bypasses the quorum percentage, never them.
	if !e.FloorMet {
		if err := s.postponeSeason(ctx, season); err != nil {
			return nil, err
		}
		log.Info().
			Str("season_id", seasonID).
			Int64("members", e.MemberCount).
			Int64("voted", e.VotedCount).
			Int64("members_needed", e.MembersNeeded).
			Int64("voters_needed", e.VotersNeeded).
			Msg("reveal postponed: participation floor not met")
		return &RevealResult{
			Revealed:     false,
			Postponed:    true,
			GroupID:      season.GroupID,
			Kind:         season.Kind,
			VotersNeeded: e.VotersNeeded,
		}, nil
	}

	// Quorum met or forced reveal after max attempts — aggregate and update status atomically
	if err := s.aggregateAndReveal(ctx, seasonID); err != nil {
		return nil, err
	}

	log.Info().
		Str("season_id", seasonID).
		Int64("voters", e.VotedCount).
		Int64("members", e.MemberCount).
		Bool("quorum_met", e.QuorumMet).
		Int("attempt", attempt).
		Msg("season revealed")

	return &RevealResult{
		Revealed: true,
		GroupID:  season.GroupID,
		Kind:     season.Kind,
	}, nil
}

// ScheduleKickoffIfEligible brings a kickoff season's Reveal forward to
// schedule.KickoffRevealDelay from now, if the group has just become reveal-eligible.
// It reports whether this call was the one that scheduled it.
//
// The underlying UPDATE only fires when the season is still a VOTING kickoff whose
// reveal_at is later than the new time, which makes the operation idempotent and keeps a
// later voter from pushing the Reveal back.
func (s *Service) ScheduleKickoffIfEligible(ctx context.Context, seasonID string) (bool, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return false, err
	}
	if season.Kind != db.SeasonKindKICKOFF || season.Status != db.SeasonStatusVOTING {
		return false, nil
	}

	e, err := s.eligibilityForSeason(ctx, season)
	if err != nil {
		return false, err
	}
	if !e.Eligible {
		return false, nil
	}

	revealAt := time.Now().Add(schedule.KickoffRevealDelay)
	if _, err := s.queries.ScheduleKickoffReveal(ctx, db.ScheduleKickoffRevealParams{
		ID:       seasonID,
		RevealAt: revealAt,
		EndsAt:   schedule.EndOfRevealWeek(revealAt),
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Already scheduled by a concurrent completion.
			return false, nil
		}
		return false, err
	}

	log.Info().
		Str("season_id", seasonID).
		Time("reveal_at", revealAt).
		Msg("kickoff reveal scheduled")

	return true, nil
}

// allowedShareChannels is the closed set of channels a share can be attributed to. An
// unrecognised channel is recorded as "other" rather than refused: the share has already
// happened in the OS share sheet by the time we hear about it.
var allowedShareChannels = map[string]bool{
	"card":     true,
	"telegram": true,
	"link":     true,
	"code":     true,
}

// NormalizeShareChannel maps a client-supplied channel onto the allowed set.
func NormalizeShareChannel(raw string) string {
	c := strings.ToLower(strings.TrimSpace(raw))
	if allowedShareChannels[c] {
		return c
	}
	return "other"
}

// RecordShare notes that a member shared their card for a season.
//
// Best-effort by design: the user's share has already happened by the time this is called, so
// a failure here is logged and swallowed rather than reported as an error for something that
// succeeded. Validation is still real — a non-member or an unrevealed season is refused.
func (s *Service) RecordShare(ctx context.Context, seasonID, userID, channel string) error {
	if _, err := s.ValidateRevealAccess(ctx, seasonID, userID); err != nil {
		return err
	}

	_, err := s.queries.CreateShareEvent(ctx, db.CreateShareEventParams{
		ID:       uuid.New().String(),
		UserID:   userID,
		SeasonID: seasonID,
		Channel:  NormalizeShareChannel(channel),
	})
	if err != nil {
		log.Error().Err(err).
			Str("season_id", seasonID).
			Str("user_id", userID).
			Msg("failed to record share event")
	}
	// Deliberately not surfaced: see the doc comment.
	return nil
}

// detectorTooSmall reports whether a group is below the size where a voter list is still anonymous
// enough to sell *for this viewer*.
//
// Counts effective size — members minus anyone blocked in either direction — rather than raw
// membership. A group of five with three blocks presents itself as safely anonymous while behaving
// like a group of two, which is precisely the case the floor exists for.
func (s *Service) detectorTooSmall(ctx context.Context, groupID, userID string) (bool, error) {
	memberCount, err := s.queries.CountGroupMembers(ctx, groupID)
	if err != nil {
		return false, err
	}

	// Scoped to this group: the global block list would also subtract people the user blocked in
	// other groups, shrinking this one because of something that happened elsewhere.
	blocked, err := s.queries.CountBlockedGroupMembers(ctx, db.CountBlockedGroupMembersParams{
		GroupID:   groupID,
		BlockerID: userID,
	})
	if err != nil {
		// A failure here must not silently *open* a paid rung, so fall back to raw membership.
		return memberCount < MinDetectorMembers, nil
	}

	return memberCount-blocked < MinDetectorMembers, nil
}

// postponeSeason moves an ineligible season's reveal to the next weekly slot and keeps
// it open, so votes already cast are preserved rather than discarded.
func (s *Service) postponeSeason(ctx context.Context, season db.Season) error {
	revealAt := schedule.NextRevealFriday(time.Now())
	_, err := s.queries.PostponeSeason(ctx, db.PostponeSeasonParams{
		ID:       season.ID,
		RevealAt: revealAt,
		EndsAt:   schedule.EndOfRevealWeek(revealAt),
	})
	return err
}

func (s *Service) aggregateAndReveal(ctx context.Context, seasonID string) error {
	q := s.queries
	var tx *sql.Tx
	if s.sqlDB != nil {
		var txErr error
		tx, txErr = s.sqlDB.BeginTx(ctx, nil)
		if txErr != nil {
			return txErr
		}
		defer tx.Rollback()
		q = db.New(tx)
	}

	// Delete existing results for idempotency
	if err := q.DeleteSeasonResultsBySeason(ctx, seasonID); err != nil {
		return err
	}

	aggregated, err := q.AggregateVotesByTarget(ctx, seasonID)
	if err != nil {
		return err
	}

	uniqueVoters, err := q.CountUniqueVoters(ctx, seasonID)
	if err != nil {
		return err
	}

	totalVoters := int32(uniqueVoters)

	for _, agg := range aggregated {
		voteCount := int32(agg.VoteCount)
		percentage := 0.0
		if totalVoters > 0 {
			percentage = math.Round(float64(voteCount)/float64(totalVoters)*1000) / 10 // one decimal
		}

		_, err := q.CreateSeasonResult(ctx, db.CreateSeasonResultParams{
			ID:          uuid.New().String(),
			SeasonID:    seasonID,
			TargetID:    agg.TargetID,
			QuestionID:  agg.QuestionID,
			VoteCount:   voteCount,
			TotalVoters: totalVoters,
			Percentage:  percentage,
		})
		if err != nil {
			return err
		}
	}

	// Update season status atomically with aggregation
	if err := q.UpdateSeasonStatus(ctx, db.UpdateSeasonStatusParams{
		ID:     seasonID,
		Status: db.SeasonStatusREVEALED,
	}); err != nil {
		return err
	}

	if tx != nil {
		return tx.Commit()
	}
	return nil
}

// GetSeasonsForReveal returns seasons ready for reveal processing.
func (s *Service) GetSeasonsForReveal(ctx context.Context) ([]db.Season, error) {
	return s.queries.GetSeasonsForReveal(ctx)
}

// --- API endpoints ---

type AttributeDto struct {
	QuestionID   string  `json:"question_id"`
	QuestionText string  `json:"question_text"`
	Category     string  `json:"category"`
	Percentage   float64 `json:"percentage"`
	Rank         int     `json:"rank"`
}

type TrendDto struct {
	Attribute string  `json:"attribute"`
	Change    string  `json:"change"` // "up", "down", "same"
	Delta     float64 `json:"delta"`
}

type AchievementDto struct {
	Type     string `json:"type"`
	Metadata any    `json:"metadata,omitempty"`
}

type MyCard struct {
	TopAttributes    []AttributeDto   `json:"top_attributes"`
	HiddenAttributes []AttributeDto   `json:"hidden_attributes"`
	ReputationTitle  string           `json:"reputation_title"`
	Trend            *TrendDto        `json:"trend"`
	NewAchievements  []AchievementDto `json:"new_achievements"`
	CardImageURL     string           `json:"card_image_url"`
}

type GroupSummary struct {
	TopPerQuestion []TopAttributeDto `json:"top_per_question"`
	VoterCount     int64             `json:"voter_count"`
}

type TopAttributeDto struct {
	QuestionID   string  `json:"question_id"`
	QuestionText string  `json:"question_text"`
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	AvatarEmoji  *string `json:"avatar_emoji"`
	Percentage   float64 `json:"percentage"`
}

type RevealData struct {
	MyCard       MyCard       `json:"my_card"`
	GroupSummary GroupSummary `json:"group_summary"`
}

// ValidateRevealAccess checks that the season exists, is revealed, and the user is a member.
func (s *Service) ValidateRevealAccess(ctx context.Context, seasonID, userID string) (*db.Season, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeasonNotFound
		}
		return nil, err
	}

	if season.Status != db.SeasonStatusREVEALED {
		return nil, ErrSeasonNotRevealed
	}

	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return nil, err
	}
	if isMember == 0 {
		return nil, ErrNotMember
	}

	return &season, nil
}

func (s *Service) GetReveal(ctx context.Context, seasonID, userID string) (*RevealData, error) {
	season, err := s.ValidateRevealAccess(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	// Get user's results
	results, err := s.queries.GetSeasonResultsByUser(ctx, db.GetSeasonResultsByUserParams{
		SeasonID: seasonID,
		TargetID: userID,
	})
	if err != nil {
		return nil, err
	}

	topAttrs, hiddenAttrs := splitAttributes(results)

	// Compute trend from previous season
	trend := s.computeTrend(ctx, *season, userID, results)

	// Get group summary (top per question)
	topPerQ, err := s.queries.GetTopResultPerQuestion(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	voterCount, err := s.queries.CountUniqueVoters(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	topPerQuestion := make([]TopAttributeDto, len(topPerQ))
	for i, t := range topPerQ {
		dto := TopAttributeDto{
			QuestionID:   t.QuestionID,
			QuestionText: t.QuestionText,
			UserID:       t.TargetID,
			Username:     t.Username,
			Percentage:   t.Percentage,
		}
		if t.AvatarEmoji.Valid {
			dto.AvatarEmoji = &t.AvatarEmoji.String
		}
		topPerQuestion[i] = dto
	}

	title := generateTitle(topAttrs)

	// Look up card image URL from card_cache
	cardImageURL := ""
	cardCache, err := s.queries.GetCardCache(ctx, db.GetCardCacheParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err == nil {
		cardImageURL = cardCache.ImageUrl
	}

	return &RevealData{
		MyCard: MyCard{
			TopAttributes:    topAttrs,
			HiddenAttributes: hiddenAttrs,
			ReputationTitle:  title,
			Trend:            trend,
			NewAchievements:  s.getNewAchievements(ctx, seasonID, userID),
			CardImageURL:     cardImageURL,
		},
		GroupSummary: GroupSummary{
			TopPerQuestion: topPerQuestion,
			VoterCount:     voterCount,
		},
	}, nil
}

// --- Members cards ---

type MemberCardDto struct {
	UserID          string         `json:"user_id"`
	Username        string         `json:"username"`
	AvatarEmoji     *string        `json:"avatar_emoji"`
	AvatarURL       *string        `json:"avatar_url"`
	TopAttributes   []AttributeDto `json:"top_attributes"`
	ReputationTitle string         `json:"reputation_title"`
}

func (s *Service) GetMembersCards(ctx context.Context, seasonID, userID string) ([]MemberCardDto, error) {
	if _, err := s.ValidateRevealAccess(ctx, seasonID, userID); err != nil {
		return nil, err
	}

	// Single query: all results for this season with user info
	allResults, err := s.queries.GetAllSeasonResultsWithUsers(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	// A block hides the other person's card from this viewer. The viewer's own card is never hidden.
	hidden := map[string]bool{}
	if blocked, err := s.queries.ListBlockedUserIDs(ctx, userID); err == nil {
		for _, id := range blocked {
			hidden[id] = true
		}
	}

	// Group results by target user, preserving order (sorted by target_id, percentage DESC)
	type userInfo struct {
		username    string
		avatarEmoji sql.NullString
		avatarUrl   sql.NullString
		results     []db.GetSeasonResultsByUserRow
	}
	byUser := make(map[string]*userInfo)
	userOrder := make([]string, 0)

	for _, r := range allResults {
		if hidden[r.TargetID] && r.TargetID != userID {
			continue
		}
		ui, ok := byUser[r.TargetID]
		if !ok {
			ui = &userInfo{
				username:    r.Username,
				avatarEmoji: r.AvatarEmoji,
				avatarUrl:   r.AvatarUrl,
			}
			byUser[r.TargetID] = ui
			userOrder = append(userOrder, r.TargetID)
		}
		ui.results = append(ui.results, db.GetSeasonResultsByUserRow{
			QuestionID:       r.QuestionID,
			QuestionText:     r.QuestionText,
			QuestionCategory: r.QuestionCategory,
			// Carried across on purpose: splitAttributes orders by tone, and an unset tone would
			// read as neutral for every row — which silently degrades to plain percentage order
			// without failing anything.
			QuestionTone: r.QuestionTone,
			Percentage:   r.Percentage,
		})
	}

	cards := make([]MemberCardDto, 0, len(byUser))
	for _, uid := range userOrder {
		ui := byUser[uid]
		topAttrs, _ := splitAttributes(ui.results)

		card := MemberCardDto{
			UserID:          uid,
			Username:        ui.username,
			TopAttributes:   topAttrs,
			ReputationTitle: generateTitle(topAttrs),
		}
		if ui.avatarEmoji.Valid {
			card.AvatarEmoji = &ui.avatarEmoji.String
		}
		if ui.avatarUrl.Valid {
			card.AvatarURL = &ui.avatarUrl.String
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// --- Open hidden attributes ---

type OpenHiddenResult struct {
	AllAttributes  []AttributeDto `json:"all_attributes"`
	CrystalBalance int32          `json:"crystal_balance"`
}

func (s *Service) OpenHidden(ctx context.Context, seasonID, userID string) (*OpenHiddenResult, error) {
	if _, err := s.ValidateRevealAccess(ctx, seasonID, userID); err != nil {
		return nil, err
	}

	// Check balance and deduct atomically in a transaction
	const cost = 5

	q := s.queries
	var tx *sql.Tx
	if s.sqlDB != nil {
		var txErr error
		tx, txErr = s.sqlDB.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, txErr
		}
		defer tx.Rollback()
		q = db.New(tx)

		// Lock user row to prevent concurrent spend race
		if _, err := q.LockUserForUpdate(ctx, userID); err != nil {
			return nil, fmt.Errorf("lock user: %w", err)
		}
	}

	balance, err := q.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if balance < cost {
		return nil, ErrInsufficientFunds
	}

	newBalance := balance - cost
	_, err = q.CreateCrystalLog(ctx, db.CreateCrystalLogParams{
		ID:          uuid.New().String(),
		UserID:      userID,
		Delta:       -cost,
		Balance:     newBalance,
		Type:        db.CrystalLogTypeSPENDATTRIBUTES,
		Description: sql.NullString{String: "Open hidden attributes for season", Valid: true},
	})
	if err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	// Return all attributes
	results, err := s.queries.GetSeasonResultsByUser(ctx, db.GetSeasonResultsByUserParams{
		SeasonID: seasonID,
		TargetID: userID,
	})
	if err != nil {
		return nil, err
	}

	// Same order as the card, so the purchased "show everything" view is the card with more rows
	// rather than a differently ranked list of the same results.
	allAttrs := make([]AttributeDto, len(results))
	for i, r := range cardorder.ByTone(results) {
		allAttrs[i] = AttributeDto{
			QuestionID:   r.QuestionID,
			QuestionText: r.QuestionText,
			Category:     string(r.QuestionCategory),
			Percentage:   r.Percentage,
			Rank:         i + 1,
		}
	}

	return &OpenHiddenResult{
		AllAttributes:  allAttrs,
		CrystalBalance: newBalance,
	}, nil
}

// --- Detector ---

type VoterProfile struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarURL   *string `json:"avatar_url"`
}

type DetectorResult struct {
	Purchased bool           `json:"purchased"`
	Voters    []VoterProfile `json:"voters"`

	// The ladder. Older clients read Purchased and Voters and keep working, which is why these are
	// additive rather than a separate endpoint — the Reveal screen already fetches the detector.
	VoterCount    int64          `json:"voter_count"`
	Hints         []DetectorHint `json:"hints"`
	HintCost      int32          `json:"hint_cost"`
	FullCost      int32          `json:"full_cost"`
	HintAvailable bool           `json:"hint_available"`
	// Available is false while the group is below MinDetectorMembers, so the app can
	// disable the button instead of letting a member discover the limit by spending.
	Available      bool  `json:"available"`
	CrystalBalance int32 `json:"crystal_balance"`
}

func (s *Service) GetDetector(ctx context.Context, seasonID, userID string) (*DetectorResult, error) {
	season, err := s.ValidateRevealAccess(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	tooSmall, err := s.detectorTooSmall(ctx, season.GroupID, userID)
	if err != nil {
		return nil, err
	}

	hasDet, err := s.queries.HasDetector(ctx, db.HasDetectorParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err != nil {
		return nil, err
	}

	balance, err := s.queries.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user balance: %w", err)
	}

	// The free rung: how many people voted, named nobody. Its job is to create the question.
	voterCount, err := s.queries.CountSeasonVoters(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	hints, err := s.detectorHints(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	result := &DetectorResult{
		Purchased:      hasDet,
		Voters:         []VoterProfile{},
		Available:      !tooSmall,
		CrystalBalance: balance,
		VoterCount:     voterCount,
		Hints:          hints,
		HintCost:       DetectorHintCost,
		FullCost:       DetectorFullCost,
		// Nothing left to hint at once the list is owned, or once every voter is revealed.
		HintAvailable: !tooSmall && !hasDet && int64(len(hints)) < voterCount,
	}

	if hasDet {
		voters, err := s.queries.GetVoterProfilesBySeason(ctx, seasonID)
		if err != nil {
			return nil, err
		}
		for _, v := range voters {
			vp := VoterProfile{
				ID:       v.ID,
				Username: v.Username,
			}
			if v.AvatarEmoji.Valid {
				vp.AvatarEmoji = &v.AvatarEmoji.String
			}
			if v.AvatarUrl.Valid {
				vp.AvatarURL = &v.AvatarUrl.String
			}
			result.Voters = append(result.Voters, vp)
		}
	}

	return result, nil
}

func (s *Service) BuyDetector(ctx context.Context, seasonID, userID string) (*DetectorResult, error) {
	season, err := s.ValidateRevealAccess(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	// Checked before the transaction opens, so a refused purchase cannot touch the balance.
	tooSmall, err := s.detectorTooSmall(ctx, season.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if tooSmall {
		return nil, ErrGroupTooSmall
	}

	const cost = 10

	q := s.queries
	var tx *sql.Tx
	if s.sqlDB != nil {
		var txErr error
		tx, txErr = s.sqlDB.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, txErr
		}
		defer tx.Rollback()
		q = db.New(tx)

		// Lock user row to prevent concurrent spend race
		if _, err := q.LockUserForUpdate(ctx, userID); err != nil {
			return nil, fmt.Errorf("lock user: %w", err)
		}
	}

	// Check if already purchased (inside transaction to prevent race condition)
	hasDet, err := q.HasDetector(ctx, db.HasDetectorParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err != nil {
		return nil, err
	}
	if hasDet {
		if tx != nil {
			tx.Rollback()
		}
		return s.GetDetector(ctx, seasonID, userID)
	}

	balance, err := q.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if balance < cost {
		return nil, ErrInsufficientFunds
	}

	newBalance := balance - cost
	_, err = q.CreateCrystalLog(ctx, db.CreateCrystalLogParams{
		ID:          uuid.New().String(),
		UserID:      userID,
		Delta:       -cost,
		Balance:     newBalance,
		Type:        db.CrystalLogTypeSPENDDETECTOR,
		Description: sql.NullString{String: "Detector for season", Valid: true},
	})
	if err != nil {
		return nil, err
	}

	_, err = q.CreateDetector(ctx, db.CreateDetectorParams{
		ID:       uuid.New().String(),
		UserID:   userID,
		SeasonID: seasonID,
		GroupID:  season.GroupID,
	})
	if err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	// Fetch voters now that detector is purchased
	voters, err := s.queries.GetVoterProfilesBySeason(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	voterProfiles := make([]VoterProfile, 0, len(voters))
	for _, v := range voters {
		vp := VoterProfile{
			ID:       v.ID,
			Username: v.Username,
		}
		if v.AvatarEmoji.Valid {
			vp.AvatarEmoji = &v.AvatarEmoji.String
		}
		if v.AvatarUrl.Valid {
			vp.AvatarURL = &v.AvatarUrl.String
		}
		voterProfiles = append(voterProfiles, vp)
	}

	return &DetectorResult{
		Purchased:      true,
		Voters:         voterProfiles,
		Available:      true,
		CrystalBalance: newBalance,
	}, nil
}

// --- Helpers ---

func splitAttributes(results []db.GetSeasonResultsByUserRow) (top []AttributeDto, hidden []AttributeDto) {
	const topN = 3
	// Ordered before the split, so tone decides which near-tied attribute lands in the top three and
	// is therefore on the shared card — reordering only within the top three would be pointless.
	for i, r := range cardorder.ByTone(results) {
		attr := AttributeDto{
			QuestionID:   r.QuestionID,
			QuestionText: r.QuestionText,
			Category:     string(r.QuestionCategory),
			Percentage:   r.Percentage,
			Rank:         i + 1,
		}
		if i < topN {
			top = append(top, attr)
		} else {
			hidden = append(hidden, attr)
		}
	}
	if top == nil {
		top = []AttributeDto{}
	}
	if hidden == nil {
		hidden = []AttributeDto{}
	}
	return
}

func (s *Service) computeTrend(ctx context.Context, season db.Season, userID string, currentResults []db.GetSeasonResultsByUserRow) *TrendDto {
	if len(currentResults) == 0 {
		return nil
	}

	prevSeason, err := s.queries.GetPreviousRevealedSeason(ctx, db.GetPreviousRevealedSeasonParams{
		GroupID: season.GroupID,
		ID:      season.ID,
	})
	if err != nil {
		return nil
	}

	prevResults, err := s.queries.GetSeasonResultsByUser(ctx, db.GetSeasonResultsByUserParams{
		SeasonID: prevSeason.ID,
		TargetID: userID,
	})
	if err != nil || len(prevResults) == 0 {
		return nil
	}

	// The attribute the *card* leads with, not the highest percentage: the trend line sits next to
	// the card, and naming a different attribute than the one shown above it reads as a bug.
	currentTop := cardorder.ByTone(currentResults)[0]

	// Find same question in previous results
	for _, prev := range prevResults {
		if prev.QuestionID == currentTop.QuestionID {
			delta := currentTop.Percentage - prev.Percentage
			change := "same"
			if delta > 1 {
				change = "up"
			} else if delta < -1 {
				change = "down"
			}
			return &TrendDto{
				Attribute: currentTop.QuestionText,
				Change:    change,
				Delta:     math.Round(delta*10) / 10,
			}
		}
	}

	return nil
}

func (s *Service) getNewAchievements(ctx context.Context, seasonID, userID string) []AchievementDto {
	achievements, err := s.queries.GetSeasonAchievements(ctx, sql.NullString{String: seasonID, Valid: true})
	if err != nil {
		return []AchievementDto{}
	}

	result := []AchievementDto{}
	for _, a := range achievements {
		if a.UserID != userID {
			continue
		}
		dto := AchievementDto{
			Type: string(a.AchievementType),
		}
		if a.Metadata.Valid {
			var meta map[string]any
			if json.Unmarshal(a.Metadata.RawMessage, &meta) == nil {
				dto.Metadata = meta
			}
		}
		result = append(result, dto)
	}
	return result
}

func generateTitle(topAttrs []AttributeDto) string {
	if len(topAttrs) == 0 {
		return "Загадка века"
	}
	// Simple category-based title mapping
	cat := topAttrs[0].Category
	switch cat {
	case "HOT":
		return "Горячая штучка"
	case "FUNNY":
		return "Душа компании"
	case "SECRETS":
		return "Хранитель тайн"
	case "SKILLS":
		return "Мастер на все руки"
	case "ROMANCE":
		return "Сердцеед"
	case "STUDY":
		return "Ботан года"
	default:
		return "Загадка века"
	}
}

// detectorHints returns the voters already revealed to this member, as partial identities.
func (s *Service) detectorHints(ctx context.Context, seasonID, userID string) ([]DetectorHint, error) {
	rows, err := s.queries.GetDetectorHints(ctx, db.GetDetectorHintsParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err != nil {
		return nil, err
	}

	hints := make([]DetectorHint, 0, len(rows))
	for _, r := range rows {
		h := DetectorHint{FirstLetter: firstRune(r.Username)}
		if r.AvatarEmoji.Valid {
			h.AvatarEmoji = &r.AvatarEmoji.String
		}
		if r.AvatarUrl.Valid {
			h.AvatarURL = &r.AvatarUrl.String
		}
		hints = append(hints, h)
	}
	return hints, nil
}

// BuyDetectorHint reveals one more voter partially: enough to guess, not enough to know.
//
// Refuses — without spending — when the group is too small for any paid rung, when the full list is
// already owned, when nobody voted, and when every voter has already been revealed. The crystal
// spend and the hint record happen in one transaction, so a crash between them cannot charge for
// nothing.
func (s *Service) BuyDetectorHint(ctx context.Context, seasonID, userID string) (*DetectorResult, error) {
	season, err := s.ValidateRevealAccess(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	// A hint drawn from two or three possible voters identifies someone as surely as a list does.
	tooSmall, err := s.detectorTooSmall(ctx, season.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if tooSmall {
		return nil, ErrGroupTooSmall
	}

	hasDet, err := s.queries.HasDetector(ctx, db.HasDetectorParams{UserID: userID, SeasonID: seasonID})
	if err != nil {
		return nil, err
	}
	if hasDet {
		// There is nothing left to hint at.
		return nil, ErrNothingToReveal
	}

	q := s.queries
	var tx *sql.Tx
	if s.sqlDB != nil {
		var txErr error
		tx, txErr = s.sqlDB.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, txErr
		}
		defer tx.Rollback()
		q = db.New(tx)

		if _, err := q.LockUserForUpdate(ctx, userID); err != nil {
			return nil, fmt.Errorf("lock user: %w", err)
		}
	}

	// Pick before charging: "no voters left" must not cost crystals.
	voter, err := q.PickUnrevealedVoter(ctx, db.PickUnrevealedVoterParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNothingToReveal
		}
		return nil, err
	}

	balance, err := q.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	if balance < DetectorHintCost {
		return nil, ErrInsufficientFunds
	}

	newBalance := balance - DetectorHintCost
	if _, err := q.CreateCrystalLog(ctx, db.CreateCrystalLogParams{
		ID:          uuid.New().String(),
		UserID:      userID,
		Delta:       -DetectorHintCost,
		Balance:     newBalance,
		Type:        db.CrystalLogTypeSPENDDETECTOR,
		Description: sql.NullString{String: "Подсказка детектора", Valid: true},
	}); err != nil {
		return nil, err
	}

	if _, err := q.CreateDetectorHint(ctx, db.CreateDetectorHintParams{
		ID:             uuid.New().String(),
		UserID:         userID,
		SeasonID:       seasonID,
		RevealedUserID: voter.ID,
	}); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return s.GetDetector(ctx, seasonID, userID)
}
