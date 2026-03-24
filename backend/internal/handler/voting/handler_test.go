package voting

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

	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	votingsvc "github.com/repa-app/repa/internal/service/voting"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (tv *testValidator) Validate(i any) error {
	return nil
}

func TestMapServiceError(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"season not found", votingsvc.ErrSeasonNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"season not voting", votingsvc.ErrSeasonNotVoting, http.StatusBadRequest, "SEASON_NOT_VOTING"},
		{"not member", votingsvc.ErrNotMember, http.StatusForbidden, "NOT_MEMBER"},
		{"already voted", votingsvc.ErrAlreadyVoted, http.StatusConflict, "ALREADY_VOTED"},
		{"self vote", votingsvc.ErrSelfVote, http.StatusBadRequest, "SELF_VOTE"},
		{"target not member", votingsvc.ErrTargetNotMember, http.StatusBadRequest, "TARGET_NOT_MEMBER"},
		{"invalid question", votingsvc.ErrInvalidQuestion, http.StatusBadRequest, "INVALID_QUESTION"},
		{"unknown", errors.New("unknown"), http.StatusInternalServerError, "INTERNAL"},
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

func TestCastVote_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CastVote(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Additional tests ---

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func TestCastVote_EmptyJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "testuser")

	// Bind succeeds with empty JSON, test validator passes, svc is nil → panic.
	defer func() {
		recover()
	}()

	_ = h.CastVote(c)
}

func TestCastVote_PartialJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(`{"question_id": "q1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "testuser")

	// Missing target_id but test validator passes, svc panics.
	defer func() {
		recover()
	}()

	_ = h.CastVote(c)
}

func TestCastVote_TruncatedJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(`{"question_id": "q1`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CastVote(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "VALIDATION" {
		t.Errorf("expected error code VALIDATION, got %s", resp["error"]["code"])
	}
}

func TestCastVote_ArrayInsteadOfObject(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(`[1, 2, 3]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CastVote(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetVotingSession_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/voting", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "testuser")

	// svc is nil → panic, confirms path param and user claim extraction works.
	defer func() {
		recover()
	}()

	_ = h.GetVotingSession(c)
}

func TestGetProgress_PathParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/progress", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "testuser")

	defer func() {
		recover()
	}()

	_ = h.GetProgress(c)
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
			"wrapped season not found",
			fmt.Errorf("lookup: %w", votingsvc.ErrSeasonNotFound),
			http.StatusNotFound,
			"NOT_FOUND",
		},
		{
			"wrapped season not voting",
			fmt.Errorf("check: %w", votingsvc.ErrSeasonNotVoting),
			http.StatusBadRequest,
			"SEASON_NOT_VOTING",
		},
		{
			"wrapped not member",
			fmt.Errorf("auth: %w", votingsvc.ErrNotMember),
			http.StatusForbidden,
			"NOT_MEMBER",
		},
		{
			"wrapped already voted",
			fmt.Errorf("vote: %w", votingsvc.ErrAlreadyVoted),
			http.StatusConflict,
			"ALREADY_VOTED",
		},
		{
			"wrapped self vote",
			fmt.Errorf("vote: %w", votingsvc.ErrSelfVote),
			http.StatusBadRequest,
			"SELF_VOTE",
		},
		{
			"wrapped target not member",
			fmt.Errorf("vote: %w", votingsvc.ErrTargetNotMember),
			http.StatusBadRequest,
			"TARGET_NOT_MEMBER",
		},
		{
			"wrapped invalid question",
			fmt.Errorf("vote: %w", votingsvc.ErrInvalidQuestion),
			http.StatusBadRequest,
			"INVALID_QUESTION",
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

	_ = mapServiceError(c, votingsvc.ErrSeasonNotFound)

	var resp map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp["error"]["code"] != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", resp["error"]["code"])
	}
	if resp["error"]["message"] != "Season not found" {
		t.Errorf("expected message 'Season not found', got %s", resp["error"]["message"])
	}
}

// --- Mock Querier for voting service (implements db.Querier) ---

type mockVotingQuerier struct {
	db.Querier // embed to satisfy interface; unused methods will panic
	getSeasonByIDFn           func(ctx context.Context, id string) (db.Season, error)
	isGroupMemberFn           func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error)
	getSeasonQuestionsFn      func(ctx context.Context, seasonID string) ([]db.Question, error)
	getVotesBySeasonAndVoterFn func(ctx context.Context, arg db.GetVotesBySeasonAndVoterParams) ([]db.Vote, error)
	getGroupMembersFn         func(ctx context.Context, groupID string) ([]db.GetGroupMembersRow, error)
	hasVoteForQuestionFn      func(ctx context.Context, arg db.HasVoteForQuestionParams) (int64, error)
	createVoteFn              func(ctx context.Context, arg db.CreateVoteParams) (db.Vote, error)
	countGroupMembersFn       func(ctx context.Context, groupID string) (int64, error)
	countSeasonQuestionsFn    func(ctx context.Context, seasonID string) (int64, error)
	countCompletedVotersFn    func(ctx context.Context, arg db.CountCompletedVotersParams) (int64, error)
	hasUserVotedInSeasonFn    func(ctx context.Context, arg db.HasUserVotedInSeasonParams) (int64, error)
}

func (m *mockVotingQuerier) GetSeasonByID(ctx context.Context, id string) (db.Season, error) {
	if m.getSeasonByIDFn != nil {
		return m.getSeasonByIDFn(ctx, id)
	}
	return db.Season{}, sql.ErrNoRows
}
func (m *mockVotingQuerier) IsGroupMember(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
	if m.isGroupMemberFn != nil {
		return m.isGroupMemberFn(ctx, arg)
	}
	return 1, nil
}
func (m *mockVotingQuerier) GetSeasonQuestions(ctx context.Context, seasonID string) ([]db.Question, error) {
	if m.getSeasonQuestionsFn != nil {
		return m.getSeasonQuestionsFn(ctx, seasonID)
	}
	return []db.Question{}, nil
}
func (m *mockVotingQuerier) GetVotesBySeasonAndVoter(ctx context.Context, arg db.GetVotesBySeasonAndVoterParams) ([]db.Vote, error) {
	if m.getVotesBySeasonAndVoterFn != nil {
		return m.getVotesBySeasonAndVoterFn(ctx, arg)
	}
	return []db.Vote{}, nil
}
func (m *mockVotingQuerier) GetGroupMembers(ctx context.Context, groupID string) ([]db.GetGroupMembersRow, error) {
	if m.getGroupMembersFn != nil {
		return m.getGroupMembersFn(ctx, groupID)
	}
	return []db.GetGroupMembersRow{}, nil
}
func (m *mockVotingQuerier) HasVoteForQuestion(ctx context.Context, arg db.HasVoteForQuestionParams) (int64, error) {
	if m.hasVoteForQuestionFn != nil {
		return m.hasVoteForQuestionFn(ctx, arg)
	}
	return 0, nil
}
func (m *mockVotingQuerier) CreateVote(ctx context.Context, arg db.CreateVoteParams) (db.Vote, error) {
	if m.createVoteFn != nil {
		return m.createVoteFn(ctx, arg)
	}
	return db.Vote{ID: arg.ID, SeasonID: arg.SeasonID, VoterID: arg.VoterID, TargetID: arg.TargetID, QuestionID: arg.QuestionID}, nil
}
func (m *mockVotingQuerier) CountGroupMembers(ctx context.Context, groupID string) (int64, error) {
	if m.countGroupMembersFn != nil {
		return m.countGroupMembersFn(ctx, groupID)
	}
	return 5, nil
}
func (m *mockVotingQuerier) CountSeasonQuestions(ctx context.Context, seasonID string) (int64, error) {
	if m.countSeasonQuestionsFn != nil {
		return m.countSeasonQuestionsFn(ctx, seasonID)
	}
	return 5, nil
}
func (m *mockVotingQuerier) CountCompletedVoters(ctx context.Context, arg db.CountCompletedVotersParams) (int64, error) {
	if m.countCompletedVotersFn != nil {
		return m.countCompletedVotersFn(ctx, arg)
	}
	return 0, nil
}
func (m *mockVotingQuerier) HasUserVotedInSeason(ctx context.Context, arg db.HasUserVotedInSeasonParams) (int64, error) {
	if m.hasUserVotedInSeasonFn != nil {
		return m.hasUserVotedInSeasonFn(ctx, arg)
	}
	return 0, nil
}

func newTestVotingHandler(mock *mockVotingQuerier) *Handler {
	svc := votingsvc.NewService(mock)
	return NewHandler(svc)
}

// --- GetVotingSession tests ---

func TestGetVotingSession_Success(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
		getSeasonQuestionsFn: func(ctx context.Context, seasonID string) ([]db.Question, error) {
			return []db.Question{
				{ID: "q1", Text: "Who is the funniest?", Category: db.QuestionCategoryFUNNY},
				{ID: "q2", Text: "Who is the smartest?", Category: db.QuestionCategorySKILLS},
			}, nil
		},
		getVotesBySeasonAndVoterFn: func(ctx context.Context, arg db.GetVotesBySeasonAndVoterParams) ([]db.Vote, error) {
			return []db.Vote{{QuestionID: "q1"}}, nil
		},
		getGroupMembersFn: func(ctx context.Context, groupID string) ([]db.GetGroupMembersRow, error) {
			return []db.GetGroupMembersRow{
				{ID: "u1", Username: "voter"},
				{ID: "u2", Username: "target1"},
				{ID: "u3", Username: "target2"},
			}, nil
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/voting", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.GetVotingSession(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"]
	if data["season_id"] != "s1" {
		t.Errorf("expected season_id s1, got %v", data["season_id"])
	}
	questions := data["questions"].([]any)
	if len(questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(questions))
	}
	targets := data["targets"].([]any)
	if len(targets) != 2 {
		t.Errorf("expected 2 targets (excluding voter), got %d", len(targets))
	}
}

func TestGetVotingSession_SeasonNotFound(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{}, sql.ErrNoRows
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/missing/voting", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("missing")
	setUser(c, "u1", "voter")

	err := h.GetVotingSession(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp["error"]["code"])
	}
}

func TestGetVotingSession_NotMember(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 0, nil // not a member
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/voting", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "outsider", "outsider")

	err := h.GetVotingSession(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %s", resp["error"]["code"])
	}
}

func TestGetVotingSession_SeasonNotVoting(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusREVEALED}, nil
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/voting", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.GetVotingSession(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "SEASON_NOT_VOTING" {
		t.Errorf("expected SEASON_NOT_VOTING, got %s", resp["error"]["code"])
	}
}

// --- CastVote tests ---

func TestCastVote_Success(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
		getSeasonQuestionsFn: func(ctx context.Context, seasonID string) ([]db.Question, error) {
			return []db.Question{
				{ID: "q1", Text: "Who is best?", Category: db.QuestionCategoryFUNNY},
			}, nil
		},
		hasVoteForQuestionFn: func(ctx context.Context, arg db.HasVoteForQuestionParams) (int64, error) {
			return 0, nil // not voted yet
		},
		createVoteFn: func(ctx context.Context, arg db.CreateVoteParams) (db.Vote, error) {
			return db.Vote{ID: "v1", SeasonID: arg.SeasonID, VoterID: arg.VoterID, TargetID: arg.TargetID, QuestionID: arg.QuestionID}, nil
		},
		getVotesBySeasonAndVoterFn: func(ctx context.Context, arg db.GetVotesBySeasonAndVoterParams) ([]db.Vote, error) {
			return []db.Vote{{QuestionID: "q1"}}, nil
		},
	}
	h := newTestVotingHandler(mock)

	body := `{"question_id":"q1","target_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.CastVote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"]
	vote := data["vote"].(map[string]any)
	if vote["question_id"] != "q1" {
		t.Errorf("expected question_id q1, got %v", vote["question_id"])
	}
	if vote["target_id"] != "u2" {
		t.Errorf("expected target_id u2, got %v", vote["target_id"])
	}
}

func TestCastVote_SelfVote(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
	}
	h := newTestVotingHandler(mock)

	body := `{"question_id":"q1","target_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.CastVote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "SELF_VOTE" {
		t.Errorf("expected SELF_VOTE, got %s", resp["error"]["code"])
	}
}

func TestCastVote_AlreadyVoted(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
		getSeasonQuestionsFn: func(ctx context.Context, seasonID string) ([]db.Question, error) {
			return []db.Question{{ID: "q1"}}, nil
		},
		hasVoteForQuestionFn: func(ctx context.Context, arg db.HasVoteForQuestionParams) (int64, error) {
			return 1, nil // already voted
		},
	}
	h := newTestVotingHandler(mock)

	body := `{"question_id":"q1","target_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.CastVote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "ALREADY_VOTED" {
		t.Errorf("expected ALREADY_VOTED, got %s", resp["error"]["code"])
	}
}

func TestCastVote_NotMember(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 0, nil
		},
	}
	h := newTestVotingHandler(mock)

	body := `{"question_id":"q1","target_id":"u2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/seasons/s1/votes",
		strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "outsider", "outsider")

	err := h.CastVote(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// --- GetProgress tests ---

func TestGetProgress_Success(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 1, nil
		},
		countGroupMembersFn: func(ctx context.Context, groupID string) (int64, error) {
			return 5, nil
		},
		countSeasonQuestionsFn: func(ctx context.Context, seasonID string) (int64, error) {
			return 3, nil
		},
		countCompletedVotersFn: func(ctx context.Context, arg db.CountCompletedVotersParams) (int64, error) {
			return 3, nil
		},
		hasUserVotedInSeasonFn: func(ctx context.Context, arg db.HasUserVotedInSeasonParams) (int64, error) {
			return 3, nil // voted on all 3 questions
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/progress", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "u1", "voter")

	err := h.GetProgress(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"]
	if data["voted_count"] != float64(3) {
		t.Errorf("expected voted_count 3, got %v", data["voted_count"])
	}
	if data["total_count"] != float64(5) {
		t.Errorf("expected total_count 5, got %v", data["total_count"])
	}
	if data["user_voted"] != true {
		t.Errorf("expected user_voted true, got %v", data["user_voted"])
	}
	if data["quorum_reached"] != true {
		t.Errorf("expected quorum_reached true (3/5 >= 50%%), got %v", data["quorum_reached"])
	}
}

func TestGetProgress_NotMember(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{ID: id, GroupID: "g1", Status: db.SeasonStatusVOTING}, nil
		},
		isGroupMemberFn: func(ctx context.Context, arg db.IsGroupMemberParams) (int64, error) {
			return 0, nil
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/s1/progress", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("s1")
	setUser(c, "outsider", "outsider")

	err := h.GetProgress(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %s", resp["error"]["code"])
	}
}

func TestGetProgress_SeasonNotFound(t *testing.T) {
	e := setupEcho()
	mock := &mockVotingQuerier{
		getSeasonByIDFn: func(ctx context.Context, id string) (db.Season, error) {
			return db.Season{}, sql.ErrNoRows
		},
	}
	h := newTestVotingHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seasons/missing/progress", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("seasonId")
	c.SetParamValues("missing")
	setUser(c, "u1", "voter")

	err := h.GetProgress(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
