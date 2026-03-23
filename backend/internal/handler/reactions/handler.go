package reactions

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	reactionssvc "github.com/repa-app/repa/internal/service/reactions"
)

type Handler struct {
	svc *reactionssvc.Service
}

func NewHandler(svc *reactionssvc.Service) *Handler {
	return &Handler{svc: svc}
}

type createReactionRequest struct {
	Emoji string `json:"emoji" validate:"required"`
}

func (h *Handler) CreateReaction(c echo.Context) error {
	seasonID := c.Param("seasonId")
	targetID := c.Param("targetId")
	claims := appmw.GetCurrentUser(c)

	var req createReactionRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.svc.CreateReaction(c.Request().Context(), seasonID, claims.UserID, targetID, req.Emoji)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) GetReactions(c echo.Context) error {
	seasonID := c.Param("seasonId")
	targetID := c.Param("targetId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.GetReactions(c.Request().Context(), seasonID, targetID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": result})
}

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, reactionssvc.ErrSeasonNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Season not found")
	case errors.Is(err, reactionssvc.ErrSeasonNotRevealed):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SEASON_NOT_REVEALED", "Season results are not yet available")
	case errors.Is(err, reactionssvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	case errors.Is(err, reactionssvc.ErrInvalidEmoji):
		return handler.ErrorResponse(c, http.StatusBadRequest, "INVALID_EMOJI", "Emoji must be one of: 😂 🔥 💀 👀 🫡")
	case errors.Is(err, reactionssvc.ErrSelfReaction):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SELF_REACTION", "Cannot react to your own card")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
