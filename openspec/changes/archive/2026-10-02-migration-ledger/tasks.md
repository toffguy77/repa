# Tasks

## 1. Verify before writing anything

- [x] 1.1 Confirm the dev database's schema really is at 010 before forcing a version onto it — check the tables 006/008/009 add and the columns 004/005/007/010 add, rather than assuming; verify the check lists every one of them and reports what is missing if any is
- [x] 1.2 Confirm no `schema_migrations` table exists on the dev database, so forcing creates rather than overwrites a ledger someone else established; verify by querying `pg_tables`

## 2. Reconcile the ledger

- [x] 2.1 Run `migrate force 10` against the local Docker Postgres and confirm `migrate up` then reports `no change` with the schema untouched; verify table and row counts before and after are identical
- [x] 2.2 Run `migrate force 10` against `repa-db-dev` and confirm the same; verify the forced version reads back as 10 with `dirty = false`
- [x] 2.3 Confirm `make migrate` is now safe on the dev database: verify it exits zero and changes nothing

## 3. Make the state easy to read

- [x] 3.1 Add `make migrate-status` reporting the applied version and whether the ledger is dirty, running `migrate` from its pinned container so no local install is needed; verify it prints 10 against both databases
- [x] 3.2 Document the procedure in `docs/features/infrastructure.md`: that `migrate version` writes, that `force` reconciles without running DDL, that the dev cluster needs `target_session_attrs=read-write`, and what a dirty ledger means; verify every command in the doc has been run as written
- [x] 3.3 Note in `CLAUDE.md` that migrations are tracked from now on, and that applying one by hand means reconciling the ledger in the same session; verify the note points at `make migrate-status`

## 4. Verification

- [x] 4.1 Confirm the e2e harness is unaffected — it applies migration files directly and never consults the ledger; verify `go test ./tests/e2e/` still passes
- [x] 4.2 Run `go test ./...` and confirm `openspec validate migration-ledger --strict` passes
