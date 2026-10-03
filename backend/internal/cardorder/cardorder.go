// Package cardorder decides the order a member's attributes appear in on their reputation card.
//
// Its own package because four places build a card from the same results and must agree: the reveal
// API, the members-cards list, the purchased "show everything" view, and the worker that renders the
// shared PNG. The PNG is the one that matters most — it is the artifact a person is invited to send —
// and it is produced furthest from the API, so a rule living inside the reveal service would have been
// silently skipped there.
package cardorder

import (
	"sort"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// BandPoints is how close two attributes must be, in percentage points, for tone to decide which
// leads.
//
// Five points: wide enough to break the ties that actually occur in groups of 5–50, where a single
// vote is worth 2–20 points, and narrow enough that it cannot flip a real difference.
const BandPoints = 5.0

// ByTone orders a member's results by percentage, letting tone break near-ties.
//
// Why near-ties and not an outright preference: the card's value is that it is true. A person who
// overwhelmingly got one edgy attribute should see it leading, and a card that buried it would be a
// different product. But at 40% and 38% the ordering is arbitrary anyway, and the card is the thing
// the person is invited to share — so the arbitrary choice may as well favour what they'd want to send.
//
// Implemented as *anchor clusters* rather than as a comparison function, because "within 5 points" is
// not transitive: 50% edgy, 47% neutral, 44% warm would give warm < neutral < edgy < warm, and sorting
// with an inconsistent comparator yields an undefined order that can lift the 44% warm attribute above
// the 50% edgy one — the exact reordering this is supposed to forbid. A cluster holds the attributes
// within the band of the cluster's own strongest attribute, and tone orders only inside a cluster. The
// anchor is fixed, so no attribute is ever promoted past one more than the band stronger, however long
// the chain of near-ties.
//
// Nothing is removed and nothing is hidden: this is a permutation of its input.
func ByTone(results []db.GetSeasonResultsByUserRow) []db.GetSeasonResultsByUserRow {
	if len(results) < 2 {
		return results
	}

	ordered := make([]db.GetSeasonResultsByUserRow, len(results))
	copy(ordered, results)

	// The query already orders by percentage DESC, but sorting here keeps this function correct on its
	// own — it is the only thing the card's order depends on.
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Percentage > ordered[j].Percentage
	})

	for start := 0; start < len(ordered); {
		anchor := ordered[start].Percentage
		end := start + 1
		for end < len(ordered) && anchor-ordered[end].Percentage <= BandPoints {
			end++
		}

		cluster := ordered[start:end]
		sort.SliceStable(cluster, func(i, j int) bool {
			if ri, rj := toneRank(cluster[i].QuestionTone), toneRank(cluster[j].QuestionTone); ri != rj {
				return ri < rj
			}
			// Same tone: the stronger result leads, as it would without any of this.
			return cluster[i].Percentage > cluster[j].Percentage
		})

		start = end
	}

	return ordered
}

// toneRank is the order tones lead in within a cluster: warm, then neutral, then edgy.
func toneRank(t db.QuestionTone) int {
	switch t {
	case db.QuestionToneWARM:
		return 0
	case db.QuestionToneEDGY:
		return 2
	default:
		return 1
	}
}
