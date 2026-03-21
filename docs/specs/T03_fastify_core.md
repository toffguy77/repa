# T03 — Echo: приложение, middleware, базовая структура

## Цель
Production-ready Echo приложение: JWT auth middleware, error handling, структура модулей, asynq worker.

## internal/lib/db.go
```go
package lib

import (
    "github.com/jackc/pgx/v5/stdlib"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(dsn string) *pgxpool.Pool {
    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil { panic(err) }
    return pool
}
```

## internal/lib/redis.go
```go
package lib

import "github.com/redis/go-redis/v9"

func NewRedis(url string) *redis.Client {
    opts, _ := redis.ParseURL(url)
    return redis.NewClient(opts)
}
```

## internal/lib/asynq.go
```go
package lib

import "github.com/hibiken/asynq"

const (
    TypeRevealChecker    = "reveal:checker"
    TypeRevealProcess    = "reveal:process"
    TypeSeasonCreator    = "season:creator"
    TypeAchievements     = "achievements:calculate"
    TypePushWeekly       = "push:weekly-scheduler"
    TypePushTuesday      = "push:tuesday-signal"
    TypePushWednesday    = "push:wednesday-quorum"
    TypePushThursday     = "push:thursday-teaser"
    TypePushFriPreReveal = "push:friday-pre-reveal"
    TypePushReveal       = "push:reveal-notification"
    TypePushSundayPrev   = "push:sunday-preview"
    TypePushSundayStreak = "push:sunday-streak"
    TypeTelegramStart    = "telegram:season-start"
    TypeTelegramReveal   = "telegram:reveal-post"
    TypeTelegramShare    = "telegram:share-card"
    TypeReactionPush     = "push:reaction"
)

func NewAsynqClient(redisURL string) *asynq.Client {
    opts, _ := asynq.ParseRedisURI(redisURL)
    return asynq.NewClient(opts)
}

func NewAsynqServer(redisURL string) *asynq.Server {
    opts, _ := asynq.ParseRedisURI(redisURL)
    return asynq.NewServer(opts, asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
    })
}
```

## internal/middleware/auth.go
```go
package middleware

import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
)

type JWTClaims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

func JWTAuth(secret string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Извлечь токен из Authorization: Bearer {token}
            // Валидировать
            // Положить claims в context: c.Set("user", claims)
            // При ошибке → 401 { error: { code: "UNAUTHORIZED", message: "..." } }
        }
    }
}

// GetCurrentUser извлекает claims из context
func GetCurrentUser(c echo.Context) *JWTClaims {
    return c.Get("user").(*JWTClaims)
}
```

## internal/middleware/ratelimit.go
```go
// Rate limiting через Redis
// Ключ: `rl:{endpoint}:{identifier}`, TTL: window
// При превышении → 429 { error: { code: "RATE_LIMIT", message: "Слишком много запросов" } }

func RateLimit(rdb *redis.Client, key string, limit int, window time.Duration) echo.MiddlewareFunc
```

## Формат ошибок (app-wide)

```go
// internal/handler/errors.go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func ErrorResponse(c echo.Context, status int, code, message string) error {
    return c.JSON(status, map[string]interface{}{
        "error": AppError{Code: code, Message: message},
    })
}

// Коды ошибок:
// UNAUTHORIZED, FORBIDDEN, NOT_FOUND, CONFLICT, VALIDATION, RATE_LIMIT
// INSUFFICIENT_CRYSTALS, SEASON_NOT_REVEALED, QUORUM_NOT_REACHED

// Global error handler в main:
e.HTTPErrorHandler = func(err error, c echo.Context) {
    // echo.HTTPError → стандартный ответ
    // AppError → { error: { code, message } }
    // validation.ValidationErrors → 400 с деталями
    // default → 500
}
```

## Структура типового handler'а
```go
// internal/handler/groups/handler.go
package groups

type Handler struct {
    svc *service.GroupService
}

func NewHandler(svc *service.GroupService) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Register(g *echo.Group) {
    g.POST("", h.CreateGroup)
    g.GET("", h.ListGroups)
    g.GET("/:id", h.GetGroup)
    // ...
}

func (h *Handler) CreateGroup(c echo.Context) error {
    user := middleware.GetCurrentUser(c)
    var req CreateGroupRequest
    if err := c.Bind(&req); err != nil { return err }
    if err := c.Validate(&req); err != nil { return err }
    // вызов service
    // return c.JSON(201, map[string]interface{}{"data": result})
}
```

## Валидатор
```go
// internal/middleware/validator.go
import "github.com/go-playground/validator/v10"

type CustomValidator struct {
    validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        return echo.NewHTTPError(400, map[string]interface{}{
            "error": map[string]string{
                "code": "VALIDATION",
                "message": err.Error(),
            },
        })
    }
    return nil
}

// Зарегистрировать в main: e.Validator = &CustomValidator{validator: validator.New()}
```

## cmd/server/main.go (полный)
```go
func main() {
    cfg := config.Load()

    // Инфраструктура
    pool := lib.NewDB(cfg.DatabaseURL)
    rdb  := lib.NewRedis(cfg.RedisURL)
    asynqClient := lib.NewAsynqClient(cfg.RedisURL)

    // Echo
    e := echo.New()
    e.HideBanner = true
    e.Validator = middleware.NewValidator()
    e.HTTPErrorHandler = handler.ErrorHandler

    // Middleware
    e.Use(echomiddleware.Logger())
    e.Use(echomiddleware.Recover())
    e.Use(echomiddleware.CORS())
    e.Use(echomiddleware.SecureWithConfig(...))

    // Queries (sqlc)
    q := db.New(pool)

    // Services
    authSvc   := authservice.New(q, rdb, cfg)
    groupsSvc := groupsservice.New(q, rdb, asynqClient)
    // ...

    // Routes
    api := e.Group("/api/v1")
    api.GET("/health", healthHandler(pool, rdb))

    authG := api.Group("/auth")
    authhandler.NewHandler(authSvc).Register(authG)

    protected := api.Group("", middleware.JWTAuth(cfg.JWTSecret))
    groupshandler.NewHandler(groupsSvc).Register(protected.Group("/groups"))
    // ...

    // Telegram webhook (без JWT)
    api.POST("/telegram/webhook", telegramHandler.Webhook)

    // Asynq worker (горутина)
    go startWorker(cfg, pool, rdb)

    e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

## Health check
```go
// GET /api/v1/health
func healthHandler(pool *pgxpool.Pool, rdb *redis.Client) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Проверить pool.Ping() и rdb.Ping()
        return c.JSON(200, map[string]string{
            "status": "ok", "db": "ok", "redis": "ok",
        })
    }
}
```

## Критерии готовности
- [ ] `make dev` запускает сервер без ошибок
- [ ] `GET /api/v1/health` → 200
- [ ] Запрос без JWT токена → 401 в формате `{ error: { code, message } }`
- [ ] Невалидный body → 400 с описанием
- [ ] Asynq worker стартует в горутине
- [ ] Нет data race (проверить `go test -race ./...`)
