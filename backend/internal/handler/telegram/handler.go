package telegram

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	"github.com/repa-app/repa/internal/lib"
	appmw "github.com/repa-app/repa/internal/middleware"
	telegramsvc "github.com/repa-app/repa/internal/service/telegram"
)

type Handler struct {
	svc           *telegramsvc.Service
	webhookSecret string
}

func NewHandler(svc *telegramsvc.Service, webhookSecret string) *Handler {
	return &Handler{svc: svc, webhookSecret: webhookSecret}
}

// Webhook handles incoming Telegram updates.
// POST /api/v1/telegram/webhook (public, secret-token validated)
func (h *Handler) Webhook(c echo.Context) error {
	if h.webhookSecret != "" {
		token := c.Request().Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(h.webhookSecret)) != 1 {
			return c.NoContent(http.StatusUnauthorized)
		}
	}

	var update lib.TelegramUpdate
	if err := c.Bind(&update); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()

	// Bot removed from chat
	if update.MyChatMember != nil {
		if update.MyChatMember.NewChatMember.Status == "kicked" || update.MyChatMember.NewChatMember.Status == "left" {
			h.svc.HandleBotRemoved(ctx, update.MyChatMember.Chat.ID)
			return c.NoContent(http.StatusOK)
		}
	}

	// Text commands
	if update.Message != nil && update.Message.Text != "" {
		text := update.Message.Text
		chatID := update.Message.Chat.ID
		chatUsername := update.Message.Chat.Username

		switch {
		case hasPrefix(text, "/connect "):
			code := extractArg(text, "/connect ")
			groupName, err := h.svc.HandleConnect(ctx, chatID, chatUsername, code)
			if err != nil {
				h.sendError(chatID, err)
				return c.NoContent(http.StatusOK)
			}
			h.sendOK(chatID, "Группа «"+groupName+"» подключена к Репе!")
			return c.NoContent(http.StatusOK)

		case text == "/repa" || hasPrefix(text, "/repa@"):
			msg, err := h.svc.HandleRepaCommand(ctx, chatID)
			if err != nil {
				h.sendError(chatID, err)
				return c.NoContent(http.StatusOK)
			}
			h.sendOK(chatID, msg)
			return c.NoContent(http.StatusOK)

		}
	}

	return c.NoContent(http.StatusOK)
}

// GenerateCode generates a connect code for a group.
// POST /api/v1/groups/:id/telegram/generate-code
func (h *Handler) GenerateCode(c echo.Context) error {
	groupID := c.Param("id")
	claims := appmw.GetCurrentUser(c)

	code, expiresAt, err := h.svc.GenerateConnectCode(c.Request().Context(), claims.UserID, groupID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"connect_code": code,
			"instruction":  "Добавьте @repaapp_bot в чат и напишите /connect " + code,
			"expires_at":   expiresAt.Format(time.RFC3339),
		},
	})
}

// Disconnect unlinks Telegram from a group.
// DELETE /api/v1/groups/:id/telegram
func (h *Handler) Disconnect(c echo.Context) error {
	groupID := c.Param("id")
	claims := appmw.GetCurrentUser(c)

	if err := h.svc.DisconnectTelegram(c.Request().Context(), claims.UserID, groupID); err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"disconnected": true}})
}

// ShareToTelegram shares a user's card to the group's Telegram chat.
// POST /api/v1/seasons/:seasonId/share-to-telegram
func (h *Handler) ShareToTelegram(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	if err := h.svc.ShareCard(c.Request().Context(), claims.UserID, seasonID); err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"shared": true}})
}

// --- Helpers ---

func (h *Handler) sendOK(chatID int64, text string) {
	// Fire and forget — webhook must return quickly
	go func() {
		_ = h.svc.SendMessage(chatID, text)
	}()
}

func (h *Handler) sendError(chatID int64, err error) {
	msg := "Произошла ошибка"
	switch {
	case errors.Is(err, telegramsvc.ErrCodeNotFound):
		msg = "Код не найден или истёк"
	case errors.Is(err, telegramsvc.ErrBotNotAdmin):
		msg = "Сделайте бота администратором чата"
	}
	go func() {
		_ = h.svc.SendMessage(chatID, msg)
	}()
}

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, telegramsvc.ErrGroupNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Group not found")
	case errors.Is(err, telegramsvc.ErrNotAdmin):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_ADMIN", "Only the group admin can perform this action")
	case errors.Is(err, telegramsvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	case errors.Is(err, telegramsvc.ErrNoTelegram):
		return handler.ErrorResponse(c, http.StatusBadRequest, "NO_TELEGRAM", "Group has no linked Telegram chat")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func extractArg(s, prefix string) string {
	return s[len(prefix):]
}
