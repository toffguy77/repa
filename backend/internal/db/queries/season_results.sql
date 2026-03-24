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

-- name: DeleteSeasonResultsBySeason :exec
DELETE FROM season_results WHERE season_id = $1;

-- name: GetTopResultPerQuestion :many
SELECT DISTINCT ON (sr.question_id) sr.question_id, sr.target_id, sr.vote_count, sr.percentage,
  q.text as question_text, u.username, u.avatar_emoji
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
JOIN users u ON u.id = sr.target_id
WHERE sr.season_id = $1
ORDER BY sr.question_id, sr.percentage DESC, sr.vote_count DESC;

-- name: GetTopAttributeForUser :one
SELECT sr.question_id, sr.percentage
FROM season_results sr
WHERE sr.season_id = $1 AND sr.target_id = $2
ORDER BY sr.percentage DESC
LIMIT 1;

-- name: GetMaxPercentageForUser :one
SELECT COALESCE(MAX(sr.percentage), 0)::float as max_pct
FROM season_results sr
WHERE sr.season_id = $1 AND sr.target_id = $2;

-- name: GetAllSeasonResultsWithUsers :many
SELECT sr.target_id, sr.question_id, sr.vote_count, sr.percentage,
  q.text as question_text, q.category as question_category,
  u.username, u.avatar_emoji, u.avatar_url
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
JOIN users u ON u.id = sr.target_id
WHERE sr.season_id = $1
ORDER BY sr.target_id, sr.percentage DESC;

-- name: HasQuestionBeenToppedInGroup :one
SELECT COUNT(*) FROM season_results sr
JOIN seasons s ON s.id = sr.season_id
WHERE s.group_id = $1 AND sr.question_id = $2 AND sr.season_id != $3
AND sr.percentage = (
  SELECT MAX(sr2.percentage) FROM season_results sr2
  WHERE sr2.season_id = sr.season_id AND sr2.question_id = sr.question_id
);
