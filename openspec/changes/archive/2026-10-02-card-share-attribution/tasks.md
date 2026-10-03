# Tasks

## 1. Schema and dependency

- [x] 1.1 Add migration `006_share_attribution.up.sql` with the `join_source` enum (`LINK`, `CODE`, `CARD`, `TELEGRAM`, `UNKNOWN`), `group_members.join_source` defaulting to `UNKNOWN`, and `share_events(id, user_id, season_id, channel, created_at)` with an index on `(season_id, channel)`; plus the down migration. Verify both apply and roll back cleanly against Postgres
- [x] 1.2 Add `github.com/skip2/go-qrcode` to `go.mod`; verify `go build ./...` passes and `go.sum` is updated
- [x] 1.3 Add queries for recording a share event and counting shares by channel and joins by source; verify `make sqlc` regenerates and `go build ./...` passes

## 2. Card call-to-action

- [x] 2.1 Add `internal/service/cards/qr.go` rendering an invite link to a base64 PNG data URI at medium error correction, returning an empty string on failure; verify unit tests assert a non-empty data URI for a valid link and an empty string (not an error) when rendering fails
- [x] 2.2 Extend `CardData` with the invite code and QR data URI, and add the CTA block to `BuildCardHTML` — code, QR on a light plate, and one instruction line; verify template tests assert the code and instruction appear and that the QR `<img>` is omitted when the data URI is empty
- [x] 2.3 Populate the new fields in `GenerateCardsForSeason` from the group's invite code; verify a unit test asserts the rendered card for a group contains that group's code
- [x] 2.4 Update `docs/features/cards.md` with the CTA block, the QR decisions (light plate, medium EC, degrade-on-failure) and the new `CardData` fields; verify every documented field exists

## 3. Join attribution

- [x] 3.1 Accept an optional source on `POST /api/v1/groups/join/:inviteCode`, mapping unknown values to `UNKNOWN`, and persist it on the membership; verify handler and service tests cover each valid source, a missing source, and an unrecognised value
- [x] 3.2 Record `join_source` in the group-creation path as well, so a founder's membership is not indistinguishable from an unattributed join; verify a unit test asserts the founder's source
- [x] 3.3 Update `docs/features/groups.md` with the join-source values and the API field; verify the documented values match the enum

## 4. Share events

- [x] 4.1 Add `POST /api/v1/seasons/:seasonId/shares` recording a share with its channel, best-effort and never failing the caller's flow; verify handler tests assert a 200 for each channel and that a repository error still returns success
- [x] 4.2 Expose share counts by channel and join counts by source in the admin stats endpoint; verify a handler test asserts both appear
- [x] 4.3 Document the share endpoint and the admin counts in `docs/features/cards.md` and `docs/features/moderation.md`; verify the documented shapes match the handlers

## 5. Mobile

- [x] 5.1 Build the share text from the group's invite code with the `card` channel marker and send it on share; verify unit tests assert the text contains the code-bearing link and the channel
- [x] 5.2 Report the share to the new endpoint after a successful share, swallowing failures; verify a unit test asserts the report is attempted and that a failure does not surface to the user
- [x] 5.3 Pass the arrival source through the join flow — `CARD` when arriving from a link with the card marker, `LINK` for a plain link, `CODE` when typed; verify unit tests cover each path
- [x] 5.4 Update `docs/features/cards.md` and `docs/features/groups.md` mobile sections; verify each documented behaviour has a test

## 6. Integration verification

- [x] 6.1 Add e2e coverage: generate a card for a group and assert the HTML contains the group's invite code and a QR data URI
- [x] 6.2 Add e2e coverage: join with each source value and assert the recorded `join_source`; join with an unrecognised value and assert `UNKNOWN` and a successful join
- [x] 6.3 Add e2e coverage: record shares on a season and assert the admin stats report them by channel
- [x] 6.4 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate card-share-attribution --strict` passes
