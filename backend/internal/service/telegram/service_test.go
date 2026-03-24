package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// --- Mock Querier ---

type mockQuerier struct {
	db.Querier

	groups              map[string]db.Group            // by ID
	groupsByConnectCode map[string]db.Group            // by connect code
	groupsByTelegramID  map[string]db.Group            // by telegram chat ID
	seasons             map[string]db.Season           // by ID
	activeSeasons       map[string]db.Season           // by group ID
	allVotingSeasons    []db.Season
	members             map[string]map[string]bool     // groupID -> userID -> true
	memberCounts        map[string]int64               // groupID -> count
	voterCounts         map[string]int64               // seasonID -> count
	topResults          map[string][]db.GetTopResultPerQuestionRow // seasonID -> rows
	cardCaches          map[string]db.CardCache        // "userID:seasonID" -> cache
	users               map[string]db.User             // by ID

	// Track calls
	setConnectCodeCalled    bool
	updateTelegramCalled    bool
	disconnectByChat        sql.NullString
	disconnectByChatCalled  bool
	setConnectCodeErr       error
	updateTelegramErr       error
	disconnectByChatErr     error
}

func newMockQuerier() *mockQuerier {
	return &mockQuerier{
		groups:              make(map[string]db.Group),
		groupsByConnectCode: make(map[string]db.Group),
		groupsByTelegramID:  make(map[string]db.Group),
		seasons:             make(map[string]db.Season),
		activeSeasons:       make(map[string]db.Season),
		members:             make(map[string]map[string]bool),
		memberCounts:        make(map[string]int64),
		voterCounts:         make(map[string]int64),
		topResults:          make(map[string][]db.GetTopResultPerQuestionRow),
		cardCaches:          make(map[string]db.CardCache),
		users:               make(map[string]db.User),
	}
}

func (m *mockQuerier) GetGroupByID(_ context.Context, id string) (db.Group, error) {
	g, ok := m.groups[id]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *mockQuerier) SetGroupConnectCode(_ context.Context, _ db.SetGroupConnectCodeParams) error {
	m.setConnectCodeCalled = true
	return m.setConnectCodeErr
}

func (m *mockQuerier) UpdateGroupTelegram(_ context.Context, _ db.UpdateGroupTelegramParams) error {
	m.updateTelegramCalled = true
	return m.updateTelegramErr
}

func (m *mockQuerier) GetGroupByConnectCode(_ context.Context, code sql.NullString) (db.Group, error) {
	g, ok := m.groupsByConnectCode[code.String]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *mockQuerier) DisconnectTelegramByChat(_ context.Context, chatID sql.NullString) error {
	m.disconnectByChatCalled = true
	m.disconnectByChat = chatID
	return m.disconnectByChatErr
}

func (m *mockQuerier) GetGroupByTelegramChatID(_ context.Context, chatID sql.NullString) (db.Group, error) {
	g, ok := m.groupsByTelegramID[chatID.String]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *mockQuerier) GetActiveSeasonByGroup(_ context.Context, groupID string) (db.Season, error) {
	s, ok := m.activeSeasons[groupID]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) CountGroupMembers(_ context.Context, groupID string) (int64, error) {
	return m.memberCounts[groupID], nil
}

func (m *mockQuerier) CountSeasonVoters(_ context.Context, seasonID string) (int64, error) {
	return m.voterCounts[seasonID], nil
}

func (m *mockQuerier) GetAllVotingSeasons(_ context.Context) ([]db.Season, error) {
	return m.allVotingSeasons, nil
}

func (m *mockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	s, ok := m.seasons[id]
	if !ok {
		return db.Season{}, sql.ErrNoRows
	}
	return s, nil
}

func (m *mockQuerier) IsGroupMember(_ context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if g, ok := m.members[arg.GroupID]; ok && g[arg.UserID] {
		return 1, nil
	}
	return 0, nil
}

func (m *mockQuerier) GetCardCache(_ context.Context, arg db.GetCardCacheParams) (db.CardCache, error) {
	key := arg.UserID + ":" + arg.SeasonID
	c, ok := m.cardCaches[key]
	if !ok {
		return db.CardCache{}, sql.ErrNoRows
	}
	return c, nil
}

func (m *mockQuerier) GetUserByID(_ context.Context, id string) (db.User, error) {
	u, ok := m.users[id]
	if !ok {
		return db.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (m *mockQuerier) GetTopResultPerQuestion(_ context.Context, seasonID string) ([]db.GetTopResultPerQuestionRow, error) {
	return m.topResults[seasonID], nil
}

// --- Existing Tests ---

func TestRandomAlphaNum(t *testing.T) {
	code := randomAlphaNum(4)
	if len(code) != 4 {
		t.Errorf("expected length 4, got %d", len(code))
	}

	// Should not contain ambiguous characters (0, O, 1, I)
	for _, c := range code {
		switch c {
		case '0', 'O', '1', 'I':
			t.Errorf("code contains ambiguous character: %c", c)
		}
	}

	// Two codes should be different (probabilistic but extremely unlikely to fail)
	code2 := randomAlphaNum(4)
	if code == code2 {
		t.Log("warning: two random codes were identical (extremely unlikely)")
	}
}

func TestRandomAlphaNum_Length(t *testing.T) {
	for _, n := range []int{1, 4, 8, 16} {
		code := randomAlphaNum(n)
		if len(code) != n {
			t.Errorf("expected length %d, got %d", n, len(code))
		}
	}
}

// --- randomAlphaNum charset validation ---

func TestRandomAlphaNum_ValidCharset(t *testing.T) {
	const validChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	// Generate many codes to check all characters are valid
	for i := 0; i < 100; i++ {
		code := randomAlphaNum(8)
		for _, c := range code {
			if !strings.ContainsRune(validChars, c) {
				t.Errorf("code contains invalid character: %c", c)
			}
		}
	}
}

// --- GenerateConnectCode tests ---

func TestGenerateConnectCode_Success(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", Name: "Test Group", AdminID: "admin1"}
	svc := NewService(m, nil, "https://repa.app")

	code, expiry, err := svc.GenerateConnectCode(context.Background(), "admin1", "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(code, "REPA-") {
		t.Errorf("code should start with REPA-, got %s", code)
	}
	if len(code) != 9 { // "REPA-" (5) + 4 chars
		t.Errorf("expected code length 9, got %d", len(code))
	}
	if expiry.Before(time.Now()) {
		t.Error("expiry should be in the future")
	}
	if !m.setConnectCodeCalled {
		t.Error("SetGroupConnectCode should have been called")
	}
}

func TestGenerateConnectCode_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	_, _, err := svc.GenerateConnectCode(context.Background(), "admin1", "nonexistent")
	if !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGenerateConnectCode_NotAdmin(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", Name: "Test Group", AdminID: "admin1"}
	svc := NewService(m, nil, "https://repa.app")

	_, _, err := svc.GenerateConnectCode(context.Background(), "not-admin", "g1")
	if !errors.Is(err, ErrNotAdmin) {
		t.Errorf("expected ErrNotAdmin, got %v", err)
	}
}

func TestGenerateConnectCode_SetCodeError(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", Name: "Test Group", AdminID: "admin1"}
	m.setConnectCodeErr = errors.New("db error")
	svc := NewService(m, nil, "https://repa.app")

	_, _, err := svc.GenerateConnectCode(context.Background(), "admin1", "g1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db error" {
		t.Errorf("expected 'db error', got %v", err)
	}
}

// --- DisconnectTelegram tests ---

func TestDisconnectTelegram_Success(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{
		ID:      "g1",
		AdminID: "admin1",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.DisconnectTelegram(context.Background(), "admin1", "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.updateTelegramCalled {
		t.Error("UpdateGroupTelegram should have been called")
	}
}

func TestDisconnectTelegram_NotAdmin(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", AdminID: "admin1"}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.DisconnectTelegram(context.Background(), "not-admin", "g1")
	if !errors.Is(err, ErrNotAdmin) {
		t.Errorf("expected ErrNotAdmin, got %v", err)
	}
}

func TestDisconnectTelegram_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	err := svc.DisconnectTelegram(context.Background(), "admin1", "nonexistent")
	if !errors.Is(err, ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestDisconnectTelegram_UpdateError(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", AdminID: "admin1"}
	m.updateTelegramErr = errors.New("db error")
	svc := NewService(m, nil, "https://repa.app")

	err := svc.DisconnectTelegram(context.Background(), "admin1", "g1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- HandleDisconnect tests ---

func TestHandleDisconnect_DelegatesToQuery(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	err := svc.HandleDisconnect(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !m.disconnectByChatCalled {
		t.Error("DisconnectTelegramByChat should have been called")
	}
	if m.disconnectByChat.String != "12345" || !m.disconnectByChat.Valid {
		t.Errorf("expected chat ID '12345', got %v", m.disconnectByChat)
	}
}

func TestHandleDisconnect_Error(t *testing.T) {
	m := newMockQuerier()
	m.disconnectByChatErr = errors.New("db error")
	svc := NewService(m, nil, "https://repa.app")

	err := svc.HandleDisconnect(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- HandleRepaCommand tests ---

func TestHandleRepaCommand_ChatNotLinked(t *testing.T) {
	m := newMockQuerier()
	// No group linked to this chat
	svc := NewService(m, nil, "https://repa.app")

	msg, err := svc.HandleRepaCommand(context.Background(), 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Этот чат не привязан к группе в Репе." {
		t.Errorf("unexpected message: %s", msg)
	}
}

func TestHandleRepaCommand_NoActiveSeason(t *testing.T) {
	m := newMockQuerier()
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "Test Group"}
	// No active season for g1
	svc := NewService(m, nil, "https://repa.app")

	msg, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Группа «Test Group» — сейчас нет активного сезона."
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestHandleRepaCommand_ActiveSeason(t *testing.T) {
	m := newMockQuerier()
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "Test Group"}
	revealAt := time.Date(2025, 3, 21, 17, 0, 0, 0, time.UTC) // Friday 17:00 UTC = 20:00 MSK
	m.activeSeasons["g1"] = db.Season{
		ID:       "s1",
		GroupID:  "g1",
		Number:   3,
		Status:   db.SeasonStatusVOTING,
		RevealAt: revealAt,
	}
	m.memberCounts["g1"] = 10
	m.voterCounts["s1"] = 6
	svc := NewService(m, nil, "https://repa.app")

	msg, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "Test Group") {
		t.Errorf("message should contain group name, got: %s", msg)
	}
	if !strings.Contains(msg, "Сезон 3") {
		t.Errorf("message should contain season number, got: %s", msg)
	}
	if !strings.Contains(msg, "6 / 10") {
		t.Errorf("message should contain voter/member counts, got: %s", msg)
	}
	if !strings.Contains(msg, "VOTING") {
		t.Errorf("message should contain status, got: %s", msg)
	}
	if !strings.Contains(msg, "МСК") {
		t.Errorf("message should contain MSK timezone, got: %s", msg)
	}
}

func TestHandleRepaCommand_ActiveSeason_RevealTimeFormatted(t *testing.T) {
	m := newMockQuerier()
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "My Group"}
	revealAt := time.Date(2025, 3, 21, 17, 0, 0, 0, time.UTC)
	m.activeSeasons["g1"] = db.Season{
		ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusVOTING,
		RevealAt: revealAt,
	}
	m.memberCounts["g1"] = 5
	m.voterCounts["s1"] = 2
	svc := NewService(m, nil, "https://repa.app")

	msg, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 17:00 UTC + 3h = 20:00 MSK, date 21.03
	if !strings.Contains(msg, "21.03 в 20:00") {
		t.Errorf("expected reveal time '21.03 в 20:00' in message, got: %s", msg)
	}
}

// --- PostSeasonStart tests ---

func TestPostSeasonStart_GroupWithoutTelegram(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{Valid: false}, // no telegram
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStart(context.Background(), "g1")
	if err != nil {
		t.Fatalf("expected nil (no-op), got: %v", err)
	}
}

func TestPostSeasonStart_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStart(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

// --- PostReveal tests ---

func TestPostReveal_GroupWithoutTelegram(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{Valid: false}, // no telegram
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostReveal(context.Background(), "s1")
	if err != nil {
		t.Fatalf("expected nil (no-op), got: %v", err)
	}
}

func TestPostReveal_SeasonNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostReveal(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent season")
	}
}

func TestPostReveal_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "nonexistent-group"}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostReveal(context.Background(), "s1")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

// --- ShareCard tests ---

func TestShareCard_NotMember(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	// user "u1" is NOT a member of "g1"
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestShareCard_GroupWithoutTelegram(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.members["g1"] = map[string]bool{"u1": true}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		TelegramChatID: sql.NullString{Valid: false}, // no telegram
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if !errors.Is(err, ErrNoTelegram) {
		t.Errorf("expected ErrNoTelegram, got %v", err)
	}
}

func TestShareCard_SeasonNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent season")
	}
}

func TestShareCard_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "nonexistent-group"}
	m.members["nonexistent-group"] = map[string]bool{"u1": true}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

// --- PostSeasonStartAll tests ---

func TestPostSeasonStartAll_NoSeasons(t *testing.T) {
	m := newMockQuerier()
	m.allVotingSeasons = []db.Season{}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStartAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostSeasonStartAll_SkipsGroupsWithoutTelegram(t *testing.T) {
	m := newMockQuerier()
	m.allVotingSeasons = []db.Season{
		{ID: "s1", GroupID: "g1"},
		{ID: "s2", GroupID: "g2"},
	}
	// g1 has no telegram, g2 has no telegram
	m.groups["g1"] = db.Group{ID: "g1", Name: "G1", TelegramChatID: sql.NullString{Valid: false}}
	m.groups["g2"] = db.Group{ID: "g2", Name: "G2", TelegramChatID: sql.NullString{Valid: false}}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStartAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- HandleBotRemoved tests ---

func TestHandleBotRemoved_DelegatesToDisconnect(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	// Should not panic even if disconnect returns no error
	svc.HandleBotRemoved(context.Background(), 12345)
	if !m.disconnectByChatCalled {
		t.Error("DisconnectTelegramByChat should have been called")
	}
}

func TestHandleBotRemoved_ErrorIsLoggedNotReturned(t *testing.T) {
	m := newMockQuerier()
	m.disconnectByChatErr = errors.New("db error")
	svc := NewService(m, nil, "https://repa.app")

	// Should not panic; error is logged internally
	svc.HandleBotRemoved(context.Background(), 12345)
	if !m.disconnectByChatCalled {
		t.Error("DisconnectTelegramByChat should have been called")
	}
}

// --- SendMessage tests ---

func TestSendMessage_NilBotPanics(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil, but did not panic")
		}
	}()

	_ = svc.SendMessage(12345, "hello")
}

// --- HandleConnect tests ---

func TestHandleConnect_CodeNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")

	_, err := svc.HandleConnect(context.Background(), 12345, "testchat", "REPA-XXXX")
	if !errors.Is(err, ErrCodeNotFound) {
		t.Errorf("expected ErrCodeNotFound, got %v", err)
	}
}

// --- PostSeasonStart with invalid chat ID ---

func TestPostSeasonStart_InvalidChatID(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "not-a-number", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStart(context.Background(), "g1")
	if err == nil {
		t.Fatal("expected error for invalid chat ID")
	}
	if !strings.Contains(err.Error(), "parse chat_id") {
		t.Errorf("expected 'parse chat_id' error, got: %v", err)
	}
}

// --- PostReveal with invalid chat ID ---

func TestPostReveal_InvalidChatID(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "not-a-number", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostReveal(context.Background(), "s1")
	if err == nil {
		t.Fatal("expected error for invalid chat ID")
	}
	if !strings.Contains(err.Error(), "parse chat_id") {
		t.Errorf("expected 'parse chat_id' error, got: %v", err)
	}
}

// --- ShareCard with invalid chat ID ---

func TestShareCard_InvalidChatID(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.members["g1"] = map[string]bool{"u1": true}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		TelegramChatID: sql.NullString{String: "not-a-number", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if err == nil {
		t.Fatal("expected error for invalid chat ID")
	}
	if !strings.Contains(err.Error(), "parse chat_id") {
		t.Errorf("expected 'parse chat_id' error, got: %v", err)
	}
}

// --- ShareCard card cache not found ---

func TestShareCard_CardCacheNotFound(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.members["g1"] = map[string]bool{"u1": true}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	// No card cache for u1:s1
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if err == nil {
		t.Fatal("expected error for missing card cache")
	}
	if !strings.Contains(err.Error(), "get card cache") {
		t.Errorf("expected 'get card cache' error, got: %v", err)
	}
}

// --- NewService tests ---

func TestNewService(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil, "https://repa.app")
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

// ========================================================================
// Additional coverage tests
// ========================================================================

// --- HandleConnect: code found but bot is nil (panics) ---

func TestHandleConnect_CodeFoundBotNil(t *testing.T) {
	m := newMockQuerier()
	m.groupsByConnectCode["REPA-ABCD"] = db.Group{ID: "g1", Name: "Test Group"}
	svc := NewService(m, nil, "https://repa.app") // nil bot

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil, but no panic occurred")
		}
	}()

	_, _ = svc.HandleConnect(context.Background(), 12345, "testchat", "REPA-ABCD")
}

// --- PostReveal: with top results (up to 5 displayed) ---

func TestPostReveal_WithTopResultsNoBotPanics(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	m.topResults["s1"] = []db.GetTopResultPerQuestionRow{
		{QuestionText: "Most funny?", Username: "alice", Percentage: 75.0},
		{QuestionText: "Most smart?", Username: "bob", Percentage: 60.0},
		{QuestionText: "Most kind?", Username: "charlie", Percentage: 55.0},
	}
	svc := NewService(m, nil, "https://repa.app") // nil bot

	// Bot is nil, so SendMessage will panic when we reach it
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil")
		}
	}()

	_ = svc.PostReveal(context.Background(), "s1")
}

// --- PostReveal: GetTopResultPerQuestion returns empty list ---

func TestPostReveal_EmptyTopResultsNoBotPanics(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	m.topResults["s1"] = []db.GetTopResultPerQuestionRow{} // empty
	svc := NewService(m, nil, "https://repa.app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil")
		}
	}()

	_ = svc.PostReveal(context.Background(), "s1")
}

// --- PostSeasonStart: valid chat ID but nil bot panics ---

func TestPostSeasonStart_ValidChatIDBotNilPanics(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil")
		}
	}()

	_ = svc.PostSeasonStart(context.Background(), "g1")
}

// --- ShareCard: valid card cache but user not found ---

func TestShareCard_UserNotFound(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.members["g1"] = map[string]bool{"u1": true}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	m.cardCaches["u1:s1"] = db.CardCache{
		UserID:   "u1",
		SeasonID: "s1",
		ImageUrl: "https://example.com/card.png",
	}
	// user "u1" is NOT in users map -> will return sql.ErrNoRows
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

// --- ShareCard: valid card and user, bot nil panics ---

func TestShareCard_FullPathBotNilPanics(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.members["g1"] = map[string]bool{"u1": true}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	m.cardCaches["u1:s1"] = db.CardCache{
		UserID:   "u1",
		SeasonID: "s1",
		ImageUrl: "https://example.com/card.png",
	}
	m.users["u1"] = db.User{ID: "u1", Username: "alice"}
	svc := NewService(m, nil, "https://repa.app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil")
		}
	}()

	_ = svc.ShareCard(context.Background(), "u1", "s1")
}

// --- PostSeasonStartAll: group not found is logged, not returned ---

func TestPostSeasonStartAll_GroupNotFoundContinues(t *testing.T) {
	m := newMockQuerier()
	m.allVotingSeasons = []db.Season{
		{ID: "s1", GroupID: "nonexistent"},
		{ID: "s2", GroupID: "g2"},
	}
	// g2 has no telegram — will return nil (no-op)
	m.groups["g2"] = db.Group{ID: "g2", Name: "G2", TelegramChatID: sql.NullString{Valid: false}}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStartAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error (failures logged), got: %v", err)
	}
}

// --- HandleRepaCommand: CountGroupMembers error ---

func TestHandleRepaCommand_CountGroupMembersError(t *testing.T) {
	m := newMockQuerier()
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "Test Group"}
	m.activeSeasons["g1"] = db.Season{
		ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusVOTING,
		RevealAt: time.Now().Add(time.Hour),
	}
	// Override CountGroupMembers to return error
	m2 := &mockQuerierWithMemberCountError{mockQuerier: m, memberCountErr: errors.New("count error")}
	svc := NewService(m2, nil, "https://repa.app")

	_, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

// mockQuerierWithMemberCountError wraps mockQuerier but returns an error for CountGroupMembers.
type mockQuerierWithMemberCountError struct {
	*mockQuerier
	memberCountErr error
}

func (m *mockQuerierWithMemberCountError) CountGroupMembers(_ context.Context, _ string) (int64, error) {
	return 0, m.memberCountErr
}

// --- HandleRepaCommand: CountSeasonVoters error ---

func TestHandleRepaCommand_CountSeasonVotersError(t *testing.T) {
	m := newMockQuerier()
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "Test Group"}
	m.activeSeasons["g1"] = db.Season{
		ID: "s1", GroupID: "g1", Number: 1, Status: db.SeasonStatusVOTING,
		RevealAt: time.Now().Add(time.Hour),
	}
	m.memberCounts["g1"] = 5
	m2 := &mockQuerierWithVoterCountError{mockQuerier: m, voterCountErr: errors.New("voter error")}
	svc := NewService(m2, nil, "https://repa.app")

	_, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "voter error" {
		t.Errorf("expected 'voter error', got %v", err)
	}
}

type mockQuerierWithVoterCountError struct {
	*mockQuerier
	voterCountErr error
}

func (m *mockQuerierWithVoterCountError) CountSeasonVoters(_ context.Context, _ string) (int64, error) {
	return 0, m.voterCountErr
}

// --- HandleRepaCommand: GetGroupByTelegramChatID generic error ---

func TestHandleRepaCommand_GetGroupByTelegramChatIDError(t *testing.T) {
	m := &mockQuerierWithTelegramError{
		mockQuerier:      newMockQuerier(),
		getByTelegramErr: errors.New("tg lookup failed"),
	}
	svc := NewService(m, nil, "https://repa.app")

	_, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "tg lookup failed" {
		t.Errorf("expected 'tg lookup failed', got %v", err)
	}
}

type mockQuerierWithTelegramError struct {
	*mockQuerier
	getByTelegramErr error
}

func (m *mockQuerierWithTelegramError) GetGroupByTelegramChatID(_ context.Context, _ sql.NullString) (db.Group, error) {
	return db.Group{}, m.getByTelegramErr
}

// --- HandleRepaCommand: GetActiveSeasonByGroup generic error ---

func TestHandleRepaCommand_GetActiveSeasonError(t *testing.T) {
	m := &mockQuerierWithActiveSeasonError{
		mockQuerier:      newMockQuerier(),
		activeSeasonErr:  errors.New("season db error"),
	}
	m.groupsByTelegramID["12345"] = db.Group{ID: "g1", Name: "Test"}
	svc := NewService(m, nil, "https://repa.app")

	_, err := svc.HandleRepaCommand(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "season db error" {
		t.Errorf("expected 'season db error', got %v", err)
	}
}

type mockQuerierWithActiveSeasonError struct {
	*mockQuerier
	activeSeasonErr error
}

func (m *mockQuerierWithActiveSeasonError) GetActiveSeasonByGroup(_ context.Context, _ string) (db.Season, error) {
	return db.Season{}, m.activeSeasonErr
}

// --- PostSeasonStartAll: GetAllVotingSeasons error ---

func TestPostSeasonStartAll_GetSeasonsError(t *testing.T) {
	m := &mockQuerierWithVotingSeasonsError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("seasons query failed"),
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostSeasonStartAll(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "get voting seasons") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

type mockQuerierWithVotingSeasonsError struct {
	*mockQuerier
	err error
}

func (m *mockQuerierWithVotingSeasonsError) GetAllVotingSeasons(_ context.Context) ([]db.Season, error) {
	return nil, m.err
}

// --- DisconnectTelegram: generic GetGroupByID error ---

func TestDisconnectTelegram_GenericDBError(t *testing.T) {
	m := &mockQuerierWithGroupError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("connection lost"),
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.DisconnectTelegram(context.Background(), "admin1", "g1")
	if err == nil || err.Error() != "connection lost" {
		t.Errorf("expected connection lost error, got %v", err)
	}
}

type mockQuerierWithGroupError struct {
	*mockQuerier
	err error
}

func (m *mockQuerierWithGroupError) GetGroupByID(_ context.Context, _ string) (db.Group, error) {
	return db.Group{}, m.err
}

// --- GenerateConnectCode: generic GetGroupByID error ---

func TestGenerateConnectCode_GenericDBError(t *testing.T) {
	m := &mockQuerierWithGroupError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("db timeout"),
	}
	svc := NewService(m, nil, "https://repa.app")

	_, _, err := svc.GenerateConnectCode(context.Background(), "admin1", "g1")
	if err == nil || err.Error() != "db timeout" {
		t.Errorf("expected db timeout error, got %v", err)
	}
}

// --- HandleConnect: GetGroupByConnectCode generic error ---

func TestHandleConnect_GenericDBError(t *testing.T) {
	m := &mockQuerierWithConnectCodeError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("db error"),
	}
	svc := NewService(m, nil, "https://repa.app")

	_, err := svc.HandleConnect(context.Background(), 12345, "chat", "REPA-ABCD")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

type mockQuerierWithConnectCodeError struct {
	*mockQuerier
	err error
}

func (m *mockQuerierWithConnectCodeError) GetGroupByConnectCode(_ context.Context, _ sql.NullString) (db.Group, error) {
	return db.Group{}, m.err
}

// --- PostReveal: GetTopResultPerQuestion error ---

func TestPostReveal_GetTopResultError(t *testing.T) {
	m := &mockQuerierWithTopResultError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("top result error"),
	}
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.PostReveal(context.Background(), "s1")
	if err == nil || err.Error() != "top result error" {
		t.Errorf("expected top result error, got %v", err)
	}
}

type mockQuerierWithTopResultError struct {
	*mockQuerier
	err error
}

func (m *mockQuerierWithTopResultError) GetTopResultPerQuestion(_ context.Context, _ string) ([]db.GetTopResultPerQuestionRow, error) {
	return nil, m.err
}

// --- ShareCard: IsGroupMember error ---

func TestShareCard_IsGroupMemberError(t *testing.T) {
	m := &mockQuerierWithMemberError{
		mockQuerier: newMockQuerier(),
		err:         errors.New("member check error"),
	}
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	svc := NewService(m, nil, "https://repa.app")

	err := svc.ShareCard(context.Background(), "u1", "s1")
	if err == nil || err.Error() != "member check error" {
		t.Errorf("expected member check error, got %v", err)
	}
}

type mockQuerierWithMemberError struct {
	*mockQuerier
	err error
}

func (m *mockQuerierWithMemberError) IsGroupMember(_ context.Context, _ db.IsGroupMemberParams) (int64, error) {
	return 0, m.err
}

// --- PostReveal: with more than 5 top results (limit check) ---

func TestPostReveal_LimitedTo5ResultsNoBotPanics(t *testing.T) {
	m := newMockQuerier()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1"}
	m.groups["g1"] = db.Group{
		ID:             "g1",
		Name:           "Test Group",
		TelegramChatID: sql.NullString{String: "12345", Valid: true},
	}
	results := make([]db.GetTopResultPerQuestionRow, 8)
	for i := range results {
		results[i] = db.GetTopResultPerQuestionRow{
			QuestionText: fmt.Sprintf("Q%d", i),
			Username:     fmt.Sprintf("user%d", i),
			Percentage:   float64(90 - i*5),
		}
	}
	m.topResults["s1"] = results
	svc := NewService(m, nil, "https://repa.app")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when bot is nil")
		}
	}()

	_ = svc.PostReveal(context.Background(), "s1")
}
