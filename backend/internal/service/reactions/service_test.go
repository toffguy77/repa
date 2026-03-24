package reactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// --- Mock ---

type mockQuerier struct {
	db.Querier
	seasons        map[string]db.Season
	members        map[string]map[string]bool // groupID -> userID -> true
	reactions      []db.GetReactionsForUserRow
	createdParams  []db.CreateReactionParams
	failOnCreate   bool
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

func (m *mockQuerier) CreateReaction(_ context.Context, arg db.CreateReactionParams) (db.Reaction, error) {
	if m.failOnCreate {
		return db.Reaction{}, errors.New("db error")
	}
	m.createdParams = append(m.createdParams, arg)
	return db.Reaction{ID: arg.ID}, nil
}

func (m *mockQuerier) GetReactionsForUser(_ context.Context, _ db.GetReactionsForUserParams) ([]db.GetReactionsForUserRow, error) {
	return m.reactions, nil
}

func newMock() *mockQuerier {
	return &mockQuerier{
		seasons: map[string]db.Season{
			"s1": {ID: "s1", GroupID: "g1", Status: db.SeasonStatusREVEALED},
		},
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true, "u3": true},
		},
		reactions: nil,
	}
}

// --- CreateReaction tests ---

func TestCreateReaction_InvalidEmoji(t *testing.T) {
	tests := []struct {
		name  string
		emoji string
	}{
		{"thumbs up", "\U0001F44D"},
		{"heart", "\u2764\uFE0F"},
		{"empty", ""},
		{"plain text", "hello"},
		{"concatenated", "\U0001F602\U0001F525"},
	}
	svc := NewService(newMock(), nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", tt.emoji)
			if err != ErrInvalidEmoji {
				t.Errorf("expected ErrInvalidEmoji, got %v", err)
			}
		})
	}
}

func TestCreateReaction_SelfReaction(t *testing.T) {
	svc := NewService(newMock(), nil)
	_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u1", "\U0001F525")
	if err != ErrSelfReaction {
		t.Errorf("expected ErrSelfReaction, got %v", err)
	}
}

func TestCreateReaction_SeasonNotFound(t *testing.T) {
	svc := NewService(newMock(), nil)
	_, err := svc.CreateReaction(context.Background(), "nonexistent", "u1", "u2", "\U0001F525")
	if err != ErrSeasonNotFound {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestCreateReaction_SeasonNotRevealed(t *testing.T) {
	m := newMock()
	m.seasons["s1"] = db.Season{ID: "s1", GroupID: "g1", Status: db.SeasonStatusVOTING}
	svc := NewService(m, nil)
	_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", "\U0001F525")
	if err != ErrSeasonNotRevealed {
		t.Errorf("expected ErrSeasonNotRevealed, got %v", err)
	}
}

func TestCreateReaction_NotMember(t *testing.T) {
	svc := NewService(newMock(), nil)
	_, err := svc.CreateReaction(context.Background(), "s1", "outsider", "u2", "\U0001F525")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestCreateReaction_Success(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)
	counts, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", "\U0001F525")
	if err != nil {
		t.Fatal(err)
	}
	if counts == nil {
		t.Fatal("expected non-nil counts")
	}
	if len(m.createdParams) != 1 {
		t.Errorf("expected 1 created reaction, got %d", len(m.createdParams))
	}
	p := m.createdParams[0]
	if p.SeasonID != "s1" || p.ReactorID != "u1" || p.TargetID != "u2" || p.Emoji != "\U0001F525" {
		t.Errorf("unexpected params: %+v", p)
	}
}

func TestCreateReaction_DBError(t *testing.T) {
	m := newMock()
	m.failOnCreate = true
	svc := NewService(m, nil)
	_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", "\U0001F525")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreateReaction_AllValidEmojis(t *testing.T) {
	emojis := []string{"\U0001F602", "\U0001F525", "\U0001F480", "\U0001F440", "\U0001FAE1"}
	for _, emoji := range emojis {
		m := newMock()
		svc := NewService(m, nil)
		_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", emoji)
		if err != nil {
			t.Errorf("emoji %q should be valid, got %v", emoji, err)
		}
	}
}

func TestCreateReaction_ValidationOrder(t *testing.T) {
	svc := NewService(newMock(), nil)
	// Both invalid emoji AND self-reaction — emoji check runs first
	_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u1", "invalid")
	if err != ErrInvalidEmoji {
		t.Errorf("expected ErrInvalidEmoji (checked first), got %v", err)
	}
}

func TestCreateReaction_NoAsynqClient(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)
	// Should succeed even without asynq client
	_, err := svc.CreateReaction(context.Background(), "s1", "u1", "u2", "\U0001F525")
	if err != nil {
		t.Errorf("expected success with nil asynq, got %v", err)
	}
}

// --- GetReactions tests ---

func TestGetReactions_Success(t *testing.T) {
	m := newMock()
	m.reactions = []db.GetReactionsForUserRow{
		{ReactorID: "u1", Emoji: "\U0001F525"},
		{ReactorID: "u2", Emoji: "\U0001F525"},
		{ReactorID: "u3", Emoji: "\U0001F602"},
	}
	svc := NewService(m, nil)
	counts, err := svc.GetReactions(context.Background(), "s1", "u2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if counts.Counts["\U0001F525"] != 2 {
		t.Errorf("expected fire count 2, got %d", counts.Counts["\U0001F525"])
	}
	if counts.Counts["\U0001F602"] != 1 {
		t.Errorf("expected laugh count 1, got %d", counts.Counts["\U0001F602"])
	}
	if counts.MyEmoji == nil || *counts.MyEmoji != "\U0001F525" {
		t.Error("expected my_emoji to be fire")
	}
}

func TestGetReactions_NoMyEmoji(t *testing.T) {
	m := newMock()
	m.reactions = []db.GetReactionsForUserRow{
		{ReactorID: "u2", Emoji: "\U0001F525"},
	}
	svc := NewService(m, nil)
	counts, err := svc.GetReactions(context.Background(), "s1", "u2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if counts.MyEmoji != nil {
		t.Error("expected nil my_emoji when current user hasn't reacted")
	}
}

func TestGetReactions_Empty(t *testing.T) {
	m := newMock()
	svc := NewService(m, nil)
	counts, err := svc.GetReactions(context.Background(), "s1", "u2", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(counts.Counts) != 0 {
		t.Errorf("expected empty counts, got %d", len(counts.Counts))
	}
}

func TestGetReactions_SeasonNotFound(t *testing.T) {
	svc := NewService(newMock(), nil)
	_, err := svc.GetReactions(context.Background(), "nonexistent", "u2", "u1")
	if err != ErrSeasonNotFound {
		t.Errorf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestGetReactions_NotMember(t *testing.T) {
	svc := NewService(newMock(), nil)
	_, err := svc.GetReactions(context.Background(), "s1", "u2", "outsider")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- Sentinel errors ---

func TestSentinelErrors_Distinct(t *testing.T) {
	errs := []error{ErrSeasonNotFound, ErrSeasonNotRevealed, ErrNotMember, ErrSelfReaction, ErrInvalidEmoji}
	for i := range errs {
		for j := i + 1; j < len(errs); j++ {
			if errs[i] == errs[j] {
				t.Errorf("errors %d and %d are the same: %v", i, j, errs[i])
			}
		}
	}
}

// --- AllowedEmojis ---

func TestAllowedEmojis_Count(t *testing.T) {
	if len(allowedEmojis) != 5 {
		t.Errorf("expected 5 allowed emojis, got %d", len(allowedEmojis))
	}
}

// --- ReactionCounts JSON ---

func TestReactionCounts_JSON_RoundTrip(t *testing.T) {
	e := "\U0001F525"
	rc := ReactionCounts{
		Counts:  map[string]int{"\U0001F525": 3, "\U0001F602": 1},
		MyEmoji: &e,
	}
	data, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ReactionCounts
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.MyEmoji == nil || *decoded.MyEmoji != e {
		t.Errorf("expected my_emoji=%q", e)
	}
	if decoded.Counts["\U0001F525"] != 3 {
		t.Errorf("expected fire=3, got %d", decoded.Counts["\U0001F525"])
	}
}

// --- NewService ---

func TestNewService(t *testing.T) {
	svc := NewService(nil, nil)
	if svc == nil {
		t.Error("expected non-nil service")
	}
}
