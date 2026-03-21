# Authentication

## Overview

Authentication for the Repa app. Users can sign in via phone OTP, Apple ID, or Google. After auth, a JWT token is issued (90-day expiry). New users are prompted to complete their profile (username, avatar emoji, birth year). Supports avatar image upload to S3, push notification preferences, account deletion, and app version enforcement.

## API Endpoints

### Public

#### `POST /api/v1/auth/otp/send`
Send OTP code to a phone number.
- **Body:** `{ "phone": "+79001234567" }` (E.164 format, required)
- **Success 200:** `{ "data": { "sent": true } }` (in dev mode also returns `"code": "123456"`)
- **Error 429:** `RATE_LIMIT` — max 3 OTP sends per phone per hour
- **Rate limit:** Redis counter `rl:otp:{phone}`, 1-hour TTL

#### `POST /api/v1/auth/otp/verify`
Verify OTP code and authenticate.
- **Body:** `{ "phone": "+79001234567", "code": "123456" }` (code: 6 digits, required)
- **Success 200:** `{ "data": { "token": "jwt...", "user": { UserDto } } }`
- **Error 401:** `INVALID_OTP` — wrong code
- **Error 429:** `OTP_BLOCKED` — max 5 attempts per phone per 5 minutes
- **Behavior:** Creates user if phone not found (generates `user_XXXXXXXX` username)

#### `POST /api/v1/auth/apple`
Sign in with Apple ID token.
- **Body:** `{ "id_token": "..." }` (required)
- **Success 200:** `{ "data": { "token": "jwt...", "user": { UserDto } } }`
- **Error 401:** `INVALID_TOKEN` — token validation failed
- **Behavior:** Validates JWT signature against Apple's JWKS, extracts `sub`, creates/finds user by `apple_id`

#### `POST /api/v1/auth/google`
Sign in with Google ID token.
- **Body:** `{ "id_token": "..." }` (required)
- **Success 200:** `{ "data": { "token": "jwt...", "user": { UserDto } } }`
- **Error 401:** `INVALID_TOKEN` — token validation failed
- **Behavior:** Validates via Google's tokeninfo endpoint, extracts `sub`, creates/finds user by `google_id`

#### `GET /api/v1/auth/username-check?username=...`
Check if username is available.
- **Query:** `username` (required)
- **Success 200:** `{ "data": { "available": true } }`
- **Error 400:** `VALIDATION` — invalid format
- **Rate limit:** 20 requests per minute per IP

#### `GET /api/v1/app/version`
Check app version and force update status.
- **Header:** `X-App-Version: 1.0.0` (optional)
- **Success 200:** `{ "data": { "min_version": "1.0.0", "latest_version": "1.0.0", "force_update": false } }`

### Protected (requires Bearer JWT)

#### `GET /api/v1/auth/me`
Get current user profile.
- **Success 200:** `{ "data": { UserDto } }`
- **Error 404:** `NOT_FOUND` — user deleted

#### `PATCH /api/v1/auth/profile`
Update user profile.
- **Body:** `{ "username?": "newname", "avatar_emoji?": "...", "birth_year?": 2005 }`
- **Validation:** username 3-20 chars, birth_year 1990-2012
- **Success 200:** `{ "data": { UserDto } }`
- **Error 409:** `USERNAME_TAKEN` or `USERNAME_COOLDOWN` (30-day change limit)
- **Error 400:** `VALIDATION` — invalid format

#### `POST /api/v1/auth/avatar`
Upload avatar image.
- **Body:** multipart form, field `file` (JPEG or PNG, max 5MB)
- **Success 200:** `{ "data": { UserDto } }` (avatar_url populated)
- **Error 400:** `FILE_TOO_LARGE` or `INVALID_IMAGE`
- **Error 503:** `UNAVAILABLE` — S3 not configured
- **Behavior:** Resizes to 256x256 (Lanczos), converts to JPEG 85% quality, uploads to S3

#### `PATCH /api/v1/push/preferences`
Update push notification preference.
- **Body:** `{ "category": "SEASON_START|REMINDER|REVEAL|REACTION|NEXT_SEASON", "enabled": true }`
- **Success 200:** `{ "data": { "category": "...", "enabled": true } }`

#### `DELETE /api/v1/auth/account`
Delete user account (hard delete, cascades to all related data).
- **Success 200:** `{ "data": { "deleted": true } }`

### Response DTOs

```json
// UserDto
{
  "id": "uuid",
  "username": "string",
  "avatar_url": "string | null",
  "avatar_emoji": "string | null",
  "birth_year": 1234 | null,
  "created_at": "2026-01-01T00:00:00Z"
}
```

## Data Model

### Tables

- **users** — id (PK), phone (UNIQUE), apple_id (UNIQUE), google_id (UNIQUE), username (UNIQUE), avatar_url, avatar_emoji, birth_year, created_at, updated_at
- **push_preferences** — id (PK), user_id (FK users), category (push_category enum), enabled (bool). UNIQUE(user_id, category)
- **fcm_tokens** — id (PK), user_id (FK users), token (UNIQUE), platform, created_at

### Enums

- `push_category`: SEASON_START, REMINDER, REVEAL, REACTION, NEXT_SEASON

## Business Rules

- **OTP rate limiting:** max 3 sends per phone per hour, max 5 verify attempts per 5 minutes
- **Username format:** regex `^[a-zA-Za-яА-ЯёЁ0-9_]{3,20}$` — letters (Latin + Cyrillic), digits, underscore
- **Username cooldown:** can only change username once every 30 days
- **New user detection (mobile):** if `avatar_emoji == null || birth_year == null` after auth → redirect to profile setup
- **JWT:** HS256, 90-day expiry, claims: UserID, Username
- **Avatar:** max 5MB, JPEG/PNG only (validated by magic bytes), resized to 256x256, stored as JPEG 85% in S3
- **Dev mode:** when `DEV_MODE=true`, OTP code is returned in the send response
- **Push preferences:** default to enabled if no preference record exists
- **Account deletion:** hard delete with CASCADE — removes all user data

## Mobile Screens

### Phone Screen (`/auth/phone`)
- Text field with `+7 (___) ___-__-__` mask (mask_text_input_formatter)
- "Получить код" button, disabled until 10 digits entered
- Apple/Google sign-in buttons (stubs, disabled)
- Loading spinner during OTP send
- Error message display

### OTP Screen (`/auth/otp`)
- 6-digit Pinput input, auto-submits on completion
- 5-minute countdown timer (Отправить повторно через MM:SS)
- "Отправить код повторно" button appears after timer expires
- Error message display for invalid code
- Navigation: success → router redirect to `/home` or `/auth/setup`

### Profile Setup Screen (`/auth/setup`)
- Shown only when `needsProfileSetup == true` (no avatar_emoji or no birth_year)
- Emoji avatar picker: 20 emojis in a grid, tap to select
- Username field with 500ms debounce availability check (green checkmark / red X)
- Birth year field: digits only, 4 chars max
- "Готово" button, disabled until valid form (username >= 3 chars, birth year in range)
- Validation: birth year between `currentYear - 22` and `currentYear - 14`

### Home Screen (`/home`) — stub
- Welcome message, logout button
- Placeholder for groups (T07)

### Navigation (go_router)
- Auth redirect: no token → `/auth/phone`
- Profile setup redirect: authenticated + needsProfileSetup → `/auth/setup`
- Authenticated + complete profile → `/home`
- Uses `ChangeNotifier` + `refreshListenable` pattern to avoid GoRouter recreation

## Architecture

### Backend

```
backend/
├── cmd/server/main.go                    # Entrypoint, route registration
├── internal/
│   ├── config/config.go                  # Env-based config (34 vars)
│   ├── handler/
│   │   ├── errors.go                     # ErrorResponse(), ErrorHandler()
│   │   └── auth/handler.go              # 11 handler methods
│   ├── service/auth/service.go          # Business logic, JWT signing, S3 upload
│   ├── middleware/
│   │   ├── auth.go                      # JWTAuth middleware, GetCurrentUser()
│   │   ├── ratelimit.go                 # Redis-based rate limiter
│   │   └── validator.go                 # go-playground/validator wrapper
│   ├── db/
│   │   ├── migrations/001_init.up.sql   # Full schema (14 tables, 7 enums)
│   │   ├── queries/users.sql            # User CRUD queries
│   │   ├── queries/push_preferences.sql # Push pref queries
│   │   └── sqlc/                        # Generated Go code (DO NOT EDIT)
│   └── lib/
│       ├── db.go                        # pgxpool connection
│       ├── redis.go                     # go-redis client
│       ├── s3.go                        # S3 upload client
│       └── asynq.go                     # Task queue client + 15 task types
```

### Mobile

```
mobile/lib/
├── main.dart                                         # App entrypoint, ProviderScope
├── core/
│   ├── api/
│   │   ├── api_client.dart                          # Dio factory, auth interceptor, parseError
│   │   └── api_service.dart                         # API methods (Dio wrapper)
│   ├── providers/
│   │   ├── api_provider.dart                        # Dio, ApiService, SecureStorage providers
│   │   └── auth_provider.dart                       # AuthNotifier, AuthState, AuthStatus
│   ├── router/app_router.dart                       # GoRouter with refreshListenable
│   └── theme/
│       ├── app_colors.dart                          # #7C3AED purple, surface, text colors
│       ├── app_text_styles.dart                     # headline1/2, body, caption, button
│       └── app_theme.dart                           # MaterialApp ThemeData
└── features/
    ├── auth/
    │   ├── data/auth_repository.dart                # sendOtp, verifyOtp, checkUsername, updateProfile
    │   ├── domain/user.dart                         # Freezed User entity
    │   └── presentation/
    │       ├── auth_notifier.dart                   # OtpSendNotifier, OtpVerifyNotifier, ProfileSetupNotifier
    │       ├── phone_screen.dart                    # Phone input screen
    │       ├── otp_screen.dart                      # OTP verification screen
    │       └── profile_setup_screen.dart            # Profile completion screen
    └── home/home_screen.dart                        # Stub home screen
```

### Key Dependencies

- **Backend:** Dio interceptor adds `Authorization: Bearer {token}` from FlutterSecureStorage
- **401 handling:** Dio interceptor on 401 → deletes token → calls `AuthNotifier.logout()` → router redirects to `/auth/phone`
- **Token validation on startup:** `AuthNotifier.checkAuth()` reads token from storage → calls `GET /auth/me` → populates user or clears token
- **State management:** All auth state via Riverpod `StateNotifier` providers, no setState for business logic
