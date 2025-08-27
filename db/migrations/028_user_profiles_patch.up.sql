-- Ajusta user_profiles para o formato atual usado pelos handlers
ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS birth_year  INT,
  ADD COLUMN IF NOT EXISTS gender      TEXT,
  ADD COLUMN IF NOT EXISTS level       TEXT,
  ADD COLUMN IF NOT EXISTS goal        TEXT,
  ADD COLUMN IF NOT EXISTS updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Índice (idempotente)
CREATE INDEX IF NOT EXISTS idx_user_profiles_updated_at
  ON user_profiles (updated_at DESC);
