package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// Matches both well-formed tags like <script> and malformed ones like <script src=x (no closing >)
var htmlTagRe = regexp.MustCompile(`<[^>]*>|<[a-zA-Z][^>]*$`)

func Sanitize() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ct := c.Request().Header.Get("Content-Type")
			if c.Request().Body == nil || !strings.Contains(ct, "application/json") {
				return next(c)
			}

			body, err := io.ReadAll(c.Request().Body)
			c.Request().Body.Close()
			if err != nil {
				return next(c)
			}

			if bytes.ContainsRune(body, 0) {
				return c.JSON(http.StatusBadRequest, map[string]any{
					"error": map[string]string{
						"code":    "VALIDATION",
						"message": "Request contains invalid characters",
					},
				})
			}

			var data any
			if err := json.Unmarshal(body, &data); err != nil {
				c.Request().Body = io.NopCloser(bytes.NewReader(body))
				return next(c)
			}

			sanitized := sanitizeValue(data)
			newBody, err := json.Marshal(sanitized)
			if err != nil {
				c.Request().Body = io.NopCloser(bytes.NewReader(body))
				return next(c)
			}

			c.Request().Body = io.NopCloser(bytes.NewReader(newBody))
			c.Request().ContentLength = int64(len(newBody))
			return next(c)
		}
	}
}

func sanitizeValue(v any) any {
	switch val := v.(type) {
	case string:
		s := strings.TrimSpace(val)
		s = htmlTagRe.ReplaceAllString(s, "")
		return s
	case map[string]any:
		for k, v2 := range val {
			val[k] = sanitizeValue(v2)
		}
		return val
	case []any:
		for i, v2 := range val {
			val[i] = sanitizeValue(v2)
		}
		return val
	default:
		return v
	}
}
