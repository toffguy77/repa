# Proposal

## Why

A new group's first Reveal — the entire payoff of the product — is 6 to 11 days away.
`getNextSeasonDates()` (`backend/internal/service/groups/service.go:533-554`) always
targets the *next* Monday, and when a group is created on a Monday the
`daysUntilMonday == 0 → 7` branch (`:538-541`) pushes it a full week out, so Season #1
reveals on the Friday of the following week. For a 14–22 audience whose attention span
is hours, the product is empty for its entire first week — and because the shared card is
the only organic acquisition channel, nothing is shareable during the window when a new
user is most motivated to talk about the app.

At the same time nothing stops a Reveal from being worthless or non-anonymous:
`ProcessReveal` (`backend/internal/service/reveal/service.go:38-85`) has no floor on
member or voter count, so a one-member group force-reveals an empty card after three
retries (`:67-70`), and in a 2–3 member group percentages trivially identify who said
what. The constant block (`backend/internal/service/groups/service.go:28-35`) enforces
`MaxMembersPerGroup = 50` but has no lower bound at all, even though the PRD specifies
5–50.

## What Changes

- A new group's Season #1 becomes a **kickoff season**: it opens for voting immediately
  on group creation and reveals roughly an hour after the group first reaches reveal
  eligibility, instead of waiting for a calendar Friday.
- **BREAKING** (internal scheduling contract): weekly seasons target the *upcoming*
  Friday rather than the Friday of the following week, with a guaranteed minimum voting
  window. A season is created as soon as the group has none, not only on Sunday evening.
- Reveal gains hard eligibility floors: a season SHALL NOT reveal with fewer than 3
  members or fewer than 3 completed voters. When the floor is unmet the season is
  postponed instead of force-revealed, and members are told how many more people are
  needed. No user is ever shown an empty card.
- Groups below 5 members get explicit small-group handling: an anonymity notice in the
  app and the detector disabled (with no crystals charged), because a voter list drawn
  from 2–4 possible voters is not anonymous.

## Capabilities

### New Capabilities

- `season-lifecycle`: when seasons open, when they reveal, how the kickoff season differs
  from the weekly cadence, and what happens when a Reveal cannot legitimately happen.
- `small-group-anonymity`: the guarantees Repa makes about anonymity and card quality in
  groups too small for anonymous aggregation to hold.

### Modified Capabilities

None — this is the first OpenSpec change in the project, so both capabilities are new.

## Impact

- **Schema:** new migration adding `seasons.kind` (`KICKOFF` | `WEEKLY`) and
  `seasons.postpone_count`; `make sqlc` regeneration.
- **Backend:** `internal/service/groups/service.go` (season date calculation, kickoff
  creation, season creator), `internal/service/reveal/service.go` (eligibility floors,
  postponement), `internal/service/voting/service.go` (schedule kickoff reveal when
  eligibility is first reached), `internal/worker/tasks/reveal.go`,
  `internal/db/queries/seasons.sql`, cron registration in `cmd/server/main.go`.
- **API:** `GET /groups/:id` and `GET /groups` season payloads gain the information the
  client needs to explain a pending or postponed Reveal; detector endpoints gain a
  `GROUP_TOO_SMALL` error.
- **Mobile:** group screen and reveal waiting state must render "need N more people" and
  the small-group anonymity notice.
- **Docs:** `docs/features/groups.md`, `docs/features/reveal.md`, `docs/features/voting.md`.
- **Non-goals** are listed below.

## Non-goals

- Changing the Friday 20:00 MSK weekly Reveal ritual itself, or moving to per-timezone
  reveals — the weekly ritual stays, only the first cycle and the gap-filling change.
- Introducing instant per-vote feedback during the week (separate change).
- Changing quorum percentages (50% / 40%) or the 3-retry/2-hour retry behaviour, beyond
  adding the absolute voter floor on top of them.
- Enforcing a minimum of 5 members to *create* or *join* a group — the floor applies to
  revealing, not to group existence, so a founder can still create a group alone and
  invite people.
