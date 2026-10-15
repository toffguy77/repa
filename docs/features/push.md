# Push Notifications (T17 + T18)

## Overview

FCM-based push notifications with rate limiting, quiet hours, and preference controls. Backend handles scheduling and delivery; Flutter handles token registration, foreground display, and tap-to-navigate.

## API Endpoints

### `POST /api/v1/push/register`
Register an FCM token for the current user.
- **Request:** `{ "token": "fcm_token", "platform": "ios"|"android" }`
- **Success 200:** `{ "data": { "registered": true } }`
- **Behavior:** Upserts token in `fcm_tokens` table. If token already exists for another user, reassigns it.

### `PATCH /api/v1/push/preferences`
Update push notification preferences per category.
- **Request:** `{ "category": "SEASON_START"|"REMINDER"|"REVEAL"|"REACTION"|"NEXT_SEASON", "enabled": bool }`
- **Success 200:** `{ "data": { "category": "SEASON_START", "enabled": false } }`
- **Note:** Implemented in `internal/handler/auth/handler.go`, not the push handler.

### `GET /api/v1/groups/:id/next-season/question-candidates`

Get question candidates for the upcoming season's community vote.

### `POST /api/v1/groups/:id/next-season/vote-question`

Cast a vote for a question candidate.

## Push Schedule (weekly cycle)

All scheduled tasks run via the asynq scheduler. Times are MSK (UTC+3).

| Day | MSK | UTC cron | Category | Content |
| --- | --- | -------- | -------- | ------- |
| Mon | 17:00 | `0 14 * * 1` | SEASON_START | "New season started" to all VOTING groups |
| Tue | 19:00 | `0 16 * * 2` | REMINDER | Social proof — "someone voted" signal, **only to members who actually have a non-zero count** |
| Wed | 18:00 | `0 15 * * 3` | REMINDER | Quorum status: near-done to non-voters, at-risk to everyone |
| Thu | 20:00 | `0 17 * * 4` | REMINDER | Leading category emoji hint, **only to members who have a teaser** |
| Fri | 19:00 | `0 16 * * 5` | REVEAL | 1h pre-reveal reminder to everyone |
| Fri | 20:00 | triggered by reveal worker | REVEAL | Reveal ready (enqueued after reveal processing) |
| Sun | 12:00 | `0 9 * * 0` | NEXT_SEASON | Question voting invite to all active groups |
| Sun | 18:00 | `0 15 * * 0` | REMINDER | Streak reminder to users who voted last season |

## Push Data Payload

All pushes include a `data` map for client-side navigation:
```json
{
  "screen": "reveal|reveal-waiting|vote|question-vote|shop",
  "groupId": "uuid",
  "seasonId": "uuid"
}
```

## Business Rules

- Max 3 pushes per user per day (Redis counter keyed by `push-count:{userID}:{MSK-date}`, TTL to midnight MSK).
- Quiet hours: 23:00–09:00 MSK — pushes are dropped, not queued.
- Per-category opt-out via `push_preferences` table (all categories enabled by default).
- Invalid/unregistered FCM tokens are auto-cleaned on send failure (`IsUnregistered` check).
- FCM client is optional at startup — if Firebase credentials are absent, pushes are silently skipped.

## Mobile (Flutter)

### Initialization
Push is initialized via `ref.listenManual` on `authProvider` in `main.dart` — triggers `pushService.init()` on first authenticated state, avoiding side effects during the build phase.

`init()` does:

1. `Firebase.initializeApp()`.
2. `requestPermission()` (iOS; alert + badge + sound).
3. Register token via `POST /push/register` if permission granted.
4. Set up foreground, background-tap, and terminated-tap listeners.
5. Register token refresh listener for automatic re-registration.

### Message Handling

- **Foreground:** `FirebaseMessaging.onMessage` — message data is logged (`debugPrint`). No in-app UI shown in MVP.
- **Background tap:** `FirebaseMessaging.onMessageOpenedApp` — navigates via `data['screen']`.
- **Terminated tap:** `FirebaseMessaging.instance.getInitialMessage()` — same navigation as background tap.
- **Background handler:** Top-level `_firebaseMessagingBackgroundHandler` annotated `@pragma('vm:entry-point')`, re-initializes Firebase.

### Deep-link Navigation (tap routing)

| `screen` value | Route |
| --- | --- |
| `reveal` | `/groups/:groupId/reveal/:seasonId` |
| `reveal-waiting` | `/groups/:groupId` |
| `vote` | `/groups/:groupId/vote/:seasonId` |
| `question-vote` | `/groups/:groupId` |
| `shop` | `/shop` |

### Architecture
```
mobile/lib/main.dart                              # ref.listenManual triggers push init on auth
mobile/lib/core/services/push_service.dart        # FCM init, token registration, message routing
```

## Backend Architecture
```
backend/internal/handler/push/handler.go          # POST /push/register, question-candidate endpoints
backend/internal/handler/auth/handler.go          # PATCH /push/preferences (UpdatePushPreferences)
backend/internal/service/push/service.go          # SendToUser, SendToUsers, SendToGroupMembers
backend/internal/lib/firebase.go                  # FCM client (multicast, invalid token cleanup)
backend/internal/worker/tasks/push.go             # All scheduled + event push task handlers
backend/internal/db/queries/push.sql              # FCM token queries, top-category query
backend/internal/db/queries/push_preferences.sql  # Preference upsert + lookup
```

## A push only promises what the app can show

The Tuesday and Thursday fan-outs filter their recipients by whether the screen behind the push will
actually contain the thing the push mentions (`withVotesAbout`). Previously both went to every member
of every voting group, so a member with no votes about them was promised intrigue and shown a progress
bar. That costs more than the missed send: it teaches the recipient to ignore the next notification,
including the ones that work.

The filter costs one query per candidate, bounded by group size and run twice a week off-peak.

## The immediate vote signal

`push:vote-signal` — enqueued when someone answers a question about a member, for that member only.

- Says the number went up; carries **no** voter name and **no** attribute.
- **Debounced to once per recipient per MSK day** via `signal-sent:{userID}:{date}`, the same
  Redis-key-expiring-at-midnight mechanism as the daily cap. A 20-person group would otherwise produce
  19 notifications in an evening.
- Counts against the 3-per-day cap and respects quiet hours like any other push.
- Routes to `reveal-waiting`, which now opens the anticipation screen rather than the group.

See `docs/features/reveal.md` → Anticipation.
