-- name: GetDetectorHints :many
-- Voters already revealed to this member for this season, with just enough of each to show a
-- partial identity. Deliberately selects no username: the hint DTO has no field for one.
SELECT dh.revealed_user_id, u.username, u.avatar_emoji, u.avatar_url, dh.created_at
FROM detector_hints dh
JOIN users u ON u.id = dh.revealed_user_id
WHERE dh.user_id = $1 AND dh.season_id = $2
ORDER BY dh.created_at;

-- name: CountDetectorHints :one
SELECT COUNT(*)::bigint FROM detector_hints WHERE user_id = $1 AND season_id = $2;

-- name: PickUnrevealedVoter :one
-- A voter of this season who has not yet been revealed to this member, chosen at random.
--
-- Random rather than ordered: a fixed order would let a player infer the rule, and ordering by
-- anything meaningful (join date, vote time) would leak a second fact beyond identity.
-- Postgres rejects ORDER BY random() alongside SELECT DISTINCT, so the de-duplication happens in a
-- subquery and the random pick sits outside it.
SELECT v.id, v.username, v.avatar_emoji, v.avatar_url
FROM (
  SELECT DISTINCT u.id, u.username, u.avatar_emoji, u.avatar_url
  FROM votes vo
  JOIN users u ON u.id = vo.voter_id
  WHERE vo.season_id = $1
    AND vo.voter_id <> $2
    AND NOT EXISTS (
      SELECT 1 FROM detector_hints dh
      WHERE dh.user_id = $2 AND dh.season_id = $1 AND dh.revealed_user_id = u.id
    )
) v
ORDER BY random()
LIMIT 1;

-- name: CreateDetectorHint :one
INSERT INTO detector_hints (id, user_id, season_id, revealed_user_id)
VALUES ($1, $2, $3, $4) RETURNING *;
