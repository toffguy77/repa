# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Repa ("Репа") — a mobile app (Flutter, iOS + Android) where users in closed groups anonymously vote on each other weekly via funny questions. Friday 20:00 MSK = Reveal: everyone sees their reputation card. Audience: Russian students 14–22.

## Architecture

Monorepo with two top-level directories:

- `backend/` — Node.js 22, Fastify 4, TypeScript strict, Prisma ORM, PostgreSQL 16, Redis 7 (ioredis + BullMQ)
- `mobile/` — Flutter 3.x, Dart 3, Riverpod 2, go_router, Dio + retrofit, Freezed

Backend modules live in `backend/src/modules/{module}/` with `*.router.ts`, `*.service.ts`, `*.schema.ts` (Zod), `*.test.ts`.
Infrastructure clients in `backend/src/lib/` (prisma, redis, bullmq, firebase, s3, anthropic, telegram).
BullMQ workers in `backend/src/jobs/`.

Mobile follows feature-first: `mobile/lib/features/{feature}/` with `data/`, `domain/`, `presentation/` layers.
Core utilities in `mobile/lib/core/` (api, router, theme, providers).

## Commands

```bash
# Backend
cd backend && npm install
cd backend && npm run lint
cd backend && npm test                    # Vitest
cd backend && npx prisma migrate dev      # apply migrations
cd backend && npx prisma db seed          # seed questions

# Mobile
cd mobile && flutter pub get
cd mobile && flutter analyze
cd mobile && flutter test
cd mobile && flutter test test/path/to/specific_test.dart   # single test
cd mobile && flutter build ios --no-codesign

# Local infra
docker compose up -d   # postgres + redis
```

## Specs

All feature specs live in `docs/specs/T*.md` (T01–T26), ordered by implementation phase. `docs/specs/master_context.md` is the architecture reference — always read it before implementing. `docs/repa_prd.md` is the full product requirements document.

Specs are source of truth. Read before implementing, update after changing behavior.

## Key Conventions

- **API format:** `{ "data": { ... } }` on success, `{ "error": { "code": "...", "message": "..." } }` on error. Base URL `/api/v1`.
- **Auth:** Bearer JWT in `Authorization` header.
- **Validation:** Zod schemas for all backend inputs. Freezed models for all mobile domain entities and API responses.
- **State:** Riverpod only. No setState for business logic.
- **UI language:** Hardcoded Russian strings in MVP (no arb/l10n files).
- **Design:** Color accent `#7C3AED`, system font, white/dark system background.
- **Crystal balance:** Calculated as sum of `CrystalLog.delta` — never stored as a separate field. Spend operations must use Prisma transactions.

## Critical Business Rules

- **Anonymity:** `voterId` is stored in DB (for detector feature) but NEVER returned in API voting results. Detector reveals only a list of voter IDs without question binding.
- **Reveal:** Friday 17:00 UTC (20:00 MSK). Requires quorum: ≥50% voted (≥40% for groups <8).
- **Push limits:** Max 3/day per user, silent 23:00–09:00 MSK.
- **Age gate:** Users under 18 cannot access ROMANCE category.
