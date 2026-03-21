-- name: AddSeasonQuestion :one
INSERT INTO season_questions (id, season_id, question_id, ord)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSeasonQuestions :many
SELECT q.* FROM questions q
JOIN season_questions sq ON sq.question_id = q.id
WHERE sq.season_id = $1
ORDER BY sq.ord ASC;
