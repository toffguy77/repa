.PHONY: dev build migrate migrate-down sqlc seed test

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
