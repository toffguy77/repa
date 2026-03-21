-- name: CreateReport :one
INSERT INTO reports (id, question_id, reporter_id, reason)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: HasUserReported :one
SELECT EXISTS(SELECT 1 FROM reports WHERE question_id = $1 AND reporter_id = $2) AS has_reported;
