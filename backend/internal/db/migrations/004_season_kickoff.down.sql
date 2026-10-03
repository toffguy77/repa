-- 004_season_kickoff.down.sql

DROP INDEX IF EXISTS idx_seasons_status_reveal_at;

ALTER TABLE seasons DROP COLUMN IF EXISTS postpone_count;
ALTER TABLE seasons DROP COLUMN IF EXISTS kind;

DROP TYPE IF EXISTS season_kind;
