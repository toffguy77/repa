package safety

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	safetysvc "github.com/repa-app/repa/internal/service/safety"
)

type Handler struct {
	svc *safetysvc.Service
}

func NewHandler(svc *safetysvc.Service) *Handler {
	return &Handler{svc: svc}
}

// BlockMember hides two members from each other.
func (h *Handler) BlockMember(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	targetID := c.Param("userId")

	if err := h.svc.Block(c.Request().Context(), claims.UserID, targetID); err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"blocked": true}})
}

// UnblockMember reverses a block the caller created.
func (h *Handler) UnblockMember(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	targetID := c.Param("userId")

	if err := h.svc.Unblock(c.Request().Context(), claims.UserID, targetID); err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"blocked": false}})
}

type reportMemberRequest struct {
	Reason  string `json:"reason"`
	GroupID string `json:"group_id"`
}

// ReportMember files a report with the administrative queue. Deliberately silent to the reported
// member: telling them would make reporting an act of confrontation.
func (h *Handler) ReportMember(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	targetID := c.Param("userId")

	var req reportMemberRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}

	if err := h.svc.ReportMember(c.Request().Context(), claims.UserID, targetID, req.GroupID, req.Reason); err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"reported": true}})
}

// RemoveMember takes a member out of a group and bans them from re-joining it.
func (h *Handler) RemoveMember(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	groupID := c.Param("id")
	memberID := c.Param("userId")

	if err := h.svc.RemoveMember(c.Request().Context(), claims.UserID, groupID, memberID); err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"removed": true}})
}

// LeavePermanently records that the departure is irreversible, so an invite cannot bring the member
// back. A flag on the leave route rather than a second endpoint, so the admin-transfer logic that
// leaving already performs stays in one place.
func (h *Handler) LeavePermanently(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)
	groupID := c.Param("id")

	if err := h.svc.BanSelf(c.Request().Context(), claims.UserID, groupID); err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]bool{"banned": true}})
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, safetysvc.ErrSelfTarget):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SELF_TARGET",
			"Нельзя выбрать себя")
	case errors.Is(err, safetysvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER",
			"Вы не участник этой группы")
	case errors.Is(err, safetysvc.ErrNotAdmin):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_ADMIN",
			"Только админ может это сделать")
	case errors.Is(err, safetysvc.ErrGroupNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Группа не найдена")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
