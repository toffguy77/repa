-- name: CreateGroup :one
INSERT INTO groups (id, name, invite_code, admin_id, telegram_chat_username)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: GetGroupByInviteCode :one
SELECT * FROM groups WHERE invite_code = $1;

-- name: GetUserGroups :many
SELECT g.* FROM groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE gm.user_id = $1
ORDER BY gm.joined_at DESC;

-- name: CountUserGroups :one
SELECT COUNT(*) FROM group_members WHERE user_id = $1;

-- name: CountGroupMembers :one
SELECT COUNT(*) FROM group_members WHERE group_id = $1;

-- name: AddGroupMember :one
INSERT INTO group_members (id, user_id, group_id)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetGroupMembers :many
SELECT u.id, u.username, u.avatar_emoji, u.avatar_url
FROM users u
JOIN group_members gm ON gm.user_id = u.id
WHERE gm.group_id = $1
ORDER BY gm.joined_at ASC;

-- name: UpdateGroupTelegram :exec
UPDATE groups SET telegram_chat_id = $2, telegram_chat_username = $3,
  telegram_connect_code = NULL, telegram_connect_expiry = NULL WHERE id = $1;

-- name: SetGroupConnectCode :exec
UPDATE groups SET telegram_connect_code = $2, telegram_connect_expiry = $3 WHERE id = $1;

-- name: GetGroupByTelegramChatID :one
SELECT * FROM groups WHERE telegram_chat_id = $1;
