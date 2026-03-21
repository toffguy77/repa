-- name: UpsertFCMToken :one
INSERT INTO fcm_tokens (id, user_id, token, platform)
VALUES ($1, $2, $3, $4)
ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id, platform = EXCLUDED.platform
RETURNING *;

-- name: GetUserFCMTokens :many
SELECT * FROM fcm_tokens WHERE user_id = $1;

-- name: DeleteFCMToken :exec
DELETE FROM fcm_tokens WHERE token = $1;
