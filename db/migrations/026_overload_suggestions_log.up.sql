-- log de sugestões de overload
CREATE TABLE IF NOT EXISTS overload_suggestions_log (
  id                     BIGSERIAL PRIMARY KEY,
  requested_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id                TEXT,
  ip                     INET,
  user_agent             TEXT,
  exercicio_id           BIGINT NOT NULL,
  window_size            INT NOT NULL,               -- <== renomeado (evita palavra reservada)
  avg_carga_kg           NUMERIC(10,2),
  avg_rir                NUMERIC(10,2),
  sample_count           INT,
  suggested_carga_kg     NUMERIC(10,2),
  suggested_repeticoes   INT,
  rationale              TEXT
);

-- se por acaso já existir a coluna "window", renomeia para window_size
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'overload_suggestions_log' AND column_name = 'window'
  ) THEN
    EXECUTE 'ALTER TABLE overload_suggestions_log RENAME COLUMN "window" TO window_size';
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_overload_log_exercicio_at
  ON overload_suggestions_log (exercicio_id, requested_at DESC);

CREATE INDEX IF NOT EXISTS idx_overload_log_user_at
  ON overload_suggestions_log (user_id, requested_at DESC);
