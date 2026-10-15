-- 005_invite_code_short.up.sql
-- Short, transcribable invite codes. Uniqueness becomes case-insensitive so a code typed in
-- any case resolves to exactly one group, and two simultaneous creations cannot land on the
-- same code.
--
-- Codes created before this change (36-char UUIDs) are left untouched: validating stored
-- codes would mean rewriting them, and rewriting them would break links already shared.

CREATE UNIQUE INDEX IF NOT EXISTS groups_invite_code_upper_idx ON groups (upper(invite_code));

-- The plain unique constraint is superseded by the functional index above.
ALTER TABLE groups DROP CONSTRAINT IF EXISTS groups_invite_code_key;
