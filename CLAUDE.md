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

## Task Specs

Implementation is organized in phases (T01–T26) documented in `docs/specs/`. The `docs/specs/master_context.md` file contains the complete context including DB schema, API conventions, and business rules.
