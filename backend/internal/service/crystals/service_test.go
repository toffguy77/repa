package crystals

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
)

// ---------- mockQuerier ----------

type mockQuerier struct {
	db.Querier
	balance      map[string]int32
	createdLogs  []db.CreateCrystalLogParams
	failOnCreate bool
	crystalLogs  map[string][]db.CrystalLog // keyed by userID
}

func (m *mockQuerier) GetUserBalance(_ context.Context, userID string) (int32, error) {
	b, ok := m.balance[userID]
	if !ok {
		return 0, nil
	}
	return b, nil
}

func (m *mockQuerier) CreateCrystalLog(_ context.Context, arg db.CreateCrystalLogParams) (db.CrystalLog, error) {
	if m.failOnCreate {
		return db.CrystalLog{}, sql.ErrConnDone
	}
	m.createdLogs = append(m.createdLogs, arg)
	if m.balance == nil {
		m.balance = make(map[string]int32)
	}
	m.balance[arg.UserID] = arg.Balance
	return db.CrystalLog{
		ID:     arg.ID,
		UserID: arg.UserID,
		Delta:  arg.Delta,
		Type:   arg.Type,
	}, nil
}

func (m *mockQuerier) GetUserCrystalLogs(_ context.Context, arg db.GetUserCrystalLogsParams) ([]db.CrystalLog, error) {
	if m.crystalLogs == nil {
		return nil, nil
	}
	logs, ok := m.crystalLogs[arg.UserID]
	if !ok {
		return nil, nil
	}
	return logs, nil
}

// ---------- Existing tests (kept) ----------

func TestGetBalance(t *testing.T) {
	m := &mockQuerier{balance: map[string]int32{"user1": 42}}
	svc := &Service{queries: m}
	balance, err := svc.GetBalance(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 42 {
		t.Errorf("expected balance 42, got %d", balance)
	}
}

func TestGetBalanceZero(t *testing.T) {
	m := &mockQuerier{balance: map[string]int32{}}
	svc := &Service{queries: m}
	balance, err := svc.GetBalance(context.Background(), "newuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
}

func TestGetPackages(t *testing.T) {
	svc := &Service{}
	packages := svc.GetPackages()
	if len(packages) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(packages))
	}
	if packages[0].ID != "starter" {
		t.Errorf("expected first package 'starter', got %q", packages[0].ID)
	}
	if packages[3].ID != "max" {
		t.Errorf("expected last package 'max', got %q", packages[3].ID)
	}
	if packages[0].Bonus != 0 {
		t.Errorf("starter bonus should be 0, got %d", packages[0].Bonus)
	}
	if packages[1].Bonus != 5 {
		t.Errorf("popular bonus should be 5, got %d", packages[1].Bonus)
	}
}

func TestFindPackage(t *testing.T) {
	pkg := findPackage("popular")
	if pkg == nil {
		t.Fatal("expected to find 'popular' package")
	}
	if pkg.Crystals != 30 {
		t.Errorf("expected 30 crystals, got %d", pkg.Crystals)
	}
	missing := findPackage("nonexistent")
	if missing != nil {
		t.Error("expected nil for nonexistent package")
	}
}

func TestInitPurchaseNoYukassa(t *testing.T) {
	svc := &Service{}
	_, err := svc.InitPurchase(context.Background(), "user1", "starter")
	if err != ErrPaymentsUnavailable {
		t.Errorf("expected ErrPaymentsUnavailable, got %v", err)
	}
}

func TestProcessWebhookIgnoresNonSucceeded(t *testing.T) {
	svc := &Service{}
	event := WebhookEvent{Event: "payment.waiting_for_capture"}
	err := svc.ProcessWebhook(context.Background(), event)
	if err != nil {
		t.Errorf("expected nil error for non-succeeded event, got %v", err)
	}
}

func TestWebhookEventParsing(t *testing.T) {
	paymentObj, _ := json.Marshal(map[string]any{
		"id":     "pay_123",
		"status": "succeeded",
		"amount": map[string]any{"value": "59.00", "currency": "RUB"},
	})
	event := WebhookEvent{Event: "payment.succeeded", Object: paymentObj}
	if event.Event != "payment.succeeded" {
		t.Errorf("expected event 'payment.succeeded', got %q", event.Event)
	}
	var parsed map[string]any
	if err := json.Unmarshal(event.Object, &parsed); err != nil {
		t.Fatalf("failed to unmarshal payment object: %v", err)
	}
	if parsed["id"] != "pay_123" {
		t.Errorf("expected payment id 'pay_123', got %v", parsed["id"])
	}
}

func TestVerifyResultJSON(t *testing.T) {
	balance := int32(50)
	result := VerifyResult{Status: "succeeded", NewBalance: &balance}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed["status"] != "succeeded" {
		t.Errorf("expected status 'succeeded', got %v", parsed["status"])
	}
	if parsed["new_balance"] != float64(50) {
		t.Errorf("expected new_balance 50, got %v", parsed["new_balance"])
	}
}

func TestVerifyResultJSONNoBalance(t *testing.T) {
	result := VerifyResult{Status: "pending", NewBalance: nil}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["new_balance"]; ok {
		t.Error("expected new_balance to be omitted when nil")
	}
}

func TestPackagePricing(t *testing.T) {
	for _, pkg := range Packages {
		if pkg.PriceKopecks <= 0 {
			t.Errorf("package %q has invalid price: %d", pkg.ID, pkg.PriceKopecks)
		}
		if pkg.Crystals <= 0 {
			t.Errorf("package %q has invalid crystal count: %d", pkg.ID, pkg.Crystals)
		}
		if pkg.Bonus < 0 {
			t.Errorf("package %q has negative bonus: %d", pkg.ID, pkg.Bonus)
		}
	}
	for i := 1; i < len(Packages); i++ {
		if Packages[i].PriceKopecks <= Packages[i-1].PriceKopecks {
			t.Errorf("packages not sorted by price: %s (%d) <= %s (%d)",
				Packages[i].ID, Packages[i].PriceKopecks,
				Packages[i-1].ID, Packages[i-1].PriceKopecks)
		}
	}
}

func TestIdempotencyExternalID(t *testing.T) {
	m := &mockQuerier{balance: map[string]int32{"user1": 10}, failOnCreate: true}
	_, err := m.CreateCrystalLog(context.Background(), db.CreateCrystalLogParams{
		ExternalID: sql.NullString{String: "pay_dup", Valid: true},
	})
	if err == nil {
		t.Error("expected error on duplicate create")
	}
}

func TestInitPurchasePackageNotFound(t *testing.T) {
	svc := &Service{yukassa: &lib.YukassaClient{}}
	_, err := svc.InitPurchase(context.Background(), "user1", "nonexistent_package")
	if err != ErrPackageNotFound {
		t.Errorf("expected ErrPackageNotFound, got %v", err)
	}
}

func TestVerifyPurchaseNoYukassa(t *testing.T) {
	svc := &Service{}
	_, err := svc.VerifyPurchase(context.Background(), "user1", "pay_123")
	if err != ErrPaymentsUnavailable {
		t.Errorf("expected ErrPaymentsUnavailable, got %v", err)
	}
}

func TestIsDuplicateKeyError_DuplicateKeyMessage(t *testing.T) {
	err := fmt.Errorf("pq: duplicate key value violates unique constraint")
	if !isDuplicateKeyError(err) {
		t.Error("expected isDuplicateKeyError to return true for 'duplicate key' message")
	}
}

func TestIsDuplicateKeyError_SQLState23505(t *testing.T) {
	err := fmt.Errorf("pq: ERROR: unique violation (SQLSTATE 23505)")
	if !isDuplicateKeyError(err) {
		t.Error("expected isDuplicateKeyError to return true for '23505' message")
	}
}

func TestIsDuplicateKeyError_UnrelatedError(t *testing.T) {
	err := fmt.Errorf("connection refused")
	if isDuplicateKeyError(err) {
		t.Error("expected isDuplicateKeyError to return false for unrelated error")
	}
}

func TestFindPackageAllIDs(t *testing.T) {
	expected := map[string]struct {
		crystals int
		bonus    int
		price    int
	}{
		"starter":  {10, 0, 5900},
		"popular":  {30, 5, 14900},
		"advanced": {70, 15, 29900},
		"max":      {160, 40, 59900},
	}
	for id, want := range expected {
		pkg := findPackage(id)
		if pkg == nil {
			t.Fatalf("findPackage(%q) returned nil", id)
		}
		if pkg.Crystals != want.crystals {
			t.Errorf("findPackage(%q).Crystals = %d, want %d", id, pkg.Crystals, want.crystals)
		}
		if pkg.Bonus != want.bonus {
			t.Errorf("findPackage(%q).Bonus = %d, want %d", id, pkg.Bonus, want.bonus)
		}
		if pkg.PriceKopecks != want.price {
			t.Errorf("findPackage(%q).PriceKopecks = %d, want %d", id, pkg.PriceKopecks, want.price)
		}
	}
}

func TestCrystalPackageJSONMarshal(t *testing.T) {
	pkg := CrystalPackage{ID: "starter", Crystals: 10, Bonus: 0, PriceKopecks: 5900}
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Fatalf("failed to marshal CrystalPackage: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed["id"] != "starter" {
		t.Errorf("expected id 'starter', got %v", parsed["id"])
	}
	if parsed["crystals"] != float64(10) {
		t.Errorf("expected crystals 10, got %v", parsed["crystals"])
	}
}

func TestCrystalPackageJSONUnmarshal(t *testing.T) {
	raw := `{"id":"popular","crystals":30,"bonus":5,"price_kopecks":14900}`
	var pkg CrystalPackage
	if err := json.Unmarshal([]byte(raw), &pkg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if pkg.ID != "popular" || pkg.Crystals != 30 || pkg.Bonus != 5 || pkg.PriceKopecks != 14900 {
		t.Errorf("unexpected package after unmarshal: %+v", pkg)
	}
}

// ---------- NEW: creditCrystals with sqlmock ----------

func TestCreditCrystals_Success(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	svc := &Service{sqlDB: sqlDB}

	mock.ExpectBegin()
	// GetUserBalance query
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10))
	// CreateCrystalLog insert - returns a full row
	mock.ExpectQuery("INSERT INTO crystal_logs").
		WithArgs(
			sqlmock.AnyArg(), // id (uuid)
			"user1",          // user_id
			int32(20),        // delta
			int32(30),        // balance (10 + 20)
			db.CrystalLogTypePURCHASE,
			sqlmock.AnyArg(), // description
			sqlmock.AnyArg(), // external_id
		).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "user_id", "delta", "balance", "type", "description", "external_id", "created_at"},
		).AddRow("log-1", "user1", 20, 30, db.CrystalLogTypePURCHASE, "Purchase: 20 crystals", "pay-1", time.Now()))
	mock.ExpectCommit()

	err = svc.creditCrystals(context.Background(), "user1", 20, "pay-1")
	if err != nil {
		t.Fatalf("creditCrystals: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreditCrystals_DuplicatePayment(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	svc := &Service{sqlDB: sqlDB}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10))
	// Simulate duplicate key error from Postgres
	mock.ExpectQuery("INSERT INTO crystal_logs").
		WithArgs(
			sqlmock.AnyArg(), "user1", int32(20), int32(30),
			db.CrystalLogTypePURCHASE, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint"))
	mock.ExpectRollback()

	err = svc.creditCrystals(context.Background(), "user1", 20, "pay-dup")
	if !errors.Is(err, ErrDuplicatePayment) {
		t.Errorf("expected ErrDuplicatePayment, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreditCrystals_BeginTxError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	svc := &Service{sqlDB: sqlDB}

	mock.ExpectBegin().WillReturnError(fmt.Errorf("tx begin failed"))

	err = svc.creditCrystals(context.Background(), "user1", 20, "pay-1")
	if err == nil {
		t.Fatal("expected error from BeginTx")
	}
}

func TestCreditCrystals_GetBalanceError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	svc := &Service{sqlDB: sqlDB}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("user1").
		WillReturnError(fmt.Errorf("db connection lost"))
	mock.ExpectRollback()

	err = svc.creditCrystals(context.Background(), "user1", 20, "pay-1")
	if err == nil {
		t.Fatal("expected error from GetUserBalance")
	}
}

// ---------- NEW: ProcessWebhook full flow with miniredis ----------

func TestProcessWebhook_FullSuccess(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	mq := &mockQuerier{balance: map[string]int32{"user1": 0}}
	svc := NewService(mq, sqlDB, rdb, nil)

	// Pre-populate redis with payment info
	info := paymentInfo{UserID: "user1", PackageID: "starter"}
	data, _ := json.Marshal(info)
	rdb.Set(context.Background(), paymentKeyPrefix+"pay-abc", data, paymentTTL)

	// Set up sqlmock for creditCrystals (called inside ProcessWebhook)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0))
	mock.ExpectQuery("INSERT INTO crystal_logs").
		WithArgs(
			sqlmock.AnyArg(), "user1", int32(10), int32(10),
			db.CrystalLogTypePURCHASE, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "user_id", "delta", "balance", "type", "description", "external_id", "created_at"},
		).AddRow("log-1", "user1", 10, 10, db.CrystalLogTypePURCHASE, "Purchase: 10 crystals", "pay-abc", time.Now()))
	mock.ExpectCommit()

	// Build webhook event
	paymentObj, _ := json.Marshal(lib.YukassaPayment{
		ID:     "pay-abc",
		Status: "succeeded",
		Amount: lib.YukassaAmount{Value: "59.00", Currency: "RUB"},
	})
	event := WebhookEvent{Event: "payment.succeeded", Object: paymentObj}

	err = svc.ProcessWebhook(context.Background(), event)
	if err != nil {
		t.Fatalf("ProcessWebhook: %v", err)
	}

	// Verify redis key was cleaned up
	exists := rdb.Exists(context.Background(), paymentKeyPrefix+"pay-abc").Val()
	if exists != 0 {
		t.Error("expected redis key to be deleted after successful processing")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestProcessWebhook_PaymentNotInRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}

	paymentObj, _ := json.Marshal(lib.YukassaPayment{
		ID:     "pay-unknown",
		Status: "succeeded",
	})
	event := WebhookEvent{Event: "payment.succeeded", Object: paymentObj}

	err := svc.ProcessWebhook(context.Background(), event)
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestProcessWebhook_InvalidPackageInRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	svc := &Service{rdb: rdb}

	// Store payment info with an invalid package ID
	info := paymentInfo{UserID: "user1", PackageID: "nonexistent"}
	data, _ := json.Marshal(info)
	rdb.Set(context.Background(), paymentKeyPrefix+"pay-bad", data, paymentTTL)

	paymentObj, _ := json.Marshal(lib.YukassaPayment{
		ID:     "pay-bad",
		Status: "succeeded",
	})
	event := WebhookEvent{Event: "payment.succeeded", Object: paymentObj}

	err := svc.ProcessWebhook(context.Background(), event)
	if !errors.Is(err, ErrPackageNotFound) {
		t.Errorf("expected ErrPackageNotFound, got %v", err)
	}
}

func TestProcessWebhook_InvalidPaymentObject(t *testing.T) {
	svc := &Service{}
	event := WebhookEvent{
		Event:  "payment.succeeded",
		Object: json.RawMessage(`{invalid json`),
	}
	err := svc.ProcessWebhook(context.Background(), event)
	if err == nil {
		t.Error("expected error for invalid payment object JSON")
	}
}

func TestProcessWebhook_DuplicatePaymentIsIdempotent(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer sqlDB.Close()

	mq := &mockQuerier{balance: map[string]int32{"user1": 10}}
	svc := NewService(mq, sqlDB, rdb, nil)

	// Store valid payment info
	info := paymentInfo{UserID: "user1", PackageID: "starter"}
	data, _ := json.Marshal(info)
	rdb.Set(context.Background(), paymentKeyPrefix+"pay-dup", data, paymentTTL)

	// creditCrystals will hit duplicate key
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10))
	mock.ExpectQuery("INSERT INTO crystal_logs").
		WithArgs(
			sqlmock.AnyArg(), "user1", int32(10), int32(20),
			db.CrystalLogTypePURCHASE, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint"))
	mock.ExpectRollback()

	paymentObj, _ := json.Marshal(lib.YukassaPayment{
		ID:     "pay-dup",
		Status: "succeeded",
	})
	event := WebhookEvent{Event: "payment.succeeded", Object: paymentObj}

	err = svc.ProcessWebhook(context.Background(), event)
	if err != nil {
		t.Fatalf("expected nil for duplicate payment (idempotent), got %v", err)
	}
}

func TestProcessWebhook_NonSucceededEvents(t *testing.T) {
	events := []string{
		"payment.waiting_for_capture",
		"payment.canceled",
		"refund.succeeded",
		"",
	}
	svc := &Service{}
	for _, evt := range events {
		err := svc.ProcessWebhook(context.Background(), WebhookEvent{Event: evt})
		if err != nil {
			t.Errorf("event %q should be ignored, got error: %v", evt, err)
		}
	}
}

// ---------- NEW: VerifyPurchase paths ----------

func TestVerifyPurchase_RedisHasInfo_WrongUser(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	mq := &mockQuerier{}
	yukassa := lib.NewYukassaClient("shop", "secret", "https://return.example.com")
	svc := &Service{queries: mq, rdb: rdb, yukassa: yukassa}

	// Store payment info for user1
	info := paymentInfo{UserID: "user1", PackageID: "starter"}
	data, _ := json.Marshal(info)
	rdb.Set(context.Background(), paymentKeyPrefix+"pay-1", data, paymentTTL)

	// user2 tries to verify
	_, err := svc.VerifyPurchase(context.Background(), "user2", "pay-1")
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound when wrong user, got %v", err)
	}
}

func TestVerifyPurchase_RedisExpired_FoundInLogs(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	mq := &mockQuerier{
		balance: map[string]int32{"user1": 50},
		crystalLogs: map[string][]db.CrystalLog{
			"user1": {
				{
					ID:         "log-1",
					UserID:     "user1",
					ExternalID: sql.NullString{String: "pay-old", Valid: true},
				},
			},
		},
	}
	// We need a yukassa client that can respond. Since YukassaClient hits real HTTP,
	// we test up to the point where it would call yukassa. We'll verify the redis-expired
	// path reaches the log check.
	yukassa := lib.NewYukassaClient("shop", "secret", "https://return.example.com")
	svc := &Service{queries: mq, rdb: rdb, yukassa: yukassa}

	// No redis key exists. The code falls through to crystal_logs check.
	// The log contains "pay-old", so ownership is verified.
	// Then it calls yukassa.GetPayment which will fail (no real server).
	_, err := svc.VerifyPurchase(context.Background(), "user1", "pay-old")
	// We expect a yukassa HTTP error, not ErrPaymentNotFound
	if errors.Is(err, ErrPaymentNotFound) {
		t.Error("should have found payment in crystal_logs, got ErrPaymentNotFound")
	}
	if err == nil {
		t.Error("expected error from yukassa HTTP call")
	}
}

func TestVerifyPurchase_RedisExpired_NotFoundInLogs(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	mq := &mockQuerier{
		crystalLogs: map[string][]db.CrystalLog{
			"user1": {}, // no logs
		},
	}
	yukassa := lib.NewYukassaClient("shop", "secret", "https://return.example.com")
	svc := &Service{queries: mq, rdb: rdb, yukassa: yukassa}

	_, err := svc.VerifyPurchase(context.Background(), "user1", "pay-missing")
	if !errors.Is(err, ErrPaymentNotFound) {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

// ---------- NEW: NewService constructor ----------

func TestNewService(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	mq := &mockQuerier{}
	svc := NewService(mq, nil, rdb, nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.queries != mq {
		t.Error("queries not set correctly")
	}
	if svc.rdb != rdb {
		t.Error("rdb not set correctly")
	}
}

// ---------- NEW: InitPurchase with valid package but nil yukassa fields ----------

func TestInitPurchase_ValidPackageButYukassaHTTPFails(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Use a YukassaClient that will fail on HTTP (no real server)
	yukassa := lib.NewYukassaClient("shop-test", "secret-test", "https://return.example.com")
	svc := &Service{rdb: rdb, yukassa: yukassa}

	_, err := svc.InitPurchase(context.Background(), "user1", "starter")
	if err == nil {
		t.Error("expected error from yukassa HTTP call")
	}
	// Should not be ErrPackageNotFound or ErrPaymentsUnavailable
	if errors.Is(err, ErrPackageNotFound) {
		t.Error("should not be ErrPackageNotFound for valid package")
	}
	if errors.Is(err, ErrPaymentsUnavailable) {
		t.Error("should not be ErrPaymentsUnavailable when yukassa is set")
	}
}

// ---------- NEW: WebhookEvent JSON marshal/unmarshal ----------

func TestWebhookEventJSONRoundTrip(t *testing.T) {
	original := WebhookEvent{
		Type:   "notification",
		Event:  "payment.succeeded",
		Object: json.RawMessage(`{"id":"p1","status":"succeeded"}`),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded WebhookEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Event != original.Event {
		t.Errorf("Event = %q, want %q", decoded.Event, original.Event)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}
}

// ---------- NEW: Error sentinel values ----------

func TestErrorSentinels(t *testing.T) {
	errs := []error{
		ErrPackageNotFound,
		ErrInsufficientFunds,
		ErrPaymentNotFound,
		ErrDuplicatePayment,
		ErrPaymentsUnavailable,
	}
	seen := make(map[string]bool)
	for _, e := range errs {
		msg := e.Error()
		if msg == "" {
			t.Error("error sentinel has empty message")
		}
		if seen[msg] {
			t.Errorf("duplicate error message: %q", msg)
		}
		seen[msg] = true
	}
}

// ---------- NEW: paymentInfo JSON ----------

func TestPaymentInfoJSON(t *testing.T) {
	info := paymentInfo{UserID: "u1", PackageID: "popular"}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded paymentInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.UserID != "u1" || decoded.PackageID != "popular" {
		t.Errorf("unexpected: %+v", decoded)
	}
}

// ---------- NEW: isDuplicateKeyError with pgErr interface ----------

type mockPGError struct {
	state string
}

func (e *mockPGError) Error() string   { return "pg error" }
func (e *mockPGError) SQLState() string { return e.state }

func TestIsDuplicateKeyError_PGInterface_23505(t *testing.T) {
	err := &mockPGError{state: "23505"}
	if !isDuplicateKeyError(err) {
		t.Error("expected true for pg error with SQLState 23505")
	}
}

func TestIsDuplicateKeyError_PGInterface_Other(t *testing.T) {
	err := &mockPGError{state: "42601"}
	if isDuplicateKeyError(err) {
		t.Error("expected false for pg error with non-23505 SQLState")
	}
}

func TestIsDuplicateKeyError_WrappedPGError(t *testing.T) {
	pgErr := &mockPGError{state: "23505"}
	wrapped := fmt.Errorf("create crystal log: %w", pgErr)
	if !isDuplicateKeyError(wrapped) {
		t.Error("expected true for wrapped pg error with SQLState 23505")
	}
}
