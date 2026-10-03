# Voting

## Overview

Users vote on group members by answering fun questions during the VOTING phase of a season. Each question requires selecting one group member as the answer. Voting progress is saved server-side, so interrupted sessions can be resumed. After all questions are answered, a completion screen shows group-wide voting progress.

## API Endpoints

All endpoints require Bearer JWT authentication.

### `GET /api/v1/seasons/:seasonId/voting-session`
Get current voting session with questions, targets, and progress.
- **Success 200:** `{ "data": { "season_id": "...", "questions": [VotingQuestionDto], "targets": [TargetDto], "progress": { "answered": 2, "total": 5 } } }`
- **VotingQuestionDto:** question_id, text, category, answered (bool)
- **TargetDto:** user_id, username, avatar_emoji?, avatar_url?
- **Error 404:** `NOT_FOUND` — season not found
- **Error 400:** `SEASON_NOT_VOTING` — season is not in VOTING phase
- **Error 403:** `NOT_MEMBER` — user is not a member of the group
- **Behavior:** Returns all season questions with answered flags, all group members except the voter as targets, and the voter's progress count.

### `POST /api/v1/seasons/:seasonId/votes`
Cast a vote for a specific question.
- **Body:** `{ "question_id": "...", "target_id": "..." }`
- **Success 201:** `{ "data": { "vote": { "question_id": "...", "target_id": "..." }, "progress": { "answered": 3, "total": 5 } } }`
- **Error 409:** `ALREADY_VOTED` — already voted for this question
- **Error 400:** `SELF_VOTE` — cannot vote for yourself
- **Error 400:** `TARGET_NOT_MEMBER` — target not in group
- **Error 400:** `INVALID_QUESTION` — question not part of this season
- **Behavior:** Creates a vote record. voter_id is NEVER exposed in responses (anonymity).

### `GET /api/v1/seasons/:seasonId/progress`
Get group-wide voting progress for a season.
- **Success 200:** `{ "data": { "voted_count": 3, "total_count": 5, "quorum_reached": true, "quorum_threshold": 0.5, "user_voted": true } }`
- **Behavior:** voted_count = members who completed all questions. quorum_threshold is 0.5 (50%) for groups >= 8, or 0.4 (40%) for smaller groups.

## Flutter Screens

### VotingScreen (`/groups/:id/vote/:seasonId`)
Sequential question-answer flow:
- Progress bar at top showing current question number out of total (e.g. "2 из 5" in AppBar title)
- `SmallGroupNotice` above the question while the group has fewer than 5 members (group size
  is `targets.length + 1`, since targets exclude the voter) — shown *before* the first vote
  so members can decide what to answer rather than find out after the Reveal
- QuestionCard displays question text with category emoji, keyed by question ID to trigger slide-in animation on change
- Participant grid: 2x2 fixed grid for <= 4 targets (non-scrollable), scrollable GridView for more
- Selection flow: tap participant → purple border highlight → API call and 400ms delay run in parallel → advance to next question
- Back swipe blocked (PopScope canPop: false) — answers cannot be changed once submitted
- Exit via close button in AppBar shows a confirmation dialog; notes that progress is saved and voting can be resumed later
- On load: skips already-answered questions, starts from the first unanswered one
- Error state shows a retry button; errors during vote submission appear as a SnackBar

### VotingCompleteScreen (`/groups/:id/vote/:seasonId/complete`)
Shown automatically after the last vote is submitted (VotingScreen navigates via `context.go`):
- Party emoji with elastic scale-in + fade-in animation (flutter_animate)
- "Ты проголосовал!" headline with slide-up fade-in
- "Reveal в пятницу в 20:00" subtitle
- Group progress card: linear progress bar, "X из N проголосовали", "Кворум достигнут!" if quorum reached
- Progress is fetched independently via `groupVotingProgressProvider` (separate FutureProvider, not from VotingNotifier)
- "Назад в группу" button navigates to `/groups/:id`
- Back swipe blocked (PopScope canPop: false)
- Heavy haptic feedback on init

## State Management

**VotingNotifier** (Riverpod StateNotifier, `votingProvider.autoDispose.family` keyed by seasonId):
- `loadSession()` — fetches voting session, filters unanswered questions, shuffles targets
- `selectTarget(targetId)` — guards against double-tap (checks `submitting` and `selectedTargetId`), submits vote via API with a parallel 400ms minimum delay, advances `currentIndex` or sets `completed = true`
- Targets are re-shuffled after each question advances
- If session loads with all questions already answered, sets `completed = true` immediately

**groupVotingProgressProvider** — separate `FutureProvider.autoDispose.family` keyed by seasonId, used by VotingCompleteScreen to fetch progress independently of the autoDispose VotingNotifier.

## File Structure

```
mobile/lib/features/voting/
  domain/voting.dart               — Freezed models: VotingQuestion, VotingTarget, VotingProgress,
                                     VotingSession, VoteResultData, VoteInfo, GroupVotingProgress
  data/voting_repository.dart      — getVotingSession, castVote, getVotingProgress via ApiService
  presentation/
    voting_notifier.dart           — VotingState, VotingNotifier, votingProvider, groupVotingProgressProvider
    voting_screen.dart             — Main voting flow screen
    voting_complete_screen.dart    — Completion screen with animations
    widgets/
      question_card.dart           — Question text + category emoji display
      participant_card.dart        — Tappable member card with selected/disabled states
```

## Tests

31 tests across 4 files (added in T09):
- `test/features/voting/domain/voting_test.dart` — Freezed model JSON serialization
- `test/features/voting/presentation/voting_notifier_test.dart` — notifier logic (load, selectTarget, double-tap guard, completion)
- `test/features/voting/presentation/widgets/question_card_test.dart` — widget rendering
- `test/features/voting/presentation/widgets/participant_card_test.dart` — widget rendering + tap

## Business Rules

- Voter cannot vote for themselves (enforced server-side)
- **Blocked members are excluded from targets in both directions.** Targets come from
  `GetVotingTargets(group_id, voter_id)` — members minus the voter minus anyone blocked either way —
  rather than from filtering `GetGroupMembers` in Go, so the session's list and the vote endpoint cannot
  disagree. A vote for a blocked member is refused with `TARGET_BLOCKED` (400).
- **Votes cast before a block are kept.** Removing them would shift the target's card at the moment of
  the block, which is both a worse result and a signal that a block happened.
- **A kind-only group is never asked an edgy question.** A season's questions are drawn by
  `GetRandomSystemQuestionsByCategories`, which filters on `groups.kind_only` inside the query — before
  the `min(10, members*2)` limit, so a kind-only group gets a full season rather than a short one. See
  `docs/features/groups.md` → **Question tone** for the setting, its age-based default, and the
  `NO_KIND_CATEGORIES` refusal.
- **Changing the setting does not change an open session.** Season questions are materialised in
  `season_questions` at creation; the setting applies from the next season, so a member who already
  answered half a session is never asked a different set.
- Each question must have exactly one answer
- Answers cannot be changed once submitted
- Session can be interrupted and resumed — progress persists server-side
- Quorum: >= 50% of members must complete voting (>= 40% for groups < 8 members)
- voter_id is stored in the `votes` table but NEVER exposed in API responses tied to specific votes

### Completing a session can schedule a kickoff Reveal

When a vote completes a member's session (`answered == total`) **and** the season is a
`KICKOFF` season (see `docs/features/groups.md`), the voting service calls
`reveal.ScheduleKickoffIfEligible`. If that call is the one that makes the group
reveal-eligible for the first time — at least 3 members and at least 3 completed voters —
the season's `reveal_at` moves to one hour from now (`schedule.KickoffRevealDelay`) and a
`push:kickoff-scheduled` notification goes to the group.

- The scheduling UPDATE only fires for a `VOTING` kickoff season whose `reveal_at` is
  later than the new time, so it is idempotent and a later voter cannot push the Reveal back.
- The push is enqueued with a per-season asynq `TaskID`, so concurrent completions cannot
  produce two announcements.
- A scheduling failure is logged and swallowed: the vote is already recorded and must not
  be rolled back by a scheduling hiccup.
- Weekly seasons are untouched — they keep their Friday 20:00 MSK Reveal.
