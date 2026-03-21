-- 001_init.down.sql

DROP TABLE IF EXISTS next_season_votes, push_preferences, reports, reactions,
  card_cache, fcm_tokens, crystal_logs, detectors, user_group_stats,
  achievements, season_results, votes, season_questions, questions,
  seasons, group_members, groups, users CASCADE;

DROP TYPE IF EXISTS push_category, achievement_type, crystal_log_type,
  question_status, question_source, question_category, season_status;
