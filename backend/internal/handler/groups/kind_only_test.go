package groups

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// kindGroupRow is a group row with an explicit categories list and setting.
func kindGroupRow(adminID, categories string, kindOnly bool) *sqlmock.Rows {
	return sqlmock.NewRows(handlerGroupCols).AddRow(
		"g1", "9Б", "ABC123", adminID,
		nil, nil, nil, nil,
		time.Now(), categories, kindOnly,
	)
}

func patchGroup(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := setupEcho()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, "u1", "testuser")

	if err := h.UpdateGroup(c); err != nil {
		// A validation error is returned rather than written; render it so the test can read a status.
		e.HTTPErrorHandler(err, c)
	}
	return rec
}

// --- The field is in the response ---

func TestUpdateGroup_KindOnlyIsInTheResponse(t *testing.T) {
	h, mock, _ := newHandlerWithMock(t)

	// SetKindOnly: admin lookup, bank check, update.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"FUNNY"}`, false))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(89)))
	mock.ExpectExec("UPDATE groups SET kind_only").
		WithArgs("g1", true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// UpdateGroup: admin check, then the re-read that produces the response.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"FUNNY"}`, true))
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"FUNNY"}`, true))

	rec := patchGroup(t, h, `{"kind_only":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Group struct {
				KindOnly bool `json:"kind_only"`
			} `json:"group"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The response must show the value that was just stored, not the one read before the update —
	// otherwise the client's switch snaps back to its old position.
	if !resp.Data.Group.KindOnly {
		t.Error("expected kind_only true in the response")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- Only the admin may change it ---

func TestUpdateGroup_KindOnlyNonAdminIsRefused(t *testing.T) {
	h, mock, _ := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("someone-else", `{"FUNNY"}`, false))
	// No bank check and no update: the refusal happens before either.

	rec := patchGroup(t, h, `{"kind_only":true}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "NOT_ADMIN") {
		t.Errorf("expected NOT_ADMIN, got %s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- The refusal surfaces as something the client can act on ---

func TestUpdateGroup_KindOnlyWithNoKindQuestions(t *testing.T) {
	h, mock, _ := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"HOT","SECRETS"}`, false))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	rec := patchGroup(t, h, `{"kind_only":true}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	// A distinct code, so the client can point at the switch or the categories rather than showing a
	// generic validation failure the member cannot resolve.
	if !strings.Contains(rec.Body.String(), "NO_KIND_CATEGORIES") {
		t.Errorf("expected NO_KIND_CATEGORIES, got %s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- An update that does not mention the field leaves it alone ---

func TestUpdateGroup_WithoutKindOnlyDoesNotTouchIt(t *testing.T) {
	h, mock, _ := newHandlerWithMock(t)

	// Only UpdateGroup's own statements. Expectations are matched in order against the SQL, so a
	// stray kind_only update here fails the test.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"FUNNY"}`, true))
	mock.ExpectExec("UPDATE groups SET name").
		WithArgs("g1", "10А").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("g1").
		WillReturnRows(kindGroupRow("u1", `{"FUNNY"}`, true))

	rec := patchGroup(t, h, `{"name":"10А"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- Creation carries the field through ---

func TestCreateGroup_KindOnlyRefusalSurfaces(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	userCols := []string{"id", "phone", "apple_id", "google_id", "username", "avatar_url", "avatar_emoji", "birth_year", "created_at", "updated_at", "username_changed_at"}
	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"u1", nil, nil, nil, "testuser", nil, nil,
			sql.NullInt32{Int32: int32(time.Now().Year() - 15), Valid: true},
			time.Now(), time.Now(), nil,
		))
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// kind_only is omitted on purpose: a 15-year-old creator gets the setting from their age, which
	// is the path that produces this refusal without anyone having touched a switch.
	body := `{"name":"9Б","categories":["HOT","SECRETS"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "u1", "testuser")

	if err := h.CreateGroup(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "NO_KIND_CATEGORIES") {
		t.Errorf("expected NO_KIND_CATEGORIES, got %s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
