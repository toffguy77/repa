-- name: CreateReaction :one
INSERT INTO reactions (id, season_id, reactor_id, target_id, emoji)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (season_id, reactor_id, target_id) DO UPDATE SET emoji = EXCLUDED.emoji
RETURNING *;

-- name: GetReactionsForUser :many
SELECT r.*, u.username as reactor_username, u.avatar_emoji as reactor_avatar_emoji
FROM reactions r
JOIN users u ON u.id = r.reactor_id
WHERE r.season_id = $1 AND r.target_id = $2
ORDER BY r.created_at DESC;
