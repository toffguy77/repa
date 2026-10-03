# Tasks

## 1. Schema and grants service

- [x] 1.1 Add migration `007_referrals.up.sql` adding nullable `group_members.invited_by` referencing `users(id)` with an index, plus the down migration; verify both apply and roll back cleanly against Postgres
- [x] 1.2 Add queries for recording a referrer and for counting a user's referrals; verify `make sqlc` regenerates and `go build ./...` passes
- [x] 1.3 Add `internal/service/crystals/grants.go` with grant kinds, amounts, deterministic `external_id`s and a `Grant` method that treats a unique-violation as success; verify unit tests cover a first grant, a repeated grant paying once, a duplicate not erroring, and two different grants to one user both paying
- [x] 1.4 Document the grant kinds, amounts, idempotency keys and the reasoning in `docs/features/crystals.md`; verify every documented amount matches the code

## 2. Welcome grant

- [x] 2.1 Grant the welcome amount on successful registration in the auth service, best-effort and logged on failure; verify unit tests assert the grant on a new user and no second grant on a repeat attempt
- [x] 2.2 Verify registration still succeeds when the grant fails, with a unit test that forces a grant error

## 3. Referral

- [x] 3.1 Carry a referrer through the join path: accept it on `POST /groups/join/:code`, validate it is an existing member of that group and not the joining user, and persist it; verify tests cover a valid referrer, a stranger, a self-referral and an absent one
- [x] 3.2 Grant the inviter when an invitee completes their first voting session, keyed on the invitee so it pays once; verify unit tests assert payment on first completion, no payment on the join alone, and no second payment on later completions
- [x] 3.3 Put the sharer into the card's invite link and the client's share text; verify a card test asserts the rendered link contains the member's id and a mobile test asserts the share link carries it
- [x] 3.4 Update `docs/features/groups.md` and `docs/features/cards.md` with the referrer field, its validation and the exposure trade-off; verify the documented behaviour matches the tests

## 4. Achievement grants

- [x] 4.1 Grant crystals when a granting achievement is created, keyed on the achievement id; verify unit tests assert a streak milestone and the recruiter achievement pay, and that a non-granting achievement does not
- [x] 4.2 Update `docs/features/achievements.md` with which achievements pay and how much; verify the documented amounts match the code

## 5. Mobile

- [x] 5.1 Include the sharer in the share link built for a card; verify unit tests assert the referrer appears and that a link without one still parses
- [x] 5.2 Show grant entries in the crystal history with their reason, distinguishable from purchases; verify a widget test asserts a grant and a purchase render differently
- [x] 5.3 Update `docs/features/crystals.md` mobile section; verify each documented element has a test

## 6. Integration verification

- [x] 6.1 Add e2e coverage: a new user's balance covers one detector, and the welcome grant is not paid twice
- [x] 6.2 Add e2e coverage: join with a referrer, complete voting, assert the inviter was paid once; repeat the session and assert no second payment
- [x] 6.3 Add e2e coverage: join with an invalid referrer and with none, asserting the join succeeds and nobody is paid
- [x] 6.4 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate crystal-free-economy --strict` passes
