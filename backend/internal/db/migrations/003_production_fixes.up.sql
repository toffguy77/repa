-- 003_production_fixes.up.sql

-- Add username_changed_at column (fixes: updated_at used as username cooldown)
ALTER TABLE users ADD COLUMN username_changed_at TIMESTAMPTZ;

-- Add missing indexes for performance
CREATE INDEX IF NOT EXISTS idx_crystal_logs_user_id ON crystal_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_seasons_status ON seasons(status);
