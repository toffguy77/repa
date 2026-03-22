# Reveal

## Overview

The Reveal engine processes voting results every Friday at 17:00 UTC (20:00 MSK). It checks quorum, aggregates votes into per-user reputation cards, and triggers downstream jobs (achievements, push notifications). Each user sees their top attributes, a reputation title, trend vs. previous season, and can pay crystals to unlock hidden attributes.

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
  "voter_count": 5
}
```

## Data Model

### Tables

- **season_results** — id, season_id (FK seasons), target_id (FK users), question_id (FK questions), vote_count, total_voters, percentage. Stores aggregated vote counts per (target, question) pair.
- **crystal_logs** — id, user_id (FK users), delta (integer), type (crystal_log_type enum), ref_id, created_at. Balance = `SUM(delta)`.

### Key Queries (sqlc)

- `GetSeasonResultsByUser` — results for a specific user in a season, ordered by percentage DESC
- `GetTopResultPerQuestion` — DISTINCT ON (question_id), highest percentage per question with user info
- `AggregateVotesByTarget` — GROUP BY (target_id, question_id) vote counts for a season
- `DeleteSeasonResultsBySeason` — idempotent cleanup before re-aggregation
- `CreateSeasonResult` — insert aggregated result row
- `GetUserBalance` — `SUM(delta)` from crystal_logs
- `CreateCrystalLog` — insert crystal transaction

## Business Rules

- **Quorum:** >= 50% of group members must have completed voting for groups >= 8 members; >= 40% for smaller groups.
- **Retry on quorum miss:** up to 3 attempts, 2 hours apart. After 3rd attempt, reveal proceeds regardless (forced).
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

## Architecture

```
backend/
├── internal/
│   ├── handler/reveal/handler.go          # 3 endpoints: GetReveal, GetMembersCards, OpenHidden
│   ├── service/reveal/service.go          # ProcessReveal, aggregateResults, GetReveal, GetMembersCards, OpenHidden
│   ├── worker/tasks/reveal.go             # HandleRevealChecker, HandleRevealProcess
│   └── db/
│       ├── queries/season_results.sql     # Result CRUD + aggregation queries
│       ├── queries/votes.sql              # CountUniqueVoters, AggregateVotesByTarget
│       └── queries/crystal_logs.sql       # GetUserBalance, CreateCrystalLog
```

### `GET /api/v1/seasons/:seasonId/detector`
Get detector status (purchased or not) and voter list.
- **Success 200:** `{ "data": { "purchased": bool, "voters": [VoterProfile], "crystal_balance": int } }`
- **Voters only returned if purchased.** Otherwise empty array.

### `POST /api/v1/seasons/:seasonId/detector`
Buy a detector for 10 crystals.
- **Success 200:** `{ "data": { "purchased": true, "voters": [VoterProfile], "crystal_balance": int } }`
- **Error 402:** `INSUFFICIENT_FUNDS`
- **Idempotent:** if already purchased, returns existing detector result.

## Mobile (Flutter)

### Screens

- **RevealScreen** (`/groups/:id/reveal/:seasonId`) — 4 phases: loading, waiting (timer), ready (pulsing emoji + button), opening (3s animation), revealed (card + actions)
- **MembersRevealScreen** (`/groups/:id/reveal/:seasonId/members`) — list of member cards with top-3 attributes
- **DetectorBottomSheet** — blurred voter list until purchased (10 crystals), then reveals voter profiles
- **AchievementPopup** — full-screen overlay with bounce animation for new achievements, tap to cycle/dismiss

### Architecture

```
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
        └── achievement_popup.dart           # Full-screen achievement overlay
```

### Key Dependencies

- Reveal handler -> Reveal service -> sqlc Queries
- Reveal worker -> Reveal service (ProcessReveal, GetSeasonsForReveal)
- OpenHidden -> crystal_logs table (transactional deduct + log)
- Detector -> detectors table + crystal_logs (transactional deduct + create)
- Downstream: achievements (T11, implemented), push notifications (T17)
