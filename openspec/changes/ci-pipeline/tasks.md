# Tasks

## 1. Make the commands CI will run exist locally first

- [x] 1.1 Add `make test-unit` excluding `./tests/e2e/`, so the workflow and a developer exclude it the same way; verify it runs the 34 non-e2e packages and that `make test` still runs everything
- [x] 1.2 Confirm each command the workflow will invoke succeeds on this machine, exactly as written; verify by running the full list and recording the result

## 2. The workflow

- [x] 2.1 Add `.github/workflows/ci.yml` triggered on pull requests and pushes to `main`, with independent jobs so slow feedback does not block fast; verify the YAML parses and every job's steps are commands proven in task 1
- [x] 2.2 Backend job: Go 1.26.1 pinned to `go.mod`, module cache, `go vet`, `make test-unit`; verify the pin matches `go.mod`
- [x] 2.3 e2e job: the testcontainers suite on every pull request, separate so a container flake is distinguishable from a test failure; verify it runs `make test-e2e`
- [x] 2.4 Mobile job: Flutter 3.47.6, `flutter analyze`, `flutter test`; verify the pin satisfies `pubspec.yaml`'s `sdk` constraint
- [x] 2.5 Restore `mobile/analysis_options.yaml` as the analyzer's known side effect, then fail if the working tree is otherwise dirty — `flutter pub get` moving `pubspec.lock` means the committed lock does not match what the pinned SDK resolves; verify the check passes on a clean tree and fails on the lock drift present today
- [x] 2.6 Generated-code job: `sqlc` v1.31.1 and `make sqlc-check`; verify it fails on a deliberately stale generated file and passes on a clean tree
- [x] 2.7 Secrets job: fail if anything under `secrets/` is tracked; verify it passes now and fails against a tracked file

## 3. Docs

- [x] 3.1 Document in `CLAUDE.md` what CI enforces and the one command to reproduce each failure locally; verify every command in the note has been run
- [x] 3.2 Document in `docs/features/infrastructure.md` the pinned versions and where each pin comes from; verify each matches its source file

## 4. Verification

- [x] 4.1 Validate the workflow file against GitHub's schema, or an equivalent check that catches a malformed job graph rather than only malformed YAML
- [x] 4.2 Run `go test ./...`, `flutter analyze`, `flutter test`, and confirm `openspec validate ci-pipeline --strict` passes
- [ ] 4.3 Confirm the first real run on GitHub is green, or report exactly what failed — a workflow that has never executed is not evidence of anything
