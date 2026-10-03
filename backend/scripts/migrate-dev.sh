#!/usr/bin/env bash
# Run golang-migrate against repa-db-dev.
#
# Exists because three things about this cluster each fail in a way that looks like something else:
#
#   1. PG_HOST holds two hosts, and golang-migrate's driver (lib/pq) takes only one. Pointing it at the
#      wrong one makes every DDL statement fail with "cannot execute CREATE TYPE in a read-only
#      transaction", which reads like a permissions problem. So the primary is detected, not assumed —
#      a managed cluster fails over, and a hardcoded host would be wrong on the day that happens.
#   2. sslmode=verify-full needs the Yandex CA at ~/.postgresql/root.crt
#      (https://storage.yandexcloud.net/cloud-certs/CA.pem). Do not lower sslmode to get past it.
#   3. `migrate` is not installed locally and does not need to be: it runs from its own pinned image,
#      so CI and every operator use the same version.
#
# The password is never printed. Usage:
#
#   backend/scripts/migrate-dev.sh version
#   backend/scripts/migrate-dev.sh up
#   backend/scripts/migrate-dev.sh force 10
#
# Note `version` is NOT read-only: golang-migrate creates schema_migrations before reporting on it.
set -euo pipefail

MIGRATE_IMAGE="migrate/migrate:v4.17.1"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CREDS="$REPO_ROOT/secrets/pg_credentials.env"
CA="$HOME/.postgresql/root.crt"
MIGRATIONS="$REPO_ROOT/backend/internal/db/migrations"

[ -f "$CREDS" ] || { echo "missing $CREDS" >&2; exit 1; }
[ -f "$CA" ] || {
  echo "missing $CA - fetch it with:" >&2
  echo "  curl -sS https://storage.yandexcloud.net/cloud-certs/CA.pem -o $CA && chmod 0600 $CA" >&2
  exit 1
}
[ $# -gt 0 ] || { echo "usage: $(basename "$0") <migrate args...>" >&2; exit 1; }

# Read the credentials without sourcing them: the password contains shell metacharacters.
eval "$(python3 - "$CREDS" <<'PY'
import shlex, sys
for line in open(sys.argv[1]):
    line = line.strip()
    if line and not line.startswith('#') and '=' in line:
        k, v = line.split('=', 1)
        print(f"{k.strip()}={shlex.quote(v.strip().strip(chr(34)).strip(chr(39)))}")
PY
)"

# Which of the configured hosts accepts writes right now.
PRIMARY=""
IFS=',' read -ra HOSTS <<< "$PG_HOST"
for h in "${HOSTS[@]}"; do
  h="$(echo "$h" | tr -d '[:space:]')"
  recovery="$(docker run --rm \
      -v "$CA:/ca.crt:ro" -e PGPASSWORD="$PG_PASSWORD" -e PGSSLROOTCERT=/ca.crt \
      postgres:16-alpine psql \
      "host=$h port=${PG_PORT:-6432} user=$PG_USER dbname=$PG_DBNAME sslmode=verify-full" \
      -tAc 'SELECT pg_is_in_recovery();' 2>/dev/null | tr -d '[:space:]')" || true
  if [ "$recovery" = "f" ]; then PRIMARY="$h"; break; fi
done
[ -n "$PRIMARY" ] || { echo "no read-write host among the configured hosts" >&2; exit 1; }

# Percent-encode the userinfo: golang-migrate takes a URL, and the password is not URL-safe.
enc() { python3 -c 'import sys,urllib.parse as u; print(u.quote(sys.argv[1], safe=""))' "$1"; }

docker run --rm \
  -v "$CA:/ca.crt:ro" \
  -v "$MIGRATIONS:/m:ro" \
  "$MIGRATE_IMAGE" \
  -path=/m \
  -database "postgres://$(enc "$PG_USER"):$(enc "$PG_PASSWORD")@$PRIMARY:${PG_PORT:-6432}/$(enc "$PG_DBNAME")?sslmode=verify-full&sslrootcert=/ca.crt" \
  "$@"
