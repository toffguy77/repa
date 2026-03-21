package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrorResponse(c echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": AppError{Code: code, Message: message},
	})
}

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		status := he.Code
		// Echo validation errors come as HTTPError with a map body
		if msg, ok := he.Message.(map[string]any); ok {
			_ = c.JSON(status, msg)
			return
		}
		code := "ERROR"
		switch status {
		case http.StatusBadRequest:
			code = "VALIDATION"
		case http.StatusNotFound:
			code = "NOT_FOUND"
		case http.StatusMethodNotAllowed:
			code = "METHOD_NOT_ALLOWED"
		}
		message := http.StatusText(status)
		if m, ok := he.Message.(string); ok {
			message = m
		}
		_ = ErrorResponse(c, status, code, message)
		return
	}

	_ = ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
}
