-- 008_detector_hints.up.sql
-- The detector becomes a ladder: a free count, cheap partial hints, then the full list. Hints are
-- drawn without replacement, so what was already revealed has to be recorded.
--
-- Recording the revealed voter (rather than counting hints and deriving an order) also keeps the
-- next hint unguessable: any stable ordering would have to come from the voter ids themselves.

CREATE TABLE detector_hints (
  id               TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id        TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  revealed_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, season_id, revealed_user_id)
);

CREATE INDEX idx_detector_hints_user_season ON detector_hints(user_id, season_id);
