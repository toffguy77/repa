-- 006_share_attribution.down.sql

DROP INDEX IF EXISTS idx_group_members_join_source;
DROP TABLE IF EXISTS share_events;

ALTER TABLE group_members DROP COLUMN IF EXISTS join_source;

DROP TYPE IF EXISTS join_source;
