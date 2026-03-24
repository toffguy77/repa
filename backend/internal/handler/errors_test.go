package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestErrorResponse(t *testing.T) {
	t.Run("returns JSON with error.code and error.message", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "invalid input")

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		errObj, ok := body["error"]
		if !ok {
			t.Fatal("expected 'error' key in response")
		}
		if errObj["code"] != "VALIDATION" {
			t.Errorf("expected code VALIDATION, got %s", errObj["code"])
		}
		if errObj["message"] != "invalid input" {
			t.Errorf("expected message 'invalid input', got %s", errObj["message"])
		}
	})

	t.Run("uses correct HTTP status code", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "not found")

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("Echo HTTPError 400 maps to VALIDATION", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		he := echo.NewHTTPError(http.StatusBadRequest, "bad request")
		ErrorHandler(he, c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if body["error"]["code"] != "VALIDATION" {
			t.Errorf("expected code VALIDATION, got %s", body["error"]["code"])
		}
		if body["error"]["message"] != "bad request" {
			t.Errorf("expected message 'bad request', got %s", body["error"]["message"])
		}
	})

	t.Run("Echo HTTPError 404 maps to NOT_FOUND", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		he := echo.NewHTTPError(http.StatusNotFound, "page not found")
		ErrorHandler(he, c)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if body["error"]["code"] != "NOT_FOUND" {
			t.Errorf("expected code NOT_FOUND, got %s", body["error"]["code"])
		}
	})

	t.Run("Echo HTTPError 405 maps to METHOD_NOT_ALLOWED", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		he := echo.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		ErrorHandler(he, c)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if body["error"]["code"] != "METHOD_NOT_ALLOWED" {
			t.Errorf("expected code METHOD_NOT_ALLOWED, got %s", body["error"]["code"])
		}
	})

	t.Run("Echo HTTPError with map message passes through", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		msgMap := map[string]any{
			"error": map[string]string{
				"code":    "VALIDATION",
				"message": "field is required",
			},
		}
		he := echo.NewHTTPError(http.StatusBadRequest, msgMap)
		ErrorHandler(he, c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if body["error"]["code"] != "VALIDATION" {
			t.Errorf("expected code VALIDATION, got %s", body["error"]["code"])
		}
		if body["error"]["message"] != "field is required" {
			t.Errorf("expected message 'field is required', got %s", body["error"]["message"])
		}
	})

	t.Run("non-Echo error returns 500 INTERNAL", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ErrorHandler(errors.New("something broke"), c)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		var body map[string]map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if body["error"]["code"] != "INTERNAL" {
			t.Errorf("expected code INTERNAL, got %s", body["error"]["code"])
		}
		if body["error"]["message"] != "Internal server error" {
			t.Errorf("expected message 'Internal server error', got %s", body["error"]["message"])
		}
	})

	t.Run("already committed response is a no-op", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Write a response to mark it as committed
		_ = c.String(http.StatusOK, "already sent")
		committed := rec.Body.String()

		ErrorHandler(errors.New("should be ignored"), c)

		// Body should remain unchanged — no additional JSON written
		if rec.Body.String() != committed {
			t.Error("expected response body to remain unchanged after committed response")
		}
	})
}
