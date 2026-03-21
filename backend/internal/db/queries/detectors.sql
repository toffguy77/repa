-- name: CreateDetector :one
INSERT INTO detectors (id, user_id, season_id, group_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetDetector :one
SELECT * FROM detectors WHERE user_id = $1 AND season_id = $2;

-- name: HasDetector :one
SELECT EXISTS(SELECT 1 FROM detectors WHERE user_id = $1 AND season_id = $2) AS has_detector;
