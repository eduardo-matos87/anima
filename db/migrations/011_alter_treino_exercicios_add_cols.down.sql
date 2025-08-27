-- Remove colunas adicionadas
ALTER TABLE public.treino_exercicios
  DROP COLUMN IF EXISTS repeticoes,
  DROP COLUMN IF EXISTS series;
