-- CUIDADO: down remove colunas adicionais (use só se necessário reverter)
ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS birth_year,
  DROP COLUMN IF EXISTS gender,
  DROP COLUMN IF EXISTS level,
  DROP COLUMN IF EXISTS goal;
