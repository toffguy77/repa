package groups

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	db "github.com/repa-app/repa/internal/db/sqlc"
)

// stopHere is returned from the invite-code probe to end CreateGroup before it opens a transaction.
// The probe is the first call after the kind-only guard, so reaching it *is* the assertion that the
// guard let the group through.
var stopHere = errors.New("stop: reached the invite-code probe")

func birthYear(age int) sql.NullInt32 {
	// Computed from the current year rather than written as a literal: a literal silently becomes a
	// different age every January.
	return sql.NullInt32{Int32: int32(time.Now().Year() - age), Valid: true}
}

func TestDefaultKindOnly(t *testing.T) {
	yes, no := true, false

	tests := []struct {
		name      string
		explicit  *bool
		birthYear sql.NullInt32
		want      bool
	}{
		{"a 15-year-old creator gets it on", nil, birthYear(15), true},
		{"a 17-year-old creator gets it on", nil, birthYear(17), true},
		{"an 18-year-old creator gets it off", nil, birthYear(18), false},
		{"a 30-year-old creator gets it off", nil, birthYear(30), false},
		// Unknown age is treated as under 18, matching how ROMANCE is handled — one notion of
		// "under 18" in the product.
		{"an unknown birth year is treated as under 18", nil, sql.NullInt32{}, true},
		{"a minor may turn it off", &no, birthYear(15), false},
		{"an adult may turn it on", &yes, birthYear(30), true},
		{"an explicit value wins with an unknown age", &no, sql.NullInt32{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultKindOnly(tt.explicit, tt.birthYear); got != tt.want {
				t.Errorf("defaultKindOnly = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- CreateGroup: the guard against an empty bank ---

func TestCreateGroup_KindOnlyWithNoKindQuestionsIsRefused(t *testing.T) {
	svc, mock := newMockService(t)
	on := true

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	// HOT and SECRETS are edgy end to end, so the bank offers this group nothing.
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "9Б",
		Categories:    []string{"HOT", "SECRETS"},
		UserID:        "user-1",
		UserBirthYear: birthYear(15),
		KindOnly:      &on,
	})
	if !errors.Is(err, ErrNoKindCategories) {
		t.Fatalf("expected ErrNoKindCategories, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateGroup_OneKindCategoryIsEnough(t *testing.T) {
	svc, mock := newMockService(t)
	on := true

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(50)))
	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(stopHere)

	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "9Б",
		Categories:    []string{"HOT", "FUNNY"},
		UserID:        "user-1",
		UserBirthYear: birthYear(15),
		KindOnly:      &on,
	})
	if !errors.Is(err, stopHere) {
		t.Fatalf("expected creation to get past the guard, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateGroup_OrdinaryGroupSkipsTheKindCheck(t *testing.T) {
	svc, mock := newMockService(t)
	off := false

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	// No question count expected: a group that did not ask for kind questions must not pay for the
	// check. Expectations are matched in order against the SQL, so a stray count here fails the test.
	mock.ExpectQuery("SELECT .+ FROM groups WHERE upper.invite_code.").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(stopHere)

	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "Поток",
		Categories:    []string{"HOT", "SECRETS"},
		UserID:        "user-1",
		UserBirthYear: birthYear(30),
		KindOnly:      &off,
	})
	if !errors.Is(err, stopHere) {
		t.Fatalf("expected creation to skip the guard, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- SetKindOnly ---

// kindOnlyGroupRow is a group row with chosen categories and setting, for the update tests.
func kindOnlyGroupRow(adminID, categories string, kindOnly bool) *sqlmock.Rows {
	return sqlmock.NewRows(groupCols).AddRow(
		"group-1", "9Б", "ABC123", adminID,
		nil, nil, nil, nil,
		time.Now(), categories, kindOnly,
	)
}

func TestSetKindOnly_AdminTurnsItOn(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(kindOnlyGroupRow("admin-1", `{"HOT","FUNNY"}`, false))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(50)))
	mock.ExpectExec("UPDATE groups SET kind_only").
		WithArgs("group-1", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.SetKindOnly(context.Background(), "admin-1", "group-1", true); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// This is also the open-season assertion: the three expectations above are the *only* statements
	// allowed to run, so any touch of seasons or season_questions would fail here. The setting applies
	// from the next season; the open one is left exactly as members found it.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSetKindOnly_RefusedWhenItWouldEmptyTheBank(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(kindOnlyGroupRow("admin-1", `{"HOT","SECRETS"}`, false))
	mock.ExpectQuery("SELECT COUNT.+ FROM questions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	err := svc.SetKindOnly(context.Background(), "admin-1", "group-1", true)
	if !errors.Is(err, ErrNoKindCategories) {
		t.Fatalf("expected ErrNoKindCategories, got %v", err)
	}
	// No UPDATE expected: the setting stays off rather than being applied and then starving the group.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSetKindOnly_TurningItOffSkipsTheCheck(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(kindOnlyGroupRow("admin-1", `{"HOT","SECRETS"}`, true))
	// Turning it off can only widen the bank, so there is nothing to check — and a group that got
	// into this state must be able to get out of it.
	mock.ExpectExec("UPDATE groups SET kind_only").
		WithArgs("group-1", false).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.SetKindOnly(context.Background(), "admin-1", "group-1", false); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSetKindOnly_NonAdminIsRefused(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(kindOnlyGroupRow("admin-1", `{"FUNNY"}`, false))

	err := svc.SetKindOnly(context.Background(), "member-2", "group-1", true)
	if !errors.Is(err, ErrNotAdmin) {
		t.Fatalf("expected ErrNotAdmin, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSetKindOnly_MissingGroup(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("nope").
		WillReturnError(sql.ErrNoRows)

	if err := svc.SetKindOnly(context.Background(), "admin-1", "nope", true); !errors.Is(err, ErrGroupNotFound) {
		t.Fatalf("expected ErrGroupNotFound, got %v", err)
	}
}

// TestSelectionPassesTheSettingToTheQuery pins the wiring between the group's setting and the
// selection query's tone argument — the one place a mistake would be invisible, since the query would
// still return questions, just the wrong ones.
func TestSelectionPassesTheSettingToTheQuery(t *testing.T) {
	for _, kindOnly := range []bool{true, false} {
		svc, mock := newMockService(t)

		mock.ExpectQuery("SELECT .+ FROM questions").
			WithArgs(sqlmock.AnyArg(), "group-1", int32(MinSeasonQuestions), kindOnly).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		err := svc.selectAndAssignQuestionsTx(context.Background(), svc.queries, "season-1", "group-1",
			[]db.QuestionCategory{db.QuestionCategoryFUNNY}, 1, kindOnly)
		if err != nil {
			t.Fatalf("kindOnly=%v: %v", kindOnly, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("kindOnly=%v: unmet expectations: %v", kindOnly, err)
		}
	}
}
