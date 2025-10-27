-- Adiciona colunas de forma idempotente
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'completed'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN completed BOOLEAN NOT NULL DEFAULT FALSE;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'duration_min'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN duration_min INT;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'rpe_session'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN rpe_session INT;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'intensity_level'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN intensity_level TEXT;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'workout_sessions' AND column_name = 'updated_at'
  ) THEN
    ALTER TABLE public.workout_sessions ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
  END IF;
END$$;

-- Regras de integridade leves
-- Recria constraint de forma idempotente
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE c.conname = 'ck_workout_sessions_rpe' AND n.nspname = 'public' AND t.relname = 'workout_sessions'
  ) THEN
    ALTER TABLE public.workout_sessions DROP CONSTRAINT ck_workout_sessions_rpe;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE c.conname = 'ck_workout_sessions_rpe' AND n.nspname = 'public' AND t.relname = 'workout_sessions'
  ) THEN
    ALTER TABLE public.workout_sessions
      ADD CONSTRAINT ck_workout_sessions_rpe CHECK (rpe_session IS NULL OR (rpe_session BETWEEN 1 AND 10));
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_workout_sessions_updated_at
  ON public.workout_sessions (updated_at DESC);
