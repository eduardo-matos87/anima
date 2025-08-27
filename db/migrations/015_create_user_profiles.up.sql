CREATE TABLE IF NOT EXISTS public.user_profiles (
  user_id uuid PRIMARY KEY,
  height_cm integer,
  weight_kg numeric(6,2),
  birth_date date,
  training_goal text,
  experience_level text,
  notes text,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_goal  ON public.user_profiles (training_goal);
CREATE INDEX IF NOT EXISTS idx_user_profiles_level ON public.user_profiles (experience_level);
