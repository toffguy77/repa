package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

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
	appmw "github.com/repa-app/repa/internal/middleware"
	authsvc "github.com/repa-app/repa/internal/service/auth"
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
)

const (
	testJWTSecret     = "e2e-test-secret-key-32chars-long!"
	testAdminUsername = "admin"
	testAdminPassword = "admin-test-pass"
)

// Suite holds all shared test infrastructure.
type Suite struct {
	ctx     context.Context
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	rdb     *redis.Client
	queries *db.Queries
	echo    *echo.Echo
	server  *httptest.Server
	pgC     testcontainers.Container
	redisC  testcontainers.Container
}

var suite *Suite

// Set during setupSuite so tests can drive the pipeline the worker normally drives.
var (
	suiteRevealService *revealsvc.Service
	suiteGroupsService *groupssvc.Service
	suitePushService   *pushsvc.Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	s, err := setupSuite(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup e2e suite: %v\n", err)
		os.Exit(1)
	}
	suite = s

	code := m.Run()

	suite.teardown(ctx)
	os.Exit(code)
}

func setupSuite(ctx context.Context) (*Suite, error) {
	// Start PostgreSQL container
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("repa_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres: %w", err)
	}

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}

	// Start Redis container
	redisContainer, err := tcredis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start redis: %w", err)
	}

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis connection string: %w", err)
	}

	// Connect to PostgreSQL
	pool, err := pgxpool.New(ctx, pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	sqlDB := stdlib.OpenDBFromPool(pool)

	// Apply migrations
	if err := applyMigrations(ctx, sqlDB); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	// Seed system questions
	if err := seedSystemQuestions(ctx, sqlDB); err != nil {
		return nil, fmt.Errorf("seed questions: %w", err)
	}

	// Connect to Redis
	opts, err := redis.ParseURL(redisConnStr)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	queries := db.New(sqlDB)

	// Build Echo server
	e := buildEchoServer(queries, sqlDB, rdb)

	// Start test HTTP server
	server := httptest.NewServer(e)

	return &Suite{
		ctx:     ctx,
		pool:    pool,
		sqlDB:   sqlDB,
		rdb:     rdb,
		queries: queries,
		echo:    e,
		server:  server,
		pgC:     pgContainer,
		redisC:  redisContainer,
	}, nil
}

func (s *Suite) teardown(ctx context.Context) {
	if s.server != nil {
		s.server.Close()
	}
	if s.sqlDB != nil {
		s.sqlDB.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
	if s.rdb != nil {
		s.rdb.Close()
	}
	if s.pgC != nil {
		s.pgC.Terminate(ctx)
	}
	if s.redisC != nil {
		s.redisC.Terminate(ctx)
	}
}

func applyMigrations(ctx context.Context, sqlDB *sql.DB) error {
	migrationsDir := findMigrationsDir()
	// Keep in step with internal/db/migrations — a missing file shows up as a confusing
	// INTERNAL error from the first handler that touches the new column.
	files := []string{
		"001_init.up.sql",
		"002_groups_categories.up.sql",
		"003_production_fixes.up.sql",
		"004_season_kickoff.up.sql",
		"005_invite_code_short.up.sql",
		"006_share_attribution.up.sql",
		"007_referrals.up.sql",
		"008_detector_hints.up.sql",
		"009_member_safety.up.sql",
		"010_question_tone.up.sql",
	}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if _, err := sqlDB.ExecContext(ctx, string(data)); err != nil {
			return fmt.Errorf("exec migration %s: %w", f, err)
		}
	}
	return nil
}

func findMigrationsDir() string {
	// Walk up from test file to find backend/internal/db/migrations
	candidates := []string{
		"../../internal/db/migrations",
		"../../../backend/internal/db/migrations",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "001_init.up.sql")); err == nil {
			return c
		}
	}
	// Absolute fallback
	return "/Users/thatguy/src/repa/backend/internal/db/migrations"
}

func seedSystemQuestions(ctx context.Context, sqlDB *sql.DB) error {
	categories := []string{"HOT", "FUNNY", "SECRETS", "SKILLS", "ROMANCE", "STUDY"}
	for i, cat := range categories {
		for j := 0; j < 3; j++ {
			text := fmt.Sprintf("System question %d for %s", j+1, cat)
			// Tone is set here rather than left to migration 010: the migration runs before this
			// seed, so it has nothing to classify. Using the production classifier means the test
			// bank's tones are the ones the real bank would get — HOT and SECRETS edgy, SKILLS warm,
			// the rest neutral — which is what makes the kind-only tests meaningful.
			tone := string(questionssvc.ClassifyTone(text, db.QuestionCategory(cat)))
			_, err := sqlDB.ExecContext(ctx,
				`INSERT INTO questions (id, text, category, source, status, tone) VALUES ($1, $2, $3, 'SYSTEM', 'ACTIVE', $4)`,
				uuid.New().String(), text, cat, tone,
			)
			if err != nil {
				return fmt.Errorf("seed question %d-%d: %w", i, j, err)
			}
		}
	}
	return nil
}

func buildEchoServer(queries *db.Queries, sqlDB *sql.DB, rdb *redis.Client) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = appmw.NewValidator()
	e.HTTPErrorHandler = handler.ErrorHandler

	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(appmw.Sanitize())

	cfg := &config.Config{
		JWTSecret:        testJWTSecret,
		DevMode:          true,
		AppMinVersion:    "1.0.0",
		AppLatestVersion: "1.2.0",
		AdminUsername:    testAdminUsername,
		AdminPassword:    testAdminPassword,
	}

	authService := authsvc.NewService(queries, rdb, nil, testJWTSecret, true)
	authHandler := authhandler.NewHandler(authService, cfg)

	groupsService := groupssvc.NewService(queries, sqlDB)
	groupsHandler := groupshandler.NewHandler(groupsService)

	votingService := votingsvc.NewService(queries)
	votingHandler := votinghandler.NewHandler(votingService)

	revealService := revealsvc.NewService(queries, sqlDB)
	revealHandler := revealhandler.NewHandler(revealService, nil) // no cards service in e2e

	// Completing a voting session is what brings a kickoff Reveal forward (see
	// docs/features/voting.md). No asynq client in e2e, so no kickoff push is enqueued.
	votingService.WithKickoff(revealService, nil)

	profileService := profilesvc.NewService(queries)
	profileHandler := profilehandler.NewHandler(profileService)

	chronicleHandler := chroniclehandler.NewHandler(chroniclesvc.NewService(queries))

	crystalsService := crystalssvc.NewService(queries, sqlDB, rdb, nil) // no yukassa
	crystalsHandler := crystalshandler.NewHandler(crystalsService)

	// Free crystal grants: welcome on registration, referral on a completed session.
	authService.WithWelcomeGrant(crystalsService)
	votingService.WithReferralGrants(crystalsService)

	reactionsService := reactionssvc.NewService(queries, nil) // no asynq
	reactionsHandler := reactionshandler.NewHandler(reactionsService)

	questionsService := questionssvc.NewService(queries, nil) // no AI moderator
	questionsHandler := questionshandler.NewHandler(questionsService, queries)

	pushHandler := pushhandler.NewHandler(queries)

	// Telegram with a nil bot client: SendMessage is a no-op without one, which is enough to
	// exercise webhook routing and the database side effects. /connect is not driven through the
	// webhook here because it requires the bot to be a chat administrator.
	safetyService := safetysvc.NewService(queries, sqlDB)
	safetyHandler := safetyhandler.NewHandler(safetyService)
	groupsHandler.WithSafety(safetyService)

	telegramService := telegramsvc.NewService(queries, nil, "https://repa.app")
	telegramHandler := telegramhandler.NewHandler(telegramService, "")

	suiteRevealService = revealService
	suiteGroupsService = groupsService
	// Exposed so the signal debounce can be exercised against a real Redis.
	suitePushService = pushsvc.NewService(queries, rdb, nil)

	// Health endpoint
	api := e.Group("/api/v1")
	api.POST("/telegram/webhook", telegramHandler.Webhook)
	api.GET("/health", func(c echo.Context) error {
		dbStatus := "ok"
		if err := sqlDB.PingContext(c.Request().Context()); err != nil {
			dbStatus = "error"
		}
		redisStatus := "ok"
		if err := rdb.Ping(c.Request().Context()).Err(); err != nil {
			redisStatus = "error"
		}
		status := http.StatusOK
		if dbStatus != "ok" || redisStatus != "ok" {
			status = http.StatusServiceUnavailable
		}
		return c.JSON(status, map[string]any{
			"data": map[string]string{"status": "ok", "db": dbStatus, "redis": redisStatus},
		})
	})

	// Public auth routes
	api.POST("/auth/otp/send", authHandler.OTPSend)
	api.POST("/auth/otp/verify", authHandler.OTPVerify)
	api.GET("/auth/username-check", authHandler.UsernameCheck)
	api.GET("/app/version", authHandler.AppVersion)

	// Protected routes
	protected := api.Group("", appmw.JWTAuth(testJWTSecret))
	protected.GET("/auth/me", authHandler.GetMe)
	protected.PATCH("/auth/profile", authHandler.UpdateProfile)
	protected.POST("/auth/avatar", authHandler.UploadAvatar)
	protected.PATCH("/push/preferences", authHandler.UpdatePushPreferences)
	protected.DELETE("/auth/account", authHandler.DeleteAccount)

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

	protected.GET("/seasons/:seasonId/voting-session", votingHandler.GetVotingSession)
	protected.POST("/seasons/:seasonId/votes", votingHandler.CastVote)
	protected.GET("/seasons/:seasonId/progress", votingHandler.GetProgress)

	protected.GET("/seasons/:seasonId/reveal", revealHandler.GetReveal)
	protected.GET("/seasons/:seasonId/members-cards", revealHandler.GetMembersCards)
	protected.POST("/seasons/:seasonId/reveal/open-hidden", revealHandler.OpenHidden)
	protected.GET("/seasons/:seasonId/anticipation", revealHandler.GetAnticipation)
	protected.GET("/seasons/:seasonId/detector", revealHandler.GetDetector)
	protected.POST("/seasons/:seasonId/detector", revealHandler.BuyDetector)
	protected.POST("/seasons/:seasonId/detector/hint", revealHandler.BuyDetectorHint)
	protected.POST("/seasons/:seasonId/shares", revealHandler.RecordShare)

	protected.GET("/crystals/balance", crystalsHandler.GetBalance)
	protected.GET("/crystals/packages", crystalsHandler.GetPackages)
	protected.GET("/crystals/history", crystalsHandler.GetHistory)
	api.POST("/crystals/purchase/webhook", crystalsHandler.Webhook)

	protected.GET("/groups/:id/members/:userId/profile", profileHandler.GetProfile)
	protected.GET("/groups/:id/chronicle", chronicleHandler.Get)

	protected.POST("/push/register", pushHandler.RegisterToken)

	protected.POST("/seasons/:seasonId/members/:targetId/reactions", reactionsHandler.CreateReaction)
	protected.GET("/seasons/:seasonId/members/:targetId/reactions", reactionsHandler.GetReactions)

	protected.POST("/groups/:groupId/questions", questionsHandler.CreateQuestion)
	protected.GET("/groups/:groupId/questions", questionsHandler.ListQuestions)
	protected.DELETE("/groups/:groupId/questions/:questionId", questionsHandler.DeleteQuestion)
	protected.POST("/groups/:groupId/questions/:questionId/report", questionsHandler.ReportQuestion)

	protected.GET("/groups/:id/next-season/question-candidates", pushHandler.GetQuestionCandidates)
	protected.POST("/groups/:id/next-season/vote-question", pushHandler.VoteQuestion)

	// Admin routes
	adminH := adminhandler.NewHandler(queries, testAdminUsername, testAdminPassword)
	admin := api.Group("/admin", adminH.BasicAuth)
	admin.GET("/reports", adminH.ListReports)
	admin.GET("/user-reports", adminH.ListUserReports)
	admin.PATCH("/reports/:id", adminH.ResolveReport)
	admin.GET("/stats", adminH.GetStats)

	return e
}

// --- Helpers ---

func mintToken(userID, username string) string {
	claims := appmw.JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

type jsonResponse struct {
	StatusCode int
	Body       map[string]any
	RawBody    []byte
}

func doRequest(t *testing.T, method, path string, body any, token string) jsonResponse {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, suite.server.URL+path, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	_ = json.Unmarshal(raw, &result)

	return jsonResponse{
		StatusCode: resp.StatusCode,
		Body:       result,
		RawBody:    raw,
	}
}

func doRequestBasicAuth(t *testing.T, method, path string, body any, user, pass string) jsonResponse {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, suite.server.URL+path, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]any
	_ = json.Unmarshal(raw, &result)

	return jsonResponse{
		StatusCode: resp.StatusCode,
		Body:       result,
		RawBody:    raw,
	}
}

// createTestUser creates a user directly in DB and returns (userID, token).
func createTestUser(t *testing.T, username string) (string, string) {
	t.Helper()
	userID := uuid.New().String()
	uname := shortUsername(username)
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO users (id, username, phone) VALUES ($1, $2, $3)`,
		userID, uname, "+7900"+uuid.New().String()[:7],
	)
	require.NoError(t, err)
	return userID, mintToken(userID, uname)
}

// createTestUserWithBirthYear creates a user with a specific birth year.
func createTestUserWithBirthYear(t *testing.T, username string, birthYear int) (string, string) {
	t.Helper()
	userID := uuid.New().String()
	uname := shortUsername(username)
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO users (id, username, phone, birth_year) VALUES ($1, $2, $3, $4)`,
		userID, uname, "+7900"+uuid.New().String()[:7], birthYear,
	)
	require.NoError(t, err)
	return userID, mintToken(userID, uname)
}

// createTestGroup creates a group via the API and returns group data.
//
// kind_only is sent explicitly as false: test users are created without a birth year, and an unknown
// age defaults the setting to on — which would refuse any group whose categories are all edgy. Tests
// about tone pass their own value; every other test wants an ordinary group.
func createTestGroup(t *testing.T, token string, name string, categories []string) map[string]any {
	return createTestGroupWithKindOnly(t, token, name, categories, false)
}

func createTestGroupWithKindOnly(t *testing.T, token, name string, categories []string, kindOnly bool) map[string]any {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/groups", map[string]any{
		"name":       shortName(name),
		"categories": categories,
		"kind_only":  kindOnly,
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create group failed: %s", string(resp.RawBody))
	data := resp.Body["data"].(map[string]any)
	return data["group"].(map[string]any)
}

// joinGroup joins a user to a group via the API.
func joinGroup(t *testing.T, token string, inviteCode string) {
	t.Helper()
	resp := doRequest(t, "POST", "/api/v1/groups/join/"+inviteCode, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "join group failed: %s", string(resp.RawBody))
}

// createSeasonDirectly creates a VOTING season in DB for a group.
func createSeasonDirectly(t *testing.T, groupID string) string {
	t.Helper()
	seasonID := uuid.New().String()
	now := time.Now()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO seasons (id, group_id, number, status, starts_at, reveal_at, ends_at)
		 VALUES ($1, $2, 1, 'VOTING', $3, $4, $5)`,
		seasonID, groupID, now.Add(-24*time.Hour), now.Add(7*24*time.Hour), now.Add(14*24*time.Hour),
	)
	require.NoError(t, err)
	return seasonID
}

// createRevealedSeason creates a REVEALED season for testing reveal endpoints.
func createRevealedSeason(t *testing.T, groupID string) string {
	t.Helper()
	seasonID := uuid.New().String()
	now := time.Now()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO seasons (id, group_id, number, status, starts_at, reveal_at, ends_at)
		 VALUES ($1, $2, 1, 'REVEALED', $3, $4, $5)`,
		seasonID, groupID, now.Add(-14*24*time.Hour), now.Add(-7*24*time.Hour), now.Add(-1*time.Hour),
	)
	require.NoError(t, err)
	return seasonID
}

// addSeasonQuestions assigns system questions to a season.
func addSeasonQuestions(t *testing.T, seasonID string, count int) []string {
	t.Helper()
	rows, err := suite.sqlDB.QueryContext(context.Background(),
		`SELECT id FROM questions WHERE source = 'SYSTEM' LIMIT $1`, count,
	)
	require.NoError(t, err)
	defer rows.Close()

	var questionIDs []string
	ord := 1
	for rows.Next() {
		var qid string
		require.NoError(t, rows.Scan(&qid))
		_, err := suite.sqlDB.ExecContext(context.Background(),
			`INSERT INTO season_questions (id, season_id, question_id, ord) VALUES ($1, $2, $3, $4)`,
			uuid.New().String(), seasonID, qid, ord,
		)
		require.NoError(t, err)
		questionIDs = append(questionIDs, qid)
		ord++
	}
	return questionIDs
}

// addCrystals grants crystals to a user directly in DB.
func addCrystals(t *testing.T, userID string, amount int32) {
	t.Helper()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`INSERT INTO crystal_logs (id, user_id, delta, balance, type, description)
		 VALUES ($1, $2, $3, $3, 'BONUS', 'test grant')`,
		uuid.New().String(), userID, amount,
	)
	require.NoError(t, err)
}

// createSeasonResults creates aggregated results for a revealed season.
func createSeasonResults(t *testing.T, seasonID string, targetID string, questionIDs []string) {
	t.Helper()
	for _, qid := range questionIDs {
		_, err := suite.sqlDB.ExecContext(context.Background(),
			`INSERT INTO season_results (id, season_id, target_id, question_id, vote_count, total_voters, percentage)
			 VALUES ($1, $2, $3, $4, 2, 3, 66.67)`,
			uuid.New().String(), seasonID, targetID, qid,
		)
		require.NoError(t, err)
	}
}

// shortName truncates a name to fit within the 40-char group name limit.
func shortName(prefix string) string {
	if len(prefix) > 40 {
		return prefix[:40]
	}
	return prefix
}

// shortUsername generates a unique username that fits within 20 chars.
func shortUsername(prefix string) string {
	if len(prefix) <= 20 {
		return prefix
	}
	// Truncate and add short random suffix for uniqueness
	suffix := uuid.New().String()[:4]
	maxPrefix := 20 - len(suffix) - 1
	if maxPrefix < 0 {
		maxPrefix = 0
	}
	if len(prefix) > maxPrefix {
		prefix = prefix[:maxPrefix]
	}
	return prefix + "_" + suffix
}

// getData extracts the "data" field from a response body.
func getData(t *testing.T, resp jsonResponse) map[string]any {
	t.Helper()
	data, ok := resp.Body["data"].(map[string]any)
	require.True(t, ok, "response has no 'data' field: %s", string(resp.RawBody))
	return data
}

// getError extracts the "error" field from a response body.
func getError(t *testing.T, resp jsonResponse) map[string]any {
	t.Helper()
	errObj, ok := resp.Body["error"].(map[string]any)
	require.True(t, ok, "response has no 'error' field: %s", string(resp.RawBody))
	return errObj
}
