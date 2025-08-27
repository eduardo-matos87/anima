-- db/migrations/030_user_profiles_checks.up.sql (versão resilient)
BEGIN;

-- saneia dados fora da faixa
UPDATE public.user_profiles
  SET height_cm = NULL
  WHERE height_cm IS NOT NULL AND (height_cm < 50 OR height_cm > 250);

UPDATE public.user_profiles
  SET weight_kg = NULL
  WHERE weight_kg IS NOT NULL AND (weight_kg < 20 OR weight_kg > 500);

UPDATE public.user_profiles
  SET birth_year = NULL
  WHERE birth_year IS NOT NULL AND (birth_year < 1900 OR birth_year > EXTRACT(YEAR FROM NOW())::INT);

-- recria constraints (idempotência simples: drop e add)
ALTER TABLE public.user_profiles
  DROP CONSTRAINT IF EXISTS ck_user_profiles_height_cm,
  DROP CONSTRAINT IF EXISTS ck_user_profiles_weight_kg,
  DROP CONSTRAINT IF EXISTS ck_user_profiles_birth_year;

ALTER TABLE public.user_profiles
  ADD CONSTRAINT ck_user_profiles_height_cm
    CHECK (height_cm IS NULL OR (height_cm >= 50 AND height_cm <= 250)),
  ADD CONSTRAINT ck_user_profiles_weight_kg
    CHECK (weight_kg IS NULL OR (weight_kg >= 20 AND weight_kg <= 500)),
  ADD CONSTRAINT ck_user_profiles_birth_year
    CHECK (birth_year IS NULL OR (birth_year BETWEEN 1900 AND EXTRACT(YEAR FROM NOW())::INT));

COMMIT;
