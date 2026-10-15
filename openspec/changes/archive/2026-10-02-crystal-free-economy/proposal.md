# Proposal

## Why

`crystal_log_type` has included a `BONUS` value since the first migration
(`backend/internal/db/migrations/001_init.up.sql`), and **no production code path ever writes
it** — the only Go code that uses `BONUS` is a test helper. Every crystal in the product comes
from a YuKassa purchase.

So the detector — the single most compelling hook in the game, "кто за меня голосовал" — is
reachable only by paying with a card. The audience is 14–22; the younger half does not have
one. The hook is therefore never *tasted*, which removes both the motive to pay later and the
motive to bring anyone in: the `RECRUITER` achievement already exists and awards nothing.

The previous change made arrivals attributable. This change gives those arrivals a value.

## What Changes

- Every new user receives a one-time welcome grant large enough for exactly one detector, so
  the core hook is tried before it is sold.
- Inviting someone pays: when a member a user brought in **completes their first vote**, the
  inviter is granted crystals. Rewarding the completed vote rather than the join is what keeps
  it from being an account-farming payout.
- A shared card carries who shared it, so the reward reaches the person who actually invited
  rather than the group's admin.
- Streak and recruiter achievements carry crystal grants, so the retention loop and the
  acquisition loop both pay out in the same currency.
- Every grant is idempotent: a grant can be attempted repeatedly and pays once.

## Capabilities

### New Capabilities

- `crystal-grants`: how a user obtains crystals without paying — what is granted, what triggers
  it, and the guarantee that a grant is never paid twice.

### Modified Capabilities

- `share-attribution`: a shared card additionally identifies who shared it, and a membership
  records who brought that member in.

## Impact

- **Schema:** `group_members.invited_by`; a referral index. `crystal_logs.external_id` (already
  UNIQUE) becomes the idempotency key for grants.
- **Backend:** a new grants service; wiring in auth (welcome), voting (referral trigger),
  achievements (milestone grants); `cards` and `groups` carry the referrer.
- **API:** `POST /groups/join/:code` accepts a referrer; the crystal history shows grants with
  readable reasons.
- **Mobile:** the share link carries the sharer; the shop and balance surfaces explain where
  free crystals came from.
- **Docs:** `docs/features/crystals.md`, `docs/features/groups.md`, `docs/features/cards.md`,
  `docs/features/achievements.md`.

## Non-goals

- Changing crystal prices, package contents, or what crystals buy — `detector-ladder` owns the
  spending side.
- Earning crystals by watching ads or completing surveys.
- Transferring or gifting crystals between users.
- Retroactive grants to users who registered before this change; the welcome grant is for new
  registrations, and backfilling would mean paying out for sessions that already happened.
