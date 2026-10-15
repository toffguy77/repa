# Tasks

## 1. Chronicle query and service

- [x] 1.1 Add `GetGroupChronicle(group_id, limit)` returning flat (season, question, winner, percentage, vote_count, total_voters) rows for the group's last N revealed seasons, newest first, one winner per question; verify `make sqlc` regenerates and the query runs against Postgres with a group that has several revealed seasons, an open season, and a postponed one — only the revealed ones appear
- [x] 1.1b Treat CLOSED as revealed everywhere a season's history is read — the profile queries (`GetUserSeasonHistory`, `GetTopAttributeAllTime`), `GetPreviousRevealedSeason` behind the card's trend line, and `GetLastNRevealedSeasons` behind the multi-season achievements, but *not* `GetRevealedSeasonsForGroup`, which is the closing step itself: a REVEALED season becomes CLOSED when the next one opens, so every one of those reads returned at most one season; verify a Postgres check shows multiple seasons for a group with several past reveals, and that the chronicle, the revealed-season count and the profile history all agree on which seasons count
- [x] 1.2 Add the chronicle service: membership check, grouping the flat rows into seasons in Go (not N+1 per season), withholding entries whose winner the reader has blocked; verify unit tests cover a non-member refusal, correct grouping, a blocked winner's entry being absent, and a departed winner's entry still being present
- [x] 1.3 Mark entries whose percentages were computed over fewer than `eligibility.MinDetectorMembers` voters, reading `total_voters` from the row rather than the group's size today; verify unit tests cover a marked entry, an unmarked one, and that the threshold comes from `eligibility` rather than a literal
- [x] 1.4 Add `GET /api/v1/groups/:id/chronicle`; verify handler tests cover the member success shape, the non-member 403, and an empty chronicle returning an empty list rather than an error
- [x] 1.5 Update `docs/features/groups.md` with the chronicle endpoint, what it contains, the season cap, the block rule and the anonymity marking; verify every documented rule has a test

## 2. Знатоки standing

- [x] 2.1 Add `GetGroupGuessStanding(group_id)` joining `group_members` to `user_group_stats`, ordered by accuracy, with members lacking a stats row included as unranked; verify the query runs against Postgres and a member who has never been through a reveal comes back unranked rather than at zero
- [x] 2.2 Return the standing from the chronicle endpoint, withheld entirely below two revealed seasons with a stated reason; verify unit tests cover one revealed season (withheld), two (shown), the ordering, the seasons-count per entry, and an unranked member
- [x] 2.3 Compute ranks before filtering blocked members, so a block does not renumber the list differently from what other members see; verify a unit test asserts a blocked member's absence leaves the remaining ranks unchanged
- [x] 2.4b Expose only the position and the seasons it is based on, not the accuracy figure: with the chronicle publishing each question's winner, an extreme figure recovers a member's individual votes, and the rolling average can be diffed across weeks into a match count; verify unit tests assert the figure is absent from both the struct and the serialised form
- [x] 2.4c Scope the block count to the group (`CountBlockedGroupMembers`) and judge the Reveal threshold on actual membership while the detector threshold uses the effective count; verify unit tests cover a 3-member group with one block reporting DETECTOR rather than REVEAL, and a block made outside the group leaving the effective count untouched
- [x] 2.4 Confirm the standing carries no vote: verify a unit test asserts the response contains no question, target or vote field, and an e2e test asserts the same over the real JSON
- [x] 2.5 Update `docs/features/profile.md` with the standing, its source (`user_group_stats.guess_accuracy`), the rolling window, and why it is withheld early; verify the documented window matches the computation in `internal/service/achievements`

## 3. Growth thresholds

- [x] 3.1 Add the next-threshold rule to `internal/eligibility`: the sizes that actually change behaviour (3 to reveal, 5 for the detector and anonymity), each with what it changes, and nothing above the last; verify unit tests cover a group below each threshold, one exactly on each, one above all of them, and that the 40%→50% quorum change at 8 members is not reported as a threshold
- [x] 3.2 Expose the next threshold on the group response, counted against `effective_member_count` so it agrees with the anonymity warning; verify handler tests cover a group of two, three and twelve, and a group of five with two blocks
- [x] 3.3 Update `docs/features/groups.md` with the thresholds, their source, and the rule that nothing above the last one is named; verify the documented sizes match the constants

## 4. Mobile

- [x] 4.1 Add the chronicle Freezed models and repository call; verify model tests round-trip the response, including an empty chronicle, a withheld standing, and a marked entry
- [x] 4.2 Add the chronicle screen: seasons newest first, the standout result per question, the anonymity mark where present, and an empty state naming what will fill it; verify widget tests cover a populated chronicle, an empty one, and a marked entry
- [x] 4.3 Add the standing to the chronicle screen, including the withheld state and unranked members; verify widget tests cover all three states
- [x] 4.4 Add the chronicle entry point to the group screen and the growth line, using the server's threshold rather than a client-side copy of it; verify widget tests cover the line below each threshold and its absence above the last
- [x] 4.5 Update `docs/features/groups.md` mobile section; verify each documented element has a test

## 5. Integration verification

- [x] 5.1 Add e2e coverage: a group with two revealed seasons exposes both in the chronicle while its open season is absent, and a non-member is refused
- [x] 5.2 Add e2e coverage: the standing is withheld after one reveal and present after two, and ordered by accuracy
- [x] 5.3 Add e2e coverage: the growth threshold reported for a group of two, three and five matches `eligibility`
- [x] 5.5 Add e2e coverage for the trend line against a CLOSED previous season — the regression for the status defect, which looked like "no trend yet" and so was never flagged; verify the test fails when the query filters on REVEALED alone
- [x] 5.6 Update `docs/specs/master_context.md`: the post-MVP migrations, the CLOSED-counts-as-revealed rule, the group-size thresholds and the question-tone rules; verify the documented thresholds match `internal/eligibility`
- [x] 5.4 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate season-chronicle --strict` passes
