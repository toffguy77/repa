.PHONY: dev build migrate migrate-down sqlc seed test test-e2e

dev:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

migrate:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" down 1

sqlc:
	sqlc generate

seed:
	go run ./cmd/seed

test:
	go test ./...

test-e2e: ## Run e2e tests (requires Docker/Podman)
	TESTCONTAINERS_RYUK_DISABLED=true go test -v -count=1 -timeout 300s ./tests/e2e/...
