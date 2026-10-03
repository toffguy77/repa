# Design

## Context

See proposal.md — Why. What exists: `votes(season_id, voter_id, target_id, question_id)` already
holds everything needed to count distinct voters per target, and `questions.category` carries the
category. The push worker has a Tuesday signal task and a Thursday teaser task that both fan out to
every member of every voting group. `season-cold-start` gave the reveal screen a pending-Reveal
explanation; this change gives it something to say about *you*.

## Goals / Non-Goals

**Goals:**

- Something to come back for between Monday and Friday, without touching the Friday ritual.
- A push that is true: the screen behind it has the thing the push mentioned.
- No path by which a pre-Reveal signal narrows down an individual or an attribute.

**Non-Goals:**

- Per-vote attributed feedback (Gas's model) — it trades the ritual for a drip.
- A staged multi-screen Reveal.

## Decisions

### One endpoint returning a tiny, deliberately boring payload

`GET /seasons/:seasonId/anticipation` → `{ voters_about_me, teaser_emoji, reveal_at }`.

The payload is designed to be *incapable* of leaking rather than merely not leaking today: there is
no identity field and no attribute field, so no future addition to a shared DTO can turn it into a
leak. `teaser_emoji` is a string holding one emoji — not a category code the client could map back
to a name, and not a question id.

Served only while the season is open: after the Reveal the member has the real thing, and a partial
signal next to a full result is noise.

### The teaser is gated on the weekday, server-side

`time.Now().In(MSK).Weekday() >= time.Thursday` decides whether the teaser is included. Gating on
the client would make the teaser a client-version property and would ship the leading category to
every device from Monday — the restraint has to be where the data is.

Thursday is the PRD's choice (RET-09) and it is the right one for a reason worth recording: a teaser
on Monday has four days to become boring, while a teaser on Thursday has one night.

### Leading category, resolved by a single grouped query

`GetLeadingCategoryForTarget(season_id, target_id)` groups the member's received votes by the
question's category and returns the top one. Ties break by the category's own ordering, which is
arbitrary but stable — a random tiebreak would make the teaser flicker between two emojis on
consecutive loads, which reads as a bug.

The emoji mapping lives in Go beside the category enum, not in the database and not on the client:
the client must not be able to render a category it was not told about.

### The immediate signal is debounced to once per day, in Redis

A vote about a member enqueues a push. Without a bound, a 20-person group produces 19 notifications
in an evening and the app gets muted — so the signal is sent at most once per recipient per MSK day,
tracked with a Redis key that expires at midnight MSK. This is the same mechanism the daily push cap
already uses, so there is one idea in the codebase rather than two.

Debounce rather than batch: a batched digest ("3 people answered") is more information than the
product wants to give before Friday, because it tells you how fast interest is arriving.

The enqueue happens from the voting service after a vote is recorded, as a side effect that cannot
fail the vote — the same pattern as the kickoff scheduling and the referral payout.

### The mid-week pushes are filtered by what the recipient will see

`HandleTuesdaySignal` and `HandleThursdayTeaser` currently send to every member of every voting
group. They now check the recipient's own count (and, for Thursday, that a teaser exists) and skip
the ones for whom the screen would be empty.

This costs a query per member. That is acceptable: the alternative is the current behaviour, where a
push that promises intrigue and delivers nothing teaches the recipient to ignore the next one — which
costs the pushes that do work.

### The waiting state becomes the anticipation screen

The reveal screen's `waiting` phase renders the count, the teaser when present, and a live countdown,
on top of the pending-Reveal explanation `season-cold-start` added. Push `screen: "reveal-waiting"`
already routes to the group; it is re-pointed at the reveal route so the tap lands on the thing the
push mentioned.

## Risks / Trade-offs

- **A count can be correlated with knowing who was online** → it is a count of *voters about you* in
  a group of up to 50; at the small end the group-size anonymity notice already warns that a tiny
  group leaks. Nothing here is sharper than what the progress bar already shows.
- **The emoji narrows the result space** → to one of six categories, with no percentage, no question
  and no attribute. That is the intended tease: it should be enough to think about and not enough to
  know.
- **A per-member query in the push fan-out** → bounded by group size, runs twice a week off-peak.
- **The signal push competes for the 3-per-day cap** → by design; it is debounced to one per day so
  it cannot crowd out the Friday pushes.

## Migration Plan

No schema changes. Deploy the backend, then the client. Until the client ships, the endpoint is
unused and the mid-week pushes simply reach fewer people — which is already an improvement over
promising what cannot be shown.
