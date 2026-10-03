# Design

## Context

See proposal.md → Why. `GetProfile(ctx, groupID, userID, requesterID)` already receives both the viewed
member and the viewer, so the decision needs no new plumbing — it was simply never made.

`stats.guess_accuracy` is a `float64` on `StatsDto`, populated from `user_group_stats` when a stats row
exists and left at Go's zero value when it does not. That zero already conflates two different facts
("no revealed season yet" and "matched nothing"), and withholding the field would add a third.

## Goals / Non-Goals

**Goals**

- A member's match rate reaches only that member.
- Withheld is distinguishable from zero on the wire.
- The decision is made once, in the service, not re-derived in the handler or the client.

**Non-Goals**

- Not changing how accuracy is computed or stored. The rolling average stays; what changes is who may
  read it.
- Not touching accuracy-derived achievements. An achievement records a threshold crossed, so it carries
  at most one bit about a season rather than a figure, and it cannot be differenced across weeks.
- Not auditing every other statistic in the product. This change states the rule in the spec; applying
  it to statistics that do not exist yet is the job of whoever adds them.

## Decisions

### A pointer, so the field can actually be absent

`GuessAccuracy` becomes `*float64` with `json:"guess_accuracy,omitempty"`. Alternatives considered:

- **Zero it out.** Rejected: `0` is indistinguishable from a member who genuinely matched nothing, so a
  client would display a lie rather than nothing. It also silently defeats the purpose the moment someone
  adds a "worst predictor" view.
- **A separate self-profile endpoint.** Rejected: two endpoints returning overlapping shapes is how they
  drift, and the viewer identity is already a parameter of the one that exists.
- **Coarsen into buckets** ("часто / иногда / редко"). Rejected as insufficient rather than wrong: the
  figure is differenced across weeks, and bucket *transitions* still leak — a jump from "редко" to "часто"
  bounds that week's match count. Coarsening also keeps a measurement where an order already does the job
  (the chronicle standing), so it would add a second comparison surface with weaker properties.

### The service decides, not the handler

The omission happens where `requesterID` is already compared against `userID`, in `GetProfile`. Putting it
in the handler would mean a future caller of the service — a worker, an admin view, a second endpoint —
gets the figure by default, which is the wrong direction for a privacy decision to fail in.

### Why the other statistics stay

The attack needs a count of *matches* to combine with the published winners. `seasons_played`,
`voting_streak`, `max_voting_streak`, `total_votes_cast` and `total_votes_received` count participation,
not agreement: knowing a member cast 40 votes says nothing about which targets they chose. They are left
visible rather than removed on suspicion, because stripping a profile of everything is its own cost and
the rule being applied is specific.

Worth noting what `seasons_played` *does* do: it is the weight `w` in the rolling average. It is harmless
alone and useful for interpreting a rank, but it is only harmless because the figure it weighted is now
gone. If the figure is ever reintroduced in any form, this pairing is the first thing to re-examine.

## Risks / Trade-offs

- **A member loses a number they could previously see about a friend** → the comparison that number was
  used for is the chronicle standing, which is ordinal and public. Nothing a member could reasonably want
  from the figure is lost.
- **An older client crashes on the missing field** → it would have, since it reads a non-nullable double;
  the Freezed model is made nullable in this change. There is no released client yet, so no deployed
  version is affected.
- **The rule is easy to break again** → it is recorded as a requirement in `small-group-anonymity` and as
  a convention in `CLAUDE.md`, framed as "prefer an order over a measurement" so it generalises to the
  next statistic rather than naming this one field.

## Migration Plan

No schema change and no stored value changes, so nothing to migrate. The change is a narrowing of a
response: deploying the backend before the client leaves the client reading a missing field, which is why
the model becomes nullable in the same change.

## Open Questions

None.
