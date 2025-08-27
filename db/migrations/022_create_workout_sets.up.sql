-- cria workout_sets (idempotente)
CREATE TABLE IF NOT EXISTS workout_sets (
  id           BIGSERIAL PRIMARY KEY,
  session_id   BIGINT NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
  exercicio_id BIGINT NOT NULL,
  series       INT NOT NULL,
  repeticoes   INT NOT NULL,
  carga_kg     NUMERIC(10,2),
  rir          INT,
  completed    BOOLEAN NOT NULL DEFAULT FALSE,
  notes        TEXT
);

CREATE INDEX IF NOT EXISTS idx_workout_sets_session_id   ON workout_sets (session_id);
CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_id ON workout_sets (exercicio_id);
