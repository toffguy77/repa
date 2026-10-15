package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	pushsvc "github.com/repa-app/repa/internal/service/push"
)

// ---------- Payload marshaling / unmarshaling ----------

func TestRevealPayloadMarshalUnmarshal(t *testing.T) {
	original := RevealPayload{SeasonID: "season-abc", Attempt: 3}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded RevealPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SeasonID != original.SeasonID {
		t.Errorf("SeasonID = %q, want %q", decoded.SeasonID, original.SeasonID)
	}
	if decoded.Attempt != original.Attempt {
		t.Errorf("Attempt = %d, want %d", decoded.Attempt, original.Attempt)
	}
}

func TestCardsPayloadMarshalUnmarshal(t *testing.T) {
	original := CardsPayload{SeasonID: "season-xyz"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded CardsPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SeasonID != original.SeasonID {
		t.Errorf("SeasonID = %q, want %q", decoded.SeasonID, original.SeasonID)
	}
}

func TestAchievementsPayloadMarshalUnmarshal(t *testing.T) {
	original := AchievementsPayload{SeasonID: "season-ach"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded AchievementsPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SeasonID != original.SeasonID {
		t.Errorf("SeasonID = %q, want %q", decoded.SeasonID, original.SeasonID)
	}
}

func TestTelegramPayloadMarshalUnmarshal(t *testing.T) {
	original := TelegramPayload{
		GroupID:  "group-1",
		SeasonID: "season-2",
		UserID:   "user-3",
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded TelegramPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.GroupID != original.GroupID {
		t.Errorf("GroupID = %q, want %q", decoded.GroupID, original.GroupID)
	}
	if decoded.SeasonID != original.SeasonID {
		t.Errorf("SeasonID = %q, want %q", decoded.SeasonID, original.SeasonID)
	}
	if decoded.UserID != original.UserID {
		t.Errorf("UserID = %q, want %q", decoded.UserID, original.UserID)
	}
}

func TestTelegramPayloadOmitsEmpty(t *testing.T) {
	p := TelegramPayload{GroupID: "g1"}
	data, _ := json.Marshal(p)
	var m map[string]any
	json.Unmarshal(data, &m)
	if _, ok := m["season_id"]; ok {
		t.Error("expected season_id to be omitted when empty")
	}
	if _, ok := m["user_id"]; ok {
		t.Error("expected user_id to be omitted when empty")
	}
}

func TestRevealPushPayloadMarshalUnmarshal(t *testing.T) {
	original := RevealPushPayload{SeasonID: "s-push"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded RevealPushPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.SeasonID != original.SeasonID {
		t.Errorf("SeasonID = %q, want %q", decoded.SeasonID, original.SeasonID)
	}
}

// ---------- Invalid payload → error on unmarshal ----------

func TestCardsHandlerInvalidPayload(t *testing.T) {
	proc := &CardsProcessor{} // nil service is fine; we won't reach svc call
	task := asynq.NewTask(lib.TypeCardsGenerate, []byte("not-json"))
	err := proc.HandleCardsGenerate(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestAchievementsHandlerInvalidPayload(t *testing.T) {
	proc := &AchievementsProcessor{}
	task := asynq.NewTask(lib.TypeAchievements, []byte("{invalid"))
	err := proc.HandleAchievements(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestRevealProcessorInvalidPayload(t *testing.T) {
	proc := &RevealProcessor{}
	task := asynq.NewTask(lib.TypeRevealProcess, []byte("bad"))
	err := proc.HandleRevealProcess(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestTelegramRevealPostInvalidPayload(t *testing.T) {
	proc := &TelegramProcessor{}
	task := asynq.NewTask(lib.TypeTelegramReveal, []byte("bad"))
	err := proc.HandleRevealPost(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestTelegramShareCardInvalidPayload(t *testing.T) {
	proc := &TelegramProcessor{}
	task := asynq.NewTask(lib.TypeTelegramShare, []byte("bad"))
	err := proc.HandleShareCard(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestTelegramSeasonStartInvalidPayload(t *testing.T) {
	proc := &TelegramProcessor{}
	task := asynq.NewTask(lib.TypeTelegramStart, []byte("bad"))
	err := proc.HandleSeasonStart(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestPushRevealNotificationInvalidPayload(t *testing.T) {
	proc := &PushProcessor{}
	task := asynq.NewTask(lib.TypePushReveal, []byte("bad"))
	err := proc.HandleRevealNotification(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestPushReactionInvalidPayload(t *testing.T) {
	proc := &PushProcessor{}
	task := asynq.NewTask(lib.TypeReactionPush, []byte("{broken"))
	err := proc.HandleReactionPush(t.Context(), task)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

// ---------- Valid payload unmarshaling in handlers ----------

func TestCardsHandlerValidPayloadUnmarshal(t *testing.T) {
	payload, _ := json.Marshal(CardsPayload{SeasonID: "s1"})
	task := asynq.NewTask(lib.TypeCardsGenerate, payload)
	// We can't call HandleCardsGenerate to completion because svc is nil,
	// but we verified invalid payloads are rejected above.
	// Verify the payload round-trips through asynq.Task.
	var p CardsPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("unmarshal from task payload: %v", err)
	}
	if p.SeasonID != "s1" {
		t.Errorf("SeasonID = %q, want %q", p.SeasonID, "s1")
	}
}

func TestAchievementsHandlerValidPayloadUnmarshal(t *testing.T) {
	payload, _ := json.Marshal(AchievementsPayload{SeasonID: "s2"})
	task := asynq.NewTask(lib.TypeAchievements, payload)
	var p AchievementsPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("unmarshal from task payload: %v", err)
	}
	if p.SeasonID != "s2" {
		t.Errorf("SeasonID = %q, want %q", p.SeasonID, "s2")
	}
}

func TestRevealPayloadViaAsynqTask(t *testing.T) {
	payload, _ := json.Marshal(RevealPayload{SeasonID: "s3", Attempt: 5})
	task := asynq.NewTask(lib.TypeRevealProcess, payload)
	var p RevealPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("unmarshal from task payload: %v", err)
	}
	if p.SeasonID != "s3" || p.Attempt != 5 {
		t.Errorf("unexpected payload: %+v", p)
	}
}

func TestTelegramPayloadViaAsynqTask(t *testing.T) {
	payload, _ := json.Marshal(TelegramPayload{GroupID: "g1", SeasonID: "s1", UserID: "u1"})
	task := asynq.NewTask(lib.TypeTelegramReveal, payload)
	var p TelegramPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.GroupID != "g1" || p.SeasonID != "s1" || p.UserID != "u1" {
		t.Errorf("unexpected payload: %+v", p)
	}
}

// ---------- Constructor tests ----------

func TestNewRevealChecker(t *testing.T) {
	rc := NewRevealChecker(nil, nil)
	if rc == nil {
		t.Fatal("NewRevealChecker returned nil")
	}
}

func TestNewRevealProcessor(t *testing.T) {
	rp := NewRevealProcessor(nil, nil)
	if rp == nil {
		t.Fatal("NewRevealProcessor returned nil")
	}
}

func TestNewCardsProcessor(t *testing.T) {
	cp := NewCardsProcessor(nil)
	if cp == nil {
		t.Fatal("NewCardsProcessor returned nil")
	}
}

func TestNewAchievementsProcessor(t *testing.T) {
	ap := NewAchievementsProcessor(nil)
	if ap == nil {
		t.Fatal("NewAchievementsProcessor returned nil")
	}
}

func TestNewTelegramProcessor(t *testing.T) {
	tp := NewTelegramProcessor(nil)
	if tp == nil {
		t.Fatal("NewTelegramProcessor returned nil")
	}
}

func TestNewPushProcessor(t *testing.T) {
	pp := NewPushProcessor(nil)
	if pp == nil {
		t.Fatal("NewPushProcessor returned nil")
	}
}

func TestNewSeasonCreator(t *testing.T) {
	sc := NewSeasonCreator(nil)
	if sc == nil {
		t.Fatal("NewSeasonCreator returned nil")
	}
}

// ---------- Empty / nil payload edge cases ----------

func TestTelegramSeasonStartNilPayload(t *testing.T) {
	proc := &TelegramProcessor{} // nil svc — will fail on svc.PostSeasonStartAll
	task := asynq.NewTask(lib.TypeTelegramStart, nil)
	// With nil payload, HandleSeasonStart calls svc.PostSeasonStartAll which panics on nil svc.
	// We just verify it doesn't error on unmarshal — that's the cron-mode path.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from nil svc, unless implementation changed")
		}
	}()
	_ = proc.HandleSeasonStart(t.Context(), task)
}

func TestTelegramSeasonStartEmptyPayload(t *testing.T) {
	proc := &TelegramProcessor{}
	task := asynq.NewTask(lib.TypeTelegramStart, []byte{})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from nil svc for cron-mode path")
		}
	}()
	_ = proc.HandleSeasonStart(t.Context(), task)
}

// ---------- Reaction payload test ----------

func TestReactionPayloadMarshalRoundTrip(t *testing.T) {
	payload := struct {
		TargetID  string `json:"target_id"`
		ReactorID string `json:"reactor_id"`
		Emoji     string `json:"emoji"`
		GroupID   string `json:"group_id"`
		SeasonID  string `json:"season_id"`
	}{
		TargetID:  "t1",
		ReactorID: "r1",
		Emoji:     "fire",
		GroupID:   "g1",
		SeasonID:  "s1",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	task := asynq.NewTask(lib.TypeReactionPush, data)
	var decoded struct {
		TargetID  string `json:"target_id"`
		ReactorID string `json:"reactor_id"`
		Emoji     string `json:"emoji"`
		GroupID   string `json:"group_id"`
		SeasonID  string `json:"season_id"`
	}
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.TargetID != "t1" || decoded.ReactorID != "r1" || decoded.Emoji != "fire" {
		t.Errorf("unexpected decoded payload: %+v", decoded)
	}
}

// ---------- Task type constants ----------

func TestTaskTypeConstants(t *testing.T) {
	// Verify task type constants are non-empty and unique
	types := []string{
		lib.TypeRevealChecker,
		lib.TypeRevealProcess,
		lib.TypeSeasonCreator,
		lib.TypeAchievements,
		lib.TypePushWeekly,
		lib.TypePushTuesday,
		lib.TypePushWednesday,
		lib.TypePushThursday,
		lib.TypePushFriPreReveal,
		lib.TypePushReveal,
		lib.TypePushSundayPrev,
		lib.TypePushSundayStreak,
		lib.TypeTelegramStart,
		lib.TypeTelegramReveal,
		lib.TypeTelegramShare,
		lib.TypeReactionPush,
		lib.TypeCardsGenerate,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		if typ == "" {
			t.Error("found empty task type constant")
		}
		if seen[typ] {
			t.Errorf("duplicate task type constant: %q", typ)
		}
		seen[typ] = true
	}
}

// --- Mid-week push honesty ---

// signalQuerier answers only what withVotesAbout and the signal need.
type signalQuerier struct {
	db.Querier
	votersAbout map[string]int64 // "seasonID:userID" -> count
	err         error
}

func (q *signalQuerier) CountVotersAboutTarget(_ context.Context, arg db.CountVotersAboutTargetParams) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	return q.votersAbout[arg.SeasonID+":"+arg.TargetID], nil
}

func TestWithVotesAbout_KeepsOnlyMembersWithSomethingToSee(t *testing.T) {
	q := &signalQuerier{votersAbout: map[string]int64{
		"s1:has-votes": 2,
		"s1:also-has":  1,
		"s1:no-votes":  0,
	}}
	p := NewPushProcessor(pushsvc.NewService(q, nil, nil))

	got := p.withVotesAbout(context.Background(), "s1", []string{"has-votes", "no-votes", "also-has"})

	want := map[string]bool{"has-votes": true, "also-has": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want the two members with votes", got)
	}
	for _, id := range got {
		if !want[id] {
			t.Errorf("unexpected recipient %q: a push must only promise what the screen can show", id)
		}
	}
}

func TestWithVotesAbout_SkipsOnQueryFailure(t *testing.T) {
	q := &signalQuerier{err: errors.New("db down")}
	p := NewPushProcessor(pushsvc.NewService(q, nil, nil))

	got := p.withVotesAbout(context.Background(), "s1", []string{"a", "b"})

	if len(got) != 0 {
		t.Errorf("a failed check should skip rather than send a possibly-empty promise, got %v", got)
	}
}

func TestVoteSignalPayload_CarriesNoAnswer(t *testing.T) {
	payload := VoteSignalPayload{TargetID: "u2", SeasonID: "s1", GroupID: "g1"}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for key := range decoded {
		switch key {
		case "target_id", "season_id", "group_id":
		default:
			t.Errorf("unexpected key %q: the signal must carry no voter and no answer", key)
		}
	}
}

func TestVoteSignalHandlerInvalidPayload(t *testing.T) {
	p := NewPushProcessor(nil)
	err := p.HandleVoteSignal(context.Background(), asynq.NewTask(lib.TypePushVoteSignal, []byte("not json")))
	if err == nil {
		t.Error("expected an error for a malformed payload")
	}
}
