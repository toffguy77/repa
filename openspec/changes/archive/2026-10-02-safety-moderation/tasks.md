# Tasks

## 1. Schema

- [x] 1.1 Add migration `009_member_safety.up.sql` creating `blocks(blocker_id, blocked_id)` with `UNIQUE(blocker_id, blocked_id)`, `group_bans(group_id, user_id)` with `UNIQUE(group_id, user_id)`, and `user_reports(reported_id, reporter_id, group_id, reason)` with `UNIQUE(reported_id, reporter_id)`, plus lookup indexes and the down migration; verify both apply and roll back cleanly against Postgres
- [x] 1.2 Add queries for creating and removing a block, listing blocks involving a user, checking a block between two users, creating and checking a ban, creating a user report, and listing user reports; verify `make sqlc` regenerates and `go build ./...` passes
- [x] 1.3 Add `GetVotingTargets(group_id, voter_id)` returning members minus the voter minus anyone blocked in either direction; verify a query test asserts a blocked member is absent in both directions

## 2. Safety service

- [x] 2.1 Add `internal/service/safety/` with block, unblock, report, remove-member and leave-permanently operations, each validating membership and authority; verify unit tests cover every refusal — self-block, non-member, non-admin removal, admin self-removal
- [x] 2.2 Make blocking and reporting idempotent, and make reporting independent of blocking; verify unit tests assert a second block and a second report both succeed without creating a duplicate
- [x] 2.3 Write the ban on admin removal and on permanent departure, leaving ordinary leaving reversible; verify unit tests assert a ban exists after removal and not after an ordinary leave
- [x] 2.4 Update `docs/features/moderation.md` with the person-report queue and `docs/features/groups.md` with removal, permanent departure and the ban rule; verify the documented behaviour matches the tests

## 3. Voting and results filtering

- [x] 3.1 Use `GetVotingTargets` for the voting session's target list; verify unit tests assert a blocked member is absent in both directions
- [x] 3.2 Refuse a vote whose target is blocked in either direction, with its own error; verify unit tests assert the refusal and that votes cast before the block are untouched
- [x] 3.3 Exclude blocked members from the members' cards view, while leaving the viewer's own card intact; verify unit tests cover both
- [x] 3.4 Update `docs/features/voting.md` and `docs/features/reveal.md` with the filtering and the keep-earlier-votes rule; verify each documented rule has a test

## 4. Join path and anonymity floor

- [x] 4.1 Refuse a join when a ban exists for that group, with a distinct error, including after the invite is regenerated; verify unit tests cover a banned join, a regenerated-invite join, and that other groups are unaffected
- [x] 4.2 Compute the anonymity floor from effective size (members minus blocks involving the viewer) rather than raw membership; verify unit tests assert a five-member group with three blocks is treated as small
- [x] 4.3 Update `docs/features/groups.md` and `docs/features/reveal.md` with the effective-size rule; verify the documented rule matches the tests

## 5. API

- [x] 5.1 Add block, unblock, report-member, admin remove-member and leave-permanently endpoints with their error mappings; verify handler tests cover success and each refusal
- [x] 5.2 Expose user reports in the admin queue alongside question reports; verify a handler test asserts both lists appear
- [x] 5.3 Document every endpoint in `docs/features/groups.md` and `docs/features/moderation.md`; verify the documented shapes match the handlers

## 6. Mobile

- [x] 6.1 Add block, unblock and report actions on a member, each behind a confirmation that states the consequence; verify widget tests cover the actions and their confirmations
- [x] 6.2 Add admin member-removal to the group screen, visible only to the admin, behind a confirmation that says it cannot be undone; verify widget tests assert visibility and the confirmation copy
- [x] 6.3 Add permanent leaving alongside ordinary leaving, with the difference stated; verify a widget test asserts both options and their copy
- [x] 6.4 Update `docs/features/groups.md` mobile sections; verify each documented element has a widget test

## 7. Integration verification

- [x] 7.1 Add e2e coverage: a block removes the member from voting targets in both directions and refuses a vote for them; earlier votes survive
- [x] 7.2 Add e2e coverage: an admin removes a member, who then cannot re-join with the invite or with a regenerated one, while an ordinary leaver can
- [x] 7.3 Add e2e coverage: a member report reaches the admin queue and does not notify the reported member
- [x] 7.4 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate safety-moderation --strict` passes
