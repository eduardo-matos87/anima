DROP INDEX IF EXISTS idx_workout_sessions_user_id;
ALTER TABLE public.workout_sessions
  DROP COLUMN IF EXISTS user_id;
