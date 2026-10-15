package schedule

import (
	"testing"
	"time"
)

func TestWeeklySeason_Table(t *testing.T) {
	msk := MSK()

	tests := []struct {
		name          string
		now           time.Time
		wantRevealDay int // day of month in MSK
	}{
		// Week of Mon 2026-10-05 .. Sun 2026-10-11; Fridays are the 9th and 16th.
		{"monday reveals same week", time.Date(2026, 10, 5, 10, 0, 0, 0, msk), 9},
		{"tuesday reveals same week", time.Date(2026, 10, 6, 10, 0, 0, 0, msk), 9},
		{"wednesday morning reveals same week", time.Date(2026, 10, 7, 9, 0, 0, 0, msk), 9},
		// Wed 20:01 -> Fri 20:00 is 47h59m away, under the 48h minimum.
		{"wednesday late reveals next week", time.Date(2026, 10, 7, 20, 1, 0, 0, msk), 16},
		{"thursday reveals next week", time.Date(2026, 10, 8, 10, 0, 0, 0, msk), 16},
		{"friday before reveal still next week", time.Date(2026, 10, 9, 10, 0, 0, 0, msk), 16},
		{"saturday reveals next week", time.Date(2026, 10, 10, 10, 0, 0, 0, msk), 16},
		{"sunday reveals next week", time.Date(2026, 10, 11, 22, 0, 0, 0, msk), 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startsAt, revealAt, endsAt := WeeklySeason(tt.now)

			if !startsAt.Equal(tt.now.UTC()) {
				t.Errorf("startsAt should equal now, got %s want %s", startsAt, tt.now.UTC())
			}

			r := revealAt.In(msk)
			if r.Weekday() != time.Friday {
				t.Errorf("revealAt should be Friday, got %s", r.Weekday())
			}
			if r.Hour() != 20 || r.Minute() != 0 {
				t.Errorf("revealAt should be 20:00 MSK, got %02d:%02d", r.Hour(), r.Minute())
			}
			if r.Day() != tt.wantRevealDay {
				t.Errorf("revealAt day = %d, want %d (revealAt %s)", r.Day(), tt.wantRevealDay, r)
			}

			if revealAt.Sub(tt.now) < MinVotingWindow {
				t.Errorf("voting window %s is shorter than the %s minimum", revealAt.Sub(tt.now), MinVotingWindow)
			}

			e := endsAt.In(msk)
			if e.Weekday() != time.Sunday {
				t.Errorf("endsAt should be Sunday, got %s", e.Weekday())
			}
			if e.Hour() != 23 || e.Minute() != 59 {
				t.Errorf("endsAt should be 23:59 MSK, got %02d:%02d", e.Hour(), e.Minute())
			}
			if !e.After(r) {
				t.Errorf("endsAt %s should be after revealAt %s", e, r)
			}
			if e.Sub(r) > 72*time.Hour {
				t.Errorf("endsAt should be in the same week as revealAt, gap was %s", e.Sub(r))
			}
		})
	}
}

func TestWeeklySeason_VotingOpensImmediately(t *testing.T) {
	now := time.Date(2026, 10, 5, 13, 37, 0, 0, MSK())

	startsAt, _, _ := WeeklySeason(now)

	if startsAt.After(now) {
		t.Errorf("startsAt %s must not be in the future relative to now %s", startsAt, now)
	}
	if startsAt.Weekday() != now.UTC().Weekday() {
		t.Errorf("startsAt should be the creation day, not a later Monday: got %s", startsAt.Weekday())
	}
}

func TestKickoffSeason_UsesWeeklyFallback(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, MSK())

	ks, kr, ke := KickoffSeason(now)
	ws, wr, we := WeeklySeason(now)

	if !ks.Equal(ws) || !kr.Equal(wr) || !ke.Equal(we) {
		t.Errorf("kickoff fallback schedule should match the weekly schedule:\n kickoff %s/%s/%s\n weekly  %s/%s/%s",
			ks, kr, ke, ws, wr, we)
	}
}

func TestNextRevealFriday(t *testing.T) {
	msk := MSK()

	// Exactly at a reveal moment, the next reveal is a week later.
	atReveal := time.Date(2026, 10, 9, 20, 0, 0, 0, msk)
	got := NextRevealFriday(atReveal).In(msk)
	if got.Day() != 16 {
		t.Errorf("at Friday 20:00 the next reveal should be the 16th, got %s", got)
	}

	// A minute before, the same Friday still counts.
	justBefore := time.Date(2026, 10, 9, 19, 59, 0, 0, msk)
	got = NextRevealFriday(justBefore).In(msk)
	if got.Day() != 9 {
		t.Errorf("a minute before reveal the same Friday should count, got %s", got)
	}
}

func TestEndOfRevealWeek(t *testing.T) {
	msk := MSK()
	friday := time.Date(2026, 10, 9, 20, 0, 0, 0, msk)

	got := EndOfRevealWeek(friday).In(msk)

	if got.Weekday() != time.Sunday {
		t.Errorf("want Sunday, got %s", got.Weekday())
	}
	if got.Day() != 11 {
		t.Errorf("want the 11th, got %s", got)
	}
}
