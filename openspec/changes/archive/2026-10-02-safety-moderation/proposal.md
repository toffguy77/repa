# Proposal

## Why

Repa is an anonymous peer-rating app for 14–22-year-olds, and it currently has no way for a member
to get away from another member. The only report in the product is on a *question*
(`POST /groups/:groupId/questions/:questionId/report`); there is no report on a person, no block, and
no way for a group admin to remove anyone — the groups API has `leave` and nothing else
(`docs/features/groups.md`). A member being targeted can only leave the group, and nothing stops them
being invited straight back.

Two consequences. The product one: the PRD's own risk table lists bullying as medium-probability and
high-impact, mitigated by "AI-модерация вопросов, жалобы, жёсткий ToS" — two of which do not exist for
people. The business one: App Store Guideline 1.2 requires a method for filtering objectionable
content, a mechanism to report it, **and the ability to block abusive users**. Anonymous-feedback apps
are rejected under it routinely, and NGL was fined $5M by the FTC over this category of product.

## What Changes

- A member can **block** another member. A block is mutual in effect: neither can be shown the other
  as a voting target, and neither sees the other's card.
- A member can **report** another member, with a reason, to the same admin queue that question
  reports already use.
- A group admin can **remove** a member, and a removed member cannot re-join that group with the same
  invite.
- A member can **leave permanently** — the same irreversibility, chosen by themselves.
- Blocked pairs are excluded from voting targets, so the product never asks someone to rate a person
  they blocked.
- **BREAKING** (behaviour): the voting target list is no longer "every member except me" — it is
  every member except me, those I blocked, and those who blocked me.

## Capabilities

### New Capabilities

- `member-safety`: what a member can do to get away from another member, and what the system
  guarantees once they have.

### Modified Capabilities

- `group-invites`: a removed or permanently-departed member cannot re-join with an invite.
- `small-group-anonymity`: blocks reduce the effective group size, and the anonymity floor must count
  what is actually voteable.

## Impact

- **Schema:** a `blocks` table, a `group_bans` table, and `user_reports` rows in the existing report
  queue.
- **Backend:** a new safety service; voting target filtering; join-path ban checks; admin queue.
- **API:** block/unblock, report a member, admin remove, leave permanently; the voting session's
  target list changes.
- **Mobile:** block/report actions on a member, admin remove in the group screen, confirmations.
- **Docs:** `docs/features/groups.md`, `docs/features/voting.md`, `docs/features/moderation.md`,
  `docs/features/reveal.md`.

## Non-goals

- Automated detection of abusive *voting patterns*. Detecting "this person is being piled on" from
  anonymous votes is a research problem, and acting on it would require exposing who voted.
- A full moderation console. Reports land in the existing admin queue; triage stays manual, as it
  already is for questions.
- Account-level suspension or deletion by an admin — that is an account-lifecycle change with its own
  legal surface.
- Changing the question moderation pipeline or the AI moderator.
