-- name: GetAllVotingSeasons :many
SELECT * FROM seasons WHERE status = 'VOTING';

-- name: GetNonVotersByseason :many
SELECT gm.user_id FROM group_members gm
WHERE gm.group_id = $1
AND gm.user_id NOT IN (
  SELECT DISTINCT voter_id FROM votes WHERE season_id = $2
);

-- name: GetVotedUsersBySeason :many
SELECT DISTINCT voter_id as user_id FROM votes WHERE season_id = $1;

-- name: GetTopVotedCategoryBySeason :one
SELECT q.category FROM votes v
JOIN questions q ON q.id = v.question_id
WHERE v.season_id = $1
GROUP BY q.category
ORDER BY COUNT(*) DESC
LIMIT 1;

-- name: GetUsersWithStreakInGroup :many
SELECT user_id FROM user_group_stats
WHERE group_id = $1 AND voting_streak >= $2;

-- name: GetAllGroups :many
SELECT * FROM groups;

-- name: GetAllGroupsWithMembers :many
SELECT g.id, g.name FROM groups g
JOIN group_members gm ON gm.group_id = g.id
GROUP BY g.id, g.name
HAVING COUNT(gm.id) >= 3;
