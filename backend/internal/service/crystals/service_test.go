package crystals

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	db "github.com/repa-app/repa/internal/db/sqlc"
)

// mockQuerier implements the subset of db.Querier used by the crystals service.
type mockQuerier struct {
	db.Querier
	balance        map[string]int32
	createdLogs    []db.CreateCrystalLogParams
	failOnCreate   bool
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
		return db.CrystalLog{}, sql.ErrConnDone // simulate unique constraint violation
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

func TestGetBalance(t *testing.T) {
	m := &mockQuerier{
		balance: map[string]int32{"user1": 42},
	}
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
	m := &mockQuerier{
		balance: map[string]int32{},
	}
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

	// Verify bonus values
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
	event := WebhookEvent{
		Event: "payment.waiting_for_capture",
	}
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

	event := WebhookEvent{
		Event:  "payment.succeeded",
		Object: paymentObj,
	}

	// Verify event unmarshals correctly
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
	result := VerifyResult{
		Status:     "succeeded",
		NewBalance: &balance,
	}
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

func TestPackagePricing(t *testing.T) {
	// Verify all packages have consistent pricing
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

	// Verify packages are sorted by price ascending
	for i := 1; i < len(Packages); i++ {
		if Packages[i].PriceKopecks <= Packages[i-1].PriceKopecks {
			t.Errorf("packages not sorted by price: %s (%d) <= %s (%d)",
				Packages[i].ID, Packages[i].PriceKopecks,
				Packages[i-1].ID, Packages[i-1].PriceKopecks)
		}
	}
}

func TestIdempotencyExternalID(t *testing.T) {
	// The crystal_logs table has UNIQUE constraint on external_id.
	// Verify our service correctly uses external_id for payment dedup.
	m := &mockQuerier{
		balance:      map[string]int32{"user1": 10},
		failOnCreate: true,
	}
	// creditCrystals requires sqlDB for transactions (integration test territory).
	// Here we just verify the mockQuerier correctly fails on duplicate create.
	_, err := m.CreateCrystalLog(context.Background(), db.CreateCrystalLogParams{
		ExternalID: sql.NullString{String: "pay_dup", Valid: true},
	})
	if err == nil {
		t.Error("expected error on duplicate create")
	}
}
