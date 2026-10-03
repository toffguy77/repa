// Package eligibility owns the rules that decide whether a season may reveal, and the
// state the app shows while it may not. Both the reveal worker (which gates on it) and
// the groups API (which explains it to the member) depend on this one implementation —
// two copies would let the app's explanation drift from the worker's behaviour.
package eligibility

import db "github.com/repa-app/repa/internal/db/sqlc"

const (
	// MinMembers and MinVoters are absolute floors. A season never reveals below them,
	// not even through the forced-reveal retry path: the card would be empty, or the
	// percentages would identify individual voters.
	MinMembers = 3
	MinVoters  = 3

	// MinDetectorMembers is the smallest group where a voter list is still anonymous
	// enough to sell. Below it the list is close to "everyone except you".
	MinDetectorMembers = 5
)

// Status is the full picture of where a season stands relative to the reveal rules.
type Status struct {
	MemberCount     int64
	VotedCount      int64
	MembersNeeded   int64
	VotersNeeded    int64
	QuorumThreshold float64
	QuorumMet       bool
	// FloorMet reports the absolute member/voter floors, independent of quorum.
	FloorMet bool
	// Eligible means the season can reveal right now: floors and quorum both satisfied.
	Eligible bool
}

// Evaluate derives the status from the two counts the rules depend on. votedCount is the
// number of members who completed every question, matching the progress shown in the app.
func Evaluate(memberCount, votedCount int64) Status {
	threshold := 0.5
	if memberCount < 8 {
		threshold = 0.4
	}

	s := Status{
		MemberCount:     memberCount,
		VotedCount:      votedCount,
		QuorumThreshold: threshold,
		QuorumMet:       float64(votedCount) >= float64(memberCount)*threshold,
	}
	if memberCount < MinMembers {
		s.MembersNeeded = MinMembers - memberCount
	}
	if votedCount < MinVoters {
		s.VotersNeeded = MinVoters - votedCount
	}
	s.FloorMet = s.MembersNeeded == 0 && s.VotersNeeded == 0
	s.Eligible = s.FloorMet && s.QuorumMet

	return s
}

// RevealState is what the app renders for a season's pending Reveal.
type RevealState string

const (
	// StateRevealed — results are available.
	StateRevealed RevealState = "REVEALED"
	// StateWaitingForMembers — the group itself is too small to reveal.
	StateWaitingForMembers RevealState = "WAITING_FOR_MEMBERS"
	// StateWaitingForVoters — the group is big enough but too few have finished voting.
	StateWaitingForVoters RevealState = "WAITING_FOR_VOTERS"
	// StatePostponed — a reveal time already passed without the floors being met.
	StatePostponed RevealState = "POSTPONED"
	// StateScheduled — the Reveal will happen at the season's reveal_at.
	StateScheduled RevealState = "SCHEDULED"
)

// StateFor reports what the app should say about a season's Reveal. A missing floor wins
// over POSTPONED, because "3 more people need to vote" is actionable and "postponed" is not.
func StateFor(seasonStatus db.SeasonStatus, postponeCount int32, s Status) RevealState {
	if seasonStatus == db.SeasonStatusREVEALED || seasonStatus == db.SeasonStatusCLOSED {
		return StateRevealed
	}
	switch {
	case s.MembersNeeded > 0:
		return StateWaitingForMembers
	case s.VotersNeeded > 0:
		return StateWaitingForVoters
	case postponeCount > 0:
		return StatePostponed
	default:
		return StateScheduled
	}
}

// GrowthThreshold is the next group size that changes what the group can do, and what it changes.
//
// Returned to the app so one place owns both the number and the consequence. The arrangement mirrors
// RevealState, and for the same reason: the worker's behaviour and the app's explanation must not
// drift, and a second list of sizes in the client is exactly how they would.
type GrowthThreshold struct {
	// Size is the member count at which the change happens.
	Size int64 `json:"size"`
	// Needed is how many more members are required to reach it.
	Needed int64 `json:"needed"`
	// Unlocks names what crossing it changes. A key rather than copy: the wording belongs to the app.
	Unlocks string `json:"unlocks"`
}

// Growth unlock keys.
const (
	// UnlockReveal — below MinMembers a season cannot reveal at all.
	UnlockReveal = "REVEAL"
	// UnlockDetector — below MinDetectorMembers no rung of the detector is purchasable, and the
	// percentages are small enough to identify individual voters.
	UnlockDetector = "DETECTOR"
)

// NextGrowthThreshold reports the next size threshold the group has not crossed, or nil when it has
// crossed them all.
//
// Two counts, because the two thresholds are judged against different things, and conflating them puts
// two contradictory statements on the same screen:
//
//   - memberCount is the group's actual membership. The Reveal is a group-level event — Evaluate gates
//     it on this — so a reader who has blocked someone must not be told the Reveal cannot happen when
//     it is about to.
//   - effectiveCount is the members this reader can be rated by: membership minus blocks in either
//     direction, within this group. The detector and the anonymity warning are both judged on this
//     (see reveal.detectorTooSmall), so the threshold has to be too.
//
// Only those two sizes actually change behaviour. The quorum share rising from 40% to 50% at 8 members
// is deliberately absent: it makes revealing *harder*, and presenting it as something to unlock would
// be a lie told in the user's own interest. Above the last real threshold this returns nil, which the
// app renders as "big enough" rather than inventing a goal.
func NextGrowthThreshold(memberCount, effectiveCount int64) *GrowthThreshold {
	for _, t := range []struct {
		size    int64
		against int64
		unlocks string
	}{
		{MinMembers, memberCount, UnlockReveal},
		{MinDetectorMembers, effectiveCount, UnlockDetector},
	} {
		if t.against < t.size {
			return &GrowthThreshold{
				Size:    t.size,
				Needed:  t.size - t.against,
				Unlocks: t.unlocks,
			}
		}
	}
	return nil
}
