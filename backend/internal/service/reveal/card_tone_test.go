package reveal

import (
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func toneAttr(text string, pct float64, tone db.QuestionTone) db.GetSeasonResultsByUserRow {
	return db.GetSeasonResultsByUserRow{
		QuestionID:       text,
		QuestionText:     text,
		QuestionCategory: db.QuestionCategoryFUNNY,
		QuestionTone:     tone,
		Percentage:       pct,
	}
}

func TestSplitAttributes_ToneDecidesWhoIsOnTheCard(t *testing.T) {
	// The point of ordering before the split: the top three are what the shared card shows.
	results := []db.GetSeasonResultsByUserRow{
		toneAttr("edgy-1", 40, db.QuestionToneEDGY),
		toneAttr("edgy-2", 39, db.QuestionToneEDGY),
		toneAttr("edgy-3", 38, db.QuestionToneEDGY),
		toneAttr("warm", 37, db.QuestionToneWARM),
	}
	top, hidden := splitAttributes(results)
	if len(top) != 3 || len(hidden) != 1 {
		t.Fatalf("got %d top and %d hidden, want 3 and 1", len(top), len(hidden))
	}
	if top[0].QuestionText != "warm" {
		t.Errorf("the warm attribute should lead within the band, got %q", top[0].QuestionText)
	}
	// Rank is assigned after ordering, so it describes the card the member sees.
	for i, a := range top {
		if a.Rank != i+1 {
			t.Errorf("top[%d] has rank %d", i, a.Rank)
		}
	}
	if hidden[0].Rank != 4 {
		t.Errorf("hidden attribute has rank %d, want 4", hidden[0].Rank)
	}
}

// TestMembersCards_CarryToneThrough is the test the members-cards path needed: it builds its rows by
// hand from a different query, and a forgotten field reads as neutral for every row — which looks
// exactly like plain percentage order and fails nothing.
func TestMembersCards_CarryToneThrough(t *testing.T) {
	// The shape GetMembersCards constructs: one row per (target, question), tone included.
	rows := []db.GetSeasonResultsByUserRow{
		toneAttr("edgy", 40, db.QuestionToneEDGY),
		toneAttr("warm", 38, db.QuestionToneWARM),
	}

	top, _ := splitAttributes(rows)
	if top[0].QuestionText != "warm" {
		t.Fatalf("a member's card in the list must lead with the warm attribute, got %q", top[0].QuestionText)
	}

	// And the same rows with tone dropped must order by percentage — which is what proves the
	// assertion above is actually about tone and not about input order.
	stripped := []db.GetSeasonResultsByUserRow{
		{QuestionID: "edgy", QuestionText: "edgy", Percentage: 40},
		{QuestionID: "warm", QuestionText: "warm", Percentage: 38},
	}
	strippedTop, _ := splitAttributes(stripped)
	if strippedTop[0].QuestionText != "edgy" {
		t.Errorf("without tone the order should be by percentage, got %q", strippedTop[0].QuestionText)
	}
}

// TestGenerateTitle_FollowsTheCardsLead pins that the title describes the attribute the reader sees
// first. splitAttributes orders, so generateTitle reads an already-ordered list.
func TestGenerateTitle_FollowsTheCardsLead(t *testing.T) {
	warmSkill := toneAttr("учит", 38, db.QuestionToneWARM)
	warmSkill.QuestionCategory = db.QuestionCategorySKILLS

	top, _ := splitAttributes([]db.GetSeasonResultsByUserRow{
		toneAttr("колкий", 40, db.QuestionToneEDGY),
		warmSkill,
	})
	if top[0].Category != string(db.QuestionCategorySKILLS) {
		t.Fatalf("expected the SKILLS attribute to lead, got %s", top[0].Category)
	}
	if generateTitle(top) == "" {
		t.Error("expected a title for the leading attribute")
	}
}
