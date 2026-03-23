package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrNotAdmin      = errors.New("only admin can perform this action")
	ErrCodeNotFound  = errors.New("connect code not found or expired")
	ErrBotNotAdmin   = errors.New("bot must be an administrator of the chat")
	ErrNoTelegram    = errors.New("group has no linked telegram chat")
	ErrNotMember     = errors.New("user is not a member of this group")
)

type Service struct {
	queries  db.Querier
	bot      *lib.TelegramClient
	baseURL  string
}

func NewService(queries db.Querier, bot *lib.TelegramClient, baseURL string) *Service {
	return &Service{
		queries:  queries,
		bot:      bot,
		baseURL:  baseURL,
	}
}

// GenerateConnectCode creates a REPA-XXXX code for the group admin to use in Telegram.
func (s *Service) GenerateConnectCode(ctx context.Context, userID, groupID string) (string, time.Time, error) {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, ErrGroupNotFound
		}
		return "", time.Time{}, err
	}

	if group.AdminID != userID {
		return "", time.Time{}, ErrNotAdmin
	}

	code := fmt.Sprintf("REPA-%s", randomAlphaNum(4))
	expiry := time.Now().Add(24 * time.Hour)

	if err := s.queries.SetGroupConnectCode(ctx, db.SetGroupConnectCodeParams{
		ID:                    groupID,
		TelegramConnectCode:   sql.NullString{String: code, Valid: true},
		TelegramConnectExpiry: sql.NullTime{Time: expiry, Valid: true},
	}); err != nil {
		return "", time.Time{}, err
	}

	return code, expiry, nil
}

// DisconnectTelegram unlinks the Telegram chat from the group.
func (s *Service) DisconnectTelegram(ctx context.Context, userID, groupID string) error {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}

	if group.AdminID != userID {
		return ErrNotAdmin
	}

	return s.queries.UpdateGroupTelegram(ctx, db.UpdateGroupTelegramParams{
		ID:                   groupID,
		TelegramChatID:       sql.NullString{},
		TelegramChatUsername: sql.NullString{},
	})
}

// HandleConnect processes /connect CODE from a Telegram chat.
func (s *Service) HandleConnect(ctx context.Context, chatID int64, chatUsername, code string) (string, error) {
	group, err := s.queries.GetGroupByConnectCode(ctx, sql.NullString{String: code, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCodeNotFound
		}
		return "", err
	}

	// Check bot is admin in the chat
	member, err := s.bot.GetChatMember(ctx, chatID, s.bot.BotID())
	if err != nil {
		return "", fmt.Errorf("check bot status: %w", err)
	}
	if member.Status != "administrator" && member.Status != "creator" {
		return "", ErrBotNotAdmin
	}

	if err := s.queries.UpdateGroupTelegram(ctx, db.UpdateGroupTelegramParams{
		ID:                   group.ID,
		TelegramChatID:       sql.NullString{String: strconv.FormatInt(chatID, 10), Valid: true},
		TelegramChatUsername: sql.NullString{String: chatUsername, Valid: chatUsername != ""},
	}); err != nil {
		return "", err
	}

	return group.Name, nil
}

// HandleDisconnect processes /disconnect from a Telegram chat.
func (s *Service) HandleDisconnect(ctx context.Context, chatID int64) error {
	return s.queries.DisconnectTelegramByChat(ctx, sql.NullString{
		String: strconv.FormatInt(chatID, 10),
		Valid:  true,
	})
}

// HandleBotRemoved auto-unlinks when the bot is removed from a chat.
func (s *Service) HandleBotRemoved(ctx context.Context, chatID int64) {
	if err := s.HandleDisconnect(ctx, chatID); err != nil {
		log.Warn().Err(err).Int64("chat_id", chatID).Msg("failed to unlink on bot removal")
	}
}

// HandleRepaCommand returns current season status for /repa command.
func (s *Service) HandleRepaCommand(ctx context.Context, chatID int64) (string, error) {
	chatIDStr := strconv.FormatInt(chatID, 10)
	group, err := s.queries.GetGroupByTelegramChatID(ctx, sql.NullString{String: chatIDStr, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "Этот чат не привязан к группе в Репе.", nil
		}
		return "", err
	}

	season, err := s.queries.GetActiveSeasonByGroup(ctx, group.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Sprintf("Группа «%s» — сейчас нет активного сезона.", group.Name), nil
		}
		return "", err
	}

	memberCount, err := s.queries.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return "", err
	}

	voterCount, err := s.queries.CountSeasonVoters(ctx, season.ID)
	if err != nil {
		return "", err
	}

	msk := time.FixedZone("MSK", 3*60*60)
	revealTime := season.RevealAt.In(msk).Format("02.01 в 15:04")

	return fmt.Sprintf(
		"Группа «%s» — Сезон %d\nСтатус: %s\nПроголосовало: %d / %d\nReveal: %s МСК",
		group.Name, season.Number, season.Status, voterCount, memberCount, revealTime,
	), nil
}

// PostSeasonStartAll posts season-start to all groups with active voting seasons.
func (s *Service) PostSeasonStartAll(ctx context.Context) error {
	seasons, err := s.queries.GetAllVotingSeasons(ctx)
	if err != nil {
		return fmt.Errorf("get voting seasons: %w", err)
	}
	for _, season := range seasons {
		if err := s.PostSeasonStart(ctx, season.GroupID); err != nil {
			log.Warn().Err(err).Str("group_id", season.GroupID).Msg("telegram season-start post failed")
		}
	}
	return nil
}

// PostSeasonStart sends a "new season" message to the group's Telegram chat.
func (s *Service) PostSeasonStart(ctx context.Context, groupID string) error {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !group.TelegramChatID.Valid {
		return nil
	}

	chatID, err := strconv.ParseInt(group.TelegramChatID.String, 10, 64)
	if err != nil {
		return fmt.Errorf("parse chat_id: %w", err)
	}

	text := fmt.Sprintf("Новый сезон в группе «%s»!\n\nГолосуй в приложении", group.Name)
	markup := &lib.InlineKeyboardMarkup{
		InlineKeyboard: [][]lib.InlineKeyboardButton{{
			{Text: "Проголосовать", URL: fmt.Sprintf("%s/group/%s", s.baseURL, groupID)},
		}},
	}

	return s.bot.SendMessage(ctx, chatID, text, markup)
}

// PostReveal sends reveal results summary to the group's Telegram chat.
func (s *Service) PostReveal(ctx context.Context, seasonID string) error {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
	}

	group, err := s.queries.GetGroupByID(ctx, season.GroupID)
	if err != nil {
		return err
	}
	if !group.TelegramChatID.Valid {
		return nil
	}

	chatID, err := strconv.ParseInt(group.TelegramChatID.String, 10, 64)
	if err != nil {
		return fmt.Errorf("parse chat_id: %w", err)
	}

	topResults, err := s.queries.GetTopResultPerQuestion(ctx, seasonID)
	if err != nil {
		return err
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Репа подвела итоги — %s\n", group.Name))
	lines = append(lines, "Топ атрибуты этой недели:")

	limit := min(5, len(topResults))
	for _, r := range topResults[:limit] {
		lines = append(lines, fmt.Sprintf("  «%s» — %s (%d%%)", r.QuestionText, r.Username, int(r.Percentage)))
	}

	text := strings.Join(lines, "\n")

	markup := &lib.InlineKeyboardMarkup{
		InlineKeyboard: [][]lib.InlineKeyboardButton{{
			{Text: "Открыть свою репу", URL: fmt.Sprintf("%s/group/%s/reveal", s.baseURL, group.ID)},
			{Text: "Посмотреть всех", URL: fmt.Sprintf("%s/group/%s/reveal/members", s.baseURL, group.ID)},
		}},
	}

	return s.bot.SendMessage(ctx, chatID, text, markup)
}

// ShareCard shares a user's card to the group's Telegram chat.
func (s *Service) ShareCard(ctx context.Context, userID, seasonID string) error {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
	}

	// Verify user is a member of the group
	isMember, err := s.queries.IsGroupMember(ctx, db.IsGroupMemberParams{
		UserID:  userID,
		GroupID: season.GroupID,
	})
	if err != nil {
		return err
	}
	if isMember == 0 {
		return ErrNotMember
	}

	group, err := s.queries.GetGroupByID(ctx, season.GroupID)
	if err != nil {
		return err
	}
	if !group.TelegramChatID.Valid {
		return ErrNoTelegram
	}

	chatID, err := strconv.ParseInt(group.TelegramChatID.String, 10, 64)
	if err != nil {
		return fmt.Errorf("parse chat_id: %w", err)
	}

	card, err := s.queries.GetCardCache(ctx, db.GetCardCacheParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err != nil {
		return fmt.Errorf("get card cache: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	caption := fmt.Sprintf("@%s поделился своей репой", user.Username)
	return s.bot.SendPhoto(ctx, chatID, card.ImageUrl, caption)
}

// SendMessage sends a plain text message to a Telegram chat (fire-and-forget from webhook).
func (s *Service) SendMessage(chatID int64, text string) error {
	return s.bot.SendMessage(context.Background(), chatID, text, nil)
}

func randomAlphaNum(n int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
