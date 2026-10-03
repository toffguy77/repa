# Design

## Context

See proposal.md → Why. What the suites actually need:

- Backend: Go 1.26.1. The e2e suite uses testcontainers-go, so it needs a Docker daemon — present on
  `ubuntu-latest`. It takes ~4s of test time but ~30s of container startup.
- Mobile: Flutter 3.47.6 / Dart SDK `^3.11.1`. 414 tests, ~7s.
- `make sqlc-check` needs the `sqlc` binary (v1.31.1).

## Goals / Non-Goals

**Goals**

- Every pull request runs everything that can run without credentials.
- A CI failure is reproducible locally with one documented command.
- Toolchain versions are the project's, pinned.

**Non-Goals**

- No build or signing jobs. They need a keystore and a deploy target; that is the release change.
- No coverage gate. Coverage went 15% → 82% without one, and a threshold invites tests written to move a
  number.
- Not making CI the only way to run tests. `make test` stays the local path.

## Decisions

### e2e is its own job, on every pull request

Asked and answered by the project owner: every PR. The reason to pay for it is that this session's worst
defects were ones mocks could not see — a sqlmock fixture happily returns whatever shape the test author
wrote, so a query that no longer matches the schema passes. Only real Postgres catches that.

A separate job rather than part of the backend job, so a container-startup flake is distinguishable from a
test failure, and so the fast feedback does not wait on the slow.

### Versions pinned, with the pin next to the reason

`go-version: 1.26.1` matches `go.mod`; `flutter-version: 3.47.6` matches the installed SDK and satisfies
`sdk: ^3.11.1`; `sqlc` is installed at `v1.31.1`.

Floating versions are tempting and wrong here: a toolchain that updates itself turns an unrelated upstream
change into a red build on someone's unrelated PR, and the first instinct is to distrust CI rather than the
pin. Dependabot can raise the pins as PRs, which is the right place to see a toolchain change — as a
change.

### One known tool side effect is restored; every other one fails the build

`flutter analyze` rewrites `mobile/analysis_options.yaml` on every run ("Upgrading analysis_options.yaml
to exclude build and platform directories"). That single file is restored from the index and not reported.

The first version of this step also hashed the file before and after to "catch a real edit arriving in the
same PR". That check cannot fire and was removed: `git checkout -- <file>` restores from the index, which
in CI already holds the PR's committed content, so before and after are equal by construction. A check
that can never fail is noise that teaches people to ignore the step.

What replaces it is a check that *can* fail: after analyze and test, assert the working tree is otherwise
clean. The case this catches is real and present in this repository today — `flutter pub get` moves four
transitive versions in `pubspec.lock` under the pinned SDK, which means the committed lock does not
describe what actually gets resolved. A lockfile exists precisely to make that impossible, so CI failing
on it is the correct outcome: either the lock is committed or the pin is wrong.

The step runs with `if: always()` so a test failure does not also hide a dirty tree.

### `make test-unit` rather than a path list in YAML

The backend job must exclude `./tests/e2e/`. Spelling that out in YAML means the workflow and the developer
disagree the first time a second integration package appears. A Makefile target keeps one definition.

### The secrets check asserts, rather than scans

`git ls-files secrets/` returning anything is a failure. This is not secret *scanning* — it does not look
for keys in source, which is a different job with a different tool. It asserts the one property the
project already relies on: that the directory `CLAUDE.md` designates for credentials is not tracked.
Cheap, exact, and it fails loudly the first time someone runs `git add -f`.

## Risks / Trade-offs

- **testcontainers flakes in CI** → the e2e job is separate, so a flake is legible as such. One
  unreproducible failure was seen locally in ~17 runs this session; if that recurs in CI it will at least
  be recorded with output, which is more than we have now.
- **CI becomes slow enough to route around** → the fast jobs are independent of e2e, so a PR gets unit and
  analyze feedback without waiting. If total time becomes a problem the answer is caching, not dropping
  e2e.
- **Pinned versions go stale** → visible as an outdated pin rather than as a mysterious failure, and
  Dependabot is already configured for modules in this repo.

## Migration Plan

Additive: a new workflow file. Nothing depends on it, and deleting it returns the repo to today.

The first run will be on whichever branch carries this change, which is also the first real check that the
suites pass somewhere other than this machine.

## Open Questions

None.
