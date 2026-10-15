package cards

import (
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func result(text string, pct float64, tone db.QuestionTone, cat db.QuestionCategory) db.GetSeasonResultsByUserRow {
	return db.GetSeasonResultsByUserRow{
		QuestionID:       text,
		QuestionText:     text,
		QuestionCategory: cat,
		QuestionTone:     tone,
		Percentage:       pct,
	}
}

// The rendered PNG is the shared artifact the tone rule exists for, and it is produced furthest from
// the API — so it gets its own test rather than relying on the reveal service's.
func TestCardTopAttributes_LeadsWithTheWarmerNearTie(t *testing.T) {
	attrs, _ := cardTopAttributes([]db.GetSeasonResultsByUserRow{
		result("колкий", 40, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("тёплый", 38, db.QuestionToneWARM, db.QuestionCategorySKILLS),
	})

	if len(attrs) != 2 {
		t.Fatalf("got %d attributes, want 2", len(attrs))
	}
	if attrs[0].QuestionText != "тёплый" {
		t.Errorf("the shared card should lead with the warm attribute, got %q", attrs[0].QuestionText)
	}
}

func TestCardTopAttributes_KeepsAClearlyStrongerResultFirst(t *testing.T) {
	attrs, _ := cardTopAttributes([]db.GetSeasonResultsByUserRow{
		result("колкий", 46, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("тёплый", 38, db.QuestionToneWARM, db.QuestionCategorySKILLS),
	})
	// 8 points apart: a real difference, and the card must stay truthful even when sharing it stings.
	if attrs[0].QuestionText != "колкий" {
		t.Errorf("expected the stronger result to lead, got %q", attrs[0].QuestionText)
	}
}

func TestCardTopAttributes_TitleFollowsTheLeadingAttribute(t *testing.T) {
	_, title := cardTopAttributes([]db.GetSeasonResultsByUserRow{
		result("колкий", 40, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("тёплый", 38, db.QuestionToneWARM, db.QuestionCategorySKILLS),
	})
	// The SKILLS attribute leads after ordering, so the title must be the SKILLS one — not HOT's.
	if want := titleForCategory("SKILLS"); title != want {
		t.Errorf("title = %q, want %q (the leading attribute's category)", title, want)
	}
}

func TestCardTopAttributes_TakesAtMostThree(t *testing.T) {
	attrs, _ := cardTopAttributes([]db.GetSeasonResultsByUserRow{
		result("a", 50, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("b", 30, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("c", 20, db.QuestionToneEDGY, db.QuestionCategoryHOT),
		result("d", 10, db.QuestionToneEDGY, db.QuestionCategoryHOT),
	})
	if len(attrs) != 3 {
		t.Errorf("got %d attributes, want 3", len(attrs))
	}
}

func TestCardTopAttributes_NoResultsStillYieldsATitle(t *testing.T) {
	attrs, title := cardTopAttributes(nil)
	if len(attrs) != 0 {
		t.Errorf("got %d attributes, want 0", len(attrs))
	}
	if title == "" {
		t.Error("a member nobody voted for still needs a title on their card")
	}
}
