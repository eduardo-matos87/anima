-- 017_add_profile_prefs.down.sql
ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS activity_level,
  DROP COLUMN IF EXISTS use_ai;
