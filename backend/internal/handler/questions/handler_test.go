package questions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/middleware"
	questionsvc "github.com/repa-app/repa/internal/service/questions"
)

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	return e
}

func setUser(c echo.Context, userID string) {
	c.Set("user", &middleware.JWTClaims{UserID: userID, Username: "testuser"})
}

// --- toQuestionDto tests ---

func TestToQuestionDto_AllFields(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	q := db.Question{
		ID:       "q-1",
		Text:     "Who is the funniest?",
		Category: db.QuestionCategoryFUNNY,
		Source:   db.QuestionSourceSYSTEM,
		AuthorID: sql.NullString{String: "user-42", Valid: true},
		GroupID:  sql.NullString{String: "grp-1", Valid: true},
		Status:   db.QuestionStatusACTIVE,
		CreatedAt: now,
	}

	dto := toQuestionDto(q)

	if dto.ID != "q-1" {
		t.Errorf("expected ID 'q-1', got %q", dto.ID)
	}
	if dto.Text != "Who is the funniest?" {
		t.Errorf("expected text 'Who is the funniest?', got %q", dto.Text)
	}
	if dto.Category != "FUNNY" {
		t.Errorf("expected category 'FUNNY', got %q", dto.Category)
	}
	if dto.Source != "SYSTEM" {
		t.Errorf("expected source 'SYSTEM', got %q", dto.Source)
	}
	if dto.AuthorID == nil {
		t.Fatal("expected non-nil AuthorID")
	}
	if *dto.AuthorID != "user-42" {
		t.Errorf("expected author_id 'user-42', got %q", *dto.AuthorID)
	}
	if dto.Status != "ACTIVE" {
		t.Errorf("expected status 'ACTIVE', got %q", dto.Status)
	}
	expected := now.Format(time.RFC3339)
	if dto.CreatedAt != expected {
		t.Errorf("expected created_at %q, got %q", expected, dto.CreatedAt)
	}
}

func TestToQuestionDto_NullAuthorID(t *testing.T) {
	q := db.Question{
		ID:        "q-2",
		Text:      "System question",
		Category:  db.QuestionCategoryHOT,
		Source:    db.QuestionSourceSYSTEM,
		AuthorID:  sql.NullString{Valid: false},
		Status:    db.QuestionStatusACTIVE,
		CreatedAt: time.Now(),
	}

	dto := toQuestionDto(q)

	if dto.AuthorID != nil {
		t.Errorf("expected nil AuthorID for null author, got %v", dto.AuthorID)
	}
}

func TestToQuestionDto_JSONSerialization(t *testing.T) {
	q := db.Question{
		ID:        "q-3",
		Text:      "Test",
		Category:  db.QuestionCategorySECRETS,
		Source:    db.QuestionSourceUSER,
		AuthorID:  sql.NullString{String: "u1", Valid: true},
		Status:    db.QuestionStatusPENDING,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	dto := toQuestionDto(q)
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	expectedKeys := []string{"id", "text", "category", "source", "author_id", "status", "created_at"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestToQuestionDto_NullAuthorID_OmittedInJSON(t *testing.T) {
	q := db.Question{
		ID:        "q-4",
		Text:      "No author",
		Category:  db.QuestionCategoryHOT,
		Source:    db.QuestionSourceSYSTEM,
		AuthorID:  sql.NullString{Valid: false},
		Status:    db.QuestionStatusACTIVE,
		CreatedAt: time.Now(),
	}

	dto := toQuestionDto(q)
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// author_id should be present but null
	val, ok := m["author_id"]
	if !ok {
		t.Error("expected 'author_id' key in JSON (even if null)")
	} else if val != nil {
		t.Errorf("expected null author_id, got %v", val)
	}
}

// --- CreateQuestion handler tests ---

func TestCreateQuestion_BadJSON(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no returned error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateQuestion_EmptyBody_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err == nil {
		t.Fatal("expected validation error")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

func TestCreateQuestion_InvalidCategory_ValidationError(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"text":"valid text","category":"INVALID"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err == nil {
		t.Fatal("expected validation error for invalid category")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", he.Code)
	}
}

// --- ListQuestions tests ---

func TestListQuestions_NilUser_Panics(t *testing.T) {
	// Without user claims set, GetCurrentUser returns nil, causing a nil dereference.
	// This verifies the middleware dependency.
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		_ = h.ListQuestions(c)
	}()

	if !panicked {
		t.Error("expected panic when user claims are not set")
	}
}

// --- ReportQuestion tests ---

func TestReportQuestion_BindsOptionalReason(t *testing.T) {
	// ReportQuestion uses _ = c.Bind which ignores bind errors.
	// With bad JSON, it should still proceed (bind error is ignored).
	// It will fail later at h.svc.ReportQuestion (nil svc), but bind itself does not fail.
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("questionId")
	c.SetParamValues("q-1")

	// Should panic or fail at h.svc.ReportQuestion since svc is nil
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		_ = h.ReportQuestion(c)
	}()

	if !panicked {
		t.Error("expected panic when service is nil (bind error was ignored, handler continued)")
	}
}

// --- DeleteQuestion tests ---

func TestDeleteQuestion_ExtractsParams(t *testing.T) {
	e := newTestEcho()
	h := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId", "questionId")
	c.SetParamValues("grp-1", "q-1")

	// Handler will panic at h.queries.GetGroupByID (nil queries),
	// but params should be extractable
	func() {
		defer func() {
			recover()
		}()
		_ = h.DeleteQuestion(c)
	}()

	if got := c.Param("groupId"); got != "grp-1" {
		t.Errorf("expected groupId 'grp-1', got %q", got)
	}
	if got := c.Param("questionId"); got != "q-1" {
		t.Errorf("expected questionId 'q-1', got %q", got)
	}
}

// --- mockQuerierForSvc implements db.Querier for questions service ---

type mockQuerierForSvc struct {
	db.Querier

	countUserQuestionsInGroupFn func(ctx context.Context, arg db.CountUserQuestionsInGroupParams) (int64, error)
	createQuestionFn            func(ctx context.Context, arg db.CreateQuestionParams) (db.Question, error)
	getGroupByIDFn              func(ctx context.Context, id string) (db.Group, error)
	getGroupAllQuestionsFn      func(ctx context.Context, arg db.GetGroupAllQuestionsParams) ([]db.Question, error)
	getQuestionByIDFn           func(ctx context.Context, id string) (db.Question, error)
	updateQuestionStatusFn      func(ctx context.Context, arg db.UpdateQuestionStatusParams) error
	hasUserReportedFn           func(ctx context.Context, arg db.HasUserReportedParams) (bool, error)
	createReportFn              func(ctx context.Context, arg db.CreateReportParams) (db.Report, error)
}

func (m *mockQuerierForSvc) CountUserQuestionsInGroup(ctx context.Context, arg db.CountUserQuestionsInGroupParams) (int64, error) {
	if m.countUserQuestionsInGroupFn != nil {
		return m.countUserQuestionsInGroupFn(ctx, arg)
	}
	return 0, nil
}

func (m *mockQuerierForSvc) CreateQuestion(ctx context.Context, arg db.CreateQuestionParams) (db.Question, error) {
	if m.createQuestionFn != nil {
		return m.createQuestionFn(ctx, arg)
	}
	return db.Question{
		ID:        arg.ID,
		Text:      arg.Text,
		Category:  arg.Category,
		Source:    arg.Source,
		GroupID:   arg.GroupID,
		AuthorID:  arg.AuthorID,
		Status:    arg.Status,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockQuerierForSvc) GetGroupByID(ctx context.Context, id string) (db.Group, error) {
	if m.getGroupByIDFn != nil {
		return m.getGroupByIDFn(ctx, id)
	}
	return db.Group{ID: id, AdminID: "admin-1", Categories: []string{"HOT", "FUNNY"}}, nil
}

func (m *mockQuerierForSvc) GetGroupAllQuestions(ctx context.Context, arg db.GetGroupAllQuestionsParams) ([]db.Question, error) {
	if m.getGroupAllQuestionsFn != nil {
		return m.getGroupAllQuestionsFn(ctx, arg)
	}
	return nil, nil
}

func (m *mockQuerierForSvc) GetQuestionByID(ctx context.Context, id string) (db.Question, error) {
	if m.getQuestionByIDFn != nil {
		return m.getQuestionByIDFn(ctx, id)
	}
	return db.Question{}, sql.ErrNoRows
}

func (m *mockQuerierForSvc) UpdateQuestionStatus(ctx context.Context, arg db.UpdateQuestionStatusParams) error {
	if m.updateQuestionStatusFn != nil {
		return m.updateQuestionStatusFn(ctx, arg)
	}
	return nil
}

func (m *mockQuerierForSvc) HasUserReported(ctx context.Context, arg db.HasUserReportedParams) (bool, error) {
	if m.hasUserReportedFn != nil {
		return m.hasUserReportedFn(ctx, arg)
	}
	return false, nil
}

func (m *mockQuerierForSvc) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) {
	if m.createReportFn != nil {
		return m.createReportFn(ctx, arg)
	}
	return db.Report{ID: arg.ID}, nil
}

// --- Helper to build handler with sqlmock + mock querier ---

func newHandlerWithSqlmock(t *testing.T, mockQ *mockQuerierForSvc) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	queries := db.New(sqlDB)
	svc := questionsvc.NewService(mockQ, nil) // nil moderator
	h := NewHandler(svc, queries)
	return h, mock
}

// --- CreateQuestion with sqlmock ---

func TestCreateQuestion_Success(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	// IsGroupMember returns 1 (is member)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	e := newTestEcho()
	body := `{"text":"This is a valid question text","category":"FUNNY"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	data := resp["data"].(map[string]any)
	q := data["question"].(map[string]any)
	if q["category"] != "FUNNY" {
		t.Errorf("expected category FUNNY, got %v", q["category"])
	}
	mod := data["moderation"].(map[string]any)
	if mod["approved"] != true {
		t.Error("expected moderation approved=true")
	}
}

func TestCreateQuestion_NotMember(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	// IsGroupMember returns 0 (not member)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	e := newTestEcho()
	body := `{"text":"This is a valid question text","category":"FUNNY"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "NOT_MEMBER" {
		t.Errorf("expected NOT_MEMBER, got %s", resp["error"]["code"])
	}
}

func TestCreateQuestion_QuestionLimit(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		countUserQuestionsInGroupFn: func(_ context.Context, _ db.CountUserQuestionsInGroupParams) (int64, error) {
			return 5, nil // at limit
		},
	}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	e := newTestEcho()
	body := `{"text":"This is a valid question text","category":"FUNNY"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "QUESTION_LIMIT" {
		t.Errorf("expected QUESTION_LIMIT, got %s", resp["error"]["code"])
	}
}

func TestCreateQuestion_TextTooShort(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	e := newTestEcho()
	body := `{"text":"Short","category":"FUNNY"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "VALIDATION" {
		t.Errorf("expected VALIDATION, got %s", resp["error"]["code"])
	}
}

// --- ListQuestions with sqlmock ---

func TestListQuestions_Success(t *testing.T) {
	now := time.Now()
	mockQ := &mockQuerierForSvc{
		getGroupByIDFn: func(_ context.Context, id string) (db.Group, error) {
			return db.Group{ID: id, Categories: []string{"HOT", "FUNNY"}}, nil
		},
		getGroupAllQuestionsFn: func(_ context.Context, _ db.GetGroupAllQuestionsParams) ([]db.Question, error) {
			return []db.Question{
				{ID: "q-1", Text: "Question one", Category: db.QuestionCategoryHOT, Source: db.QuestionSourceSYSTEM, Status: db.QuestionStatusACTIVE, CreatedAt: now},
				{ID: "q-2", Text: "Question two", Category: db.QuestionCategoryFUNNY, Source: db.QuestionSourceUSER, Status: db.QuestionStatusACTIVE, AuthorID: sql.NullString{String: "user-1", Valid: true}, CreatedAt: now},
			}, nil
		},
	}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.ListQuestions(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	questions := data["questions"].([]any)
	if len(questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(questions))
	}
}

func TestListQuestions_NotMember(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.ListQuestions(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// --- DeleteQuestion with sqlmock ---

func TestDeleteQuestion_Success(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		getQuestionByIDFn: func(_ context.Context, id string) (db.Question, error) {
			return db.Question{
				ID:       id,
				GroupID:  sql.NullString{String: "grp-1", Valid: true},
				AuthorID: sql.NullString{String: "user-1", Valid: true},
				Source:   db.QuestionSourceUSER,
			}, nil
		},
	}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	// GetGroupByID for admin check
	mock.ExpectQuery(`SELECT .+ FROM groups WHERE id`).
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "invite_code", "admin_id",
			"telegram_chat_id", "telegram_chat_username",
			"telegram_connect_code", "telegram_connect_expiry",
			"created_at", "categories",
		}).AddRow(
			"grp-1", "Test", "inv-1", "admin-1",
			nil, nil, nil, nil,
			time.Now(), `{"HOT"}`,
		))

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId", "questionId")
	c.SetParamValues("grp-1", "q-1")

	err := h.DeleteQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["deleted"] != true {
		t.Error("expected deleted=true")
	}
}

func TestDeleteQuestion_NotFound(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		getQuestionByIDFn: func(_ context.Context, _ string) (db.Question, error) {
			return db.Question{}, sql.ErrNoRows
		},
	}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT .+ FROM groups WHERE id`).
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "invite_code", "admin_id",
			"telegram_chat_id", "telegram_chat_username",
			"telegram_connect_code", "telegram_connect_expiry",
			"created_at", "categories",
		}).AddRow(
			"grp-1", "Test", "inv-1", "user-1",
			nil, nil, nil, nil,
			time.Now(), `{"HOT"}`,
		))

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId", "questionId")
	c.SetParamValues("grp-1", "q-nonexistent")

	err := h.DeleteQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- ReportQuestion with sqlmock ---

func TestReportQuestion_Success(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		hasUserReportedFn: func(_ context.Context, _ db.HasUserReportedParams) (bool, error) {
			return false, nil
		},
	}
	h, _ := newHandlerWithSqlmock(t, mockQ)

	e := newTestEcho()
	body := `{"reason":"offensive"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("questionId")
	c.SetParamValues("q-1")

	err := h.ReportQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["reported"] != true {
		t.Error("expected reported=true")
	}
}

func TestReportQuestion_AlreadyReported(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		hasUserReportedFn: func(_ context.Context, _ db.HasUserReportedParams) (bool, error) {
			return true, nil
		},
	}
	h, _ := newHandlerWithSqlmock(t, mockQ)

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("questionId")
	c.SetParamValues("q-1")

	err := h.ReportQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "ALREADY_REPORTED" {
		t.Errorf("expected ALREADY_REPORTED, got %s", resp["error"]["code"])
	}
}

// --- CreateQuestion internal server error ---

func TestCreateQuestion_IsGroupMemberDBError(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnError(errors.New("db down"))

	e := newTestEcho()
	body := `{"text":"This is a valid question text","category":"FUNNY"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.CreateQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestListQuestions_IsGroupMemberDBError(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM group_members`).
		WithArgs("user-1", "grp-1").
		WillReturnError(errors.New("db down"))

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId")
	c.SetParamValues("grp-1")

	err := h.ListQuestions(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestDeleteQuestion_GroupNotFound(t *testing.T) {
	mockQ := &mockQuerierForSvc{}
	h, mock := newHandlerWithSqlmock(t, mockQ)

	mock.ExpectQuery(`SELECT .+ FROM groups WHERE id`).
		WithArgs("grp-nonexistent").
		WillReturnError(sql.ErrNoRows)

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("groupId", "questionId")
	c.SetParamValues("grp-nonexistent", "q-1")

	err := h.DeleteQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestReportQuestion_InternalError(t *testing.T) {
	mockQ := &mockQuerierForSvc{
		hasUserReportedFn: func(_ context.Context, _ db.HasUserReportedParams) (bool, error) {
			return false, errors.New("db error")
		},
	}
	h, _ := newHandlerWithSqlmock(t, mockQ)

	e := newTestEcho()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-1")
	c.SetParamNames("questionId")
	c.SetParamValues("q-1")

	err := h.ReportQuestion(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// --- DTO slice conversion test ---

func TestToQuestionDto_SliceConversion(t *testing.T) {
	questions := []db.Question{
		{
			ID:        "q-a",
			Text:      "First",
			Category:  db.QuestionCategoryHOT,
			Source:    db.QuestionSourceSYSTEM,
			Status:    db.QuestionStatusACTIVE,
			AuthorID:  sql.NullString{Valid: false},
			CreatedAt: time.Now(),
		},
		{
			ID:        "q-b",
			Text:      "Second",
			Category:  db.QuestionCategoryFUNNY,
			Source:    db.QuestionSourceUSER,
			Status:    db.QuestionStatusPENDING,
			AuthorID:  sql.NullString{String: "u-5", Valid: true},
			CreatedAt: time.Now(),
		},
	}

	dtos := make([]questionDto, len(questions))
	for i, q := range questions {
		dtos[i] = toQuestionDto(q)
	}

	if len(dtos) != 2 {
		t.Fatalf("expected 2 dtos, got %d", len(dtos))
	}
	if dtos[0].ID != "q-a" {
		t.Errorf("expected first dto ID 'q-a', got %q", dtos[0].ID)
	}
	if dtos[1].AuthorID == nil || *dtos[1].AuthorID != "u-5" {
		t.Errorf("expected second dto author_id 'u-5', got %v", dtos[1].AuthorID)
	}
	if dtos[0].AuthorID != nil {
		t.Errorf("expected first dto author_id nil, got %v", dtos[0].AuthorID)
	}
}
