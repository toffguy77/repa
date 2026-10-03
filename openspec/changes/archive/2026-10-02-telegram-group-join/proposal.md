# Proposal

## Why

Telegram is where this audience already is — the PRD names class and group chats as the primary
communication channel — and Repa's bot is already in those chats. But the bot only *announces*:
`docs/features/telegram.md` and §12.5 of the PRD both state the rule "Telegram — афиша,
приложение — концерт", and the webhook handles exactly two commands, `/connect <code>` and
`/repa` (`backend/internal/handler/telegram/handler.go`).

That leaves the product's hardest problem untouched. A group is worth nothing until it has three
members, and today the founder's only tools are a link and a code they have to distribute by
hand. Meanwhile a chat that has the bot in it already contains the exact people who should be in
the group, and the bot can see the chat.

This is the one acquisition path Gas never had, and it is currently unused.

## What Changes

- The bot gains a `/join` command: anyone in a connected chat can run it and get a one-tap way
  into that chat's group, without the founder sending anything to anyone.
- The bot's reply to `/join` carries the group's invite link with Telegram attributed as the
  channel, so chat-sourced joins are measurable alongside card-sourced ones.
- **BREAKING** (bot behaviour): `/connect` in a chat that is already connected to a different
  group now reports the conflict instead of silently re-pointing the chat.
- The bot answers `/help` and an unknown command with the short list of what it can do, because a
  command nobody knows exists is not a channel.
- When a chat is connected, the bot posts once explaining that members can now use `/join` —
  the only moment the whole chat is guaranteed to be looking.

## Capabilities

### New Capabilities

- `telegram-group-join`: how a person in a connected Telegram chat gets into that chat's Repa
  group, and what the bot will and will not say while doing it.

### Modified Capabilities

- `share-attribution`: a join originating from a Telegram chat is attributed to Telegram.

## Impact

- **Backend:** `internal/service/telegram/` (the `/join`, `/help` and unknown-command replies,
  connect-conflict handling), `internal/handler/telegram/handler.go` (command routing).
- **No schema changes** — the chat-to-group link and the `TELEGRAM` join source both already
  exist.
- **Docs:** `docs/features/telegram.md`, `docs/features/groups.md`.

## Non-goals

- Voting inside Telegram. The anonymity model and the Reveal ritual both depend on the app being
  the only place results exist; §12.5's rule stands, and this change does not touch it.
- Creating a Repa group from a chat, or importing a chat's member list. Telegram does not give a
  bot the member list of a group chat, and a group built without consent is the opposite of what
  this product is.
- Inline keyboards, deep-link start parameters, or a Telegram Mini App.
- Posting individual reputation cards into a chat — that remains explicitly forbidden.
