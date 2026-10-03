# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Repa (Репа) — a mobile app (Flutter, iOS + Android) where users join private groups and anonymously vote on each other weekly by fun questions. Friday at 20:00 MSK is "Reveal" — each user sees their reputation card. Target audience: Russian students 14–22.

Monorepo with two workspaces: `backend/` (Go) and `mobile/` (Flutter).

## Tech Stack

**Backend:** Go 1.22+, Echo v4, PostgreSQL 16 (sqlc for queries, golang-migrate for migrations), Redis 7 (go-redis, asynq for job queues), zerolog, JWT (golang-jwt/v5), validator/v10.

**Mobile:** Flutter 3.x / Dart 3, Riverpod 2, go_router, Dio + retrofit, Freezed + json_serializable, flutter_animate.

## CI

`.github/workflows/ci.yml` runs on every pull request and on pushes to `main`. Five independent jobs, each
reproducible locally with one command:

| Job | What it checks | Reproduce locally |
| --- | --- | --- |
| `backend` | `go vet`, all packages except e2e | `cd backend && go vet ./... && make test-unit` |
| `backend-e2e` | the testcontainers suite against real Postgres + Redis | `cd backend && make test-e2e` |
| `mobile` | `flutter analyze`, `flutter test` | `cd mobile && flutter analyze && flutter test` |
| `generated-code` | `internal/db/sqlc` matches the `.sql` files | `cd backend && make sqlc-check` |
| `secrets` | nothing under `secrets/` is tracked | `git ls-files secrets/` (must be empty) |

Toolchain versions are pinned in the workflow's `env` block to what the project uses — Go 1.26.1 from
`go.mod`, Flutter 3.47.6, sqlc v1.31.1. Raise them as a deliberate change, not by floating to `latest`.

The mobile job ends by restoring `mobile/analysis_options.yaml` — `flutter analyze` rewrites it every run —
and then failing if the tree is **otherwise** dirty. The case that catches: `flutter pub get` moving
`pubspec.lock` means the committed lock does not describe what the pinned SDK resolves, so builds are not
reproducible. Commit the lock, or fix the pin.

e2e runs on every PR deliberately: a sqlmock fixture returns whatever shape the test author wrote, so a
query that no longer matches the schema passes a unit test. Only real Postgres catches that.

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
- `internal/schedule/`, `internal/eligibility/`, `internal/cardorder/` — rule packages with no
  dependencies on handlers or services, so the worker that acts on a rule and the API that explains it
  cannot drift apart. See **Business Rules**.

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
- **Anonymity is about inference, not just fields.** Removing `voter_id` from a response is necessary and
  not sufficient: a *derived* number can reconstruct a vote when combined with something else the API
  already publishes. Per-question winners are public (reveal summary, group chronicle), so any
  per-member accuracy figure placed next to them narrows or pins how that member voted — and a figure
  that is a rolling average can be differenced across two periods to recover a single period's count.
  Before adding any per-member statistic, ask what it can be combined with. Prefer publishing an
  **order** (a rank) over a **measurement** (a percentage): a rank cannot be arithmetically reduced to
  a vote.
- **Crystal balance:** Computed as `SUM(delta)` from `crystal_logs` — no separate balance field.
- **No GORM:** Only sqlc for type-safe DB queries.
- **UI language:** Hardcoded Russian strings (no arb/l10n files in MVP).
- **State management:** Riverpod only. No `setState` for business logic.
- **Domain models:** Freezed for all entities and API responses.
- **Color accent:** `#7C3AED` (purple), system font, white/dark system background.

## Business Rules

Three Go packages own rules that the app also has to explain. Never restate a number from them in a
handler, a client, or a doc — read it from the package, or the explanation will drift from the behaviour:

- `internal/schedule` — season timing
- `internal/eligibility` — whether a season may reveal, and what the app says while it may not
- `internal/cardorder` — the order a card's attributes appear in

### Seasons and Reveal

- A group's **first** season is a *kickoff* (`seasons.kind = 'KICKOFF'`): open immediately, revealing one
  hour after the group first satisfies every reveal rule — not on the next Friday.
- Weekly seasons reveal on the next Friday 20:00 MSK (17:00 UTC) that is at least 48 hours after the
  season opened (`schedule.MinVotingWindow`).
- **Absolute floors: >=3 members and >=3 completed voters** (`eligibility.MinMembers`, `MinVoters`).
  Never revealed below them, not even via the retry path.
- Quorum on top of the floors: >=50% voted (>=40% for groups <8).
- A season that reaches its reveal time and cannot reveal is **postponed** to the next Friday
  (`seasons.postpone_count`); votes already cast are kept.
- `REVEALED` becomes `CLOSED` when the next season opens, so a group has at most one `REVEALED` season.
  **Any query over a group's past must match `status IN ('REVEALED', 'CLOSED')`** — filtering on
  `REVEALED` alone returns exactly one season at any history length, and looks correct while doing it.
  The sole exception is `GetRevealedSeasonsForGroup`, which *is* the closing step.

### Groups

- 5–50 members, max 10 groups per user, activates at >=3 members.
- **Size thresholds** (`eligibility.NextGrowthThreshold`): 3 — a season can reveal at all; 5 — the
  detector becomes purchasable and percentages stop identifying voters. Nothing above 5 changes
  anything, and the app says so rather than naming a further target. The quorum share rising at 8
  members is **not** a threshold: it makes revealing harder.
- The detector/anonymity threshold counts members the reader can be rated by (membership minus blocks
  **within this group**); the reveal threshold counts actual membership, as `eligibility.Evaluate` does.
- Admin privileges: group name, Telegram link, custom questions, the kind-only setting, banning a member.

### Questions

- Every question carries a tone: `WARM`, `NEUTRAL`, `EDGY` (`internal/service/questions/tone.go`).
- `groups.kind_only` excludes `EDGY` from season selection. Defaults to on for a creator under 18 — an
  unknown birth year counts as under 18, as it does for `ROMANCE` — and an explicit value always wins.
- The setting applies **from the next season**: rewriting an open one would change the questions under
  members who already answered.
- `HOT` and `SECRETS` are entirely `EDGY`, so kind-only with only those categories is refused
  (`NO_KIND_CATEGORIES`) rather than producing an empty season.
- Users under 18: `ROMANCE` category unavailable.

### Crystals

- Balance is `SUM(delta)` over `crystal_logs`; grants are idempotent via `crystal_logs.external_id`.
- **The detector is a ladder, not one purchase:** a free rung (how many voted, no names), then one voter
  at a time (`DetectorHintCost`), then the full list (`DetectorFullCost`). Revealed hints live in
  `detector_hints` with `UNIQUE(user_id, season_id, revealed_user_id)`, so the same person is never
  charged twice.
- **No rung** is purchasable below `eligibility.MinDetectorMembers` — a list of "everyone but you" is
  not anonymous, so it must not be sold.
- Hidden attributes cost 5 crystals.
- Free sources exist (welcome grant, referral reward) because a paid mechanic nobody can try does not
  monetise.

### Push

- Max 3/day per user, quiet hours 23:00–09:00 MSK.
- A push must only promise what the screen it opens will actually contain.

## Database (Dev)

- Dev database is a remote Yandex Cloud Managed PostgreSQL cluster. **Never touch the cluster itself or other databases — only the `repa-db-dev` database.**
- Credentials and connection details are in `secrets/pg_credentials.env`. The `backend/.env` has `DATABASE_URL` configured.
- CA cert: `~/.postgresql/root.crt` (Yandex Cloud CA).
- When implementing features that require schema changes:
  1. Create new migration files in `internal/db/migrations/`.
  2. Apply migrations to the dev DB with `make migrate-dev` (wrapper: `backend/scripts/migrate-dev.sh`).
     It resolves the read-write host — the cluster has two, and the replica fails every DDL statement as
     "read-only transaction" — mounts the Yandex CA, and runs `migrate` from a pinned container, so
     nothing needs installing. Plain `make migrate` is for a single-host `DATABASE_URL`.
  2a. **Migrations are tracked.** `make migrate-dev-status` reports the applied version. If you ever apply
     a migration by hand, reconcile the ledger in the same session (`migrate-dev.sh force <N>`) — an
     absent or stale ledger makes the next `migrate up` start from 001 and mark itself dirty. See
     `docs/features/infrastructure.md` → **Migration ledger**.
  3. Run `make sqlc` to regenerate Go code after migration changes. `make sqlc-check` verifies the
     generated package matches the `.sql` files — a stale one compiles and silently uses the wrong query
     shape.
  4. If a migration fails — fix it. **Never leave the dev database in a broken state.**
- When running seed or any DB command, use the credentials from `backend/.env` or `secrets/pg_credentials.env`.

## Secrets

- All secrets (API keys, credentials, tokens) go into the `secrets/` directory at the repo root.
- `secrets/` **must** be in `.gitignore` — verify before committing. Never commit secrets to git.
- As new secrets appear during development (e.g., Firebase keys, YuKassa creds, Telegram token), save them to a descriptive file in `secrets/`.

## Task Specs

Implementation is organized in phases (T01–T26) documented in `docs/specs/`. The `docs/specs/master_context.md` file contains the complete context including DB schema, API conventions, and business rules.

Work after T26 is organized as OpenSpec changes rather than T-numbers. Archived changes live in
`openspec/changes/archive/`; the current requirements they produced are in `openspec/specs/`. When a
change alters behaviour, the requirement belongs in `openspec/specs/`, the behaviour in
`docs/features/`, and the context in `docs/specs/master_context.md` — all three, not one of them.
