# Tasks

## 1. Code generation and normalisation

- [x] 1.1 Add `internal/service/groups/invite.go` with the 31-symbol alphabet, `generateInviteCode()` using `crypto/rand`, and `NormalizeInviteCode(raw)` that upper-cases, trims, strips an invite-URL prefix and drops spaces/hyphens; verify table-driven unit tests cover each input form from the spec plus the excluded-character rule
- [x] 1.2 Add a statistical unit test asserting that 1000 generated codes are all 6 characters, all within the alphabet, and all distinct; verify it fails if the alphabet or length regresses

## 2. Schema and queries

- [x] 2.1 Add migration `005_invite_code_short.up.sql` creating `UNIQUE INDEX groups_invite_code_upper_idx ON groups (upper(invite_code))` and dropping the superseded plain unique constraint, plus the matching down migration; verify both apply and roll back cleanly against a Postgres instance
- [x] 2.2 Change `GetGroupByInviteCode` in `internal/db/queries/groups.sql` to match on `upper(invite_code) = upper($1)`; verify `make sqlc` regenerates and `go build ./...` passes

## 3. Service behaviour

- [x] 3.1 Use the short generator in `CreateGroup` with bounded retries on unique-violation, failing with a dedicated error when attempts are exhausted; verify unit tests cover the happy path, one collision then success, and exhaustion
- [x] 3.2 Apply `NormalizeInviteCode` in the join and preview paths and in `RegenerateInviteLink`; verify unit tests assert a lowercase code, a padded code and a full URL all resolve to the same group
- [x] 3.3 Make `RegenerateInviteLink` issue a short code and confirm the previous code no longer resolves; verify a unit test asserts the old code returns not-found afterwards
- [x] 3.4 Update `docs/features/groups.md` with the code format, the accepted input forms, the uniqueness guarantee and the legacy-code behaviour; verify every documented input form has a corresponding test

## 4. Mobile

- [x] 4.1 Upper-case input in `JoinGroupScreen` as the user types (display only — the backend stays the authority on validity, per design.md) and update the field hint to show the code form; verify unit tests cover typed-code upper-casing and that a pasted URL is left untouched
- [x] 4.2 Show the grouped code (`AB2 CD3`) alongside the link in the invite sheet on `CreateGroupScreen` and `GroupScreen`, each separately copyable; verify widget tests assert both are present and that the grouped display strips to the stored code when copied
- [x] 4.3 Update `docs/features/groups.md` mobile sections with the invite presentation; verify each documented element has a widget test

## 5. Integration verification

- [x] 5.1 Add e2e coverage: create a group, assert the returned code is 6 characters from the alphabet, join by lowercase code, join by full URL, and confirm a regenerated code invalidates the old one
- [x] 5.2 Add e2e coverage asserting a legacy UUID code inserted directly into the database still resolves and still allows joining
- [x] 5.3 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate short-invite-codes --strict` passes
