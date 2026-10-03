package chronicle

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	chroniclesvc "github.com/repa-app/repa/internal/service/chronicle"
)

func newHandlerWithMock(t *testing.T) (*Handler, sqlmock.Sqlmock, *echo.Echo) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return NewHandler(chroniclesvc.NewService(db.New(sqlDB))), mock, echo.New()
}

func setUser(c echo.Context, userID string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: "tester"})
}

func doGet(t *testing.T, h *Handler, e *echo.Echo, userID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/g1/chronicle", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("g1")
	setUser(c, userID)
	if err := h.Get(c); err != nil {
		t.Fatalf("Get: %v", err)
	}
	return rec
}

var chronicleCols = []string{
	"season_id", "season_number", "reveal_at",
	"question_id", "question_text", "question_category",
	"target_id", "username", "avatar_emoji",
	"percentage", "vote_count", "total_voters",
}

func TestGet_MemberSuccess(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT .+ FROM blocks").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
	mock.ExpectQuery("SELECT DISTINCT ON").
		WithArgs("g1", int32(chroniclesvc.MaxSeasons)).
		WillReturnRows(sqlmock.NewRows(chronicleCols).AddRow(
			"s1", int32(1), time.Now(), "q1", "Кто чаще всех смеётся?", "FUNNY",
			"u2", "bob", nil, 66.7, int32(2), int32(10),
		))
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	mock.ExpectQuery("SELECT .+ FROM group_members").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "avatar_emoji", "guess_accuracy", "seasons_played"}).
			AddRow("u1", "alice", nil, 0.8, int32(2)))

	rec := doGet(t, h, e, "u1")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Seasons []struct {
				Number  int32 `json:"number"`
				Entries []struct {
					Username    string `json:"username"`
					SmallSample bool   `json:"small_sample"`
				} `json:"entries"`
			} `json:"seasons"`
			Standing struct {
				Available bool `json:"available"`
				Members   []struct {
					Rank int `json:"rank"`
				} `json:"members"`
			} `json:"standing"`
			SeasonsTotal int64 `json:"seasons_total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data.Seasons) != 1 || len(resp.Data.Seasons[0].Entries) != 1 {
		t.Fatalf("expected one season with one entry, got %+v", resp.Data.Seasons)
	}
	if resp.Data.Seasons[0].Entries[0].Username != "bob" {
		t.Errorf("entry names %q", resp.Data.Seasons[0].Entries[0].Username)
	}
	if resp.Data.Seasons[0].Entries[0].SmallSample {
		t.Error("ten voters is not a small sample")
	}
	if !resp.Data.Standing.Available || len(resp.Data.Standing.Members) != 1 {
		t.Errorf("expected the standing to be shown, got %+v", resp.Data.Standing)
	}
	if resp.Data.SeasonsTotal != 2 {
		t.Errorf("seasons_total = %d, want 2", resp.Data.SeasonsTotal)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGet_NonMemberIsRefused(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("outsider", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	// Nothing else expected: the chronicle is never read for a non-member.

	rec := doGet(t, h, e, "outsider")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGet_EmptyChronicleIsNotAnError(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT .+ FROM blocks").
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
	mock.ExpectQuery("SELECT DISTINCT ON").
		WithArgs("g1", int32(chroniclesvc.MaxSeasons)).
		WillReturnRows(sqlmock.NewRows(chronicleCols))
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	// No standing query: below two revealed seasons it is withheld rather than fetched and dropped.

	rec := doGet(t, h, e, "u1")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	// Arrays, not nulls: an empty history is a new group, not a broken response.
	body := rec.Body.String()
	for _, want := range []string{`"seasons":[]`, `"members":[]`, `"available":false`} {
		if !contains(body, want) {
			t.Errorf("expected %s in %s", want, body)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGet_DatabaseErrorIsInternal(t *testing.T) {
	h, mock, e := newHandlerWithMock(t)
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("u1", "g1").
		WillReturnError(http.ErrBodyNotAllowed)

	rec := doGet(t, h, e, "u1")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
