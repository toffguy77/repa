package questions

import (
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func TestClassifyTone_PhraseMarkersWinOverCategory(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		category db.QuestionCategory
		want     db.QuestionTone
	}{
		// A category is a coarse signal; the phrase decides.
		{"warm phrase in an edgy category", "Кто первым поздравит с днём рождения?", db.QuestionCategoryHOT, db.QuestionToneWARM},
		{"edgy phrase in a warm category", "Кто хуже всех водит машину?", db.QuestionCategorySKILLS, db.QuestionToneEDGY},
		{"edgy phrase in a neutral category", "Кто тайно читает чужие переписки?", db.QuestionCategoryFUNNY, db.QuestionToneEDGY},
		{"warm phrase in a neutral category", "Кто всегда поможет с домашкой?", db.QuestionCategorySTUDY, db.QuestionToneWARM},

		// Category fallback when nothing matched.
		{"skills fallback is warm", "Кто быстрее всех бегает?", db.QuestionCategorySKILLS, db.QuestionToneWARM},
		{"hot fallback is edgy", "Кто заведёт роман на работе?", db.QuestionCategoryHOT, db.QuestionToneEDGY},
		{"secrets fallback is edgy", "У кого самый неожиданный секрет?", db.QuestionCategorySECRETS, db.QuestionToneEDGY},
		{"funny fallback is neutral", "Кто громче всех смеётся в кино?", db.QuestionCategoryFUNNY, db.QuestionToneNEUTRAL},
		{"study fallback is neutral", "Кто сдаёт работу последним?", db.QuestionCategorySTUDY, db.QuestionToneNEUTRAL},
		{"romance fallback is neutral", "Кто первым напишет после свидания?", db.QuestionCategoryROMANCE, db.QuestionToneNEUTRAL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyTone(tt.text, tt.category); got != tt.want {
				t.Errorf("ClassifyTone(%q, %s) = %s, want %s", tt.text, tt.category, got, tt.want)
			}
		})
	}
}

func TestClassifyTone_EdgyWinsWhenBothMarkersPresent(t *testing.T) {
	// A kind-only group should not receive a question that is partly a jab.
	got := ClassifyTone("Кто поможет, а потом сольёт всем?", db.QuestionCategoryFUNNY)
	if got != db.QuestionToneEDGY {
		t.Errorf("got %s, want EDGY — the edgy half is the half that matters", got)
	}
}

func TestClassifyTone_IsCaseInsensitive(t *testing.T) {
	if ClassifyTone("КТО ТАЙНО ЧИТАЕТ ЧУЖИЕ ПЕРЕПИСКИ?", db.QuestionCategoryFUNNY) != db.QuestionToneEDGY {
		t.Error("classification must not depend on case")
	}
}

func TestClassifyTone_NeverReturnsAnInvalidTone(t *testing.T) {
	valid := map[db.QuestionTone]bool{
		db.QuestionToneWARM:    true,
		db.QuestionToneNEUTRAL: true,
		db.QuestionToneEDGY:    true,
	}
	for _, text := range []string{"", "???", "Кто?"} {
		for _, c := range []db.QuestionCategory{
			db.QuestionCategoryHOT, db.QuestionCategoryFUNNY, db.QuestionCategorySECRETS,
			db.QuestionCategorySKILLS, db.QuestionCategoryROMANCE, db.QuestionCategorySTUDY,
			db.QuestionCategory("UNKNOWN"),
		} {
			if got := ClassifyTone(text, c); !valid[got] {
				t.Errorf("ClassifyTone(%q, %s) returned %q, which is not a valid tone", text, c, got)
			}
		}
	}
}

func TestIsKindTone(t *testing.T) {
	if !IsKindTone(db.QuestionToneWARM) || !IsKindTone(db.QuestionToneNEUTRAL) {
		t.Error("warm and neutral are both kind")
	}
	if IsKindTone(db.QuestionToneEDGY) {
		t.Error("edgy is the one a kind-only group must not receive")
	}
}

func TestEdgyMarkers_AreTheCompleteSet(t *testing.T) {
	// The property that matters is that edgyMarkers is *complete*, not precise: a warm question
	// labelled neutral costs nothing, while an edgy question reaching a kind-only group is the failure
	// this exists to prevent. These are the jabs in the shipped bank.
	mustBeEdgy := []string{
		"Кто первым побежит при пожаре?",
		"Кто наябедничает учителю первым?",
		"Кто скорее всего съест чужую еду из холодильника?",
		"Кто притворяется, что всё знает, но на самом деле нет?",
		"Кто тайно читает чужие переписки?",
	}
	for _, text := range mustBeEdgy {
		if got := ClassifyTone(text, db.QuestionCategoryHOT); got != db.QuestionToneEDGY {
			t.Errorf("%q classified as %s, want EDGY", text, got)
		}
	}
}
