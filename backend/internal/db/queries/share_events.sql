-- name: CreateShareEvent :one
INSERT INTO share_events (id, user_id, season_id, channel)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: CountSharesByChannel :many
-- Shares grouped by channel, for the acquisition funnel in admin stats.
SELECT channel, COUNT(*)::bigint AS count
FROM share_events
GROUP BY channel
ORDER BY count DESC;

-- name: CountSharesBySeason :one
SELECT COUNT(*)::bigint FROM share_events WHERE season_id = $1;

-- name: CountJoinsBySource :many
-- Memberships grouped by how the member arrived. Compared against CountSharesByChannel to
-- judge the loop from a shared card to a joined member.
SELECT join_source, COUNT(*)::bigint AS count
FROM group_members
GROUP BY join_source
ORDER BY count DESC;
