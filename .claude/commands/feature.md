---
description: Use when the user wants to implement a new feature, change behavior, or fix a bug in the Repa app. Triggers on /feature command or when user describes a feature request.
---

# Feature Development

Systematic feature development: deep understanding, architecture design, implementation with tests, code review, ship.

## Flow

### Phase 1: Discovery

**Goal:** Understand what needs to be built.

1. Create TodoWrite checklist tracking all phases.
2. If the request is vague, ask: what problem, what behavior, any constraints?
3. Summarize understanding and confirm with user.

---

### Phase 2: Codebase Exploration

**Goal:** Deeply understand relevant existing code and specs before making decisions.

1. **Read specs** — always read `docs/specs/master_context.md`. Grep `docs/specs/T*.md` for tasks related to the feature.
2. **Read feature docs** — check `docs/features/` for documentation of affected features.
3. **Launch 2-3 explore subagents in parallel**, each targeting a different aspect:
   - Similar features in the codebase (trace implementation end-to-end)
   - Architecture and abstractions in the affected area
   - Existing tests and patterns in affected modules
   - Each agent must return a list of 5-10 key files to read.
4. **Read all key files** returned by agents — build deep context before proceeding.
5. Present summary of findings: patterns discovered, relevant code, integration points.

---

### Phase 3: Clarifying Questions

**Goal:** Resolve all ambiguities before designing. DO NOT SKIP.

1. Review codebase findings + original request.
2. Identify underspecified aspects: edge cases, error handling, API contract details, scope boundaries, backend vs mobile split, anonymity implications.
3. **Present all questions in one organized list** (not one at a time).
4. **Wait for answers before proceeding.**

If user says "on your discretion" — state your recommendation and get explicit confirmation.

---

### Phase 4: Architecture & Plan

**Goal:** Design the approach. For trivial changes, keep this brief. For non-trivial features — compare options.

**For small changes (bug fix, single endpoint, minor UI tweak):**
- Present a concise plan: what changes, what tests, what docs update.

**For non-trivial features (new domain, multi-layer, backend+mobile):**
1. Launch 2-3 architect subagents with different approaches:
   - Minimal: smallest change, maximum reuse of existing code
   - Clean: best abstractions, long-term maintainability
   - Pragmatic: balance of speed and quality
2. Present trade-offs comparison and **your recommendation with reasoning**.
3. Ask user which approach they prefer.

Plan must include: code changes, new files, tests, spec updates, feature doc updates.

**CHECKPOINT 1:** Wait for user approval before writing any code.

---

### Phase 5: Implementation (autonomous)

**Goal:** Build the feature with tests and documentation.

1. **Write code** — follow monorepo structure: `backend/` for Go, `mobile/` for Flutter.
2. **Write tests** — MANDATORY. See Test Requirements below.
3. **Update specs** — modify affected `docs/specs/` files. If behavior in ANY existing spec changes, that spec MUST be updated.
4. **Write/update feature docs** — MANDATORY, NON-NEGOTIABLE. See Feature Documentation below.
   - After writing code, IMMEDIATELY create or update the relevant `docs/features/*.md` file.
   - Run `ls docs/features/` to verify the doc exists before moving to verification.
   - If the doc is missing or outdated, this is a **blocker** — do NOT proceed to verification.
5. **Verify locally:**
   - Backend: `cd backend && make test` + coverage check (>=80%)
   - Mobile: `cd mobile && flutter analyze && flutter test --coverage` + coverage check
   - Mobile build: `cd mobile && flutter build ios --no-codesign`
6. **If verification fails** — investigate root cause. Don't blindly fix tests — a failing test may indicate a code bug. Don't proceed until green.

---

### Phase 6: Code Review

**Goal:** Catch issues before presenting to user.

1. Launch 3 review subagents in parallel:
   - **Simplicity & DRY** — duplicated code, unnecessary abstractions, readability
   - **Bugs & correctness** — logic errors, race conditions, missing edge cases, anonymity violations
   - **Conventions** — Repa patterns (sqlc not GORM, Riverpod not setState, Freezed for entities, API response format)
2. Consolidate findings, prioritize by severity.
3. Fix critical issues (bugs, anonymity leaks, convention violations) autonomously.
4. Note minor suggestions for CHECKPOINT 2 discussion.

---

### Phase 7: Present Results

**CHECKPOINT 2:** Show:
- Test results + coverage %
- Build output
- `git diff`
- Any minor review findings not auto-fixed
- **Feature doc diff** — show exactly what was created or changed in `docs/features/`

**DOC GATE:** If no `docs/features/*.md` file was created or updated in this session, STOP. Go back and write the doc before presenting results. Undocumented code is a hard blocker.

Wait for explicit "ok" from user.

---

### Phase 8: Ship (after approval)

1. **Commit** — code + specs + tests + feature docs in one commit.
2. **Push** — to feature branch.
3. **Create PR** — against main with summary of changes.

---

## Feature Documentation

Every implemented feature must be documented in `docs/features/`.

### Structure

```
docs/features/
├── auth.md              # Authentication (OTP, Apple, Google, tokens)
├── groups.md            # Groups (create, join, invite, manage)
├── voting.md            # Voting flow (seasons, questions, votes)
├── reveal.md            # Reveal engine (results, achievements, cards)
├── crystals.md          # Crystal economy (purchase, spend, balance)
├── push.md              # Push notifications (FCM, preferences, scheduling)
├── telegram.md          # Telegram bot integration
└── moderation.md        # AI moderation of user questions
```

Files are organized by **domain feature**, not by task number.

### What each doc must contain

```markdown
# Feature Name

## Overview
One paragraph: what this feature does, who uses it, why it exists.

## API Endpoints
For each endpoint: method + path, request/response format, auth requirement.

## Data Model
Relevant DB tables, relationships, enums.

## Business Rules
Constraints, limits, edge cases, invariants.

## Mobile Screens (if applicable)
Screen name, route, key interactions, providers used.

## Architecture
File locations, dependencies, external services.
```

### Rules

- **Create** a new doc when implementing a feature that doesn't have one yet.
- **Update** an existing doc when modifying behavior of a documented feature.
- Feature docs describe **what IS implemented**, not what is planned.
- Backend + mobile go in the SAME doc file.
- At CHECKPOINT 2, undocumented new code is a **blocker**.

---

## Test Requirements

Tests are not optional. Every feature must include tests proportional to its complexity.

### Backend (Go)

Test files live next to the code: `internal/handler/{feature}/handler_test.go`, `internal/service/{feature}/service_test.go`.

**What to test:**
- **Handler tests** (every new endpoint): `httptest` + Echo test context. Happy path, validation errors (400), auth errors (401/403), not found (404), conflict (409). Verify `{ "data": ... }` / `{ "error": ... }` response format.
- **Service tests** (business logic): core logic with mocked dependencies. Edge cases: empty inputs, boundary values, concurrent state.
- **Integration tests** (non-trivial DB queries): test actual SQL via sqlc against test database.

**Minimum per endpoint:** 3 test cases — happy path, validation error, business rule violation.

**Tools:** `testify/assert`, `testify/require`, `httptest`, `echo.New()`. Mock interfaces for service dependencies.

### Mobile (Flutter)

Test files mirror source: `test/features/{feature}/...`.

**What to test:**
- **Unit tests** (notifiers, repositories): state transitions, error handling, edge cases. Mock with `mocktail`.
- **Widget tests** (screens): renders correctly, user interactions work, loading/error states. Use `ProviderScope` overrides.

**Minimum per screen:** 2 widget tests — renders without errors, primary action works.
**Minimum per notifier:** 3 unit tests — happy path, error case, edge case.

**Tools:** `flutter_test`, `mocktail`, `ProviderScope(overrides: [...])`.

### Coverage target: 80%

New and changed code must reach **>=80% line coverage**. Blocker at CHECKPOINT 2.

- **Backend:** `go test -coverprofile=coverage.out ./internal/handler/{feature}/... ./internal/service/{feature}/...` then `go tool cover -func=coverage.out`.
- **Mobile:** `flutter test --coverage` then `lcov --summary coverage/lcov.info`. Exclude `*.g.dart`, `*.freezed.dart`.

### General rules

- Read existing tests before writing new ones — match style.
- Broken existing tests: investigate — bug in new code or outdated test?
- Test real scenarios, not implementation details.
- 0 new tests or coverage < 80% is a blocker.

---

## Rules

- Specs in `docs/specs/` are source of truth. Always read before, always update after.
- Feature docs in `docs/features/` document current implementation. Always create or update.
- UI text is hardcoded in Russian (no arb files in MVP).
- Use Riverpod for state. No setState in business logic.
- Backend: Go + Echo handlers in `backend/internal/handler/{feature}/`, services in `backend/internal/service/{feature}/`, sqlc for DB queries.
- Mobile: feature-first architecture `mobile/lib/features/{feature}/` with data/domain/presentation layers.
- Mobile: Freezed for all domain entities and API responses.
- Color accent `#7C3AED`, system font, white/dark system background.
- **Anonymity is critical:** never expose `voter_id` in connection to specific votes in API responses.
