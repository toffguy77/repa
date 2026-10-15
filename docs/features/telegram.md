# Telegram Bot Integration (T19 + T20)


## Bot commands

The chat the bot already sits in contains the exact people who should be in the group, and a group
is worth nothing until it has three members. Before this, the founder's only tools were a link and
a code they had to distribute by hand — so the chat was an announcement board rather than a way in.

| Command | Reply |
|---|---|
| `/join`, `/join@<bot>` | The group's name and `https://repa.app/join/{code}?s=telegram`. Works for **any** member of the chat, not only the founder — the chat membership *is* the access control, which is the same trust boundary the founder applied when they connected it. |
| `/repa`, `/repa@<bot>` | Current season status: number, status, votes/members, Reveal time |
| `/connect <code>` | Links the chat to a group (code from the app) |
| `/disconnect` | Unlinks the chat |
| `/help`, `/start` | Lists the commands above |
| any other `/command` | The same command list — a command nobody knows exists is not a channel |
| a non-command message | **Nothing.** A bot that answers everything in a class chat gets removed from it. |

### `/join` in an unconnected chat

Replies with how to connect and **reveals no group**. Resolving some other group for a chat that
was never connected would be a cross-chat information leak; the failure mode worth designing
against here is a bot that is too helpful.

### Why a link rather than an inline button

An inline keyboard needs `reply_markup` handling and gives nothing a tappable link does not:
Telegram auto-links URLs, the universal link already handles the path, and the plain text keeps
working when the message is forwarded out of the chat. The reply names the group because a person
may be in several connected chats and a bare link does not say which group it leads to.

### Attribution

Every invite the bot hands out carries `?s=telegram`, and a join made with `source=TELEGRAM` is
counted separately from card-driven joins in `GET /api/v1/admin/stats` → `joins_by_source`. See
`docs/features/cards.md` → Sharing and attribution.

### Connect conflicts

Connecting a chat that is **already linked to a different group** is refused with
`ErrChatAlreadyConnected` ("Этот чат уже привязан к другой группе. Сначала напишите
/disconnect."). Previously the chat was silently re-pointed, which moved every future announcement
to the new group — invisible to the first group's members, who simply stopped hearing from the bot.
Re-running connect with the **same** group's code is idempotent and succeeds.

The connect confirmation announces the join command, because that is the one moment the whole chat
is reliably looking at the bot; a second message in quick succession is how a bot earns a mute.

### What the bot still never does

The rule from the PRD §12.5 — "Telegram — афиша, приложение — концерт" — is unchanged. No reply to
any command contains a member's attributes, percentages, or individual results, and individual
reputation cards are never posted into a chat.


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
- Telegram button colour: `BrandColors.telegram` (`lib/core/theme/brand_colors.dart`). It is a
  third-party brand colour, not a design token — it must not be themed or adjusted.

### API Endpoints Used

- `POST /groups/:id/telegram/generate-code` — generates connect code
- `DELETE /groups/:id/telegram` — disconnects Telegram chat
- `POST /seasons/:id/share-to-telegram` — shares card to linked chat
- `GET /groups/:id` — used to verify connection after `/connect` is sent in Telegram
