# Proposal

## Why

Nothing accumulates for the *group*. A season reveals on Friday, everyone looks at their own card, and
by Saturday there is no trace that it happened: the only history in the product is per-member, inside
`GET /profile` (`GetUserSeasonHistory`), and it shows you your own results in a list only you can see.
A group of friends playing for six weeks has no shared record of those six weeks — no "remember when
Катя was voted the most reliable three weeks running" — which is exactly the artifact that would make
the group worth returning to between Fridays, and the thing tbh and Gas lean on hardest.

The second half of the same gap: the product already measures who predicts the group best
(`user_group_stats.guess_accuracy`, a 5-season rolling average, computed every reveal) and then shows
it to one person privately. A number about how well you know your friends is only interesting next to
your friends' numbers. As a private stat it motivates nothing; as a group standing it is a reason to
vote every week even in a week when you do not care about your own card.

Third: group size has real mechanical consequences — 3 members to activate, 5 for the detector and for
percentages that do not identify voters, 5 before a season carries its full 10 questions — and none of
them are visible. Members have no idea that inviting two more people unlocks anything, so groups sit at
the quorum minimum, which is also where the product is at its weakest.

## What Changes

- **A group chronicle.** A per-group, members-visible timeline of revealed seasons: for each past
  season, the standout result per question — who led it and at what percentage — the same shape the
  current season's `group_summary.top_per_question` already returns, extended backwards over history.
  Readable by any member, covering only seasons that legitimately revealed.
- **A «знатоки» standing.** The group's members ranked by guess accuracy, from data already computed
  at reveal. Shown with the number of seasons each standing is based on, and withheld entirely until
  the group has enough revealed history for the ranking to mean anything.
- **Visible unlocks.** The group screen states the next real threshold and what crossing it changes,
  derived from the thresholds already in `internal/eligibility` rather than from a second list. Above
  the last real threshold it says so instead of inventing a goal.
- **Mobile:** a chronicle screen reachable from the group, the standing on it, and the unlock line on
  the group screen.

Not in scope: new achievements, per-member chronicle pages (the profile already has one), notifications
about the chronicle (push has a budget of 3/day and a reveal already spends one), and any new paid
mechanic.

## Capabilities

### New Capabilities

- `group-chronicle`: what a group's shared history contains, who may read it, which seasons appear in
  it, and how the guess-accuracy standing is derived and gated.
- `group-growth`: what the product tells members about the consequences of group size, and the rule
  that it may only name thresholds that actually exist.

### Modified Capabilities

- `small-group-anonymity`: the chronicle is a second surface that exposes per-question winners, so the
  existing rule that small-group results identify voters has to cover it; and the standing must not
  become a way to infer who voted how in a tiny group.

## Impact

- **Backend:** new queries over `seasons` / `season_results` / `user_group_stats` (no schema change —
  every number the chronicle shows is already stored); a new handler package or endpoints under
  `/api/v1/groups/:id/chronicle`; `internal/eligibility` gains the threshold descriptions so the
  unlock line and the mechanics cannot disagree.
- **Mobile:** a new chronicle screen and Freezed models, an entry point on the group screen, and the
  unlock line.
- **Docs:** `docs/features/groups.md`, `docs/features/profile.md`, `docs/features/reveal.md`.
- **No migrations.** This change adds no columns and no tables.
