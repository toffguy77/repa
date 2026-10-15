# Design

## Context

See proposal.md — Why. What exists: `HandleConnect` links a chat to a group by a one-time code,
`HandleRepaCommand` reports season status, and the webhook routes on the message text in a
`switch` (`internal/handler/telegram/handler.go`). The chat-to-group link lives on
`groups.telegram_chat_id`, and `join_source` already has a `TELEGRAM` value from
`card-share-attribution`. No schema work is needed.

## Goals / Non-Goals

**Goals:**

- Turn the chat the bot is already sitting in into a way into the group.
- Keep the "Telegram is the poster, the app is the concert" rule intact while doing it.
- Make chat-driven joins measurable next to card-driven ones.

**Non-Goals:**

- Voting, results, or cards in Telegram.
- Reading a chat's member list (Telegram does not offer it to bots) or adding people without
  their action.

## Decisions

### `/join` replies with a link, not a button

The reply is a plain message containing `https://repa.app/join/{code}?s=telegram`. An inline
keyboard would look better, but it requires `reply_markup` handling in the Telegram client
wrapper and gives nothing a tappable link does not: Telegram auto-links URLs, the app's
universal link already handles the path, and the same text works when the bot's message is
forwarded out of the chat.

The reply names the group, because a person may be in several chats and the link alone does not
say which group they are about to join.

### An unconnected chat gets instructions, never a group

`/join` in an unconnected chat replies with how to connect and nothing else. Resolving some
*other* group for a chat that has not been connected would be a cross-chat information leak, and
the failure mode worth designing against here is a bot that is too helpful.

### Command routing keeps the existing `switch`, with a default arm

The webhook's routing stays a `switch` on the text; `/join`, `/help` and a `default` arm are
added. The default replies with the command list only for messages that start with `/` — an
ordinary chat message gets silence, because a bot that answers everything in a class chat gets
removed from it.

Alternative considered: a command registry map. Rejected for five commands; the `switch` is still
the most readable form and the mention-suffix handling (`/join@repaapp_bot`) is already expressed
there as a prefix test.

### Connect conflicts are reported instead of silently re-pointing

`HandleConnect` currently writes the chat id onto whichever group presented a valid code. If a
chat is already connected to a different group, that silently moves every future announcement to
the new group — invisible to the first group's members, who simply stop hearing from the bot.
Reporting the conflict makes it a decision someone has to make deliberately (disconnect first),
rather than a side effect of a code being pasted in the wrong chat.

Reconnecting the *same* group stays a success: it is idempotent, and a code pasted twice is not
an error worth surfacing.

### The connect confirmation is the announcement

Rather than a separate "the bot has joined" post, the existing connect confirmation gains a line
about `/join`. The connect moment is the one time the whole chat is reliably looking at the bot,
and a second message in quick succession is how a bot earns a mute.

## Risks / Trade-offs

- **`/join` lets anyone in the chat into the group** → that is the intent; the chat membership
  *is* the access control, which is the same trust boundary the founder applied when they
  connected the chat. The group's 50-member cap and the admin's ability to regenerate the code
  remain the backstops.
- **The bot becomes chattier** → only on commands and once on connect; the default arm stays
  silent for non-commands specifically to avoid this.
- **A forwarded `/join` reply works outside the chat** → the link is the same invite that is
  already shareable by design, so this leaks nothing the invite does not.
- **Reporting connect conflicts changes existing behaviour** → it only affects the case where a
  chat is pointed at a second group, which today loses the first group silently; no legitimate
  flow depends on the old behaviour.

## Migration Plan

No schema or data changes. Deploy the backend; the new commands are live immediately for chats
that already have the bot. Rollback is reverting the binary — the chat-to-group links it reads are
unchanged.
