-- name: UpsertPushPreference :one
INSERT INTO push_preferences (id, user_id, category, enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, category) DO UPDATE SET enabled = EXCLUDED.enabled
RETURNING *;

-- name: GetUserPushPreferences :many
SELECT * FROM push_preferences WHERE user_id = $1;

-- name: IsPushEnabled :one
SELECT COALESCE(
  (SELECT enabled FROM push_preferences WHERE user_id = $1 AND category = $2),
  TRUE
) AS enabled;
