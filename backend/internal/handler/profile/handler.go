package profile

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	profilesvc "github.com/repa-app/repa/internal/service/profile"
)

type Handler struct {
	svc *profilesvc.Service
}

func NewHandler(svc *profilesvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetProfile(c echo.Context) error {
	groupID := c.Param("id")
	userID := c.Param("userId")
	claims := appmw.GetCurrentUser(c)

	data, err := h.svc.GetProfile(c.Request().Context(), groupID, userID, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, profilesvc.ErrUserNotFound):
			return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		case errors.Is(err, profilesvc.ErrNotMember):
			return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "User is not a member of this group")
		default:
			return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": data,
	})
}
