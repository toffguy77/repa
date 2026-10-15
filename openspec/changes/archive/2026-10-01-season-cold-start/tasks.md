# Tasks

## 1. Schema and generated code

- [x] 1.1 Add migration `backend/internal/db/migrations/004_season_kickoff.up.sql` creating the `season_kind` enum (`KICKOFF`, `WEEKLY`) and adding `seasons.kind season_kind NOT NULL DEFAULT 'WEEKLY'` plus `seasons.postpone_count integer NOT NULL DEFAULT 0`; verify by applying it to the dev DB with `psql -f` and confirming `\d seasons` shows both columns
- [x] 1.2 Add the matching `004_season_kickoff.down.sql` dropping both columns and the enum; verify by rolling it back and re-applying it cleanly against the dev DB
- [x] 1.3 Update `backend/internal/db/queries/seasons.sql`: have `CreateSeason` accept `kind`, add `ScheduleKickoffReveal` (conditional update that only moves `reveal_at` earlier for a `KICKOFF` season), `PostponeSeason` (moves `reveal_at`/`ends_at`, increments `postpone_count`), and `GetGroupsNeedingNewSeason` variant that also requires the latest season to have ended more than an hour ago; verify `make sqlc` regenerates without error and `go build ./...` passes

## 2. Season date and creation logic

- [x] 2.1 Replace `getNextSeasonDates()` in `backend/internal/service/groups/service.go` with `getWeeklySeasonDates(now)` returning `starts_at = now`, `reveal_at` = next Friday 20:00 MSK at least 48h out, `ends_at` = Sunday 23:59 MSK of the reveal week; verify with table-driven unit tests in `service_test.go` covering creation on each weekday including the Monday and Thursday boundary cases from the spec
- [x] 2.2 Make group creation produce a `KICKOFF` season with `starts_at = now` and `reveal_at` set to the weekly fallback Friday; verify a unit test asserts the created season's kind and that its voting window has already started
- [x] 2.3 Add `CreateWeeklySeasonForGroup` used by all three creation paths, and point `CreateNewSeasons` at it; verify existing season-creator tests still pass and a new test asserts the created season is `WEEKLY`
- [x] 2.4 Update `docs/features/groups.md` season-creation and season-creator sections to describe the kickoff season, the 48-hour rule, and the three creation paths; verify the documented reveal-time rule matches the behaviour asserted by the tests in 2.1

## 3. Reveal eligibility and postponement

- [x] 3.1 Add `MinRevealMembers = 3` / `MinRevealVoters = 3` constants and an `Eligibility(ctx, seasonID)` method to `backend/internal/service/reveal/service.go` returning member count, voted count, members needed, voters needed, and eligibility; verify unit tests cover below-floor, at-floor, and above-floor cases
- [x] 3.2 Gate `ProcessReveal` on `Eligibility` so the floors apply to the forced-reveal path as well, and postpone the season (next Friday, `postpone_count++`, status stays `VOTING`) instead of revealing when they are unmet; verify unit tests assert a 1-member group is never revealed across all retry attempts and that a 4-member/2-voter season at attempt 3 is postponed rather than revealed
- [x] 3.3 Make the reveal worker enqueue weekly-season creation for the group after a `KICKOFF` season reveals; verify a unit test on the worker asserts the follow-up task is enqueued for kickoff reveals and not for weekly reveals
- [x] 3.4 Ensure card generation and achievement calculation are only triggered for seasons that actually revealed; verify a unit test asserts no card-generation task is enqueued for a postponed season
- [x] 3.5 Add a postponement push informing members the Reveal is waiting and how many more people need to vote, routed through the existing push service so the daily cap and quiet hours apply; verify a unit test asserts the push is composed with the correct missing-voter count
- [x] 3.6 Update `docs/features/reveal.md` quorum and business-rules sections with the participation floors, the postponement behaviour, and `postpone_count`; verify the documented retry/postpone sequence matches the tests in 3.2

## 4. Kickoff reveal scheduling from voting

- [x] 4.1 In `backend/internal/service/voting/service.go`, when a vote completes a member's session and that completion first makes a `KICKOFF` season eligible, schedule `reveal_at = now + 1h` via the conditional query from 1.3; verify unit tests assert scheduling happens on the third completing voter and that a later voter does not move the time
- [x] 4.2 Send a "Reveal within the hour" push to group members when a kickoff reveal is scheduled; verify a unit test asserts the push is enqueued exactly once per season
- [x] 4.3 Update `docs/features/voting.md` with the kickoff-scheduling side effect of completing a voting session; verify the documented trigger matches the test in 4.1

## 5. API surface for pending Reveals

- [x] 5.1 Add `reveal_state`, `members_needed`, and `voters_needed` to the active-season objects returned by `GET /api/v1/groups` and `GET /api/v1/groups/:id`; verify handler tests assert each state value for a representative season
- [x] 5.2 Return `GROUP_TOO_SMALL` (403) from `POST /api/v1/seasons/:seasonId/detector` when the group has fewer than 5 members, checked before any crystal deduction, and report the same condition from `GET .../detector`; verify handler and service tests assert the error code and that the crystal balance is unchanged
- [x] 5.3 Update `docs/features/groups.md` and `docs/features/reveal.md` endpoint sections with the new fields and the `GROUP_TOO_SMALL` error; verify every documented field appears in the handler DTOs

## 6. Mobile: pending Reveal and small-group notice

- [x] 6.1 Extend the Freezed `ActiveSeason` models in `mobile/lib/features/groups/domain/group.dart` with the three new fields and regenerate; verify `dart run build_runner build --delete-conflicting-outputs` succeeds and the model JSON test round-trips the new fields
- [x] 6.2 Render the pending-Reveal state on `GroupScreen` — scheduled countdown, "нужно ещё N человек", or postponed — replacing the unconditional "Ждём пятницы" copy; verify widget tests cover each `reveal_state`
- [x] 6.3 Replace the static waiting text in `mobile/lib/features/reveal/presentation/reveal_screen.dart:198` with the same explanation of what the Reveal is waiting for; verify a widget test asserts the waiting state shows the missing-voter count
- [x] 6.4 Show the small-group anonymity notice on `GroupScreen` and at the start of the voting flow while the group has fewer than 5 members; verify widget tests assert the notice appears at 4 members and is absent at 5
- [x] 6.5 Disable the detector action with an explanatory label when the group is below 5 members; verify a widget test asserts the button is disabled and labelled for a 4-member group
- [x] 6.6 Update `docs/features/groups.md`, `docs/features/reveal.md`, and `docs/features/voting.md` mobile sections with the new screen states; verify each documented state corresponds to a widget test

## 7. Integration verification

- [x] 7.1 Add e2e coverage in `backend/tests/e2e` for the kickoff path: create group, add 3 members, all vote, assert `reveal_at` was brought forward and the season reveals with cards for every member
- [x] 7.2 Add e2e coverage for the floor path: a 2-member group whose reveal time has passed stays `VOTING`, has no cards, and reports a waiting `reveal_state`
- [x] 7.3 Run `go test ./...` and `flutter analyze && flutter test`, and confirm `openspec validate season-cold-start --strict` passes
