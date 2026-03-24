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
- **OTP storage:** code stored in Redis with 5-minute TTL; attempt counter stored with 5-minute TTL; both deleted on successful verify
- **Username format:** regex `^[a-zA-Za-яА-ЯёЁ0-9_]{3,20}$` — letters (Latin + Cyrillic), digits, underscore
- **Username cooldown:** can only change username once every 30 days (enforced against `updated_at`)
- **New user detection (mobile):** if `avatar_emoji == null || birth_year == null` after auth → redirect to profile setup
- **JWT:** HS256, 90-day expiry, claims: UserID, Username
- **Avatar:** max 5MB, JPEG/PNG only (validated by magic bytes: `FF D8 FF` for JPEG, `89 50 4E 47` for PNG), resized to 256x256 using `imaging.Fill` (Lanczos), stored as JPEG 85% in S3
- **Dev mode:** when `DEV_MODE=true`, OTP code is returned in the send response
- **Push preferences:** default to enabled if no preference record exists
- **Account deletion:** hard delete with CASCADE — removes all user data
- **App version:** `compareSemver` does simple numeric part comparison; `force_update` is true when client version is below `AppMinVersion`

## Mobile Screens

### Phone Screen (`/auth/phone`)
- Text field with `+7 (___) ___-__-__` mask (mask_text_input_formatter)
- "Получить код" button, disabled until 10 digits entered
- Apple/Google sign-in buttons (stubs, `onPressed: null`)
- Loading spinner inline in button during OTP send
- Error message display below text field

### OTP Screen (`/auth/otp`)
- Receives phone number via GoRouter `extra` parameter
- 6-digit Pinput input, auto-submits on completion (`onCompleted` callback)
- 5-minute countdown timer (Отправить повторно через MM:SS)
- "Отправить код повторно" button appears after timer expires; resets timer and clears pin on success
- `OtpVerifyNotifier.reset()` called on resend to clear previous error state
- Error message display for invalid code
- Navigation: success → router redirect to `/home` or `/auth/setup` (driven by `AuthNotifier.login`)

### Profile Setup Screen (`/auth/setup`)
- Shown only when `needsProfileSetup == true` (no avatar_emoji or no birth_year)
- Emoji avatar picker: 20 hardcoded emojis in a `Wrap` grid, tap to select (first emoji pre-selected)
- Username field with 500ms debounce availability check (green checkmark / red X / spinner suffix icon)
- Birth year field: digits only, 4 chars max (FilteringTextInputFormatter + LengthLimitingTextInputFormatter)
- "Готово" button, disabled until valid form (username >= 3 chars, birth year in range)
- Validation: birth year between `currentYear - 22` and `currentYear - 14`
- Navigation handled by router redirect after `AuthNotifier.profileCompleted`

### Home Screen (`/home`) — stub
- Peach emoji, welcome text, "Экран групп появится в T07" label
- Logout button (calls `AuthNotifier.logout()`)

### Navigation (go_router)
- Auth redirect: no token → `/auth/phone`
- Profile setup redirect: authenticated + needsProfileSetup → `/auth/setup`
- Authenticated + complete profile → `/home`
- Uses `_RouterNotifier extends ChangeNotifier` + `refreshListenable` pattern to avoid GoRouter recreation on every provider rebuild; only notifies when `status` or `needsProfileSetup` changes

## Architecture

### Backend

```
backend/
├── cmd/server/main.go                    # Entrypoint, route registration
├── internal/
│   ├── config/config.go                  # Env-based config (includes DevMode bool)
│   ├── handler/
│   │   ├── errors.go                     # ErrorResponse(), ErrorHandler()
│   │   ├── auth/handler.go               # 11 handler methods
│   │   └── auth/handler_test.go          # Handler-level tests (toUserDto, AppVersion, etc.)
│   ├── service/auth/service.go           # Business logic, JWT signing, S3 upload
│   └── service/auth/service_test.go      # Unit tests (isValidImage, usernameRegex, generateUsername)
│   ├── middleware/
│   │   ├── auth.go                       # JWTAuth middleware, GetCurrentUser()
│   │   ├── ratelimit.go                  # Redis-based rate limiter
│   │   └── validator.go                  # go-playground/validator wrapper
│   ├── db/
│   │   ├── migrations/001_init.up.sql    # Full schema (14 tables, 7 enums)
│   │   ├── queries/users.sql             # User CRUD queries
│   │   ├── queries/push_preferences.sql  # Push pref queries
│   │   └── sqlc/                         # Generated Go code (DO NOT EDIT)
│   └── lib/
│       ├── db.go                         # pgxpool connection
│       ├── redis.go                      # go-redis client
│       ├── s3.go                         # S3 upload client
│       └── asynq.go                      # Task queue client
```

### Mobile

```
mobile/lib/
├── main.dart                                         # App entrypoint, ProviderScope
├── core/
│   ├── api/
│   │   ├── api_client.dart                          # Dio factory, auth interceptor, parseError, AppException
│   │   └── api_service.dart                         # Raw API methods (Dio wrapper, no code generation)
│   ├── providers/
│   │   ├── api_provider.dart                        # Dio, ApiService, SecureStorage providers
│   │   └── auth_provider.dart                       # AuthNotifier, AuthState, AuthStatus
│   ├── router/app_router.dart                       # GoRouter with _RouterNotifier + refreshListenable
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

### Key Dependencies (Mobile)

- **Dio interceptor (onRequest):** reads token from `FlutterSecureStorage`, adds `Authorization: Bearer {token}` header
- **Dio interceptor (onError):** on 401 → deletes token from storage → calls `onUnauthorized` callback → `AuthNotifier.logout()` → router redirects to `/auth/phone`
- **Token validation on startup:** `AuthNotifier.checkAuth()` reads token from storage → calls `GET /auth/me` via raw Dio (not `ApiService`) → populates user or clears token
- **State management:** All auth state via Riverpod `StateNotifier` providers, no `setState` for business logic
- **ApiService** is a plain Dio wrapper (not retrofit/code-gen); methods return `Map<String, dynamic>`

### Mobile/Backend Path Alignment

All mobile API paths now match the backend routes. Previously mismatched paths were fixed:
- `GET /auth/username-check` (was `/auth/username/check`)
- `PATCH /auth/profile` (was `PUT`)
- `GET /app/version` (was `/auth/version`)

## Tests

### Backend (T04)
- `backend/internal/handler/auth/handler_test.go` — handler-level tests: `toUserDto` field mapping, `AppVersion` semver logic (6 cases), missing param / bad JSON error handling
- `backend/internal/service/auth/service_test.go` — unit tests: `isValidImage` magic byte detection (7 cases), `usernameRegex` (11 cases), `generateUsername` uniqueness and prefix

### Mobile (T05)
- 60 tests across 10 test files, 83% line coverage (excluding generated files)
- Mocking via `mocktail`; widget tests use `ProviderScope` overrides
- `test/core/api/api_client_test.dart` — `parseError` (API error, generic, timeout cases), `AppException.toString`
- `test/core/api/api_service_test.dart` — `ApiService` method calls
- `test/core/providers/auth_provider_test.dart` — `AuthNotifier` (checkAuth, login, logout, profileCompleted, needsProfileSetup detection)
- `test/core/router/app_router_test.dart` — redirect logic
- `test/features/auth/data/auth_repository_test.dart` — `AuthRepository` (sendOtp, verifyOtp, checkUsername, updateProfile; error propagation)
- `test/features/auth/presentation/auth_notifier_test.dart` — `OtpSendNotifier`, `OtpVerifyNotifier`, `ProfileSetupNotifier` (loading states, error states, success flows, reset)
- `test/features/auth/presentation/phone_screen_test.dart` — widget test for phone screen
- `test/features/auth/presentation/otp_screen_test.dart` — widget test for OTP screen
- `test/features/auth/presentation/profile_setup_screen_test.dart` — widget test for profile setup
- `test/features/home/home_screen_test.dart` — widget test for home screen stub
