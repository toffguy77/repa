package reveal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

var (
	ErrSeasonNotFound    = errors.New("season not found")
	ErrSeasonNotRevealed = errors.New("season is not revealed")
	ErrNotMember         = errors.New("user is not a member of this group")
	ErrInsufficientFunds = errors.New("insufficient crystal balance")
)

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
}

func (s *Service) ProcessReveal(ctx context.Context, seasonID string, attempt int) (*RevealResult, error) {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	if season.Status != db.SeasonStatusVOTING {
		return &RevealResult{Revealed: false}, nil
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, season.GroupID)
	if err != nil {
		return nil, err
	}

	uniqueVoters, err := s.queries.CountUniqueVoters(ctx, seasonID)
	if err != nil {
		return nil, err
	}

	threshold := 0.5
	if memberCount < 8 {
		threshold = 0.4
	}

	quorumMet := float64(uniqueVoters) >= float64(memberCount)*threshold

	// If quorum not met and we haven't exhausted retries, defer
	maxAttempts := 3
	if !quorumMet && attempt < maxAttempts {
		return &RevealResult{Revealed: false, Retry: true}, nil
	}

	// Quorum met or forced reveal after max attempts — aggregate results
	if err := s.aggregateResults(ctx, seasonID); err != nil {
		return nil, err
	}

	// Update season status to REVEALED
	if err := s.queries.UpdateSeasonStatus(ctx, db.UpdateSeasonStatusParams{
		ID:     seasonID,
		Status: db.SeasonStatusREVEALED,
	}); err != nil {
		return nil, err
	}

	log.Info().
		Str("season_id", seasonID).
		Int64("voters", uniqueVoters).
		Int64("members", memberCount).
		Bool("quorum_met", quorumMet).
		Int("attempt", attempt).
		Msg("season revealed")

	return &RevealResult{Revealed: true}, nil
}

func (s *Service) aggregateResults(ctx context.Context, seasonID string) error {
	// Delete existing results for idempotency
	if err := s.queries.DeleteSeasonResultsBySeason(ctx, seasonID); err != nil {
		return err
	}

	aggregated, err := s.queries.AggregateVotesByTarget(ctx, seasonID)
	if err != nil {
		return err
	}

	uniqueVoters, err := s.queries.CountUniqueVoters(ctx, seasonID)
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

		_, err := s.queries.CreateSeasonResult(ctx, db.CreateSeasonResultParams{
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
	season, err := s.ValidateRevealAccess(ctx, seasonID, userID)
	if err != nil {
		return nil, err
	}

	members, err := s.queries.GetGroupMembers(ctx, season.GroupID)
	if err != nil {
		return nil, err
	}

	cards := make([]MemberCardDto, 0, len(members))
	for _, m := range members {
		results, err := s.queries.GetSeasonResultsByUser(ctx, db.GetSeasonResultsByUserParams{
			SeasonID: seasonID,
			TargetID: m.ID,
		})
		if err != nil {
			return nil, err
		}

		topAttrs, _ := splitAttributes(results)

		card := MemberCardDto{
			UserID:          m.ID,
			Username:        m.Username,
			TopAttributes:   topAttrs,
			ReputationTitle: generateTitle(topAttrs),
		}
		if m.AvatarEmoji.Valid {
			card.AvatarEmoji = &m.AvatarEmoji.String
		}
		if m.AvatarUrl.Valid {
			card.AvatarURL = &m.AvatarUrl.String
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

	allAttrs := make([]AttributeDto, len(results))
	for i, r := range results {
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
	Purchased      bool            `json:"purchased"`
	Voters         []VoterProfile  `json:"voters"`
	CrystalBalance int32           `json:"crystal_balance"`
}

func (s *Service) GetDetector(ctx context.Context, seasonID, userID string) (*DetectorResult, error) {
	if _, err := s.ValidateRevealAccess(ctx, seasonID, userID); err != nil {
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
		balance = 0
	}

	result := &DetectorResult{
		Purchased:      hasDet,
		Voters:         []VoterProfile{},
		CrystalBalance: balance,
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
		CrystalBalance: newBalance,
	}, nil
}

// --- Helpers ---

func splitAttributes(results []db.GetSeasonResultsByUserRow) (top []AttributeDto, hidden []AttributeDto) {
	const topN = 3
	for i, r := range results {
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

	// Compare top attribute
	currentTop := currentResults[0]

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
