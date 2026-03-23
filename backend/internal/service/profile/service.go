package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNotMember    = errors.New("user is not a member of this group")
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

type ProfileResponse struct {
	User           UserDto          `json:"user"`
	Stats          StatsDto         `json:"stats"`
	Achievements   []AchievementDto `json:"achievements"`
	Legend         string           `json:"legend"`
	SeasonHistory []SeasonCardDto  `json:"season_history"`
}

type UserDto struct {
	Username    string  `json:"username"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarUrl   *string `json:"avatar_url"`
}

type StatsDto struct {
	SeasonsPlayed      int32              `json:"seasons_played"`
	VotingStreak       int32              `json:"voting_streak"`
	MaxVotingStreak    int32              `json:"max_voting_streak"`
	GuessAccuracy      float64            `json:"guess_accuracy"`
	TotalVotesCast     int32              `json:"total_votes_cast"`
	TotalVotesReceived int32              `json:"total_votes_received"`
	TopAttributeAllTime *TopAttributeDto  `json:"top_attribute_all_time"`
}

type TopAttributeDto struct {
	QuestionText string  `json:"question_text"`
	Percentage   float64 `json:"percentage"`
}

type AchievementDto struct {
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	EarnedAt string                 `json:"earned_at"`
}

type SeasonCardDto struct {
	SeasonID     string  `json:"season_id"`
	SeasonNumber int32   `json:"season_number"`
	TopAttribute string  `json:"top_attribute"`
	Category     string  `json:"category"`
	Percentage   float64 `json:"percentage"`
}

func (s *Service) GetProfile(ctx context.Context, groupID, userID, requesterID string) (*ProfileResponse, error) {
	// Verify requester is a member of the group
	reqCount, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{UserID: requesterID, GroupID: groupID})
	if err != nil {
		return nil, err
	}
	if reqCount == 0 {
		return nil, ErrNotMember
	}

	// Verify target user exists
	user, err := s.queries.GetUserProfileInfo(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Verify target is a member
	count, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{UserID: userID, GroupID: groupID})
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrNotMember
	}

	// Fetch stats
	stats, err := s.queries.GetUserGroupStats(ctx, db.GetUserGroupStatsParams{UserID: userID, GroupID: groupID})
	statsFound := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Fetch top attribute all time
	var topAttr *TopAttributeDto
	topRow, err := s.queries.GetTopAttributeAllTime(ctx, db.GetTopAttributeAllTimeParams{TargetID: userID, GroupID: groupID})
	if err == nil {
		topAttr = &TopAttributeDto{
			QuestionText: topRow.QuestionText,
			Percentage:   math.Round(topRow.Percentage*10) / 10,
		}
	}

	// Fetch achievements
	achievements, err := s.queries.GetUserAchievements(ctx, db.GetUserAchievementsParams{UserID: userID, GroupID: groupID})
	if err != nil {
		return nil, err
	}

	achieveDtos := make([]AchievementDto, 0, len(achievements))
	for _, a := range achievements {
		dto := AchievementDto{
			Type:     string(a.AchievementType),
			EarnedAt: a.EarnedAt.Format("2006-01-02"),
		}
		if a.Metadata.Valid {
			var meta map[string]interface{}
			if json.Unmarshal(a.Metadata.RawMessage, &meta) == nil {
				dto.Metadata = meta
			}
		}
		achieveDtos = append(achieveDtos, dto)
	}

	// Fetch season history (last 5)
	history, err := s.queries.GetUserSeasonHistory(ctx, db.GetUserSeasonHistoryParams{
		TargetID: userID,
		GroupID:  groupID,
		Limit:   5,
	})
	if err != nil {
		return nil, err
	}

	seasonCards := make([]SeasonCardDto, 0, len(history))
	for _, h := range history {
		seasonCards = append(seasonCards, SeasonCardDto{
			SeasonID:     h.SeasonID,
			SeasonNumber: h.SeasonNumber,
			TopAttribute: h.QuestionText,
			Category:     string(h.QuestionCategory),
			Percentage:   math.Round(h.Percentage*10) / 10,
		})
	}

	statsDto := StatsDto{
		TopAttributeAllTime: topAttr,
	}
	if statsFound {
		statsDto.SeasonsPlayed = stats.SeasonsPlayed
		statsDto.VotingStreak = stats.VotingStreak
		statsDto.MaxVotingStreak = stats.MaxVotingStreak
		statsDto.GuessAccuracy = math.Round(stats.GuessAccuracy*10) / 10
		statsDto.TotalVotesCast = stats.TotalVotesCast
		statsDto.TotalVotesReceived = stats.TotalVotesReceived
	}

	legend := generateLegend(user.Username, statsDto, achieveDtos)

	var avatarEmoji, avatarUrl *string
	if user.AvatarEmoji.Valid {
		avatarEmoji = &user.AvatarEmoji.String
	}
	if user.AvatarUrl.Valid {
		avatarUrl = &user.AvatarUrl.String
	}

	return &ProfileResponse{
		User: UserDto{
			Username:    user.Username,
			AvatarEmoji: avatarEmoji,
			AvatarUrl:   avatarUrl,
		},
		Stats:         statsDto,
		Achievements:  achieveDtos,
		Legend:        legend,
		SeasonHistory: seasonCards,
	}, nil
}

func generateLegend(username string, stats StatsDto, achievements []AchievementDto) string {
	types := make(map[string]bool)
	for _, a := range achievements {
		types[a.Type] = true
	}

	var parts []string

	if types["LEGEND"] && types["SNIPER"] {
		parts = append(parts, fmt.Sprintf("%s — настоящий Снайпер и Легенда группы", username))
	} else if types["LEGEND"] {
		parts = append(parts, fmt.Sprintf("%s — Легенда группы, неизменный лидер", username))
	} else if types["TELEPATH"] {
		parts = append(parts, fmt.Sprintf("%s читает мысли участников", username))
	} else if types["ORACLE"] {
		parts = append(parts, fmt.Sprintf("%s — Оракул, который всегда прав", username))
	} else if types["SNIPER"] {
		parts = append(parts, fmt.Sprintf("%s — Снайпер, попадающий в цель", username))
	} else if types["MONOPOLIST"] {
		parts = append(parts, fmt.Sprintf("%s — Монополист, которого не перепутаешь", username))
	} else if types["STREAK_VOTER"] {
		parts = append(parts, fmt.Sprintf("%s не пропускает ни одного сезона", username))
	} else if types["NIGHT_OWL"] {
		parts = append(parts, fmt.Sprintf("%s — Ночная сова, голосует под звёздами", username))
	} else if types["RECRUITER"] {
		parts = append(parts, fmt.Sprintf("%s — душа компании, привёл друзей", username))
	} else if stats.SeasonsPlayed >= 5 {
		parts = append(parts, fmt.Sprintf("%s — опытный участник, %d сезонов за плечами", username, stats.SeasonsPlayed))
	} else if stats.SeasonsPlayed > 0 {
		parts = append(parts, fmt.Sprintf("%s начинает свой путь в группе", username))
	} else {
		parts = append(parts, fmt.Sprintf("%s — загадочная личность", username))
	}

	legend := strings.Join(parts, ". ")
	if len([]rune(legend)) > 150 {
		runes := []rune(legend)
		legend = string(runes[:147]) + "..."
	}
	return legend
}

