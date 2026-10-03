-- name: CreateSeason :one
INSERT INTO seasons (id, group_id, number, starts_at, reveal_at, ends_at, kind)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

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
-- CLOSED as well as REVEALED. This runs at reveal time, when the current season is the REVEALED one
-- and every earlier season has already been closed by createSeasonForGroup — so filtering on REVEALED
-- alone returned no rows, and the trend line it feeds never rendered.
SELECT * FROM seasons
WHERE group_id = $1 AND status IN ('REVEALED', 'CLOSED') AND id != $2
ORDER BY number DESC LIMIT 1;

-- name: GetRevealedSeasonsForGroup :many
-- Deliberately REVEALED only: this is the closing step itself, finding the seasons to mark CLOSED.
SELECT * FROM seasons
WHERE group_id = $1 AND status = 'REVEALED'
ORDER BY number DESC;

-- name: GetLastNRevealedSeasons :many
-- CLOSED as well as REVEALED, for the same reason as GetPreviousRevealedSeason: the multi-season
-- achievements that read this found no prior seasons and could never fire.
SELECT * FROM seasons
WHERE group_id = $1 AND status IN ('REVEALED', 'CLOSED')
ORDER BY number DESC
LIMIT $2;

-- name: ScheduleKickoffReveal :one
-- Brings a kickoff season's reveal forward. The reveal_at > @reveal_at guard makes
-- this idempotent and keeps later voters from pushing the time back.
UPDATE seasons
SET reveal_at = @reveal_at, ends_at = GREATEST(ends_at, @ends_at)
WHERE id = @id
  AND kind = 'KICKOFF'
  AND status = 'VOTING'
  AND reveal_at > @reveal_at
RETURNING *;

-- name: PostponeSeason :one
UPDATE seasons
SET reveal_at = @reveal_at, ends_at = @ends_at, postpone_count = postpone_count + 1
WHERE id = @id AND status = 'VOTING'
RETURNING *;

-- name: GetGroupsNeedingSeasonMaintenance :many
-- Safety net for the hourly cron: eligible groups with no open season whose most
-- recent season finished more than an hour ago (so it never races the Sunday cron
-- or the post-kickoff follow-up).
SELECT g.* FROM groups g
WHERE (SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.id) >= 3
  AND NOT EXISTS (SELECT 1 FROM seasons s WHERE s.group_id = g.id AND s.status = 'VOTING')
  AND COALESCE(
        (SELECT MAX(s.ends_at) FROM seasons s WHERE s.group_id = g.id),
        TIMESTAMPTZ '-infinity'
      ) < NOW() - INTERVAL '1 hour';
