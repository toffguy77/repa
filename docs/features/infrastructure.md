# Infrastructure

## Overview

Monorepo foundation for the Repa app. Contains Docker Compose for local development, Go backend server setup with Echo framework, database schema and migrations, question seed data, and the Flutter project scaffolding.

## API Endpoints

### `GET /api/v1/health`
Health check for monitoring.
- **Public** (no auth required)
- **Success 200:** `{ "data": { "status": "ok", "db": "ok", "redis": "ok" } }`
- **Checks:** PostgreSQL connectivity via `SELECT 1`, Redis connectivity via `PING`

## Data Model

### Full Schema (001_init.up.sql)

14 tables, 7 enums, 6 indexes. See `docs/specs/master_context.md` section 4 for complete SQL.

**Enums:**
- `season_status` — VOTING, REVEALED, CLOSED
- `question_category` — HOT, FUNNY, SECRETS, SKILLS, ROMANCE, STUDY
- `question_source` — SYSTEM, USER
- `question_status` — ACTIVE, PENDING, REJECTED
- `crystal_log_type` — PURCHASE, SPEND_DETECTOR, SPEND_ATTRIBUTES, SPEND_QUESTION, BONUS
- `achievement_type` — 23 types
- `push_category` — SEASON_START, REMINDER, REVEAL, REACTION, NEXT_SEASON

**Tables:** users, groups, group_members, seasons, questions, season_questions, votes, season_results, achievements, user_group_stats, detectors, crystal_logs, fcm_tokens, card_cache, reactions, reports, push_preferences, next_season_votes

### Seed Data (cmd/seed/main.go)

185 system questions across 6 categories:
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

### Backend Server (cmd/server/main.go)

Startup sequence:
1. Load config from env (godotenv)
2. Connect to PostgreSQL (pgxpool)
3. Connect to Redis (go-redis)
4. Initialize S3 client (optional — avatar uploads disabled if not configured)
5. Initialize Asynq client + server (15 task types registered, worker goroutine started)
6. Create Echo instance with middleware stack
7. Register routes (public + protected groups)
8. Start HTTP server on `PORT` (default 3000)

**Middleware stack (in order):**
1. Logger (zerolog)
2. Recover
3. CORS (allow all origins in dev)
4. Security headers
5. Custom error handler
6. Request validator (go-playground/validator)

### Makefile Targets

| Target | Command | Description |
|--------|---------|-------------|
| `dev` | `go run ./cmd/server` | Run server locally |
| `build` | `go build -o bin/server ./cmd/server` | Build binary |
| `migrate` | `migrate -path ... up` | Apply all migrations |
| `migrate-down` | `migrate -path ... down 1` | Rollback 1 migration |
| `sqlc` | `sqlc generate` | Regenerate Go code from SQL queries |
| `seed` | `go run ./cmd/seed` | Seed question bank (185 questions) |
| `test` | `go test ./...` | Run all tests |

### Asynq Worker (background jobs)

15 task types registered (implementations pending for future tasks):
- **Reveal:** TypeRevealChecker, TypeRevealProcess
- **Seasons:** TypeSeasonCreator
- **Achievements:** TypeAchievements
- **Push notifications:** 7 scheduled push types (weekly, daily, pre-reveal, etc.)
- **Telegram:** TypeTelegramStart, TypeTelegramReveal, TypeTelegramShare
- **Reactions:** TypeReactionPush

Worker configuration: 3 queues (critical:6, default:3, low:1), concurrency 10.

### External Clients (internal/lib/)

| Client | File | Purpose |
|--------|------|---------|
| PostgreSQL | `db.go` | pgxpool connection pool |
| Redis | `redis.go` | go-redis/v9 client |
| S3 | `s3.go` | Yandex Object Storage (S3-compatible) |
| Asynq | `asynq.go` | Redis-based task queue client + server |

### Flutter Project (mobile/)

Created with `flutter create --org app.repa --project-name repa`.

**Dependencies:**
- flutter_riverpod ^2.5.1 — state management
- go_router ^13.2.4 — navigation
- dio ^5.4.3 — HTTP client
- freezed_annotation ^2.4.1 + json_annotation ^4.9.0 — code generation for models
- flutter_secure_storage ^9.0.0 — token storage
- pinput ^3.0.1 — OTP input
- mask_text_input_formatter ^2.9.0 — phone mask
- flutter_animate ^4.5.0 — animations
- cached_network_image ^3.3.1 — image caching

**Dev dependencies:** build_runner, freezed, json_serializable, mocktail

**Note:** retrofit/retrofit_generator were dropped — generator is broken with Dart 3.11. Using plain Dio methods instead.
