-- 017_add_profile_prefs.up.sql
ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS use_ai boolean DEFAULT true,
  ADD COLUMN IF NOT EXISTS activity_level text;
