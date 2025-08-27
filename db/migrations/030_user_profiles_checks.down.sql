ALTER TABLE public.user_profiles
  DROP CONSTRAINT IF EXISTS ck_user_profiles_height_cm,
  DROP CONSTRAINT IF EXISTS ck_user_profiles_weight_kg,
  DROP CONSTRAINT IF EXISTS ck_user_profiles_birth_year;
