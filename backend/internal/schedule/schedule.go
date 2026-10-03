// Package schedule owns every season timing rule in one place: the MVP schedules
// exclusively in MSK (see docs/specs/master_context.md), and both the groups service
// (season creation) and the reveal service (postponement) must agree on when the next
// Reveal is, or a postponed season drifts off the weekly rhythm.
package schedule

import "time"

const (
	// MinVotingWindow is the shortest voting window a weekly season may have. A season
	// created closer than this to the upcoming Friday reveals the Friday after instead,
	// so a season is never so short that most of the group misses it.
	MinVotingWindow = 48 * time.Hour

	// KickoffRevealDelay is how long after a new group first becomes reveal-eligible its
	// kickoff Reveal fires. It is deliberately not zero: the delay gives the "Reveal
	// within the hour" push time to land and pull the group back in together, and leaves
	// room for the fourth and fifth member to finish voting.
	KickoffRevealDelay = time.Hour
)

// MSK is the fixed UTC+3 zone the product schedules in.
func MSK() *time.Location {
	return time.FixedZone("MSK", 3*60*60)
}

// NextRevealFriday returns the first Friday 20:00 MSK strictly after now.
func NextRevealFriday(now time.Time) time.Time {
	return NextRevealFridayAfter(now, 0)
}

// NextRevealFridayAfter returns the first Friday 20:00 MSK strictly after now+minLead.
func NextRevealFridayAfter(now time.Time, minLead time.Duration) time.Time {
	msk := MSK()
	cutoff := now.In(msk).Add(minLead)

	daysUntilFriday := (int(time.Friday) - int(cutoff.Weekday()) + 7) % 7
	friday := time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day()+daysUntilFriday, 20, 0, 0, 0, msk)
	for !friday.After(cutoff) {
		friday = friday.AddDate(0, 0, 7)
	}
	return friday
}

// EndOfRevealWeek returns Sunday 23:59 MSK of the week containing revealAt.
func EndOfRevealWeek(revealAt time.Time) time.Time {
	msk := MSK()
	r := revealAt.In(msk)
	daysUntilSunday := (int(time.Sunday) - int(r.Weekday()) + 7) % 7
	if daysUntilSunday == 0 {
		daysUntilSunday = 7 // a Sunday reveal would otherwise end the same moment
	}
	return time.Date(r.Year(), r.Month(), r.Day()+daysUntilSunday, 23, 59, 0, 0, msk)
}

// WeeklySeason returns the schedule for a weekly season created at now: voting opens
// immediately and reveals on the next Friday 20:00 MSK at least MinVotingWindow away.
func WeeklySeason(now time.Time) (startsAt, revealAt, endsAt time.Time) {
	revealAt = NextRevealFridayAfter(now, MinVotingWindow)
	return now.UTC(), revealAt.UTC(), EndOfRevealWeek(revealAt).UTC()
}

// KickoffSeason returns the schedule for a new group's first season. Voting opens
// immediately; revealAt is the weekly fallback Friday, which the voting service brings
// forward to KickoffRevealDelay once the group first becomes reveal-eligible.
func KickoffSeason(now time.Time) (startsAt, revealAt, endsAt time.Time) {
	return WeeklySeason(now)
}
