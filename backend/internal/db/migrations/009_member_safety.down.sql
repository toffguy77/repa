-- 009_member_safety.down.sql
-- Target filtering falls back to "everyone except me", which is the pre-change behaviour.

DROP INDEX IF EXISTS idx_user_reports_created;
DROP TABLE IF EXISTS user_reports;

DROP INDEX IF EXISTS idx_group_bans_group_user;
DROP TABLE IF EXISTS group_bans;

DROP INDEX IF EXISTS idx_blocks_blocked;
DROP INDEX IF EXISTS idx_blocks_blocker;
DROP TABLE IF EXISTS blocks;
