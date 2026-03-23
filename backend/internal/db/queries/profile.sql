-- name: GetTopAttributeAllTime :one
SELECT q.text as question_text, COALESCE(MAX(sr.percentage), 0)::float as percentage
FROM season_results sr
JOIN questions q ON q.id = sr.question_id
JOIN seasons s ON s.id = sr.season_id
WHERE sr.target_id = $1 AND s.group_id = $2 AND s.status = 'REVEALED'
GROUP BY q.id, q.text
ORDER BY MAX(sr.percentage) DESC
LIMIT 1;

-- name: GetUserSeasonHistory :many
SELECT h.season_id, h.season_number, h.reveal_at,
  h.question_id, h.question_text, h.question_category,
  h.percentage, h.vote_count, h.total_voters
FROM (
  SELECT DISTINCT ON (s.id) s.id as season_id, s.number as season_number, s.reveal_at,
    sr.question_id, q.text as question_text, q.category as question_category,
    sr.percentage, sr.vote_count, sr.total_voters
  FROM seasons s
  JOIN season_results sr ON sr.season_id = s.id AND sr.target_id = $1
  JOIN questions q ON q.id = sr.question_id
  WHERE s.group_id = $2 AND s.status = 'REVEALED'
  ORDER BY s.id, sr.percentage DESC, sr.vote_count DESC
) h
ORDER BY h.season_number DESC
LIMIT $3;

-- name: GetUserProfileInfo :one
SELECT id, username, avatar_emoji, avatar_url FROM users WHERE id = $1;
