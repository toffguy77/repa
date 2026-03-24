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
- **GroupListItemDto:** id, name, member_count, invite_code, telegram_username, active_season (id, status, reveal_at, voted_count, total_count, user_voted)

### `GET /api/v1/groups/:id`
Get group detail with members and active season.
- **Success 200:** `{ "data": { "group": GroupDto, "members": [MemberDto], "active_season?": { id, number, status, starts_at, reveal_at, ends_at } } }`
- **Error 403:** `NOT_MEMBER`
- **Error 404:** `NOT_FOUND`
- **MemberDto:** id, username, avatar_emoji, avatar_url, is_admin
- **Note:** The active_season fields differ from the list endpoint (`GET /groups`), which returns `id, status, reveal_at, voted_count, total_count, user_voted`.

### `GET /api/v1/groups/join/:inviteCode/preview`
Preview group before joining (no membership required).
- **Success 200:** `{ "data": { "name": "...", "member_count": 5, "admin_username": "..." } }`
- **Error 404:** `NOT_FOUND`

### `POST /api/v1/groups/join/:inviteCode`
Join a group by invite code.
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

## Data Model

### Tables

- **groups** — id, name, invite_code (UNIQUE), admin_id (FK users), categories (text[]), telegram_chat_id, telegram_chat_username, telegram_connect_code, telegram_connect_expiry, created_at
- **group_members** — id, user_id (FK users), group_id (FK groups), joined_at. UNIQUE(user_id, group_id)
- **seasons** — id, group_id (FK groups), number, status (season_status enum), starts_at, reveal_at, ends_at, created_at
- **season_questions** — id, season_id (FK seasons), question_id (FK questions), ord

### Enums

- `season_status`: VOTING, REVEALED, CLOSED
- `question_category`: HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY

### Migration 002

Added `categories text[] NOT NULL DEFAULT '{}'::text[]` column to `groups` table.

## Business Rules

- **Group limits:** max 10 groups per user, max 50 members per group
- **Admin:** creator becomes admin. Only admin can update group name, telegram, and regenerate invite link
- **Admin transfer:** when admin leaves, the next member by join date becomes admin
- **Group deletion:** when last member leaves, group is deleted (CASCADE)
- **ROMANCE restriction:** users under 18 (by birth_year) cannot include ROMANCE category when creating a group. Enforced both server-side (handler returns `ROMANCE_BLOCKED`) and client-side (category hidden in CreateGroupScreen).
- **Valid categories:** HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY

### Season Creation on Group Create

1. Season #1 is created with status VOTING
2. Dates: startsAt = next Monday 00:00 MSK, revealAt = Friday 20:00 MSK (17:00 UTC), endsAt = Sunday 23:59 MSK (20:59 UTC)
3. Questions: `min(10, memberCount * 2)` random system questions from group's categories, minimum 5
4. Questions from last 3 seasons of the same group are excluded (rotation)

### Season Creator Job

- Cron: every Sunday 18:00 UTC (21:00 MSK)
- Creates new VOTING season for all active groups (>= 3 members) that don't have one
- Closes previous REVEALED season
- Asynq task type: `TypeSeasonCreator`

## Flutter Screens

### HomeScreen (`/home`)
- TabBar with groups list tab and profile placeholder tab
- FAB navigates to `/groups/create`
- Pull-to-refresh on groups list
- Empty state when user has no groups

### CreateGroupScreen (`/groups/create`)
- Name field (3-40 chars), category FilterChips, optional Telegram username field
- ROMANCE category is hidden from the chip list for users under 18 (checked via `authProvider` using `birth_year`)
- On success: shows invite share bottom sheet with copy-to-clipboard and system share sheet (`share_plus`)
- After sharing: navigates to `/home`

### JoinGroupScreen (`/groups/join`)
- Single input for invite code or full link (`https://repa.app/join/{code}`)
- 500ms debounced preview: shows group name, member count, and admin username before joining
- On join success: navigates to `/groups/:id`

### GroupScreen (`/groups/:id`)
- Pull-to-refresh
- Season card: progress bar (voted / total), voting CTA button ("Проголосовать" / "Ждём пятницы" when already voted), "Результаты готовы!" when REVEALED
- Members list with MemberAvatar widgets and "Админ" badge
- Share invite button in AppBar (builds URL from `inviteCode`)
- Telegram button in AppBar when `telegramUsername` is set (opens `t.me/` link via `url_launcher`)
- Voting CTA wired to the voting flow in T09: navigates to `/groups/:id/vote/:seasonId`

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
  presentation/
    groups_notifier.dart                    # 4 StateNotifiers + providers
    create_group_screen.dart                # Group creation form + invite share sheet
    group_screen.dart                       # Group detail: season card, members list
    join_group_screen.dart                  # Invite code/link input + join preview
    widgets/
      group_card.dart                       # List card with progress + shimmer
      member_avatar.dart                    # Emoji/photo avatar with streak badge

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
