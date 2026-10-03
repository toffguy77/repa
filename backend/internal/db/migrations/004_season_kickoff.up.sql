-- 004_season_kickoff.up.sql
-- Kickoff seasons: a new group's first season opens immediately and reveals
-- about an hour after the group first becomes eligible, instead of waiting for
-- a calendar Friday. Postponement replaces the empty forced reveal.

CREATE TYPE season_kind AS ENUM ('KICKOFF', 'WEEKLY');

ALTER TABLE seasons ADD COLUMN kind season_kind NOT NULL DEFAULT 'WEEKLY';
ALTER TABLE seasons ADD COLUMN postpone_count integer NOT NULL DEFAULT 0;

-- reveal-checker selects on (status, reveal_at) every minute; kickoff scheduling
-- makes reveal_at mutable, so keep that lookup indexed.
CREATE INDEX IF NOT EXISTS idx_seasons_status_reveal_at ON seasons(status, reveal_at);
