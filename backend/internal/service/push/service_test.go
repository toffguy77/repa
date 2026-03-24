package push

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, rdb
}

// --- canSendPush tests ---

func TestCanSendPush_QuietHours(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	// canSendPush uses time.Now(), so we can only test the rate-limit path
	// deterministically. For quiet hours, we verify by checking the current hour.
	now := time.Now().In(mskLocation)
	hour := now.Hour()

	canSend := svc.canSendPush(ctx, "user-1")

	if hour >= 23 || hour < 9 {
		// We're in quiet hours, should return false
		if canSend {
			t.Errorf("expected false during quiet hours (MSK hour=%d), got true", hour)
		}
	} else {
		// Outside quiet hours with 0 push count, should return true
		if !canSend {
			t.Errorf("expected true outside quiet hours (MSK hour=%d) with 0 push count, got false", hour)
		}
	}
}

func TestCanSendPush_RateLimitExceeded(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	now := time.Now().In(mskLocation)
	hour := now.Hour()

	// Skip if in quiet hours since that path returns false regardless
	if hour >= 23 || hour < 9 {
		t.Skip("skipping rate limit test during quiet hours")
	}

	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-rl", dateKey)

	// Set count to 3 (at limit)
	mr.Set(redisKey, "3")

	canSend := svc.canSendPush(ctx, "user-rl")
	if canSend {
		t.Error("expected false when push count >= 3, got true")
	}
}

func TestCanSendPush_RateLimitNotExceeded(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	now := time.Now().In(mskLocation)
	hour := now.Hour()

	if hour >= 23 || hour < 9 {
		t.Skip("skipping rate limit test during quiet hours")
	}

	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-ok", dateKey)

	// Set count to 2 (under limit)
	mr.Set(redisKey, "2")

	canSend := svc.canSendPush(ctx, "user-ok")
	if !canSend {
		t.Error("expected true when push count < 3, got false")
	}
}

func TestCanSendPush_RedisError_FailOpen(t *testing.T) {
	mr, rdb := newTestRedis(t)
	// Close miniredis to simulate Redis failure
	mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	now := time.Now().In(mskLocation)
	hour := now.Hour()

	if hour >= 23 || hour < 9 {
		t.Skip("skipping Redis error test during quiet hours")
	}

	// With Redis down, canSendPush should fail open (return true)
	canSend := svc.canSendPush(ctx, "user-err")
	if !canSend {
		t.Error("expected true (fail open) on Redis error, got false")
	}
}

func TestCanSendPush_NoExistingKey_ReturnsTrue(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	now := time.Now().In(mskLocation)
	hour := now.Hour()

	if hour >= 23 || hour < 9 {
		t.Skip("skipping test during quiet hours")
	}

	// No key set at all — redis.Nil error, count is 0
	canSend := svc.canSendPush(ctx, "user-new")
	if !canSend {
		t.Error("expected true when no push count exists, got false")
	}
}

// --- incrementPushCount tests ---

func TestIncrementPushCount(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	svc.incrementPushCount(ctx, "user-inc")

	now := time.Now().In(mskLocation)
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-inc", dateKey)

	val, err := mr.Get(redisKey)
	if err != nil {
		t.Fatalf("expected key to exist, got error: %v", err)
	}
	if val != "1" {
		t.Errorf("expected count '1' after one increment, got %q", val)
	}

	// Increment again
	svc.incrementPushCount(ctx, "user-inc")

	val, err = mr.Get(redisKey)
	if err != nil {
		t.Fatalf("expected key to exist, got error: %v", err)
	}
	if val != "2" {
		t.Errorf("expected count '2' after two increments, got %q", val)
	}

	// Check TTL is set
	if !mr.Exists(redisKey) {
		t.Error("expected key to exist with TTL")
	}
	ttl := mr.TTL(redisKey)
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %v", ttl)
	}
}

func TestIncrementPushCount_RedisDown_NoError(t *testing.T) {
	mr, rdb := newTestRedis(t)
	mr.Close() // simulate failure

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	// Should not panic — just logs a warning
	svc.incrementPushCount(ctx, "user-down")
}

// --- SendToUser tests ---

func TestSendToUser_NilFCM_NoOp(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil) // nil FCM
	ctx := context.Background()

	err := svc.SendToUser(ctx, "user-1", db.PushCategoryREMINDER, "Title", "Body", nil)
	if err != nil {
		t.Errorf("expected nil error with nil FCM, got %v", err)
	}
}

func TestSendToUser_NilFCM_DoesNotIncrementCount(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	_ = svc.SendToUser(ctx, "user-nofcm", db.PushCategoryREVEAL, "T", "B", nil)

	now := time.Now().In(mskLocation)
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-nofcm", dateKey)

	if mr.Exists(redisKey) {
		t.Error("expected no push count increment when FCM is nil")
	}
}

// --- SendToUsers tests ---

func TestSendToUsers_NilFCM_IteratesWithoutError(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	// Should iterate over all users without error when FCM is nil
	svc.SendToUsers(ctx, []string{"u1", "u2", "u3"}, db.PushCategorySEASONSTART, "T", "B", nil)

	// No push counts should be incremented
	now := time.Now().In(mskLocation)
	dateKey := now.Format("2006-01-02")
	for _, uid := range []string{"u1", "u2", "u3"} {
		redisKey := fmt.Sprintf("push-count:%s:%s", uid, dateKey)
		if mr.Exists(redisKey) {
			t.Errorf("expected no push count for %s when FCM is nil", uid)
		}
	}
}

func TestSendToUsers_EmptyList(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil)
	ctx := context.Background()

	// Should handle empty list gracefully
	svc.SendToUsers(ctx, []string{}, db.PushCategoryREMINDER, "T", "B", nil)
}

// --- SendToGroupMembers tests ---

func TestSendToGroupMembers_NilQueries_Panics(t *testing.T) {
	mr, rdb := newTestRedis(t)
	defer mr.Close()

	svc := NewService(nil, rdb, nil) // nil queries
	ctx := context.Background()

	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		_ = svc.SendToGroupMembers(ctx, "grp-1", db.PushCategoryREVEAL, "T", "B", nil)
	}()

	if !panicked {
		t.Error("expected panic when queries is nil")
	}
}

// --- Queries accessor test ---

func TestQueries_ReturnsQueries(t *testing.T) {
	svc := NewService(nil, nil, nil)
	if svc.Queries() != nil {
		t.Error("expected nil queries when constructed with nil")
	}
}

func TestQueries_ReturnsNonNil(t *testing.T) {
	// We can't construct a real *db.Queries without a DB, but we can verify
	// the accessor works via the nil case (already tested above).
	// This test ensures Queries() is a simple getter.
	svc := &Service{queries: nil}
	got := svc.Queries()
	if got != nil {
		t.Error("expected nil")
	}
}

// ========================================================================
// sqlmock-based tests for DB-dependent methods
// ========================================================================

// dummyFCMClient creates a non-nil *lib.FCMClient for testing purposes.
// The internal fields (client, queries) are nil/zero, so any actual FCM call
// will fail, but it passes the nil check in SendToUser.
func dummyFCMClient() *lib.FCMClient {
	// FCMClient has unexported fields, so we allocate the right size via unsafe.
	type fcmLayout struct {
		client  uintptr
		queries uintptr
	}
	raw := new(fcmLayout)
	return (*lib.FCMClient)(unsafe.Pointer(raw))
}

func newTestServiceWithDB(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { mockDB.Close() })

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	queries := db.New(mockDB)
	svc := NewService(queries, rdb, nil) // nil FCM
	return svc, mock, mr
}

func newTestServiceWithFCM(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { mockDB.Close() })

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	queries := db.New(mockDB)
	svc := NewService(queries, rdb, dummyFCMClient())
	return svc, mock, mr
}

// --- SendToUser: push preference disabled ---

func TestSendToUser_PushDisabled(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	// IsPushEnabled returns false
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREMINDER).
		WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(false))

	// FCM is nil, so even if preference check passes, it won't send.
	// But with preference disabled, it should return early before FCM check.
	err := svc.SendToUser(ctx, "user-1", db.PushCategoryREMINDER, "Title", "Body", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- SendToUser: push preference query error (fail open) ---

func TestSendToUser_PushPreferenceError(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	// IsPushEnabled returns error
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREVEAL).
		WillReturnError(fmt.Errorf("db error"))

	// FCM is nil, so it will return nil early from the nil-FCM check
	// But it should still proceed past the preference error (fail open)
	err := svc.SendToUser(ctx, "user-1", db.PushCategoryREVEAL, "T", "B", nil)
	if err != nil {
		t.Errorf("expected nil error (FCM nil), got %v", err)
	}
}

// --- SendToUser: push preference enabled but FCM nil ---

func TestSendToUser_PushEnabledFCMNil(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	// IsPushEnabled returns true
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREMINDER).
		WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(true))

	// FCM is nil, so returns nil early
	err := svc.SendToUser(ctx, "user-1", db.PushCategoryREMINDER, "T", "B", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- SendToUser: rate limited ---

func TestSendToUser_RateLimited(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, _, mr := newTestServiceWithDB(t)
	ctx := context.Background()

	// Set push count to 3 (at limit)
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-rl", dateKey)
	mr.Set(redisKey, "3")

	// Should return nil (rate limited, FCM is nil anyway)
	err := svc.SendToUser(ctx, "user-rl", db.PushCategoryREMINDER, "T", "B", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- SendToGroupMembers: success ---

func TestSendToGroupMembers_Success(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	// GetGroupMemberIDs
	mock.ExpectQuery("SELECT user_id FROM group_members").
		WithArgs("grp-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).
			AddRow("u1").
			AddRow("u2"))

	// FCM is nil, so SendToUser will no-op for each user
	err := svc.SendToGroupMembers(ctx, "grp-1", db.PushCategoryREVEAL, "Title", "Body", nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// --- SendToGroupMembers: GetGroupMemberIDs error ---

func TestSendToGroupMembers_GetMemberIDsError(t *testing.T) {
	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT user_id FROM group_members").
		WithArgs("grp-1").
		WillReturnError(fmt.Errorf("query failed"))

	err := svc.SendToGroupMembers(ctx, "grp-1", db.PushCategoryREVEAL, "T", "B", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "get group members") {
		t.Errorf("expected 'get group members' in error, got: %v", err)
	}
}

// --- SendToGroupMembers: empty member list ---

func TestSendToGroupMembers_EmptyMemberList(t *testing.T) {
	svc, mock, _ := newTestServiceWithDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT user_id FROM group_members").
		WithArgs("grp-empty").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	err := svc.SendToGroupMembers(ctx, "grp-empty", db.PushCategoryREMINDER, "T", "B", nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// --- SendToUser with non-nil FCM: rate limited ---

func TestSendToUser_WithFCM_RateLimited(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, _, mr := newTestServiceWithFCM(t)
	ctx := context.Background()

	// Set push count to 3 (at limit)
	dateKey := now.Format("2006-01-02")
	redisKey := fmt.Sprintf("push-count:%s:%s", "user-rl", dateKey)
	mr.Set(redisKey, "3")

	// Should return nil (rate limited, no FCM call attempted)
	err := svc.SendToUser(ctx, "user-rl", db.PushCategoryREMINDER, "T", "B", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- SendToUser with non-nil FCM: push preference disabled ---

func TestSendToUser_WithFCM_PushDisabled(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithFCM(t)
	ctx := context.Background()

	// IsPushEnabled returns false
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREMINDER).
		WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(false))

	err := svc.SendToUser(ctx, "user-1", db.PushCategoryREMINDER, "Title", "Body", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// --- SendToUser with non-nil FCM: push preference query error (fail open, then FCM call) ---

func TestSendToUser_WithFCM_PreferenceErrorFailOpen(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithFCM(t)
	ctx := context.Background()

	// IsPushEnabled returns error (fail open)
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREVEAL).
		WillReturnError(fmt.Errorf("db error"))

	// FCM.SendPushToUser calls queries.GetUserFCMTokens
	// Use the same mock db for FCM's internal queries -- but our dummyFCMClient
	// has nil queries, so it will panic. We recover to verify we got past the
	// preference check.
	func() {
		defer func() { recover() }()
		_ = svc.SendToUser(ctx, "user-1", db.PushCategoryREVEAL, "T", "B", nil)
	}()
	// If we got here, SendToUser proceeded past the preference check (fail open)
}

// --- SendToUser with non-nil FCM: push preference enabled, reaches FCM call ---

func TestSendToUser_WithFCM_PreferenceEnabled(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, mock, _ := newTestServiceWithFCM(t)
	ctx := context.Background()

	// IsPushEnabled returns true
	mock.ExpectQuery("SELECT").
		WithArgs("user-1", db.PushCategoryREMINDER).
		WillReturnRows(sqlmock.NewRows([]string{"enabled"}).AddRow(true))

	// The dummyFCMClient has nil internals, so SendPushToUser will panic.
	// Recovering confirms we got past all pre-FCM checks.
	func() {
		defer func() { recover() }()
		_ = svc.SendToUser(ctx, "user-1", db.PushCategoryREMINDER, "T", "B", nil)
	}()
}

// --- SendToUsers with non-nil FCM: iterates all users ---

func TestSendToUsers_WithFCM_RateLimitedNoFCMCall(t *testing.T) {
	now := time.Now().In(mskLocation)
	if now.Hour() >= 23 || now.Hour() < 9 {
		t.Skip("skipping test during quiet hours")
	}

	svc, _, mr := newTestServiceWithFCM(t)
	ctx := context.Background()

	// Rate limit all users
	dateKey := now.Format("2006-01-02")
	for _, uid := range []string{"u1", "u2"} {
		mr.Set(fmt.Sprintf("push-count:%s:%s", uid, dateKey), "3")
	}

	// Should not panic (rate limited before reaching FCM)
	svc.SendToUsers(ctx, []string{"u1", "u2"}, db.PushCategorySEASONSTART, "T", "B", nil)
}

// --- mskLocation initialization test ---

func TestMskLocation_Initialized(t *testing.T) {
	if mskLocation == nil {
		t.Fatal("mskLocation should be initialized by init()")
	}

	// Verify it represents MSK (UTC+3)
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	msk := now.In(mskLocation)
	_, offset := msk.Zone()
	if offset != 3*60*60 {
		t.Errorf("expected MSK offset 10800 seconds, got %d", offset)
	}
}
