-- name: CreateAchievement :one
INSERT INTO achievements (id, user_id, group_id, season_id, achievement_type, metadata)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetUserAchievements :many
SELECT * FROM achievements
WHERE user_id = $1 AND group_id = $2
ORDER BY earned_at DESC;

-- name: GetSeasonAchievements :many
SELECT * FROM achievements
WHERE season_id = $1
ORDER BY earned_at DESC;
