package lib

import (
	"context"
	"encoding/json"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

type FCMClient struct {
	client  *messaging.Client
	queries *db.Queries
}

type serviceAccountKey struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

func NewFCMClient(ctx context.Context, projectID, privateKey, clientEmail string, queries *db.Queries) (*FCMClient, error) {
	creds, err := json.Marshal(serviceAccountKey{
		Type:        "service_account",
		ProjectID:   projectID,
		PrivateKey:  strings.ReplaceAll(privateKey, `\n`, "\n"),
		ClientEmail: clientEmail,
	})
	if err != nil {
		return nil, err
	}

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(creds)) //nolint:staticcheck // WithCredentialsJSON is the standard way to pass service account JSON
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &FCMClient{client: client, queries: queries}, nil
}

// SendPush sends a multicast push to the given tokens.
// Note: FCM SendEachForMulticast supports max 500 tokens per call.
// Currently safe because this is called per-user (few tokens each).
func (f *FCMClient) SendPush(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	resp, err := f.client.SendEachForMulticast(ctx, msg)
	if err != nil {
		return err
	}

	for i, r := range resp.Responses {
		if r.Error == nil {
			continue
		}
		if messaging.IsUnregistered(r.Error) ||
			messaging.IsInvalidArgument(r.Error) {
			token := tokens[i]
			if delErr := f.queries.DeleteFCMToken(ctx, token); delErr != nil {
				log.Warn().Err(delErr).Str("token", token[:min(len(token), 20)]+"...").Msg("failed to delete invalid FCM token")
			} else {
				log.Info().Str("token", token[:min(len(token), 20)]+"...").Msg("deleted invalid FCM token")
			}
		}
	}

	return nil
}

func (f *FCMClient) SendPushToUser(ctx context.Context, userID, title, body string, data map[string]string) error {
	fcmTokens, err := f.queries.GetUserFCMTokens(ctx, userID)
	if err != nil {
		return err
	}

	tokens := make([]string, len(fcmTokens))
	for i, t := range fcmTokens {
		tokens[i] = t.Token
	}

	return f.SendPush(ctx, tokens, title, body, data)
}

func (f *FCMClient) SendPushToUsers(ctx context.Context, userIDs []string, title, body string, data map[string]string) error {
	for _, userID := range userIDs {
		if err := f.SendPushToUser(ctx, userID, title, body, data); err != nil {
			log.Warn().Err(err).Str("user_id", userID).Msg("failed to send push to user")
		}
	}
	return nil
}
