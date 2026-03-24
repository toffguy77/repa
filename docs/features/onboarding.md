# Onboarding & Question Voting

## Overview

New users see a 3-slide onboarding flow after registration. Groups also have a question voting feature where members vote on which questions to include in the next season.

## Onboarding Flow

### Trigger

- `isNewUser` flag on `AuthState`, set during registration
- Cleared by `onboardingCompleted()` after the user finishes or skips onboarding
- Route guard in `go_router` redirects new users to `/onboarding`

### Screens

3-slide `PageView` with dark purple background:

1. **Create a group** — explains how to create or join a group
2. **Anonymous voting** — describes the weekly voting mechanic
3. **Friday reveal** — introduces the reveal event

- `flutter_animate` entrance animations on each slide
- Haptic feedback on page transitions
- "Skip" and "Next" navigation buttons, "Start" on final slide

## Question Voting

### Concept

Between seasons, group members can vote on which user-submitted questions should be included in the next season's question pool.

### Voting Window

- Opens: Sunday 12:00 MSK
- Closes: Monday 17:00 MSK
- Client-side check determines availability

### API Endpoints

#### `GET /api/v1/groups/:groupId/next-season/question-candidates`
Get candidate questions available for voting.
- **Success 200:** `{ "data": { "candidates": [QuestionCandidateDto] } }`

#### `POST /api/v1/groups/:groupId/next-season/vote-question`
Cast a vote for a question candidate.
- **Body:** `{ "question_id": "uuid" }`
- **Success 200:** `{ "data": {} }`

### Screen States

The `QuestionVoteScreen` (`/groups/:id/question-vote`) has 5 states:

1. **Loading** — fetching candidates
2. **Voting** — showing candidate cards, user can vote
3. **Voted** — confirmation after voting
4. **Unavailable** — outside the voting window
5. **Error** — network or server error

### Mobile Architecture

```text
mobile/lib/features/questions/
├── data/questions_repository.dart
├── domain/question_candidate.dart        # Freezed: QuestionCandidate, VoteState
└── presentation/
    ├── question_vote_notifier.dart       # .family StateNotifier (per group)
    ├── question_vote_screen.dart         # Main screen with 5 states
    └── widgets/
        └── question_candidate_card.dart  # Swipeable vote card
```

### Onboarding Architecture

```text
mobile/lib/features/onboarding/
└── presentation/
    └── onboarding_screen.dart            # PageView with 3 slides, _SlideWidget defined inline
```
