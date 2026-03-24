package admin

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

	"github.com/labstack/echo/v4"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/middleware"
)

// mockQuerier implements the subset of db.Querier used by admin.Handler.
type mockQuerier struct {
	db.Querier

	listReportsFunc          func(ctx context.Context, arg db.ListReportsParams) ([]db.ListReportsRow, error)
	countReportsFunc         func(ctx context.Context) (int64, error)
	getReportByIDFunc        func(ctx context.Context, id string) (db.GetReportByIDRow, error)
	updateQuestionStatusFunc func(ctx context.Context, arg db.UpdateQuestionStatusParams) error
	countActiveUsers7DFunc   func(ctx context.Context) (int64, error)
	countActiveUsers30DFunc  func(ctx context.Context) (int64, error)
	countGroupsFunc          func(ctx context.Context) (int64, error)
	sumRevenue7DaysFunc      func(ctx context.Context) (int64, error)
}

func (m *mockQuerier) ListReports(ctx context.Context, arg db.ListReportsParams) ([]db.ListReportsRow, error) {
	if m.listReportsFunc != nil {
		return m.listReportsFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockQuerier) CountReports(ctx context.Context) (int64, error) {
	if m.countReportsFunc != nil {
		return m.countReportsFunc(ctx)
	}
	return 0, nil
}

func (m *mockQuerier) GetReportByID(ctx context.Context, id string) (db.GetReportByIDRow, error) {
	if m.getReportByIDFunc != nil {
		return m.getReportByIDFunc(ctx, id)
	}
	return db.GetReportByIDRow{}, nil
}

func (m *mockQuerier) UpdateQuestionStatus(ctx context.Context, arg db.UpdateQuestionStatusParams) error {
	if m.updateQuestionStatusFunc != nil {
		return m.updateQuestionStatusFunc(ctx, arg)
	}
	return nil
}

func (m *mockQuerier) CountActiveUsers7Days(ctx context.Context) (int64, error) {
	if m.countActiveUsers7DFunc != nil {
		return m.countActiveUsers7DFunc(ctx)
	}
	return 0, nil
}

func (m *mockQuerier) CountActiveUsers30Days(ctx context.Context) (int64, error) {
	if m.countActiveUsers30DFunc != nil {
		return m.countActiveUsers30DFunc(ctx)
	}
	return 0, nil
}

func (m *mockQuerier) CountGroups(ctx context.Context) (int64, error) {
	if m.countGroupsFunc != nil {
		return m.countGroupsFunc(ctx)
	}
	return 0, nil
}

func (m *mockQuerier) SumRevenue7Days(ctx context.Context) (int64, error) {
	if m.sumRevenue7DaysFunc != nil {
		return m.sumRevenue7DaysFunc(ctx)
	}
	return 0, nil
}

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	return e
}

// --- BasicAuth Tests ---

func TestBasicAuth_NoPasswordConfigured(t *testing.T) {
	h := NewHandler(&mockQuerier{}, "admin", "")
	e := setupEcho()

	called := false
	handler := h.BasicAuth(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if called {
		t.Error("next handler should not have been called")
	}
}

func TestBasicAuth_MissingCredentials(t *testing.T) {
	h := NewHandler(&mockQuerier{}, "admin", "secret")
	e := setupEcho()

	handler := h.BasicAuth(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

func TestBasicAuth_WrongCredentials(t *testing.T) {
	h := NewHandler(&mockQuerier{}, "admin", "secret")
	e := setupEcho()

	handler := h.BasicAuth(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestBasicAuth_CorrectCredentials(t *testing.T) {
	h := NewHandler(&mockQuerier{}, "admin", "secret")
	e := setupEcho()

	called := false
	handler := h.BasicAuth(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("next handler should have been called")
	}
}

// --- ListReports Tests ---

func TestListReports_SuccessWithDefaults(t *testing.T) {
	now := time.Now()
	mock := &mockQuerier{
		listReportsFunc: func(_ context.Context, arg db.ListReportsParams) ([]db.ListReportsRow, error) {
			if arg.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected default offset 0, got %d", arg.Offset)
			}
			return []db.ListReportsRow{
				{
					ID:               "r1",
					QuestionID:       "q1",
					ReporterID:       "u1",
					Reason:           sql.NullString{String: "offensive", Valid: true},
					CreatedAt:        now,
					QuestionText:     "Bad question?",
					QuestionCategory: db.QuestionCategoryFUNNY,
					QuestionStatus:   db.QuestionStatusPENDING,
					ReporterUsername:  "alice",
				},
			}, nil
		},
		countReportsFunc: func(_ context.Context) (int64, error) {
			return 1, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/reports", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListReports(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)

	data, ok := resp["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("expected 1 report in data, got %v", resp["data"])
	}

	meta := resp["meta"].(map[string]any)
	if meta["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", meta["total"])
	}
	if meta["page"].(float64) != 1 {
		t.Errorf("expected page 1, got %v", meta["page"])
	}
	if meta["limit"].(float64) != 20 {
		t.Errorf("expected limit 20, got %v", meta["limit"])
	}
}

func TestListReports_CustomPageLimit(t *testing.T) {
	mock := &mockQuerier{
		listReportsFunc: func(_ context.Context, arg db.ListReportsParams) ([]db.ListReportsRow, error) {
			if arg.Limit != 50 {
				t.Errorf("expected limit 50, got %d", arg.Limit)
			}
			if arg.Offset != 100 {
				t.Errorf("expected offset 100 ((3-1)*50), got %d", arg.Offset)
			}
			return []db.ListReportsRow{}, nil
		},
		countReportsFunc: func(_ context.Context) (int64, error) {
			return 150, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/reports?page=3&limit=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListReports(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)

	meta := resp["meta"].(map[string]any)
	if meta["page"].(float64) != 3 {
		t.Errorf("expected page 3, got %v", meta["page"])
	}
	if meta["limit"].(float64) != 50 {
		t.Errorf("expected limit 50, got %v", meta["limit"])
	}
}

func TestListReports_LimitCappedAt100(t *testing.T) {
	mock := &mockQuerier{
		listReportsFunc: func(_ context.Context, arg db.ListReportsParams) ([]db.ListReportsRow, error) {
			// limit > 100 should be reset to default 20
			if arg.Limit != 20 {
				t.Errorf("expected limit capped to 20 (default), got %d", arg.Limit)
			}
			return []db.ListReportsRow{}, nil
		},
		countReportsFunc: func(_ context.Context) (int64, error) {
			return 0, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/reports?limit=200", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListReports(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListReports_ReportWithNilReason(t *testing.T) {
	mock := &mockQuerier{
		listReportsFunc: func(_ context.Context, _ db.ListReportsParams) ([]db.ListReportsRow, error) {
			return []db.ListReportsRow{
				{
					ID:               "r1",
					QuestionID:       "q1",
					ReporterID:       "u1",
					Reason:           sql.NullString{Valid: false},
					CreatedAt:        time.Now(),
					QuestionText:     "Some question",
					QuestionCategory: db.QuestionCategoryHOT,
					QuestionStatus:   db.QuestionStatusACTIVE,
					ReporterUsername:  "bob",
				},
			}, nil
		},
		countReportsFunc: func(_ context.Context) (int64, error) {
			return 1, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/reports", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListReports(c)
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)

	data := resp["data"].([]any)
	report := data[0].(map[string]any)
	if report["reason"] != nil {
		t.Errorf("expected reason to be nil, got %v", report["reason"])
	}
}

func TestListReports_ReportWithReason(t *testing.T) {
	mock := &mockQuerier{
		listReportsFunc: func(_ context.Context, _ db.ListReportsParams) ([]db.ListReportsRow, error) {
			return []db.ListReportsRow{
				{
					ID:               "r1",
					QuestionID:       "q1",
					ReporterID:       "u1",
					Reason:           sql.NullString{String: "spam content", Valid: true},
					CreatedAt:        time.Now(),
					QuestionText:     "Some question",
					QuestionCategory: db.QuestionCategorySECRETS,
					QuestionStatus:   db.QuestionStatusPENDING,
					ReporterUsername:  "charlie",
				},
			}, nil
		},
		countReportsFunc: func(_ context.Context) (int64, error) {
			return 1, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/reports", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListReports(c)
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)

	data := resp["data"].([]any)
	report := data[0].(map[string]any)
	if report["reason"] != "spam content" {
		t.Errorf("expected reason 'spam content', got %v", report["reason"])
	}
}

// --- ResolveReport Tests ---

func TestResolveReport_ApproveAction(t *testing.T) {
	var capturedStatus db.QuestionStatus
	mock := &mockQuerier{
		getReportByIDFunc: func(_ context.Context, id string) (db.GetReportByIDRow, error) {
			return db.GetReportByIDRow{
				ID:         id,
				QuestionID: "q1",
				ReporterID: "u1",
			}, nil
		},
		updateQuestionStatusFunc: func(_ context.Context, arg db.UpdateQuestionStatusParams) error {
			capturedStatus = arg.Status
			return nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	body := `{"action":"approve"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/reports/r1/resolve", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("r1")

	err := h.ResolveReport(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedStatus != db.QuestionStatusACTIVE {
		t.Errorf("expected ACTIVE status, got %s", capturedStatus)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["new_status"] != "ACTIVE" {
		t.Errorf("expected new_status ACTIVE, got %v", data["new_status"])
	}
	if data["action"] != "approve" {
		t.Errorf("expected action approve, got %v", data["action"])
	}
}

func TestResolveReport_RejectAction(t *testing.T) {
	var capturedStatus db.QuestionStatus
	mock := &mockQuerier{
		getReportByIDFunc: func(_ context.Context, id string) (db.GetReportByIDRow, error) {
			return db.GetReportByIDRow{
				ID:         id,
				QuestionID: "q1",
				ReporterID: "u1",
			}, nil
		},
		updateQuestionStatusFunc: func(_ context.Context, arg db.UpdateQuestionStatusParams) error {
			capturedStatus = arg.Status
			return nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	body := `{"action":"reject"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/reports/r1/resolve", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("r1")

	err := h.ResolveReport(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedStatus != db.QuestionStatusREJECTED {
		t.Errorf("expected REJECTED status, got %s", capturedStatus)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	if data["new_status"] != "REJECTED" {
		t.Errorf("expected new_status REJECTED, got %v", data["new_status"])
	}
}

func TestResolveReport_NotFound(t *testing.T) {
	mock := &mockQuerier{
		getReportByIDFunc: func(_ context.Context, _ string) (db.GetReportByIDRow, error) {
			return db.GetReportByIDRow{}, errors.New("sql: no rows in result set")
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	body := `{"action":"approve"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/reports/nonexistent/resolve", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.ResolveReport(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestResolveReport_BadRequestBody(t *testing.T) {
	h := NewHandler(&mockQuerier{}, "admin", "secret")
	e := setupEcho()

	body := `{"action":"invalid_action"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/reports/r1/resolve", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("r1")

	err := h.ResolveReport(c)
	// The validator returns an echo.HTTPError for validation failures
	if err == nil {
		// If no error, the response itself should indicate bad request
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	} else {
		// echo.HTTPError from validator
		he, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatalf("expected echo.HTTPError, got %T: %v", err, err)
		}
		if he.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", he.Code)
		}
	}
}

// --- GetStats Tests ---

func TestGetStats_Success(t *testing.T) {
	mock := &mockQuerier{
		countActiveUsers7DFunc: func(_ context.Context) (int64, error) {
			return 42, nil
		},
		countActiveUsers30DFunc: func(_ context.Context) (int64, error) {
			return 128, nil
		},
		countGroupsFunc: func(_ context.Context) (int64, error) {
			return 15, nil
		},
		sumRevenue7DaysFunc: func(_ context.Context) (int64, error) {
			return 9900, nil
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetStats(c)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)

	if data["dau_7d"].(float64) != 42 {
		t.Errorf("expected dau_7d 42, got %v", data["dau_7d"])
	}
	if data["mau_30d"].(float64) != 128 {
		t.Errorf("expected mau_30d 128, got %v", data["mau_30d"])
	}
	if data["groups_count"].(float64) != 15 {
		t.Errorf("expected groups_count 15, got %v", data["groups_count"])
	}
	if data["revenue_7d_rub"].(float64) != 9900 {
		t.Errorf("expected revenue_7d_rub 9900, got %v", data["revenue_7d_rub"])
	}
}

func TestGetStats_DBErrorsSilentlyIgnored(t *testing.T) {
	dbErr := errors.New("connection refused")
	mock := &mockQuerier{
		countActiveUsers7DFunc: func(_ context.Context) (int64, error) {
			return 0, dbErr
		},
		countActiveUsers30DFunc: func(_ context.Context) (int64, error) {
			return 0, dbErr
		},
		countGroupsFunc: func(_ context.Context) (int64, error) {
			return 0, dbErr
		},
		sumRevenue7DaysFunc: func(_ context.Context) (int64, error) {
			return 0, dbErr
		},
	}

	h := NewHandler(mock, "admin", "secret")
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/admin/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetStats(c)
	if err != nil {
		t.Fatal(err)
	}
	// Should still return 200 even when all DB calls fail
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)

	// All values should be zero (the zero-value returned alongside errors)
	if data["dau_7d"].(float64) != 0 {
		t.Errorf("expected dau_7d 0, got %v", data["dau_7d"])
	}
	if data["mau_30d"].(float64) != 0 {
		t.Errorf("expected mau_30d 0, got %v", data["mau_30d"])
	}
	if data["groups_count"].(float64) != 0 {
		t.Errorf("expected groups_count 0, got %v", data["groups_count"])
	}
	if data["revenue_7d_rub"].(float64) != 0 {
		t.Errorf("expected revenue_7d_rub 0, got %v", data["revenue_7d_rub"])
	}
}
