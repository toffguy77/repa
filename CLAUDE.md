# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Repa (Репа) — a mobile app (Flutter, iOS + Android) where users join private groups and anonymously vote on each other weekly by fun questions. Friday at 20:00 MSK is "Reveal" — each user sees their reputation card. Target audience: Russian students 14–22.

Monorepo with two workspaces: `backend/` (Go) and `mobile/` (Flutter).

## Tech Stack

**Backend:** Go 1.22+, Echo v4, PostgreSQL 16 (sqlc for queries, golang-migrate for migrations), Redis 7 (go-redis, asynq for job queues), zerolog, JWT (golang-jwt/v5), validator/v10.

**Mobile:** Flutter 3.x / Dart 3, Riverpod 2, go_router, Dio + retrofit, Freezed + json_serializable, flutter_animate.

## Common Commands

```bash
# Infrastructure
docker compose up -d                # Start postgres + redis

# Backend (from backend/)
make dev                            # Run server (go run ./cmd/server)
make build                          # Build binary to bin/server
make migrate                        # Apply migrations (needs DATABASE_URL)
make migrate-down                   # Roll back 1 migration
make sqlc                           # Regenerate Go code from SQL queries
make seed                           # Seed question bank (go run ./cmd/seed)
make test                           # go test ./...

# Mobile (from mobile/)
flutter analyze                     # Lint
flutter test                        # Run all tests
flutter test test/path_to_test.dart # Single test file
flutter build ios --no-codesign     # Verify iOS build
dart run build_runner build --delete-conflicting-outputs  # Code generation (freezed, retrofit, json_serializable)
```

## Architecture

### Backend (`backend/`)

Layered architecture per feature:
- `cmd/server/main.go` — entrypoint
- `internal/config/` — env-based config (godotenv)
- `internal/handler/{feature}/` — Echo HTTP handlers (routing + validation)
- `internal/service/{feature}/` — business logic
- `internal/db/migrations/` — SQL migration files (golang-migrate)
- `internal/db/queries/` — SQL queries for sqlc
- `internal/db/sqlc/` — **auto-generated, never edit manually** — run `make sqlc` to regenerate
- `internal/worker/` — asynq task handlers (reveal-checker, season-creator, push scheduler)
- `internal/middleware/` — auth, rate limiting, security
- `internal/lib/` — external client singletons (redis, firebase, s3, telegram, asynq)

### Mobile (`mobile/`)

Feature-first clean architecture:
- `lib/core/` — shared: API client (Dio), router (go_router), theme, global Riverpod providers
- `lib/features/{feature}/data/` — repository, API models
- `lib/features/{feature}/domain/` — use cases, entities (freezed)
- `lib/features/{feature}/presentation/` — screens, widgets, notifiers

## Key Conventions

- **Specs are source of truth:** Always read `docs/specs/master_context.md` and relevant `docs/specs/T*.md` before implementing. Update specs when behavior changes.
- **API format:** All responses use `{ "data": { ... } }` or `{ "error": { "code": "...", "message": "..." } }`. Base URL: `/api/v1`.
- **Anonymity is critical:** `votes` table stores `voter_id` (needed for detector), but API responses NEVER expose `voter_id` in connection to specific votes. Detector returns only a list of voter IDs, without question/answer binding.
- **Crystal balance:** Computed as `SUM(delta)` from `crystal_logs` — no separate balance field.
- **No GORM:** Only sqlc for type-safe DB queries.
- **UI language:** Hardcoded Russian strings (no arb/l10n files in MVP).
- **State management:** Riverpod only. No `setState` for business logic.
- **Domain models:** Freezed for all entities and API responses.
- **Color accent:** `#7C3AED` (purple), system font, white/dark system background.

## Business Rules

- Reveal: Friday 17:00 UTC (20:00 MSK). Requires quorum: >=50% voted (>=40% for groups <8 members).
- Groups: 5–50 members, max 10 groups per user, activates at >=3 members.
- Push: Max 3/day per user, quiet hours 23:00–09:00 MSK.
- Seasons: First created with group; subsequent ones auto-created Sunday 18:00 UTC via asynq job.
- Users under 18: ROMANCE category unavailable.
- Detector costs 10 crystals, hidden attributes cost 5 crystals.

## Database (Dev)

- Dev database is a remote Yandex Cloud Managed PostgreSQL cluster. **Never touch the cluster itself or other databases — only the `repa-db-dev` database.**
- Credentials and connection details are in `secrets/pg_credentials.env`. The `backend/.env` has `DATABASE_URL` configured.
- CA cert: `~/.postgresql/root.crt` (Yandex Cloud CA).
- When implementing features that require schema changes:
  1. Create new migration files in `internal/db/migrations/`.
  2. Apply migrations to the dev DB automatically using `psql -f` (the `make migrate` DSN format may not work with multi-host + special chars — use the psql connection string from `secrets/pg_credentials.env`).
  3. Run `make sqlc` to regenerate Go code after migration changes.
  4. If a migration fails — fix it. **Never leave the dev database in a broken state.**
- When running seed or any DB command, use the credentials from `backend/.env` or `secrets/pg_credentials.env`.

## Secrets

- All secrets (API keys, credentials, tokens) go into the `secrets/` directory at the repo root.
- `secrets/` **must** be in `.gitignore` — verify before committing. Never commit secrets to git.
- As new secrets appear during development (e.g., Firebase keys, YuKassa creds, Telegram token), save them to a descriptive file in `secrets/`.

## Task Specs

Implementation is organized in phases (T01–T26) documented in `docs/specs/`. The `docs/specs/master_context.md` file contains the complete context including DB schema, API conventions, and business rules.
