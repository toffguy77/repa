# Tasks

## 1. Join command

- [x] 1.1 Add `HandleJoinCommand(ctx, chatID)` to the telegram service, returning the group's name and `https://repa.app/join/{code}?s=telegram` for a connected chat, and connect instructions for an unconnected one; verify unit tests cover a connected chat, an unconnected chat, and that the reply contains no results data
- [x] 1.2 Route `/join` and `/join@<bot>` in the webhook handler; verify handler tests assert both forms reach the service and that the reply is sent to the originating chat
- [x] 1.3 Update `docs/features/telegram.md` with the command, its replies, and the attribution it carries; verify the documented reply shapes match the tests

## 2. Help and unknown commands

- [x] 2.1 Add a help reply listing the join, status and connect commands; verify a unit test asserts all three are named
- [x] 2.2 Route `/help` and answer any other `/`-prefixed command with the same list, while leaving non-command messages unanswered; verify handler tests cover a known command, an unknown command, and a plain message
- [x] 2.3 Document the help behaviour and the deliberate silence on non-commands in `docs/features/telegram.md`; verify each documented case has a test

## 3. Connect behaviour

- [x] 3.1 Make `HandleConnect` refuse a code for a different group when the chat is already connected, while treating a reconnect of the same group as success; verify unit tests cover the conflict, the idempotent reconnect, and that the original link survives the refusal
- [x] 3.2 Add the join-command line to the connect confirmation; verify a unit test asserts the confirmation mentions it
- [x] 3.3 Update `docs/features/telegram.md` connect section with the conflict rule and the confirmation copy; verify the documented behaviour matches the tests

## 4. Integration verification

- [x] 4.1 Add e2e coverage: connect a chat via the webhook, send `/join`, and assert the reply carries that group's code and the Telegram channel marker
- [x] 4.2 Add e2e coverage: `/join` in an unconnected chat reveals no group; an unknown command is answered; a plain message is not
- [x] 4.3 Add e2e coverage: joining through the bot's link records `TELEGRAM` as the source and appears in the admin funnel
- [x] 4.4 Run `go test ./...` and confirm `openspec validate telegram-group-join --strict` passes
