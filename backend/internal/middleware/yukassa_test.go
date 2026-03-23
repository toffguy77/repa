package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestYukassaIPAllowlist_AllowedIP(t *testing.T) {
	e := echo.New()
	e.POST("/webhook", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, YukassaIPAllowlist())

	// 185.71.76.1 is within 185.71.76.0/27
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Real-Ip", "185.71.76.1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for allowed IP, got %d", rec.Code)
	}
}

func TestYukassaIPAllowlist_BlockedIP(t *testing.T) {
	e := echo.New()
	e.POST("/webhook", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, YukassaIPAllowlist())

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Real-Ip", "1.2.3.4")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for blocked IP, got %d", rec.Code)
	}
}

func TestYukassaIPAllowlist_AnotherAllowedRange(t *testing.T) {
	e := echo.New()
	e.POST("/webhook", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, YukassaIPAllowlist())

	// 77.75.153.10 is within 77.75.153.0/25
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Real-Ip", "77.75.153.10")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for allowed IP, got %d", rec.Code)
	}
}
