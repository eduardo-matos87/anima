ALTER TABLE public.workout_sessions
  ADD COLUMN IF NOT EXISTS completed      BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS duration_min   INT,
  ADD COLUMN IF NOT EXISTS rpe_session    INT,
  ADD COLUMN IF NOT EXISTS updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Regras de integridade leves
ALTER TABLE public.workout_sessions
  DROP CONSTRAINT IF EXISTS ck_workout_sessions_rpe;

ALTER TABLE public.workout_sessions
  ADD CONSTRAINT ck_workout_sessions_rpe
    CHECK (rpe_session IS NULL OR (rpe_session BETWEEN 1 AND 10));

CREATE INDEX IF NOT EXISTS idx_workout_sessions_updated_at
  ON public.workout_sessions (updated_at DESC);
