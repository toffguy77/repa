package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/repa-app/repa/internal/config"
	"github.com/repa-app/repa/internal/handler"
	"github.com/repa-app/repa/internal/lib"
	appmw "github.com/repa-app/repa/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.Load()
	ctx := context.Background()

	// Database
	pool, err := lib.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("connected to database")

	sqlDB := lib.NewDBFromPool(pool)
	defer sqlDB.Close()

	// Redis
	rdb, err := lib.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()
	log.Info().Msg("connected to redis")

	// Asynq client
	asynqClient, err := lib.NewAsynqClient(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create asynq client")
	}
	defer asynqClient.Close()

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.Validator = appmw.NewValidator()
	e.HTTPErrorHandler = handler.ErrorHandler

	// Global middleware
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Routes
	api := e.Group("/api/v1")
	api.GET("/health", healthHandler(pool, rdb))

	// Protected routes (placeholder for future handlers)
	_ = api.Group("", appmw.JWTAuth(cfg.JWTSecret))

	// Keep references for future wiring
	_ = sqlDB
	_ = asynqClient

	// Asynq worker
	go startWorker(cfg)

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("server stopped")
}

func healthHandler(pool *pgxpool.Pool, rdb *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		dbStatus := "ok"
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "error"
		}

		redisStatus := "ok"
		if err := rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "error"
		}

		status := http.StatusOK
		if dbStatus != "ok" || redisStatus != "ok" {
			status = http.StatusServiceUnavailable
		}

		return c.JSON(status, map[string]any{
			"data": map[string]string{
				"status": "ok",
				"db":     dbStatus,
				"redis":  redisStatus,
			},
		})
	}
}

func startWorker(cfg *config.Config) {
	srv, err := lib.NewAsynqServer(cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to create asynq server")
		return
	}

	mux := asynq.NewServeMux()
	// Task handlers will be registered here in future tasks

	if err := srv.Run(mux); err != nil {
		log.Error().Err(err).Msg("asynq worker error")
	}
}

