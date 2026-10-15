# Proposal

## Why

The detector is a single purchase that answers the question completely: 10 crystals, and the full
list of voters appears (`GET/POST /api/v1/seasons/:seasonId/detector`). One shot, and the curiosity
that drove it is gone — there is nothing left to buy for that season, and nothing to want.

Gas monetised the same curiosity for roughly ten times as much by **not** answering it: "God Mode"
showed the *first letter* of a name. The tease is what sells; the answer ends the transaction. Repa
currently sells the answer.

The PRD targets ARPU above 180 ₽/month from detectors, while the cheapest package is 59 ₽ for one
detector. At one purchase per season that target is arithmetically out of reach.

## What Changes

- The detector becomes a ladder of three steps instead of one purchase:
  1. **Free** — how many people voted about you. No payment, no names.
  2. **A hint** — for a small amount, one voter is partially revealed: their avatar and the first
     letter of their name, not their identity.
  3. **The list** — the current full voter list, at the current price.
- Each step is bought independently and remembered per season, so a player can stop at any rung.
- **BREAKING** (API): the detector response gains the ladder's state; clients that only read
  `purchased` and `voters` keep working.
- Hints are drawn without replacement, so paying twice reveals two different people rather than the
  same one.
- The free rung is visible without any purchase, so the curiosity is created before anything is
  sold.

## Capabilities

### New Capabilities

- `detector-ladder`: the rungs of the detector, what each one reveals, and the anonymity limits on
  the partial ones.

### Modified Capabilities

- `small-group-anonymity`: the group-size floor applies to every rung of the ladder, not only the
  full list.

## Impact

- **Schema:** a `detector_hints` table recording which voter each hint revealed, per user and season.
- **Backend:** `internal/service/reveal/` (the ladder, hint selection, pricing),
  `internal/handler/reveal/` (the new step endpoint and the extended response).
- **API:** `GET /seasons/:id/detector` reports the ladder; `POST /seasons/:id/detector/hint` buys a
  hint; the existing `POST /seasons/:id/detector` keeps buying the full list.
- **Mobile:** the detector sheet becomes a ladder with the free count always visible.
- **Docs:** `docs/features/reveal.md`, `docs/features/crystals.md`.

## Non-goals

- Changing crystal prices or package contents.
- A subscription tier.
- Revealing *what* a voter answered — the anonymity rule that the detector never binds a voter to a
  question or an answer is unchanged and applies to every rung.
- Hints on anything other than the voter list (no partial attribute reveals).
