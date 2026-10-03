-- 005_invite_code_short.down.sql

ALTER TABLE groups ADD CONSTRAINT groups_invite_code_key UNIQUE (invite_code);

DROP INDEX IF EXISTS groups_invite_code_upper_idx;
