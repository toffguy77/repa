package chronicle

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	chroniclesvc "github.com/repa-app/repa/internal/service/chronicle"
)

type Handler struct {
	svc *chroniclesvc.Service
}

func NewHandler(svc *chroniclesvc.Service) *Handler {
	return &Handler{svc: svc}
}

// Get returns the group's chronicle.
// GET /api/v1/groups/:id/chronicle
func (h *Handler) Get(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)

	chron, err := h.svc.Get(c.Request().Context(), claims.UserID, c.Param("id"))
	if err != nil {
		if errors.Is(err, chroniclesvc.ErrNotMember) {
			return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}

	return c.JSON(http.StatusOK, map[string]any{"data": chron})
}
