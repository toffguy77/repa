# Telegram Bot Integration (T19 + T20)

## Overview

Telegram bot (@repaapp_bot) for publishing posts to linked group chats. Supports connect/disconnect flow, season-start and reveal auto-posts, manual card sharing, and status commands.

## API Endpoints

### `POST /api/v1/telegram/webhook`

Telegram Bot API webhook (public, secret-token validated).

- **Header:** `X-Telegram-Bot-Api-Secret-Token` must match `TELEGRAM_WEBHOOK_SECRET` env var. Validated with `crypto/subtle.ConstantTimeCompare`.
- **Handles:** `/connect CODE`, `/repa`, bot removal events (`my_chat_member` with status `kicked`/`left`).

### `POST /api/v1/groups/:id/telegram/generate-code`

Generate a connect code for linking a Telegram chat (admin only).

- **Success 200:** `{ "data": { "connect_code": "REPA-X7K2", "instruction": "...", "expires_at": "..." } }`
- **Errors:** 403 NOT_ADMIN, 404 NOT_FOUND

### `DELETE /api/v1/groups/:id/telegram`

Unlink Telegram chat from group (admin only).

- **Success 200:** `{ "data": { "disconnected": true } }`
- **Errors:** 403 NOT_ADMIN, 404 NOT_FOUND

### `POST /api/v1/seasons/:seasonId/share-to-telegram`

Share user's card to the group's Telegram chat.

- **Success 200:** `{ "data": { "shared": true } }`
- **Errors:** 400 NO_TELEGRAM (group has no linked chat)

All Telegram routes are only registered when `TELEGRAM_TOKEN` is set in config.

## Connect Flow

1. Admin calls `POST /groups/:id/telegram/generate-code` — gets `REPA-XXXX` code (24h TTL).
2. Admin adds @repaapp_bot to Telegram group chat, makes it admin.
3. Someone types `/connect REPA-XXXX` in the chat.
4. Bot calls `getChatMember` to verify it holds `administrator` or `creator` status in the chat.
5. Bot links `telegram_chat_id` and `telegram_chat_username` to the group, sends confirmation.

## Bot Commands

| Command | Description |
| --- | --- |
| `/connect CODE` | Link this chat to a Repa group |
| `/repa` | Show group name, season number and status, vote progress (voted/total), and reveal time MSK |

Disconnect is only available via the API (`DELETE /groups/:id/telegram`) to enforce admin-only access — there is no Telegram-to-Repa user mapping to verify admin status from a bot command.

## Auto-Posts

| Job Type | Trigger | Content |
| --- | --- | --- |
| `telegram:season-start` | Cron Mon 14:00 UTC (17:00 MSK) | New season announcement with inline "Проголосовать" button linking to the app |
| `telegram:reveal-post` | Enqueued by reveal worker after processing | Top 5 attributes (question + winner username + percentage), two inline buttons for card/members |
| `telegram:share-card` | User-initiated via API | Card photo sent as photo message with username caption |

## Auto-Unlink

When the bot is removed from a chat (`my_chat_member` update with status `kicked` or `left`), `HandleBotRemoved` calls `DisconnectTelegramByChat`, clearing `telegram_chat_id` and `telegram_chat_username` automatically.

## Business Rules

- Connect codes expire after 24 hours.
- Bot must be an administrator of the chat to complete connection (verified via `getChatMember` API call).
- Only group admin can generate connect codes and disconnect via API.
- Reveal post shows top 5 attributes (question text + winner username + integer percentage).
- Season-start post includes inline button linking to `APP_BASE_URL/group/:id`.
- Share card verifies the requesting user is a member of the group before posting.

## Backend Architecture

```text
backend/internal/lib/telegram.go                 # Telegram Bot API HTTP client (net/http, 15s timeout)
backend/internal/service/telegram/service.go     # Business logic: connect, disconnect, posts, share
backend/internal/handler/telegram/handler.go     # Webhook + REST endpoints
backend/internal/worker/tasks/telegram.go        # Asynq task handlers
```

## DB Queries Used

- `SetGroupConnectCode` — save connect code with expiry
- `GetGroupByConnectCode` — find group by unexpired code
- `UpdateGroupTelegram` — set `telegram_chat_id` + `telegram_chat_username`, or clear both on disconnect
- `GetGroupByTelegramChatID` — lookup for `/repa` command
- `DisconnectTelegramByChat` — clear chat fields on bot removal
- `GetActiveSeasonByGroup` — season data for `/repa` command response
- `CountGroupMembers` — member count for `/repa` command response
- `CountSeasonVoters` — voter count for `/repa` command response
- `GetAllVotingSeasons` — enumerate groups for season-start broadcast
- `GetSeasonByID` — resolve group from season for reveal post and share card
- `GetTopResultPerQuestion` — reveal post content (top attributes)
- `GetCardCache` — card image URL for share-to-chat
- `GetUserByID` — sender username for card share caption
- `IsGroupMember` — membership check before share card

## Flutter UI (T20)

### Screens & Widgets

- `TelegramSetupScreen` — admin-only screen at `/groups/:id/telegram`. Shows "not connected" state with connect button, or "connected" state (displaying `@chatUsername`) with disconnect button.
- `ConnectInstructionSheet` — bottom sheet shown after code generation. Shows 3 steps, copyable `/connect CODE` command, Open Telegram button, Verify button with countdown timer (24h expiry).
- `GroupScreen` — settings gear icon (admin-only) navigates to TelegramSetupScreen.
- `RevealScreen` — "Отправить в Telegram-чат" button (only rendered when `group.telegramUsername != null`). Native share upgraded to download PNG card to temp file via `path_provider` and share using `Share.shareXFiles`.

### Architecture

```text
mobile/lib/features/telegram/
├── data/telegram_repository.dart         # API calls: generate code, disconnect, share
├── domain/telegram_connect.dart          # Freezed model for connect code response
└── presentation/
    ├── telegram_notifier.dart            # StateNotifier + providers (setup state + shareToTelegramProvider)
    ├── telegram_setup_screen.dart        # Main setup screen
    └── connect_instruction_sheet.dart    # Bottom sheet with instructions
```

### Key Implementation Details

- `verifyConnection()` in the notifier calls `GET /groups/:id` (via `GroupsRepository.getGroup`) and checks `group.telegramUsername != null` — no dedicated verify endpoint.
- `shareToTelegramProvider` is a standalone `Provider.autoDispose` that returns the repository, used from `RevealScreen` outside the setup flow.
- Telegram button color: `Color(0xFF2AABEE)` (Telegram blue), distinct from app primary purple.

### API Endpoints Used

- `POST /groups/:id/telegram/generate-code` — generates connect code
- `DELETE /groups/:id/telegram` — disconnects Telegram chat
- `POST /seasons/:id/share-to-telegram` — shares card to linked chat
- `GET /groups/:id` — used to verify connection after `/connect` is sent in Telegram
