# Tasks

## 1. Schema

- [x] 1.1 Add migration `008_detector_hints.up.sql` creating `detector_hints(id, user_id, season_id, revealed_user_id, created_at)` with `UNIQUE(user_id, season_id, revealed_user_id)` and a `(user_id, season_id)` index, plus the down migration; verify both apply and roll back cleanly against Postgres
- [x] 1.2 Add queries to list a member's hints for a season, to pick an unrevealed voter at random, and to record a hint; verify `make sqlc` regenerates and `go build ./...` passes

## 2. Ladder service

- [x] 2.1 Add hint and full-list price constants to the reveal service with the hint below both the full list and the referral grant; verify a unit test asserts both relationships
- [x] 2.2 Implement `BuyDetectorHint` spending the hint price and recording the revealed voter in one transaction, refusing when the group is too small, when no unrevealed voter remains, when nobody voted, and when the balance is short; verify unit tests cover each refusal and assert the balance is untouched in every one
- [x] 2.3 Return a hint DTO carrying avatar and first character only, with no username field at all; verify a unit test asserts the type has no full-name field and that the first character is a rune rather than a byte
- [x] 2.4 Extend `GetDetector` with the ladder state — revealed hints, both prices, and whether a further hint is available; verify unit tests cover an untouched, a partially climbed, and a completed ladder
- [x] 2.5 Update `docs/features/reveal.md` and `docs/features/crystals.md` with the rungs, prices, the no-replacement rule and the anonymity limits; verify every documented price matches the constants

## 3. API

- [x] 3.1 Add `POST /api/v1/seasons/:seasonId/detector/hint` mapping each refusal to its error code; verify handler tests assert a successful hint, `GROUP_TOO_SMALL`, `INSUFFICIENT_FUNDS`, and the no-hints-left case
- [x] 3.2 Confirm the existing detector endpoints keep their current response fields so older clients still work; verify a handler test asserts `purchased` and `voters` are still present and shaped as before
- [x] 3.3 Document both endpoints and the extended response in `docs/features/reveal.md`; verify the documented fields match the DTOs

## 4. Mobile

- [x] 4.1 Extend the detector models with the ladder state and the hint shape; verify model tests round-trip the new fields and default safely when they are absent
- [x] 4.2 Rebuild the detector sheet as a ladder: the free count always visible, a hint action, revealed hints shown as partial rows, and the full list as the top action; verify widget tests cover each ladder state including small-group refusal
- [x] 4.3 Update `docs/features/reveal.md` mobile section; verify each documented state has a widget test

## 5. Integration verification

- [x] 5.1 Add e2e coverage: the free count is visible without a purchase; a hint reveals one voter partially and never the full name; two hints reveal two different voters
- [x] 5.2 Add e2e coverage: hints exhaust, a small group refuses every paid rung while still showing the count, and the full list still works after hints
- [x] 5.3 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate detector-ladder --strict` passes
