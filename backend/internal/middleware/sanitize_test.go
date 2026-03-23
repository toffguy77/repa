package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSanitize_StripHTMLTags(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		var body map[string]string
		if err := c.Bind(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, body)
	})

	body := `{"name":"<script>alert('xss')</script>Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<script>") {
		t.Error("HTML tags were not stripped")
	}
	if !strings.Contains(rec.Body.String(), "alert('xss')") {
		t.Error("text content should be preserved")
	}
}

func TestSanitize_TrimWhitespace(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		var body map[string]string
		if err := c.Bind(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, body)
	})

	body := `{"name":"  hello  "}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// Should contain trimmed "hello" without leading/trailing spaces
	resp := rec.Body.String()
	if strings.Contains(resp, `"  hello  "`) {
		t.Error("whitespace was not trimmed")
	}
}

func TestSanitize_RejectNullBytes(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	body := "{\"name\":\"hello\x00world\"}"
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for null bytes, got %d", rec.Code)
	}
}

func TestSanitize_NonJSON_PassThrough(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	body := `not json at all`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, "text/plain")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for non-JSON, got %d", rec.Code)
	}
}

func TestSanitize_MalformedHTMLTag(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		var body map[string]string
		if err := c.Bind(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, body)
	})

	// Malformed tag without closing >
	body := `{"name":"<script src=x"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	resp := rec.Body.String()
	if strings.Contains(resp, "<script") {
		t.Error("malformed HTML tag was not stripped")
	}
}

func TestSanitize_NestedObjects(t *testing.T) {
	e := echo.New()
	e.Use(Sanitize())
	e.POST("/test", func(c echo.Context) error {
		var body map[string]any
		if err := c.Bind(&body); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, body)
	})

	body := `{"outer":{"inner":"<b>bold</b>"},"list":["<i>item</i>"]}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	resp := rec.Body.String()
	if strings.Contains(resp, "<b>") || strings.Contains(resp, "<i>") {
		t.Error("nested HTML tags were not stripped")
	}
}
