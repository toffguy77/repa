-- name: CreateVote :one
INSERT INTO votes (id, season_id, voter_id, target_id, question_id)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetVotesBySeasonAndVoter :many
SELECT * FROM votes WHERE season_id = $1 AND voter_id = $2;

-- name: CountUniqueVoters :one
SELECT COUNT(DISTINCT voter_id) FROM votes WHERE season_id = $1;

-- name: GetVotersBySeason :many
SELECT DISTINCT voter_id FROM votes WHERE season_id = $1;

-- name: HasVoteForQuestion :one
SELECT COUNT(*) FROM votes WHERE season_id = $1 AND voter_id = $2 AND question_id = $3;

-- name: CountCompletedVoters :one
SELECT COUNT(*) FROM (
  SELECT voter_id FROM votes
  WHERE season_id = $1
  GROUP BY voter_id
  HAVING COUNT(*) >= $2::bigint
) sub;

-- name: AggregateVotesByTarget :many
SELECT target_id, question_id, COUNT(*) as vote_count
FROM votes WHERE season_id = $1
GROUP BY target_id, question_id;

-- name: GetWinnerPerQuestion :many
SELECT DISTINCT ON (question_id) question_id, target_id, COUNT(*) as vote_count
FROM votes WHERE season_id = $1
GROUP BY question_id, target_id
ORDER BY question_id, vote_count DESC;

-- name: GetVotesByVoterInSeason :many
SELECT v.question_id, v.target_id FROM votes v
WHERE v.season_id = $1 AND v.voter_id = $2;

-- name: GetFirstVoteTimeByUser :one
SELECT COALESCE(MIN(created_at), NOW())::timestamptz as first_vote_at FROM votes
WHERE season_id = $1 AND voter_id = $2;

-- name: GetFirstCompletedVoter :one
SELECT voter_id FROM (
  SELECT voter_id, MAX(created_at) as last_vote_at
  FROM votes WHERE season_id = $1
  GROUP BY voter_id
  HAVING COUNT(*) >= $2::bigint
) sub
ORDER BY last_vote_at ASC
LIMIT 1;

-- name: CountVotesCastByUser :one
SELECT COUNT(*) FROM votes WHERE season_id = $1 AND voter_id = $2;

-- name: CountVotesReceivedByUser :one
SELECT COUNT(*) FROM votes WHERE season_id = $1 AND target_id = $2;
