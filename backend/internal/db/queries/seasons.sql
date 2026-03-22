-- name: CreateSeason :one
INSERT INTO seasons (id, group_id, number, starts_at, reveal_at, ends_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetActiveSeasonByGroup :one
SELECT * FROM seasons WHERE group_id = $1 AND status = 'VOTING' LIMIT 1;

-- name: GetSeasonByID :one
SELECT * FROM seasons WHERE id = $1;

-- name: UpdateSeasonStatus :exec
UPDATE seasons SET status = $2 WHERE id = $1;

-- name: GetSeasonsForReveal :many
SELECT * FROM seasons WHERE status = 'VOTING' AND reveal_at <= NOW();

-- name: GetGroupsNeedingNewSeason :many
SELECT DISTINCT g.* FROM groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE NOT EXISTS (SELECT 1 FROM seasons s WHERE s.group_id = g.id AND s.status = 'VOTING')
GROUP BY g.id HAVING COUNT(gm.id) >= 3;

-- name: GetLastSeasonNumber :one
SELECT COALESCE(MAX(number), 0)::int FROM seasons WHERE group_id = $1;

-- name: CountSeasonVoters :one
SELECT COUNT(DISTINCT voter_id) FROM votes WHERE season_id = $1;

-- name: HasUserVotedInSeason :one
SELECT COUNT(*) FROM votes WHERE season_id = $1 AND voter_id = $2;

-- name: GetPreviousRevealedSeason :one
SELECT * FROM seasons
WHERE group_id = $1 AND status = 'REVEALED' AND id != $2
ORDER BY number DESC LIMIT 1;

-- name: GetRevealedSeasonsForGroup :many
SELECT * FROM seasons
WHERE group_id = $1 AND status = 'REVEALED'
ORDER BY number DESC;

-- name: GetLastNRevealedSeasons :many
SELECT * FROM seasons
WHERE group_id = $1 AND status = 'REVEALED'
ORDER BY number DESC
LIMIT $2;
