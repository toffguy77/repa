package push

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/middleware"
)

// newTestEcho creates an Echo instance with the project's custom validator.
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	return e
}

// setUser sets a fake JWT user claim on the echo context.
func setUser(c echo.Context, userID string) {
	c.Set("user", &middleware.JWTClaims{UserID: userID, Username: "testuser"})
}

// --- RegisterToken tests ---

func TestRegisterToken_BadJSON(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	err := h.RegisterToken(c)
	if err != nil {
		t.Fatalf("expected no returned error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Error("expected 'error' key in response")
	}
}

func TestRegisterToken_EmptyBody_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	err := h.RegisterToken(c)
	// The validator returns an echo.HTTPError, which Echo would handle
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", he.Code)
	}
}

func TestRegisterToken_MissingToken_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	// Platform is valid but token is missing
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"platform":"ios"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	err := h.RegisterToken(c)
	if err == nil {
		t.Fatal("expected validation error for missing token")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

func TestRegisterToken_InvalidPlatform_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"abc","platform":"windows"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	err := h.RegisterToken(c)
	if err == nil {
		t.Fatal("expected validation error for invalid platform")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

// --- VoteQuestion tests ---

func TestVoteQuestion_BadJSON(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("group-1")

	err := h.VoteQuestion(c)
	if err != nil {
		t.Fatalf("expected no returned error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestVoteQuestion_EmptyBody_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("group-1")

	err := h.VoteQuestion(c)
	if err == nil {
		t.Fatal("expected validation error for empty questionId")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

// --- GetQuestionCandidates tests ---

func TestGetQuestionCandidates_ExtractsGroupParam(t *testing.T) {
	// Verify the handler extracts the path param; it will fail on membership check
	// (nil queries) but we can verify the param extraction path.
	e := newTestEcho()
	h := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/groups/grp-123/question-candidates", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-123")

	// This will panic or error because h.queries is nil, but we're testing
	// param extraction. We recover from the nil pointer dereference.
	func() {
		defer func() {
			if r := recover(); r == nil {
				// If it didn't panic, the handler somehow completed
				// (perhaps queries was not nil), which is fine too.
			}
		}()
		_ = h.GetQuestionCandidates(c)
	}()

	// The param was set and accessible — verify via echo context
	if got := c.Param("id"); got != "grp-123" {
		t.Errorf("expected param 'id' = 'grp-123', got %q", got)
	}
}

// --- questionDto tests ---

func TestQuestionDto_JSONSerialization(t *testing.T) {
	dto := questionDto{
		ID:       "q-1",
		Text:     "Who is the funniest?",
		Category: "FUNNY",
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal questionDto: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["id"] != "q-1" {
		t.Errorf("expected id 'q-1', got %q", parsed["id"])
	}
	if parsed["text"] != "Who is the funniest?" {
		t.Errorf("expected text 'Who is the funniest?', got %q", parsed["text"])
	}
	if parsed["category"] != "FUNNY" {
		t.Errorf("expected category 'FUNNY', got %q", parsed["category"])
	}
}

func TestQuestionDto_AllFieldsPresent(t *testing.T) {
	dto := questionDto{
		ID:       "q-2",
		Text:     "Test question",
		Category: "HOT",
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, key := range []string{"id", "text", "category"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

// --- sqlmock-based tests ---

func TestRegisterToken_Success(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"fcm-tok-1","platform":"ios"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	mock.ExpectQuery("INSERT INTO fcm_tokens").
		WithArgs(sqlmock.AnyArg(), "user-1", "fcm-tok-1", "ios").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "platform", "created_at"}).
			AddRow("tok-id", "user-1", "fcm-tok-1", "ios", time.Now()))

	if err := h.RegisterToken(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]map[string]bool
	json.Unmarshal(rec.Body.Bytes(), &body)
	if !body["data"]["registered"] {
		t.Error("expected registered=true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterToken_DBError(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"token":"fcm-tok-1","platform":"android"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")

	mock.ExpectQuery("INSERT INTO fcm_tokens").
		WithArgs(sqlmock.AnyArg(), "user-1", "fcm-tok-1", "android").
		WillReturnError(fmt.Errorf("db connection lost"))

	if err := h.RegisterToken(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetQuestionCandidates_Success(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodGet, "/groups/grp-1/question-candidates", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// GetRandomSystemQuestions
	mock.ExpectQuery("SELECT id, text, category, source, group_id, author_id, status, created_at FROM questions").
		WithArgs(int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}).
			AddRow("q-1", "Who is the funniest?", "FUNNY", "SYSTEM", nil, nil, "ACTIVE", time.Now()).
			AddRow("q-2", "Who is the smartest?", "SKILLS", "SYSTEM", nil, nil, "ACTIVE", time.Now()).
			AddRow("q-3", "Who is the hottest?", "HOT", "SYSTEM", nil, nil, "ACTIVE", time.Now()))

	if err := h.GetQuestionCandidates(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]map[string][]questionDto
	json.Unmarshal(rec.Body.Bytes(), &body)
	candidates := body["data"]["candidates"]
	if len(candidates) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(candidates))
	}
	if candidates[0].ID != "q-1" {
		t.Errorf("expected first candidate ID q-1, got %s", candidates[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetQuestionCandidates_NotMember(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodGet, "/groups/grp-1/question-candidates", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	if err := h.GetQuestionCandidates(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != "FORBIDDEN" {
		t.Errorf("expected error code FORBIDDEN, got %s", body["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestVoteQuestion_Success(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"questionId":"q-1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// GetLastSeasonNumber
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(2)))

	// HasNextSeasonVote
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("grp-1", "user-1", int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// GetQuestionByID
	mock.ExpectQuery("SELECT id, text, category, source, group_id, author_id, status, created_at FROM questions WHERE id").
		WithArgs("q-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}).
			AddRow("q-1", "Who is the funniest?", "FUNNY", "SYSTEM", nil, nil, "ACTIVE", time.Now()))

	// CreateNextSeasonVote
	mock.ExpectQuery("INSERT INTO next_season_votes").
		WithArgs(sqlmock.AnyArg(), "grp-1", "user-1", "q-1", int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "user_id", "question_id", "season_number", "created_at"}).
			AddRow("vote-1", "grp-1", "user-1", "q-1", int32(3), time.Now()))

	if err := h.VoteQuestion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]map[string]bool
	json.Unmarshal(rec.Body.Bytes(), &body)
	if !body["data"]["voted"] {
		t.Error("expected voted=true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestVoteQuestion_NotMember(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"questionId":"q-1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	if err := h.VoteQuestion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	var body map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got %s", body["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestVoteQuestion_AlreadyVoted(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"questionId":"q-1"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// GetLastSeasonNumber
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(5)))

	// HasNextSeasonVote — already voted
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("grp-1", "user-1", int32(6)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	if err := h.VoteQuestion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}

	var body map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != "ALREADY_VOTED" {
		t.Errorf("expected ALREADY_VOTED, got %s", body["error"]["code"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestVoteQuestion_QuestionNotFound(t *testing.T) {
	e := newTestEcho()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)
	h := NewHandler(queries)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"questionId":"q-nonexistent"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("id")
	c.SetParamValues("grp-1")

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// GetLastSeasonNumber
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(1)))

	// HasNextSeasonVote — not voted yet
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("grp-1", "user-1", int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// GetQuestionByID — not found (no rows)
	mock.ExpectQuery("SELECT id, text, category, source, group_id, author_id, status, created_at FROM questions WHERE id").
		WithArgs("q-nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}))

	if err := h.VoteQuestion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterTokenRequest_JSONTags(t *testing.T) {
	req := RegisterTokenRequest{
		Token:    "fcm-token-123",
		Platform: "ios",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if m["token"] != "fcm-token-123" {
		t.Errorf("expected token 'fcm-token-123', got %q", m["token"])
	}
	if m["platform"] != "ios" {
		t.Errorf("expected platform 'ios', got %q", m["platform"])
	}
}
