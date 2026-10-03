# Design

## Context

See proposal.md → Why. The relevant facts, each verified against the local Docker Postgres before writing
this:

- golang-migrate's ledger is `schema_migrations(version bigint primary key, dirty boolean not null)`,
  holding exactly one row.
- `migrate version` is **not** read-only: it creates the table if absent, then reports `error: no
  migration`. So "just checking" already changes the database.
- `migrate force N` sets the version to N and clears `dirty` **without running any migration** — it exists
  precisely for a schema that was brought to N by other means.
- After `force 10`, `migrate up` reports `no change` and touches nothing (confirmed: 24 tables and 205
  questions unchanged).

## Goals / Non-Goals

**Goals**

- `make migrate` is safe to run against `repa-db-dev` and does nothing.
- The applied version is readable without inventing a query.
- The next person who applies a migration by hand knows they have to reconcile.

**Non-Goals**

- Not changing how migrations are written or applied going forward. `make migrate` becomes usable; it does
  not become mandatory.
- Not back-filling a row per version. golang-migrate keeps one row by design; a fabricated history would
  be a different lie from the current one.
- Not touching production: it does not exist yet.

## Decisions

### `migrate force 10`, not a hand-written INSERT

Both produce one row. `force` is better for a reason that matters later: the table's shape belongs to
golang-migrate, not to us. A hand-written `CREATE TABLE` encodes today's column names into a one-off
command that nobody re-reads, and silently diverges if the tool's ledger format ever changes. `force` asks
the tool to write its own bookkeeping.

The version is `10` because `010_question_tone` is the highest applied — golang-migrate takes the integer
prefix of the filename, so the version is `10`, not `010`.

### Run `migrate` from its own container

`migrate` is not installed on this machine and does not need to be. `migrate/migrate:v4.17.1` is the
official image; mounting `internal/db/migrations` read-only means the container can only read what it is
meant to. This also means CI and any operator run the identical version, rather than whatever their
package manager had.

Pinned to a tag rather than `latest`, because a migration tool that silently changes its ledger format
between runs is the one thing worse than no ledger.

### `make migrate-status` rather than documenting a query

The question "what is applied?" currently has no answer short of inspecting
`information_schema.columns`. A target makes it one command, and makes the answer the *tool's* answer
rather than an inference from the schema — which is the whole point of having a ledger.

It reports the version and whether the ledger is dirty. A dirty ledger is the state a failed migration
leaves behind, and it is worth seeing by name rather than as a confusing refusal from the next command.

### The read-write host requirement is documented, not worked around

`PG_HOST` holds two hosts. Without `target_session_attrs=read-write` the driver may connect to the replica,
where every DDL statement fails with `cannot execute CREATE TYPE in a read-only transaction` — which reads
like a permissions problem and is not one. This cost real time to diagnose, so it is written down next to
the command rather than left for the next person to rediscover.

## Risks / Trade-offs

- **`force` can be used to lie.** Forcing a version the schema has not actually reached would make
  `migrate up` skip real migrations, silently. Mitigated by verifying the schema *before* forcing — the
  tasks check the columns and tables 004–010 add, rather than assuming — and by documenting `force` as a
  reconciliation tool rather than a fix for a failed migration.
- **Two databases could drift again.** The ledger makes drift visible rather than preventing it; that is
  the improvement. `make migrate-status` is what makes checking cheap.
- **A dirty ledger still needs a human.** Deliberately: a half-applied migration is a decision, not
  something to automate away.

## Migration Plan

Forward: `migrate force 10` on each database, then `migrate up` to confirm `no change`.

Rollback: `DROP TABLE schema_migrations`. That returns both databases exactly to today's state — which is
the state this change exists to fix, so rolling back is only sensible if the forced version turns out to be
wrong, and then the fix is to force the right one rather than to drop the table.

## Open Questions

None.
