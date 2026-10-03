# Proposal

## Why

`GET /api/v1/groups/:id/members/:userId/profile` returns `stats.guess_accuracy` — how often that member's
votes matched the result the group arrived at — to **any** member of the group. On its own it reads like a
harmless stat. Next to what the API already publishes, it is a way to work out how someone voted.

Two routes, both open today:

1. **Winners are public.** The reveal summary (`group_summary.top_per_question`) and the group chronicle
   both name which member led each question. A viewed member at or near 100% voted for those winners; at
   or near 0% they voted for someone else on every question. In a season of 3–10 questions that pins or
   badly narrows each of their individual votes.
2. **The figure is a rolling average, so it can be differenced.** It is recomputed at every reveal as
   `new = (old*w + seasonAccuracy) / (w + 1)` with `w = min(seasons_played, 4)`, and `seasons_played` is
   returned alongside it. Reading the same member's profile on two consecutive weeks solves for that
   week's accuracy exactly, and therefore for how many of that week's questions they matched.

This is the rule in CLAUDE.md — *API responses never expose `voter_id` in connection to specific votes* —
being satisfied in its letter and broken in its substance: no vote is in the response, and a vote comes
out of it anyway. The `season-chronicle` change already removed the figure from the group standing for
exactly this reason; the profile endpoint is the same leak by a different door, and was left alone then
because it predates that work.

## What Changes

- `stats.guess_accuracy` is returned **only on the viewer's own profile**. A member's own accuracy tells
  them nothing they do not already know — they cast the votes — while another member's is the leak.
- On someone else's profile the field is absent rather than zeroed, so a client cannot mistake "withheld"
  for "never matched anything".
- `seasons_played`, `voting_streak`, `max_voting_streak`, `total_votes_cast` and `total_votes_received`
  stay visible: none of them is a count of *matches*, so none can be combined with the published winners
  to recover a vote.
- Mobile shows the accuracy tile only on one's own profile.

Not in scope: changing how accuracy is computed or stored, the chronicle standing (already handled), and
the achievements derived from accuracy — an achievement is a threshold crossed, not a figure, and cannot
be differenced.

## Capabilities

### Modified Capabilities

- `small-group-anonymity`: the capability already holds the rules that stop results identifying voters.
  It gains the rule that a *derived* per-member figure counts as identifying when the API publishes
  something it can be combined with.

## Impact

- **Backend:** `internal/service/profile` — `StatsDto.GuessAccuracy` becomes a pointer so it can be
  omitted, populated only when `requesterID == userID` (the service already receives both).
- **Mobile:** `StatsDto.guessAccuracy` becomes nullable; the profile screen hides the tile when absent.
- **Docs:** `docs/features/profile.md`, `docs/specs/master_context.md`, and the anonymity convention in
  `CLAUDE.md`.
- **No migrations**, and no change to any stored value.
- **Breaking for a client that assumes the field is always present** — the Freezed model is updated in
  the same change, and an older client reading a missing double would have crashed, so the field is
  omitted rather than renamed.
