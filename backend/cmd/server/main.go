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
	adminhandler "github.com/repa-app/repa/internal/handler/admin"
	authhandler "github.com/repa-app/repa/internal/handler/auth"
	chroniclehandler "github.com/repa-app/repa/internal/handler/chronicle"
	crystalshandler "github.com/repa-app/repa/internal/handler/crystals"
	groupshandler "github.com/repa-app/repa/internal/handler/groups"
	profilehandler "github.com/repa-app/repa/internal/handler/profile"
	pushhandler "github.com/repa-app/repa/internal/handler/push"
	questionshandler "github.com/repa-app/repa/internal/handler/questions"
	reactionshandler "github.com/repa-app/repa/internal/handler/reactions"
	revealhandler "github.com/repa-app/repa/internal/handler/reveal"
	safetyhandler "github.com/repa-app/repa/internal/handler/safety"
	telegramhandler "github.com/repa-app/repa/internal/handler/telegram"
	votinghandler "github.com/repa-app/repa/internal/handler/voting"
	"github.com/repa-app/repa/internal/lib"
	appmw "github.com/repa-app/repa/internal/middleware"
	achievesvc "github.com/repa-app/repa/internal/service/achievements"
	authsvc "github.com/repa-app/repa/internal/service/auth"
	cardssvc "github.com/repa-app/repa/internal/service/cards"
	chroniclesvc "github.com/repa-app/repa/internal/service/chronicle"
	crystalssvc "github.com/repa-app/repa/internal/service/crystals"
	groupssvc "github.com/repa-app/repa/internal/service/groups"
	profilesvc "github.com/repa-app/repa/internal/service/profile"
	pushsvc "github.com/repa-app/repa/internal/service/push"
	questionssvc "github.com/repa-app/repa/internal/service/questions"
	reactionssvc "github.com/repa-app/repa/internal/service/reactions"
	revealsvc "github.com/repa-app/repa/internal/service/reveal"
	safetysvc "github.com/repa-app/repa/internal/service/safety"
	telegramsvc "github.com/repa-app/repa/internal/service/telegram"
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
		AllowOrigins: []string{"https://repa.app", "https://*.repa.app"},
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
	e.Use(appmw.Sanitize())

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
	// A completed voting session is what brings a new group's kickoff Reveal forward, so
	// the voting service needs the reveal service's eligibility predicate.
	votingService.WithKickoff(revealService, asynqClient)
	achieveService := achievesvc.NewService(queries)
	cardsService := cardssvc.NewService(queries, s3Client)
	revealHandler := revealhandler.NewHandler(revealService, cardsService)
	safetyService := safetysvc.NewService(queries, sqlDB)
	safetyHandler := safetyhandler.NewHandler(safetyService)
	// Leaving permanently is a flag on the leave route, so the groups handler needs the safety
	// service to record it.
	groupsHandler.WithSafety(safetyService)

	profileService := profilesvc.NewService(queries)
	profileHandler := profilehandler.NewHandler(profileService)

	chronicleService := chroniclesvc.NewService(queries)
	chronicleHandler := chroniclehandler.NewHandler(chronicleService)

	var yukassaClient *lib.YukassaClient
	if cfg.YukassaShopID != "" {
		yukassaClient = lib.NewYukassaClient(cfg.YukassaShopID, cfg.YukassaSecret, cfg.YukassaReturn)
	}
	crystalsService := crystalssvc.NewService(queries, sqlDB, rdb, yukassaClient)
	crystalsHandler := crystalshandler.NewHandler(crystalsService)

	// Free crystal grants. Wired after the crystals service exists, so the collaborators that
	// trigger a grant (registration, a completed session, an awarded achievement) can reach it.
	authService.WithWelcomeGrant(crystalsService)
	votingService.WithReferralGrants(crystalsService)
	achieveService.WithGrants(crystalsService)

	// FCM + Push
	var fcmClient *lib.FCMClient
	if cfg.FirebaseProjectID != "" {
		var fcmErr error
		fcmClient, fcmErr = lib.NewFCMClient(ctx, cfg.FirebaseProjectID, cfg.FirebasePrivateKey, cfg.FirebaseClientEmail, queries)
		if fcmErr != nil {
			log.Warn().Err(fcmErr).Msg("FCM client not available, push notifications disabled")
		} else {
			log.Info().Msg("FCM client initialized")
		}
	}
	pushService := pushsvc.NewService(queries, rdb, fcmClient)
	pushHandler := pushhandler.NewHandler(queries)

	reactionsService := reactionssvc.NewService(queries, asynqClient)
	reactionsHandler := reactionshandler.NewHandler(reactionsService)

	// Questions + AI moderation
	var anthropicClient *lib.AnthropicClient
	if cfg.AnthropicKey != "" {
		anthropicClient = lib.NewAnthropicClient(cfg.AnthropicKey)
		log.Info().Msg("Anthropic client initialized for AI moderation")
	}
	var moderator questionssvc.Moderator
	if anthropicClient != nil {
		moderator = anthropicClient
	}
	questionsService := questionssvc.NewService(queries, moderator)
	questionsHandler := questionshandler.NewHandler(questionsService, queries)

	// Telegram
	var telegramService *telegramsvc.Service
	var telegramHandler *telegramhandler.Handler
	if cfg.TelegramToken != "" {
		tgClient, tgErr := lib.NewTelegramClient(cfg.TelegramToken)
		if tgErr != nil {
			log.Warn().Err(tgErr).Msg("Telegram client not available")
		} else {
			log.Info().Msg("Telegram client initialized")
			telegramService = telegramsvc.NewService(queries, tgClient, cfg.AppBaseURL)
			telegramHandler = telegramhandler.NewHandler(telegramService, cfg.TelegramSecret)
		}
	}

	// Well-known routes (Universal Links / App Links)
	e.GET("/.well-known/apple-app-site-association", wellKnownApple)
	e.GET("/.well-known/assetlinks.json", wellKnownAndroid)

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
	protected.DELETE("/groups/:id/members/:userId", safetyHandler.RemoveMember)
	protected.POST("/members/:userId/block", safetyHandler.BlockMember)
	protected.DELETE("/members/:userId/block", safetyHandler.UnblockMember)
	protected.POST("/members/:userId/report", safetyHandler.ReportMember)
	protected.PATCH("/groups/:id", groupsHandler.UpdateGroup)
	protected.POST("/groups/:id/invite-link", groupsHandler.RegenerateInviteLink)

	// Voting routes
	protected.GET("/seasons/:seasonId/voting-session", votingHandler.GetVotingSession)
	protected.POST("/seasons/:seasonId/votes", votingHandler.CastVote,
		appmw.RateLimit(rdb, "vote", 30, time.Minute))
	protected.GET("/seasons/:seasonId/progress", votingHandler.GetProgress)

	// Reveal routes
	protected.GET("/seasons/:seasonId/reveal", revealHandler.GetReveal)
	protected.GET("/seasons/:seasonId/members-cards", revealHandler.GetMembersCards)
	protected.POST("/seasons/:seasonId/reveal/open-hidden", revealHandler.OpenHidden)
	protected.GET("/seasons/:seasonId/my-card-url", revealHandler.GetMyCardURL,
		appmw.RateLimit(rdb, "card", 5, time.Hour))
	protected.GET("/seasons/:seasonId/detector", revealHandler.GetDetector)
	protected.POST("/seasons/:seasonId/detector", revealHandler.BuyDetector)
	protected.GET("/seasons/:seasonId/anticipation", revealHandler.GetAnticipation)
	protected.POST("/seasons/:seasonId/detector/hint", revealHandler.BuyDetectorHint)
	protected.POST("/seasons/:seasonId/shares", revealHandler.RecordShare)

	// Crystal routes
	protected.GET("/crystals/balance", crystalsHandler.GetBalance)
	protected.GET("/crystals/packages", crystalsHandler.GetPackages)
	protected.GET("/crystals/history", crystalsHandler.GetHistory)
	protected.POST("/crystals/purchase/init", crystalsHandler.InitPurchase)
	protected.GET("/crystals/purchase/verify/:paymentId", crystalsHandler.VerifyPurchase)
	// Webhook — no JWT (called by YuKassa), IP-restricted
	api.POST("/crystals/purchase/webhook", crystalsHandler.Webhook, appmw.YukassaIPAllowlist())

	// Profile routes
	protected.GET("/groups/:id/members/:userId/profile", profileHandler.GetProfile)
	protected.GET("/groups/:id/chronicle", chronicleHandler.Get)

	// Push routes
	protected.POST("/push/register", pushHandler.RegisterToken)

	// Reaction routes
	protected.POST("/seasons/:seasonId/members/:targetId/reactions", reactionsHandler.CreateReaction)
	protected.GET("/seasons/:seasonId/members/:targetId/reactions", reactionsHandler.GetReactions)

	// Question moderation routes
	protected.POST("/groups/:groupId/questions", questionsHandler.CreateQuestion,
		appmw.RateLimit(rdb, "question", 10, time.Hour))
	protected.GET("/groups/:groupId/questions", questionsHandler.ListQuestions)
	protected.DELETE("/groups/:groupId/questions/:questionId", questionsHandler.DeleteQuestion)
	protected.POST("/groups/:groupId/questions/:questionId/report", questionsHandler.ReportQuestion)

	// Next-season question voting
	protected.GET("/groups/:id/next-season/question-candidates", pushHandler.GetQuestionCandidates)
	protected.POST("/groups/:id/next-season/vote-question", pushHandler.VoteQuestion)

	// Admin routes (Basic Auth)
	adminHandler := adminhandler.NewHandler(queries, cfg.AdminUsername, cfg.AdminPassword)
	admin := api.Group("/admin", adminHandler.BasicAuth)
	admin.GET("/reports", adminHandler.ListReports)
	admin.GET("/user-reports", adminHandler.ListUserReports)
	admin.PATCH("/reports/:id", adminHandler.ResolveReport)
	admin.GET("/stats", adminHandler.GetStats)

	// Telegram routes
	if telegramHandler != nil {
		api.POST("/telegram/webhook", telegramHandler.Webhook)
		protected.POST("/groups/:id/telegram/generate-code", telegramHandler.GenerateCode)
		protected.DELETE("/groups/:id/telegram", telegramHandler.Disconnect)
		protected.POST("/seasons/:seasonId/share-to-telegram", telegramHandler.ShareToTelegram)
	}

	// Asynq worker
	go startWorker(cfg, revealService, achieveService, cardsService, pushService, telegramService, groupsService, asynqClient)

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

func wellKnownApple(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"applinks": map[string]any{
			"apps": []string{},
			"details": []map[string]any{
				{
					"appID": "3GWRDGA8B7.app.repa.repa",
					"paths": []string{"/join/*", "/invite/*"},
				},
			},
		},
	})
}

func wellKnownAndroid(c echo.Context) error {
	return c.JSON(http.StatusOK, []map[string]any{
		{
			"relation": []string{"delegate_permission/common.handle_all_urls"},
			"target": map[string]any{
				"namespace":    "android_app",
				"package_name": "app.repa.repa",
				"sha256_cert_fingerprints": []string{
					"86:20:33:3D:E5:38:74:C5:79:72:03:41:46:39:FE:AA:B8:FC:3A:29:45:FA:BC:95:C0:0D:23:87:AC:02:87:A3",
					"91:34:01:79:20:DD:10:5F:08:86:24:AD:71:2B:80:57:DA:5B:CD:57:45:5E:82:BE:51:F5:98:77:09:4C:E2:5C",
				},
			},
		},
	})
}

func startWorker(cfg *config.Config, revealSvc *revealsvc.Service, achieveSvc *achievesvc.Service, cardsSvc *cardssvc.Service, pushSvc *pushsvc.Service, telegramSvc *telegramsvc.Service, groupsSvc *groupssvc.Service, asynqClient *asynq.Client) {
	srv, err := lib.NewAsynqServer(cfg.RedisURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to create asynq server")
		return
	}

	revealChecker := tasks.NewRevealChecker(revealSvc, asynqClient)
	revealProcessor := tasks.NewRevealProcessor(revealSvc, asynqClient)
	achieveProcessor := tasks.NewAchievementsProcessor(achieveSvc)
	cardsProcessor := tasks.NewCardsProcessor(cardsSvc)
	pushProcessor := tasks.NewPushProcessor(pushSvc)

	mux := asynq.NewServeMux()
	mux.HandleFunc(lib.TypeRevealChecker, revealChecker.HandleRevealChecker)
	mux.HandleFunc(lib.TypeRevealProcess, revealProcessor.HandleRevealProcess)
	mux.HandleFunc(lib.TypeAchievements, achieveProcessor.HandleAchievements)
	mux.HandleFunc(lib.TypeCardsGenerate, cardsProcessor.HandleCardsGenerate)
	mux.HandleFunc(lib.TypePushWeekly, pushProcessor.HandleWeeklyScheduler)
	mux.HandleFunc(lib.TypePushTuesday, pushProcessor.HandleTuesdaySignal)
	mux.HandleFunc(lib.TypePushWednesday, pushProcessor.HandleWednesdayQuorum)
	mux.HandleFunc(lib.TypePushThursday, pushProcessor.HandleThursdayTeaser)
	mux.HandleFunc(lib.TypePushFriPreReveal, pushProcessor.HandleFridayPreReveal)
	mux.HandleFunc(lib.TypePushReveal, pushProcessor.HandleRevealNotification)
	mux.HandleFunc(lib.TypePushSundayPrev, pushProcessor.HandleSundayPreview)
	mux.HandleFunc(lib.TypePushSundayStreak, pushProcessor.HandleSundayStreak)
	mux.HandleFunc(lib.TypeReactionPush, pushProcessor.HandleReactionPush)
	mux.HandleFunc(lib.TypePushPostponed, pushProcessor.HandleRevealPostponed)
	mux.HandleFunc(lib.TypePushKickoff, pushProcessor.HandleKickoffScheduled)
	mux.HandleFunc(lib.TypePushVoteSignal, pushProcessor.HandleVoteSignal)

	seasonCreator := tasks.NewSeasonCreator(groupsSvc)
	mux.HandleFunc(lib.TypeSeasonCreator, seasonCreator.HandleSeasonCreator)
	mux.HandleFunc(lib.TypeSeasonMaintain, seasonCreator.HandleSeasonMaintain)
	mux.HandleFunc(lib.TypeSeasonForGroup, seasonCreator.HandleSeasonForGroup)

	if telegramSvc != nil {
		telegramProcessor := tasks.NewTelegramProcessor(telegramSvc)
		mux.HandleFunc(lib.TypeTelegramStart, telegramProcessor.HandleSeasonStart)
		mux.HandleFunc(lib.TypeTelegramReveal, telegramProcessor.HandleRevealPost)
		mux.HandleFunc(lib.TypeTelegramShare, telegramProcessor.HandleShareCard)
	}

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
	registerCron(scheduler, "* * * * *", lib.TypeRevealChecker, "critical", "reveal-checker (every minute)")

	// Push schedule (all times in UTC, MSK = UTC+3)
	registerCron(scheduler, "0 14 * * 1", lib.TypePushWeekly, "default", "weekly-scheduler (Mon 17:00 MSK)")
	registerCron(scheduler, "0 16 * * 2", lib.TypePushTuesday, "default", "tuesday-signal (Tue 19:00 MSK)")
	registerCron(scheduler, "0 15 * * 3", lib.TypePushWednesday, "default", "wednesday-quorum (Wed 18:00 MSK)")
	registerCron(scheduler, "0 17 * * 4", lib.TypePushThursday, "default", "thursday-teaser (Thu 20:00 MSK)")
	registerCron(scheduler, "0 16 * * 5", lib.TypePushFriPreReveal, "default", "friday-pre-reveal (Fri 19:00 MSK)")
	registerCron(scheduler, "0 9 * * 0", lib.TypePushSundayPrev, "default", "sunday-preview (Sun 12:00 MSK)")
	registerCron(scheduler, "0 15 * * 0", lib.TypePushSundayStreak, "default", "sunday-streak (Sun 18:00 MSK)")

	// Season creator: Sunday 18:00 UTC (21:00 MSK) — create new seasons for active groups
	registerCron(scheduler, "0 18 * * 0", lib.TypeSeasonCreator, "critical", "season-creator (Sun 21:00 MSK)")
	registerCron(scheduler, "7 * * * *", lib.TypeSeasonMaintain, "critical", "season-maintenance (hourly)")

	// Telegram season-start: same time as weekly push (Mon 17:00 MSK)
	registerCron(scheduler, "0 14 * * 1", lib.TypeTelegramStart, "default", "telegram-season-start (Mon 17:00 MSK)")

	if err := scheduler.Run(); err != nil {
		log.Error().Err(err).Msg("asynq scheduler error")
	}
}

func registerCron(scheduler *asynq.Scheduler, cronExpr, taskType, queue, label string) {
	task := asynq.NewTask(taskType, nil)
	_, err := scheduler.Register(cronExpr, task, asynq.Queue(queue))
	if err != nil {
		log.Error().Err(err).Str("task", taskType).Msg("failed to register cron schedule")
	} else {
		log.Info().Msgf("registered cron: %s", label)
	}
}
