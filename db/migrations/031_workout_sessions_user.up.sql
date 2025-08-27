-- Add user_id às sessões (compatível com dados existentes)
ALTER TABLE public.workout_sessions
  ADD COLUMN IF NOT EXISTS user_id TEXT;

CREATE INDEX IF NOT EXISTS idx_workout_sessions_user_id
  ON public.workout_sessions (user_id);
