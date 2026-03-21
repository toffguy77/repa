---
description: Use when the user wants to implement a new feature, change behavior, or fix a bug in the Repa app. Triggers on /feature command or when user describes a feature request.
---

# Feature Development

Full feature development cycle with two checkpoints.

## Flow

### Phase 1: Understanding (before code)

1. **Find relevant specs** — look in `docs/specs/T*.md` for tasks related to the feature. Always read `docs/specs/master_context.md` for architecture and conventions.
2. **Read existing feature docs** — look in `docs/features/` for documentation of affected features. Understand what is already documented about current behavior.
3. **Read matched specs** — understand current app behavior in affected areas.
4. **Ask clarifying questions** — one question at a time until the task is clear.
5. **Present plan** — show: what code changes, what tests, what specs will be updated, what feature docs will be created/updated.

**CHECKPOINT 1:** Wait for user approval before writing any code.

### Phase 2: Implementation (autonomous)

6. **Write code** — implementation following monorepo structure: `backend/` for Go server, `mobile/` for Flutter.
7. **Write tests** — tests are MANDATORY for every feature. See Test Requirements below.
8. **Update specs** — modify affected files in `docs/specs/` to reflect new behavior. If behavior described in ANY existing spec changes, that spec MUST be updated.
9. **Write/update feature documentation** — see Feature Documentation below. This is MANDATORY.
10. **Verify locally:**
    - Backend: `cd backend && make test` + coverage check (see Coverage target below)
    - Mobile: `cd mobile && flutter analyze && flutter test --coverage` + coverage check
    - Mobile build: `cd mobile && flutter build ios --no-codesign`
11. **If verification fails** — investigate root cause. Don't blindly fix tests — a failing test may indicate a code bug. Don't present results until everything is green.

**CHECKPOINT 2:** Show test results, build output, and `git diff`. Wait for explicit "ok" from user.

### Phase 3: Ship (after approval)

12. **Commit** — code + updated specs + tests + feature docs in one commit.
13. **Push** — to feature branch.
14. **Create PR** — against main with summary of changes.

## Feature Documentation

Documentation is not optional. Every implemented feature must be documented in `docs/features/`.

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

Files are organized by **domain feature**, not by task number. One feature doc may span multiple tasks (e.g., `auth.md` covers T04 backend + T05 mobile).

### What each doc must contain

```markdown
# Feature Name

## Overview
One paragraph: what this feature does, who uses it, why it exists.

## API Endpoints
For each endpoint:
- Method + path (e.g., `POST /api/v1/auth/otp/send`)
- Request body / query params (with types)
- Response format (success + error cases)
- Auth: required or public

## Data Model
- Relevant DB tables and key fields
- Relationships between tables
- Enums used

## Business Rules
- Constraints, limits, edge cases
- What is allowed/forbidden
- Key invariants (e.g., "voter_id is never exposed in vote results")

## Mobile Screens (if applicable)
- Screen name → route path
- Key UI elements and user interactions
- State management: which providers/notifiers are used
- Navigation flow between screens

## Architecture
- File locations: handlers, services, repositories, screens
- Key dependencies between components
- External services used (Redis, S3, FCM, etc.)
```

### Rules

- **Create** a new doc when implementing a feature that doesn't have one yet.
- **Update** an existing doc when modifying behavior of a documented feature.
- **Every endpoint, screen, notifier, and business rule** must appear in a feature doc. If you wrote code that isn't documented — add it.
- Feature docs describe **what IS implemented** (current state), not what is planned. Don't document unbuilt features.
- If a feature spans backend + mobile, both sides go in the SAME doc file.
- Keep docs concise and factual. No tutorials, no marketing — just reference material for the next developer (or AI agent) working on this code.
- At CHECKPOINT 2, if any new code is not covered by feature docs, that's a **blocker**.

## Test Requirements

Tests are not optional. Every feature must include tests proportional to its complexity.

### Backend (Go)

Test files live next to the code they test: `internal/handler/{feature}/handler_test.go`, `internal/service/{feature}/service_test.go`.

**What to test:**
- **Handler tests** (every new endpoint): use `httptest` + Echo test context. Test happy path, validation errors (400), auth errors (401/403), not found (404), conflict (409). Verify response JSON structure matches `{ "data": ... }` / `{ "error": ... }` convention.
- **Service tests** (business logic): test core logic with mocked dependencies (DB queries, Redis). Test edge cases: empty inputs, boundary values, concurrent state.
- **Integration tests** (when touching DB queries): test actual SQL via sqlc against a test database when the logic is non-trivial (aggregations, upserts, race conditions).

**Minimum per endpoint:** at least 3 test cases — happy path, one validation error, one business rule violation.

**Tools:** `testify/assert`, `testify/require`, `httptest`, `echo.New()` for handler tests. Mock interfaces for service dependencies.

### Mobile (Flutter)

Test files mirror source: `test/features/{feature}/...`, `test/core/...`.

**What to test:**
- **Unit tests** (notifiers, repositories): test state transitions, error handling, edge cases. Mock API calls with fake implementations or `mocktail`.
- **Widget tests** (screens): test that screens render correctly, user interactions trigger expected state changes, loading/error states display properly. Use `ProviderScope` overrides for dependency injection.
- **No integration/E2E tests** in MVP — focus on unit and widget tests.

**Minimum per screen:** at least 2 widget tests — renders without errors, primary user action works.
**Minimum per notifier:** at least 3 unit tests — happy path, error case, edge case.

**Tools:** `flutter_test`, `mocktail` for mocking, `ProviderScope(overrides: [...])` for Riverpod.

### Coverage target: 80%

New and changed code must reach **>=80% line coverage**. This is a blocker at CHECKPOINT 2.

**Backend:** Run `cd backend && go test -coverprofile=coverage.out ./internal/handler/{feature}/... ./internal/service/{feature}/...` then check with `go tool cover -func=coverage.out`. Report coverage % for each changed package.

**Mobile:** Run `cd mobile && flutter test --coverage` then check with `lcov --summary coverage/lcov.info` (or parse the file). Report coverage % for changed files. Exclude generated files (`*.g.dart`, `*.freezed.dart`).

If coverage is below 80% — add more tests before proceeding. Do NOT reduce the threshold.

### General rules

- Read existing tests in affected modules BEFORE writing new ones — match style and patterns.
- If existing tests break — investigate: is it a bug in new code or an outdated test? If bug — fix code, not test. If outdated — update test with understanding of new behavior.
- Test real scenarios, not implementation details. A test that breaks on every refactor is worse than no test.
- Never skip tests to save time. If CHECKPOINT 2 shows 0 new test files or coverage < 80%, that's a blocker.

## Rules

- Specs in `docs/specs/` are source of truth. Always read before, always update after.
- Feature docs in `docs/features/` document current implementation. Always create or update after implementing.
- UI text is hardcoded in Russian (no arb files in MVP).
- Use Riverpod for state. No setState in business logic.
- Backend: Go + Echo handlers in `backend/internal/handler/{feature}/`, services in `backend/internal/service/{feature}/`, sqlc for DB queries.
- Mobile: feature-first architecture `mobile/lib/features/{feature}/` with data/domain/presentation layers.
- Mobile: Freezed for all domain entities and API responses.
- Color accent `#7C3AED`, system font, white/dark system background.
