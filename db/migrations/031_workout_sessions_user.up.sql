-- Garantimos coluna e índice de forma idempotente
DO $$
BEGIN
  -- coluna user_id
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'user_id'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN user_id TEXT;
  END IF;

  -- índice por usuário
  IF NOT EXISTS (
    SELECT 1 FROM pg_class i
    JOIN pg_namespace n ON n.oid = i.relnamespace
    WHERE i.relkind = 'i' AND i.relname = 'idx_workout_sessions_user_id' AND n.nspname = 'public'
  ) THEN
    CREATE INDEX idx_workout_sessions_user_id ON public.workout_sessions (user_id);
  END IF;
END$$;
