package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrystals_GetBalance_InitialZero(t *testing.T) {
	_, token := createTestUser(t, "crystal_balance_"+t.Name())

	resp := doRequest(t, "GET", "/api/v1/crystals/balance", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, float64(0), data["balance"])
}

func TestCrystals_GetBalance_AfterGrant(t *testing.T) {
	userID, token := createTestUser(t, "crystal_granted_"+t.Name())
	addCrystals(t, userID, 25)

	resp := doRequest(t, "GET", "/api/v1/crystals/balance", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, float64(25), data["balance"])
}

func TestCrystals_GetPackages(t *testing.T) {
	_, token := createTestUser(t, "crystal_pkgs_"+t.Name())

	resp := doRequest(t, "GET", "/api/v1/crystals/packages", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	packages := data["packages"].([]any)
	assert.Len(t, packages, 4)

	// Verify package structure
	pkg := packages[0].(map[string]any)
	assert.NotEmpty(t, pkg["id"])
	assert.NotNil(t, pkg["crystals"])
	assert.NotNil(t, pkg["price_kopecks"])
}

func TestCrystals_Webhook_PaymentSucceeded(t *testing.T) {
	userID, _ := createTestUser(t, "crystal_webhook_"+t.Name())
	paymentID := uuid.New().String()

	// Store payment info in Redis (simulating InitPurchase)
	info := map[string]string{
		"user_id":    userID,
		"package_id": "starter", // 10 crystals
	}
	data, _ := json.Marshal(info)
	suite.rdb.Set(suite.ctx, "payment:"+paymentID, data, 0)

	// Send webhook
	webhookBody := map[string]any{
		"type":  "notification",
		"event": "payment.succeeded",
		"object": map[string]any{
			"id":     paymentID,
			"status": "succeeded",
		},
	}

	resp := doRequest(t, "POST", "/api/v1/crystals/purchase/webhook", webhookBody, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify balance updated
	var balance int32
	err := suite.sqlDB.QueryRowContext(suite.ctx,
		`SELECT COALESCE(SUM(delta), 0) FROM crystal_logs WHERE user_id = $1`, userID,
	).Scan(&balance)
	require.NoError(t, err)
	assert.Equal(t, int32(10), balance) // starter package = 10 crystals
}

func TestCrystals_Webhook_Idempotent(t *testing.T) {
	userID, _ := createTestUser(t, "crystal_idempotent_"+t.Name())
	paymentID := uuid.New().String()

	info := map[string]string{
		"user_id":    userID,
		"package_id": "starter",
	}
	data, _ := json.Marshal(info)
	suite.rdb.Set(suite.ctx, "payment:"+paymentID, data, 0)

	webhookBody := map[string]any{
		"type":  "notification",
		"event": "payment.succeeded",
		"object": map[string]any{
			"id":     paymentID,
			"status": "succeeded",
		},
	}

	// First webhook
	resp1 := doRequest(t, "POST", "/api/v1/crystals/purchase/webhook", webhookBody, "")
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	// Store payment info again (in case redis was cleaned)
	suite.rdb.Set(suite.ctx, "payment:"+paymentID, data, 0)

	// Second webhook — should be idempotent
	resp2 := doRequest(t, "POST", "/api/v1/crystals/purchase/webhook", webhookBody, "")
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Balance should still be 10 (not 20)
	var balance int32
	err := suite.sqlDB.QueryRowContext(suite.ctx,
		`SELECT COALESCE(SUM(delta), 0) FROM crystal_logs WHERE user_id = $1`, userID,
	).Scan(&balance)
	require.NoError(t, err)
	assert.Equal(t, int32(10), balance)
}
