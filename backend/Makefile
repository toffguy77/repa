.PHONY: dev build migrate migrate-down migrate-status migrate-dev migrate-dev-status sqlc sqlc-check seed test test-unit test-e2e

dev:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

# These two need a local `migrate` binary and a DATABASE_URL with a single host. For repa-db-dev use the
# migrate-dev targets below instead: that cluster has two hosts and golang-migrate's driver takes one.
migrate:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" down 1

# What the ledger says is applied, and whether a failed migration left it dirty. A dirty ledger makes
# every later `migrate` refuse to run until a human decides what happened — seeing it named is better
# than meeting it as a confusing refusal.
migrate-status:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" version

# repa-db-dev: resolves the read-write host, mounts the Yandex CA, and runs migrate from its pinned
# image, so nothing needs installing. See docs/features/infrastructure.md.
migrate-dev:
	./scripts/migrate-dev.sh up

migrate-dev-status:
	./scripts/migrate-dev.sh version

sqlc:
	sqlc generate

# Fails if internal/db/sqlc does not match what the .sql files produce. A stale generated package
# compiles fine and silently uses the wrong query shape, so nothing else catches this.
#
# Compares the directory against itself across a regeneration rather than against git HEAD: the git
# form reports a false failure for anyone with uncommitted work, which makes the target useless
# locally and therefore never run.
sqlc-check:
	@before=$$(find internal/db/sqlc -name '*.go' -exec sha256sum {} + | sort | sha256sum); \
	$(MAKE) --no-print-directory sqlc; \
	after=$$(find internal/db/sqlc -name '*.go' -exec sha256sum {} + | sort | sha256sum); \
	if [ "$$before" != "$$after" ]; then \
		echo "internal/db/sqlc is out of date — 'make sqlc' changed it; commit the result" >&2; \
		exit 1; \
	fi; \
	echo "sqlc generated code is up to date"

seed:
	go run ./cmd/seed

test:
	go test ./...

# Everything except the testcontainers suite. One definition, used by CI and by anyone who wants fast
# feedback — spelling the exclusion out in both places means they disagree the first time a second
# integration package appears.
test-unit:
	go test $$(go list ./... | grep -v '/tests/e2e')

test-e2e: ## Run e2e tests (requires Docker/Podman)
	TESTCONTAINERS_RYUK_DISABLED=true go test -v -count=1 -timeout 300s ./tests/e2e/...
