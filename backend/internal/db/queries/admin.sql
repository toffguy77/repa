-- name: ListReports :many
SELECT r.id, r.question_id, r.reporter_id, r.reason, r.created_at,
       q.text AS question_text, q.category AS question_category, q.status AS question_status,
       u.username AS reporter_username
FROM reports r
JOIN questions q ON q.id = r.question_id
JOIN users u ON u.id = r.reporter_id
ORDER BY r.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountReports :one
SELECT count(*) FROM reports;

-- name: GetReportByID :one
SELECT r.id, r.question_id, r.reporter_id, r.reason, r.created_at,
       q.text AS question_text, q.status AS question_status
FROM reports r
JOIN questions q ON q.id = r.question_id
WHERE r.id = $1;

-- name: CountActiveUsers7Days :one
SELECT count(DISTINCT user_id) FROM group_members
WHERE joined_at > NOW() - INTERVAL '7 days';

-- name: CountActiveUsers30Days :one
SELECT count(DISTINCT user_id) FROM group_members
WHERE joined_at > NOW() - INTERVAL '30 days';

-- name: CountGroups :one
SELECT count(*) FROM groups;

-- name: SumRevenue7Days :one
SELECT COALESCE(SUM(delta), 0)::bigint FROM crystal_logs
WHERE type = 'PURCHASE' AND created_at > NOW() - INTERVAL '7 days';
