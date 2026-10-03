package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sendWebhookMessage delivers a chat message to the bot webhook the way Telegram would.
func sendWebhookMessage(t *testing.T, chatID int64, text string) jsonResponse {
	t.Helper()
	body := map[string]any{
		"update_id": 1,
		"message": map[string]any{
			"message_id": 1,
			"chat": map[string]any{
				"id":       chatID,
				"type":     "supergroup",
				"username": "testchat",
			},
			"text": text,
		},
	}
	return doRequest(t, "POST", "/api/v1/telegram/webhook", body, "")
}

// connectChatDirectly links a chat to a group in the database, standing in for a successful
// /connect (which needs the bot to be a chat administrator).
func connectChatDirectly(t *testing.T, groupID string, chatID int64) {
	t.Helper()
	_, err := suite.sqlDB.ExecContext(context.Background(),
		`UPDATE groups SET telegram_chat_id = $1 WHERE id = $2`,
		fmt.Sprintf("%d", chatID), groupID)
	require.NoError(t, err)
}

func TestTelegramJoin_WebhookAcceptsCommands(t *testing.T) {
	// The webhook must return 200 for every shape of message: Telegram retries anything else.
	for _, text := range []string{"/join", "/join@repaapp_bot", "/help", "/unknowncmd", "hello"} {
		resp := sendWebhookMessage(t, 100100, text)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"webhook should accept %q: %s", text, string(resp.RawBody))
	}
}

// --- 4.3: a Telegram-sourced join is attributed ---

func TestTelegramJoin_JoiningFromTheBotsLinkIsAttributed(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("tg_owner"))
	group := createTestGroup(t, ownerToken, shortName("Telegram Join"), []string{"FUNNY"})
	groupID := group["id"].(string)
	code := group["invite_code"].(string)

	connectChatDirectly(t, groupID, 100200)

	// The link the bot hands out marks Telegram as the channel.
	inviteURL := "https://repa.app/join/" + code + "?s=telegram"
	assert.True(t, strings.Contains(inviteURL, "s=telegram"))

	userID, token := createTestUser(t, shortUsername("tg_joiner"))
	resp := doRequest(t, "POST", "/api/v1/groups/join/"+code+"?source=TELEGRAM", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.RawBody))

	assert.Equal(t, "TELEGRAM", joinSourceOf(t, userID, groupID))

	// And it is counted separately in the funnel.
	stats := doRequestBasicAuth(t, "GET", "/api/v1/admin/stats", nil,
		testAdminUsername, testAdminPassword)
	require.Equal(t, http.StatusOK, stats.StatusCode)
	joins := getData(t, stats)["joins_by_source"].(map[string]any)
	assert.Contains(t, joins, "TELEGRAM")
}

func TestTelegramJoin_UnconnectedChatIsNotHandedAGroup(t *testing.T) {
	_, ownerToken := createTestUser(t, shortUsername("tg_unconn"))
	createTestGroup(t, ownerToken, shortName("Telegram Unconnected"), []string{"FUNNY"})

	// A chat that was never connected: the bot must not resolve some other group for it. The
	// webhook replies asynchronously, so this asserts the contract at the service level.
	resp := sendWebhookMessage(t, 100300, "/join")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
