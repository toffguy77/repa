package voting

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	votingsvc "github.com/repa-app/repa/internal/service/voting"
)

type Handler struct {
	svc *votingsvc.Service
}

func NewHandler(svc *votingsvc.Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs ---

type votingQuestionDto struct {
	QuestionID string `json:"question_id"`
	Text       string `json:"text"`
	Category   string `json:"category"`
	Answered   bool   `json:"answered"`
}

type targetDto struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarURL   *string `json:"avatar_url"`
}

type progressDto struct {
	Answered int `json:"answered"`
	Total    int `json:"total"`
}

// --- Handlers ---

func (h *Handler) GetVotingSession(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	session, err := h.svc.GetVotingSession(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	questions := make([]votingQuestionDto, len(session.Questions))
	for i, q := range session.Questions {
		questions[i] = votingQuestionDto{
			QuestionID: q.QuestionID,
			Text:       q.Text,
			Category:   q.Category,
			Answered:   q.Answered,
		}
	}

	targets := make([]targetDto, len(session.Targets))
	for i, t := range session.Targets {
		td := targetDto{
			UserID:   t.UserID,
			Username: t.Username,
		}
		if t.AvatarEmoji.Valid {
			td.AvatarEmoji = &t.AvatarEmoji.String
		}
		if t.AvatarURL.Valid {
			td.AvatarURL = &t.AvatarURL.String
		}
		targets[i] = td
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"season_id": session.SeasonID,
			"questions": questions,
			"targets":   targets,
			"progress": progressDto{
				Answered: session.Answered,
				Total:    session.Total,
			},
		},
	})
}

type castVoteRequest struct {
	QuestionID string `json:"question_id" validate:"required"`
	TargetID   string `json:"target_id" validate:"required"`
}

func (h *Handler) CastVote(c echo.Context) error {
	var req castVoteRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	result, err := h.svc.CastVote(c.Request().Context(), seasonID, claims.UserID, req.QuestionID, req.TargetID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": map[string]any{
			"vote": map[string]string{
				"question_id": result.QuestionID,
				"target_id":   result.TargetID,
			},
			"progress": progressDto{
				Answered: result.Answered,
				Total:    result.Total,
			},
		},
	})
}

func (h *Handler) GetProgress(c echo.Context) error {
	seasonID := c.Param("seasonId")
	claims := appmw.GetCurrentUser(c)

	progress, err := h.svc.GetProgress(c.Request().Context(), seasonID, claims.UserID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{
			"voted_count":      progress.VotedCount,
			"total_count":      progress.TotalCount,
			"quorum_reached":   progress.QuorumReached,
			"quorum_threshold": progress.QuorumThreshold,
			"user_voted":       progress.UserVoted,
		},
	})
}

// --- Helpers ---

func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, votingsvc.ErrSeasonNotFound):
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Season not found")
	case errors.Is(err, votingsvc.ErrSeasonNotVoting):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SEASON_NOT_VOTING", "Season is not in voting phase")
	case errors.Is(err, votingsvc.ErrNotMember):
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	case errors.Is(err, votingsvc.ErrAlreadyVoted):
		return handler.ErrorResponse(c, http.StatusConflict, "ALREADY_VOTED", "You already voted for this question")
	case errors.Is(err, votingsvc.ErrSelfVote):
		return handler.ErrorResponse(c, http.StatusBadRequest, "SELF_VOTE", "You cannot vote for yourself")
	case errors.Is(err, votingsvc.ErrTargetNotMember):
		return handler.ErrorResponse(c, http.StatusBadRequest, "TARGET_NOT_MEMBER", "Target is not a member of this group")
	case errors.Is(err, votingsvc.ErrInvalidQuestion):
		return handler.ErrorResponse(c, http.StatusBadRequest, "INVALID_QUESTION", "Question is not part of this season")
	default:
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Something went wrong")
	}
}
