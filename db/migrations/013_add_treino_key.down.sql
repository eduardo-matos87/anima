DROP INDEX IF EXISTS idx_treinos_treino_key_unique;

ALTER TABLE public.treinos
  DROP COLUMN IF EXISTS treino_key;
