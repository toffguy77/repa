# Design

## Context

See proposal.md — Why. What exists: `questions(text, category, source, group_id, author_id, status)`
with 205 seeded rows in six categories; `GetRandomSystemQuestionsByCategories` picks a season's
questions filtered by the group's categories and excluding the last three seasons' questions;
`questions.Service` runs the Anthropic moderator on a user-submitted question and sets `ACTIVE`,
`REJECTED` or leaves `PENDING`.

## Goals / Non-Goals

**Goals:**

- Let a group choose the kind of thing it will be asked, instead of shipping one bank to a class of
  14-year-olds and a flat share of 21-year-olds.
- Make the card lead with something a person would want to share, without making it untruthful.

**Non-Goals:**

- Deleting the edgy questions, or rewriting the bank's copy.
- Per-member banks — voting is group-wide, and different members answering different questions would
  break aggregation.

## Decisions

### Three tones, as an enum on the question

`question_tone`: `WARM`, `NEUTRAL`, `EDGY`. Three rather than two because the interesting distinction
is not "nice / not nice": most of the bank is neither a compliment nor an insult («Кто чаще всех
опаздывает?»), and collapsing that into either bucket would either empty a kind-only group's bank or
let real jabs into it.

Default `NEUTRAL` at the column level, so a question can never exist without a tone and an unclassified
one is usable rather than lost.

### Seed classification is by curated keyword sets, reviewed once

The 205 seeded questions are classified by matching against curated phrase lists, then the result is
pinned by a test that asserts the distribution and spot-checks specific questions. This is a one-off
data problem: an LLM pass would be unreproducible in CI, and hand-labelling 205 rows in a migration
makes the migration unreviewable.

The important property is not that every label is perfect — it is that `kind_only` groups cannot
receive anything from the edgy list, and that list is explicit and auditable in one place.

### `groups.kind_only`, defaulting from the creator's age

A boolean on the group, not a per-season setting: the bank a group wants is a property of the group.
Defaulting from the creator's `birth_year` reuses the same signal the ROMANCE restriction already uses,
so there is one notion of "under 18" in the product rather than two.

An explicit value in the request always wins — including an adult turning it on and a minor turning it
off. The default is a sensible starting point, not a restriction: the product already lets an under-18
user choose their categories, and a restriction they cannot see or change would be worse than the
current state.

### Changing the setting does not rewrite an open season

Season questions are materialised in `season_questions` at creation. Rewriting an open season would
change the questions out from under members who already answered some of them, and their existing votes
would reference questions no longer in the season. The setting applies from the next season.

### An all-edgy category set is refused rather than allowed to empty the season

HOT and SECRETS exist in order to provoke, so every question in them is edgy. A kind-only group
restricted to those categories would draw zero questions and could not vote at all — and the failure
would be silent, appearing as an empty season rather than as an answer to anything the member chose.
So the combination is refused at group creation and when the setting is turned on, with an error that
names both the setting and the categories, since those are the two things the member can change.

The check asks the bank (`CountKindQuestionsByCategories`) rather than comparing against a hardcoded
list of "kind categories". What makes a combination unusable is the absence of questions, which is a
property of the bank's contents: adding one gentle SECRETS question should make SECRETS legal for a
kind-only group without a code change, and removing the last kind FUNNY question should make it illegal.

### Card ordering: tone as a tiebreaker, not an override

Attributes are ordered by percentage, with tone breaking near-ties — a warm attribute wins over an edgy
one when they are within a small band of each other. Tone never *reorders past* a clearly stronger
result, because the card's value is that it is true: a person who overwhelmingly got one edgy attribute
should see it, and a card that hid it would be a different product.

The band is 5 percentage points: wide enough to break the ties that actually occur in groups of 5–50
(where one vote is 2–20 points), narrow enough that it cannot flip a real difference.

The rule lives in its own package, `internal/cardorder`, rather than inside the reveal service. Four
places build a card from the same results and must agree: the reveal API, the members-cards list, the
purchased "show everything" view, and the worker that renders the shared PNG. The PNG is the one the
requirement is actually about — it is the artifact a person is invited to send — and it is produced
furthest from the API, so a rule living in the reveal service would have been silently skipped there.
The failure mode is quiet: an unset tone reads as neutral for every row, which is indistinguishable
from plain percentage order and fails nothing.

The band cannot be implemented as a comparison function, because "within 5 points" is not transitive:
50% edgy, 47% neutral and 44% warm would give warm < neutral < edgy < warm, and a sort over an
inconsistent comparator produces an undefined order that can move the 44% warm attribute above the 50%
edgy one — exactly the reordering-past-a-stronger-result this decision forbids. Instead the list is
sorted by percentage and then walked into *anchor clusters*: a cluster holds the attributes within 5
points of the cluster's own first (strongest) attribute, and tone orders within a cluster only. The
anchor is fixed, so no attribute can be promoted past one more than 5 points stronger, however long
the chain of near-ties.

### Filtering happens in the selection query

`GetRandomSystemQuestionsByCategories` gains a tone filter, rather than selecting then filtering in Go.
Filtering afterwards would under-fill a season: the query already limits to `min(10, members*2)`, so
dropping rows after the limit would hand a kind-only group fewer questions than an ordinary one.

## Risks / Trade-offs

- **The keyword classifier mislabels some questions** → the risk that matters is an edgy question
  reaching a kind-only group, so the edgy list is the one that must be complete rather than precise; a
  warm question labelled neutral costs nothing.
- **A kind-only group has a smaller bank** → ~2/3 of the bank remains, which at 5–10 questions a week
  is many months before rotation pressure; the rotation window is 3 seasons.
- **The refusal blocks a combination a member deliberately chose** → it is the only alternative to an
  empty season. The client does not pre-empt it by disabling categories: which categories are all
  edgy is a property of the bank's contents, and duplicating that judgement in the client would be a
  second source of truth that silently goes stale. Instead the client shows the server's explanation
  next to the switch and the category chips, both of which the member can change on the spot.
- **Tone-aware ordering could be seen as hiding results** → nothing is removed, the band is 5 points,
  and the hidden-attributes purchase still shows everything.
- **`kind_only` defaults from self-reported age** → the same limitation the ROMANCE restriction already
  has; this change does not make it worse.

## Migration Plan

1. Migration `010_question_tone`: `question_tone` enum, `questions.tone` defaulting to `NEUTRAL`,
   `groups.kind_only` defaulting to false, and an UPDATE classifying the seeded rows.
2. `make sqlc`.
3. Re-run `make seed` is **not** required: the migration classifies the rows already present, and the
   seeder is updated so a fresh database arrives classified.
4. Rollback: drop both columns and the enum. Selection falls back to category-only, which is the
   current behaviour.
