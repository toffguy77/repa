-- 001_init.up.sql

CREATE TYPE season_status AS ENUM ('VOTING', 'REVEALED', 'CLOSED');
CREATE TYPE question_category AS ENUM ('HOT', 'FUNNY', 'SECRETS', 'SKILLS', 'ROMANCE', 'STUDY');
CREATE TYPE question_source AS ENUM ('SYSTEM', 'USER');
CREATE TYPE question_status AS ENUM ('ACTIVE', 'PENDING', 'REJECTED');
CREATE TYPE crystal_log_type AS ENUM ('PURCHASE', 'SPEND_DETECTOR', 'SPEND_ATTRIBUTES', 'SPEND_QUESTION', 'BONUS');
CREATE TYPE achievement_type AS ENUM (
  'SNIPER', 'ORACLE', 'TELEPATH', 'BLIND', 'RANDOM',
  'EXPERT_OF', 'BEST_FRIEND', 'DETECTIVE', 'STRANGER',
  'LEGEND', 'CHANGEABLE', 'MONOPOLIST', 'ENIGMA', 'RISING', 'PIONEER',
  'STREAK_VOTER', 'FIRST_VOTER', 'LAST_VOTER', 'NIGHT_OWL', 'ANALYST',
  'MEDIA', 'CONSPIRATOR', 'RECRUITER'
);
CREATE TYPE push_category AS ENUM ('SEASON_START', 'REMINDER', 'REVEAL', 'REACTION', 'NEXT_SEASON');

CREATE TABLE users (
  id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  phone         TEXT UNIQUE,
  apple_id      TEXT UNIQUE,
  google_id     TEXT UNIQUE,
  username      TEXT UNIQUE NOT NULL,
  avatar_url    TEXT,
  avatar_emoji  TEXT,
  birth_year    INT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE groups (
  id                      TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  name                    TEXT NOT NULL,
  invite_code             TEXT UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
  admin_id                TEXT NOT NULL REFERENCES users(id),
  telegram_chat_id        TEXT,
  telegram_chat_username  TEXT,
  telegram_connect_code   TEXT,
  telegram_connect_expiry TIMESTAMPTZ,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
  id        TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id  TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE seasons (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  number     INT NOT NULL,
  status     season_status NOT NULL DEFAULT 'VOTING',
  starts_at  TIMESTAMPTZ NOT NULL,
  reveal_at  TIMESTAMPTZ NOT NULL,
  ends_at    TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  text       TEXT NOT NULL,
  category   question_category NOT NULL,
  source     question_source NOT NULL DEFAULT 'SYSTEM',
  group_id   TEXT REFERENCES groups(id) ON DELETE CASCADE,
  author_id  TEXT REFERENCES users(id),
  status     question_status NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE season_questions (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  question_id TEXT NOT NULL REFERENCES questions(id),
  ord         INT NOT NULL,
  UNIQUE(season_id, question_id)
);

CREATE TABLE votes (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id   TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  voter_id    TEXT NOT NULL REFERENCES users(id),
  target_id   TEXT NOT NULL REFERENCES users(id),
  question_id TEXT NOT NULL REFERENCES questions(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, voter_id, question_id)
);

CREATE TABLE season_results (
  id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id    TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  target_id    TEXT NOT NULL REFERENCES users(id),
  question_id  TEXT NOT NULL REFERENCES questions(id),
  vote_count   INT NOT NULL,
  total_voters INT NOT NULL,
  percentage   FLOAT NOT NULL,
  UNIQUE(season_id, target_id, question_id)
);

CREATE TABLE achievements (
  id               TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id         TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  season_id        TEXT REFERENCES seasons(id),
  achievement_type achievement_type NOT NULL,
  metadata         JSONB,
  earned_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_group_stats (
  id                   TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id              TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  group_id             TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  seasons_played       INT NOT NULL DEFAULT 0,
  voting_streak        INT NOT NULL DEFAULT 0,
  max_voting_streak    INT NOT NULL DEFAULT 0,
  guess_accuracy       FLOAT NOT NULL DEFAULT 0,
  total_votes_cast     INT NOT NULL DEFAULT 0,
  total_votes_received INT NOT NULL DEFAULT 0,
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, group_id)
);

CREATE TABLE detectors (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  group_id   TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE crystal_logs (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  delta       INT NOT NULL,
  balance     INT NOT NULL,
  type        crystal_log_type NOT NULL,
  description TEXT,
  external_id TEXT UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fcm_tokens (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token      TEXT UNIQUE NOT NULL,
  platform   TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE card_cache (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  image_url  TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, season_id)
);

CREATE TABLE reactions (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  season_id  TEXT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  reactor_id TEXT NOT NULL REFERENCES users(id),
  target_id  TEXT NOT NULL REFERENCES users(id),
  emoji      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(season_id, reactor_id, target_id)
);

CREATE TABLE reports (
  id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  question_id TEXT NOT NULL REFERENCES questions(id),
  reporter_id TEXT NOT NULL REFERENCES users(id),
  reason      TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(question_id, reporter_id)
);

CREATE TABLE push_preferences (
  id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category   push_category NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(user_id, category)
);

CREATE TABLE next_season_votes (
  id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  group_id      TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id       TEXT NOT NULL REFERENCES users(id),
  question_id   TEXT NOT NULL REFERENCES questions(id),
  season_number INT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(group_id, user_id, season_number)
);

-- Indexes
CREATE INDEX idx_votes_season ON votes(season_id);
CREATE INDEX idx_votes_target_season ON votes(target_id, season_id);
CREATE INDEX idx_season_results_season ON season_results(season_id, target_id);
CREATE INDEX idx_achievements_user_group ON achievements(user_id, group_id);
CREATE INDEX idx_user_group_stats ON user_group_stats(user_id, group_id);
CREATE INDEX idx_seasons_group_status ON seasons(group_id, status);
