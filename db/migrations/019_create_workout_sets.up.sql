-- 019: Cria tabela de séries por sessão (workout_sets)

CREATE TABLE IF NOT EXISTS workout_sets (
  id            BIGSERIAL PRIMARY KEY,
  session_id    BIGINT NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
  exercicio_id  INTEGER NOT NULL REFERENCES exercises(id),
  set_index     INTEGER NOT NULL,                -- 1..N dentro da sessão/exercício
  weight_kg     NUMERIC(6,2) NULL,               -- carga usada (se aplicável)
  reps          INTEGER NULL,                     -- repetições realizadas
  rir           INTEGER NULL,                     -- Reps In Reserve (opcional)
  completed     BOOLEAN NOT NULL DEFAULT TRUE,    -- marca se a série foi concluída
  rest_sec      INTEGER NULL,                     -- descanso após a série (se medido)
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Evita duplicidade lógica de (session_id, exercicio_id, set_index)
CREATE UNIQUE INDEX IF NOT EXISTS uq_workout_sets_session_exercicio_idx
  ON workout_sets(session_id, exercicio_id, set_index);

CREATE INDEX IF NOT EXISTS idx_workout_sets_session_id    ON workout_sets(session_id);
CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_id  ON workout_sets(exercicio_id);
