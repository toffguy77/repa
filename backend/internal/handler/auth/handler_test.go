package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/repa-app/repa/internal/config"
	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
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

func TestToUserDto(t *testing.T) {
	now := time.Now()
	user := db.User{
		ID:          "user-123",
		Username:    "testuser",
		AvatarUrl:   sql.NullString{String: "https://example.com/avatar.jpg", Valid: true},
		AvatarEmoji: sql.NullString{String: "🍆", Valid: true},
		BirthYear:   sql.NullInt32{Int32: 2000, Valid: true},
		CreatedAt:   now,
	}

	dto := toUserDto(user)

	if dto.ID != "user-123" {
		t.Errorf("expected ID user-123, got %s", dto.ID)
	}
	if dto.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", dto.Username)
	}
	if dto.AvatarURL == nil || *dto.AvatarURL != "https://example.com/avatar.jpg" {
		t.Error("expected avatar URL")
	}
	if dto.AvatarEmoji == nil || *dto.AvatarEmoji != "🍆" {
		t.Error("expected avatar emoji")
	}
	if dto.BirthYear == nil || *dto.BirthYear != 2000 {
		t.Error("expected birth year 2000")
	}
	if dto.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("expected created_at %s, got %s", now.Format(time.RFC3339), dto.CreatedAt)
	}
}

func TestToUserDto_NullFields(t *testing.T) {
	user := db.User{
		ID:        "user-456",
		Username:  "nulluser",
		CreatedAt: time.Now(),
	}

	dto := toUserDto(user)

	if dto.AvatarURL != nil {
		t.Error("expected nil avatar URL")
	}
	if dto.AvatarEmoji != nil {
		t.Error("expected nil avatar emoji")
	}
	if dto.BirthYear != nil {
		t.Error("expected nil birth year")
	}
}

func TestAppVersion(t *testing.T) {
	e := setupEcho()
	cfg := &config.Config{
		AppMinVersion:    "1.0.0",
		AppLatestVersion: "1.2.0",
	}
	h := &Handler{cfg: cfg}

	tests := []struct {
		name         string
		appVersion   string
		forceUpdate  bool
	}{
		{"no header", "", false},
		{"current version", "1.2.0", false},
		{"old version", "0.9.0", true},
		{"min version", "1.0.0", false},
		{"double digit major", "10.0.0", false},
		{"double digit vs single", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/app/version", nil)
			if tt.appVersion != "" {
				req.Header.Set("X-App-Version", tt.appVersion)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := h.AppVersion(c); err != nil {
				t.Fatal(err)
			}

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}

			var resp map[string]map[string]any
			json.Unmarshal(rec.Body.Bytes(), &resp)

			data := resp["data"]
			if data["min_version"] != "1.0.0" {
				t.Errorf("expected min_version 1.0.0, got %v", data["min_version"])
			}
			if data["latest_version"] != "1.2.0" {
				t.Errorf("expected latest_version 1.2.0, got %v", data["latest_version"])
			}
			if data["force_update"] != tt.forceUpdate {
				t.Errorf("expected force_update %v, got %v", tt.forceUpdate, data["force_update"])
			}
		})
	}
}

func TestUsernameCheck_MissingParam(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/username-check", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UsernameCheck(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestOTPSend_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.OTPSend(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUploadAvatar_NoFile(t *testing.T) {
	e := setupEcho()
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/avatar", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &appmw.JWTClaims{UserID: "user-123", Username: "testuser"})

	err := h.UploadAvatar(c)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

