# Design

## Context

See proposal.md — Why. What already exists: `crystal_logs` has a UNIQUE `external_id` (used
today for YuKassa payment ids), a `BONUS` enum value nothing writes, and balance computed as
`SUM(delta)`. `card-share-attribution` added `group_members.join_source` and a per-channel
share link. Achievements are created by `achievements.Service` through one `CreateAchievement`
call site.

## Goals / Non-Goals

**Goals:**

- The detector gets tried before it gets sold.
- A referral reward that cannot be farmed by creating accounts.
- Grants that are safe to retry, because every trigger here is a path that can run twice.

**Non-Goals:**

- Changing what crystals cost or buy.
- Backfilling grants for existing users.

## Decisions

### `external_id` is the idempotency key, not a new table

Every grant carries a deterministic `external_id`: `welcome:{userID}`,
`referral:{inviteeID}`, `achievement:{achievementID}`. The column is already UNIQUE, so a
duplicate insert fails on the constraint and the grants service treats that specific failure as
success — the grant exists, which is what the caller wanted.

This is why the identity is derived rather than random: a random id would make every retry a new
payment. `referral:{inviteeID}` keyed on the *invitee* is what makes "once per invited member"
true by construction, rather than by a count query that races.

Alternatives considered: a `grants` table with a unique constraint (rejected — a second source
of truth for balance, which the product deliberately does not have), and an application-level
"have we paid?" check (rejected — it races, and the constraint already exists).

### The referral pays on the invitee's first *completed* session

Rewarding the join would pay for an account, and accounts are free. Rewarding a completed
voting session means the invitee answered every question about real people in a real group that
met the participation floor — work that is not worth faking for 5 crystals.

The trigger lives where completion is already detected: the voting service already knows when a
vote completes a session (it is the same place the kickoff Reveal is scheduled). The grant is
best-effort and logged on failure — a payout problem must not roll back a recorded vote.

### The referrer travels in the card's link, and the card is per-member

A group's invite code identifies the *group*, so a code alone cannot say who invited. The card,
however, is generated per member, so its QR link carries `&ref={userID}` and the shared text
does the same. The join records it in `group_members.invited_by`.

The referrer is validated at join time: it must be an existing member of that group and must not
be the joining user. An invalid referrer is dropped rather than refused — the same reasoning as
`join_source`, where losing an acquisition over an attribution field is the worse trade.

**Trade-off accepted:** a publicly shared card exposes the sharer's user id. It is a random
UUID that grants nothing without authentication, and it is already visible to fellow group
members through the members API. The alternative — a per-user-per-group referral token — adds a
table and a lookup to protect an identifier that is not secret.

### Achievement grants are a table in code, not a column in the database

`grantForAchievement(type) int` maps achievement types to amounts. Achievement types are a Go
enum already; putting the amounts beside them keeps the payout visible next to the rule that
earns it, and a change to an amount is a code review rather than a data migration.

Only `STREAK_VOTER` and `RECRUITER` pay. Paying for every achievement would make the currency
meaningless and would reward `BLIND` (accuracy under 20%), which is a joke badge.

### Amounts

| Grant | Amount | Reason |
|---|---|---|
| Welcome | 10 | Exactly one detector — a taste, not a stockpile |
| Referral | 5 | Two referrals buy a detector, so inviting is a real path to the hook |
| `STREAK_VOTER` milestone | 5 | Paid at the 5/10/20 milestones the achievement already uses |
| `RECRUITER` | 10 | A whole detector, because bringing three people in is the behaviour the product most wants |

The welcome grant deliberately equals one detector rather than two: the goal is to create the
want, not to satisfy it.

## Risks / Trade-offs

- **Free crystals cannibalise sales** → the welcome grant is one detector in a product with a
  weekly Reveal; a player who wants the hook every week needs more than the grant provides. The
  prior state — where the hook was never experienced — is the larger revenue problem.
- **Referral farming** → the reward requires a completed session in a group that can actually
  reveal; the participation floors from `season-cold-start` make a fake group unable to produce
  the Reveal that makes the farming worthwhile.
- **A shared card exposes the sharer's user id** → accepted, with the reasoning above.
- **A grant failure is invisible to the user** → grants are logged and idempotent, so a missed
  grant can be re-driven; failing the user's vote or registration over a payout would be worse.

## Migration Plan

1. Migration `007_referrals`: `group_members.invited_by TEXT REFERENCES users(id)` (nullable)
   with an index for the referral lookup.
2. `make sqlc`.
3. Deploy. Existing users receive no welcome grant (see Non-goals); new registrations do.
4. Rollback: drop the column. Grants already written stay valid — they are ordinary
   `crystal_logs` rows.
