-- name: UpsertUserGroupStats :one
INSERT INTO user_group_stats (id, user_id, group_id, seasons_played, voting_streak, max_voting_streak, guess_accuracy, total_votes_cast, total_votes_received)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (user_id, group_id) DO UPDATE SET
  seasons_played = EXCLUDED.seasons_played,
  voting_streak = EXCLUDED.voting_streak,
  max_voting_streak = EXCLUDED.max_voting_streak,
  guess_accuracy = EXCLUDED.guess_accuracy,
  total_votes_cast = EXCLUDED.total_votes_cast,
  total_votes_received = EXCLUDED.total_votes_received,
  updated_at = NOW()
RETURNING *;

-- name: GetUserGroupStats :one
SELECT * FROM user_group_stats WHERE user_id = $1 AND group_id = $2;
