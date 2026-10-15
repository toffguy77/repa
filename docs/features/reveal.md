# Reveal

## Overview

The Reveal engine processes voting results every Friday at 17:00 UTC (20:00 MSK). It checks quorum, aggregates votes into per-user reputation cards, and triggers downstream jobs (achievements, card generation, push notifications). Each user sees their top attributes, a reputation title, trend vs. previous season, and can pay crystals to unlock hidden attributes or buy a detector to see who voted.

## API Endpoints

All endpoints require Bearer JWT authentication.

### `GET /api/v1/seasons/:seasonId/reveal`
Get current user's reveal card and group summary.
- **Success 200:** `{ "data": { "my_card": MyCardDto, "group_summary": GroupSummaryDto } }`
- **Error 404:** `NOT_FOUND` — season not found
- **Error 400:** `SEASON_NOT_REVEALED` — season is still in VOTING phase
- **Error 403:** `NOT_MEMBER` — user is not a member of the group
- **Behavior:** Returns user's top 3 attributes (open), remaining attributes (hidden/blurred), reputation title, trend vs. previous season, and group summary with top result per question.

### `GET /api/v1/seasons/:seasonId/members-cards`
Get all group members' reveal cards (top attributes only).
- **Success 200:** `{ "data": { "members": [MemberCardDto] } }`
- **Error 400:** `SEASON_NOT_REVEALED`
- **Error 403:** `NOT_MEMBER`
- **MemberCardDto:** user_id, username, avatar_emoji, avatar_url, top_attributes (AttributeDto[]), reputation_title

### `POST /api/v1/seasons/:seasonId/reveal/open-hidden`
Unlock hidden attributes for 5 crystals.
- **Success 200:** `{ "data": { "all_attributes": [AttributeDto], "crystal_balance": number } }`
- **Error 402:** `INSUFFICIENT_FUNDS` — balance < 5 crystals
- **Error 400:** `SEASON_NOT_REVEALED`
- **Error 403:** `NOT_MEMBER`
- **Behavior:** Deducts 5 crystals atomically (transaction), creates crystal_log with type `SPEND_ATTRIBUTES`, returns all attributes (top + hidden) with updated balance.

## Detector ladder

The detector was one purchase that answered the question completely — and a question, once
answered, stops being worth paying for. Gas monetised the same curiosity for roughly ten times as
much by *not* answering it. The detector is now three rungs:

| Rung | Price | Reveals |
|---|---|---|
| Count | **free** | "N человек проголосовали про тебя" — names nobody |
| Hint | 3 💎 | One voter's avatar and the **first character** of their name |
| Full list | 10 💎 | Every voter, by name (unchanged) |

- The count is free because its job is to *create* the question. Charging for the question is how a
  single-purchase detector ends up selling only the answer.
- The hint sits **below** the 5-crystal referral grant on purpose: a player who invited one friend
  can afford a rung without a bank card.
- **Hints are drawn without replacement.** `detector_hints` records which voter each hint revealed
  (`UNIQUE(user_id, season_id, revealed_user_id)`), so paying twice reveals two different people.
  Recording what was revealed — rather than counting hints and deriving an order — also keeps the
  next hint unguessable: any stable ordering would have to come from the voter ids themselves.
- Selection is **random** among the unrevealed. A fixed order would let a player infer the rule, and
  ordering by anything meaningful (join date, vote time) would leak a second fact beyond identity.
- The hint DTO has **no username field at all**, so no code path can leak a full name through a hint
  and no future change to a shared shape can add one. `first_letter` is the first *rune*: usernames
  can begin with a digit or a Cyrillic letter.
- The group-size floor applies to **every paid rung** — a hint drawn from two or three possible
  voters identifies someone as surely as a list does. The free count stays visible at any size.
- No rung ever binds a voter to a question or an answer. That rule is unchanged.

### `POST /api/v1/seasons/:seasonId/detector/hint`
Buys one partial reveal.
- **Success 200:** the full detector payload, with one more entry in `hints`
- **Error 402:** `INSUFFICIENT_FUNDS`
- **Error 403:** `GROUP_TOO_SMALL`, `NOT_MEMBER`
- **Error 409:** `NOTHING_TO_REVEAL` — every voter is revealed, nobody voted, or the full list is
  already owned
- **Error 400:** `SEASON_NOT_REVEALED`
- **Behavior:** the voter is picked *before* the crystals are spent, and the spend and the hint
  record happen in one transaction — so a refusal never costs crystals and a crash between the two
  cannot charge for nothing.

### `GET /api/v1/seasons/:seasonId/detector`
Get detector status for the current user.
- **Success 200:** `{ "data": { "purchased": bool, "voters": [VoterProfile], "available": bool, "crystal_balance": number, "voter_count": number, "hints": [DetectorHint], "hint_cost": number, "full_cost": number, "hint_available": bool } }`
- `DetectorHint` is `{ first_letter, avatar_emoji, avatar_url }` — there is no `username` field.
- The ladder fields are **additive**: a client that only reads `purchased` and `voters` keeps
  working, which is why this is not a separate endpoint — the Reveal screen already fetches the
  detector, and a second round trip on the screen that most needs to feel instant is not worth the
  tidier shape.
- `available` is false while the group has fewer than `MinDetectorMembers` (5) members, so the app can disable the button instead of letting a member discover the limit by spending.
- **Error 400:** `SEASON_NOT_REVEALED`
- **Error 403:** `NOT_MEMBER`
- **Behavior:** If purchased, returns voter profiles (IDs only, no question/answer binding). If not purchased, returns empty voters list.

### `POST /api/v1/seasons/:seasonId/detector`
Buy a detector for 10 crystals.
- **Success 200:** `{ "data": { "purchased": true, "voters": [VoterProfile], "crystal_balance": number } }`
- **Error 402:** `INSUFFICIENT_FUNDS` — balance < 10 crystals
- **Error 403:** `GROUP_TOO_SMALL` — group has fewer than `MinDetectorMembers` (5) members. Checked before any crystal deduction, so the balance is unchanged and no detector record is created.
- **Error 409:** `ALREADY_PURCHASED` — detector already bought for this season
- **Error 400:** `SEASON_NOT_REVEALED`
- **Error 403:** `NOT_MEMBER`
- **Behavior:** Deducts 10 crystals atomically, creates detector record and crystal_log, returns voter profiles.

### `GET /api/v1/seasons/:seasonId/my-card-url`
Get the current user's card image URL for a season.
- **Success 200:** `{ "data": { "image_url": "https://...", "status": "ready" } }`
- **Generating:** `{ "data": { "image_url": null, "status": "generating" } }` — card not yet ready
- **Error 400:** `SEASON_NOT_REVEALED`
- **Error 403:** `NOT_MEMBER`
- **Rate limit:** 5 requests per hour

### Response DTOs

```json
// MyCardDto
{
  "top_attributes": [AttributeDto],
  "hidden_attributes": [AttributeDto],
  "reputation_title": "string",
  "trend": TrendDto,
  "new_achievements": [AchievementDto],
  "card_image_url": ""
}

// AttributeDto
// Ordered by percentage with tone breaking near-ties inside a 5-point band
// (cardorder.BandPoints) — see docs/features/cards.md -> Attribute ordering.
// `rank` is assigned after that ordering, so it describes the card as shown.
// Tone itself is not exposed: it decides order, it is not a fact about the member.
{
  "question_id": "uuid",
  "question_text": "string",
  "category": "HOT|FUNNY|SECRETS|SKILLS|ROMANCE|STUDY",
  "percentage": 45.5,
  "rank": 1
}

// TrendDto
{
  "attribute": "string",
  "change": "up|down|same",
  "delta": 12.3
}

// GroupSummaryDto
{
  "top_per_question": [{ "question_id", "question_text", "user_id", "username", "avatar_emoji", "percentage" }],
  // VoterProfile (used by detector endpoints)
  // { "id", "username", "avatar_emoji", "avatar_url" }
  "voter_count": 5
}
```

## Data Model

### Tables

- **season_results** — id, season_id (FK seasons), target_id (FK users), question_id (FK questions), vote_count, total_voters, percentage. Stores aggregated vote counts per (target, question) pair.
  `GetSeasonResultsByUser` joins `questions.tone` so the card's order can use it; tone is not stored on
  the result, since it belongs to the question.
- **crystal_logs** — id, user_id (FK users), delta (integer), type (crystal_log_type enum), ref_id, created_at. Balance = `SUM(delta)`.
- **detectors** — id, user_id (FK users), season_id (FK seasons), group_id (FK groups), created_at. Tracks detector purchases per user per season.

### Key Queries (sqlc)

- `GetSeasonResultsByUser` — results for a specific user in a season, ordered by percentage DESC
- `GetTopResultPerQuestion` — DISTINCT ON (question_id), highest percentage per question with user info
- `AggregateVotesByTarget` — GROUP BY (target_id, question_id) vote counts for a season
- `DeleteSeasonResultsBySeason` — idempotent cleanup before re-aggregation
- `CreateSeasonResult` — insert aggregated result row
- `GetUserBalance` — `SUM(delta)` from crystal_logs
- `CreateCrystalLog` — insert crystal transaction
- `HasDetector` — check if user already purchased detector for a season
- `CreateDetector` — insert detector record
- `GetVoterProfilesBySeason` — voter user profiles for detector result (without question/answer binding)

## Mobile Screens

### RevealScreen waiting state

When the season has not revealed, the screen reads the group's active season
(`groupDetailProvider`) and shows that season's `revealHeadline` / `revealExplanation` —
"Нужно больше людей", "Нужно больше голосов", "Репа перенесена", or the Friday promise —
instead of a generic "Результаты ещё не готовы". A live `RevealCountdownWidget` is shown
only when `reveal_state` is `SCHEDULED`; counting down to a time that cannot produce a
Reveal is what made the Tuesday/Thursday pushes feel like a lie.

### DetectorSheet

`DetectorResult.available` (false below 5 members) disables the purchase action and
replaces it with an explanation. The flag defaults to `true` when absent, so an older
backend keeps working. Unavailability takes precedence over an insufficient balance:
telling a member to buy crystals they cannot spend here would be worse than useless.


## Anticipation

The push schedule used to promise things the app could not show: Tuesday's «Кто-то уже ответил на
вопросы про тебя 👀» landed on a progress bar, and Thursday's category teaser had no screen at all.
A push that promises intrigue and delivers nothing does not merely fail to retain — it teaches people
to ignore the next one, which costs the pushes that do work. (PRD RET-04/05, RET-09/10.)

### `GET /api/v1/seasons/:seasonId/anticipation`
- **Success 200:** `{ "data": { "voters_about_me": number, "teaser_emoji": string, "reveal_at": string } }`
- **Error 400:** `SEASON_ALREADY_REVEALED` — the member has their card; a partial signal next to a
  full result is noise
- **Error 403:** `NOT_MEMBER`
- **Error 404:** `NOT_FOUND`

### What it may and may not contain

The payload is designed to be **incapable** of leaking rather than merely not leaking today: it has
no identity field and no attribute field, so no future addition to a shared DTO can turn it into a
leak. A unit test asserts the exact key set.

- `voters_about_me` counts **distinct voters**, not votes, and excludes the member's own votes —
  answering about others must not inflate your own count.
- `teaser_emoji` is the **emoji itself**, not a category code the client could map back to a name. The
  mapping lives in Go beside the category enum, so a client cannot render a category it was not told
  about.
- Available before the member has voted themselves: the count is not a reward for participating.

### The Thursday rule

The teaser is included only from **Thursday** onward in MSK, gated server-side. Gating on the client
would make the teaser a client-version property and would ship the leading category to every device
from Monday. Thursday is the PRD's choice and the right one: a teaser on Monday has four days to
become boring, while a teaser on Thursday has one night.

The leading category is resolved by a grouped query over the member's received votes, with ties
broken on the category's own ordering — arbitrary but **stable**, because a random tiebreak would make
the emoji flicker between two values on consecutive loads and read as a bug.

### The immediate signal

When someone answers a question about a member, a `push:vote-signal` task is enqueued for that
member — never for the voter. It says that the number went up and nothing else: no voter name, no
attribute. The *what* stays sealed until Friday, which is the whole point of the week in between.

Debounced to **once per recipient per MSK day** (`signal-sent:{userID}:{date}` in Redis, the same
mechanism as the daily push cap). Without a bound, a 20-person group produces 19 notifications in an
evening and the app gets muted. Debounce rather than batch: a digest ("3 people answered") tells you
how fast interest is arriving, which is more than the product wants to give away before Friday.

The enqueue is a side effect of recording a vote and can never fail the vote.

### AnticipationPanel (mobile)

The reveal screen's waiting phase renders the count as an `AppStat`, the teaser emoji when present,
and a live countdown — on top of the pending-Reveal explanation. Zero votes says so in words rather
than showing an empty panel. The `reveal-waiting` push now routes to this screen rather than to the
group, which is what makes the mid-week pushes true.


## Business Rules

- **Blocks hide cards.** `GET /members-cards` omits members blocked in either direction; the viewer's
  own card is never hidden. The detector's group-size floor is judged against **effective** size
  (membership minus blocks involving the viewer), because a group of five with three blocks behaves like
  a group of two. See `docs/features/groups.md` → Member safety.
- **Participation floors (absolute):** a season never reveals with fewer than
  `MinRevealMembers` (3) members or fewer than `MinRevealVoters` (3) members who completed
  voting. These floors sit *above* the quorum percentage — the forced-reveal path bypasses
  the percentage, never the floors. Below them a card would be empty or trivially
  deanonymising.
- **Quorum:** >= 50% of group members must have completed voting for groups >= 8 members; >= 40% for smaller groups.
- **Retry on quorum miss:** up to 3 attempts, 2 hours apart. After the 3rd attempt the
  reveal proceeds regardless of the percentage — but only if the participation floors are met.
- **Postponement:** when the floors are unmet after the retries, the season is *postponed*
  rather than revealed: `reveal_at` moves to the next Friday 20:00 MSK, `ends_at` to that
  week's Sunday, `postpone_count` increments, and the status stays `VOTING`. Votes already
  cast remain valid. No card images, achievements, or Telegram posts are produced. Members
  get a `push:reveal-postponed` notification naming how many more people need to vote.
- **Eligibility predicate:** `reveal.Eligibility(ctx, seasonID)` is the single source of
  truth, consumed by the reveal worker (gate), the voting service (kickoff scheduling) and
  the groups handler (`reveal_state` shown in the app). It reports member count, completed
  voters, `members_needed`, `voters_needed`, quorum status, floor status, and eligibility.
- **Kickoff seasons:** a `KICKOFF` season (a new group's first — see
  `docs/features/groups.md`) is scheduled to reveal `KickoffRevealDelay` (1 hour) after the
  group first becomes eligible, instead of waiting for a Friday. After a kickoff reveals,
  the worker enqueues `season:create-for-group` so the group's weekly cycle starts
  immediately rather than waiting until Sunday.
- **Top attributes:** top 3 by percentage are always visible. Remaining are hidden (blurred in UI).
- **Hidden unlock cost:** 5 crystals per season per user.
- **Reputation title:** generated from the category of the top attribute:
  - HOT -> "Горячая штучка"
  - FUNNY -> "Душа компании"
  - SECRETS -> "Хранитель тайн"
  - SKILLS -> "Мастер на все руки"
  - ROMANCE -> "Сердцеед"
  - STUDY -> "Ботан года"
  - Fallback -> "Загадка века"
- **Trend:** compares top attribute percentage with the same attribute in the previous REVEALED season. Returns "up"/"down"/"same" with delta.
- **Anonymity:** voter_id is NEVER exposed in reveal responses. Results are aggregated vote counts only.
- **Downstream jobs:** after successful reveal, enqueues `achievements:calculate`, `cards:generate`, and `push:reveal-notification` tasks.
- **Card image:** `card_image_url` in MyCardDto is populated from `card_cache` table (see [cards.md](cards.md)).

## Worker Jobs

### `reveal:checker` (cron: every minute)
- Queries seasons where `status = VOTING AND reveal_at <= NOW()`
- Enqueues `reveal:process` task for each, queue: `critical`

### `reveal:process` (queued, payload: seasonID + attempt)
1. Check quorum (unique voters / total members)
2. If quorum met OR attempt >= 3 (forced):
   - Delete old results (idempotent)
   - Aggregate votes: COUNT per (target_id, question_id)
   - Compute percentage = vote_count / total_voters * 100 (1 decimal)
   - Insert season_results rows
   - Update season status -> REVEALED
   - Enqueue downstream: `achievements:calculate`, `push:reveal-notification`
3. If quorum not met and attempt < 3:
   - Re-enqueue with attempt+1, delay 2 hours

### Task Deduplication

`reveal:process` uses `asynq.TaskID("reveal:" + seasonID)` to prevent duplicate processing of the same season.

## Architecture

```text
backend/
├── internal/
│   ├── handler/reveal/handler.go          # 6 endpoints: GetReveal, GetMembersCards, OpenHidden, GetDetector, BuyDetector, GetMyCardURL
│   ├── service/reveal/service.go          # ProcessReveal, aggregateResults, GetReveal, GetMembersCards, OpenHidden, GetDetector, BuyDetector, ValidateRevealAccess, GetSeasonsForReveal
│   ├── worker/tasks/reveal.go             # HandleRevealChecker, HandleRevealProcess
│   └── db/
│       ├── queries/season_results.sql     # Result CRUD + aggregation queries
│       ├── queries/votes.sql              # CountUniqueVoters, AggregateVotesByTarget
│       ├── queries/crystal_logs.sql       # GetUserBalance, CreateCrystalLog
│       └── queries/detectors.sql          # HasDetector, CreateDetector, GetVoterProfilesBySeason
```

## Mobile (Flutter)

### Screens

- **RevealScreen** (`/groups/:id/reveal/:seasonId`) — 4 phases: loading, waiting (timer), ready (pulsing emoji + button), opening (3s animation), revealed (card + actions)
- **MembersRevealScreen** (`/groups/:id/reveal/:seasonId/members`) — list of member cards with top-3 attributes
- **DetectorBottomSheet** — blurred voter list until purchased (10 crystals), then reveals voter profiles
- **AchievementPopup** — full-screen overlay with bounce animation for new achievements, tap to cycle/dismiss

### Flutter Architecture

```text
mobile/lib/features/reveal/
├── data/reveal_repository.dart              # API calls
├── domain/reveal.dart                       # Freezed models (RevealData, MyCard, MemberCard, DetectorResult, etc.)
└── presentation/
    ├── reveal_notifier.dart                 # StateNotifier with RevealPhase enum
    ├── reveal_screen.dart                   # Main screen with phase-based rendering
    ├── members_reveal_screen.dart           # Member cards list
    └── widgets/
        ├── reputation_card.dart             # Card with attributes + hidden section
        ├── attribute_bar.dart               # Animated progress bar
        ├── detector_sheet.dart              # Bottom sheet for detector
        ├── achievement_popup.dart           # Full-screen achievement overlay
        └── reaction_bar.dart                # Emoji reaction buttons (see [reactions.md](reactions.md))
```

### Key Dependencies

- Reveal handler -> Reveal service -> sqlc Queries
- Reveal worker -> Reveal service (ProcessReveal, GetSeasonsForReveal)
- OpenHidden -> crystal_logs table (transactional deduct + log)
- Detector -> detectors table + crystal_logs (transactional deduct + create)
- Downstream: achievements (T11, implemented), push notifications (T17), reactions (T18, see [reactions.md](reactions.md))

### DetectorSheet (mobile)

The sheet renders the ladder:

- The free count is always on screen as an `AppStat`, before anything is bought.
- Revealed hints appear as rows showing the avatar and `М•••` — one character, never a name.
- The hint action disappears once `hint_available` is false (every voter revealed, or the list
  owned).
- An insufficient balance relabels a rung rather than hiding it, and routes to the shop.
- A small group shows the count and a single disabled action explaining the 5-member floor.
