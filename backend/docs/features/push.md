# Push Notifications (T17)

## Overview
FCM-based push notification system with weekly scheduler for retention.

## Components

### Firebase Client (`internal/lib/firebase.go`)
- Initializes Firebase Admin SDK from env credentials
- `SendPush(tokens, title, body, data)` — multicast to token list, auto-deletes invalid tokens
- `SendPushToUser(userID, ...)` — fetches user's FCM tokens from DB then sends
- `SendPushToUsers(userIDs, ...)` — batch send to multiple users

### Push Service (`internal/service/push/service.go`)
- Rate limiting: max 3 pushes per user per day (Redis counter, keyed by MSK date)
- Quiet hours: 23:00-09:00 MSK — no pushes sent
- Respects per-user push preferences (`push_preferences` table)
- `SendToUser(userID, category, title, body, data)` — single user with all checks
- `SendToUsers(userIDs, ...)` — batch with per-user checks
- `SendToGroupMembers(groupID, ...)` — send to all group members

### Push Handler (`internal/handler/push/handler.go`)
- `POST /api/v1/push/register` — register/update FCM token (upsert by token)
- `GET /api/v1/groups/:id/next-season/question-candidates` — 3 random system questions for voting
- `POST /api/v1/groups/:id/next-season/vote-question` — cast one vote per group per next season

### Push Worker Tasks (`internal/worker/tasks/push.go`)
Weekly cron schedule (all times MSK):
| Day | Time | Job | Description |
|---|---|---|---|
| Mon | 17:00 | weekly-scheduler | "New season started" to all VOTING groups |
| Tue | 19:00 | tuesday-signal | "Someone answered about you" to voters |
| Wed | 18:00 | wednesday-quorum | Quorum status — nag non-voters or warn all |
| Thu | 20:00 | thursday-teaser | Leading category hint to voters |
| Fri | 19:00 | friday-pre-reveal | "1 hour until Reveal" to everyone |
| Fri | 20:00 | reveal-notification | "Your repa is ready" (triggered by reveal process) |
| Sun | 12:00 | sunday-preview | "Vote for next week's questions" |
| Sun | 18:00 | sunday-streak | "Don't break your streak" to streak >= 3 |

Additional event-driven push:
- `reaction-push` — when someone reacts to a user's card

## Business Rules
- Max 3 pushes per user per day
- Quiet hours: 23:00-09:00 MSK
- Each push category can be disabled per user via `push_preferences`
- Invalid FCM tokens auto-deleted on send failure

## Environment Variables
- `FIREBASE_PROJECT_ID` — Firebase project ID
- `FIREBASE_PRIVATE_KEY` — service account private key
- `FIREBASE_CLIENT_EMAIL` — service account email
