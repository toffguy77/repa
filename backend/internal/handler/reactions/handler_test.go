package reactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	reactionssvc "github.com/repa-app/repa/internal/service/reactions"
)

// ---------------------------------------------------------------------------
// Mock Querier — only the 4 methods used by reactions service are implemented.
// All other Querier methods panic with "not implemented".
// ---------------------------------------------------------------------------

type mockQuerier struct {
	db.Querier

	getSeasonByIDFn      func(ctx context.Context, id string) (db.Season, error)
	isGroupMemberFn      func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error)
	createReactionFn     func(ctx context.Context, arg db.CreateReactionParams) (db.Reaction, error)
	getReactionsForUserFn func(ctx context.Context, arg db.GetReactionsForUserParams) ([]db.GetReactionsForUserRow, error)
}

func (m *mockQuerier) GetSeasonByID(ctx context.Context, id string) (db.Season, error) {
	return m.getSeasonByIDFn(ctx, id)
}

func (m *mockQuerier) IsGroupMember(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
	return m.isGroupMemberFn(ctx, arg)
}

func (m *mockQuerier) CreateReaction(ctx context.Context, arg db.CreateReactionParams) (db.Reaction, error) {
	return m.createReactionFn(ctx, arg)
}

func (m *mockQuerier) GetReactionsForUser(ctx context.Context, arg db.GetReactionsForUserParams) ([]db.GetReactionsForUserRow, error) {
	return m.getReactionsForUserFn(ctx, arg)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = appmw.NewValidator()
	return e
}

func setUser(c echo.Context, userID string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: "testuser"})
}

func revealedSeason() db.Season {
	return db.Season{
		ID:       "season-1",
		GroupID:  "group-1",
		Number:   1,
		Status:   db.SeasonStatusREVEALED,
		StartsAt: time.Now().Add(-7 * 24 * time.Hour),
		RevealAt: time.Now().Add(-1 * time.Hour),
		EndsAt:   time.Now().Add(6 * 24 * time.Hour),
	}
}

func votingSeason() db.Season {
	s := revealedSeason()
	s.Status = db.SeasonStatusVOTING
	return s
}

// defaultMock returns a mockQuerier pre-wired for the happy path.
func defaultMock() *mockQuerier {
	return &mockQuerier{
		getSeasonByIDFn: func(_ context.Context, _ string) (db.Season, error) {
			return revealedSeason(), nil
		},
		isGroupMemberFn: func(_ context.Context, _ db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
		createReactionFn: func(_ context.Context, arg db.CreateReactionParams) (db.Reaction, error) {
			return db.Reaction{
				ID:        arg.ID,
				SeasonID:  arg.SeasonID,
				ReactorID: arg.ReactorID,
				TargetID:  arg.TargetID,
				Emoji:     arg.Emoji,
				CreatedAt: time.Now(),
			}, nil
		},
		getReactionsForUserFn: func(_ context.Context, arg db.GetReactionsForUserParams) ([]db.GetReactionsForUserRow, error) {
			return []db.GetReactionsForUserRow{
				{
					ID:              "r1",
					SeasonID:        arg.SeasonID,
					ReactorID:       "user-1",
					TargetID:        arg.TargetID,
					Emoji:           "\U0001F525",
					CreatedAt:       time.Now(),
					ReactorUsername: "testuser",
				},
			}, nil
		},
	}
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func parseError(t *testing.T, rec *httptest.ResponseRecorder) errorBody {
	t.Helper()
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse error body: %v\nbody: %s", err, rec.Body.String())
	}
	return body
}

// ---------------------------------------------------------------------------
// CreateReaction tests
// ---------------------------------------------------------------------------

func TestCreateReaction_Success(t *testing.T) {
	mq := defaultMock()
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"` + "\U0001F525" + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data struct {
			Counts  map[string]int `json:"counts"`
			MyEmoji *string        `json:"my_emoji"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data.Counts["\U0001F525"] != 1 {
		t.Errorf("expected fire count 1, got %d", resp.Data.Counts["\U0001F525"])
	}
	if resp.Data.MyEmoji == nil || *resp.Data.MyEmoji != "\U0001F525" {
		t.Errorf("expected my_emoji to be fire emoji, got %v", resp.Data.MyEmoji)
	}
}

func TestCreateReaction_InvalidEmoji(t *testing.T) {
	mq := defaultMock()
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "INVALID_EMOJI" {
		t.Errorf("expected INVALID_EMOJI, got %q", errBody.Error.Code)
	}
}

func TestCreateReaction_SelfReaction(t *testing.T) {
	mq := defaultMock()
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"` + "\U0001F525" + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-1") // targetId == current user
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "SELF_REACTION" {
		t.Errorf("expected SELF_REACTION, got %q", errBody.Error.Code)
	}
}

func TestCreateReaction_SeasonNotFound(t *testing.T) {
	mq := defaultMock()
	mq.getSeasonByIDFn = func(_ context.Context, _ string) (db.Season, error) {
		return db.Season{}, sql.ErrNoRows
	}
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"` + "\U0001F525" + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("no-such-season", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %q", errBody.Error.Code)
	}
}

func TestCreateReaction_SeasonNotRevealed(t *testing.T) {
	mq := defaultMock()
	mq.getSeasonByIDFn = func(_ context.Context, _ string) (db.Season, error) {
		return votingSeason(), nil
	}
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"` + "\U0001F525" + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "SEASON_NOT_REVEALED" {
		t.Errorf("expected SEASON_NOT_REVEALED, got %q", errBody.Error.Code)
	}
}

func TestCreateReaction_NotMember(t *testing.T) {
	mq := defaultMock()
	mq.isGroupMemberFn = func(_ context.Context, _ db.IsGroupMemberParams) (int64, error) {
		return 0, nil
	}
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	body := `{"emoji":"` + "\U0001F525" + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %q", errBody.Error.Code)
	}
}

func TestCreateReaction_BadRequestBody(t *testing.T) {
	mq := defaultMock()
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.CreateReaction(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "VALIDATION" {
		t.Errorf("expected VALIDATION, got %q", errBody.Error.Code)
	}
}

// ---------------------------------------------------------------------------
// GetReactions tests
// ---------------------------------------------------------------------------

func TestGetReactions_Success(t *testing.T) {
	mq := defaultMock()
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.GetReactions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data struct {
			Counts  map[string]int `json:"counts"`
			MyEmoji *string        `json:"my_emoji"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data.Counts["\U0001F525"] != 1 {
		t.Errorf("expected fire count 1, got %d", resp.Data.Counts["\U0001F525"])
	}
	if resp.Data.MyEmoji == nil || *resp.Data.MyEmoji != "\U0001F525" {
		t.Errorf("expected my_emoji to be fire emoji, got %v", resp.Data.MyEmoji)
	}
}

func TestGetReactions_NotMember(t *testing.T) {
	mq := defaultMock()
	mq.isGroupMemberFn = func(_ context.Context, _ db.IsGroupMemberParams) (int64, error) {
		return 0, nil
	}
	h := NewHandler(reactionssvc.NewService(mq, nil))
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId", "targetId")
	c.SetParamValues("season-1", "user-2")
	setUser(c, "user-1")

	if err := h.GetReactions(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	errBody := parseError(t, rec)
	if errBody.Error.Code != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %q", errBody.Error.Code)
	}
}
