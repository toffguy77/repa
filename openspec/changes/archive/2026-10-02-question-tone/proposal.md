# Proposal

## Why

The question bank has a category but no notion of *tone*. 205 seeded questions span 🔥 Горячие and
🤫 Секреты alongside 😂 Смешные, with items like «Кто первым побежит при пожаре?» — mockery, not a
compliment. Nothing in the schema distinguishes "this is a nice thing to receive" from "this is a
thing you would rather not be told in front of your class."

Two consequences:

- **Shareability.** The shared card is the product's only organic acquisition channel, and people
  share flattering things and hide unflattering ones. A card whose top attribute was chosen purely by
  vote count is as likely to be an insult as a compliment — and an insult does not get posted to
  Stories.
- **Store risk.** tbh was strictly positive and survived App Store review and the press; Gas aimed at
  compliments and still drew a wave of harassment reports. The PRD's bullying mitigation is
  "AI-модерация вопросов" — which judges whether a question is *allowed*, not whether a group wanted
  that kind of question at all.

A 14-year-old's class group and a 21-year-old's flat share do not want the same bank, and today they
get the same one.

## What Changes

- Every question carries a **tone**: warm, neutral, or edgy. Seeded questions are classified; new
  user questions default to neutral and are classified by the existing moderation step.
- A group can be set to **kind questions only**, which restricts its seasons to warm and neutral
  questions. School-age groups get this on by default.
- The card's headline attribute is chosen with tone as a tiebreaker, so a warm result wins over an
  edgy one at comparable vote counts — the card leads with something a person would want to share.
- **BREAKING** (behaviour): season question selection now filters by the group's tone setting, so a
  kind-only group will never receive an edgy question.
- The group creation screen states what the setting does, rather than burying it.

## Capabilities

### New Capabilities

- `question-tone`: what tone a question has, how a group constrains it, and how tone affects what a
  card leads with.

### Modified Capabilities

- `season-lifecycle`: season question selection respects the group's tone setting.

## Impact

- **Schema:** `questions.tone`, `groups.kind_only`; the seed gains tone values.
- **Backend:** question selection, the moderation step's tone classification, card headline ordering.
- **API:** group creation and update accept `kind_only`; group responses expose it.
- **Mobile:** a tone switch on group creation and in group settings, with its consequence stated.
- **Docs:** `docs/features/groups.md`, `docs/features/moderation.md`, `docs/features/cards.md`, and a
  new section in `docs/features/voting.md`.

## Non-goals

- Removing the edgy questions. A 21-year-old flat share wanting 🔥 Горячие is a legitimate use; the
  point is that the group chooses.
- Per-member tone preferences. Voting is group-wide, so a per-member bank would mean members answering
  different questions about each other, which breaks aggregation.
- Re-writing the question bank's copy, or changing the categories.
- Changing the AI moderation verdicts for *allowed* versus *rejected* — tone is orthogonal to that.
