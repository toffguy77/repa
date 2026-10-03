# Proposal

## Why

`make migrate` runs golang-migrate, which tracks applied versions in a `schema_migrations` table. On
`repa-db-dev` that table **did not exist**: migrations 001–003 were applied by hand before this work, and
004–010 were applied by hand during it. The database's schema is at version 10 and its ledger says
nothing at all.

golang-migrate reads an absent ledger as "version 0, nothing applied" and starts from `001_init.up.sql`.
So the first person to run `make migrate` against that database gets `CREATE TYPE ... already exists`,
and — because `migrate` marks the version **dirty** when a step fails — leaves the ledger in a state
where every subsequent `migrate` refuses to run until someone forces it by hand. The same trap is waiting
on production the first time anyone deploys: a `migrate up` in a release pipeline would fail on a
database that is actually perfectly up to date.

This is not a theoretical risk. It is one command away, the command is the documented one
(`CLAUDE.md` → *Apply migrations to the dev DB*), and it looks like the right thing to run.

## What Changes

- The ledger on `repa-db-dev` is reconciled to version 10 using `migrate force`, which is golang-migrate's
  own tool for "the schema is already at this version". No DDL runs.
- The same reconciliation is applied to the local Docker Postgres, so a developer's database and the dev
  database agree about what is applied.
- `docs/features/infrastructure.md` gains the procedure, including two things that cost an hour to
  discover: `migrate version` **writes** (it creates the table before reporting "no migration"), and the
  dev cluster needs `target_session_attrs=read-write` or every DDL fails as read-only on the replica.
- A `make migrate-status` target, so "what is actually applied" is one command rather than a query
  someone has to invent.
- A note in `CLAUDE.md` that migrations are tracked from now on, and that applying one by hand means
  reconciling the ledger in the same session.

No code, no schema, and no product behaviour change — so this change declares `skip_specs: true`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. `openspec/changes/migration-ledger/.openspec.yaml` sets `skip_specs: true`: nothing observable to a
user of the app changes, so inventing a requirement to satisfy validation would be worse than declaring
the absence honestly.

## Impact

- **`repa-db-dev`:** one new table, `schema_migrations`, holding one row. No existing table or row touched.
- **Local Docker Postgres:** the same.
- **`backend/Makefile`:** a `migrate-status` target.
- **Docs:** `docs/features/infrastructure.md`, `CLAUDE.md`.
- **Not production:** there is no production database yet. The procedure documented here is what the
  deploy change (next) will need, which is why this comes first.
