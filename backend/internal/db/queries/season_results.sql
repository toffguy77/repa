-- name: CreateSeasonResult :one
INSERT INTO season_results (id, season_id, target_id, question_id, vote_count, total_voters, percentage)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetSeasonResultsByUser :many
-- Tone comes along because the card's order depends on it: within a near-tie the warmer attribute
-- leads, since the card is the thing a person is invited to share.
SELECT sr.*, q.text as question_text, q.category as question_category, q.tone as question_tone
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
WHERE sr.season_id = $1 AND sr.target_id = $2
ORDER BY sr.percentage DESC;

-- name: GetSeasonResults :many
SELECT * FROM season_results WHERE season_id = $1
ORDER BY percentage DESC;

-- name: DeleteSeasonResultsBySeason :exec
DELETE FROM season_results WHERE season_id = $1;

-- name: GetTopResultPerQuestion :many
SELECT DISTINCT ON (sr.question_id) sr.question_id, sr.target_id, sr.vote_count, sr.percentage,
  q.text as question_text, u.username, u.avatar_emoji
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
JOIN users u ON u.id = sr.target_id
WHERE sr.season_id = $1
ORDER BY sr.question_id, sr.percentage DESC, sr.vote_count DESC;

-- name: GetTopAttributeForUser :one
SELECT sr.question_id, sr.percentage
FROM season_results sr
WHERE sr.season_id = $1 AND sr.target_id = $2
ORDER BY sr.percentage DESC
LIMIT 1;

-- name: GetMaxPercentageForUser :one
SELECT COALESCE(MAX(sr.percentage), 0)::float as max_pct
FROM season_results sr
WHERE sr.season_id = $1 AND sr.target_id = $2;

-- name: GetAllSeasonResultsWithUsers :many
-- Tone is selected for the same reason as in GetSeasonResultsByUser: this feeds the members-cards
-- list, whose cards are ordered by the same rule as a member's own.
SELECT sr.target_id, sr.question_id, sr.vote_count, sr.percentage,
  q.text as question_text, q.category as question_category, q.tone as question_tone,
  u.username, u.avatar_emoji, u.avatar_url
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
JOIN users u ON u.id = sr.target_id
WHERE sr.season_id = $1
ORDER BY sr.target_id, sr.percentage DESC;

-- name: HasQuestionBeenToppedInGroup :one
SELECT COUNT(*) FROM season_results sr
JOIN seasons s ON s.id = sr.season_id
WHERE s.group_id = $1 AND sr.question_id = $2 AND sr.season_id != $3
AND sr.percentage = (
  SELECT MAX(sr2.percentage) FROM season_results sr2
  WHERE sr2.season_id = sr.season_id AND sr2.question_id = sr.question_id
);

-- name: GetGroupChronicle :many
-- The group's shared record: for each of the last $2 revealed seasons, the standout result per
-- question. Flat rows, grouped into seasons by the service.
--
-- One query rather than GetTopResultPerQuestion per season: a group that has played for a year has
-- ~50 seasons, and the per-season loop would be the second N+1 in this codebase.
--
-- total_voters comes along because it decides the anonymity marking: a share of two voters identifies
-- them whatever the group's size was, and the group's historic membership is not recoverable — the
-- group_members table records joined_at and no departure.
--
-- The season window is a subquery so DISTINCT ON can order by (season, question, percentage) while the
-- window itself is ordered by season number.
SELECT DISTINCT ON (sr.season_id, sr.question_id)
  s.id AS season_id, s.number AS season_number, s.reveal_at,
  sr.question_id, q.text AS question_text, q.category AS question_category,
  sr.target_id, u.username, u.avatar_emoji,
  sr.percentage, sr.vote_count, sr.total_voters
FROM season_results sr
JOIN (
  -- Aliased and qualified: sqlc's analyzer otherwise reads the unqualified group_id against the
  -- outer query's tables and reports it as ambiguous.
  SELECT sn.id, sn.number, sn.reveal_at FROM seasons sn
  -- Both statuses, because CLOSED is what a REVEALED season becomes when the next one opens
  -- (see groups.createSeasonForGroup). A group therefore has at most one REVEALED season at any
  -- moment, and every older season that legitimately revealed is CLOSED — filtering on REVEALED
  -- alone would make the chronicle a one-entry list that silently looked correct.
  WHERE sn.group_id = $1 AND sn.status IN ('REVEALED', 'CLOSED')
  ORDER BY sn.number DESC
  LIMIT $2
) s ON s.id = sr.season_id
JOIN questions q ON q.id = sr.question_id
JOIN users u ON u.id = sr.target_id
ORDER BY sr.season_id, sr.question_id, sr.percentage DESC, sr.vote_count DESC;

-- name: CountRevealedSeasons :one
-- How many seasons this group has actually revealed. CLOSED counts: see GetGroupChronicle above.
SELECT COUNT(*)::bigint FROM seasons
WHERE group_id = $1 AND status IN ('REVEALED', 'CLOSED');
