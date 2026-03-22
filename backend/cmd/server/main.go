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
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/handler"
	authhandler "github.com/repa-app/repa/internal/handler/auth"
	groupshandler "github.com/repa-app/repa/internal/handler/groups"
	revealhandler "github.com/repa-app/repa/internal/handler/reveal"
	votinghandler "github.com/repa-app/repa/internal/handler/voting"
	"github.com/repa-app/repa/internal/lib"
	appmw "github.com/repa-app/repa/internal/middleware"
	authsvc "github.com/repa-app/repa/internal/service/auth"
	groupssvc "github.com/repa-app/repa/internal/service/groups"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
	votingsvc "github.com/repa-app/repa/internal/service/voting"
	"github.com/repa-app/repa/internal/worker/tasks"
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

	// Services
	queries := db.New(sqlDB)

	var s3Client *lib.S3Client
	if cfg.S3AccessKey != "" {
		var s3Err error
		s3Client, s3Err = lib.NewS3Client(cfg.S3Endpoint, cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
		if s3Err != nil {
			log.Warn().Err(s3Err).Msg("S3 client not available, avatar uploads disabled")
		}
	}

	authService := authsvc.NewService(queries, rdb, s3Client, cfg.JWTSecret, cfg.DevMode)
	authHandler := authhandler.NewHandler(authService, cfg)

	groupsService := groupssvc.NewService(queries, sqlDB)
	groupsHandler := groupshandler.NewHandler(groupsService)

	votingService := votingsvc.NewService(queries)
	votingHandler := votinghandler.NewHandler(votingService)

	revealService := revealsvc.NewService(queries, sqlDB)
	revealHandler := revealhandler.NewHandler(revealService)

	// Routes
	api := e.Group("/api/v1")
	api.GET("/health", healthHandler(pool, rdb))

	// Public auth routes
	api.POST("/auth/apple", authHandler.AppleAuth)
	api.POST("/auth/google", authHandler.GoogleAuth)
	api.POST("/auth/otp/send", authHandler.OTPSend)
	api.POST("/auth/otp/verify", authHandler.OTPVerify)
	api.GET("/auth/username-check", authHandler.UsernameCheck,
		appmw.RateLimit(rdb, "username-check", 20, time.Minute))
	api.GET("/app/version", authHandler.AppVersion)

	// Protected routes
	protected := api.Group("", appmw.JWTAuth(cfg.JWTSecret))
	protected.GET("/auth/me", authHandler.GetMe)
	protected.PATCH("/auth/profile", authHandler.UpdateProfile)
	protected.POST("/auth/avatar", authHandler.UploadAvatar)
	protected.PATCH("/push/preferences", authHandler.UpdatePushPreferences)
	protected.DELETE("/auth/account", authHandler.DeleteAccount)

	// Group routes (specific paths before :id wildcard)
	protected.POST("/groups", groupsHandler.CreateGroup)
	protected.GET("/groups", groupsHandler.ListGroups)
	protected.GET("/groups/join/:inviteCode/preview", groupsHandler.JoinPreview)
	protected.POST("/groups/join/:inviteCode", groupsHandler.JoinGroup)
	protected.GET("/groups/:id", groupsHandler.GetGroup)
	protected.DELETE("/groups/:id/leave", groupsHandler.LeaveGroup)
	protected.PATCH("/groups/:id", groupsHandler.UpdateGroup)
	protected.POST("/groups/:id/invite-link", groupsHandler.RegenerateInviteLink)

	// Voting routes
	protected.GET("/seasons/:seasonId/voting-session", votingHandler.GetVotingSession)
	protected.POST("/seasons/:seasonId/votes", votingHandler.CastVote)
	protected.GET("/seasons/:seasonId/progress", votingHandler.GetProgress)

	// Reveal routes
	protected.GET("/seasons/:seasonId/reveal", revealHandler.GetReveal)
	protected.GET("/seasons/:seasonId/members-cards", revealHandler.GetMembersCards)
	protected.POST("/seasons/:seasonId/reveal/open-hidden", revealHandler.OpenHidden)

	// Asynq worker
	go startWorker(cfg, revealService, asynqClient)

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

func startWorker(cfg *config.Config, revealSvc *revealsvc.Service, asynqClient *asynq.Client) {
	srv, err := lib.NewAsynqServer(cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to create asynq server")
		return
	}

	revealChecker := tasks.NewRevealChecker(revealSvc, asynqClient)
	revealProcessor := tasks.NewRevealProcessor(revealSvc, asynqClient)

	mux := asynq.NewServeMux()
	mux.HandleFunc(lib.TypeRevealChecker, revealChecker.HandleRevealChecker)
	mux.HandleFunc(lib.TypeRevealProcess, revealProcessor.HandleRevealProcess)

	// Start asynq scheduler for periodic tasks
	go startScheduler(cfg)

	if err := srv.Run(mux); err != nil {
		log.Error().Err(err).Msg("asynq worker error")
	}
}

func startScheduler(cfg *config.Config) {
	opts, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse redis URI for scheduler")
		return
	}

	scheduler := asynq.NewScheduler(opts, nil)

	// reveal-checker: every minute, check for seasons ready for reveal
	task := asynq.NewTask(lib.TypeRevealChecker, nil)
	_, err = scheduler.Register("* * * * *", task, asynq.Queue("critical"))
	if err != nil {
		log.Error().Err(err).Msg("failed to register reveal-checker schedule")
		return
	}
	log.Info().Msg("registered reveal-checker cron (every minute)")

	if err := scheduler.Run(); err != nil {
		log.Error().Err(err).Msg("asynq scheduler error")
	}
}

