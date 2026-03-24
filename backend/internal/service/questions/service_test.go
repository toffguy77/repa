package questions

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
)

// --- Mock Querier ---

type mockQuerier struct {
	db.Querier
	questionCounts map[string]int64          // "authorID:groupID" -> count
	questions      map[string]db.Question    // questionID -> Question
	groups         map[string]db.Group       // groupID -> Group
	groupQuestions map[string][]db.Question  // groupID -> questions
	reported       map[string]bool           // "questionID:reporterID" -> reported
	createdQuestions []db.CreateQuestionParams
	createdReports   []db.CreateReportParams
	updatedStatuses  []db.UpdateQuestionStatusParams

	countErr          error
	createQuestionErr error
	getGroupErr       error
	getAllQuestionsErr error
	getQuestionErr    error
	updateStatusErr   error
	hasReportedErr    error
	createReportErr   error
}

func (m *mockQuerier) CountUserQuestionsInGroup(_ context.Context, arg db.CountUserQuestionsInGroupParams) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	key := arg.AuthorID.String + ":" + arg.GroupID.String
	return m.questionCounts[key], nil
}

func (m *mockQuerier) CreateQuestion(_ context.Context, arg db.CreateQuestionParams) (db.Question, error) {
	if m.createQuestionErr != nil {
		return db.Question{}, m.createQuestionErr
	}
	m.createdQuestions = append(m.createdQuestions, arg)
	return db.Question{
		ID:       arg.ID,
		Text:     arg.Text,
		Category: arg.Category,
		Source:   arg.Source,
		GroupID:  arg.GroupID,
		AuthorID: arg.AuthorID,
		Status:   arg.Status,
	}, nil
}

func (m *mockQuerier) GetGroupByID(_ context.Context, id string) (db.Group, error) {
	if m.getGroupErr != nil {
		return db.Group{}, m.getGroupErr
	}
	g, ok := m.groups[id]
	if !ok {
		return db.Group{}, sql.ErrNoRows
	}
	return g, nil
}

func (m *mockQuerier) GetGroupAllQuestions(_ context.Context, arg db.GetGroupAllQuestionsParams) ([]db.Question, error) {
	if m.getAllQuestionsErr != nil {
		return nil, m.getAllQuestionsErr
	}
	return m.groupQuestions[arg.GroupID.String], nil
}

func (m *mockQuerier) GetQuestionByID(_ context.Context, id string) (db.Question, error) {
	if m.getQuestionErr != nil {
		return db.Question{}, m.getQuestionErr
	}
	q, ok := m.questions[id]
	if !ok {
		return db.Question{}, sql.ErrNoRows
	}
	return q, nil
}

func (m *mockQuerier) UpdateQuestionStatus(_ context.Context, arg db.UpdateQuestionStatusParams) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updatedStatuses = append(m.updatedStatuses, arg)
	return nil
}

func (m *mockQuerier) HasUserReported(_ context.Context, arg db.HasUserReportedParams) (bool, error) {
	if m.hasReportedErr != nil {
		return false, m.hasReportedErr
	}
	key := arg.QuestionID + ":" + arg.ReporterID
	return m.reported[key], nil
}

func (m *mockQuerier) CreateReport(_ context.Context, arg db.CreateReportParams) (db.Report, error) {
	if m.createReportErr != nil {
		return db.Report{}, m.createReportErr
	}
	m.createdReports = append(m.createdReports, arg)
	return db.Report{
		ID:         arg.ID,
		QuestionID: arg.QuestionID,
		ReporterID: arg.ReporterID,
		Reason:     arg.Reason,
	}, nil
}

// --- Mock Moderator ---

type mockModerator struct {
	approved bool
	reason   *string
	err      error
	called   bool
}

func (m *mockModerator) ModerateQuestion(_ context.Context, _ string) (*lib.ModerationResult, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &lib.ModerationResult{
		Approved: m.approved,
		Reason:   m.reason,
	}, nil
}

// --- Fixtures ---

func newMockQuerier() *mockQuerier {
	return &mockQuerier{
		questionCounts: map[string]int64{},
		questions:      map[string]db.Question{},
		groups:         map[string]db.Group{},
		groupQuestions: map[string][]db.Question{},
		reported:       map[string]bool{},
	}
}

func validText() string {
	return "Это валидный вопрос для теста?" // > 10 runes
}

// --- Constants and Sentinels ---

func TestMaxQuestionsPerUserPerGroup(t *testing.T) {
	if MaxQuestionsPerUserPerGroup != 5 {
		t.Errorf("expected MaxQuestionsPerUserPerGroup = 5, got %d", MaxQuestionsPerUserPerGroup)
	}
}

func TestErrorSentinelsAreDistinct(t *testing.T) {
	errs := []error{
		ErrQuestionLimit,
		ErrTextLength,
		ErrNotAuthorized,
		ErrQuestionNotFound,
		ErrAlreadyReported,
		ErrNotGroupQuestion,
	}
	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			if errors.Is(errs[i], errs[j]) {
				t.Errorf("error sentinels %d and %d should be distinct", i, j)
			}
		}
	}
}

// --- CreateQuestion tests ---

func TestCreateQuestion_TextTooShort(t *testing.T) {
	svc := NewService(newMockQuerier(), nil)
	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", "short", db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrTextLength) {
		t.Errorf("expected ErrTextLength, got %v", err)
	}
}

func TestCreateQuestion_TextTooShortExact9Runes(t *testing.T) {
	svc := NewService(newMockQuerier(), nil)
	// 9 ASCII characters
	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", "123456789", db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrTextLength) {
		t.Errorf("expected ErrTextLength for 9-rune text, got %v", err)
	}
}

func TestCreateQuestion_TextTooLong(t *testing.T) {
	svc := NewService(newMockQuerier(), nil)
	longText := strings.Repeat("a", 121)
	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", longText, db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrTextLength) {
		t.Errorf("expected ErrTextLength, got %v", err)
	}
}

func TestCreateQuestion_TextExactly10Runes(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)
	text := "1234567890" // exactly 10
	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", text, db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error for 10-rune text, got %v", err)
	}
	if result.Question.Text != text {
		t.Errorf("expected text %q, got %q", text, result.Question.Text)
	}
}

func TestCreateQuestion_TextExactly120Runes(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)
	text := strings.Repeat("x", 120)
	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", text, db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error for 120-rune text, got %v", err)
	}
	if result.Question.Text != text {
		t.Errorf("expected text length 120, got %d", len(result.Question.Text))
	}
}

func TestCreateQuestion_MultiByteRuneCounting(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)
	// 10 Cyrillic characters = 20 bytes but only 10 runes
	text := "Абвгдежзик"
	if len([]rune(text)) != 10 {
		t.Fatalf("test setup: expected 10 runes, got %d", len([]rune(text)))
	}
	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", text, db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error for 10-rune Cyrillic text, got %v", err)
	}
	if result.Question.Text != text {
		t.Errorf("expected text %q, got %q", text, result.Question.Text)
	}
}

func TestCreateQuestion_MultiByteRuneTooShort(t *testing.T) {
	svc := NewService(newMockQuerier(), nil)
	// 9 Cyrillic characters = 18 bytes but only 9 runes
	text := "Абвгдежзи"
	if len([]rune(text)) != 9 {
		t.Fatalf("test setup: expected 9 runes, got %d", len([]rune(text)))
	}
	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", text, db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrTextLength) {
		t.Errorf("expected ErrTextLength for 9-rune Cyrillic text, got %v", err)
	}
}

func TestCreateQuestion_QuestionLimit(t *testing.T) {
	m := newMockQuerier()
	m.questionCounts["u1:g1"] = 5
	svc := NewService(m, nil)

	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrQuestionLimit) {
		t.Errorf("expected ErrQuestionLimit, got %v", err)
	}
}

func TestCreateQuestion_QuestionLimitExceeded(t *testing.T) {
	m := newMockQuerier()
	m.questionCounts["u1:g1"] = 10
	svc := NewService(m, nil)

	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrQuestionLimit) {
		t.Errorf("expected ErrQuestionLimit, got %v", err)
	}
}

func TestCreateQuestion_JustUnderLimit(t *testing.T) {
	m := newMockQuerier()
	m.questionCounts["u1:g1"] = 4
	svc := NewService(m, nil)

	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error at count 4, got %v", err)
	}
	if !result.Approved {
		t.Error("expected approved=true when no moderator")
	}
}

func TestCreateQuestion_SuccessNoModerator(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)

	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Approved {
		t.Error("expected approved=true without moderator")
	}
	if result.Reason != nil {
		t.Errorf("expected nil reason, got %v", result.Reason)
	}
	if result.Question.Status != db.QuestionStatusACTIVE {
		t.Errorf("expected ACTIVE status, got %s", result.Question.Status)
	}
	if result.Question.Source != db.QuestionSourceUSER {
		t.Errorf("expected USER source, got %s", result.Question.Source)
	}
	if len(m.createdQuestions) != 1 {
		t.Fatalf("expected 1 created question, got %d", len(m.createdQuestions))
	}
	cq := m.createdQuestions[0]
	if cq.AuthorID.String != "u1" || !cq.AuthorID.Valid {
		t.Errorf("expected authorID u1, got %v", cq.AuthorID)
	}
	if cq.GroupID.String != "g1" || !cq.GroupID.Valid {
		t.Errorf("expected groupID g1, got %v", cq.GroupID)
	}
	if cq.Category != db.QuestionCategoryFUNNY {
		t.Errorf("expected FUNNY category, got %s", cq.Category)
	}
}

func TestCreateQuestion_ModerationApproved(t *testing.T) {
	m := newMockQuerier()
	mod := &mockModerator{approved: true}
	svc := NewService(m, mod)

	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !mod.called {
		t.Error("expected moderator to be called")
	}
	if !result.Approved {
		t.Error("expected approved=true")
	}
	if result.Question.Status != db.QuestionStatusACTIVE {
		t.Errorf("expected ACTIVE status, got %s", result.Question.Status)
	}
}

func TestCreateQuestion_ModerationRejected(t *testing.T) {
	m := newMockQuerier()
	reason := "inappropriate content"
	mod := &mockModerator{approved: false, reason: &reason}
	svc := NewService(m, mod)

	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Approved {
		t.Error("expected approved=false")
	}
	if result.Reason == nil || *result.Reason != reason {
		t.Errorf("expected reason %q, got %v", reason, result.Reason)
	}
	if result.Question.Status != db.QuestionStatusREJECTED {
		t.Errorf("expected REJECTED status, got %s", result.Question.Status)
	}
}

func TestCreateQuestion_ModerationError(t *testing.T) {
	m := newMockQuerier()
	mod := &mockModerator{err: errors.New("API timeout")}
	svc := NewService(m, mod)

	result, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err != nil {
		t.Fatalf("expected no error (should fallback to PENDING), got %v", err)
	}
	if result.Approved {
		t.Error("expected approved=false on moderation error")
	}
	if result.Question.Status != db.QuestionStatusPENDING {
		t.Errorf("expected PENDING status on moderation error, got %s", result.Question.Status)
	}
}

func TestCreateQuestion_CountError(t *testing.T) {
	m := newMockQuerier()
	m.countErr = errors.New("db error")
	svc := NewService(m, nil)

	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestCreateQuestion_CreateDBError(t *testing.T) {
	m := newMockQuerier()
	m.createQuestionErr = errors.New("insert failed")
	svc := NewService(m, nil)

	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", validText(), db.QuestionCategoryFUNNY)
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected insert error, got %v", err)
	}
}

func TestCreateQuestion_EmptyText(t *testing.T) {
	svc := NewService(newMockQuerier(), nil)
	_, err := svc.CreateQuestion(context.Background(), "u1", "g1", "", db.QuestionCategoryFUNNY)
	if !errors.Is(err, ErrTextLength) {
		t.Errorf("expected ErrTextLength for empty text, got %v", err)
	}
}

// --- ListGroupQuestions tests ---

func TestListGroupQuestions_Success(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{
		ID:         "g1",
		Categories: []string{"FUNNY", "HOT"},
	}
	expectedQuestions := []db.Question{
		{ID: "q1", Text: "Question 1", Category: db.QuestionCategoryFUNNY},
		{ID: "q2", Text: "Question 2", Category: db.QuestionCategoryHOT},
	}
	m.groupQuestions["g1"] = expectedQuestions
	svc := NewService(m, nil)

	questions, err := svc.ListGroupQuestions(context.Background(), "g1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(questions))
	}
}

func TestListGroupQuestions_GroupNotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)

	_, err := svc.ListGroupQuestions(context.Background(), "nonexistent")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestListGroupQuestions_GetGroupError(t *testing.T) {
	m := newMockQuerier()
	m.getGroupErr = errors.New("connection lost")
	svc := NewService(m, nil)

	_, err := svc.ListGroupQuestions(context.Background(), "g1")
	if err == nil || err.Error() != "connection lost" {
		t.Errorf("expected connection error, got %v", err)
	}
}

func TestListGroupQuestions_GetAllQuestionsError(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", Categories: []string{"FUNNY"}}
	m.getAllQuestionsErr = errors.New("query failed")
	svc := NewService(m, nil)

	_, err := svc.ListGroupQuestions(context.Background(), "g1")
	if err == nil || err.Error() != "query failed" {
		t.Errorf("expected query error, got %v", err)
	}
}

func TestListGroupQuestions_EmptyCategories(t *testing.T) {
	m := newMockQuerier()
	m.groups["g1"] = db.Group{ID: "g1", Categories: []string{}}
	svc := NewService(m, nil)

	questions, err := svc.ListGroupQuestions(context.Background(), "g1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if questions == nil {
		// nil is acceptable for empty result
	}
}

// --- DeleteQuestion tests ---

func TestDeleteQuestion_SuccessAsAuthor(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g1", Valid: true},
		AuthorID: sql.NullString{String: "u1", Valid: true},
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(m.updatedStatuses) != 1 {
		t.Fatalf("expected 1 status update, got %d", len(m.updatedStatuses))
	}
	if m.updatedStatuses[0].ID != "q1" {
		t.Errorf("expected question ID q1, got %s", m.updatedStatuses[0].ID)
	}
	if m.updatedStatuses[0].Status != db.QuestionStatusREJECTED {
		t.Errorf("expected REJECTED status, got %s", m.updatedStatuses[0].Status)
	}
}

func TestDeleteQuestion_SuccessAsAdmin(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g1", Valid: true},
		AuthorID: sql.NullString{String: "u2", Valid: true}, // different author
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", true)
	if err != nil {
		t.Fatalf("expected no error as admin, got %v", err)
	}
}

func TestDeleteQuestion_NotFound(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "nonexistent", false)
	if !errors.Is(err, ErrQuestionNotFound) {
		t.Errorf("expected ErrQuestionNotFound, got %v", err)
	}
}

func TestDeleteQuestion_NotGroupQuestion(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g2", Valid: true}, // different group
		AuthorID: sql.NullString{String: "u1", Valid: true},
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if !errors.Is(err, ErrNotGroupQuestion) {
		t.Errorf("expected ErrNotGroupQuestion, got %v", err)
	}
}

func TestDeleteQuestion_NotGroupQuestionNullGroupID(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{Valid: false}, // system question, no group
		AuthorID: sql.NullString{String: "u1", Valid: true},
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if !errors.Is(err, ErrNotGroupQuestion) {
		t.Errorf("expected ErrNotGroupQuestion for null group_id, got %v", err)
	}
}

func TestDeleteQuestion_NotAuthorized(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g1", Valid: true},
		AuthorID: sql.NullString{String: "u2", Valid: true}, // different author
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if !errors.Is(err, ErrNotAuthorized) {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestDeleteQuestion_NotAuthorizedNullAuthor(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g1", Valid: true},
		AuthorID: sql.NullString{Valid: false}, // system question, no author
	}
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if !errors.Is(err, ErrNotAuthorized) {
		t.Errorf("expected ErrNotAuthorized for null author_id, got %v", err)
	}
}

func TestDeleteQuestion_GetQuestionDBError(t *testing.T) {
	m := newMockQuerier()
	m.getQuestionErr = errors.New("db connection error")
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if err == nil || err.Error() != "db connection error" {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestDeleteQuestion_UpdateStatusError(t *testing.T) {
	m := newMockQuerier()
	m.questions["q1"] = db.Question{
		ID:       "q1",
		GroupID:  sql.NullString{String: "g1", Valid: true},
		AuthorID: sql.NullString{String: "u1", Valid: true},
	}
	m.updateStatusErr = errors.New("update failed")
	svc := NewService(m, nil)

	err := svc.DeleteQuestion(context.Background(), "u1", "g1", "q1", false)
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected update error, got %v", err)
	}
}

// --- ReportQuestion tests ---

func TestReportQuestion_Success(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)
	reason := "offensive"

	err := svc.ReportQuestion(context.Background(), "u1", "q1", &reason)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(m.createdReports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(m.createdReports))
	}
	r := m.createdReports[0]
	if r.QuestionID != "q1" {
		t.Errorf("expected questionID q1, got %s", r.QuestionID)
	}
	if r.ReporterID != "u1" {
		t.Errorf("expected reporterID u1, got %s", r.ReporterID)
	}
	if !r.Reason.Valid || r.Reason.String != "offensive" {
		t.Errorf("expected reason 'offensive', got %v", r.Reason)
	}
}

func TestReportQuestion_SuccessNilReason(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)

	err := svc.ReportQuestion(context.Background(), "u1", "q1", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(m.createdReports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(m.createdReports))
	}
	if m.createdReports[0].Reason.Valid {
		t.Error("expected null reason when nil is passed")
	}
}

func TestReportQuestion_AlreadyReported(t *testing.T) {
	m := newMockQuerier()
	m.reported["q1:u1"] = true
	svc := NewService(m, nil)

	err := svc.ReportQuestion(context.Background(), "u1", "q1", nil)
	if !errors.Is(err, ErrAlreadyReported) {
		t.Errorf("expected ErrAlreadyReported, got %v", err)
	}
}

func TestReportQuestion_HasReportedDBError(t *testing.T) {
	m := newMockQuerier()
	m.hasReportedErr = errors.New("db error")
	svc := NewService(m, nil)

	err := svc.ReportQuestion(context.Background(), "u1", "q1", nil)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestReportQuestion_CreateReportDBError(t *testing.T) {
	m := newMockQuerier()
	m.createReportErr = errors.New("insert failed")
	svc := NewService(m, nil)

	err := svc.ReportQuestion(context.Background(), "u1", "q1", nil)
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected insert error, got %v", err)
	}
}

// --- NewService test ---

func TestNewService(t *testing.T) {
	m := newMockQuerier()
	mod := &mockModerator{approved: true}
	svc := NewService(m, mod)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewService_NilModerator(t *testing.T) {
	m := newMockQuerier()
	svc := NewService(m, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
