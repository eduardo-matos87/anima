-- perfis de usuário (vinculados ao user_id vindo do JWT/HEADER)
CREATE TABLE IF NOT EXISTS user_profiles (
  user_id     TEXT PRIMARY KEY,
  height_cm   NUMERIC(5,2),
  weight_kg   NUMERIC(6,2),
  birth_year  INT,
  gender      TEXT,           -- livre; ex.: male/female/other
  level       TEXT,           -- iniciante/intermediario/avancado
  goal        TEXT,           -- hipertrofia/emagrecimento/etc
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_updated_at
  ON user_profiles (updated_at DESC);
