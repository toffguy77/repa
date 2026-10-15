# Proposal

## Why

The push schedule promises things the app cannot show. Tuesday sends *«Кто-то уже ответил на
вопросы про тебя 👀»* (`backend/internal/worker/tasks/push.go`) and the tap lands on the group
screen, where no such counter exists. Thursday sends a category teaser; there is no teaser screen.
The reveal screen's waiting state was a static line of text until `season-cold-start` gave it the
pending-Reveal explanation — but it still has nothing to say about *you*.

The PRD specifies all of this: RET-04/05 (a counter of votes about you, number only), RET-09/10 (a
category teaser), RET-12 (a staged Reveal). None of it exists. A push that promises intrigue and
delivers a progress bar does not merely fail to retain — it teaches people to ignore the next one,
which costs the pushes that *do* work.

Underneath that is the structural problem: the loop pays out once a week. Gas closed its loop in
minutes, which is how it got dozens of sessions a day out of the same mechanic. Repa's weekly Reveal
is the brand and should stay, but it cannot be the *only* feedback a player ever receives.

## What Changes

- A member can see, at any point during the week, **how many people have answered questions about
  them** — a number, never a name, never an attribute.
- When someone votes about a member, that member gets an immediate, non-attributed signal: a count
  went up. The *what* stays sealed until Friday.
- The leading category of a member's results is exposed as **an emoji only**, from Thursday onward —
  the teaser the Thursday push has been promising.
- The waiting screen becomes the anticipation screen: the count, the teaser when available, and a
  live countdown.
- **BREAKING** (push behaviour): the Tuesday and Thursday pushes are only sent to members for whom
  the screen will actually have something to show.

## Capabilities

### New Capabilities

- `anticipation`: what a member may learn about votes concerning them *before* the Reveal, and the
  hard limits on it.

### Modified Capabilities

- `push-notifications`: the mid-week pushes may only promise what the app can show.

## Impact

- **Backend:** a new anticipation endpoint and service, a leading-category query, push-task gating.
- **No schema changes** — votes and season questions already carry everything needed.
- **API:** `GET /api/v1/seasons/:seasonId/anticipation`.
- **Mobile:** the reveal waiting state becomes the anticipation screen; push routing points at it.
- **Docs:** `docs/features/reveal.md`, `docs/features/push.md`, `docs/features/voting.md`.

## Non-goals

- Revealing *who* voted, or what they answered, before the Reveal. The detector stays a
  post-Reveal purchase and the anonymity rules are unchanged.
- Instant per-vote attribution of the kind Gas used ("someone said you have the best smile") — that
  trades the Friday ritual for a drip, and the ritual is the product's identity.
- A multi-screen staged Reveal (PRD RET-12/13). That is a Reveal-flow change, not an anticipation
  one, and it belongs with the chronicle work.
- Changing the weekly cadence, the quorum, or the push cap.
