package questions

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var (
	ErrQuestionLimit   = errors.New("maximum 5 custom questions per group")
	ErrTextLength      = errors.New("question text must be 10-120 characters")
	ErrNotAuthorized   = errors.New("only admin or author can delete this question")
	ErrQuestionNotFound = errors.New("question not found")
	ErrAlreadyReported = errors.New("you already reported this question")
	ErrNotGroupQuestion = errors.New("question does not belong to this group")
)

const MaxQuestionsPerUserPerGroup = 5

type CreateResult struct {
	Question db.Question
	Approved bool
	Reason   *string
}

type Service struct {
	queries   *db.Queries
	anthropic *lib.AnthropicClient
}

func NewService(queries *db.Queries, anthropic *lib.AnthropicClient) *Service {
	return &Service{queries: queries, anthropic: anthropic}
}

func (s *Service) CreateQuestion(ctx context.Context, userID, groupID, text string, category db.QuestionCategory) (*CreateResult, error) {
	textLen := len([]rune(text))
	if textLen < 10 || textLen > 120 {
		return nil, ErrTextLength
	}

	count, err := s.queries.CountUserQuestionsInGroup(ctx, db.CountUserQuestionsInGroupParams{
		AuthorID: sql.NullString{String: userID, Valid: true},
		GroupID:  sql.NullString{String: groupID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	if count >= MaxQuestionsPerUserPerGroup {
		return nil, ErrQuestionLimit
	}

	// AI moderation
	status := db.QuestionStatusACTIVE
	var reason *string

	if s.anthropic != nil {
		modResult, err := s.anthropic.ModerateQuestion(ctx, text)
		if err != nil {
			log.Warn().Err(err).Str("text", text).Msg("moderation timeout/error, setting PENDING")
			status = db.QuestionStatusPENDING
		} else if !modResult.Approved {
			status = db.QuestionStatusREJECTED
			reason = modResult.Reason
		}
	}

	question, err := s.queries.CreateQuestion(ctx, db.CreateQuestionParams{
		ID:       uuid.New().String(),
		Text:     text,
		Category: category,
		Source:   db.QuestionSourceUSER,
		GroupID:  sql.NullString{String: groupID, Valid: true},
		AuthorID: sql.NullString{String: userID, Valid: true},
		Status:   status,
	})
	if err != nil {
		return nil, err
	}

	return &CreateResult{
		Question: question,
		Approved: status == db.QuestionStatusACTIVE,
		Reason:   reason,
	}, nil
}

func (s *Service) ListGroupQuestions(ctx context.Context, groupID string) ([]db.Question, error) {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	categories := make([]db.QuestionCategory, len(group.Categories))
	for i, c := range group.Categories {
		categories[i] = db.QuestionCategory(c)
	}

	return s.queries.GetGroupAllQuestions(ctx, db.GetGroupAllQuestionsParams{
		GroupID: sql.NullString{String: groupID, Valid: true},
		Column2: categories,
	})
}

func (s *Service) DeleteQuestion(ctx context.Context, userID, groupID, questionID string, isAdmin bool) error {
	question, err := s.queries.GetQuestionByID(ctx, questionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrQuestionNotFound
		}
		return err
	}

	if !question.GroupID.Valid || question.GroupID.String != groupID {
		return ErrNotGroupQuestion
	}

	isAuthor := question.AuthorID.Valid && question.AuthorID.String == userID
	if !isAdmin && !isAuthor {
		return ErrNotAuthorized
	}

	return s.queries.UpdateQuestionStatus(ctx, db.UpdateQuestionStatusParams{
		ID:     questionID,
		Status: db.QuestionStatusREJECTED,
	})
}

func (s *Service) ReportQuestion(ctx context.Context, userID, questionID string, reason *string) error {
	reported, err := s.queries.HasUserReported(ctx, db.HasUserReportedParams{
		QuestionID: questionID,
		ReporterID: userID,
	})
	if err != nil {
		return err
	}
	if reported {
		return ErrAlreadyReported
	}

	var reasonNull sql.NullString
	if reason != nil {
		reasonNull = sql.NullString{String: *reason, Valid: true}
	}

	_, err = s.queries.CreateReport(ctx, db.CreateReportParams{
		ID:         uuid.New().String(),
		QuestionID: questionID,
		ReporterID: userID,
		Reason:     reasonNull,
	})
	return err
}
