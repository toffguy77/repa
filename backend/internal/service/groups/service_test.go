package groups

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

// --- CreateGroup: valid params but nil sqlDB panics after validation passes ---

func TestCreateGroup_ValidParamsNilDB(t *testing.T) {
	svc := NewService(nil, nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when sqlDB is nil and validation passes, but no panic occurred")
		}
	}()

	birthYear := sql.NullInt32{Int32: 2000, Valid: true}
	_, _ = svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "Valid Group",
		Categories:    []string{"HOT"},
		UserID:        "user-1",
		UserBirthYear: birthYear,
	})
}

// --- CreateGroup name boundary: exactly MinGroupName and MaxGroupName chars ---

func TestCreateGroup_NameBoundary(t *testing.T) {
	svc := NewService(nil, nil)

	// Exactly 3 chars (MinGroupName) — should pass name validation, fail on nil queries
	t.Run("exactly MinGroupName chars", func(t *testing.T) {
		defer func() { recover() }() // nil queries will panic
		_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
			Name:          "Abc",
			Categories:    []string{"HOT"},
			UserID:        "user-1",
			UserBirthYear: sql.NullInt32{Int32: 2000, Valid: true},
		})
		// If we got ErrInvalidName, name validation rejected it incorrectly
		if err == ErrInvalidName {
			t.Error("name of exactly 3 chars should not be rejected as invalid")
		}
	})

	// Exactly 40 chars (MaxGroupName) — should pass name validation
	t.Run("exactly MaxGroupName chars", func(t *testing.T) {
		defer func() { recover() }()
		name40 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ12345678901234" // 40 chars
		if len(name40) != 40 {
			t.Fatalf("test setup: expected 40 chars, got %d", len(name40))
		}
		_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
			Name:          name40,
			Categories:    []string{"HOT"},
			UserID:        "user-1",
			UserBirthYear: sql.NullInt32{Int32: 2000, Valid: true},
		})
		if err == ErrInvalidName {
			t.Error("name of exactly 40 chars should not be rejected as invalid")
		}
	})

	// 2 chars — rejected
	t.Run("below MinGroupName", func(t *testing.T) {
		_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
			Name:       "Ab",
			Categories: []string{"HOT"},
		})
		if err != ErrInvalidName {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
	})

	// 41 chars — rejected
	t.Run("above MaxGroupName", func(t *testing.T) {
		name41 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789012345" // 41 chars
		_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
			Name:       name41,
			Categories: []string{"HOT"},
		})
		if err != ErrInvalidName {
			t.Errorf("expected ErrInvalidName, got %v", err)
		}
	})
}

// --- getNextSeasonDates: startsAt is always a Monday ---

func TestGetNextSeasonDates_StartsAtIsMonday(t *testing.T) {
	startsAt, _, _ := getNextSeasonDates()

	msk := time.FixedZone("MSK", 3*60*60)
	startsAtMSK := startsAt.In(msk)

	if startsAtMSK.Weekday() != time.Monday {
		t.Errorf("startsAt in MSK should be Monday, got %s", startsAtMSK.Weekday())
	}
}

// --- getNextSeasonDates: all dates are in the future ---

func TestGetNextSeasonDates_AllInFuture(t *testing.T) {
	now := time.Now()
	startsAt, revealAt, endsAt := getNextSeasonDates()

	if !startsAt.After(now) {
		t.Errorf("startsAt %v should be after now %v", startsAt, now)
	}
	if !revealAt.After(now) {
		t.Errorf("revealAt %v should be after now %v", revealAt, now)
	}
	if !endsAt.After(now) {
		t.Errorf("endsAt %v should be after now %v", endsAt, now)
	}
}

// --- getNextSeasonDates: startsAt is at 00:00 MSK (21:00 UTC previous day) ---

func TestGetNextSeasonDates_StartsAtMidnightMSK(t *testing.T) {
	startsAt, _, _ := getNextSeasonDates()

	msk := time.FixedZone("MSK", 3*60*60)
	startsAtMSK := startsAt.In(msk)

	if startsAtMSK.Hour() != 0 || startsAtMSK.Minute() != 0 || startsAtMSK.Second() != 0 {
		t.Errorf("startsAt MSK should be 00:00:00, got %02d:%02d:%02d",
			startsAtMSK.Hour(), startsAtMSK.Minute(), startsAtMSK.Second())
	}
}

// --- getNextSeasonDates: endsAt hour/minute is 23:59 MSK ---

func TestGetNextSeasonDates_EndsAt2359MSK(t *testing.T) {
	_, _, endsAt := getNextSeasonDates()

	msk := time.FixedZone("MSK", 3*60*60)
	endsAtMSK := endsAt.In(msk)

	if endsAtMSK.Hour() != 23 || endsAtMSK.Minute() != 59 {
		t.Errorf("endsAt MSK should be 23:59, got %02d:%02d",
			endsAtMSK.Hour(), endsAtMSK.Minute())
	}
}

// --- getNextSeasonDates: revealAt is exactly 4 days after startsAt (Monday→Friday) ---

func TestGetNextSeasonDates_RevealIs4DaysAfterStart(t *testing.T) {
	startsAt, revealAt, _ := getNextSeasonDates()

	msk := time.FixedZone("MSK", 3*60*60)
	startMSK := startsAt.In(msk)
	revealMSK := revealAt.In(msk)

	// Compare calendar days: Monday(1) to Friday(5) = 4 days
	startYD := startMSK.YearDay()
	revealYD := revealMSK.YearDay()

	diff := revealYD - startYD
	// Handle year boundary
	if diff < 0 {
		diff += 365
		if startMSK.Year()%4 == 0 {
			diff++
		}
	}
	if diff != 4 {
		t.Errorf("revealAt should be 4 calendar days after startsAt, got %d (start=%s, reveal=%s)",
			diff, startMSK.Format("Mon 2006-01-02"), revealMSK.Format("Mon 2006-01-02"))
	}
}

// --- validateCategories: all valid categories at once ---

func TestValidateCategories_AllValidAtOnce(t *testing.T) {
	all := []string{"HOT", "FUNNY", "SECRETS", "SKILLS", "ROMANCE", "STUDY"}
	birthYear := sql.NullInt32{Int32: 2000, Valid: true}

	cats, err := validateCategories(all, birthYear)
	if err != nil {
		t.Fatalf("expected no error with all valid categories, got %v", err)
	}
	if len(cats) != len(all) {
		t.Errorf("expected %d categories, got %d", len(all), len(cats))
	}
}

// --- validateCategories: duplicate categories (should work, no dedup) ---

func TestValidateCategories_Duplicates(t *testing.T) {
	cats, err := validateCategories([]string{"HOT", "HOT", "FUNNY"}, sql.NullInt32{})
	if err != nil {
		t.Fatalf("expected no error with duplicate categories, got %v", err)
	}
	if len(cats) != 3 {
		t.Errorf("expected 3 categories (including duplicates), got %d", len(cats))
	}
	if cats[0] != db.QuestionCategoryHOT || cats[1] != db.QuestionCategoryHOT {
		t.Errorf("expected first two to be HOT, got %v", cats)
	}
}

// --- validateCategories: ROMANCE for exactly 18 years old (boundary) ---

func TestValidateCategories_RomanceExactly18(t *testing.T) {
	birthYear := sql.NullInt32{Int32: int32(time.Now().Year() - 18), Valid: true}
	cats, err := validateCategories([]string{"ROMANCE"}, birthYear)
	if err != nil {
		t.Fatalf("ROMANCE should be allowed for exactly 18 years old, got %v", err)
	}
	if len(cats) != 1 || cats[0] != db.QuestionCategoryROMANCE {
		t.Errorf("expected [ROMANCE], got %v", cats)
	}
}

// --- validateCategories: ROMANCE for exactly 17 years old (boundary, blocked) ---

func TestValidateCategories_RomanceExactly17(t *testing.T) {
	birthYear := sql.NullInt32{Int32: int32(time.Now().Year() - 17), Valid: true}
	_, err := validateCategories([]string{"ROMANCE"}, birthYear)
	if err != ErrRomanceBlocked {
		t.Errorf("ROMANCE should be blocked for 17 years old, got %v", err)
	}
}

// --- Error sentinel values are distinct ---

func TestErrorSentinelsDistinct(t *testing.T) {
	errs := []error{
		ErrGroupNotFound,
		ErrNotMember,
		ErrAlreadyMember,
		ErrGroupLimitUser,
		ErrGroupLimitSize,
		ErrNotAdmin,
		ErrInvalidName,
		ErrNoCategories,
		ErrInvalidCategory,
		ErrRomanceBlocked,
	}

	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			if errs[i] == errs[j] {
				t.Errorf("error sentinels at index %d and %d should be distinct, both are: %v", i, j, errs[i])
			}
			if errs[i].Error() == errs[j].Error() {
				t.Errorf("error messages at index %d and %d should be distinct, both are: %q", i, j, errs[i].Error())
			}
		}
	}
}

// --- categoryStrings: empty input ---

func TestCategoryStrings_Empty(t *testing.T) {
	strs := categoryStrings(nil)
	if len(strs) != 0 {
		t.Errorf("expected empty slice, got %v", strs)
	}
}

// --- categoryStrings: single element ---

func TestCategoryStrings_Single(t *testing.T) {
	strs := categoryStrings([]db.QuestionCategory{db.QuestionCategoryROMANCE})
	if len(strs) != 1 || strs[0] != "ROMANCE" {
		t.Errorf("expected [ROMANCE], got %v", strs)
	}
}

// --- categoryStrings: all categories ---

func TestCategoryStrings_AllCategories(t *testing.T) {
	all := []db.QuestionCategory{
		db.QuestionCategoryHOT,
		db.QuestionCategoryFUNNY,
		db.QuestionCategorySECRETS,
		db.QuestionCategorySKILLS,
		db.QuestionCategoryROMANCE,
		db.QuestionCategorySTUDY,
	}
	strs := categoryStrings(all)
	if len(strs) != 6 {
		t.Fatalf("expected 6 strings, got %d", len(strs))
	}
	expected := []string{"HOT", "FUNNY", "SECRETS", "SKILLS", "ROMANCE", "STUDY"}
	for i, e := range expected {
		if strs[i] != e {
			t.Errorf("index %d: expected %s, got %s", i, e, strs[i])
		}
	}
}

// --- validateCategories: empty string category ---

func TestValidateCategories_EmptyString(t *testing.T) {
	_, err := validateCategories([]string{""}, sql.NullInt32{})
	if err != ErrInvalidCategory {
		t.Errorf("expected ErrInvalidCategory for empty string, got %v", err)
	}
}

// --- validateCategories: case sensitivity ---

func TestValidateCategories_CaseSensitive(t *testing.T) {
	_, err := validateCategories([]string{"hot"}, sql.NullInt32{})
	if err != ErrInvalidCategory {
		t.Errorf("expected ErrInvalidCategory for lowercase 'hot', got %v", err)
	}

	_, err = validateCategories([]string{"Hot"}, sql.NullInt32{})
	if err != ErrInvalidCategory {
		t.Errorf("expected ErrInvalidCategory for mixed case 'Hot', got %v", err)
	}
}

// --- NewService returns non-nil ---

func TestNewService(t *testing.T) {
	svc := NewService(nil, nil)
	if svc == nil {
		t.Error("NewService should return non-nil service")
	}
}

// --- CreateGroup: invalid category is caught before DB access ---

func TestCreateGroup_InvalidCategoryBeforeDB(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:       "Valid Name",
		Categories: []string{"INVALID_CAT"},
		UserID:     "user-1",
	})
	if err != ErrInvalidCategory {
		t.Errorf("expected ErrInvalidCategory, got %v", err)
	}
}

// --- CreateGroup: ROMANCE blocked before DB access ---

func TestCreateGroup_RomanceBlockedBeforeDB(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "Valid Name",
		Categories:    []string{"ROMANCE"},
		UserID:        "user-1",
		UserBirthYear: sql.NullInt32{Int32: int32(time.Now().Year() - 15), Valid: true},
	})
	if err != ErrRomanceBlocked {
		t.Errorf("expected ErrRomanceBlocked, got %v", err)
	}
}

// ========================================================================
// sqlmock-based tests for service methods that require DB access
// ========================================================================

// helper: create a service backed by go-sqlmock
func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { mockDB.Close() })
	queries := db.New(mockDB)
	svc := NewService(queries, mockDB)
	return svc, mock
}

// group row columns used by GetGroupByID, GetGroupByInviteCode, CreateGroup etc.
var groupCols = []string{
	"id", "name", "invite_code", "admin_id",
	"telegram_chat_id", "telegram_chat_username",
	"telegram_connect_code", "telegram_connect_expiry",
	"created_at", "categories",
}

func mockGroupRow(id, name, inviteCode, adminID string) *sqlmock.Rows {
	return sqlmock.NewRows(groupCols).AddRow(
		id, name, inviteCode, adminID,
		nil, nil, nil, nil,
		time.Now(), `{"HOT"}`,
	)
}

// --- CreateGroup: user at group limit ---

func TestCreateGroup_GroupLimitUser(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))

	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "My Group",
		Categories:    []string{"HOT"},
		UserID:        "user-1",
		UserBirthYear: sql.NullInt32{Int32: 2000, Valid: true},
	})
	if err != ErrGroupLimitUser {
		t.Errorf("expected ErrGroupLimitUser, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- CreateGroup: CountUserGroups DB error ---

func TestCreateGroup_CountUserGroupsDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("db connection lost"))

	_, err := svc.CreateGroup(context.Background(), CreateGroupParams{
		Name:          "My Group",
		Categories:    []string{"HOT"},
		UserID:        "user-1",
		UserBirthYear: sql.NullInt32{Int32: 2000, Valid: true},
	})
	if err == nil || err.Error() != "db connection lost" {
		t.Errorf("expected db error, got %v", err)
	}
}

// --- GetGroup: not a member ---

func TestGetGroup_NotMember(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	_, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- GetGroup: group not found ---

func TestGetGroup_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	// IsGroupMember returns 1 (is a member)
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// GetGroupByID returns no rows
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- JoinGroup: group not found ---

func TestJoinGroup_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("bad-code").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.JoinGroup(context.Background(), "user-1", "bad-code")
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- JoinGroup: already a member ---

func TestJoinGroup_AlreadyMember(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err != ErrAlreadyMember {
		t.Errorf("expected ErrAlreadyMember, got %v", err)
	}
}

// --- JoinGroup: group at member limit ---

func TestJoinGroup_GroupLimitSize(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	// IsGroupMember → not a member
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// CountGroupMembers → at limit
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(50)))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err != ErrGroupLimitSize {
		t.Errorf("expected ErrGroupLimitSize, got %v", err)
	}
}

// --- JoinGroup: user at group limit ---

func TestJoinGroup_UserGroupLimit(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// CountGroupMembers → room available
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	// CountUserGroups → at limit
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err != ErrGroupLimitUser {
		t.Errorf("expected ErrGroupLimitUser, got %v", err)
	}
}

// --- JoinGroup: success ---

func TestJoinGroup_Success(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test Group", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// AddGroupMember
	mock.ExpectQuery("INSERT INTO group_members").
		WithArgs(sqlmock.AnyArg(), "user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "group_id", "joined_at"}).
			AddRow("member-1", "user-1", "group-1", time.Now()))

	group, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if group.ID != "group-1" {
		t.Errorf("expected group ID group-1, got %s", group.ID)
	}
	if group.Name != "Test Group" {
		t.Errorf("expected group name 'Test Group', got %s", group.Name)
	}
}

// --- LeaveGroup: not a member ---

func TestLeaveGroup_NotMember(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err != ErrNotMember {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

// --- LeaveGroup: group not found ---

func TestLeaveGroup_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(sql.ErrNoRows)

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- LeaveGroup: last member deletes group ---

func TestLeaveGroup_LastMemberDeletesGroup(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "user-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectExec("DELETE FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// --- UpdateGroup: group not found ---

func TestUpdateGroup_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(sql.ErrNoRows)

	name := "New Name"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "user-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- UpdateGroup: not admin ---

func TestUpdateGroup_NotAdmin(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	name := "New Name"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "not-admin",
		GroupID: "group-1",
		Name:    &name,
	})
	if err != ErrNotAdmin {
		t.Errorf("expected ErrNotAdmin, got %v", err)
	}
}

// --- UpdateGroup: name too short ---

func TestUpdateGroup_NameTooShort(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	name := "Ab"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

// --- UpdateGroup: name too long ---

func TestUpdateGroup_NameTooLong(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	name := string(make([]byte, 41))
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err != ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

// --- RegenerateInviteLink: group not found ---

func TestRegenerateInviteLink_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.RegenerateInviteLink(context.Background(), "user-1", "group-1")
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- RegenerateInviteLink: not admin ---

func TestRegenerateInviteLink_NotAdmin(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	_, err := svc.RegenerateInviteLink(context.Background(), "not-admin", "group-1")
	if err != ErrNotAdmin {
		t.Errorf("expected ErrNotAdmin, got %v", err)
	}
}

// --- RegenerateInviteLink: success ---

func TestRegenerateInviteLink_Success(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectExec("UPDATE groups SET invite_code").
		WithArgs("group-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	url, err := svc.RegenerateInviteLink(context.Background(), "admin-1", "group-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(url) == 0 {
		t.Error("expected non-empty URL")
	}
	if url[:len("https://repa.app/join/")] != "https://repa.app/join/" {
		t.Errorf("expected URL to start with https://repa.app/join/, got %s", url)
	}
}

// --- GetJoinPreview: group not found ---

func TestGetJoinPreview_GroupNotFound(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("bad-code").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetJoinPreview(context.Background(), "bad-code")
	if err != ErrGroupNotFound {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

// --- GetJoinPreview: success ---

func TestGetJoinPreview_Success(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Cool Group", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(7)))

	mock.ExpectQuery("SELECT username FROM users WHERE id").
		WithArgs("admin-1").
		WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("admin_user"))

	preview, err := svc.GetJoinPreview(context.Background(), "invite-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if preview.Name != "Cool Group" {
		t.Errorf("expected name 'Cool Group', got %s", preview.Name)
	}
	if preview.MemberCount != 7 {
		t.Errorf("expected 7 members, got %d", preview.MemberCount)
	}
	if preview.AdminUsername != "admin_user" {
		t.Errorf("expected admin username 'admin_user', got %s", preview.AdminUsername)
	}
}

// --- GetUser: DB error ---

func TestGetUser_DBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := svc.GetUser(context.Background(), "user-1")
	if err == nil || err.Error() != "connection refused" {
		t.Errorf("expected connection refused error, got %v", err)
	}
}

// --- CreateNewSeasons: no groups needing season ---

func TestCreateNewSeasons_NoGroups(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols))

	err := svc.CreateNewSeasons(context.Background())
	if err != nil {
		t.Errorf("expected no error when no groups need seasons, got %v", err)
	}
}

// --- CreateNewSeasons: DB error ---

func TestCreateNewSeasons_DBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnError(fmt.Errorf("query failed"))

	err := svc.CreateNewSeasons(context.Background())
	if err == nil || err.Error() != "query failed" {
		t.Errorf("expected query failed error, got %v", err)
	}
}

// --- GetUser: success ---

func TestGetUser_Success(t *testing.T) {
	svc, mock := newMockService(t)

	userCols := []string{
		"id", "phone", "apple_id", "google_id", "username",
		"avatar_url", "avatar_emoji", "birth_year",
		"created_at", "updated_at", "username_changed_at",
	}
	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(
			"user-1", nil, nil, nil, "testuser",
			nil, nil, int32(2000),
			now, now, nil,
		))

	user, err := svc.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}
}

// --- GetGroup: success with active season ---

func TestGetGroup_Success(t *testing.T) {
	svc, mock := newMockService(t)

	// IsGroupMember
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// GetGroupByID
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test Group", "invite-1", "admin-1"))

	// GetGroupMembers
	memberCols := []string{"id", "username", "avatar_emoji", "avatar_url"}
	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows(memberCols).
			AddRow("user-1", "alice", nil, nil).
			AddRow("user-2", "bob", nil, nil))

	// GetActiveSeasonByGroup
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at"}
	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM seasons WHERE group_id").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"season-1", "group-1", int32(1), "VOTING",
			now, now.Add(4*24*time.Hour), now.Add(6*24*time.Hour), now,
		))

	detail, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detail.Group.ID != "group-1" {
		t.Errorf("expected group ID group-1, got %s", detail.Group.ID)
	}
	if len(detail.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(detail.Members))
	}
	if detail.ActiveSeason == nil {
		t.Error("expected active season to be set")
	} else if detail.ActiveSeason.ID != "season-1" {
		t.Errorf("expected season ID season-1, got %s", detail.ActiveSeason.ID)
	}
}

// --- GetGroup: success without active season ---

func TestGetGroup_NoActiveSeason(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test Group", "invite-1", "admin-1"))

	memberCols := []string{"id", "username", "avatar_emoji", "avatar_url"}
	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows(memberCols))

	// GetActiveSeasonByGroup returns no rows
	mock.ExpectQuery("SELECT .+ FROM seasons WHERE group_id").
		WithArgs("group-1").
		WillReturnError(sql.ErrNoRows)

	detail, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detail.ActiveSeason != nil {
		t.Error("expected no active season")
	}
}

// --- UpdateGroup: success with name update ---

func TestUpdateGroup_SuccessNameUpdate(t *testing.T) {
	svc, mock := newMockService(t)

	// GetGroupByID (first call)
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Old Name", "invite-1", "admin-1"))

	// UpdateGroupName
	mock.ExpectExec("UPDATE groups SET name").
		WithArgs("group-1", "New Name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// GetGroupByID (second call to return updated group)
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "New Name", "invite-1", "admin-1"))

	name := "New Name"
	updated, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", updated.Name)
	}
}

// --- UpdateGroup: success with telegram username update ---

func TestUpdateGroup_SuccessTelegramUpdate(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	// UpdateGroupTelegramUsername
	mock.ExpectExec("UPDATE groups SET telegram_chat_username").
		WithArgs("group-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// GetGroupByID (re-fetch)
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	tg := "my_chat"
	updated, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:          "admin-1",
		GroupID:         "group-1",
		TelegramUsername: &tg,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.ID != "group-1" {
		t.Errorf("expected group ID group-1, got %s", updated.ID)
	}
}

// --- UpdateGroup: no fields to update (nil name and telegram) ---

func TestUpdateGroup_NoFieldsToUpdate(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	// No UpdateGroupName or UpdateGroupTelegramUsername calls

	// GetGroupByID (re-fetch)
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	updated, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.ID != "group-1" {
		t.Errorf("expected group ID group-1, got %s", updated.ID)
	}
}

// --- LeaveGroup: non-admin member leaves (no admin transfer) ---

func TestLeaveGroup_NonAdminLeaves(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-2", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// group admin is admin-1, not user-2
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	// BeginTx
	mock.ExpectBegin()

	// RemoveGroupMember
	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("user-2", "group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// No admin transfer since user-2 is not the admin

	mock.ExpectCommit()

	err := svc.LeaveGroup(context.Background(), "user-2", "group-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- LeaveGroup: admin leaves, transfers to next member ---

func TestLeaveGroup_AdminTransfer(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("admin-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin()

	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("admin-1", "group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// GetNextAdmin
	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1", "admin-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-2"))

	// UpdateGroupAdmin
	mock.ExpectExec("UPDATE groups SET admin_id").
		WithArgs("group-1", "user-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := svc.LeaveGroup(context.Background(), "admin-1", "group-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- GetJoinPreview: CountGroupMembers error ---

func TestGetJoinPreview_CountMembersError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("count error"))

	_, err := svc.GetJoinPreview(context.Background(), "invite-1")
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected count error, got %v", err)
	}
}

// --- RegenerateInviteLink: DB update error ---

func TestRegenerateInviteLink_UpdateError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectExec("UPDATE groups SET invite_code").
		WithArgs("group-1", sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update failed"))

	_, err := svc.RegenerateInviteLink(context.Background(), "admin-1", "group-1")
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected update failed error, got %v", err)
	}
}

// ========================================================================
// Additional coverage tests
// ========================================================================

// --- ListUserGroups ---

var listGroupStatsCols = []string{
	"id", "name", "invite_code", "admin_id",
	"telegram_chat_id", "telegram_chat_username",
	"telegram_connect_code", "telegram_connect_expiry",
	"created_at", "categories",
	"member_count",
	"active_season_id", "active_season_number", "active_season_status",
	"active_season_starts_at", "active_season_reveal_at", "active_season_ends_at",
	"voted_count", "user_vote_count",
}

func TestListUserGroups_Empty(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(listGroupStatsCols))

	items, err := svc.ListUserGroups(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}

func TestListUserGroups_DBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("db down"))

	_, err := svc.ListUserGroups(context.Background(), "user-1")
	if err == nil || err.Error() != "db down" {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestListUserGroups_WithoutActiveSeason(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	mock.ExpectQuery("SELECT .+").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(listGroupStatsCols).AddRow(
			"g1", "Group One", "inv1", "admin1",
			nil, nil, nil, nil,
			now, `{"HOT","FUNNY"}`,
			int64(5),
			nil, nil, nil, nil, nil, nil, // no active season
			int64(0), int64(0),
		))

	items, err := svc.ListUserGroups(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Group.ID != "g1" {
		t.Errorf("expected group ID g1, got %s", items[0].Group.ID)
	}
	if items[0].ActiveSeason != nil {
		t.Error("expected no active season")
	}
	if items[0].MemberCount != 5 {
		t.Errorf("expected member count 5, got %d", items[0].MemberCount)
	}
}

func TestListUserGroups_WithActiveSeason(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	revealAt := now.Add(3 * 24 * time.Hour)
	endsAt := now.Add(5 * 24 * time.Hour)
	mock.ExpectQuery("SELECT .+").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(listGroupStatsCols).AddRow(
			"g1", "Group One", "inv1", "admin1",
			nil, nil, nil, nil,
			now, `{"HOT"}`,
			int64(8),
			"s1", int32(2), "VOTING", now, revealAt, endsAt,
			int64(4), int64(1),
		))

	items, err := svc.ListUserGroups(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.ActiveSeason == nil {
		t.Fatal("expected active season")
	}
	if item.ActiveSeason.ID != "s1" {
		t.Errorf("expected season ID s1, got %s", item.ActiveSeason.ID)
	}
	if item.ActiveSeason.Number != 2 {
		t.Errorf("expected season number 2, got %d", item.ActiveSeason.Number)
	}
	if item.VotedCount != 4 {
		t.Errorf("expected voted count 4, got %d", item.VotedCount)
	}
	if !item.UserVoted {
		t.Error("expected user to have voted")
	}
}

func TestListUserGroups_MultipleGroups(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	mock.ExpectQuery("SELECT .+").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows(listGroupStatsCols).
			AddRow("g1", "Group1", "inv1", "a1", nil, nil, nil, nil, now, `{"HOT"}`,
				int64(3), nil, nil, nil, nil, nil, nil, int64(0), int64(0)).
			AddRow("g2", "Group2", "inv2", "a2", nil, nil, nil, nil, now, `{"FUNNY"}`,
				int64(7), "s2", int32(1), "VOTING", now, now.Add(time.Hour), now.Add(2*time.Hour),
				int64(2), int64(0)))

	items, err := svc.ListUserGroups(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Group.ID != "g1" || items[1].Group.ID != "g2" {
		t.Errorf("unexpected group IDs: %s, %s", items[0].Group.ID, items[1].Group.ID)
	}
	if items[1].ActiveSeason == nil {
		t.Error("expected second group to have active season")
	}
	if items[1].UserVoted {
		t.Error("expected user not to have voted in second group")
	}
}

// --- GetGroup: IsGroupMember DB error ---

func TestGetGroup_IsGroupMemberDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnError(fmt.Errorf("db error"))

	_, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

// --- GetGroup: GetGroupByID generic DB error (not ErrNoRows) ---

func TestGetGroup_GetGroupByIDGenericError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("connection timeout"))

	_, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err == nil || err.Error() != "connection timeout" {
		t.Errorf("expected connection timeout error, got %v", err)
	}
}

// --- GetGroup: GetGroupMembers DB error ---

func TestGetGroup_GetGroupMembersError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("members query failed"))

	_, err := svc.GetGroup(context.Background(), "group-1", "user-1")
	if err == nil || err.Error() != "members query failed" {
		t.Errorf("expected members query failed error, got %v", err)
	}
}

// --- JoinGroup: GetGroupByInviteCode generic DB error ---

func TestJoinGroup_InviteCodeDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnError(fmt.Errorf("db timeout"))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err == nil || err.Error() != "db timeout" {
		t.Errorf("expected db timeout error, got %v", err)
	}
}

// --- JoinGroup: IsGroupMember DB error ---

func TestJoinGroup_IsGroupMemberDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnError(fmt.Errorf("member check failed"))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err == nil || err.Error() != "member check failed" {
		t.Errorf("expected member check failed error, got %v", err)
	}
}

// --- JoinGroup: CountGroupMembers DB error ---

func TestJoinGroup_CountGroupMembersDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("count members error"))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err == nil || err.Error() != "count members error" {
		t.Errorf("expected count members error, got %v", err)
	}
}

// --- JoinGroup: CountUserGroups DB error ---

func TestJoinGroup_CountUserGroupsDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("count user groups error"))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err == nil || err.Error() != "count user groups error" {
		t.Errorf("expected count user groups error, got %v", err)
	}
}

// --- JoinGroup: AddGroupMember DB error ---

func TestJoinGroup_AddGroupMemberDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	mock.ExpectQuery("INSERT INTO group_members").
		WithArgs(sqlmock.AnyArg(), "user-1", "group-1").
		WillReturnError(fmt.Errorf("insert failed"))

	_, err := svc.JoinGroup(context.Background(), "user-1", "invite-1")
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("expected insert failed error, got %v", err)
	}
}

// --- LeaveGroup: IsGroupMember DB error ---

func TestLeaveGroup_IsGroupMemberDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnError(fmt.Errorf("db error"))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

// --- LeaveGroup: GetGroupByID generic DB error ---

func TestLeaveGroup_GetGroupByIDGenericError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("connection error"))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "connection error" {
		t.Errorf("expected connection error, got %v", err)
	}
}

// --- LeaveGroup: CountGroupMembers DB error ---

func TestLeaveGroup_CountGroupMembersDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("count error"))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected count error, got %v", err)
	}
}

// --- LeaveGroup: DeleteGroup error (last member) ---

func TestLeaveGroup_DeleteGroupError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "user-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectExec("DELETE FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("delete failed"))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "delete failed" {
		t.Errorf("expected delete failed error, got %v", err)
	}
}

// --- LeaveGroup: BeginTx error ---

func TestLeaveGroup_BeginTxError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin tx error"))

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "begin tx error" {
		t.Errorf("expected begin tx error, got %v", err)
	}
}

// --- LeaveGroup: RemoveGroupMember error ---

func TestLeaveGroup_RemoveGroupMemberError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin()

	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("user-1", "group-1").
		WillReturnError(fmt.Errorf("remove member error"))

	mock.ExpectRollback()

	err := svc.LeaveGroup(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "remove member error" {
		t.Errorf("expected remove member error, got %v", err)
	}
}

// --- LeaveGroup: GetNextAdmin error ---

func TestLeaveGroup_GetNextAdminError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("admin-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin()

	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("admin-1", "group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1", "admin-1").
		WillReturnError(fmt.Errorf("no next admin"))

	mock.ExpectRollback()

	err := svc.LeaveGroup(context.Background(), "admin-1", "group-1")
	if err == nil || err.Error() != "no next admin" {
		t.Errorf("expected no next admin error, got %v", err)
	}
}

// --- LeaveGroup: UpdateGroupAdmin error ---

func TestLeaveGroup_UpdateGroupAdminError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("admin-1", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin()

	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("admin-1", "group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT .+ FROM users .+ JOIN group_members").
		WithArgs("group-1", "admin-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-2"))

	mock.ExpectExec("UPDATE groups SET admin_id").
		WithArgs("group-1", "user-2").
		WillReturnError(fmt.Errorf("update admin error"))

	mock.ExpectRollback()

	err := svc.LeaveGroup(context.Background(), "admin-1", "group-1")
	if err == nil || err.Error() != "update admin error" {
		t.Errorf("expected update admin error, got %v", err)
	}
}

// --- LeaveGroup: commit error ---

func TestLeaveGroup_CommitError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("user-2", "group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mock.ExpectBegin()

	mock.ExpectExec("DELETE FROM group_members WHERE user_id").
		WithArgs("user-2", "group-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err := svc.LeaveGroup(context.Background(), "user-2", "group-1")
	if err == nil || err.Error() != "commit error" {
		t.Errorf("expected commit error, got %v", err)
	}
}

// --- CreateNewSeasons: single group success ---

func TestCreateNewSeasons_SingleGroupSuccess(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	// GetGroupsNeedingNewSeason returns one group
	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols).AddRow(
			"g1", "Group One", "inv1", "admin1",
			nil, nil, nil, nil,
			now, `{"HOT"}`,
		))

	// createSeasonForGroup transaction:
	mock.ExpectBegin()

	// GetRevealedSeasonsForGroup - no revealed seasons
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at"}
	mock.ExpectQuery("SELECT .+ FROM seasons.+WHERE group_id.+REVEALED").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(seasonCols))

	// GetLastSeasonNumber
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(1)))

	// CreateSeason
	mock.ExpectQuery("INSERT INTO seasons").
		WithArgs(sqlmock.AnyArg(), "g1", int32(2), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"new-season", "g1", int32(2), "VOTING", now, now, now, now,
		))

	// CountGroupMembers
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	// GetRandomSystemQuestionsByCategories - returns 5 questions
	questionCols := []string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}
	qRows := sqlmock.NewRows(questionCols)
	for i := 0; i < 5; i++ {
		qRows.AddRow(fmt.Sprintf("q%d", i), fmt.Sprintf("Question %d", i), "HOT", "SYSTEM", nil, nil, "ACTIVE", now)
	}
	mock.ExpectQuery("SELECT .+ FROM questions").
		WithArgs(sqlmock.AnyArg(), "g1", int32(10)). // 5 members * 2 = 10
		WillReturnRows(qRows)

	// AddSeasonQuestion for each question
	sqCols := []string{"id", "season_id", "question_id", "ord"}
	for i := 0; i < 5; i++ {
		mock.ExpectQuery("INSERT INTO season_questions").
			WithArgs(sqlmock.AnyArg(), "new-season", fmt.Sprintf("q%d", i), int32(i)).
			WillReturnRows(sqlmock.NewRows(sqCols).AddRow(
				fmt.Sprintf("sq%d", i), "new-season", fmt.Sprintf("q%d", i), int32(i),
			))
	}

	mock.ExpectCommit()

	err := svc.CreateNewSeasons(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- CreateNewSeasons: all groups fail ---

func TestCreateNewSeasons_AllGroupsFail(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols).AddRow(
			"g1", "Group One", "inv1", "admin1",
			nil, nil, nil, nil,
			now, `{"HOT"}`,
		))

	// createSeasonForGroup: BeginTx fails
	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin tx failed"))

	err := svc.CreateNewSeasons(context.Background())
	if err == nil {
		t.Fatal("expected error when all groups fail")
	}
	if err.Error() != "season creator: all 1 groups failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- CreateNewSeasons: partial failure (some succeed, some fail) ---

func TestCreateNewSeasons_PartialFailure(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at"}

	// Two groups
	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols).
			AddRow("g1", "Group1", "inv1", "a1", nil, nil, nil, nil, now, `{"HOT"}`).
			AddRow("g2", "Group2", "inv2", "a2", nil, nil, nil, nil, now, `{"FUNNY"}`))

	// First group: BeginTx fails
	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))

	// Second group: succeeds
	mock.ExpectBegin()

	mock.ExpectQuery("SELECT .+ FROM seasons.+WHERE group_id.+REVEALED").
		WithArgs("g2").
		WillReturnRows(sqlmock.NewRows(seasonCols))

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("g2").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(0)))

	mock.ExpectQuery("INSERT INTO seasons").
		WithArgs(sqlmock.AnyArg(), "g2", int32(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"ns2", "g2", int32(1), "VOTING", now, now, now, now,
		))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g2").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	questionCols := []string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}
	qRows := sqlmock.NewRows(questionCols)
	for i := 0; i < 5; i++ {
		qRows.AddRow(fmt.Sprintf("q%d", i), fmt.Sprintf("Q %d", i), "FUNNY", "SYSTEM", nil, nil, "ACTIVE", now)
	}
	mock.ExpectQuery("SELECT .+ FROM questions").
		WithArgs(sqlmock.AnyArg(), "g2", int32(6)). // 3 * 2 = 6
		WillReturnRows(qRows)

	sqCols := []string{"id", "season_id", "question_id", "ord"}
	for i := 0; i < 5; i++ {
		mock.ExpectQuery("INSERT INTO season_questions").
			WithArgs(sqlmock.AnyArg(), "ns2", fmt.Sprintf("q%d", i), int32(i)).
			WillReturnRows(sqlmock.NewRows(sqCols).AddRow(
				fmt.Sprintf("sq%d", i), "ns2", fmt.Sprintf("q%d", i), int32(i),
			))
	}

	mock.ExpectCommit()

	// Should succeed (partial failure is logged, not returned as error)
	err := svc.CreateNewSeasons(context.Background())
	if err != nil {
		t.Fatalf("expected no error with partial failure, got %v", err)
	}
}

// --- createSeasonForGroup: with revealed seasons to close ---

func TestCreateNewSeasons_ClosesRevealedSeasons(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at"}

	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols).AddRow(
			"g1", "Group1", "inv1", "a1", nil, nil, nil, nil, now, `{"HOT"}`))

	mock.ExpectBegin()

	// GetRevealedSeasonsForGroup returns one revealed season
	mock.ExpectQuery("SELECT .+ FROM seasons.+WHERE group_id.+REVEALED").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"old-s1", "g1", int32(1), "REVEALED", now, now, now, now,
		))

	// UpdateSeasonStatus to close the revealed season
	mock.ExpectExec("UPDATE seasons SET status").
		WithArgs("old-s1", "CLOSED").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// GetLastSeasonNumber
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(1)))

	// CreateSeason
	mock.ExpectQuery("INSERT INTO seasons").
		WithArgs(sqlmock.AnyArg(), "g1", int32(2), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"new-s1", "g1", int32(2), "VOTING", now, now, now, now,
		))

	// CountGroupMembers
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(4)))

	// GetRandomSystemQuestionsByCategories
	questionCols := []string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}
	qRows := sqlmock.NewRows(questionCols)
	for i := 0; i < 5; i++ {
		qRows.AddRow(fmt.Sprintf("q%d", i), fmt.Sprintf("Q%d", i), "HOT", "SYSTEM", nil, nil, "ACTIVE", now)
	}
	mock.ExpectQuery("SELECT .+ FROM questions").
		WithArgs(sqlmock.AnyArg(), "g1", int32(8)). // 4 * 2 = 8
		WillReturnRows(qRows)

	sqCols := []string{"id", "season_id", "question_id", "ord"}
	for i := 0; i < 5; i++ {
		mock.ExpectQuery("INSERT INTO season_questions").
			WithArgs(sqlmock.AnyArg(), "new-s1", fmt.Sprintf("q%d", i), int32(i)).
			WillReturnRows(sqlmock.NewRows(sqCols).AddRow(
				fmt.Sprintf("sq%d", i), "new-s1", fmt.Sprintf("q%d", i), int32(i),
			))
	}

	mock.ExpectCommit()

	err := svc.CreateNewSeasons(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// --- UpdateGroup: UpdateGroupName DB error ---

func TestUpdateGroup_UpdateGroupNameDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Old Name", "invite-1", "admin-1"))

	mock.ExpectExec("UPDATE groups SET name").
		WithArgs("group-1", "New Name").
		WillReturnError(fmt.Errorf("update name failed"))

	name := "New Name"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err == nil || err.Error() != "update name failed" {
		t.Errorf("expected update name failed error, got %v", err)
	}
}

// --- UpdateGroup: UpdateGroupTelegramUsername DB error ---

func TestUpdateGroup_UpdateTelegramUsernameDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectExec("UPDATE groups SET telegram_chat_username").
		WithArgs("group-1", sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update tg failed"))

	tg := "my_chat"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:          "admin-1",
		GroupID:         "group-1",
		TelegramUsername: &tg,
	})
	if err == nil || err.Error() != "update tg failed" {
		t.Errorf("expected update tg failed error, got %v", err)
	}
}

// --- UpdateGroup: re-fetch after update fails ---

func TestUpdateGroup_RefetchError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	// No name or tg update, goes straight to re-fetch
	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("refetch error"))

	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "admin-1",
		GroupID: "group-1",
	})
	if err == nil || err.Error() != "refetch error" {
		t.Errorf("expected refetch error, got %v", err)
	}
}

// --- UpdateGroup: generic GetGroupByID error (not ErrNoRows) ---

func TestUpdateGroup_GenericDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("connection lost"))

	name := "New Name"
	_, err := svc.UpdateGroup(context.Background(), UpdateGroupParams{
		UserID:  "user-1",
		GroupID: "group-1",
		Name:    &name,
	})
	if err == nil || err.Error() != "connection lost" {
		t.Errorf("expected connection lost error, got %v", err)
	}
}

// --- GetJoinPreview: generic DB error from GetGroupByInviteCode ---

func TestGetJoinPreview_GenericDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("code-1").
		WillReturnError(fmt.Errorf("db broken"))

	_, err := svc.GetJoinPreview(context.Background(), "code-1")
	if err == nil || err.Error() != "db broken" {
		t.Errorf("expected db broken error, got %v", err)
	}
}

// --- GetJoinPreview: GetAdminUsername error ---

func TestGetJoinPreview_GetAdminUsernameError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE invite_code").
		WithArgs("invite-1").
		WillReturnRows(mockGroupRow("group-1", "Test", "invite-1", "admin-1"))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("group-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mock.ExpectQuery("SELECT username FROM users WHERE id").
		WithArgs("admin-1").
		WillReturnError(fmt.Errorf("admin not found"))

	_, err := svc.GetJoinPreview(context.Background(), "invite-1")
	if err == nil || err.Error() != "admin not found" {
		t.Errorf("expected admin not found error, got %v", err)
	}
}

// --- RegenerateInviteLink: generic DB error from GetGroupByID ---

func TestRegenerateInviteLink_GenericDBError(t *testing.T) {
	svc, mock := newMockService(t)

	mock.ExpectQuery("SELECT .+ FROM groups WHERE id").
		WithArgs("group-1").
		WillReturnError(fmt.Errorf("db timeout"))

	_, err := svc.RegenerateInviteLink(context.Background(), "user-1", "group-1")
	if err == nil || err.Error() != "db timeout" {
		t.Errorf("expected db timeout error, got %v", err)
	}
}

// --- selectAndAssignQuestionsTx: question count boundaries ---

func TestCreateNewSeasons_QuestionCountCappedAtMax(t *testing.T) {
	svc, mock := newMockService(t)

	now := time.Now()
	seasonCols := []string{"id", "group_id", "number", "status", "starts_at", "reveal_at", "ends_at", "created_at"}

	// Group with many members (memberCount * 2 > MaxSeasonQuestions)
	mock.ExpectQuery("SELECT .+ FROM groups").
		WillReturnRows(sqlmock.NewRows(groupCols).AddRow(
			"g1", "Group1", "inv1", "a1", nil, nil, nil, nil, now, `{"HOT"}`))

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT .+ FROM seasons.+WHERE group_id.+REVEALED").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows(seasonCols))

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(int32(0)))

	mock.ExpectQuery("INSERT INTO seasons").
		WithArgs(sqlmock.AnyArg(), "g1", int32(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(seasonCols).AddRow(
			"ns1", "g1", int32(1), "VOTING", now, now, now, now,
		))

	// 20 members -> 20 * 2 = 40, capped at MaxSeasonQuestions = 10
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("g1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(20)))

	questionCols := []string{"id", "text", "category", "source", "group_id", "author_id", "status", "created_at"}
	qRows := sqlmock.NewRows(questionCols)
	for i := 0; i < 10; i++ {
		qRows.AddRow(fmt.Sprintf("q%d", i), fmt.Sprintf("Q%d", i), "HOT", "SYSTEM", nil, nil, "ACTIVE", now)
	}
	mock.ExpectQuery("SELECT .+ FROM questions").
		WithArgs(sqlmock.AnyArg(), "g1", int32(10)).
		WillReturnRows(qRows)

	sqCols := []string{"id", "season_id", "question_id", "ord"}
	for i := 0; i < 10; i++ {
		mock.ExpectQuery("INSERT INTO season_questions").
			WithArgs(sqlmock.AnyArg(), "ns1", fmt.Sprintf("q%d", i), int32(i)).
			WillReturnRows(sqlmock.NewRows(sqCols).AddRow(
				fmt.Sprintf("sq%d", i), "ns1", fmt.Sprintf("q%d", i), int32(i),
			))
	}

	mock.ExpectCommit()

	err := svc.CreateNewSeasons(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
