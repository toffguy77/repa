# Settings, Analytics, Assets & Admin

## Settings Screen (`/settings`)
- Accessible from Profile tab via gear icon in AppBar
- Sections: Profile (avatar, username, birth year), Push preferences, Account (logout, delete), About (version, privacy, terms)
- Avatar upload via image_picker (camera/gallery), uploaded to S3 via `POST /auth/avatar`
- Push preferences toggled per category via `PATCH /push/preferences`
- Delete account with confirmation dialog via `DELETE /auth/account`

## Analytics (Firebase Analytics)
- `AnalyticsService` in `core/analytics/analytics_service.dart`
- Events: login, signup, group_created, group_joined, voting_started, voting_completed, reveal_opened, card_shared, detector_purchased, purchase_initiated, purchase_completed, achievement_unlocked, reaction_sent, question_voted
- Wired at screen level via `ref.read(analyticsProvider)`

## App Config / Flavors
- `AppConfig` in `core/config/app_config.dart` with dev/staging/prod configs
- Selected via `--dart-define=FLAVOR=dev|staging|prod`
- `Env` class reads flavor and exposes `apiBaseUrl` and `appUrl`

## Force Update
- `ForceUpdateScreen` — blocking full-screen when app version < min_version
- Backend: `GET /app/version` returns min_version, latest_version, force_update flag

## Splash Screen
- Config: `flutter_native_splash.yaml` — dark purple (#1A0533) background
- Assets needed: `assets/splash/logo.png` (eggplant + REPA logo, white on transparent)
- Firebase init runs during splash hold in `main.dart`

## App Icon
- Config: `flutter_launcher_icons.yaml` — dark purple adaptive background
- Assets needed: `assets/icon/icon.png` (1024x1024) and `assets/icon/icon_foreground.png`

## Admin Endpoints (Basic Auth)
- `GET /api/v1/admin/reports?page=1&limit=20` — paginated reports with question text, category, status, reporter
- `PATCH /api/v1/admin/reports/:id` — `{ "action": "approve" | "reject" }` updates question status
- `GET /api/v1/admin/stats` — DAU (7d), MAU (30d), groups count, revenue (7d)
- Auth: HTTP Basic Auth with `ADMIN_USERNAME` / `ADMIN_PASSWORD` env vars

## Backend Changes
- New SQL queries in `internal/db/queries/admin.sql`
- New handler in `internal/handler/admin/handler.go`
- Routes registered under `/api/v1/admin` group with BasicAuth middleware

## New Flutter Packages
- `firebase_analytics` — event tracking
- `image_picker` — avatar photo selection
- `package_info_plus` — app version display
- `flutter_launcher_icons` (dev) — icon generation
- `flutter_native_splash` (dev) — splash screen generation
