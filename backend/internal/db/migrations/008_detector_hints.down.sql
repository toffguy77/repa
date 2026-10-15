-- 008_detector_hints.down.sql
-- The full-list purchase path is unchanged, so dropping this leaves a working detector.

DROP INDEX IF EXISTS idx_detector_hints_user_season;
DROP TABLE IF EXISTS detector_hints;
