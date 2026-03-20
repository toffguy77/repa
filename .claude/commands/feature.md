---
description: Use when the user wants to implement a new feature, change behavior, or fix a bug in the Repa app. Triggers on /feature command or when user describes a feature request.
---

# Feature Development

Full feature development cycle with two checkpoints.

## Flow

### Phase 1: Understanding (before code)

1. **Find relevant specs** — look in `docs/specs/T*.md` for tasks related to the feature. Always read `docs/specs/master_context.md` for architecture and conventions.
2. **Read matched specs** — understand current app behavior in affected areas.
3. **Ask clarifying questions** — one question at a time until the task is clear.
4. **Present plan** — show: what code changes, what tests, what specs will be updated.

**CHECKPOINT 1:** Wait for user approval before writing any code.

### Phase 2: Implementation (autonomous)

5. **Write code** — implementation following monorepo structure: `backend/` for server, `mobile/` for Flutter.
6. **Tests** — think before generating tests:
   - Read existing tests in affected modules (`backend/src/modules/*/*.test.ts`, `mobile/test/`).
   - If existing tests break — investigate: is it a bug in new code or an outdated test? If bug — fix code, not test. If outdated — update test with understanding of new behavior.
   - Write new tests for added behavior: test real scenarios (happy path + edge cases), not just line coverage.
   - Backend tests must be stable (Vitest). Mobile tests must account for scroll, animations, async providers.
7. **Update specs** — modify affected files in `docs/specs/` to reflect new behavior. If behavior described in ANY existing spec changes, that spec MUST be updated.
8. **Verify locally:**
   - Backend: `cd backend && npm run lint && npm test`
   - Mobile: `cd mobile && flutter analyze && flutter test`
   - Mobile build: `cd mobile && flutter build ios --no-codesign`
9. **If verification fails** — investigate root cause. Don't blindly fix tests — a failing test may indicate a code bug. Don't present results until everything is green.

**CHECKPOINT 2:** Show test results, build output, and `git diff`. Wait for explicit "ok" from user.

### Phase 3: Ship (after approval)

10. **Commit** — code + updated specs + tests in one commit.
11. **Push** — to feature branch.
12. **Create PR** — against main with summary of changes.

## Rules

- Specs in `docs/specs/` are source of truth. Always read before, always update after.
- UI text is hardcoded in Russian (no arb files in MVP).
- Use Riverpod for state. No setState in business logic.
- Backend: Fastify modules in `backend/src/modules/{module}/`, Zod for validation, Prisma for DB.
- Mobile: feature-first architecture `mobile/lib/features/{feature}/` with data/domain/presentation layers.
- Mobile: Freezed for all domain entities and API responses.
- Color accent `#7C3AED`, system font, white/dark system background.
