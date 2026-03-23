package questions

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/handler"
	appmw "github.com/repa-app/repa/internal/middleware"
	questionsvc "github.com/repa-app/repa/internal/service/questions"
)

type Handler struct {
	svc     *questionsvc.Service
	queries *db.Queries
}

func NewHandler(svc *questionsvc.Service, queries *db.Queries) *Handler {
	return &Handler{svc: svc, queries: queries}
}

// --- DTOs ---

type questionDto struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	Category  string  `json:"category"`
	Source    string  `json:"source"`
	AuthorID  *string `json:"author_id"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func toQuestionDto(q db.Question) questionDto {
	dto := questionDto{
		ID:        q.ID,
		Text:      q.Text,
		Category:  string(q.Category),
		Source:    string(q.Source),
		Status:    string(q.Status),
		CreatedAt: q.CreatedAt.Format(time.RFC3339),
	}
	if q.AuthorID.Valid {
		dto.AuthorID = &q.AuthorID.String
	}
	return dto
}

// --- Handlers ---

type createQuestionRequest struct {
	Text     string `json:"text" validate:"required"`
	Category string `json:"category" validate:"required,oneof=HOT FUNNY SECRETS SKILLS ROMANCE STUDY"`
}

func (h *Handler) CreateQuestion(c echo.Context) error {
	user := appmw.GetCurrentUser(c)
	groupID := c.Param("groupId")

	var req createQuestionRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Check membership
	memberCount, err := h.queries.IsGroupMember(c.Request().Context(), db.IsGroupMemberParams{
		UserID:  user.UserID,
		GroupID: groupID,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
	}
	if memberCount == 0 {
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	}

	result, err := h.svc.CreateQuestion(c.Request().Context(), user.UserID, groupID, req.Text, db.QuestionCategory(req.Category))
	if err != nil {
		switch {
		case errors.Is(err, questionsvc.ErrQuestionLimit):
			return handler.ErrorResponse(c, http.StatusConflict, "QUESTION_LIMIT", err.Error())
		case errors.Is(err, questionsvc.ErrTextLength):
			return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", err.Error())
		default:
			return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		}
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": map[string]any{
			"question": toQuestionDto(result.Question),
			"moderation": map[string]any{
				"approved": result.Approved,
				"reason":   result.Reason,
			},
		},
	})
}

func (h *Handler) ListQuestions(c echo.Context) error {
	user := appmw.GetCurrentUser(c)
	groupID := c.Param("groupId")

	memberCount, err := h.queries.IsGroupMember(c.Request().Context(), db.IsGroupMemberParams{
		UserID:  user.UserID,
		GroupID: groupID,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
	}
	if memberCount == 0 {
		return handler.ErrorResponse(c, http.StatusForbidden, "NOT_MEMBER", "You are not a member of this group")
	}

	questions, err := h.svc.ListGroupQuestions(c.Request().Context(), groupID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
	}

	dtos := make([]questionDto, len(questions))
	for i, q := range questions {
		dtos[i] = toQuestionDto(q)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"questions": dtos},
	})
}

func (h *Handler) DeleteQuestion(c echo.Context) error {
	user := appmw.GetCurrentUser(c)
	groupID := c.Param("groupId")
	questionID := c.Param("questionId")

	// Check if user is admin
	group, err := h.queries.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Group not found")
	}
	isAdmin := group.AdminID == user.UserID

	err = h.svc.DeleteQuestion(c.Request().Context(), user.UserID, groupID, questionID, isAdmin)
	if err != nil {
		switch {
		case errors.Is(err, questionsvc.ErrQuestionNotFound):
			return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Question not found")
		case errors.Is(err, questionsvc.ErrNotGroupQuestion):
			return handler.ErrorResponse(c, http.StatusBadRequest, "NOT_GROUP_QUESTION", err.Error())
		case errors.Is(err, questionsvc.ErrNotAuthorized):
			return handler.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"deleted": true},
	})
}

type reportRequest struct {
	Reason *string `json:"reason"`
}

func (h *Handler) ReportQuestion(c echo.Context) error {
	user := appmw.GetCurrentUser(c)
	questionID := c.Param("questionId")

	var req reportRequest
	_ = c.Bind(&req)

	err := h.svc.ReportQuestion(c.Request().Context(), user.UserID, questionID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, questionsvc.ErrAlreadyReported):
			return handler.ErrorResponse(c, http.StatusConflict, "ALREADY_REPORTED", err.Error())
		default:
			return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"reported": true},
	})
}
