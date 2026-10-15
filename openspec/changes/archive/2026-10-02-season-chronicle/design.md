# Design

## Context

See proposal.md → Why. The relevant current state:

- `GetTopResultPerQuestion(season_id)` already produces exactly a chronicle entry's shape — one row per
  question with the leading member and their percentage — for a *single* season. It is what
  `group_summary.top_per_question` returns at reveal.
- `season_results` stores `percentage`, `vote_count` and `total_voters` per (target, question). Nothing
  the chronicle shows needs computing; it needs selecting across seasons instead of one.
- `user_group_stats.guess_accuracy` is a 5-season rolling average, recomputed for every member at every
  reveal (`internal/service/achievements/service.go`), alongside `seasons_played`.
- `internal/eligibility` owns `MinMembers = 3` and `MinDetectorMembers = 5`, and the groups API already
  reports `effective_member_count` — membership minus blocks in either direction — which the anonymity
  warning is judged against.

## Goals / Non-Goals

**Goals**

- One query per chronicle page, not one per season: a group that has played for a year has ~50 seasons.
- The chronicle reads the same numbers the reveal showed, so the record and the live view cannot
  disagree.
- The growth copy derives from `eligibility`, so a threshold change cannot leave the copy behind.

**Non-Goals**

- No denormalised chronicle table. See the decision below.
- No pagination in this change. The cap is a limit, not a cursor; a group with 50 seasons gets the most
  recent N. Adding a cursor later does not change the specs.

## Decisions

### The chronicle is computed, not stored

A `group_chronicle` table would be the obvious move and is the wrong one. Every number it would hold is
already in `season_results`, which is immutable once a season reveals — so a second copy could only
ever drift, and it would have to be backfilled for every group that has already played. The chronicle
is a different `ORDER BY` over data the product already keeps.

The cost is a wider query rather than a point lookup. `DISTINCT ON (season_id, question_id)` over a
group's revealed seasons is indexed by `season_results(season_id)` and bounded by a season cap, so the
work is proportional to what is displayed.

Alternative considered: materialise at reveal, in the same transaction that writes `season_results`.
Rejected because it buys nothing — the read is already cheap — in exchange for a table that can be
wrong.

### "Revealed" means REVEALED *or* CLOSED

`seasons.status` goes VOTING → REVEALED → CLOSED, and the last step happens when the *next* season
opens (`groups.createSeasonForGroup` closes every REVEALED season for the group). So a group has at
most one REVEALED season at any moment, and every older season that legitimately revealed is CLOSED.

Any history query filtering on `status = 'REVEALED'` therefore returns one season, no matter how long
the group has played — which looks correct, because it returns rows. The existing profile queries
(`GetUserSeasonHistory`, `GetTopAttributeAllTime`) have this defect today: the profile's "season
history" and its "all-time top attribute" have only ever covered the current week. They are corrected
here rather than left, because the chronicle rests on the same definition and shipping one next to the
other would make the two surfaces disagree about the group's past.

### One query returns the whole page, joined on seasons

`GetGroupChronicle(group_id, limit)` returns the flat (season, question, winner) rows for the last N
revealed seasons, and the service groups them in Go. The alternative — looping
`GetTopResultPerQuestion` per season — is N+1 and would be the second N+1 fixed in this codebase.

`total_voters` comes along on every row, because that is what decides the anonymity marking (see the
spec delta): a share of two voters identifies them whatever the group's size was, and the group's
historic membership is not recoverable once members leave, since `group_members` records `joined_at`
and no departure.

### The standing is read, never recomputed

`user_group_stats` already holds each member's accuracy and `seasons_played`. The standing is
`SELECT … JOIN group_members … ORDER BY guess_accuracy DESC`, with the stats row absent for a member
who has not yet been through a reveal — which is how "not yet ranked" is expressed, rather than as an
accuracy of zero. Recomputing accuracy at read time would produce a different number from the one the
profile shows, since the stored value is a rolling average over a window.

The standing is withheld below two revealed seasons. One season's accuracy is a coin flip over that
season's questions, and a leaderboard built on it says something untrue about the group. Two is the
smallest number where the rolling average has actually averaged anything.

### Growth copy is generated from the thresholds

`eligibility` gains a function returning the next threshold a group has not crossed, along with what it
changes, and `nil` above the last one. It takes **two** counts, because the two thresholds are judged
against different things and conflating them puts two contradictory statements on one screen: the
Reveal is a group-level event that `Evaluate` gates on actual membership, while the detector and the
anonymity warning are judged on the members the reader can be rated by. A single effective count would
have told a reader who blocked someone that the Reveal could not happen while it was about to.

The block count itself is scoped to the group (`CountBlockedGroupMembers`). `ListBlockedUserIDs` is
global, so subtracting its length — which both the anonymity count and `detectorTooSmall` did — shrank
one group because of a block made in another. The copy lives in the mobile layer; the *number* and *which
consequence* come from Go. This is the same arrangement `RevealState` already uses, and for the same
reason: the worker's behaviour and the app's explanation must not drift.

The thresholds that actually exist are 3 (a season can reveal at all) and 5 (the detector becomes
purchasable and percentages stop identifying voters). The quorum share rising from 40% to 50% at 8
members is deliberately **not** a threshold here — it makes revealing harder, and presenting it as an
unlock would be a lie in the user's own interest. Above 5 the honest answer is "big enough", which is
what the spec requires the app to say.

### The standing publishes a position, not the figure behind it

The query reads `guess_accuracy`, but the response does not carry it. The chronicle is what turns a
published accuracy into an inference channel: it names which member won each question, so a member at
either extreme of the scale has their vote on each of those questions recovered exactly. And because the
stored figure is a rolling average weighted by `seasons_played`, two consecutive weeks can be solved for
that week's match count — `seasonAccuracy = new*(w+1) - old*w` — which a response listing every member's
figure makes cheap across the whole group.

A rank cannot be reduced to a vote, and ranking is the whole feature. `seasons_played` stays, because
without it a place is not interpretable.

Note that `GET /groups/:id/members/:userId/profile` already exposes `guess_accuracy` and
`seasons_played` for any member to any other member, so the figure is reachable today one member at a
time. That is a pre-existing exposure, not one this change introduces; it is left alone here rather than
quietly widened.

### Blocks filter the chronicle, not the standing

A blocked member's chronicle entries are withheld from the blocker, matching how `GetMembersCards`
already hides their card. The standing, by contrast, is filtered the same way — a leaderboard that
lists someone you blocked is the thing a block is supposed to prevent — but the *ranks* are computed
before filtering, so removing a row does not silently promote everyone below it into a different
position than other members see.

## Risks / Trade-offs

- **A group with a year of history makes a big response** → capped at a fixed number of seasons, newest
  first; the cap is a constant next to the query, and pagination is a later change that does not touch
  the specs.
- **The standing could become a shaming device** ("you know us worst") → it ranks accuracy, not
  popularity, and the bottom of the list is the person who guesses badly rather than the person nobody
  likes. The number of seasons is shown next to every entry so a single bad week is visible as such.
- **Accuracy is a rolling average, so the standing moves for reasons a member cannot see** → the
  seasons count is shown alongside it, and the profile already exposes the same number, so the two
  surfaces agree.
- **Growth copy could read as nagging** → it is one line, it names a consequence rather than a target,
  and it disappears entirely once the group is big enough.

## Migration Plan

No schema change, so nothing to migrate and nothing to roll back. New endpoints are additive: an older
client that does not call them is unaffected, and a newer client against an older server gets a 404,
which the app treats as "no chronicle" rather than an error.

## Open Questions

None that affect the specs or the task breakdown. The season cap's exact value and the copy's final
wording are both adjustable without changing any requirement.
