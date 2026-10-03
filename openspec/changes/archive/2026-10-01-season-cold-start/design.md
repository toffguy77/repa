# Design

## Context

See proposal.md — Why. The relevant existing machinery:

- `getNextSeasonDates()` (`backend/internal/service/groups/service.go:533`) is the single
  place season dates are computed, used both by group creation and by `CreateNewSeasons`
  (`:556`).
- The reveal pipeline is already event-driven on a timestamp: `reveal-checker` runs every
  minute (`backend/cmd/server/main.go:414`) and selects `status = 'VOTING' AND reveal_at
  <= NOW()` (`internal/db/queries/seasons.sql:14-15`), enqueuing `reveal-process` per
  season. `ProcessReveal` (`internal/service/reveal/service.go:38`) returns
  `{Revealed, Retry}` and the worker re-enqueues with a 2-hour delay while `Retry` is set.
- `GetGroupsNeedingNewSeason` (`internal/db/queries/seasons.sql:17-21`) already filters to
  groups with `COUNT(members) >= 3`, but only runs from the Sunday 21:00 MSK cron
  (`cmd/server/main.go:426`).

This means the kickoff mechanic needs no new scheduler: scheduling a kickoff Reveal is
just writing a `reveal_at` one hour in the future, and the existing per-minute checker
picks it up.

## Goals / Non-Goals

**Goals:**

- Reuse the existing timestamp-driven reveal pipeline rather than adding a parallel
  scheduling path.
- Keep one authoritative function for weekly date math, so the cadence cannot drift
  between group creation and the season creator.
- Make reveal eligibility a single predicate that both the reveal path and the API
  read, so the app's explanation of "why no Reveal yet" can never disagree with what
  the worker does.

**Non-Goals:**

- Per-group or per-timezone reveal times.
- Reworking achievements, card generation, or push content beyond what the new states
  require.
- Backfilling or rescheduling seasons that already exist in the dev database; the
  migration only adds columns with defaults.

## Decisions

### Explicit `seasons.kind` column over inferring from `number = 1`

A new `season_kind` enum (`KICKOFF`, `WEEKLY`) with `DEFAULT 'WEEKLY'` is added to
`seasons`. Inferring "kickoff" from `number = 1` was rejected: a group whose kickoff
season is postponed for weeks would still be season 1 while behaving weekly, and future
work (seasonal events, re-activated dormant groups) would overload the number further. An
explicit column keeps the branch in the date logic readable and makes existing rows
correct by default.

### Kickoff reveal is scheduled on the eligibility transition, inside the vote transaction

`CastVote` already knows when a voter completes their last question. When that completion
makes the season eligible for the first time (member floor met, voter floor met, season
is `KICKOFF`, `reveal_at` not yet brought forward), the service sets
`reveal_at = NOW() + 1 hour`.

Alternatives considered:

- *A dedicated per-minute cron scanning kickoff seasons for eligibility.* Rejected: a
  second scanner duplicating the eligibility predicate is exactly the drift risk this
  design is trying to avoid, and the information is already in hand at vote time.
- *Revealing immediately on the third vote.* Rejected: the one-hour delay is doing real
  product work — it gives the push ("Reveal within the hour") time to land and pull the
  group back into the app together, which is what makes the Reveal feel like an event
  rather than a page refresh. It also leaves room for the fourth and fifth member to get
  their votes in.

Scheduling is a conditional `UPDATE ... WHERE kind = 'KICKOFF' AND reveal_at > $new` so
concurrent completions cannot push the time back, satisfying the "not pushed back by
later voters" scenario without a lock.

### Weekly dates: upcoming Friday, with a 48-hour minimum voting window

`getWeeklySeasonDates(now)` replaces `getNextSeasonDates()`:

- `starts_at = now`
- `reveal_at` = this week's Friday 20:00 MSK if that is at least 48 hours away, otherwise
  next week's Friday 20:00 MSK
- `ends_at` = Sunday 23:59 MSK of the reveal week

The 48-hour floor is the one knob that trades off speed against a fair voting window. A
24-hour floor would let a Wednesday-evening season reveal Friday with most members never
seeing it; a 72-hour floor would push any Tuesday season a week out, recreating the
problem this change exists to fix.

### Gap filling: event-driven for kickoff, cron for the weekly rhythm

Three paths, deliberately:

1. **After a kickoff reveal** — the reveal worker enqueues weekly-season creation for that
   group immediately. A group that reveals Wednesday starts its weekly cycle Wednesday.
2. **Sunday 21:00 MSK cron** — unchanged, and still the normal path. Friday 20:00 to
   Sunday 23:59 stays reserved for discussion, reactions, and detectors; opening a season
   right after the Friday reveal would cannibalise that window.
3. **Hourly maintenance cron** — creates a season for any eligible group that has none and
   whose latest season ended more than an hour ago. This is the safety net for dropped
   jobs and for groups that crossed the 3-member threshold mid-week; the one-hour grace
   keeps it from racing path 1 or 2.

All three call the same creation function, so the cadence rule lives in one place.

### Eligibility is one predicate, consumed by three callers

`reveal.Eligibility(ctx, seasonID) → {MemberCount, VotedCount, MembersNeeded, VotersNeeded, Eligible}`
is computed from the existing `CountGroupMembers` and `CountUniqueVoters` queries and used
by: `ProcessReveal` (gate), `CastVote` (kickoff scheduling), and the groups handler (the
`reveal_state` the app renders). One predicate is what keeps the app's explanation and the
worker's behaviour from disagreeing.

The floors are constants in the reveal service: `MinRevealMembers = 3`,
`MinRevealVoters = 3`. They sit *above* the percentage quorum — the forced-reveal path
after three retries bypasses the percentage, never the floors.

### Postponement moves `reveal_at` instead of closing the season

When the floors are unmet after retries, `reveal_at` moves to the next Friday 20:00 MSK,
`ends_at` moves to that week's Sunday, `postpone_count` increments, and the status stays
`VOTING`. Votes are preserved: discarding a small group's accumulated votes would punish
the people who did participate, which is the opposite of the intent. `postpone_count` is
stored rather than derived so the push copy can escalate ("still waiting" vs. "third week
without a Reveal") and so dormant groups are findable later.

A season that has been postponed keeps `kind = 'KICKOFF'` if it started as one — that is
what makes the "falls back to the weekly schedule" requirement true without a second
state machine: a postponed kickoff is simply a kickoff whose `reveal_at` now lands on a
Friday.

### `reveal_state` on the season payload rather than a new endpoint

The active-season object in `GET /groups` and `GET /groups/:id` gains
`reveal_state ∈ {SCHEDULED, WAITING_FOR_MEMBERS, WAITING_FOR_VOTERS, POSTPONED, REVEALED}`
plus `members_needed` and `voters_needed`. Adding fields to a payload the app already
fetches on every group open avoids a second round trip on the screen that most needs to
explain itself. Both existing group endpoints return season sub-objects with different
shapes already (noted in `docs/features/groups.md`); this change adds the same three
fields to both rather than unifying them, which is a larger refactor than this change
should carry.

### Detector gating lives in the reveal service, before the crystal deduction

`BuyDetector` checks member count first and returns `ErrGroupTooSmall` before opening the
transaction, so the "balance unchanged, no detector record" scenario holds by
construction rather than by rollback. `GetDetector` reports the same condition so the app
can disable the button instead of letting a member discover it by spending.

## Risks / Trade-offs

- **A kickoff Reveal with exactly 3 voters is weakly anonymous** → the floors make a card
  possible at 3 members, where 2 voters determine every percentage. Mitigated by the
  small-group-anonymity notice and by disabling the detector below 5 members; accepted
  because the alternative (floor of 5) reintroduces a multi-day cold start for the common
  case of a founder plus two friends.
- **Hourly maintenance cron could double-create a season** → creation is guarded by the
  same `NOT EXISTS (season WHERE status = 'VOTING')` predicate inside the creating
  transaction, and the one-hour grace window keeps it clear of the other two paths.
- **Faster cadence raises push volume** → the 3-per-day cap and quiet hours in the push
  service already bound this; the kickoff "Reveal within the hour" push is a new send that
  counts against that cap like any other.
- **Groups created mid-week now reveal the same week with a shorter window** → that is the
  intended trade: 48 hours of voting beats 11 days of nothing, and the participation
  floors stop a thin season from producing a bad card.
- **`reveal_at` becomes a mutable field with three writers** (creation, kickoff
  scheduling, postponement) → all three go through the season service, and the kickoff
  update is conditional, so the invariant "`reveal_at` only moves earlier for a kickoff,
  only later for a postponement" is enforced in SQL rather than by convention.

## Migration Plan

1. Migration `004_season_kickoff`: create `season_kind` enum, add
   `seasons.kind season_kind NOT NULL DEFAULT 'WEEKLY'` and
   `seasons.postpone_count integer NOT NULL DEFAULT 0`. Existing rows become `WEEKLY`
   with zero postponements, which matches their current behaviour.
2. `make sqlc`.
3. Deploy backend. Existing open seasons keep their `reveal_at`, so no in-flight season
   changes behaviour; only seasons created after deploy use the new cadence.
4. Register the hourly maintenance cron.
5. Rollback: `004_season_kickoff.down.sql` drops both columns and the enum. The code paths
   that read `kind` are additive, so reverting the binary alone is also safe.
