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
- **Progress bar** at top showing X of N questions answered
- **QuestionCard** displays question text with category emoji, slide-in animation
- **Participant grid** (2x2 for <= 4 targets, scrollable grid for more) showing group members
- **Selection flow:** tap participant -> purple border + checkmark animation -> 400ms delay -> slide to next question
- **Back swipe blocked** (PopScope canPop: false) — answers cannot be changed
- **Exit dialog** confirms exit, notes progress is saved
- **Resume:** loads session from API, skips already-answered questions automatically

### VotingCompleteScreen (`/groups/:id/vote/:seasonId/complete`)
Shown after last vote:
- Party emoji with elastic scale-in animation
- "Ты проголосовал!" headline
- "Reveal в пятницу в 20:00" subtitle
- Group progress card: X of N voted, quorum status
- "Назад в группу" button navigates to group detail

## State Management

**VotingNotifier** (Riverpod StateNotifier):
- `loadSession()` — fetches voting session, determines unanswered questions, shuffles targets
- `selectTarget(targetId)` — submits vote, advances to next question or completes
- Prevents double-taps during submission
- Targets are re-shuffled between questions

## File Structure

```
mobile/lib/features/voting/
  domain/voting.dart          — Freezed models (VotingQuestion, VotingTarget, VotingSession, etc.)
  data/voting_repository.dart — API calls via ApiService
  presentation/
    voting_notifier.dart       — State management + providers
    voting_screen.dart         — Main voting flow screen
    voting_complete_screen.dart — Completion screen with animations
    widgets/
      question_card.dart       — Question display with category emoji
      participant_card.dart    — Tappable member card with selection state
```

## Business Rules

- Voter cannot vote for themselves (enforced server-side)
- Each question must have exactly one answer
- Answers cannot be changed once submitted
- Session can be interrupted and resumed — progress persists server-side
- Quorum: >= 50% of members must complete voting (>= 40% for groups < 8 members)
- voter_id is stored but NEVER exposed in API responses tied to specific votes
