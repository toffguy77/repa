# Proposal

## Why

Nothing runs the tests. There is no `.github/workflows` directory, so 35 Go packages and 414 Flutter
tests exist and are never executed except by whoever remembers to. The repository has twelve archived
changes' worth of behaviour resting on those tests.

This session is the argument for why it matters. Four defects were found only because something actually
ran: the tone ordering was silently inert on three of four card paths, every history query returned one
season, the block list was unscoped, and a widget read `MediaQuery` from `initState`. Three of the four
were invisible to unit tests and surfaced in e2e or in a mounted widget. A green suite nobody runs is
indistinguishable from no suite.

Two drift risks also have no guard today:

- **`internal/db/sqlc` can go stale.** A generated package that no longer matches the `.sql` files
  compiles fine and uses the wrong query shape. `make sqlc-check` exists now; nothing invokes it.
- **`secrets/` is one `git add -f` from being committed.** It is gitignored and clean today, which is a
  fact worth asserting rather than assuming.

## What Changes

- A `.github/workflows/ci.yml` running on pull requests and pushes to `main`:
  - **backend** — `go vet`, `go test ./...` excluding e2e, with the module cache warmed
  - **backend-e2e** — the testcontainers suite, as its own job on every PR
  - **mobile** — `flutter analyze` and `flutter test`
  - **generated-code** — `make sqlc-check`, failing when the generated package is stale
  - **secrets** — asserts nothing under `secrets/` is tracked
- Versions pinned to what the project actually uses: Go 1.26.1, Flutter 3.47.6, sqlc v1.31.1. A CI that
  silently upgrades a toolchain is a CI that reports somebody else's failure.
- `flutter analyze` rewrites `mobile/analysis_options.yaml` on every run; CI must not fail on that, and
  must not let it mask a real change either.

No product behaviour changes, so this change sets `skip_specs: true`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None — `skip_specs: true`. CI observes behaviour; it does not define any.

## Impact

- **New:** `.github/workflows/ci.yml`.
- **`backend/Makefile`:** a `test-unit` target, so the workflow and a developer exclude e2e the same way
  rather than each spelling it out.
- **Docs:** `CLAUDE.md` (what CI enforces), `docs/features/infrastructure.md` (how to reproduce a CI
  failure locally).
- **Not included:** build/release jobs for iOS and Android, and anything needing credentials. Signing and
  store upload belong with the release work, which needs a keystore and a deploy target that do not exist
  in CI yet.
