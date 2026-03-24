-- 003_production_fixes.down.sql

DROP INDEX IF EXISTS idx_seasons_status;
DROP INDEX IF EXISTS idx_crystal_logs_user_id;
ALTER TABLE users DROP COLUMN IF EXISTS username_changed_at;
