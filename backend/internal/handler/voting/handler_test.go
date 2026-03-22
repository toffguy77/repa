package voting

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
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
