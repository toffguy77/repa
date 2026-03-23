# Telegram Bot Integration (T19)

## Overview

Telegram bot (@repaapp_bot) for publishing posts to linked group chats. Supports connect/disconnect flow, season-start and reveal auto-posts, manual card sharing, and status commands.

## API Endpoints

### `POST /api/v1/telegram/webhook`
Telegram Bot API webhook (public, secret-token validated).
- **Header:** `X-Telegram-Bot-Api-Secret-Token` must match `TELEGRAM_WEBHOOK_SECRET` env var.
- **Handles:** `/connect CODE`, `/repa`, `/disconnect`, bot removal events.

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

## Connect Flow

1. Admin calls `POST /groups/:id/telegram/generate-code` -> gets `REPA-XXXX` code (24h TTL).
2. Admin adds @repaapp_bot to Telegram group chat, makes it admin.
3. Someone types `/connect REPA-XXXX` in the chat.
4. Bot verifies it's an admin of the chat, links `telegram_chat_id` to the group.
5. Bot sends confirmation message.

## Bot Commands

| Command | Description |
|---------|-------------|
| `/connect CODE` | Link this chat to a Repa group |
| `/repa` | Show current season status |

Disconnect is only available via the API (`DELETE /groups/:id/telegram`) to enforce admin-only access — there is no Telegram-to-Repa user mapping to verify admin status from a bot command.

## Auto-Posts (asynq jobs)

| Job Type | Trigger | Content |
|----------|---------|---------|
| `telegram:season-start` | Cron Mon 17:00 MSK | New season announcement with vote button |
| `telegram:reveal-post` | After reveal processing | Top attributes summary with deeplinks |
| `telegram:share-card` | User-initiated | Card photo with username caption |

## Auto-Unlink

When the bot is removed from a chat (`my_chat_member` update with status `kicked`/`left`), the `telegram_chat_id` is automatically cleared.

## Business Rules

- Connect codes expire after 24 hours.
- Bot must be an administrator of the chat to complete connection.
- Only group admin can generate connect codes and disconnect.
- Reveal post shows top 5 attributes (question + winner + percentage).
- Season-start post includes inline button linking to the app.

## Backend Architecture
```
backend/internal/lib/telegram.go                 # Telegram Bot API HTTP client (net/http)
backend/internal/service/telegram/service.go     # Business logic: connect, disconnect, posts
backend/internal/handler/telegram/handler.go     # Webhook + REST endpoints
backend/internal/worker/tasks/telegram.go        # Asynq task handlers
```

## DB Queries Used
- `SetGroupConnectCode` — save connect code with expiry
- `GetGroupByConnectCode` — find group by unexpired code
- `UpdateGroupTelegram` — set chat_id + username, clear code
- `GetGroupByTelegramChatID` — lookup for /repa command
- `DisconnectTelegramByChat` — clear chat_id on bot removal
- `GetTopResultPerQuestion` — reveal post content
- `GetCardCache` — card URL for sharing
