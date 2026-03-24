package crystals

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	db "github.com/repa-app/repa/internal/db/sqlc"
	appmw "github.com/repa-app/repa/internal/middleware"
	crystalssvc "github.com/repa-app/repa/internal/service/crystals"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{}
	return e
}

type testValidator struct{}

func (tv *testValidator) Validate(i any) error { return nil }

func setUser(c echo.Context, userID, username string) {
	c.Set("user", &appmw.JWTClaims{UserID: userID, Username: username})
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestGetPackages(t *testing.T) {
	// GetPackages only calls svc.GetPackages() which returns static data.
	// The service can be constructed with nil dependencies for this method.
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/packages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetPackages(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body struct {
		Data struct {
			Packages []crystalssvc.CrystalPackage `json:"packages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(body.Data.Packages) == 0 {
		t.Fatal("expected at least one package")
	}

	// Verify the known packages are present
	expectedIDs := map[string]bool{"starter": false, "popular": false, "advanced": false, "max": false}
	for _, p := range body.Data.Packages {
		expectedIDs[p.ID] = true
	}
	for id, found := range expectedIDs {
		if !found {
			t.Errorf("expected package %q not found in response", id)
		}
	}
}

func TestGetPackages_ResponseFormat(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/packages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = h.GetPackages(c)

	// Verify the response has the expected envelope structure
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := raw["data"]; !ok {
		t.Error("response missing 'data' key")
	}
	data, ok := raw["data"].(map[string]any)
	if !ok {
		t.Fatal("'data' is not an object")
	}
	if _, ok := data["packages"]; !ok {
		t.Error("data missing 'packages' key")
	}
}

func TestInitPurchase_BadJSON(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	// Send invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/purchase", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.InitPurchase(c)
	if err != nil {
		t.Fatalf("expected handler to write response, not return error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if body.Error.Code != "VALIDATION" {
		t.Errorf("expected error code VALIDATION, got %s", body.Error.Code)
	}
}

func TestInitPurchase_EmptyBody(t *testing.T) {
	// With an empty body, Echo's Bind succeeds with zero-value struct.
	// Our test validator is a no-op, so validation passes.
	// The service is called with empty package_id and nil yukassa client,
	// which returns ErrPaymentsUnavailable (503).
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/purchase", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.InitPurchase(c)
	if err != nil {
		t.Fatalf("expected handler to write response, not return error: %v", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d for empty body with nil yukassa, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if body.Error.Code != "PAYMENTS_UNAVAILABLE" {
		t.Errorf("expected error code PAYMENTS_UNAVAILABLE, got %s", body.Error.Code)
	}
}

func TestWebhook_BadJSON(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/webhook", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	if err != nil {
		t.Fatalf("expected handler to write response, not return error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Webhook returns NoContent on bad JSON
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for webhook bad JSON, got %s", rec.Body.String())
	}
}

func TestWebhook_EmptyBody(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/webhook", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	if err != nil {
		t.Fatalf("expected handler to write response, not return error: %v", err)
	}

	// Empty body should bind to zero-value struct (no error from Echo's Bind for empty body)
	// The service will process it — event type won't match "payment.succeeded" so it returns nil
	// Handler returns 200 OK
	if rec.Code != http.StatusOK && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 200 or 400, got %d", rec.Code)
	}
}

// --- Mock Querier for crystals service ---

type mockCrystalsQuerier struct {
	db.Querier // embed to satisfy interface
	getUserBalanceFn func(ctx context.Context, userID string) (int32, error)
}

func (m *mockCrystalsQuerier) GetUserBalance(ctx context.Context, userID string) (int32, error) {
	if m.getUserBalanceFn != nil {
		return m.getUserBalanceFn(ctx, userID)
	}
	return 0, nil
}

// --- GetBalance tests ---

func TestGetBalance_Success(t *testing.T) {
	mock := &mockCrystalsQuerier{
		getUserBalanceFn: func(ctx context.Context, userID string) (int32, error) {
			return 42, nil
		},
	}
	svc := crystalssvc.NewService(mock, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/balance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.GetBalance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	data := resp["data"]
	if data["balance"] != float64(42) {
		t.Errorf("expected balance 42, got %v", data["balance"])
	}
}

func TestGetBalance_ZeroBalance(t *testing.T) {
	mock := &mockCrystalsQuerier{
		getUserBalanceFn: func(ctx context.Context, userID string) (int32, error) {
			return 0, nil
		},
	}
	svc := crystalssvc.NewService(mock, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/balance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-new", "newuser")

	err := h.GetBalance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["data"]["balance"] != float64(0) {
		t.Errorf("expected balance 0, got %v", resp["data"]["balance"])
	}
}

func TestGetBalance_DBError(t *testing.T) {
	mock := &mockCrystalsQuerier{
		getUserBalanceFn: func(ctx context.Context, userID string) (int32, error) {
			return 0, errors.New("db connection failed")
		},
	}
	svc := crystalssvc.NewService(mock, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/balance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.GetBalance(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "INTERNAL" {
		t.Errorf("expected error code INTERNAL, got %s", resp["error"]["code"])
	}
}

// --- InitPurchase tests ---

func TestInitPurchase_PackageNotFound(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	body := `{"package_id":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/purchase",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.InitPurchase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With nil yukassa, the service returns ErrPaymentsUnavailable before checking the package.
	// So we expect 503 PAYMENTS_UNAVAILABLE (yukassa is nil).
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestInitPurchase_ValidPackageButNoPayments(t *testing.T) {
	// Valid package ID but yukassa is nil => ErrPaymentsUnavailable
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	body := `{"package_id":"starter"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/purchase",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setUser(c, "user-123", "testuser")

	err := h.InitPurchase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "PAYMENTS_UNAVAILABLE" {
		t.Errorf("expected PAYMENTS_UNAVAILABLE, got %s", resp["error"]["code"])
	}
}

// --- VerifyPurchase tests ---

func TestVerifyPurchase_PaymentsUnavailable(t *testing.T) {
	// yukassa is nil => ErrPaymentsUnavailable
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crystals/purchase/p1/verify", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("paymentId")
	c.SetParamValues("p1")
	setUser(c, "user-123", "testuser")

	err := h.VerifyPurchase(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp map[string]map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"]["code"] != "PAYMENTS_UNAVAILABLE" {
		t.Errorf("expected PAYMENTS_UNAVAILABLE, got %s", resp["error"]["code"])
	}
}

// --- Webhook tests ---

func TestWebhook_NonSucceededEvent(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	// Send a valid JSON event that is not "payment.succeeded"
	body := `{"type":"notification","event":"payment.canceled","object":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/webhook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Non-succeeded events are silently acknowledged with 200
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Should be NoContent
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %s", rec.Body.String())
	}
}

func TestWebhook_WaitingForCaptureEvent(t *testing.T) {
	svc := crystalssvc.NewService(nil, nil, nil, nil)
	h := NewHandler(svc)

	e := setupEcho()
	body := `{"type":"notification","event":"payment.waiting_for_capture","object":{"id":"pay-1","status":"waiting_for_capture"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crystals/webhook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Webhook(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
