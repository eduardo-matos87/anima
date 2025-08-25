-- 018: Cria tabela de sessões de treino (workout_sessions)

CREATE TABLE IF NOT EXISTS workout_sessions (
  id            BIGSERIAL PRIMARY KEY,
  user_id       TEXT NOT NULL,                          -- alinha com X-User-ID e /api/me/*
  treino_id     INTEGER NULL REFERENCES treinos(id) ON DELETE SET NULL,
  started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ended_at      TIMESTAMPTZ NULL,
  duration_sec  INTEGER NULL,                           -- opcional (pode ser calculado), útil para filtros rápidos
  notes         TEXT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workout_sessions_user_id      ON workout_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_sessions_started_at   ON workout_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_workout_sessions_treino_id    ON workout_sessions(treino_id);
