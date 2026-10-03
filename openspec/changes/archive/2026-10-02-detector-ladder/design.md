# Design

## Context

See proposal.md — Why. Today `BuyDetector` deducts 10 crystals in a transaction that locks the user
row, writes a `detectors` row, and returns `GetVoterProfilesBySeason`. `GetDetector` reports
`purchased`, `voters`, `available` (the small-group gate from `season-cold-start`) and the balance.

## Goals / Non-Goals

**Goals:**

- Create the want before selling the answer: the count is free, the answer is the top rung.
- More transactions per season without raising any single price.
- Keep every anonymity guarantee the single-purchase detector had.

**Non-Goals:**

- New prices for crystal packages, or a subscription.
- Partial reveals of anything except the voter list.

## Decisions

### Three rungs, priced 0 / 3 / 10

| Rung | Price | Reveals |
|---|---|---|
| Count | 0 | "4 человека проголосовали про тебя" |
| Hint | 3 💎 | One voter's avatar + first character of their name |
| Full list | 10 💎 | Every voter (unchanged) |

3 for a hint is deliberately below the 5-crystal referral grant: a player who invited one friend can
afford a hint, which is the whole point of making the first rung reachable without a card. The full
list stays at 10 so the existing price and the existing purchase are untouched.

The count is free rather than 1 crystal because its job is to *create* the question. Charging for the
question is how the previous design ended up selling only the answer.

### Hints are drawn without replacement, recorded per (user, season, voter)

`detector_hints(id, user_id, season_id, revealed_user_id, created_at)` with
`UNIQUE(user_id, season_id, revealed_user_id)`. Selection excludes voters already revealed to this
member, and the row is written in the same transaction that spends the crystals — so a crash between
the two cannot charge for nothing.

Alternatives considered: a counter of hints bought plus a deterministic ordering (rejected — the
ordering would have to be stable across requests, which means deriving it from voter ids, which makes
the *next* hint predictable from the previous one). Recording what was revealed is both simpler and
not guessable.

### A hint reveals avatar + first character, by construction rather than by trimming

The hint DTO has no `username` field at all. It carries `avatar_emoji`, `avatar_url` and
`first_letter` — so there is no code path where a full name could leak through a hint, and no future
change to a shared DTO can accidentally add one.

First *character*, not first letter: usernames here can start with a digit or a Cyrillic letter, and
"first letter" would need a definition that excludes them. The first rune is what a player reads.

### Selection is random, and that is deliberate

The revealed voter is chosen at random from the unrevealed ones. Revealing in a fixed order would let
a player who buys hints in two different groups infer the ordering rule, and ordering by anything
meaningful (join date, vote time) would leak a second fact beyond identity.

### The small-group floor applies to the hint too

A hint drawn from two or three possible voters identifies someone as surely as a list. The existing
`detectorTooSmall` check is applied to the hint endpoint as well, before the transaction opens. The
free count stays available at any size, because a count names nobody.

### The ladder's state rides on the existing detector response

`GetDetector` gains `hints` (what has been revealed), `hint_price`, `full_price`, and
`hint_available`. Older clients read `purchased` and `voters` and keep working, which is why this is
additive rather than a new endpoint — the Reveal screen fetches the detector already, and a second
round trip on the screen that most needs to feel instant is not worth the tidier shape.

## Risks / Trade-offs

- **Three rungs make the sheet more complex** → the free rung is the only thing shown until the
  member acts, so the default state is simpler than today's blurred list plus a price.
- **A player may stop at hints and never buy the list** → that is a sale that previously did not
  exist; the hint is priced so two hints still cost less than the list, which keeps the list the
  better deal for someone who wants certainty.
- **Hint rows grow with usage** → one row per revealed voter per season, bounded by group size.
- **Random selection can reveal the least interesting voter first** → accepted; any ordering that
  made it more satisfying would also make it informative, which is the thing being protected.

## Migration Plan

1. Migration `008_detector_hints`: the table plus its unique constraint and a lookup index.
2. `make sqlc`.
3. Deploy. Existing `detectors` rows keep meaning "bought the full list"; a member who already owns
   the list sees a completed ladder.
4. Rollback: drop the table. The full-list purchase path is unchanged, so the old binary works
   against the new schema and the new binary works without hints having ever been bought.
