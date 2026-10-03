-- 007_referrals.down.sql
-- Grants already written stay valid: they are ordinary crystal_logs rows.

DROP INDEX IF EXISTS idx_group_members_invited_by;
ALTER TABLE group_members DROP COLUMN IF EXISTS invited_by;
