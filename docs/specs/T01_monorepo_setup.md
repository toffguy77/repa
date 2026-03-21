# T01 — Монорепозиторий, Docker, конфиги

## Цель
Создать рабочую основу проекта: монорепо, Docker Compose, Go-модуль backend, sqlc конфиг.

## Структура монорепо
```
repa/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── db/
│   │   │   ├── migrations/       # SQL файлы
│   │   │   ├── queries/          # SQL для sqlc
│   │   │   └── sqlc/             # сгенерированный код (gitignore нет — коммитим)
│   │   ├── handler/
│   │   ├── service/
│   │   ├── worker/
│   │   ├── middleware/
│   │   └── lib/
│   ├── sqlc.yaml
│   ├── .env.example
│   ├── Makefile
│   └── go.mod
├── mobile/
│   └── .gitkeep
├── docker-compose.yml
├── .gitignore
└── README.md
```

## docker-compose.yml
```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: repa
      POSTGRES_USER: repa
      POSTGRES_PASSWORD: repa
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U repa"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

## go.mod
```
module github.com/repa-app/repa

go 1.22
```

## Зависимости (go get)
```bash
go get github.com/labstack/echo/v4
go get github.com/labstack/echo-contrib/echoprometheus  # optional
go get github.com/golang-jwt/jwt/v5
go get github.com/go-playground/validator/v10
go get github.com/redis/go-redis/v9
go get github.com/hibiken/asynq
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/stdlib
go get github.com/rs/zerolog
go get github.com/joho/godotenv
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/service/s3
go get firebase.google.com/go/v4
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/chromedp/chromedp
go get github.com/stretchr/testify
go get github.com/golang-migrate/migrate/v4
```

## sqlc.yaml
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries/"
    schema: "internal/db/migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db/sqlc"
        emit_json_tags: true
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
```

## internal/config/config.go
```go
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    Port           string
    DatabaseURL    string
    RedisURL       string
    JWTSecret      string
    S3Endpoint     string
    S3Bucket       string
    S3AccessKey    string
    S3SecretKey    string
    S3Region       string
    AnthropicKey   string
    YukassaShopID  string
    YukassaSecret  string
    YukassaReturn  string
    TelegramToken  string
    TelegramSecret string
    AppBaseURL     string
    AppMinVersion  string
    AppLatestVersion string
    AdminUsername  string
    AdminPassword  string
    FirebaseProjectID    string
    FirebasePrivateKey   string
    FirebaseClientEmail  string
}

func Load() *Config {
    _ = godotenv.Load()
    return &Config{
        Port:        getEnv("PORT", "3000"),
        DatabaseURL: mustEnv("DATABASE_URL"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        JWTSecret:   mustEnv("JWT_SECRET"),
        // ... остальные поля
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" { return v }
    return fallback
}
func mustEnv(key string) string {
    v := os.Getenv(key)
    if v == "" { panic("required env var missing: " + key) }
    return v
}
```

## cmd/server/main.go
```go
package main

import (
    "context"
    "log"
    "github.com/repa-app/repa/internal/config"
)

func main() {
    cfg := config.Load()
    // Инициализация клиентов (DB, Redis)
    // Создание Echo инстанса
    // Регистрация роутов
    // Start asynq worker (отдельная горутина)
    // ListenAndServe
    log.Printf("Server starting on :%s", cfg.Port)
}
```

## Makefile
```makefile
.PHONY: dev build migrate sqlc seed

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
```

## .env.example (полный)
```env
PORT=3000
DATABASE_URL=postgresql://repa:repa@localhost:5432/repa?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=change_me_in_production

S3_ENDPOINT=https://storage.yandexcloud.net
S3_BUCKET=repa-media
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_REGION=ru-central1

FIREBASE_PROJECT_ID=
FIREBASE_PRIVATE_KEY=
FIREBASE_CLIENT_EMAIL=

ANTHROPIC_API_KEY=
YUKASSA_SHOP_ID=
YUKASSA_SECRET_KEY=
YUKASSA_RETURN_URL=https://repa.app/payment/return

TELEGRAM_BOT_TOKEN=
TELEGRAM_WEBHOOK_SECRET=

APP_BASE_URL=https://repa.app
APP_MIN_VERSION=1.0.0
APP_LATEST_VERSION=1.0.0
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change_me
```

## .gitignore
```
bin/
*.env
.env
!.env.example
# Flutter
mobile/build/
mobile/.dart_tool/
mobile/android/local.properties
mobile/ios/Flutter/Generated.xcconfig
```

## README.md
```markdown
## Запуск

docker compose up -d
cd backend
cp .env.example .env
make migrate
make seed
make dev
```

## Критерии готовности
- [ ] `docker compose up -d` поднимает postgres и redis
- [ ] `make dev` запускает сервер без ошибок
- [ ] `sqlc generate` работает без ошибок
- [ ] `make migrate` применяет миграции (пока пустые — будут в T02)
- [ ] `GET /health` → 200
