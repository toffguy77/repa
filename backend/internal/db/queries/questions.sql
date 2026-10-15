-- name: CreateQuestion :one
INSERT INTO questions (id, text, category, source, group_id, author_id, status, tone)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetSystemQuestionsByCategory :many
SELECT * FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE' AND category = $1
ORDER BY RANDOM();

-- name: GetRandomSystemQuestions :many
SELECT * FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE'
ORDER BY RANDOM()
LIMIT $1;

-- name: GetRandomSystemQuestionsByCategories :many
-- $4 is the group's kind_only setting. Filtering inside the query rather than afterwards matters: the
-- LIMIT is min(10, members*2), so dropping rows after it would hand a kind-only group fewer questions
-- than an ordinary one.
SELECT * FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE'
  AND category = ANY($1::question_category[])
  AND (NOT $4::boolean OR tone <> 'EDGY')
  AND id NOT IN (
    SELECT sq.question_id FROM season_questions sq
    JOIN seasons s ON s.id = sq.season_id
    WHERE s.group_id = $2
      AND s.number > (SELECT COALESCE(MAX(number), 0) - 3 FROM seasons WHERE group_id = $2)
  )
ORDER BY RANDOM()
LIMIT $3;

-- name: CountKindQuestionsByCategories :one
-- How many non-edgy questions the given categories offer. Used to refuse a combination that would
-- hand a kind-only group an empty season: HOT and SECRETS exist to provoke, so every question in them
-- is edgy and a kind-only group restricted to them would draw nothing.
--
-- Counted against the bank rather than compared with a hardcoded list of "kind categories", so adding
-- one gentle SECRETS question is enough to make the combination legal — no code change needed.
SELECT COUNT(*)::bigint FROM questions
WHERE source = 'SYSTEM' AND status = 'ACTIVE'
  AND category = ANY($1::question_category[])
  AND tone <> 'EDGY';

-- name: GetGroupCustomQuestions :many
SELECT * FROM questions
WHERE group_id = $1 AND source = 'USER' AND status = 'ACTIVE'
ORDER BY created_at DESC;

-- name: UpdateQuestionStatus :exec
UPDATE questions SET status = $2 WHERE id = $1;

-- name: CountUserQuestionsInGroup :one
SELECT COUNT(*) FROM questions
WHERE author_id = $1 AND group_id = $2 AND source = 'USER' AND status != 'REJECTED';

-- name: GetGroupAllQuestions :many
SELECT * FROM questions
WHERE (
  (source = 'SYSTEM' AND status = 'ACTIVE' AND category = ANY($2::question_category[]))
  OR (source = 'USER' AND group_id = $1 AND status = 'ACTIVE')
)
ORDER BY source DESC, created_at DESC;

-- name: CreateQuestionSeed :exec
INSERT INTO questions (id, text, category, source, status)
VALUES ($1, $2, $3, 'SYSTEM', 'ACTIVE')
ON CONFLICT DO NOTHING;
