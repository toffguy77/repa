-- name: CreateQuestion :one
INSERT INTO questions (id, text, category, source, group_id, author_id, status)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetSystemQuestionsByCategory :many
SELECT * FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE' AND category = $1
ORDER BY RANDOM();

-- name: GetRandomSystemQuestions :many
SELECT * FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE'
ORDER BY RANDOM()
LIMIT $1;

-- name: GetGroupCustomQuestions :many
SELECT * FROM questions
WHERE group_id = $1 AND source = 'USER' AND status = 'ACTIVE'
ORDER BY created_at DESC;

-- name: UpdateQuestionStatus :exec
UPDATE questions SET status = $2 WHERE id = $1;

-- name: CreateQuestionSeed :exec
INSERT INTO questions (id, text, category, source, status)
VALUES ($1, $2, $3, 'SYSTEM', 'ACTIVE')
ON CONFLICT DO NOTHING;
