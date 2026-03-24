# Reactions (T18)

## Overview

Users can react to other members' reveal cards with one of 5 emojis. One reaction per user per target per season (upsert). Reactions are displayed anonymously (only counts visible).

## API Endpoints

### `POST /api/v1/seasons/:seasonId/members/:targetId/reactions`

Create or update a reaction on a member's card.

- **Request:** `{ "emoji": "one of allowed emojis" }`
- **Allowed emojis:** `["😂", "🔥", "💀", "👀", "🫡"]`
- **Success 200:** `{ "data": { "counts": { "😂": 3, "🔥": 1 }, "my_emoji": "😂" } }`
- **Error 400:** `INVALID_EMOJI` — emoji not in allowed set
- **Error 400:** `SELF_REACTION` — cannot react to own card
- **Error 400:** `SEASON_NOT_REVEALED` — season status is not REVEALED
- **Error 404:** `NOT_FOUND` — season not found
- **Error 403:** `NOT_MEMBER` — user not in group
- **Behavior:** Upsert (one reaction per reactor/target/season). Enqueues `push:reaction` task for target.

### `GET /api/v1/seasons/:seasonId/members/:targetId/reactions`

Get aggregated reaction counts for a member's card.

- **Success 200:** `{ "data": { "counts": { "😂": 3, "🔥": 1 }, "my_emoji": "😂" } }`
- **`my_emoji`** is the current user's reaction (null if none).
- **Counts are anonymous** — no reactor IDs exposed.

## Data Model

### Table: `reactions`

- `id` TEXT PK
- `season_id` TEXT FK seasons
- `reactor_id` TEXT FK users
- `target_id` TEXT FK users
- `emoji` TEXT
- `created_at` TIMESTAMPTZ
- `UNIQUE(season_id, reactor_id, target_id)`

### Key Queries

- `CreateReaction` — INSERT ON CONFLICT UPDATE emoji (upsert)
- `GetReactionsForUser` — all reactions for a target in a season, with reactor info

## Business Rules

- One reaction per reactor per target per season (changing emoji replaces previous).
- Only allowed when the season status is REVEALED — returns `SEASON_NOT_REVEALED` otherwise.
- Cannot react to own card.
- Reactions are anonymous in API responses (only aggregated counts + own emoji).
- Creating a reaction triggers `push:reaction` notification to target.

## Push Notification

When a reaction is created, a `push:reaction` task is enqueued:

- Title: "Кто-то отреагировал на твою карточку {emoji}"
- Body: "Зайди и посмотри!"
- Data: `{ screen: "reveal", groupId, seasonId }`

## Mobile (Flutter)

### ReactionBar Widget

- 5 emoji buttons in a row with count badges.
- Selected reaction (user's own) highlighted with purple border.
- Tap triggers `sendReaction` with optimistic local update, confirmed by server response.
- Haptic feedback (`HapticFeedback.lightImpact`) on tap.
- Displayed under each member card in `MembersRevealScreen`.

### Architecture

```text
mobile/lib/features/reveal/
├── domain/reveal.dart                          # ReactionCounts model
├── data/reveal_repository.dart                 # getReactions, createReaction
├── presentation/
│   ├── reveal_notifier.dart                    # loadReactions, sendReaction (optimistic)
│   ├── members_reveal_screen.dart              # Integrates ReactionBar per member
│   └── widgets/reaction_bar.dart               # ReactionBar widget
```

## Backend Architecture

```text
backend/internal/
├── handler/reactions/handler.go                # CreateReaction, GetReactions
├── service/reactions/service.go                # Business logic + push enqueue
├── db/queries/reactions.sql                    # SQL queries
└── db/sqlc/reactions.sql.go                    # Generated (do not edit)
```
