-- name: CreateNextSeasonVote :one
INSERT INTO next_season_votes (id, group_id, user_id, question_id, season_number)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetNextSeasonVotes :many
SELECT nsv.*, q.text as question_text, q.category as question_category
FROM next_season_votes nsv
JOIN questions q ON q.id = nsv.question_id
WHERE nsv.group_id = $1 AND nsv.season_number = $2
ORDER BY nsv.created_at DESC;

-- name: CountNextSeasonVotesByQuestion :many
SELECT question_id, COUNT(*) as vote_count
FROM next_season_votes
WHERE group_id = $1 AND season_number = $2
GROUP BY question_id
ORDER BY vote_count DESC;

-- name: HasNextSeasonVote :one
SELECT COUNT(*) FROM next_season_votes
WHERE group_id = $1 AND user_id = $2 AND season_number = $3;
