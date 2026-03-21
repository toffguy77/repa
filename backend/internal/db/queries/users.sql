-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: GetUserByAppleID :one
SELECT * FROM users WHERE apple_id = $1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (id, phone, apple_id, google_id, username, avatar_emoji, birth_year)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users SET username = $2, avatar_emoji = $3, avatar_url = $4,
  birth_year = $5, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: UpdateUserAvatarURL :one
UPDATE users SET avatar_url = $2, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
