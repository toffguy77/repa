# Design

## Context

See proposal.md — Why. What exists: `reports(question_id NOT NULL, reporter_id, reason)` with a
`UNIQUE(question_id, reporter_id)`, surfaced by `ListReports` in the admin handler. Group membership is
`group_members(user_id, group_id)` with `UNIQUE(user_id, group_id)`; `LeaveGroup` deletes the row and
transfers admin if needed. Voting targets come from `GetGroupMembers` minus the voter.

## Goals / Non-Goals

**Goals:**

- A member can get away from another member, immediately, without leaving the group.
- A removal that an invite cannot undo.
- Compliance with the store requirement that there be a block, a report, and content filtering.

**Non-Goals:**

- Inferring abuse from anonymous voting patterns — acting on that would require exposing who voted.
- An admin moderation console, or account-level suspension.

## Decisions

### A separate `user_reports` table, not a nullable `question_id`

`reports.question_id` is `NOT NULL` and every existing query joins through it. Making it nullable would
force each of those to handle the null case for the sake of a different kind of report. `user_reports`
keeps the question queue exactly as it is and gives the person queue its own `UNIQUE(reported_id,
reporter_id)` — one report per person per reporter, which is the same shape the question queue already
uses.

Both appear in the admin queue as separate lists rather than a merged feed: they need different
columns, and a merged list would have to be sparse in both directions.

### A block is one row, but mutual in effect

`blocks(blocker_id, blocked_id)` records who created it, so it can be undone by that person. The
*effect* is symmetric: filtering uses `blocker_id = me OR blocked_id = me`.

One-directional blocking was rejected because of what it would mean here. If A blocks B but B can still
rate A, the product continues asking B to judge someone who has withdrawn — and A still receives B's
votes, which is exactly the harm the block was for. A symmetric effect costs B the ability to rate one
person, which is the right trade in a rating product.

### Votes cast before a block are kept

Removing a blocked person's earlier votes would change a member's percentages at the moment of a block,
which is both a worse result (the member's card quietly shifts) and an information leak: a careful
observer comparing results before and after would learn that a block happened and roughly when.
Filtering applies to *future* targets only.

### `group_bans` separates "left" from "cannot come back"

Ordinary leaving stays reversible — people leave groups by accident and rejoin. A removal by an admin,
and a deliberate permanent departure, write a `group_bans(group_id, user_id)` row that the join path
checks. This is why regenerating the invite does not undo a removal: the ban is on the person and the
group, not on the code.

Alternatives considered: a `banned` flag on `group_members` (rejected — the row is deleted on removal,
so the flag would have to keep a tombstone row that every membership query then has to exclude), and
tracking bans in Redis (rejected — a ban must outlive a cache).

### Target filtering happens in one query, not in Go

`GetVotingTargets(group_id, voter_id)` returns members minus the voter minus anyone blocked in either
direction. Doing it in SQL rather than filtering `GetGroupMembers` in the service means the voting
session, the vote validation, and any future caller cannot disagree about who is rateable — the bug
this shape prevents is a target list that offers someone the vote endpoint then refuses.

### The anonymity floor counts what is voteable

`small-group-anonymity` thresholds move from raw member count to *effective* size — members minus
blocks involving the viewer. A group of five with three blocks presents itself as safely anonymous
while behaving like a group of two, which is precisely the case the floor exists for.

## Risks / Trade-offs

- **Blocks can be used to shrink a group below the voting floor** → the participation floors then
  postpone the Reveal rather than producing a thin result, which is the correct outcome: a group that
  has fragmented into blocks should not be producing reputation cards.
- **A symmetric block costs the non-blocking party something they did not choose** → accepted and
  argued above; the alternative keeps the harm the block exists to stop.
- **Two report queues instead of one** → they carry different columns; merging would make both sparse.
- **A ban is per group, so a determined person can be invited to a *new* group** → that is a different
  group with a different admin, and solving it would mean account-level banning, which is out of scope.

## Migration Plan

1. Migration `009_member_safety`: `blocks`, `group_bans`, `user_reports`, each with its unique
   constraint and lookup indexes.
2. `make sqlc`.
3. Deploy backend, then client. Until the client ships, the endpoints are unused and nothing changes.
4. Rollback: drop the three tables. Target filtering falls back to "everyone except me", which is the
   current behaviour.
