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

func (h *Handler) GetDetector(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.GetDetector(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) BuyDetector(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.BuyDetector(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
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

// GetAnticipation returns what a member may know about votes concerning them mid-week: a count and,
// from Thursday, one category emoji. Nothing that identifies anyone or names an attribute.
func (h *Handler) GetAnticipation(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	state, err := h.svc.GetAnticipation(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": state})
}

// BuyDetectorHint reveals one more voter partially — the middle rung of the detector ladder.
func (h *Handler) BuyDetectorHint(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.BuyDetectorHint(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

type recordShareRequest struct {
	Channel string `json:"channel"`
}

// RecordShare notes that the member shared their card. Never fails the caller's flow for a
// telemetry write — see revealsvc.RecordShare.
func (h *Handler) RecordShare(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	var req recordShareRequest
	// A missing or malformed body is not worth failing a share over; the channel simply
	// becomes "other".
	_ = c.Bind(&req)

	if err := h.svc.RecordShare(c.Request().Context(), seasonID, claims.UserID, req.Channel); err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"recorded": true}})
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
	case errors.Is(err, revealsvc.ErrSeasonAlreadyRevealed):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SEASON_ALREADY_REVEALED",
			"Результаты уже готовы")
	case errors.Is(err, revealsvc.ErrNothingToReveal):
		return handler.ErrorResponse(c, http.StatusConflict, "NOTHING_TO_REVEAL",
			"Больше подсказок нет")
	case errors.Is(err, revealsvc.ErrGroupTooSmall):
		return handler.ErrorResponse(c, http.StatusForbidden, "GROUP_TOO_SMALL",
			"Детектор доступен в группах от 5 человек")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
