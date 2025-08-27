ALTER TABLE public.treinos
  ADD COLUMN IF NOT EXISTS coach_notes text;
