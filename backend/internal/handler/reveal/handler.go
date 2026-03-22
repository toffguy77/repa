package reveal

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	cardssvc "github.com/repa-app/repa/internal/service/cards"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
)

type Handler struct {
	svc      *revealsvc.Service
	cardsSvc *cardssvc.Service
}

func NewHandler(svc *revealsvc.Service, cardsSvc *cardssvc.Service) *Handler {
	return &Handler{svc: svc, cardsSvc: cardsSvc}
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

func (h *Handler) GetMyCardURL(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	// Validate season exists, is revealed, and user is a member
	if _, err := h.svc.ValidateRevealAccess(c.Request().Context(), seasonID, claims.UserID); err != nil {
		return mapServiceError(c, err)
	}

	url, err := h.cardsSvc.GetCardURL(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusOK, map[string]any{
				"data": map[string]any{
					"image_url": nil,
					"status":    "generating",
				},
			})
		}
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"image_url": url,
			"status":    "ready",
		},
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
