-- cria workout_sessions + trigger de updated_at (idempotente)
CREATE TABLE IF NOT EXISTS workout_sessions (
  id          BIGSERIAL PRIMARY KEY,
  treino_id   BIGINT NOT NULL,
  started_at  TIMESTAMPTZ NOT NULL,
  notes       TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS IF NOT EXISTS idx_workout_sessions_treino_id ON workout_sessions (treino_id);
CREATE INDEX IF NOT EXISTS IF NOT EXISTS idx_workout_sessions_session_at ON workout_sessions (started_at);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_updated_at') THEN
    CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $BODY$
    BEGIN
      NEW.updated_at = NOW();
      RETURN NEW;
    END;
    $BODY$ LANGUAGE plpgsql;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tr_workout_sessions_updated_at') THEN
    CREATE TRIGGER tr_workout_sessions_updated_at
    BEFORE UPDATE ON workout_sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;
