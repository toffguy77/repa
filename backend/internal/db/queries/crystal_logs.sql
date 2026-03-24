-- name: CreateCrystalLog :one
INSERT INTO crystal_logs (id, user_id, delta, balance, type, description, external_id)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetUserBalance :one
SELECT COALESCE(SUM(delta), 0)::int AS balance FROM crystal_logs WHERE user_id = $1;

-- name: LockUserForUpdate :one
SELECT id FROM users WHERE id = $1 FOR UPDATE;

-- name: GetUserCrystalLogs :many
SELECT * FROM crystal_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
