package groups

import (
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
	appmw "github.com/repa-app/repa/internal/middleware"
	groupsvc "github.com/repa-app/repa/internal/service/groups"
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

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func TestToGroupDto(t *testing.T) {
	now := time.Now()
	g := db.Group{
		ID:                   "g1",
		Name:                 "Test Group",
		AdminID:              "u1",
		InviteCode:           "inv-123",
		Categories:           []string{"HOT", "FUNNY"},
		TelegramChatUsername: sql.NullString{String: "testchat", Valid: true},
		CreatedAt:            now,
	}

	dto := toGroupDto(g)

	if dto.ID != "g1" {
		t.Errorf("expected ID g1, got %s", dto.ID)
	}
	if dto.Name != "Test Group" {
		t.Errorf("expected name Test Group, got %s", dto.Name)
	}
	if dto.AdminID != "u1" {
		t.Errorf("expected admin_id u1, got %s", dto.AdminID)
	}
	if len(dto.Categories) != 2 || dto.Categories[0] != "HOT" {
		t.Errorf("expected categories [HOT FUNNY], got %v", dto.Categories)
	}
	if dto.TelegramUsername == nil || *dto.TelegramUsername != "testchat" {
		t.Error("expected telegram username testchat")
	}
	if dto.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("expected created_at %s, got %s", now.Format(time.RFC3339), dto.CreatedAt)
	}
}

func TestToGroupDto_NullTelegram(t *testing.T) {
	g := db.Group{
		ID:         "g1",
		Name:       "No TG",
		Categories: []string{"HOT"},
		CreatedAt:  time.Now(),
	}

	dto := toGroupDto(g)
	if dto.TelegramUsername != nil {
		t.Error("expected nil telegram username")
	}
}

func TestMapServiceError(t *testing.T) {
	e := setupEcho()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", groupsvc.ErrGroupNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"not member", groupsvc.ErrNotMember, http.StatusForbidden, "NOT_MEMBER"},
		{"already member", groupsvc.ErrAlreadyMember, http.StatusConflict, "ALREADY_MEMBER"},
		{"group limit user", groupsvc.ErrGroupLimitUser, http.StatusConflict, "GROUP_LIMIT"},
		{"group limit size", groupsvc.ErrGroupLimitSize, http.StatusConflict, "MEMBER_LIMIT"},
		{"not admin", groupsvc.ErrNotAdmin, http.StatusForbidden, "NOT_ADMIN"},
		{"invalid name", groupsvc.ErrInvalidName, http.StatusBadRequest, "VALIDATION"},
		{"no categories", groupsvc.ErrNoCategories, http.StatusBadRequest, "VALIDATION"},
		{"invalid category", groupsvc.ErrInvalidCategory, http.StatusBadRequest, "VALIDATION"},
		{"romance blocked", groupsvc.ErrRomanceBlocked, http.StatusForbidden, "ROMANCE_BLOCKED"},
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

func TestCreateGroup_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGroup(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateGroup_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups/g1",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UpdateGroup(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMapServiceError_Unknown(t *testing.T) {
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mapServiceError(c, errors.New("unknown error"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
