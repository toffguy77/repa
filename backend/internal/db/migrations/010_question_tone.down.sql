-- 010_question_tone.down.sql
-- Selection falls back to category-only, which is the pre-change behaviour.

DROP INDEX IF EXISTS idx_questions_tone;

ALTER TABLE groups DROP COLUMN IF EXISTS kind_only;
ALTER TABLE questions DROP COLUMN IF EXISTS tone;

DROP TYPE IF EXISTS question_tone;
