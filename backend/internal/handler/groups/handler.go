package groups

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/eligibility"
	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	groupsvc "github.com/repa-app/repa/internal/service/groups"
	"github.com/rs/zerolog/log"
)

// SelfBanner records a permanent departure. Implemented by the safety service; a narrow interface
// keeps the groups handler from depending on the whole safety surface.
type SelfBanner interface {
	BanSelf(ctx context.Context, userID, groupID string) error
}

type Handler struct {
	svc    *groupsvc.Service
	safety SelfBanner
}

func NewHandler(svc *groupsvc.Service) *Handler {
	return &Handler{svc: svc}
}

// WithSafety wires permanent departure. Optional: nil means the flag is ignored, which is what the
// handler's own unit tests want.
func (h *Handler) WithSafety(s SelfBanner) *Handler {
	h.safety = s
	return h
}

// --- DTOs ---

type groupDto struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	AdminID          string   `json:"admin_id"`
	InviteCode       string   `json:"invite_code"`
	Categories       []string `json:"categories"`
	KindOnly         bool     `json:"kind_only"`
	TelegramUsername *string  `json:"telegram_username"`
	CreatedAt        string   `json:"created_at"`
}

func toGroupDto(g db.Group) groupDto {
	dto := groupDto{
		ID:         g.ID,
		Name:       g.Name,
		AdminID:    g.AdminID,
		InviteCode: g.InviteCode,
		Categories: g.Categories,
		KindOnly:   g.KindOnly,
		CreatedAt:  g.CreatedAt.Format(time.RFC3339),
	}
	if g.TelegramChatUsername.Valid {
		dto.TelegramUsername = &g.TelegramChatUsername.String
	}
	return dto
}

type memberDto struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarURL   *string `json:"avatar_url"`
	IsAdmin     bool    `json:"is_admin"`
}

type seasonDto struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	RevealAt   string `json:"reveal_at"`
	VotedCount int64  `json:"voted_count"`
	TotalCount int64  `json:"total_count"`
	UserVoted  bool   `json:"user_voted"`
	// RevealState tells the app whether the Reveal is SCHEDULED, WAITING_FOR_MEMBERS,
	// WAITING_FOR_VOTERS, POSTPONED or REVEALED, and MembersNeeded/VotersNeeded say how
	// far off it is. See internal/eligibility.
	RevealState   string `json:"reveal_state"`
	MembersNeeded int64  `json:"members_needed"`
	VotersNeeded  int64  `json:"voters_needed"`
}

type groupListItemDto struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	MemberCount      int64      `json:"member_count"`
	InviteCode       string     `json:"invite_code"`
	TelegramUsername *string    `json:"telegram_username"`
	ActiveSeason     *seasonDto `json:"active_season"`
}

// --- Handlers ---

type createGroupRequest struct {
	Name             string   `json:"name" validate:"required,min=3,max=40"`
	Categories       []string `json:"categories" validate:"required,min=1"`
	TelegramUsername string   `json:"telegram_username"`
	// KindOnly restricts the group to warm and neutral questions. Omitted means "use the age-based
	// default"; a pointer rather than a bool so an explicit false is distinguishable from absence.
	KindOnly *bool `json:"kind_only"`
}

func (h *Handler) CreateGroup(c echo.Context) error {
	var req createGroupRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := appmw.GetCurrentUser(c)

	user, err := h.svc.GetUser(c.Request().Context(), claims.UserID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to get user")
	}

	result, err := h.svc.CreateGroup(c.Request().Context(), groupsvc.CreateGroupParams{
		UserID:           claims.UserID,
		Name:             req.Name,
		Categories:       req.Categories,
		TelegramUsername: req.TelegramUsername,
		KindOnly:         req.KindOnly,
		UserBirthYear:    user.BirthYear,
	})
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": map[string]any{
			"group":      toGroupDto(result.Group),
			"invite_url": result.InviteURL,
		},
	})
}

func (h *Handler) ListGroups(c echo.Context) error {
	claims := appmw.GetCurrentUser(c)

	items, err := h.svc.ListUserGroups(c.Request().Context(), claims.UserID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to list groups")
	}

	dtos := make([]groupListItemDto, len(items))
	for i, item := range items {
		dto := groupListItemDto{
			ID:          item.Group.ID,
			Name:        item.Group.Name,
			MemberCount: item.MemberCount,
			InviteCode:  item.Group.InviteCode,
		}
		if item.Group.TelegramChatUsername.Valid {
			dto.TelegramUsername = &item.Group.TelegramChatUsername.String
		}
		if item.ActiveSeason != nil {
			season := &seasonDto{
				ID:         item.ActiveSeason.ID,
				Status:     string(item.ActiveSeason.Status),
				RevealAt:   item.ActiveSeason.RevealAt.Format(time.RFC3339),
				VotedCount: item.VotedCount,
				TotalCount: item.MemberCount,
				UserVoted:  item.UserVoted,
			}
			if item.RevealStatus != nil {
				season.RevealState = string(eligibility.StateFor(
					item.ActiveSeason.Status, item.ActiveSeason.PostponeCount, *item.RevealStatus))
				season.MembersNeeded = item.RevealStatus.MembersNeeded
				season.VotersNeeded = item.RevealStatus.VotersNeeded
			}
			dto.ActiveSeason = season
		}
		dtos[i] = dto
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"groups": dtos}})
}

func (h *Handler) GetGroup(c echo.Context) error {
	groupID := c.Param("id")
	claims := appmw.GetCurrentUser(c)

	detail, err := h.svc.GetGroup(c.Request().Context(), groupID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	members := make([]memberDto, len(detail.Members))
	for i, m := range detail.Members {
		md := memberDto{
			ID:       m.ID,
			Username: m.Username,
			IsAdmin:  m.ID == detail.Group.AdminID,
		}
		if m.AvatarEmoji.Valid {
			md.AvatarEmoji = &m.AvatarEmoji.String
		}
		if m.AvatarUrl.Valid {
			md.AvatarURL = &m.AvatarUrl.String
		}
		members[i] = md
	}

	resp := map[string]any{
		"group":   toGroupDto(detail.Group),
		"members": members,
		// Anonymity thresholds are judged against this, not the raw member count — see
		// docs/features/groups.md.
		"effective_member_count": detail.EffectiveMemberCount,
		// The next group size that changes what the group can do, or null once every threshold is
		// crossed. Derived from eligibility so the number and the consequence have one owner, and
		// given both counts: the Reveal threshold is judged on actual membership (as the worker does)
		// while the detector threshold is judged on the effective count (as the detector does).
		// Null means "big enough" — the app must not invent a target.
		"next_threshold": eligibility.NextGrowthThreshold(
			int64(len(detail.Members)), int64(detail.EffectiveMemberCount)),
	}

	if detail.ActiveSeason != nil {
		revealState := string(eligibility.StateRevealed)
		var membersNeeded, votersNeeded, votedCount int64
		if detail.RevealStatus != nil {
			revealState = string(eligibility.StateFor(
				detail.ActiveSeason.Status, detail.ActiveSeason.PostponeCount, *detail.RevealStatus))
			membersNeeded = detail.RevealStatus.MembersNeeded
			votersNeeded = detail.RevealStatus.VotersNeeded
			votedCount = detail.RevealStatus.VotedCount
		}
		resp["active_season"] = map[string]any{
			"id":             detail.ActiveSeason.ID,
			"number":         detail.ActiveSeason.Number,
			"status":         string(detail.ActiveSeason.Status),
			"starts_at":      detail.ActiveSeason.StartsAt.Format(time.RFC3339),
			"reveal_at":      detail.ActiveSeason.RevealAt.Format(time.RFC3339),
			"ends_at":        detail.ActiveSeason.EndsAt.Format(time.RFC3339),
			"kind":           string(detail.ActiveSeason.Kind),
			"reveal_state":   revealState,
			"members_needed": membersNeeded,
			"voters_needed":  votersNeeded,
			"voted_count":    votedCount,
			"total_count":    int64(len(detail.Members)),
		}
	}

	return c.JSON(http.StatusOK, map[string]any{"data": resp})
}

func (h *Handler) JoinPreview(c echo.Context) error {
	inviteCode := c.Param("inviteCode")

	preview, err := h.svc.GetJoinPreview(c.Request().Context(), inviteCode)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"name":           preview.Name,
			"member_count":   preview.MemberCount,
			"admin_username": preview.AdminUsername,
		},
	})
}

func (h *Handler) JoinGroup(c echo.Context) error {
	inviteCode := c.Param("inviteCode")
	claims := appmw.GetCurrentUser(c)

	// `source` records how the member arrived. It is optional and never fails the join: an
	// unrecognised or missing value becomes UNKNOWN.
	source := groupsvc.ParseJoinSource(c.QueryParam("source"))

	// `ref` is the member who shared the link. Validated in the service; an invalid claim is
	// dropped rather than failing the join.
	referrer := c.QueryParam("ref")

	group, err := h.svc.JoinGroup(c.Request().Context(), claims.UserID, inviteCode, source, referrer)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"group": toGroupDto(*group)}})
}

func (h *Handler) LeaveGroup(c echo.Context) error {
	groupID := c.Param("id")
	claims := appmw.GetCurrentUser(c)

	// `permanent=true` means an invite must not bring the member back. Ordinary leaving stays
	// reversible on purpose: people leave groups by accident.
	permanent := c.QueryParam("permanent") == "true"

	if err := h.svc.LeaveGroup(c.Request().Context(), claims.UserID, groupID); err != nil {
		return mapServiceError(c, err)
	}

	if permanent && h.safety != nil {
		if err := h.safety.BanSelf(c.Request().Context(), claims.UserID, groupID); err != nil {
			// The member has already left; failing the request now would misreport what happened.
			// The ban is recoverable by leaving again with the flag.
			log.Error().Err(err).Str("group_id", groupID).Msg("failed to record permanent departure")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]bool{"left": true, "permanent": permanent},
	})
}

type updateGroupRequest struct {
	KindOnly         *bool   `json:"kind_only"`
	Name             *string `json:"name" validate:"omitempty,min=3,max=40"`
	TelegramUsername *string `json:"telegram_username"`
}

func (h *Handler) UpdateGroup(c echo.Context) error {
	groupID := c.Param("id")
	var req updateGroupRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	claims := appmw.GetCurrentUser(c)

	// Applied before the rest so a request carrying only kind_only still takes effect.
	if req.KindOnly != nil {
		if err := h.svc.SetKindOnly(c.Request().Context(), claims.UserID, groupID, *req.KindOnly); err != nil {
			return mapServiceError(c, err)
		}
	}

	group, err := h.svc.UpdateGroup(c.Request().Context(), groupsvc.UpdateGroupParams{
		UserID:           claims.UserID,
		GroupID:          groupID,
		Name:             req.Name,
		TelegramUsername: req.TelegramUsername,
	})
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"group": toGroupDto(*group)}})
}

func (h *Handler) RegenerateInviteLink(c echo.Context) error {
	groupID := c.Param("id")
	claims := appmw.GetCurrentUser(c)

	inviteURL, err := h.svc.RegenerateInviteLink(c.Request().Context(), claims.UserID, groupID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"invite_url": inviteURL}})
}

// --- Helpers ---

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, groupsvc.ErrGroupNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Group not found")
	case errors.Is(err, groupsvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	case errors.Is(err, groupsvc.ErrAlreadyMember):
		return handler.ErrorResponse(c, http.StatusConflict, "ALREADY_MEMBER", "You are already a member of this group")
	case errors.Is(err, groupsvc.ErrGroupLimitUser):
		return handler.ErrorResponse(c, http.StatusConflict, "GROUP_LIMIT", "You have reached the maximum number of groups (10)")
	case errors.Is(err, groupsvc.ErrGroupLimitSize):
		return handler.ErrorResponse(c, http.StatusConflict, "MEMBER_LIMIT", "This group has reached the maximum number of members (50)")
	case errors.Is(err, groupsvc.ErrNotAdmin):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_ADMIN", "Only the group admin can perform this action")
	case errors.Is(err, groupsvc.ErrInvalidName):
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Group name must be 3-40 characters")
	case errors.Is(err, groupsvc.ErrNoCategories):
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "At least one category is required")
	case errors.Is(err, groupsvc.ErrInvalidCategory):
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid question category")
	case errors.Is(err, groupsvc.ErrGroupBanned):
		return handler.ErrorResponse(c, http.StatusForbidden, "GROUP_BANNED",
			"Вы больше не можете вступить в эту группу")
	case errors.Is(err, groupsvc.ErrNoKindCategories):
		// A distinct code rather than a generic VALIDATION: the client can point at the switch or at
		// the categories, which is the only way the member can resolve it.
		return handler.ErrorResponse(c, http.StatusBadRequest, "NO_KIND_CATEGORIES",
			"В режиме «только добрые вопросы» эти категории недоступны — выберите ещё одну")
	case errors.Is(err, groupsvc.ErrRomanceBlocked):
		return handler.ErrorResponse(c, http.StatusForbidden, "ROMANCE_BLOCKED", "ROMANCE category is not available for users under 18")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
