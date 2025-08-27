ALTER TABLE public.workout_sessions
  DROP CONSTRAINT IF EXISTS ck_workout_sessions_rpe;

ALTER TABLE public.workout_sessions
  DROP COLUMN IF EXISTS completed,
  DROP COLUMN IF EXISTS duration_min,
  DROP COLUMN IF EXISTS rpe_session,
  DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_workout_sessions_updated_at;
