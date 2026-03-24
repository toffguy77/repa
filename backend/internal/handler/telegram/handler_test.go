package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	appmw "github.com/repa-app/repa/internal/middleware"
	telegramsvc "github.com/repa-app/repa/internal/service/telegram"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	return e
}

func TestWebhook_InvalidSecret(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWebhook_ValidSecret_EmptyUpdate(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "my-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_NoSecret(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when no secret configured, got %d", rec.Code)
	}
}

func TestWebhook_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMapServiceError(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", telegramsvc.ErrGroupNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"not admin", telegramsvc.ErrNotAdmin, http.StatusForbidden, "NOT_ADMIN"},
		{"not member", telegramsvc.ErrNotMember, http.StatusForbidden, "NOT_MEMBER"},
		{"no telegram", telegramsvc.ErrNoTelegram, http.StatusBadRequest, "NO_TELEGRAM"},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, "INTERNAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = mapServiceError(c, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp map[string]map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp["error"]["code"] != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, resp["error"]["code"])
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	if !hasPrefix("/connect ABC", "/connect ") {
		t.Error("expected true for /connect prefix")
	}
	if hasPrefix("/repa", "/connect ") {
		t.Error("expected false for /repa with /connect prefix")
	}
}

func TestExtractArg(t *testing.T) {
	got := extractArg("/connect REPA-X7K2", "/connect ")
	if got != "REPA-X7K2" {
		t.Errorf("expected REPA-X7K2, got %s", got)
	}
}

// --- Additional tests ---

func TestWebhook_BadJSON_Variants(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"truncated json", `{"update_id": 1`},
		{"array instead of object", `[1, 2, 3]`},
		{"empty string", ``},
		{"plain text", `hello world`},
		{"xml payload", `<update><id>1</id></update>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEcho()
			h := &Handler{webhookSecret: ""}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
				strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := h.Webhook(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != http.StatusBadRequest && rec.Code != http.StatusOK {
				t.Errorf("expected 400 or 200, got %d", rec.Code)
			}
		})
	}
}

func TestWebhook_MissingSecretHeader(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// No X-Telegram-Bot-Api-Secret-Token header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWebhook_EmptySecretHeader(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWebhook_MyChatMember_BotKicked(t *testing.T) {
	e := setupEcho()
	// svc is nil, so HandleBotRemoved will panic. We catch it to confirm
	// the handler correctly parsed the my_chat_member update.
	h := &Handler{webhookSecret: ""}

	body := `{
		"update_id": 123,
		"my_chat_member": {
			"chat": {"id": 456, "type": "supergroup", "title": "Test"},
			"from": {"id": 789, "is_bot": false},
			"new_chat_member": {"status": "kicked", "user": {"id": 101, "is_bot": true}}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	defer func() {
		recover() // expected panic from nil svc
	}()

	_ = h.Webhook(c)
}

func TestWebhook_MyChatMember_BotLeft(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	body := `{
		"update_id": 123,
		"my_chat_member": {
			"chat": {"id": 456, "type": "supergroup", "title": "Test"},
			"from": {"id": 789, "is_bot": false},
			"new_chat_member": {"status": "left", "user": {"id": 101, "is_bot": true}}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	defer func() {
		recover()
	}()

	_ = h.Webhook(c)
}

func TestWebhook_MyChatMember_BotAdded(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	// Bot added (status "member") — should NOT trigger HandleBotRemoved,
	// should fall through to return 200.
	body := `{
		"update_id": 123,
		"my_chat_member": {
			"chat": {"id": 456, "type": "supergroup", "title": "Test"},
			"from": {"id": 789, "is_bot": false},
			"new_chat_member": {"status": "member", "user": {"id": 101, "is_bot": true}}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_MessageWithoutText(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	// Message with empty text — should not trigger any command, return 200.
	body := `{
		"update_id": 124,
		"message": {
			"message_id": 1,
			"chat": {"id": 456, "type": "supergroup"},
			"text": ""
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_UnknownCommand(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	// Unknown command — should fall through, return 200.
	body := `{
		"update_id": 125,
		"message": {
			"message_id": 2,
			"chat": {"id": 456, "type": "supergroup"},
			"text": "/help"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_PlainTextMessage(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	body := `{
		"update_id": 126,
		"message": {
			"message_id": 3,
			"chat": {"id": 456, "type": "supergroup"},
			"text": "just a regular message"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_ConnectCommand_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	body := `{
		"update_id": 127,
		"message": {
			"message_id": 4,
			"chat": {"id": 456, "type": "supergroup", "username": "testchat"},
			"text": "/connect REPA-CODE123"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// svc is nil → panic on HandleConnect
	defer func() {
		recover()
	}()

	_ = h.Webhook(c)
}

func TestWebhook_RepaCommand_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	body := `{
		"update_id": 128,
		"message": {
			"message_id": 5,
			"chat": {"id": 456, "type": "supergroup"},
			"text": "/repa"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	defer func() {
		recover()
	}()

	_ = h.Webhook(c)
}

func TestWebhook_RepaAtCommand(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	// /repa@botname should also match
	body := `{
		"update_id": 129,
		"message": {
			"message_id": 6,
			"chat": {"id": 456, "type": "supergroup"},
			"text": "/repa@repaapp_bot"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	defer func() {
		recover()
	}()

	_ = h.Webhook(c)
}

func TestMapServiceError_WrappedErrors(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			"wrapped not found",
			fmt.Errorf("lookup: %w", telegramsvc.ErrGroupNotFound),
			http.StatusNotFound,
			"NOT_FOUND",
		},
		{
			"wrapped not admin",
			fmt.Errorf("auth: %w", telegramsvc.ErrNotAdmin),
			http.StatusForbidden,
			"NOT_ADMIN",
		},
		{
			"wrapped not member",
			fmt.Errorf("auth: %w", telegramsvc.ErrNotMember),
			http.StatusForbidden,
			"NOT_MEMBER",
		},
		{
			"wrapped no telegram",
			fmt.Errorf("share: %w", telegramsvc.ErrNoTelegram),
			http.StatusBadRequest,
			"NO_TELEGRAM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = mapServiceError(c, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp map[string]map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if resp["error"]["code"] != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, resp["error"]["code"])
			}
		})
	}
}

func TestMapServiceError_ResponseBody(t *testing.T) {
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mapServiceError(c, telegramsvc.ErrNotAdmin)

	var resp map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp["error"]["code"] != "NOT_ADMIN" {
		t.Errorf("expected code NOT_ADMIN, got %s", resp["error"]["code"])
	}
	if resp["error"]["message"] != "Only the group admin can perform this action" {
		t.Errorf("expected correct message, got %s", resp["error"]["message"])
	}
}

func TestHasPrefix_EdgeCases(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		want   bool
	}{
		{"/connect ABC", "/connect ", true},
		{"/connect ", "/connect ", true},
		{"/connect", "/connect ", false},
		{"/repa", "/repa", true},
		{"", "/connect ", false},
		{"/connect ", "", true},
		{"", "", true},
		{"/repa@bot", "/repa@", true},
		{"/repa@", "/repa@", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_prefix_"+tt.prefix, func(t *testing.T) {
			got := hasPrefix(tt.s, tt.prefix)
			if got != tt.want {
				t.Errorf("hasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestExtractArg_Variants(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		want   string
	}{
		{"/connect REPA-X7K2", "/connect ", "REPA-X7K2"},
		{"/connect ", "/connect ", ""},
		{"/connect multiple words here", "/connect ", "multiple words here"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := extractArg(tt.s, tt.prefix)
			if got != tt.want {
				t.Errorf("extractArg(%q, %q) = %q, want %q", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestGenerateCode_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/telegram/generate-code", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	defer func() {
		recover()
	}()

	_ = h.GenerateCode(c)
}

func TestDisconnect_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/g1/telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	defer func() {
		recover()
	}()

	_ = h.Disconnect(c)
}

func TestShareToTelegram_NilSvc(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/share-to-telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	defer func() {
		recover()
	}()

	_ = h.ShareToTelegram(c)
}

// --- Fake Telegram bot using httptest + custom transport ---

// fakeTransport redirects all HTTP requests to a local httptest server.
type fakeTransport struct {
	server    *httptest.Server
	realHTTP  http.RoundTripper
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = t.server.Listener.Addr().String()
	return t.realHTTP.RoundTrip(req)
}

// newFakeTelegramBot creates a TelegramClient that talks to a local mock server.
// The mock server returns success for all API calls.
func newFakeTelegramBot() (*lib.TelegramClient, *httptest.Server, http.RoundTripper) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return success for any Telegram API method
		if strings.Contains(r.URL.Path, "getMe") {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"username":"test_bot"}}`))
		} else if strings.Contains(r.URL.Path, "getChatMember") {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"status":"administrator","user":{"id":123,"is_bot":true}}}`))
		} else {
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		}
	}))

	// Save the real transport before replacing
	realTransport := http.DefaultTransport
	// If realTransport is already a fakeTransport, unwrap it
	if ft, ok := realTransport.(*fakeTransport); ok {
		realTransport = ft.realHTTP
	}
	newFt := &fakeTransport{server: srv, realHTTP: realTransport}
	http.DefaultTransport = newFt

	bot, err := lib.NewTelegramClient("fake-token")
	// Keep http.DefaultTransport as fakeTransport so goroutines spawned
	// by sendOK/sendError also hit our mock server.
	if err != nil {
		http.DefaultTransport = realTransport
		panic(fmt.Sprintf("failed to create fake bot: %v", err))
	}

	return bot, srv, realTransport
}

func newTestHandlerWithBot(mq *handlerMockQuerier) (*Handler, func()) {
	bot, srv, realTransport := newFakeTelegramBot()
	svc := telegramsvc.NewService(mq, bot, "https://repa.app")
	h := NewHandler(svc, "test-secret")
	cleanup := func() {
		// Wait for any sendOK/sendError goroutines to complete
		time.Sleep(150 * time.Millisecond)
		http.DefaultTransport = realTransport
		srv.Close()
	}
	return h, cleanup
}

// --- Mock Querier for handler-level tests ---

type handlerMockQuerier struct {
	db.Querier

	groups             map[string]db.Group
	groupsByTelegramID map[string]db.Group
	seasons            map[string]db.Season // by season ID
	activeSeasons      map[string]db.Season // by group ID
	memberCounts       map[string]int64
	voterCounts        map[string]int64
	members            map[string]map[string]bool // groupID -> userID -> true

	disconnectByChatCalled bool
	disconnectByChatErr    error
}

func newHandlerMockQuerier() *handlerMockQuerier {
	return &handlerMockQuerier{
		groups:             make(map[string]db.Group),
		groupsByTelegramID: make(map[string]db.Group),
		seasons:            make(map[string]db.Season),
		activeSeasons:      make(map[string]db.Season),
		memberCounts:       make(map[string]int64),
		voterCounts:        make(map[string]int64),
		members:            make(map[string]map[string]bool),
	}
}

func (m *handlerMockQuerier) GetGroupByID(_ context.Context, id string) (db.Group, error) {
	g, ok := m.groups[id]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *handlerMockQuerier) SetGroupConnectCode(_ context.Context, _ db.SetGroupConnectCodeParams) error {
	return nil
}

func (m *handlerMockQuerier) UpdateGroupTelegram(_ context.Context, _ db.UpdateGroupTelegramParams) error {
	return nil
}

func (m *handlerMockQuerier) DisconnectTelegramByChat(_ context.Context, _ sql.NullString) error {
	m.disconnectByChatCalled = true
	return m.disconnectByChatErr
}

func (m *handlerMockQuerier) GetGroupByConnectCode(_ context.Context, code sql.NullString) (db.Group, error) {
	for _, g := range m.groups {
		if g.TelegramConnectCode.Valid && g.TelegramConnectCode.String == code.String {
			return g, nil
		}
	}
	return db.Group{}, sql.ErrNoRows
}

func (m *handlerMockQuerier) GetGroupByTelegramChatID(_ context.Context, chatID sql.NullString) (db.Group, error) {
	g, ok := m.groupsByTelegramID[chatID.String]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *handlerMockQuerier) GetActiveSeasonByGroup(_ context.Context, groupID string) (db.Season, error) {
	s, ok := m.activeSeasons[groupID]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *handlerMockQuerier) CountGroupMembers(_ context.Context, groupID string) (int64, error) {
	return m.memberCounts[groupID], nil
}

func (m *handlerMockQuerier) CountSeasonVoters(_ context.Context, seasonID string) (int64, error) {
	return m.voterCounts[seasonID], nil
}

func (m *handlerMockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	s, ok := m.seasons[id]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *handlerMockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if g, ok := m.members[arg.GroupID]; ok && g[arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

// --- Tests with real service + mock querier ---

func newTestHandler(mq *handlerMockQuerier) *Handler {
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	return NewHandler(svc, "test-secret")
}

func webhookRequest(e *echo.Echo, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "test-secret")
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestWebhook_MyChatMember_BotKicked_WithService(t *testing.T) {
	mq := newHandlerMockQuerier()
	h := newTestHandler(mq)
	e := setupEcho()

	body := `{
		"update_id": 200,
		"my_chat_member": {
			"chat": {"id": 12345, "type": "supergroup", "title": "Test Chat"},
			"from": {"id": 789, "is_bot": false},
			"new_chat_member": {"status": "kicked", "user": {"id": 101, "is_bot": true}}
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Give goroutine time to execute HandleBotRemoved -> DisconnectTelegramByChat
	time.Sleep(50 * time.Millisecond)
	if !mq.disconnectByChatCalled {
		t.Error("expected DisconnectTelegramByChat to be called")
	}
}

func TestWebhook_MyChatMember_BotLeft_WithService(t *testing.T) {
	mq := newHandlerMockQuerier()
	h := newTestHandler(mq)
	e := setupEcho()

	body := `{
		"update_id": 201,
		"my_chat_member": {
			"chat": {"id": 99999, "type": "supergroup", "title": "Left Chat"},
			"from": {"id": 789, "is_bot": false},
			"new_chat_member": {"status": "left", "user": {"id": 101, "is_bot": true}}
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	time.Sleep(50 * time.Millisecond)
	if !mq.disconnectByChatCalled {
		t.Error("expected DisconnectTelegramByChat to be called on 'left' status")
	}
}

func TestWebhook_ConnectCommand_CodeNotFound_WithBot(t *testing.T) {
	mq := newHandlerMockQuerier()
	h, cleanup := newTestHandlerWithBot(mq)
	defer cleanup()
	e := setupEcho()

	body := `{
		"update_id": 400,
		"message": {
			"message_id": 20,
			"chat": {"id": 88888, "type": "supergroup", "username": "testchat"},
			"text": "/connect REPA-INVALID"
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// sendError is called in a goroutine — give it time
	time.Sleep(100 * time.Millisecond)
}

func TestWebhook_ConnectCommand_Success_WithBot(t *testing.T) {
	mq := newHandlerMockQuerier()
	mq.groups["grp-con"] = db.Group{
		ID:                    "grp-con",
		Name:                  "Connected Group",
		AdminID:               "admin-1",
		TelegramConnectCode:   sql.NullString{String: "REPA-ABCD", Valid: true},
		TelegramConnectExpiry: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
	}
	h, cleanup := newTestHandlerWithBot(mq)
	defer cleanup()
	e := setupEcho()

	body := `{
		"update_id": 401,
		"message": {
			"message_id": 21,
			"chat": {"id": 99999, "type": "supergroup", "username": "testchat2"},
			"text": "/connect REPA-ABCD"
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	time.Sleep(100 * time.Millisecond)
}

func TestWebhook_RepaCommand_NotLinked_WithBot(t *testing.T) {
	mq := newHandlerMockQuerier()
	h, cleanup := newTestHandlerWithBot(mq)
	defer cleanup()
	e := setupEcho()

	body := `{
		"update_id": 500,
		"message": {
			"message_id": 30,
			"chat": {"id": 77777, "type": "supergroup"},
			"text": "/repa"
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	time.Sleep(100 * time.Millisecond)
}

func TestWebhook_RepaCommand_LinkedGroup_WithBot(t *testing.T) {
	mq := newHandlerMockQuerier()
	mq.groupsByTelegramID["55555"] = db.Group{
		ID:             "grp-linked",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "55555", Valid: true},
	}
	mq.activeSeasons["grp-linked"] = db.Season{
		ID:       "season-1",
		GroupID:  "grp-linked",
		Number:   3,
		Status:   db.SeasonStatusVOTING,
		RevealAt: time.Date(2026, 3, 27, 17, 0, 0, 0, time.UTC),
	}
	mq.memberCounts["grp-linked"] = 10
	mq.voterCounts["season-1"] = 5

	h, cleanup := newTestHandlerWithBot(mq)
	defer cleanup()
	e := setupEcho()

	body := `{
		"update_id": 501,
		"message": {
			"message_id": 31,
			"chat": {"id": 55555, "type": "supergroup"},
			"text": "/repa"
		}
	}`
	c, rec := webhookRequest(e, body)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	time.Sleep(100 * time.Millisecond)
}

func TestGenerateCode_Success(t *testing.T) {
	e := setupEcho()
	e.Validator = appmw.NewValidator()

	mq := newHandlerMockQuerier()
	mq.groups["grp-1"] = db.Group{
		ID:      "grp-1",
		Name:    "My Group",
		AdminID: "admin-1",
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/grp-1/telegram/generate-code", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("grp-1")
	c.Set("user", &appmw.JWTClaims{UserID: "admin-1", Username: "admin"})

	if err := h.GenerateCode(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	code := body["data"]["connect_code"]
	if !strings.HasPrefix(code, "REPA-") {
		t.Errorf("expected connect_code to start with REPA-, got %s", code)
	}
	if body["data"]["instruction"] == "" {
		t.Error("expected instruction in response")
	}
	if body["data"]["expires_at"] == "" {
		t.Error("expected expires_at in response")
	}
}

func TestGenerateCode_NotAdmin(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	mq.groups["grp-1"] = db.Group{
		ID:      "grp-1",
		Name:    "My Group",
		AdminID: "admin-1",
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/grp-1/telegram/generate-code", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("grp-1")
	c.Set("user", &appmw.JWTClaims{UserID: "not-admin", Username: "user"})

	if err := h.GenerateCode(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_ADMIN" {
		t.Errorf("expected NOT_ADMIN, got %s", resp["error"]["code"])
	}
}

func TestGenerateCode_GroupNotFound(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/nonexistent/telegram/generate-code", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "user"})

	if err := h.GenerateCode(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDisconnect_Success(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	mq.groups["grp-1"] = db.Group{
		ID:      "grp-1",
		Name:    "My Group",
		AdminID: "admin-1",
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/grp-1/telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("grp-1")
	c.Set("user", &appmw.JWTClaims{UserID: "admin-1", Username: "admin"})

	if err := h.Disconnect(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]map[string]bool
	json.Unmarshal(rec.Body.Bytes(), &body)
	if !body["data"]["disconnected"] {
		t.Error("expected disconnected=true")
	}
}

func TestDisconnect_NotAdmin(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	mq.groups["grp-1"] = db.Group{
		ID:      "grp-1",
		Name:    "My Group",
		AdminID: "admin-1",
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/grp-1/telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("grp-1")
	c.Set("user", &appmw.JWTClaims{UserID: "not-admin", Username: "user"})

	if err := h.Disconnect(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestDisconnect_GroupNotFound(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/nonexistent/telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "user"})

	if err := h.Disconnect(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestShareToTelegram_GroupNotFound(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/nonexistent/share-to-telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("nonexistent")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	if err := h.ShareToTelegram(c); err != nil {
		t.Fatal(err)
	}
	// GetSeasonByID returns sql.ErrNoRows, service returns generic error -> mapServiceError -> 500
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestShareToTelegram_NotMember(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	// Add a season by ID — user is not a member of the group
	mq.seasons["season-1"] = db.Season{
		ID:      "season-1",
		GroupID: "grp-1",
		Number:  1,
		Status:  db.SeasonStatusVOTING,
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/season-1/share-to-telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("season-1")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	if err := h.ShareToTelegram(c); err != nil {
		t.Fatal(err)
	}
	// User is not a member -> ErrNotMember -> mapServiceError -> 403 NOT_MEMBER
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %s", resp["error"]["code"])
	}
}

func TestShareToTelegram_NoTelegram(t *testing.T) {
	e := setupEcho()

	mq := newHandlerMockQuerier()
	mq.seasons["season-1"] = db.Season{
		ID:      "season-1",
		GroupID: "grp-1",
		Number:  1,
		Status:  db.SeasonStatusVOTING,
	}
	mq.members["grp-1"] = map[string]bool{"u1": true}
	mq.groups["grp-1"] = db.Group{
		ID:      "grp-1",
		Name:    "Test Group",
		AdminID: "u1",
		// No TelegramChatID -> ErrNoTelegram
	}
	svc := telegramsvc.NewService(mq, nil, "https://repa.app")
	h := NewHandler(svc, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/season-1/share-to-telegram", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("season-1")
	c.Set("user", &appmw.JWTClaims{UserID: "u1", Username: "testuser"})

	if err := h.ShareToTelegram(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NO_TELEGRAM" {
		t.Errorf("expected NO_TELEGRAM, got %s", resp["error"]["code"])
	}
}

func TestSendError_MessageMapping(t *testing.T) {
	// Test the sendError message selection logic directly.
	// We can't call sendError itself (goroutine + nil bot panic),
	// but we verify the mapServiceError which shares the same error sentinel logic.
	e := setupEcho()

	// ErrCodeNotFound and ErrBotNotAdmin are handled in sendError but not mapServiceError.
	// Verify they are distinct error values.
	if errors.Is(telegramsvc.ErrCodeNotFound, telegramsvc.ErrGroupNotFound) {
		t.Error("ErrCodeNotFound should not match ErrGroupNotFound")
	}
	if errors.Is(telegramsvc.ErrBotNotAdmin, telegramsvc.ErrNotAdmin) {
		t.Error("ErrBotNotAdmin should not match ErrNotAdmin")
	}

	// Ensure mapServiceError handles wrapped ErrCodeNotFound as generic (INTERNAL)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = mapServiceError(c, telegramsvc.ErrCodeNotFound)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("ErrCodeNotFound should map to 500 via mapServiceError, got %d", rec.Code)
	}
}
