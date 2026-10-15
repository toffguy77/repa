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

-- name: GetGroupGuessStanding :many
-- Every current member of the group with their stored guess accuracy.
--
-- A LEFT JOIN on purpose: a member who has not yet been through a reveal has no stats row, and that is
-- how "not yet ranked" is expressed. Defaulting them to zero would rank them last on a number nobody
-- measured.
--
-- The accuracy is read, never recomputed: the stored value is a rolling average over a 5-season
-- window (see internal/service/achievements), so recomputing here would show a different number than
-- the profile does.
SELECT u.id, u.username, u.avatar_emoji,
  ugs.guess_accuracy, ugs.seasons_played
FROM group_members gm
JOIN users u ON u.id = gm.user_id
LEFT JOIN user_group_stats ugs ON ugs.user_id = gm.user_id AND ugs.group_id = gm.group_id
WHERE gm.group_id = $1
ORDER BY ugs.guess_accuracy DESC NULLS LAST, u.username ASC;
