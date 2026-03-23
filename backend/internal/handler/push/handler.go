package push

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/handler"
	"github.com/repa-app/repa/internal/middleware"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

type RegisterTokenRequest struct {
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android"`
}

func (h *Handler) RegisterToken(c echo.Context) error {
	userID := middleware.GetCurrentUser(c).UserID

	var req RegisterTokenRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	_, err := h.queries.UpsertFCMToken(c.Request().Context(), db.UpsertFCMTokenParams{
		ID:       uuid.New().String(),
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to register token")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]bool{"registered": true},
	})
}

// --- Next-season question voting ---

type questionDto struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Category string `json:"category"`
}

// GetQuestionCandidates returns 3 random system questions for next-season voting.
func (h *Handler) GetQuestionCandidates(c echo.Context) error {
	userID := middleware.GetCurrentUser(c).UserID
	groupID := c.Param("id")
	ctx := c.Request().Context()

	// Verify membership
	isMember, err := h.queries.IsGroupMember(ctx, db.IsGroupMemberParams{UserID: userID, GroupID: groupID})
	if err != nil || isMember == 0 {
		return handler.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Not a group member")
	}

	questions, err := h.queries.GetRandomSystemQuestions(ctx, 3)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to get question candidates")
	}

	dtos := make([]questionDto, len(questions))
	for i, q := range questions {
		dtos[i] = questionDto{ID: q.ID, Text: q.Text, Category: string(q.Category)}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"candidates": dtos},
	})
}

type VoteQuestionRequest struct {
	QuestionID string `json:"questionId" validate:"required"`
}

// VoteQuestion casts a vote for a question in the next season.
func (h *Handler) VoteQuestion(c echo.Context) error {
	userID := middleware.GetCurrentUser(c).UserID
	groupID := c.Param("id")
	ctx := c.Request().Context()

	var req VoteQuestionRequest
	if err := c.Bind(&req); err != nil {
		return handler.ErrorResponse(c, http.StatusBadRequest, "VALIDATION", "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Verify membership
	isMember, err := h.queries.IsGroupMember(ctx, db.IsGroupMemberParams{UserID: userID, GroupID: groupID})
	if err != nil || isMember == 0 {
		return handler.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Not a group member")
	}

	// Get next season number
	lastNumber, err := h.queries.GetLastSeasonNumber(ctx, groupID)
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to get season number")
	}
	nextSeasonNumber := lastNumber + 1

	// Check if already voted
	count, err := h.queries.HasNextSeasonVote(ctx, db.HasNextSeasonVoteParams{
		GroupID:      groupID,
		UserID:       userID,
		SeasonNumber: nextSeasonNumber,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to check vote")
	}
	if count > 0 {
		return handler.ErrorResponse(c, http.StatusConflict, "ALREADY_VOTED", "Already voted for next season question")
	}

	// Verify question exists
	if _, err := h.queries.GetQuestionByID(ctx, req.QuestionID); err != nil {
		return handler.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Question not found")
	}

	_, err = h.queries.CreateNextSeasonVote(ctx, db.CreateNextSeasonVoteParams{
		ID:           uuid.New().String(),
		GroupID:      groupID,
		UserID:       userID,
		QuestionID:   req.QuestionID,
		SeasonNumber: nextSeasonNumber,
	})
	if err != nil {
		return handler.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL", "Failed to cast vote")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]bool{"voted": true},
	})
}
