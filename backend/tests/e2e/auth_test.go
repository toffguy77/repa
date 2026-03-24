package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_OTPSend_DevMode(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/otp/send", map[string]any{
		"phone": "+79001234567",
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, true, data["sent"])
	// In dev mode, OTP code is returned in response
	assert.NotEmpty(t, data["code"], "dev mode should return OTP code")
}

func TestAuth_OTPSend_InvalidPhone(t *testing.T) {
	resp := doRequest(t, "POST", "/api/v1/auth/otp/send", map[string]any{
		"phone": "invalid",
	}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuth_OTPVerify_Success(t *testing.T) {
	phone := "+79009876543"

	// Send OTP
	sendResp := doRequest(t, "POST", "/api/v1/auth/otp/send", map[string]any{
		"phone": phone,
	}, "")
	require.Equal(t, http.StatusOK, sendResp.StatusCode)
	sendData := getData(t, sendResp)
	code := sendData["code"].(string)

	// Verify OTP
	resp := doRequest(t, "POST", "/api/v1/auth/otp/verify", map[string]any{
		"phone": phone,
		"code":  code,
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.NotEmpty(t, data["token"], "should return JWT token")
	user := data["user"].(map[string]any)
	assert.NotEmpty(t, user["id"])
	assert.NotEmpty(t, user["username"])
}

func TestAuth_OTPVerify_WrongCode(t *testing.T) {
	phone := "+79001112233"

	// Send OTP
	doRequest(t, "POST", "/api/v1/auth/otp/send", map[string]any{"phone": phone}, "")

	// Verify with wrong code
	resp := doRequest(t, "POST", "/api/v1/auth/otp/verify", map[string]any{
		"phone": phone,
		"code":  "000000",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	errObj := getError(t, resp)
	assert.Equal(t, "INVALID_OTP", errObj["code"])
}

func TestAuth_GetMe(t *testing.T) {
	userID, token := createTestUser(t, "getme_user")

	resp := doRequest(t, "GET", "/api/v1/auth/me", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, userID, data["id"])
	assert.Equal(t, "getme_user", data["username"])
}

func TestAuth_GetMe_NoToken(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/auth/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuth_UpdateProfile(t *testing.T) {
	_, token := createTestUser(t, "profile_update_user")

	emoji := "🎮"
	birthYear := 2002
	resp := doRequest(t, "PATCH", "/api/v1/auth/profile", map[string]any{
		"avatar_emoji": emoji,
		"birth_year":   birthYear,
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, emoji, data["avatar_emoji"])
	by := int(data["birth_year"].(float64))
	assert.Equal(t, birthYear, by)
}

func TestAuth_UsernameCheck_Available(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/auth/username-check?username=unique_test_user_xyz", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, true, data["available"])
}

func TestAuth_UsernameCheck_Taken(t *testing.T) {
	createTestUser(t, "taken_username")

	resp := doRequest(t, "GET", "/api/v1/auth/username-check?username=taken_username", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, false, data["available"])
}

func TestAuth_UpdatePushPreferences(t *testing.T) {
	_, token := createTestUser(t, "push_pref_user")

	resp := doRequest(t, "PATCH", "/api/v1/push/preferences", map[string]any{
		"category": "REVEAL",
		"enabled":  false,
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, "REVEAL", data["category"])
	assert.Equal(t, false, data["enabled"])
}

func TestAuth_DeleteAccount(t *testing.T) {
	userID, token := createTestUser(t, "delete_me_user")

	resp := doRequest(t, "DELETE", "/api/v1/auth/account", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, true, data["deleted"])

	// Verify user is gone
	var count int
	err := suite.sqlDB.QueryRowContext(suite.ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestAuth_AppVersion(t *testing.T) {
	resp := doRequest(t, "GET", "/api/v1/app/version", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := getData(t, resp)
	assert.Equal(t, "1.0.0", data["min_version"])
	assert.Equal(t, "1.2.0", data["latest_version"])
}
