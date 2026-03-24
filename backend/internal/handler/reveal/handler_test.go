package reveal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	cardssvc "github.com/repa-app/repa/internal/service/cards"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (tv *testValidator) Validate(i any) error { return nil }

// errorBody is used to decode the JSON error envelope.
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestMapServiceError_SeasonNotFound(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	_ = mapServiceError(c, revealsvc.ErrSeasonNotFound)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %s", body.Error.Code)
	}
	if body.Error.Message != "Season not found" {
		t.Errorf("expected message 'Season not found', got %s", body.Error.Message)
	}
}

func TestMapServiceError_SeasonNotRevealed(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	_ = mapServiceError(c, revealsvc.ErrSeasonNotRevealed)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != "SEASON_NOT_REVEALED" {
		t.Errorf("expected error code SEASON_NOT_REVEALED, got %s", body.Error.Code)
	}
	if body.Error.Message != "Season results are not yet available" {
		t.Errorf("expected message 'Season results are not yet available', got %s", body.Error.Message)
	}
}

func TestMapServiceError_NotMember(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	_ = mapServiceError(c, revealsvc.ErrNotMember)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != "NOT_MEMBER" {
		t.Errorf("expected error code NOT_MEMBER, got %s", body.Error.Code)
	}
	if body.Error.Message != "You are not a member of this group" {
		t.Errorf("expected message 'You are not a member of this group', got %s", body.Error.Message)
	}
}

func TestMapServiceError_InsufficientFunds(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	_ = mapServiceError(c, revealsvc.ErrInsufficientFunds)

	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected status %d, got %d", http.StatusPaymentRequired, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != "INSUFFICIENT_FUNDS" {
		t.Errorf("expected error code INSUFFICIENT_FUNDS, got %s", body.Error.Code)
	}
	if body.Error.Message != "Not enough crystals" {
		t.Errorf("expected message 'Not enough crystals', got %s", body.Error.Message)
	}
}

func TestMapServiceError_UnknownError(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	_ = mapServiceError(c, errors.New("some unexpected error"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != "INTERNAL" {
		t.Errorf("expected error code INTERNAL, got %s", body.Error.Code)
	}
	if body.Error.Message != "Something went wrong" {
		t.Errorf("expected message 'Something went wrong', got %s", body.Error.Message)
	}
}

func TestMapServiceError_WrappedSeasonNotFound(t *testing.T) {
	e := setupEcho()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	wrapped := errors.Join(revealsvc.ErrSeasonNotFound, errors.New("extra context"))
	_ = mapServiceError(c, wrapped)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected wrapped error to match ErrSeasonNotFound, got status %d", rec.Code)
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ---------------------------------------------------------------------------
// Mock querier for handler-level tests (constructs real services)
// ---------------------------------------------------------------------------

type handlerMockQuerier struct {
	db.Querier
	seasons         map[string]db.Season
	members         map[string]map[string]bool
	resultsByUser   map[string][]db.GetSeasonResultsByUserRow
	topPerQuestion  map[string][]db.GetTopResultPerQuestionRow
	uniqueVoters    map[string]int64
	allResultsUsers map[string][]db.GetAllSeasonResultsWithUsersRow
	balance         map[string]int32
	cardCache       map[string]db.CardCache // key = userID:seasonID
	hasDetector     map[string]bool         // key = userID:seasonID
	voterProfiles   map[string][]db.GetVoterProfilesBySeasonRow

	// error injection
	forceErr error
}

func (m *handlerMockQuerier) GetSeasonByID(_ context.Context, id string) (db.Season, error) {
	if m.forceErr != nil {
		return db.Season{}, m.forceErr
	}
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

func (m *handlerMockQuerier) GetSeasonResultsByUser(_ context.Context, arg db.GetSeasonResultsByUserParams) ([]db.GetSeasonResultsByUserRow, error) {
	return m.resultsByUser[arg.SeasonID+":"+arg.TargetID], nil
}

func (m *handlerMockQuerier) GetTopResultPerQuestion(_ context.Context, seasonID string) ([]db.GetTopResultPerQuestionRow, error) {
	return m.topPerQuestion[seasonID], nil
}

func (m *handlerMockQuerier) CountUniqueVoters(_ context.Context, seasonID string) (int64, error) {
	return m.uniqueVoters[seasonID], nil
}

func (m *handlerMockQuerier) GetPreviousRevealedSeason(_ context.Context, _ db.GetPreviousRevealedSeasonParams) (db.Season, error) {
	return db.Season{}, sql.ErrNoRows
}

func (m *handlerMockQuerier) GetSeasonAchievements(_ context.Context, _ sql.NullString) ([]db.Achievement, error) {
	return []db.Achievement{}, nil
}

func (m *handlerMockQuerier) GetCardCache(_ context.Context, arg db.GetCardCacheParams) (db.CardCache, error) {
	key := arg.UserID + ":" + arg.SeasonID
	c, ok := m.cardCache[key]
	if !ok {
		return db.CardCache{}, sql.ErrNoRows
	}
	return c, nil
}

func (m *handlerMockQuerier) GetAllSeasonResultsWithUsers(_ context.Context, seasonID string) ([]db.GetAllSeasonResultsWithUsersRow, error) {
	return m.allResultsUsers[seasonID], nil
}

func (m *handlerMockQuerier) GetUserBalance(_ context.Context, userID string) (int32, error) {
	b, ok := m.balance[userID]
	if !ok {
		return 0, nil
	}
	return b, nil
}

func (m *handlerMockQuerier) HasDetector(_ context.Context, arg db.HasDetectorParams) (bool, error) {
	return m.hasDetector[arg.UserID+":"+arg.SeasonID], nil
}

func (m *handlerMockQuerier) GetVoterProfilesBySeason(_ context.Context, seasonID string) ([]db.GetVoterProfilesBySeasonRow, error) {
	return m.voterProfiles[seasonID], nil
}

func (m *handlerMockQuerier) CreateCrystalLog(_ context.Context, _ db.CreateCrystalLogParams) (db.CrystalLog, error) {
	return db.CrystalLog{}, nil
}

func (m *handlerMockQuerier) CreateDetector(_ context.Context, _ db.CreateDetectorParams) (db.Detector, error) {
	return db.Detector{}, nil
}

func (m *handlerMockQuerier) LockUserForUpdate(_ context.Context, id string) (string, error) {
	return id, nil
}

func newHandlerMock() *handlerMockQuerier {
	return &handlerMockQuerier{
		seasons: map[string]db.Season{
			"s-revealed": {ID: "s-revealed", GroupID: "g1", Number: 2, Status: db.SeasonStatusREVEALED},
			"s-voting":   {ID: "s-voting", GroupID: "g1", Number: 1, Status: db.SeasonStatusVOTING},
		},
		members: map[string]map[string]bool{
			"g1": {"u1": true, "u2": true},
		},
		resultsByUser: map[string][]db.GetSeasonResultsByUserRow{
			"s-revealed:u1": {
				{QuestionID: "q1", QuestionText: "Funniest?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 80, VoteCount: 4, TotalVoters: 5},
				{QuestionID: "q2", QuestionText: "Hottest?", QuestionCategory: db.QuestionCategoryHOT, Percentage: 60, VoteCount: 3, TotalVoters: 5},
				{QuestionID: "q3", QuestionText: "Secrets?", QuestionCategory: db.QuestionCategorySECRETS, Percentage: 40, VoteCount: 2, TotalVoters: 5},
				{QuestionID: "q4", QuestionText: "Studies?", QuestionCategory: db.QuestionCategorySTUDY, Percentage: 20, VoteCount: 1, TotalVoters: 5},
			},
		},
		topPerQuestion: map[string][]db.GetTopResultPerQuestionRow{
			"s-revealed": {
				{QuestionID: "q1", TargetID: "u1", QuestionText: "Funniest?", Username: "alice", Percentage: 80},
			},
		},
		uniqueVoters: map[string]int64{"s-revealed": 5},
		allResultsUsers: map[string][]db.GetAllSeasonResultsWithUsersRow{
			"s-revealed": {
				{TargetID: "u1", QuestionID: "q1", QuestionText: "Funniest?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 80, Username: "alice"},
				{TargetID: "u2", QuestionID: "q1", QuestionText: "Funniest?", QuestionCategory: db.QuestionCategoryFUNNY, Percentage: 50, Username: "bob"},
			},
		},
		balance:       map[string]int32{"u1": 20, "u2": 3},
		cardCache:     map[string]db.CardCache{"u1:s-revealed": {ImageUrl: "https://cdn.example.com/card.png"}},
		hasDetector:   map[string]bool{},
		voterProfiles: map[string][]db.GetVoterProfilesBySeasonRow{},
	}
}

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func buildHandler(mq *handlerMockQuerier) *Handler {
	revealSvc := revealsvc.NewService(mq, nil)
	cardsSvc := cardssvc.NewService(mq, nil)
	return NewHandler(revealSvc, cardsSvc)
}

// ---------------------------------------------------------------------------
// GetReveal
// ---------------------------------------------------------------------------

func TestGetReveal_Success(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s-revealed/reveal", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u1", "alice")

	if err := h.GetReveal(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("missing data envelope")
	}
	myCard, ok := data["my_card"].(map[string]any)
	if !ok {
		t.Fatal("missing my_card")
	}
	topAttrs, ok := myCard["top_attributes"].([]any)
	if !ok || len(topAttrs) != 3 {
		t.Errorf("expected 3 top_attributes, got %v", topAttrs)
	}
}

func TestGetReveal_NotMember(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "outsider", "outsider")

	_ = h.GetReveal(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetReveal_SeasonNotFound(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("nonexistent")
	setUser(c, "u1", "alice")

	_ = h.GetReveal(c)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetReveal_NotRevealed(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-voting")
	setUser(c, "u1", "alice")

	_ = h.GetReveal(c)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// GetMembersCards
// ---------------------------------------------------------------------------

func TestGetMembersCards_Success(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u1", "alice")

	if err := h.GetMembersCards(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	data := body["data"].(map[string]any)
	members := data["members"].([]any)
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

// ---------------------------------------------------------------------------
// GetMyCardURL
// ---------------------------------------------------------------------------

func TestGetMyCardURL_Ready(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u1", "alice")

	if err := h.GetMyCardURL(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	if data["status"] != "ready" {
		t.Errorf("expected status ready, got %v", data["status"])
	}
	if data["image_url"] != "https://cdn.example.com/card.png" {
		t.Errorf("unexpected image_url: %v", data["image_url"])
	}
}

func TestGetMyCardURL_Generating(t *testing.T) {
	mq := newHandlerMock()
	// u2 has no card cache entry → sql.ErrNoRows
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u2", "bob")

	if err := h.GetMyCardURL(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	if data["status"] != "generating" {
		t.Errorf("expected status generating, got %v", data["status"])
	}
	if data["image_url"] != nil {
		t.Errorf("expected nil image_url, got %v", data["image_url"])
	}
}

// ---------------------------------------------------------------------------
// OpenHidden — error paths
// ---------------------------------------------------------------------------

func TestOpenHidden_InsufficientFunds(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u2", "bob") // balance = 3, cost = 5

	_ = h.OpenHidden(c)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rec.Code)
	}
}

func TestOpenHidden_NotMember(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "outsider", "outsider")

	_ = h.OpenHidden(c)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// BuyDetector — error paths
// ---------------------------------------------------------------------------

func TestBuyDetector_InsufficientFunds(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u2", "bob") // balance = 3, cost = 10

	_ = h.BuyDetector(c)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rec.Code)
	}
}

func TestBuyDetector_SeasonNotFound(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("nonexistent")
	setUser(c, "u1", "alice")

	_ = h.BuyDetector(c)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetDetector_Success(t *testing.T) {
	mq := newHandlerMock()
	h := buildHandler(mq)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s-revealed")
	setUser(c, "u1", "alice")

	if err := h.GetDetector(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	if data["purchased"] != false {
		t.Errorf("expected purchased=false, got %v", data["purchased"])
	}
}
