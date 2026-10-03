# Infrastructure

## Overview

Monorepo foundation for the Repa app. Contains Docker Compose for local development, Go backend server setup with Echo framework, database schema and migrations, question seed data, security middleware, production Dockerfile, and the Flutter project scaffolding.

## API Endpoints

### `GET /api/v1/health`
Health check for monitoring.
- **Public** (no auth required)
- **Success 200:** `{ "data": { "status": "ok", "db": "ok", "redis": "ok" } }`
- **Checks:** PostgreSQL connectivity via `SELECT 1`, Redis connectivity via `PING`

### `GET /.well-known/apple-app-site-association`
iOS Universal Links configuration. Covers `/join/*` and `/invite/*` paths for app ID `3GWRDGA8B7.app.repa.repa`. No auth required.

### `GET /.well-known/assetlinks.json`
Android App Links configuration. Declares `app.repa.repa` package with SHA-256 cert fingerprint. No auth required.

## Data Model

### Full Schema (001_init.up.sql)

18 tables, 7 enums, 6 indexes.

**Enums:**
- `season_status` — VOTING, REVEALED, CLOSED
- `question_category` — HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY
- `question_source` — SYSTEM, USER
- `question_status` — ACTIVE, PENDING, REJECTED
- `crystal_log_type` — PURCHASE, SPEND_DETECTOR, SPEND_ATTRIBUTES, SPEND_QUESTION, BONUS
- `achievement_type` — 23 types (SNIPER, ORACLE, TELEPATH, BLIND, RANDOM, EXPERT_OF, BEST_FRIEND, DETECTIVE, STRANGER, LEGEND, CHANGEABLE, MONOPOLIST, ENIGMA, RISING, PIONEER, STREAK_VOTER, FIRST_VOTER, LAST_VOTER, NIGHT_OWL, ANALYST, MEDIA, CONSPIRATOR, RECRUITER)
- `push_category` — SEASON_START, REMINDER, REVEAL, REACTION, NEXT_SEASON

**Tables:** users, groups, group_members, seasons, questions, season_questions, votes, season_results, achievements, user_group_stats, detectors, crystal_logs, fcm_tokens, card_cache, reactions, reports, push_preferences, next_season_votes

### Seed Data (cmd/seed/main.go)

205 system questions across 6 categories:
- HOT: 40 questions
- FUNNY: 50 questions
- SECRETS: 40 questions
- SKILLS: 35 questions
- ROMANCE: 20 questions
- STUDY: 20 questions

All seeded as `source=SYSTEM`, `status=ACTIVE`, with `ON CONFLICT DO NOTHING`.

## Business Rules

- **Database convention:** All IDs are `TEXT` with `gen_random_uuid()::text` default
- **Cascade deletes:** All foreign keys use `ON DELETE CASCADE`
- **API response format:** `{ "data": { ... } }` for success, `{ "error": { "code": "...", "message": "..." } }` for errors
- **Dates:** All timestamps in ISO 8601 UTC (TIMESTAMPTZ)
- **sqlc:** Type-safe query generation — generated files in `internal/db/sqlc/` must never be edited manually

## Architecture

### Docker Compose

```yaml
services:
  postgres:  # postgres:16-alpine, port 5432, db=repa, user=repa, pass=repa
  redis:     # redis:7-alpine, port 6379
volumes:
  postgres_data:  # persistent PostgreSQL data
```

Both services have healthchecks (5s interval, 5s timeout, 5 retries).

### Dockerfile (backend/Dockerfile)

Multi-stage build targeting Go 1.26:

- **Build stage:** `golang:1.26-alpine` — downloads modules, builds binary with `CGO_ENABLED=0 -ldflags="-s -w"`
- **Runtime stage:** `alpine:3.20` with Chromium, NSS, FreeType, HarfBuzz, ttf-freefont, tzdata — Chromium is required for chromedp card rendering
- Sets `CHROME_PATH=/usr/bin/chromium-browser` and `TZ=UTC`
- Copies binary and `internal/db/migrations/` into the image
- Exposes port 3000

### Backend Server (cmd/server/main.go)

Startup sequence:
1. Load config from env (godotenv)
2. Connect to PostgreSQL (pgxpool via `lib.NewPool`)
3. Bridge pgxpool to `database/sql` via `lib.NewDBFromPool` (required by sqlc)
4. Connect to Redis (`lib.NewRedis`)
5. Initialize Asynq client (`lib.NewAsynqClient`)
6. Initialize optional external clients: S3, Firebase, Anthropic, Telegram, YuKassa (all skipped gracefully if not configured)
7. Create Echo instance with middleware stack
8. Register routes (public + protected groups)
9. Start Asynq worker goroutine (`startWorker`)
10. Start Asynq scheduler goroutine (`startScheduler`)
11. Start HTTP server on `PORT` (default 3000) with graceful shutdown on SIGINT/SIGTERM

**Middleware stack (in order):**
1. Recover
2. RequestID
3. CORS (allow all origins; methods GET/POST/PUT/PATCH/DELETE; headers Origin/Content-Type/Accept/Authorization)
4. Secure headers (XSS protection, nosniff, X-Frame-Options DENY, HSTS 1yr, CSP `default-src 'self'`)
5. Sanitize (global input sanitization — see Security section)
6. Custom error handler (`handler.ErrorHandler`)
7. Request validator (go-playground/validator v10)
8. JWT auth middleware (`appmw.JWTMiddleware`) — applied to `protected` route group only

### Config (internal/config/config.go)

All config loaded from environment via godotenv. Required fields (panic if missing): `DATABASE_URL`, `JWT_SECRET`. All others have defaults or are optional.

| Env Var | Default | Purpose |
|---------|---------|---------|
| `PORT` | `3000` | HTTP listen port |
| `DATABASE_URL` | — (required) | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection |
| `JWT_SECRET` | — (required) | JWT signing key |
| `S3_ENDPOINT` / `S3_BUCKET` / `S3_ACCESS_KEY` / `S3_SECRET_KEY` / `S3_REGION` | empty / `ru-central1` | Yandex Object Storage |
| `ANTHROPIC_API_KEY` | empty | Claude AI for question moderation |
| `YUKASSA_SHOP_ID` / `YUKASSA_SECRET_KEY` / `YUKASSA_RETURN_URL` | — | YuKassa payments |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_WEBHOOK_SECRET` | empty | Telegram bot |
| `APP_BASE_URL` | `https://repa.app` | Used in deep links and card URLs |
| `APP_MIN_VERSION` / `APP_LATEST_VERSION` | `1.0.0` | Version gating |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | `admin` / empty | Admin endpoints |
| `FIREBASE_PROJECT_ID` / `FIREBASE_PRIVATE_KEY` / `FIREBASE_CLIENT_EMAIL` | empty | FCM push notifications |
| `DEV_MODE` | `false` | Development mode flag |

### Makefile Targets

| Target | Command | Description |
|--------|---------|-------------|
| `dev` | `go run ./cmd/server` | Run server locally |
| `build` | `go build -o bin/server ./cmd/server` | Build binary |
| `migrate` | `migrate -path ... up` | Apply all migrations |
| `migrate-down` | `migrate -path ... down 1` | Rollback 1 migration |
| `sqlc` | `sqlc generate` | Regenerate Go code from SQL queries |
| `seed` | `go run ./cmd/seed` | Seed question bank (205 questions) |
| `test` | `go test ./...` | Run all tests |

### Asynq Worker (background jobs)

17 task type constants defined in `internal/lib/asynq.go`. All handlers are implemented in `internal/worker/tasks/`.

**Registered handlers:**
- `reveal:checker` — RevealChecker: polls every minute for seasons ready for reveal
- `reveal:process` — RevealProcessor: runs the reveal for a season, fans out to achievements/cards/push/telegram tasks
- `achievements:calculate` — AchievementsProcessor
- `cards:generate` — CardsProcessor (uses chromedp/Chromium)
- `push:weekly-scheduler`, `push:tuesday-signal`, `push:wednesday-quorum`, `push:thursday-teaser`, `push:friday-pre-reveal`, `push:reveal-notification`, `push:sunday-preview`, `push:sunday-streak`, `push:reaction` — PushProcessor
- `telegram:season-start`, `telegram:reveal-post`, `telegram:share-card` — TelegramProcessor (only registered if `TELEGRAM_BOT_TOKEN` is configured)

**Asynq Scheduler (cron):**

| Cron | Task | Queue | Description |
|------|------|-------|-------------|
| `* * * * *` | `reveal:checker` | critical | Every minute |
| `0 14 * * 1` | `push:weekly-scheduler` | default | Mon 17:00 MSK |
| `0 16 * * 2` | `push:tuesday-signal` | default | Tue 19:00 MSK |
| `0 15 * * 3` | `push:wednesday-quorum` | default | Wed 18:00 MSK |
| `0 17 * * 4` | `push:thursday-teaser` | default | Thu 20:00 MSK |
| `0 16 * * 5` | `push:friday-pre-reveal` | default | Fri 19:00 MSK |
| `0 9 * * 0` | `push:sunday-preview` | default | Sun 12:00 MSK |
| `0 15 * * 0` | `push:sunday-streak` | default | Sun 18:00 MSK |
| `0 18 * * 0` | `season:creator` | critical | Sun 21:00 MSK |
| `0 14 * * 1` | `telegram:season-start` | default | Mon 17:00 MSK |

Worker configuration: 3 queues (critical:6, default:3, low:1), concurrency 10.

### External Clients (internal/lib/)

| File | Client | Purpose |
|------|--------|---------|
| `db.go` | pgxpool + sql.DB bridge | PostgreSQL connection pool |
| `redis.go` | go-redis/v9 | Redis client |
| `asynq.go` | asynq | Task queue client + server + 17 task type constants |
| `s3.go` | aws-sdk-go-v2 | Yandex Object Storage (S3-compatible) for avatar uploads |
| `firebase.go` | firebase-admin/go v4 | FCM push notifications |
| `anthropic.go` | Anthropic API | Claude AI for question moderation |
| `telegram.go` | Telegram Bot API | Telegram group integration |
| `yukassa.go` | YuKassa API | Payment processing |

## Security

### Input Sanitization (internal/middleware/sanitize.go)

Global middleware applied to all routes. Processes JSON request bodies:
- Rejects requests containing null bytes (returns 400 VALIDATION)
- Trims whitespace from all string values
- Strips HTML tags using regex `<[^>]*>|<[a-zA-Z][^>]*$` (handles both well-formed and malformed tags like `<script src=x`)
- Recursively processes nested objects and arrays

### Rate Limiting (internal/middleware/ratelimit.go)

Redis-backed sliding window rate limiter, applied per-route per-user. Per-route limits:
- `POST /seasons/:seasonId/votes` — 30 requests/minute
- `POST /groups/:groupId/questions` — 10 requests/hour
- `GET /seasons/:seasonId/my-card-url` — 5 requests/hour

### YuKassa IP Allowlist (internal/middleware/yukassa.go)

Applied to `POST /api/v1/crystals/purchase/webhook` (no JWT on this route — called by YuKassa). Only allows requests from official YuKassa CIDR ranges:
- `185.71.76.0/27`, `185.71.77.0/27`
- `77.75.153.0/25`, `77.75.156.11/32`, `77.75.156.35/32`, `77.75.154.128/25`
- `2a02:5180::/32`

Returns 403 FORBIDDEN for any other IP.

### Anonymity Tests (internal/handler/voting/anonymity_test.go)

Test suite enforcing anonymity guarantees:
- Source scan: verifies no file in `internal/` exposes `voter_id` in JSON output tags
- DTO reflection: checks all vote-related response structs have no `voter_id` field
- Detector binding check: verifies detector response contains only a list of voter IDs without question/answer binding
- Progress endpoint coverage: included in anonymity scans
- Service source scan: scans service layer for accidental voter ID exposure

## Flutter Project (mobile/)

Created with `flutter create --org app.repa --project-name repa`.

**Runtime dependencies:**
- flutter_riverpod ^2.5.1 — state management
- go_router ^13.2.4 — navigation
- dio ^5.4.3 — HTTP client
- freezed_annotation ^2.4.1 + json_annotation ^4.9.0 — code generation for models
- flutter_secure_storage ^9.0.0 — token storage
- pinput ^3.0.1 — OTP input
- mask_text_input_formatter ^2.9.0 — phone mask
- flutter_animate ^4.5.0 — animations
- cached_network_image ^3.3.1 — image caching
- firebase_crashlytics — crash reporting

**Dev dependencies:** build_runner, freezed, json_serializable, mocktail

**Note:** retrofit/retrofit_generator were dropped — generator is broken with Dart 3.11. Using plain Dio methods instead.

### API Base URL Configuration

`mobile/lib/core/api/env.dart` reads `API_BASE_URL` from `--dart-define` at build time, defaulting to `http://localhost:3000/api/v1`. Pass `--dart-define=API_BASE_URL=https://api.repa.app/api/v1` for production builds.

### iOS Release Config

- `Runner.entitlements` / `RunnerRelease.entitlements`: APNs (`aps-environment`) and Associated Domains (`applinks:repa.app`) entitlements
- `Info.plist`: portrait-only orientation lock, camera/photo library privacy strings
- Bundle ID: `app.repa.repa`

### Android Release Config

- `AndroidManifest.xml`: App Links intent-filter for `repa.app` covering `/join/*` and `/invite/*`
- `build.gradle.kts`: release signing via `key.properties`, Google Services and Firebase Crashlytics Gradle plugins, ProGuard enabled
- `proguard-rules.pro`: keep rules for Dio, Riverpod, json_serializable

## Dependency Security

Dependabot vulnerabilities resolved (reflected in current `go.mod`):

| Package | Version | Severity | Issue |
|---------|---------|----------|-------|
| `golang.org/x/image` | v0.38.0 | High + Medium | Panic parsing palette-color images; uncontrolled resource consumption in TIFF decoder |
| `google.golang.org/grpc` | v1.79.3 | Critical | Auth bypass |
| `go.opentelemetry.io/otel/sdk` | v1.42.0 | High | PATH hijacking |

One unresolved alert remains: `disintegration/imaging` v1.6.2 TIFF crash (low severity) — no upstream fix available.


## Dev database connection (repa-db-dev)

Yandex Cloud Managed PostgreSQL. Credentials in `secrets/pg_credentials.env` (`PG_HOST` holds **two**
comma-separated hosts), `DATABASE_URL` in `backend/.env`.

Two things make a naive connection fail in ways that look like something else:

- **`target_session_attrs=read-write` is required.** With two hosts and no preference the driver may land
  on the replica, where every DDL statement fails with `cannot execute CREATE TYPE in a read-only
  transaction`. That reads like a permissions problem and is not one.
  `SELECT pg_is_in_recovery()` confirms which node you reached.
- **`sslmode=verify-full` needs the Yandex CA** at `~/.postgresql/root.crt`
  (`https://storage.yandexcloud.net/cloud-certs/CA.pem`, `chmod 0600`). Do not lower `sslmode` to get
  past a missing certificate.

If `psql` is not installed locally, run it from a container with the CA mounted read-only, passing the
password through `PGPASSWORD` so it never reaches a command line or a log:

```
docker run --rm -i \
  -v "$HOME/.postgresql/root.crt:/ca.crt:ro" \
  -e PGPASSWORD=... -e PGSSLROOTCERT=/ca.crt \
  postgres:16-alpine \
  psql "host=$PG_HOST port=6432 user=$PG_USER dbname=repa-db-dev sslmode=verify-full target_session_attrs=read-write" \
  -v ON_ERROR_STOP=1 --single-transaction -f /in.sql
```

Apply migrations one file at a time with `ON_ERROR_STOP=1 --single-transaction`, so a failure rolls back
rather than leaving the database half-migrated.

**Touch only `repa-db-dev`** — never the cluster, never another database.

### Migration ledger

golang-migrate tracks applied versions in `schema_migrations(version, dirty)` — one row, the version being
the integer prefix of the highest applied file (so `010_question_tone` is version **10**, not `010`).

Until 2026-10-02 that table **did not exist** on `repa-db-dev`: 001–003 were applied by hand before the
post-MVP work and 004–010 by hand during it. An absent ledger reads to golang-migrate as "nothing
applied", so `make migrate` would have started from `001_init.up.sql`, failed on `CREATE TYPE ... already
exists`, and marked the version **dirty** — after which every later `migrate` refuses to run until a human
intervenes. The ledger was reconciled with `force 10`; `migrate up` now reports `no change`.

Use the wrapper rather than `make migrate` for this cluster:

```
backend/scripts/migrate-dev.sh version     # what the ledger says  (make migrate-dev-status)
backend/scripts/migrate-dev.sh up          # apply anything pending (make migrate-dev)
backend/scripts/migrate-dev.sh force 10    # reconcile: set the version, run no DDL
```

It resolves the read-write host, mounts the CA, and runs `migrate/migrate:v4.17.1` from its own container.
Three things it exists to handle, each of which fails as something else:

- **Two hosts, one driver slot.** golang-migrate uses `lib/pq`, which takes a single host in its URL, so
  `target_session_attrs` cannot help. The script asks each host `SELECT pg_is_in_recovery()` and uses the
  one that answers `f`. It detects rather than hardcodes because a managed cluster fails over.
- **`migrate version` is not read-only.** It creates `schema_migrations` before reporting `no migration`.
  "Just checking" already writes.
- **`force N` runs no DDL.** It is the tool for *the schema is already at N by other means* — which is
  what reconciliation is. It is **not** a fix for a failed migration: forcing a version the schema has not
  reached makes `migrate up` skip real migrations silently. Verify the schema first.

A **dirty** ledger is what a half-applied migration leaves behind. `make migrate-dev-status` names it;
resolving it is a decision about what actually got applied, not something to automate.

### Rules going forward

- Applying a migration by hand means reconciling the ledger **in the same session**, or the next person
  inherits the trap that was just removed.
- `make migrate-dev-status` before and after any schema work.
- `make sqlc-check` fails when `internal/db/sqlc` does not match the `.sql` files. A stale generated
  package compiles and silently uses the wrong query shape, so nothing else catches it.


## CI

`.github/workflows/ci.yml`. Jobs, pins and local equivalents are tabulated in `CLAUDE.md` → **CI**; this
section records where each pin comes from and why the mobile job ends the way it does.

| Pin | Value | Source of truth |
| --- | --- | --- |
| `GO_VERSION` | 1.26.1 | `backend/go.mod` → `go 1.26.1` |
| `FLUTTER_VERSION` | 3.47.6 | satisfies `mobile/pubspec.yaml` → `sdk: ^3.11.1` |
| `SQLC_VERSION` | v1.31.1 | the version `internal/db/sqlc` was generated with |

Pins are deliberate. A toolchain that floats to `latest` turns an unrelated upstream release into a red
build on someone's unrelated pull request, and the first instinct is to distrust CI rather than the pin.
Dependabot already watches the Go modules in this repo and can raise these the same way.

### The mobile job's last step

`flutter analyze` rewrites `mobile/analysis_options.yaml` on every invocation ("Upgrading
analysis_options.yaml to exclude build and platform directories"). That one file is restored from the index
as a known side effect. Then the job fails if anything **else** is modified.

Measured, not assumed: with the lockfile current, `pub get` + `analyze` + `test` modify exactly one tracked
file, `analysis_options.yaml`. With the lockfile as committed at the time of writing, `pub get` moves four
transitive versions — which is the drift the check exists to surface. A lockfile that does not describe what
the pinned SDK resolves defeats the purpose of having one.

The step uses `if: always()`, so a failing test does not also conceal a dirty tree.

### Reproducing a CI failure

Each row in the `CLAUDE.md` table is the whole command. Two notes:

- `make test-e2e` needs a Docker daemon; it sets `TESTCONTAINERS_RYUK_DISABLED=true`, as CI does.
- `make sqlc-check` compares `internal/db/sqlc` against itself across a regeneration rather than against
  git HEAD, so it gives the same answer with uncommitted work in the tree.
