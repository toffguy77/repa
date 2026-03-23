# Push Notifications (T17 + T18)

## Overview

FCM-based push notifications with rate limiting, quiet hours, and preference controls. Backend handles scheduling and delivery; Flutter handles token registration, foreground display, and tap-to-navigate.

## API Endpoints

### `POST /api/v1/push/register`
Register an FCM token for the current user.
- **Request:** `{ "token": "fcm_token", "platform": "ios"|"android" }`
- **Success 200:** `{ "data": { "ok": true } }`
- **Behavior:** Upserts token in `fcm_tokens` table. If token already exists for another user, reassigns it.

### `PATCH /api/v1/push/preferences`
Update push notification preferences.
- **Request:** `{ "category": "SEASON_START"|"REMINDER"|"REVEAL"|"REACTION"|"NEXT_SEASON", "enabled": bool }`
- **Success 200:** `{ "data": { "ok": true } }`

## Push Schedule (weekly cycle)

| Day | Time (MSK) | Type | Content |
|-----|-----------|------|---------|
| Mon | 17:00 | SEASON_START | New season started |
| Tue | 19:00 | REMINDER | Someone voted signal |
| Wed | 18:00 | REMINDER | Quorum status |
| Thu | 20:00 | REMINDER | Category teaser |
| Fri | 19:00 | REMINDER | Pre-reveal reminder |
| Fri | 20:00 | REVEAL | Reveal ready |
| Sun | 12:00 | NEXT_SEASON | Question voting |
| Sun | 18:00 | REMINDER | Streak reminder |

## Push Data Payload

All pushes include a `data` field for navigation:
```json
{
  "screen": "reveal|reveal-waiting|vote|question-vote|shop",
  "groupId": "uuid",
  "seasonId": "uuid"
}
```

## Business Rules

- Max 3 pushes per user per day (Redis counter, TTL midnight MSK).
- Quiet hours: 23:00-09:00 MSK — pushes are dropped, not queued.
- Per-category opt-out via `push_preferences` table (all enabled by default).
- Invalid/unregistered tokens are auto-cleaned on send failure.

## Mobile (Flutter)

### Initialization
- `Firebase.initializeApp()` on app start (after auth).
- Permission request (iOS).
- Token registration via `POST /push/register`.
- Token refresh listener re-registers automatically.

### Message Handling
- **Foreground:** `FirebaseMessaging.onMessage` — logged (snackbar TBD).
- **Background tap:** `FirebaseMessaging.onMessageOpenedApp` — navigates via `data.screen`.
- **Terminated tap:** `FirebaseMessaging.instance.getInitialMessage()` — same navigation.

### Architecture
```
mobile/lib/core/services/push_service.dart    # FCM init, token registration, message routing
```

## Backend Architecture
```
backend/internal/handler/push/handler.go       # POST /push/register, question voting endpoints
backend/internal/service/push/service.go       # SendToUser, SendToUsers, SendToGroupMembers
backend/internal/lib/firebase.go               # FCM client (multicast, cleanup)
backend/internal/worker/tasks/push.go          # All scheduled + event push handlers
```
