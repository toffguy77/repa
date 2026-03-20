---
description: Use when the user wants a code review of current changes, before committing, or to validate changes against specs. Triggers on /review command.
---

# Code Review

Review current changes against specs and coding standards.

## Flow

1. **Read diff:** `git diff` (unstaged) + `git diff --cached` (staged)

2. **Find relevant specs:**
   - Extract keywords from changed file paths (e.g., `modules/voting/` -> voting, `features/reveal/` -> reveal)
   - Match against `docs/specs/T*.md` filenames and content
   - Always reference `docs/specs/master_context.md` for conventions
   - Read all matched specs

3. **Review checklist:**
   - [ ] Code matches behavior described in specs
   - [ ] If behavior changed -> specs are updated in the same diff
   - [ ] Tests cover new/changed behavior (Vitest for backend, Flutter test for mobile)
   - [ ] Backend: Zod schemas validate all inputs
   - [ ] Backend: Prisma queries are efficient, use transactions where needed
   - [ ] Backend: API responses follow `{ data }` / `{ error: { code, message } }` format
   - [ ] Mobile: UI text is hardcoded Russian (no English left behind)
   - [ ] Mobile: Riverpod for state (no setState in business logic)
   - [ ] Mobile: Feature-first structure respected (`mobile/lib/features/{feature}/`)
   - [ ] Mobile: Freezed models for domain entities and API responses
   - [ ] Anonymity rules: `voterId` never exposed in API results
   - [ ] No security issues (no secrets in code, no SQL injection, no XSS)
   - [ ] No magic strings/numbers — use constants

4. **Output:**
   - If issues found: list each issue with file:line, what's wrong, and how to fix
   - If clean: approve with brief summary of what was reviewed

## Severity levels

- **Blocker:** Spec not updated, tests missing, security issue, anonymity violation
- **Warning:** Style issue, could be improved
- **Note:** Suggestion, not required
