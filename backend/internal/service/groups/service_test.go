package groups

import (
	"context"
	"database/sql"
	"testing"
	"time"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func TestValidateCategories_Valid(t *testing.T) {
	cats, err := validateCategories([]string{"HOT", "FUNNY", "SKILLS"}, sql.NullInt32{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 3 {
		t.Errorf("expected 3 categories, got %d", len(cats))
	}
	if cats[0] != db.QuestionCategoryHOT {
		t.Errorf("expected HOT, got %s", cats[0])
	}
}

func TestValidateCategories_Invalid(t *testing.T) {
	_, err := validateCategories([]string{"HOT", "INVALID"}, sql.NullInt32{})
	if err != ErrInvalidCategory {
		t.Errorf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestValidateCategories_RomanceBlocked(t *testing.T) {
	birthYear := sql.NullInt32{Int32: int32(time.Now().Year() - 14), Valid: true}
	_, err := validateCategories([]string{"ROMANCE"}, birthYear)
	if err != ErrRomanceBlocked {
		t.Errorf("expected ErrRomanceBlocked, got %v", err)
	}
}

func TestValidateCategories_RomanceAllowed(t *testing.T) {
	birthYear := sql.NullInt32{Int32: 2000, Valid: true}
	cats, err := validateCategories([]string{"ROMANCE"}, birthYear)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 1 || cats[0] != db.QuestionCategoryROMANCE {
		t.Errorf("expected [ROMANCE], got %v", cats)
	}
}

func TestValidateCategories_RomanceBlockedWhenNoBirthYear(t *testing.T) {
	_, err := validateCategories([]string{"ROMANCE"}, sql.NullInt32{})
	if err != ErrRomanceBlocked {
		t.Errorf("expected ErrRomanceBlocked, got %v", err)
	}
}

func TestCategoryStrings(t *testing.T) {
	cats := []db.QuestionCategory{db.QuestionCategoryHOT, db.QuestionCategoryFUNNY}
	strs := categoryStrings(cats)
	if len(strs) != 2 || strs[0] != "HOT" || strs[1] != "FUNNY" {
		t.Errorf("expected [HOT FUNNY], got %v", strs)
	}
}

func TestGetNextSeasonDates(t *testing.T) {
	startsAt, revealAt, endsAt := getNextSeasonDates()

	if startsAt.IsZero() || revealAt.IsZero() || endsAt.IsZero() {
		t.Error("dates should not be zero")
	}

	if !revealAt.After(startsAt) {
		t.Error("revealAt should be after startsAt")
	}

	if !endsAt.After(revealAt) {
		t.Error("endsAt should be after revealAt")
	}

	// Reveal should be on a Friday at 17:00 UTC
	if revealAt.Weekday() != time.Friday {
		t.Errorf("revealAt should be Friday, got %s", revealAt.Weekday())
	}
	if revealAt.Hour() != 17 {
		t.Errorf("revealAt hour should be 17, got %d", revealAt.Hour())
	}
}

func TestGetNextSeasonDates_EndsOnSunday(t *testing.T) {
	_, _, endsAt := getNextSeasonDates()

	if endsAt.Weekday() != time.Sunday {
		t.Errorf("endsAt should be Sunday, got %s", endsAt.Weekday())
	}
}

func TestCreateGroupParams_Validation(t *testing.T) {
	svc := NewService(nil, nil)

	tests := []struct {
		name string
		p    CreateGroupParams
		err  error
	}{
		{"name too short", CreateGroupParams{Name: "ab", Categories: []string{"HOT"}}, ErrInvalidName},
		{"name too long", CreateGroupParams{Name: string(make([]byte, 41)), Categories: []string{"HOT"}}, ErrInvalidName},
		{"no categories", CreateGroupParams{Name: "Valid", Categories: []string{}}, ErrNoCategories},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateGroup(context.Background(), tt.p)
			if err != tt.err {
				t.Errorf("expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if MaxGroupsPerUser != 10 {
		t.Errorf("MaxGroupsPerUser should be 10, got %d", MaxGroupsPerUser)
	}
	if MaxMembersPerGroup != 50 {
		t.Errorf("MaxMembersPerGroup should be 50, got %d", MaxMembersPerGroup)
	}
	if MinSeasonQuestions != 5 {
		t.Errorf("MinSeasonQuestions should be 5, got %d", MinSeasonQuestions)
	}
	if MaxSeasonQuestions != 10 {
		t.Errorf("MaxSeasonQuestions should be 10, got %d", MaxSeasonQuestions)
	}
}
