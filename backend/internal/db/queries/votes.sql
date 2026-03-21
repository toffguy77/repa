-- name: CreateVote :one
INSERT INTO votes (id, season_id, voter_id, target_id, question_id)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetVotesBySeasonAndVoter :many
SELECT * FROM votes WHERE season_id = $1 AND voter_id = $2;

-- name: CountUniqueVoters :one
SELECT COUNT(DISTINCT voter_id) FROM votes WHERE season_id = $1;

-- name: GetVotersBySeason :many
SELECT DISTINCT voter_id FROM votes WHERE season_id = $1;

-- name: AggregateVotesByTarget :many
SELECT target_id, question_id, COUNT(*) as vote_count
FROM votes WHERE season_id = $1
GROUP BY target_id, question_id;
