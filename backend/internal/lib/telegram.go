package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const telegramBaseURL = "https://api.telegram.org/bot"

type TelegramClient struct {
	token  string
	botID  int64
	client *http.Client
}

func NewTelegramClient(token string) (*TelegramClient, error) {
	tc := &TelegramClient{
		token: token,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	me, err := tc.GetMe(context.Background())
	if err != nil {
		return nil, fmt.Errorf("telegram getMe: %w", err)
	}
	tc.botID = me.ID

	return tc, nil
}

// Telegram API types

type TelegramUser struct {
	ID       int64  `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

type TelegramChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      TelegramChat  `json:"chat"`
	Text      string        `json:"text"`
}

type TelegramChatMember struct {
	Status string        `json:"status"`
	User   *TelegramUser `json:"user"`
}

type TelegramChatMemberUpdated struct {
	Chat          TelegramChat       `json:"chat"`
	From          TelegramUser       `json:"from"`
	NewChatMember TelegramChatMember `json:"new_chat_member"`
}

type TelegramUpdate struct {
	UpdateID     int64                      `json:"update_id"`
	Message      *TelegramMessage           `json:"message"`
	MyChatMember *TelegramChatMemberUpdated `json:"my_chat_member"`
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type telegramResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
}

// API methods

func (tc *TelegramClient) GetMe(ctx context.Context) (*TelegramUser, error) {
	var user TelegramUser
	if err := tc.apiCall(ctx, "getMe", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (tc *TelegramClient) BotID() int64 {
	return tc.botID
}

func (tc *TelegramClient) GetChatMember(ctx context.Context, chatID int64, userID int64) (*TelegramChatMember, error) {
	params := map[string]any{
		"chat_id": chatID,
		"user_id": userID,
	}
	var member TelegramChatMember
	if err := tc.apiCall(ctx, "getChatMember", params, &member); err != nil {
		return nil, err
	}
	return &member, nil
}

func (tc *TelegramClient) SendMessage(ctx context.Context, chatID int64, text string, markup *InlineKeyboardMarkup) error {
	params := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if markup != nil {
		params["reply_markup"] = markup
	}
	return tc.apiCall(ctx, "sendMessage", params, nil)
}

func (tc *TelegramClient) SendPhoto(ctx context.Context, chatID int64, photoURL string, caption string) error {
	params := map[string]any{
		"chat_id":    chatID,
		"photo":      photoURL,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	return tc.apiCall(ctx, "sendPhoto", params, nil)
}

func (tc *TelegramClient) apiCall(ctx context.Context, method string, params map[string]any, result any) error {
	url := telegramBaseURL + tc.token + "/" + method

	var body io.Reader
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var tgResp telegramResponse
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if !tgResp.OK {
		return fmt.Errorf("telegram API error: %s", tgResp.Description)
	}

	if result != nil {
		if err := json.Unmarshal(tgResp.Result, result); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}

	return nil
}
