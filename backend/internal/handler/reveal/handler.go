package reveal

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
)

type Handler struct {
	svc *revealsvc.Service
}

func NewHandler(svc *revealsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetReveal(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	data, err := h.svc.GetReveal(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": data,
	})
}

func (h *Handler) GetMembersCards(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	cards, err := h.svc.GetMembersCards(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"members": cards,
		},
	})
}

func (h *Handler) OpenHidden(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.OpenHidden(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": result,
	})
}

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, revealsvc.ErrSeasonNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Season not found")
	case errors.Is(err, revealsvc.ErrSeasonNotRevealed):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SEASON_NOT_REVEALED", "Season results are not yet available")
	case errors.Is(err, revealsvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	case errors.Is(err, revealsvc.ErrInsufficientFunds):
		return handler.ErrorResponse(c, http.StatusPaymentRequired, "INSUFFICIENT_FUNDS", "Not enough crystals")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
