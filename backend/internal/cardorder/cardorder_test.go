package cardorder

import (
	"fmt"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

func attr(text string, pct float64, tone db.QuestionTone) db.GetSeasonResultsByUserRow {
	return db.GetSeasonResultsByUserRow{
		QuestionID:   text,
		QuestionText: text,
		Percentage:   pct,
		QuestionTone: tone,
	}
}

func order(rows []db.GetSeasonResultsByUserRow) []string {
	out := make([]string, 0, len(rows))
	for _, r := range ByTone(rows) {
		out = append(out, r.QuestionText)
	}
	return out
}

func assertOrder(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestByTone_ExactTiePrefersWarm(t *testing.T) {
	assertOrder(t, order([]db.GetSeasonResultsByUserRow{
		attr("edgy", 40, db.QuestionToneEDGY),
		attr("neutral", 40, db.QuestionToneNEUTRAL),
		attr("warm", 40, db.QuestionToneWARM),
	}), []string{"warm", "neutral", "edgy"})
}

func TestByTone_WithinBandPrefersWarm(t *testing.T) {
	// 4 points apart: inside the band, so the ordering was arbitrary anyway.
	assertOrder(t, order([]db.GetSeasonResultsByUserRow{
		attr("edgy-44", 44, db.QuestionToneEDGY),
		attr("warm-40", 40, db.QuestionToneWARM),
	}), []string{"warm-40", "edgy-44"})
}

func TestByTone_OutOfBandKeepsTheStrongerResult(t *testing.T) {
	// 6 points apart: a real difference, and the card must stay truthful.
	assertOrder(t, order([]db.GetSeasonResultsByUserRow{
		attr("edgy-46", 46, db.QuestionToneEDGY),
		attr("warm-40", 40, db.QuestionToneWARM),
	}), []string{"edgy-46", "warm-40"})
}

func TestByTone_ExactlyOnTheBandEdgeIsInside(t *testing.T) {
	// 5.0 points is the band, inclusive — one boundary, stated once, rather than a gap at exactly 5.
	assertOrder(t, order([]db.GetSeasonResultsByUserRow{
		attr("edgy-45", 45, db.QuestionToneEDGY),
		attr("warm-40", 40, db.QuestionToneWARM),
	}), []string{"warm-40", "edgy-45"})
}

// TestByTone_ChainOfNearTiesCannotWalkPastTheAnchor is the case that discriminates
// anchor clusters from a comparison function. Each neighbouring pair is within the band, but the ends
// are 6 points apart. A comparator would report warm < neutral < edgy < warm and sort into an
// undefined order that can place the 44% warm attribute first.
func TestByTone_ChainOfNearTiesCannotWalkPastTheAnchor(t *testing.T) {
	got := order([]db.GetSeasonResultsByUserRow{
		attr("edgy-50", 50, db.QuestionToneEDGY),
		attr("neutral-47", 47, db.QuestionToneNEUTRAL),
		attr("warm-44", 44, db.QuestionToneWARM),
	})
	// The cluster anchors at 50 and holds 47 (3 points behind), but not 44 — which is 6 points behind
	// the anchor even though it is only 3 behind 47. So tone promotes neutral-47 over edgy-50, and
	// warm-44 stays below both. A comparator would instead be free to put warm-44 first.
	assertOrder(t, got, []string{"neutral-47", "edgy-50", "warm-44"})
}

func TestByTone_LongChainNeverPromotesPastTheBand(t *testing.T) {
	// A 2-point ladder from 50 down to 38. Clusters anchor at 50 (50…45), then 44 (44…40), then 38.
	// The 38% warm attribute is 12 points behind the leader and must stay last.
	rows := []db.GetSeasonResultsByUserRow{
		attr("edgy-50", 50, db.QuestionToneEDGY),
		attr("edgy-48", 48, db.QuestionToneEDGY),
		attr("edgy-46", 46, db.QuestionToneEDGY),
		attr("edgy-44", 44, db.QuestionToneEDGY),
		attr("edgy-42", 42, db.QuestionToneEDGY),
		attr("edgy-40", 40, db.QuestionToneEDGY),
		attr("warm-38", 38, db.QuestionToneWARM),
	}
	got := order(rows)
	if got[len(got)-1] != "warm-38" {
		t.Errorf("warm-38 should stay last, 12 points behind the leader: %v", got)
	}
	if got[0] != "edgy-50" {
		t.Errorf("the strongest result should lead: %v", got)
	}
}

func TestByTone_NeverPromotesPastTheBandForAnyInput(t *testing.T) {
	// The property, checked over a ladder of every tone at every 1-point step: an attribute may only
	// appear above a stronger one when they are within the band.
	tones := []db.QuestionTone{db.QuestionToneWARM, db.QuestionToneNEUTRAL, db.QuestionToneEDGY}
	rows := make([]db.GetSeasonResultsByUserRow, 0, 60)
	for i := 0; i < 60; i++ {
		rows = append(rows, attr(fmt.Sprintf("q%02d", i), float64(60-i), tones[i%3]))
	}

	got := ByTone(rows)
	for i := range got {
		for j := i + 1; j < len(got); j++ {
			if got[j].Percentage > got[i].Percentage &&
				got[j].Percentage-got[i].Percentage > BandPoints {
				t.Fatalf("%.1f%% appeared below %.1f%%, more than %.0f points weaker in position",
					got[i].Percentage, got[j].Percentage, BandPoints)
			}
		}
	}
}

func TestByTone_RemovesNothing(t *testing.T) {
	rows := []db.GetSeasonResultsByUserRow{
		attr("a", 50, db.QuestionToneEDGY),
		attr("b", 48, db.QuestionToneWARM),
		attr("c", 20, db.QuestionToneNEUTRAL),
		attr("d", 20, db.QuestionToneWARM),
		attr("e", 1, db.QuestionToneEDGY),
	}
	got := ByTone(rows)
	if len(got) != len(rows) {
		t.Fatalf("got %d attributes, want %d — ordering must be a permutation", len(got), len(rows))
	}
	seen := map[string]int{}
	for _, r := range got {
		seen[r.QuestionText]++
	}
	for _, r := range rows {
		if seen[r.QuestionText] != 1 {
			t.Errorf("%q appears %d times in the output", r.QuestionText, seen[r.QuestionText])
		}
	}
}

func TestByTone_DoesNotMutateItsInput(t *testing.T) {
	// splitAttributes is not the only reader of the results slice — computeTrend takes the same one.
	rows := []db.GetSeasonResultsByUserRow{
		attr("edgy-40", 40, db.QuestionToneEDGY),
		attr("warm-38", 38, db.QuestionToneWARM),
	}
	_ = ByTone(rows)
	if rows[0].QuestionText != "edgy-40" {
		t.Errorf("input was reordered: %v", rows[0].QuestionText)
	}
}

func TestByTone_EmptyAndSingle(t *testing.T) {
	if got := ByTone(nil); len(got) != 0 {
		t.Errorf("nil should stay empty, got %v", got)
	}
	one := []db.GetSeasonResultsByUserRow{attr("only", 10, db.QuestionToneEDGY)}
	if got := ByTone(one); len(got) != 1 || got[0].QuestionText != "only" {
		t.Errorf("a single attribute should pass through, got %v", got)
	}
}
