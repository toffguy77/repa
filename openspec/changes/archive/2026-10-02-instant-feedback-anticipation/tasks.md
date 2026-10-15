# Tasks

## 1. Queries

- [x] 1.1 Add a query counting distinct voters who answered about a target in a season, excluding the target's own votes; verify `make sqlc` regenerates and a unit test asserts voters rather than votes are counted
- [x] 1.2 Add a query returning the leading question category among a target's received votes, with a stable tiebreak; verify `make sqlc` regenerates and `go build ./...` passes

## 2. Anticipation service

- [x] 2.1 Add the anticipation state to the reveal service: voter count, optional teaser emoji, reveal time; refuse for a revealed season and for a non-member; verify unit tests cover the count, zero votes, a non-member, and a revealed season
- [x] 2.2 Gate the teaser on Thursday-or-later in MSK, server-side; verify unit tests assert no teaser Monday–Wednesday and a teaser Thursday–Sunday, with no teaser at all when there are no votes
- [x] 2.3 Map categories to emoji in Go beside the enum, so a client cannot render a category it was not told about; verify a unit test asserts every category has an emoji
- [x] 2.4 Assert the payload is structurally incapable of leaking — a unit test marshals it and fails if any identity or attribute key appears
- [x] 2.5 Update `docs/features/reveal.md` with the endpoint, the payload, the Thursday rule and the anonymity limits; verify every documented field exists

## 3. API

- [x] 3.1 Add `GET /api/v1/seasons/:seasonId/anticipation` with error mapping; verify handler tests cover success, `SEASON_NOT_VOTING` for a revealed season, and `NOT_MEMBER`
- [x] 3.2 Document the endpoint in `docs/features/reveal.md`; verify the documented shape matches the handler

## 4. Immediate signal

- [x] 4.1 Enqueue a "someone answered about you" push for the target when a vote is recorded, as a side effect that cannot fail the vote; verify unit tests assert the enqueue happens for the target, never for the voter, and that an enqueue failure does not fail the vote
- [x] 4.2 Debounce the signal to once per recipient per MSK day using the existing Redis push-counter mechanism; verify unit tests assert the second send on the same day is suppressed and that a new day sends again
- [x] 4.3 Compose the signal with no voter name and no attribute; verify a unit test asserts the copy contains neither
- [x] 4.4 Document the signal and its debounce in `docs/features/push.md`; verify the documented behaviour matches the tests

## 5. Push honesty

- [x] 5.1 Filter the Tuesday signal fan-out to members with a non-zero count; verify a unit test asserts a member with no votes about them is skipped and one with votes is sent
- [x] 5.2 Filter the Thursday teaser fan-out to members who have a teaser; verify a unit test asserts the same
- [x] 5.3 Re-point the `reveal-waiting` push target at the screen that holds the information; verify a mobile test asserts the route
- [x] 5.4 Update `docs/features/push.md` with the filtering rule and why it exists; verify the documented schedule matches the handlers

## 6. Mobile

- [x] 6.1 Add the anticipation model and repository call; verify model tests round-trip the fields and default safely when the teaser is absent
- [x] 6.2 Render the anticipation screen in the reveal waiting phase — count, teaser when present, live countdown, on top of the pending-Reveal explanation; verify widget tests cover zero votes, a count, a count with a teaser, and the waiting-for-people case
- [x] 6.3 Update `docs/features/reveal.md` mobile section; verify each documented state has a widget test

## 7. Integration verification

- [x] 7.1 Add e2e coverage: votes about a member raise their count, the member's own votes do not, and the payload carries no identity or attribute
- [x] 7.2 Add e2e coverage: a revealed season refuses the anticipation state, and a non-member is refused
- [x] 7.3 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate instant-feedback-anticipation --strict` passes
