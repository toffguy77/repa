-- name: UpsertCardCache :one
INSERT INTO card_cache (id, user_id, season_id, image_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, season_id) DO UPDATE SET image_url = EXCLUDED.image_url
RETURNING *;

-- name: GetCardCache :one
SELECT * FROM card_cache WHERE user_id = $1 AND season_id = $2;
