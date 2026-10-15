-- name: CreateBlock :one
-- Idempotent: blocking twice is not an error, and the second call must not fail the request.
INSERT INTO blocks (id, blocker_id, blocked_id)
VALUES ($1, $2, $3)
-- A no-op update, so the existing row is returned rather than the insert failing.
ON CONFLICT (blocker_id, blocked_id) DO UPDATE SET created_at = blocks.created_at
RETURNING *;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2;

-- name: IsBlockedEitherWay :one
-- A block's effect is symmetric even though only one person created it.
SELECT EXISTS (
  SELECT 1 FROM blocks
  WHERE (blocker_id = $1 AND blocked_id = $2)
     OR (blocker_id = $2 AND blocked_id = $1)
);

-- name: ListBlockedUserIDs :many
-- Everyone this user cannot interact with, in either direction.
-- One scan with a CASE rather than a UNION of two scans: the UNION form also confuses sqlc's
-- analyzer about which branch an unqualified column belongs to.
-- The cast is needed: without it sqlc cannot infer the CASE expression's type.
SELECT (CASE WHEN b.blocker_id = $1 THEN b.blocked_id ELSE b.blocker_id END)::text AS user_id
FROM blocks b
WHERE b.blocker_id = $1 OR b.blocked_id = $1;

-- name: GetVotingTargets :many
-- Members who may be rated by this voter: everyone else, minus anyone blocked in either direction.
--
-- Done in SQL rather than by filtering GetGroupMembers in Go so the session's target list and the
-- vote endpoint cannot disagree — the bug this prevents is offering a target the vote then refuses.
SELECT u.id, u.username, u.avatar_emoji, u.avatar_url
FROM users u
JOIN group_members gm ON gm.user_id = u.id
WHERE gm.group_id = $1
  AND u.id <> $2
  AND NOT EXISTS (
    SELECT 1 FROM blocks b
    WHERE (b.blocker_id = $2 AND b.blocked_id = u.id)
       OR (b.blocker_id = u.id AND b.blocked_id = $2)
  )
ORDER BY gm.joined_at;

-- name: CreateGroupBan :one
INSERT INTO group_bans (id, group_id, user_id, banned_by, reason)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (group_id, user_id) DO UPDATE SET reason = EXCLUDED.reason
RETURNING *;

-- name: IsGroupBanned :one
SELECT EXISTS (SELECT 1 FROM group_bans WHERE group_id = $1 AND user_id = $2);

-- name: CreateUserReport :one
-- One report per person per reporter, matching the shape the question queue already uses.
INSERT INTO user_reports (id, reported_id, reporter_id, group_id, reason)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (reported_id, reporter_id) DO UPDATE SET reason = EXCLUDED.reason
RETURNING *;

-- name: ListUserReports :many
SELECT ur.id, ur.reported_id, ur.reporter_id, ur.group_id, ur.reason, ur.created_at,
       reported.username AS reported_username,
       reporter.username AS reporter_username
FROM user_reports ur
JOIN users reported ON reported.id = ur.reported_id
JOIN users reporter ON reporter.id = ur.reporter_id
ORDER BY ur.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUserReports :one
SELECT COUNT(*)::bigint FROM user_reports;

-- name: CountBlockedGroupMembers :one
-- How many members *of this group* the user cannot interact with, in either direction.
--
-- Scoped to the group on purpose. ListBlockedUserIDs is global, so subtracting its length from a
-- group's membership also subtracts people the user blocked in other groups entirely — which silently
-- shrank one group because of something that happened in another.
SELECT COUNT(*)::bigint
FROM group_members gm
WHERE gm.group_id = $1
  AND EXISTS (
    SELECT 1 FROM blocks b
    WHERE (b.blocker_id = $2 AND b.blocked_id = gm.user_id)
       OR (b.blocker_id = gm.user_id AND b.blocked_id = $2)
  );
