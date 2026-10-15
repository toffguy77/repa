-- 009_member_safety.up.sql
-- A member can get away from another member. Required by App Store Guideline 1.2 (filtering, a report
-- mechanism, and the ability to block abusive users) and by the PRD's own bullying risk.

-- Who created the block is recorded so it can be undone by them; the *effect* is symmetric, because a
-- one-directional block would keep asking the blocked person to rate someone who withdrew, and would
-- keep delivering their votes to that person.
CREATE TABLE blocks (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  blocker_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  blocked_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (blocker_id, blocked_id),
  CHECK (blocker_id <> blocked_id)
);

CREATE INDEX idx_blocks_blocker ON blocks(blocker_id);
CREATE INDEX idx_blocks_blocked ON blocks(blocked_id);

-- Separates "left" from "cannot come back". Ordinary leaving stays reversible — people leave by
-- accident. A removal or a deliberate permanent departure writes a ban that the join path checks, which
-- is why regenerating the invite does not undo it: the ban is on the person and the group, not the code.
CREATE TABLE group_bans (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  -- Who ended the membership: the admin, or the member themselves.
  banned_by  TEXT REFERENCES users(id) ON DELETE SET NULL,
  reason     TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (group_id, user_id)
);

CREATE INDEX idx_group_bans_group_user ON group_bans(group_id, user_id);

-- A separate queue from `reports`, whose question_id is NOT NULL and whose every query joins through
-- it. Making that nullable would force each of those to handle the null case for a different kind of
-- report.
CREATE TABLE user_reports (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  reported_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reporter_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id    TEXT REFERENCES groups(id) ON DELETE SET NULL,
  reason      TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (reported_id, reporter_id),
  CHECK (reported_id <> reporter_id)
);

CREATE INDEX idx_user_reports_created ON user_reports(created_at DESC);
