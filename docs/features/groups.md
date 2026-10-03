# Groups

## Overview

Users create and join private groups where they vote on each other weekly. Each group has an admin, an invite link, a set of question categories, and auto-created weekly seasons. Groups are limited to 5-50 members, and users can join up to 10 groups.

## API Endpoints

All endpoints require Bearer JWT authentication.

### `POST /api/v1/groups`
Create a new group with first season.
- **Body:** `{ "name": "...", "categories": ["HOT", "FUNNY", ...], "telegram_username?": "..." }`
- **Validation:** name 3-40 chars, categories non-empty, valid category values
- **Success 201:** `{ "data": { "group": GroupDto, "invite_url": "https://repa.app/join/{code}" } }`
- **Error 409:** `GROUP_LIMIT` — user already in 10 groups
- **Error 400:** `VALIDATION` — invalid name or categories
- **Error 403:** `ROMANCE_BLOCKED` — user under 18 tried to include ROMANCE category
- **Behavior:** Creates group, adds creator as admin + member, creates Season #1 with random questions. All wrapped in a single SQL transaction.

### `GET /api/v1/groups`
List current user's groups with active season info.
- **Success 200:** `{ "data": { "groups": [GroupListItemDto] } }`
- **GroupListItemDto:** id, name, member_count, invite_code, telegram_username, active_season (id, status, reveal_at, voted_count, total_count, user_voted, reveal_state, members_needed, voters_needed)
- `voted_count` counts members who completed **every** question, the same number the
  progress endpoint and the reveal rules use.

### `GET /api/v1/groups/:id`
Get group detail with members and active season.
- **Success 200:** `{ "data": { "group": GroupDto, "members": [MemberDto], "active_season?": { id, number, status, starts_at, reveal_at, ends_at } } }`
- **Error 403:** `NOT_MEMBER`
- **Error 404:** `NOT_FOUND`
- **MemberDto:** id, username, avatar_emoji, avatar_url, is_admin
- **Note:** The active_season fields differ from the list endpoint (`GET /groups`), which returns `id, status, reveal_at, voted_count, total_count, user_voted, reveal_state, members_needed, voters_needed`. This endpoint additionally returns `number`, `starts_at`, `ends_at`, and `kind`.

### Pending Reveal fields

Both group endpoints expose what a pending Reveal is waiting for, derived from
`internal/eligibility` — the same rules the reveal worker enforces:

| Field | Meaning |
|---|---|
| `reveal_state` | `SCHEDULED`, `WAITING_FOR_MEMBERS`, `WAITING_FOR_VOTERS`, `POSTPONED`, or `REVEALED` |
| `members_needed` | How many more members must join before a Reveal is possible (0 when satisfied) |
| `voters_needed` | How many more members must complete voting (0 when satisfied) |

A missing floor takes precedence over `POSTPONED`, because "3 more people need to vote" is
actionable and "postponed" is not.

### `GET /api/v1/groups/join/:inviteCode/preview`
Preview group before joining (no membership required).
- **Success 200:** `{ "data": { "name": "...", "member_count": 5, "admin_username": "..." } }`
- **Error 404:** `NOT_FOUND`
- **Behavior:** the code is normalised before lookup — see **Invite codes** below.

### `POST /api/v1/groups/join/:inviteCode`
Join a group by invite code.
- **Query:** `ref` (optional) — the user id of the member who shared the link, so a referral
  reward reaches them (see `docs/features/crystals.md`). Validated server-side: the referrer must
  be an existing member of that group and must not be the joining user. An invalid claim is
  **dropped**, not refused — losing an acquisition over an attribution field is the worse trade.
- **Query:** `source` (optional) — how the member arrived: `LINK`, `CODE`, `CARD`, `TELEGRAM`.
  Case-insensitive. An absent or unrecognised value is recorded as `UNKNOWN` and **never** fails
  the join: a client sending a source the server does not know is version skew, and refusing a
  join over a telemetry field would trade an acquisition for a datum.
- **Success 200:** `{ "data": { "group": GroupDto } }`
- **Error 404:** `NOT_FOUND`
- **Error 409:** `ALREADY_MEMBER`, `GROUP_LIMIT` (user), `MEMBER_LIMIT` (group)

### `DELETE /api/v1/groups/:id/leave`
Leave a group.
- **Success 200:** `{ "data": { "left": true } }`
- **Error 403:** `NOT_MEMBER`
- **Behavior:**
  - If last member → group is deleted
  - If admin leaves → admin transferred to next member by join date
  - Leave + admin transfer wrapped in a single SQL transaction

### `PATCH /api/v1/groups/:id`
Update group (admin only).
- **Body:** `{ "name?": "...", "telegram_username?": "..." }`
- **Success 200:** `{ "data": { "group": GroupDto } }`
- **Error 403:** `NOT_ADMIN`
- **Note:** `UpdateGroupTelegramUsername` is a separate query from `UpdateGroupTelegram` to avoid clearing `telegram_chat_id` on username-only updates.

### `POST /api/v1/groups/:id/invite-link`
Regenerate invite link (admin only).
- **Success 200:** `{ "data": { "invite_url": "https://repa.app/join/{newCode}" } }`
- **Error 403:** `NOT_ADMIN`

### Response DTOs

```json
// GroupDto
{
  "id": "uuid",
  "name": "string",
  "admin_id": "uuid",
  "invite_code": "uuid",
  "categories": ["HOT", "FUNNY"],
  "telegram_username": "string | null",
  "created_at": "2026-01-01T00:00:00Z"
}
```

### `GET /api/v1/groups/:id/chronicle`

The group's shared record of its past seasons, plus the guess standing. Members only (`NOT_MEMBER`,
403). See **Chronicle** below.

## Invite codes

An invite code is **6 characters** from the 31-symbol alphabet
`23456789ABCDEFGHJKMNPQRSTUVWXYZ`. The digits `0`/`1` and the letters `O`/`I`/`L` are
excluded because those are the pairs people confuse when reading a code aloud or off a screen.
31⁶ ≈ 887 million codes.

Codes are generated with `crypto/rand` (`internal/service/groups/invite.go`): a predictable
sequence would let someone enumerate newly created groups, which is the property the alphabet
choice protects.

### Accepted input

`NormalizeInviteCode` is applied in the service — not the handler — so the API, the deep link
and any future bot command cannot disagree about what a valid code is. All of these resolve
to `AB2CD3`:

| Input | Why it is accepted |
|---|---|
| `AB2CD3` | the stored form |
| `ab2cd3`, `Ab2Cd3` | lookup is case-insensitive |
| `  AB2CD3 ` | surrounding whitespace is trimmed |
| `AB2 CD3` | the grouped display the app shows |
| `AB2-CD3`, `AB2_CD3` | separators a user may add |
| `https://repa.app/join/AB2CD3` | a pasted link |
| `https://repa.app/join/AB2CD3?utm_source=tg` | a link with tracking parameters |

### Uniqueness and revocation

- Uniqueness is enforced by `UNIQUE INDEX groups_invite_code_upper_idx ON groups
  (upper(invite_code))` (migration 005), so two simultaneous creations cannot land on the same
  code.
- `freshInviteCode` probes before using a code and retries on collision, up to
  `maxInviteCodeAttempts` (10). Exhausting them returns `ErrInviteCodeExhausted` and fails the
  request — with 887M codes that signals a fault, not a full code space.
- Regenerating the invite overwrites `groups.invite_code`, which revokes the previous code
  immediately: the lookup only ever matches the current value. Admin only.

### Legacy codes

Codes issued before migration 005 are 36-character UUIDs. They keep working, because the
lookup is `upper(invite_code) = upper($1)` and no constraint validates the stored charset or
length — the format rule applies to *generation* only. Validating stored codes would mean
rewriting them, and rewriting them would break the links this behaviour exists to protect. A
group's legacy code is replaced opportunistically the next time its admin regenerates the
invite; there is no backfill.

`FormatInviteCode` groups a short code for display (`AB2 CD3`). That gap is cosmetic — the
stored and transmitted code has none, and normalisation strips any the user types back in.


## Member safety

An anonymous peer-rating app for 14–22-year-olds had no way for a member to get away from another
member: the only report was on a *question*, there was no block, and an admin could not remove anybody.
App Store Guideline 1.2 requires a filtering method, a report mechanism **and** the ability to block
abusive users; the PRD's own risk table lists bullying as medium-probability and high-impact.

### `POST /api/v1/members/:userId/block` · `DELETE /api/v1/members/:userId/block`
- **Success 200:** `{ "data": { "blocked": true|false } }`
- **Error 400:** `SELF_TARGET`
- **Behavior:** only the blocker's action is recorded, so only they can undo it — but the **effect is
  symmetric**. A one-directional block would keep asking the blocked person to rate someone who
  withdrew, and keep delivering their votes to that person, which is the harm the block exists to stop.
  Idempotent.

### `POST /api/v1/members/:userId/report`
- **Body:** `{ "group_id": "...", "reason": "..." }`
- **Success 200:** `{ "data": { "reported": true } }`
- **Error 400:** `SELF_TARGET`; **403:** `NOT_MEMBER`
- **Behavior:** reaches the admin queue (`GET /api/v1/admin/user-reports`). The reported member is
  **not** notified — telling them would make reporting an act of confrontation. One report per person
  per reporter; a second updates the reason. Independent of blocking.

### `DELETE /api/v1/groups/:id/members/:userId`
- **Success 200:** `{ "data": { "removed": true } }`
- **Error 400:** `SELF_TARGET` (removing yourself is leaving); **403:** `NOT_ADMIN`, `NOT_MEMBER`;
  **404:** `NOT_FOUND`
- **Behavior:** removes the membership **and writes a ban**, so the member cannot re-join.

### `DELETE /api/v1/groups/:id/leave?permanent=true`
- **Success 200:** `{ "data": { "left": true, "permanent": bool } }`
- **Behavior:** `permanent=true` additionally writes a ban. Ordinary leaving stays reversible on
  purpose — people leave groups by accident — so permanence is a separate, explicit choice.

### Bans

`group_bans(group_id, user_id)` is checked on join and returns `GROUP_BANNED` (403). The ban is on the
**person and the group**, not on the code, which is why regenerating the invite does not undo it. A ban
is per group: a different group with its own admin is unaffected.

A `blocks` row is separate from a ban: a block is between two members, a ban is between a member and a
group.

### Effective group size

`GET /groups/:id` returns `effective_member_count` — the membership minus anyone blocked in either
direction. **All anonymity thresholds are judged against this**, not the raw count: a group of five with
three blocks presents itself as safely anonymous while behaving like a group of two, which is precisely
the case the floor exists for.

### Mobile

- Long-pressing a member opens `MemberActionsSheet` (block, report, and — for the admin only — remove).
  The ordinary tap still opens their profile.
- Every action is behind a confirmation that **states its consequence**, including that a removal cannot
  be undone even with a new invite link.
- The leave action offers both kinds with the difference spelled out: "Можно вернуться по ссылке" versus
  "Вернуться по ссылке не получится".
- `SmallGroupNotice` uses `effective_member_count` when present, falling back to the raw member count.


## Data Model

### Tables

- **groups** — id, name, invite_code (UNIQUE), admin_id (FK users), categories (text[]), kind_only (boolean, default false — see **Question tone** below), telegram_chat_id, telegram_chat_username, telegram_connect_code, telegram_connect_expiry, created_at
- **blocks** — id, blocker_id, blocked_id, created_at. UNIQUE(blocker_id, blocked_id), CHECK they differ.
- **group_bans** — id, group_id, user_id, banned_by, reason, created_at. UNIQUE(group_id, user_id).
- **user_reports** — id, reported_id, reporter_id, group_id, reason, created_at.
  UNIQUE(reported_id, reporter_id). A separate table from `reports`, whose `question_id` is NOT NULL and
  whose every query joins through it.
- **group_members** — id, user_id (FK users), group_id (FK groups), joined_at, invited_by (FK
  users, nullable — who brought this member in), join_source
  (`join_source` enum, default `UNKNOWN`). UNIQUE(user_id, group_id). Provenance is a property of
  the membership — written once, always read alongside it — so it is a column rather than an
  event table. A founder's own membership is `UNKNOWN`, which is accurate: they did not arrive
  through an invite.
- **share_events** — id, user_id (FK users), season_id (FK seasons), channel, created_at.
  Append-only, no uniqueness: a member can share many times per season. See
  `docs/features/cards.md` → Sharing and attribution.
- **seasons** — id, group_id (FK groups), number, status (season_status enum), starts_at, reveal_at, ends_at, created_at
- **season_questions** — id, season_id (FK seasons), question_id (FK questions), ord

### Enums

- `season_status`: VOTING, REVEALED, CLOSED
- `question_category`: HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY
- `question_tone`: WARM, NEUTRAL, EDGY (migration 010 — see **Question tone** below)

### Migration 002

Added `categories text[] NOT NULL DEFAULT '{}'::text[]` column to `groups` table.

## Business Rules

- **Group limits:** max 10 groups per user, max 50 members per group
- **Admin:** creator becomes admin. Only admin can update group name, telegram, and regenerate invite link
- **Admin transfer:** when admin leaves, the next member by join date becomes admin
- **Group deletion:** when last member leaves, group is deleted (CASCADE)
- **ROMANCE restriction:** users under 18 (by birth_year) cannot include ROMANCE category when creating a group. Enforced both server-side (handler returns `ROMANCE_BLOCKED`) and client-side (category hidden in CreateGroupScreen).
- **Valid categories:** HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY

### Chronicle

The group's record of what happened: for each past season, the standout result per question — who led
it and at what share of the vote. Computed from `season_results`, never stored: every number is already
there, so a second copy could only drift and would need backfilling for groups that have already played.

- **"Past" means `REVEALED` *or* `CLOSED`.** `seasons.status` goes VOTING → REVEALED → CLOSED, and the
  last step happens when the *next* season opens (`createSeasonForGroup` closes every REVEALED season
  for the group). A group therefore has at most one REVEALED season at any moment, and every older
  season that legitimately revealed is CLOSED. A history query filtering on `REVEALED` alone returns
  one season however long the group has played — which looks correct, because it returns rows. The
  profile's own history queries had this defect and were corrected alongside the chronicle.
- **Capped** at `chronicle.MaxSeasons` seasons, newest first. The response reports `seasons_total`
  alongside `seasons_shown`, so the app can say the history is longer than what it is showing. Not
  pagination; a cursor is a later change.
- **Blocks withhold entries, not seasons.** An entry whose winner the reader has blocked is omitted;
  the season stays, even reduced to nothing. Dropping it would tell the reader that a block is in play.
- **A departed member's entry remains.** It records something that happened.
- **Small-sample marking.** `small_sample` is set when the result's percentages were computed over
  fewer than `eligibility.MinDetectorMembers` voters — the same threshold the detector is gated on,
  rather than a second number. Judged on `season_results.total_voters`, not the group's membership:
  a share of two voters identifies them whatever the group's size was, and the historic membership is
  not recoverable, since `group_members` records `joined_at` and no departure.

### Знатоки standing

Members ranked by `user_group_stats.guess_accuracy` — read, never recomputed, because the stored value
is a rolling average over a 5-season window (`internal/service/achievements`) and recomputing would show
a different number than the profile does. See `docs/features/profile.md`.

- **Withheld below `chronicle.MinSeasonsForStanding` revealed seasons**, with a stated reason. After one
  season the rolling average has averaged nothing, and a leaderboard built on it ranks luck.
- **A member with no stats row is unranked**, not ranked at zero: "has not played here yet" and "is
  always wrong" are different facts. The LEFT JOIN is what expresses this.
- **Every entry carries `seasons_played`.** An accuracy from two seasons and one from ten are not
  comparable numbers, so the basis is shown next to the figure.
- **Ranks are assigned before blocks are filtered**, so removing a row does not renumber the rest into
  positions other members do not see.
- **No vote is exposed.** The response holds members, accuracy and a seasons count — no question,
  target or vote, in any group size.

### Group-size thresholds

`next_threshold` on the group response names the next size that changes what the group can do, or is
`null` once every threshold is crossed. Derived from `eligibility.NextGrowthThreshold` so the number and
the consequence have one owner — the same arrangement `reveal_state` uses, for the same reason.

| Size | Unlock key | What changes |
| --- | --- | --- |
| `eligibility.MinMembers` (3) | `REVEAL` | Below it a season cannot reveal at all |
| `eligibility.MinDetectorMembers` (5) | `DETECTOR` | The detector becomes purchasable, and percentages stop identifying individual voters |

- Counted against `effective_member_count`, so it agrees with the anonymity warning rather than
  contradicting it: a group of five with two blocks is told it is still short.
- **Nothing above the last threshold is named.** `null` means "big enough"; the app renders no target.
- **The quorum share rising from 40% to 50% at 8 members is deliberately not a threshold.** It makes
  revealing *harder*, and presenting it as an unlock would be a lie told in the user's own interest.

### Question tone

Every question carries a tone — `WARM`, `NEUTRAL` or `EDGY` (`questions.tone`, migration 010). A group
can restrict itself to the kind ones with `groups.kind_only`, and while it is on an `EDGY` question is
never selected for that group's seasons. Classification lives in
`internal/service/questions/tone.go` (`ClassifyTone`): curated phrase lists first — edgy before warm,
since a question carrying both is one a kind-only group should not receive — then a category fallback
(SKILLS → WARM, HOT/SECRETS → EDGY, everything else → NEUTRAL). Migration 010 applies the same rules
in SQL so databases that already exist are corrected without re-seeding; `cmd/seed` applies them in Go
so a fresh database arrives classified. `TestMigrationMarkersMatchClassifier` keeps the two in step.

- **Default:** on for a creator under 18 (by `birth_year`), off otherwise. An unknown birth year counts
  as under 18, matching the ROMANCE restriction. An explicit `kind_only` in the request always wins,
  including an adult turning it on and a minor turning it off.
- **Changing it:** `PATCH /api/v1/groups/:id` with `kind_only`, admin only (`NOT_ADMIN` otherwise).
- **Open seasons are never rewritten.** The setting applies from the next season: rewriting an open one
  would change the questions out from under members who already answered some, and their votes would
  reference questions no longer in it.
- **An all-edgy category set is refused.** HOT and SECRETS exist to provoke, so every question in them
  is edgy; a kind-only group restricted to them would draw nothing. Both creation and the setting
  change return `NO_KIND_CATEGORIES` (400) when the combination would leave no question to ask. The
  check counts the bank (`CountKindQuestionsByCategories`) rather than comparing against a hardcoded
  list of categories, so the rule follows the bank's actual contents. Turning the setting *off* is
  never refused — a group that reached that state must be able to leave it.
- **Selection:** the tone filter is applied inside `GetRandomSystemQuestionsByCategories`, not after
  it. The query limits to `min(10, members*2)`, so filtering afterwards would hand a kind-only group
  fewer questions than an ordinary one.
- **Card ordering** uses tone as a near-tie breaker — see `docs/features/cards.md`.

### Season Kinds

Seasons carry a `kind` (`seasons.kind`, migration 004):

- **KICKOFF** — a new group's Season #1. Opens for voting immediately and is scheduled to
  reveal about an hour after the group first becomes reveal-eligible, so a new group does
  not wait for a calendar Friday. See `docs/features/voting.md` for the scheduling trigger
  and `docs/features/reveal.md` for the eligibility floors.
- **WEEKLY** — every subsequent season. Follows the Friday 20:00 MSK rhythm.

`seasons.postpone_count` records how many times a season's reveal has been postponed for
lack of participation.

### Season Creation on Group Create

1. Season #1 is created with status VOTING and `kind = KICKOFF`
2. Dates: `startsAt = now` (voting is open immediately), `revealAt` = weekly fallback
   Friday (see below), `endsAt` = Sunday 23:59 MSK of the reveal week
3. Questions: `min(10, memberCount * 2)` random system questions from group's categories, minimum 5
4. Questions from last 3 seasons of the same group are excluded (rotation)

### Weekly Season Schedule

`getWeeklySeasonDates(now)` is the single source of truth for season dates:

- `startsAt = now` — voting opens at creation, never on a later Monday
- `revealAt` = the next Friday 20:00 MSK that is at least `MinVotingWindow` (48h) away.
  A season created Monday–Wednesday morning reveals that same week; one created from
  Wednesday 20:00 onward reveals the following Friday, so voting is never shorter than
  two days.
- `endsAt` = Sunday 23:59 MSK of the reveal week

### Season Creation Paths

All three paths call `CreateWeeklySeasonForGroup`, which is a no-op if the group already
has a VOTING season, so they cannot double-create:

| Path | Trigger | Purpose |
|---|---|---|
| Post-kickoff follow-up | Enqueued by the reveal worker after a KICKOFF season reveals | A group that reveals mid-week starts its weekly cycle immediately instead of waiting for Sunday |
| Season creator cron | Sunday 18:00 UTC (21:00 MSK), `TypeSeasonCreator` | The normal weekly rhythm; Friday 20:00 → Sunday 23:59 stays reserved for discussion, reactions, and detectors |
| Maintenance cron | Hourly, `TypeSeasonMaintenance` | Safety net: creates a season for any group with >= 3 members that has none and whose latest season ended more than an hour ago |

The season creator also closes previous REVEALED seasons (status → CLOSED).

## Flutter Screens

### HomeScreen (`/home`)
- TabBar with groups list tab and profile placeholder tab
- FAB navigates to `/groups/create`
- Pull-to-refresh on groups list
- Empty state when user has no groups

### CreateGroupScreen (`/groups/create`)
- Name field (3-40 chars), category FilterChips, "Только добрые вопросы" switch, optional Telegram
  username field
- ROMANCE category is hidden from the chip list for users under 18 (checked via `authProvider` using `birth_year`)
- The kind-only switch shows the server's default (on for under-18 and for an unknown birth year) but
  **sends nothing until the creator touches it**: the age rule lives on the server, and a client that
  always sent its own value would override it — wrongly whenever it does not know the birth year. The
  copy next to it states what the group will and will not be asked rather than naming the setting
  alone. Category chips are *not* disabled while it is on: which categories are all-edgy is a property
  of the bank's contents, so the client shows the server's `NO_KIND_CATEGORIES` message instead of
  duplicating that judgement.
- On success: shows `InviteShareSheet` — the grouped code (`AB2 CD3`) shown as prominently as
  the link, each separately copyable, plus the system share sheet (`share_plus`). The code is
  copied in its stored form; the gap is cosmetic. See **Invite codes** above.
- After sharing: navigates to `/home`

### JoinGroupScreen (`/groups/join`)
- Single input for invite code or full link (`https://repa.app/join/{code}`). `UpperCaseFormatter`
  upper-cases a typed code for display but leaves a pasted URL untouched; the backend remains
  the authority on what a valid code is.
- 500ms debounced preview: shows group name, member count, and admin username before joining
- On join success: navigates to `/groups/:id`

### GroupScreen (`/groups/:id`)
- Pull-to-refresh
- `SmallGroupNotice` at the top while the group has fewer than 5 members: warns that
  percentages in a tiny group reveal who voted how (`kAnonymityMinMembers`, mirrors
  `eligibility.MinDetectorMembers`)
- `KindOnlyTile`: the group's kind-only setting, shown to every member and changeable only by the
  admin (a non-admin sees a disabled switch rather than nothing — the setting determines what members
  will be asked, so it is information about the group, not an admin preference). The admin's copy
  states that the change applies from the next репа, which is the part that surprises people. A
  refusal is shown inline, and the switch keeps showing the stored value so a refused change never
  looks saved.
- Season card: progress bar (voted / total), `PendingRevealNotice` explaining a
  non-`SCHEDULED` `reveal_state`, voting CTA button ("Проголосовать" / the season's
  `revealHeadline` when already voted), "Результаты готовы!" when REVEALED
- Members list with MemberAvatar widgets and "Админ" badge
- Share invite button in AppBar opens the same `InviteShareSheet`
- Telegram button in AppBar when `telegramUsername` is set (opens `t.me/` link via `url_launcher`)
- Voting CTA wired to the voting flow in T09: navigates to `/groups/:id/vote/:seasonId`
- `GrowthNotice`: one line naming what inviting more people changes, from the server's
  `next_threshold`. Renders nothing when the field is null — a group that is big enough is told so by
  the absence of a goal. An unlock key the client does not recognise states the number without claiming
  a consequence, rather than dropping the line or inventing one.
- History icon in the AppBar opens `/groups/:id/chronicle`

### ChronicleScreen (`/groups/:id/chronicle`)
- `GuessStandingCard` at the top: the знатоки ranking, or the server's reason while it is withheld.
  An unranked member shows a dash and "ещё не играл" rather than 0%.
- `ChronicleSeasonCard` per past season, newest first: the question, who led it, their share, and the
  small-sample warning on entries that carry it. A season whose entries were all withheld by a block
  still appears, showing "Нечего показать".
- Empty state names what will fill it rather than only reporting that it is empty
- States the cap when the history is longer than what is shown
- An unparseable `reveal_at` falls back to the raw string rather than losing the season

### GroupCard widget
- Shows group name, member count, voting progress
- Shimmer animation when user has not yet voted

### MemberAvatar widget
- Displays emoji avatar or photo (avatar_url)
- Supports streak badge

## State Management

Four Riverpod StateNotifiers in `groups_notifier.dart`:
- **groupsListProvider** (`GroupsListNotifier`) — loads and refreshes the user's group list
- **createGroupProvider** (`CreateGroupNotifier`) — handles group creation form state
- **joinGroupProvider** (`JoinGroupNotifier`) — handles preview + join flow
- **groupDetailProvider** (`GroupDetailNotifier`) — `.autoDispose.family` keyed by group ID, loads detail + members + active season

## File Structure

```
backend/
├── cmd/server/main.go                       # Route registration for groups
├── internal/
│   ├── handler/groups/
│   │   ├── handler.go                       # 8 handler methods + DTOs + error mapping
│   │   └── handler_test.go                  # DTO, error mapping, validation tests
│   ├── service/groups/
│   │   ├── service.go                       # Business logic, season creation, question selection
│   │   └── service_test.go                  # Category validation, date calculation, constants
│   └── db/
│       ├── migrations/002_groups_categories.up.sql
│       ├── queries/groups.sql               # 16 queries (CRUD, membership, admin transfer)
│       ├── queries/seasons.sql              # Season queries (create, active, voters)
│       ├── queries/questions.sql            # Question selection by categories with rotation
│       └── queries/season_questions.sql     # Season-question assignment

mobile/lib/features/groups/
  data/groups_repository.dart               # API calls via ApiService
  domain/group.dart                         # Freezed models: Group, GroupListItem, ActiveSeason, Member, GroupDetail, JoinPreview
                                            # + SeasonRevealState enum and ActiveSeasonRevealCopy
                                            #   extension (Russian copy for pending Reveals)
  presentation/
    groups_notifier.dart                    # 4 StateNotifiers + providers
    create_group_screen.dart                # Group creation form + invite share sheet
    group_screen.dart                       # Group detail: season card, members list
    join_group_screen.dart                  # Invite code/link input + join preview
    widgets/
      group_card.dart                       # List card with progress + shimmer
      member_avatar.dart                    # Emoji/photo avatar with streak badge
      pending_reveal_notice.dart            # Explains a pending Reveal (waiting / postponed)
      small_group_notice.dart               # Anonymity caution below 5 members

mobile/lib/features/home/home_screen.dart   # TabBar: groups list + profile placeholder
mobile/lib/core/router/app_router.dart      # Routes: /groups/create, /groups/join, /groups/:id, /join/:code deeplink
```

## Deeplink Handling

Route `/join/:code` is registered in go_router. When an unauthenticated user opens the link:
1. `_RouterNotifier` intercepts the deeplink and saves the invite code to `flutter_secure_storage` under key `pending_invite_code`
2. User is redirected to `/auth/phone`
3. After successful authentication, the `/home` route redirect reads `pending_invite_code`, clears it, and redirects to `/groups/join?code={code}`

## Key Dependencies

- Group handler → Group service → sqlc Queries
- Season creation uses `questions.GetRandomSystemQuestionsByCategories` with category filter and recent-question exclusion
- Invite URLs: `https://repa.app/join/{inviteCode}` where inviteCode is a UUID
- Mobile: `share_plus` for system share sheet, `url_launcher` for Telegram links
