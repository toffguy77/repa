-- name: CreateSeasonResult :one
INSERT INTO season_results (id, season_id, target_id, question_id, vote_count, total_voters, percentage)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetSeasonResultsByUser :many
SELECT sr.*, q.text as question_text, q.category as question_category
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
WHERE sr.season_id = $1 AND sr.target_id = $2
ORDER BY sr.percentage DESC;

-- name: GetSeasonResults :many
SELECT * FROM season_results WHERE season_id = $1
ORDER BY percentage DESC;
