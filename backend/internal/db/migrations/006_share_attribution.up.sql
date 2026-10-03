-- 006_share_attribution.up.sql
-- Close the loop on the shared card: record how each member arrived, and count share actions,
-- so the path from a shared card to a joined member is measured rather than assumed.

CREATE TYPE join_source AS ENUM ('LINK', 'CODE', 'CARD', 'TELEGRAM', 'UNKNOWN');

-- Provenance is a property of the membership: written once, always read alongside it.
-- Existing rows read as UNKNOWN, which is accurate — their source was never recorded.
ALTER TABLE group_members
  ADD COLUMN join_source join_source NOT NULL DEFAULT 'UNKNOWN';

-- A share can happen many times per member per season, so this is append-only with no
-- uniqueness constraint. It deliberately holds no vote data: a table keyed by user and season
-- must not acquire anything correlatable with answers.
CREATE TABLE share_events (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  channel    TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_share_events_season_channel ON share_events(season_id, channel);
CREATE INDEX idx_group_members_join_source ON group_members(join_source);
