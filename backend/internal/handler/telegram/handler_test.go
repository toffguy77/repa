package telegram

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	telegramsvc "github.com/repa-app/repa/internal/service/telegram"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	return e
}

func TestWebhook_InvalidSecret(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWebhook_ValidSecret_EmptyUpdate(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: "my-secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "my-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestWebhook_NoSecret(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`{"update_id": 1}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when no secret configured, got %d", rec.Code)
	}
}

func TestWebhook_BadJSON(t *testing.T) {
	e := setupEcho()
	h := &Handler{webhookSecret: ""}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telegram/webhook",
		strings.NewReader(`not json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Webhook(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
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
		{"not found", telegramsvc.ErrGroupNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"not admin", telegramsvc.ErrNotAdmin, http.StatusForbidden, "NOT_ADMIN"},
		{"not member", telegramsvc.ErrNotMember, http.StatusForbidden, "NOT_MEMBER"},
		{"no telegram", telegramsvc.ErrNoTelegram, http.StatusBadRequest, "NO_TELEGRAM"},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, "INTERNAL"},
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

func TestHasPrefix(t *testing.T) {
	if !hasPrefix("/connect ABC", "/connect ") {
		t.Error("expected true for /connect prefix")
	}
	if hasPrefix("/repa", "/connect ") {
		t.Error("expected false for /repa with /connect prefix")
	}
}

func TestExtractArg(t *testing.T) {
	got := extractArg("/connect REPA-X7K2", "/connect ")
	if got != "REPA-X7K2" {
		t.Errorf("expected REPA-X7K2, got %s", got)
	}
}
