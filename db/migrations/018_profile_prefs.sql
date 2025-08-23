-- 018_profile_prefs.sql
-- Adiciona preferências de geração no perfil (idempotente)
BEGIN;

-- preferred_division: fullbody | upper-lower | ppl
ALTER TABLE public.user_profiles
  ADD COLUMN IF NOT EXISTS preferred_division TEXT;

UPDATE public.user_profiles
  SET preferred_division = 'fullbody'
  WHERE preferred_division IS NULL;

ALTER TABLE public.user_profiles
  ALTER COLUMN preferred_division SET DEFAULT 'fullbody';

-- cria a constraint se ainda não existir
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'chk_user_profiles_preferred_division'
  ) THEN
    ALTER TABLE public.user_profiles
      ADD CONSTRAINT chk_user_profiles_preferred_division
      CHECK (preferred_division IN ('fullbody','upper-lower','ppl'));
  END IF;
END$$;

ALTER TABLE public.user_profiles
  ALTER COLUMN preferred_division SET NOT NULL;

-- days_per_week: 1..7
ALTER TABLE public.user_profiles
  ADD COLUMN IF NOT EXISTS days_per_week INT;

UPDATE public.user_profiles
  SET days_per_week = 3
  WHERE days_per_week IS NULL;

ALTER TABLE public.user_profiles
  ALTER COLUMN days_per_week SET DEFAULT 3;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'chk_user_profiles_days_per_week'
  ) THEN
    ALTER TABLE public.user_profiles
      ADD CONSTRAINT chk_user_profiles_days_per_week
      CHECK (days_per_week BETWEEN 1 AND 7);
  END IF;
END$$;

ALTER TABLE public.user_profiles
  ALTER COLUMN days_per_week SET NOT NULL;

-- equipment: JSONB {}
ALTER TABLE public.user_profiles
  ADD COLUMN IF NOT EXISTS equipment JSONB;

UPDATE public.user_profiles
  SET equipment = '{}'::jsonb
  WHERE equipment IS NULL;

ALTER TABLE public.user_profiles
  ALTER COLUMN equipment SET DEFAULT '{}'::jsonb,
  ALTER COLUMN equipment SET NOT NULL;

-- injuries: JSONB []
ALTER TABLE public.user_profiles
  ADD COLUMN IF NOT EXISTS injuries JSONB;

UPDATE public.user_profiles
  SET injuries = '[]'::jsonb
  WHERE injuries IS NULL;

ALTER TABLE public.user_profiles
  ALTER COLUMN injuries SET DEFAULT '[]'::jsonb,
  ALTER COLUMN injuries SET NOT NULL;

-- use_ai já existe no seu schema; se não existir, descomente abaixo:
-- ALTER TABLE public.user_profiles
--   ADD COLUMN IF NOT EXISTS use_ai BOOLEAN NOT NULL DEFAULT false;

COMMIT;
