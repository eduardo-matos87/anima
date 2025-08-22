-- Adiciona uma chave lógica para lookup por string do /generate
ALTER TABLE public.treinos
  ADD COLUMN IF NOT EXISTS treino_key text;

-- Garante unicidade quando presente (sem bloquear NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_treinos_treino_key_unique
  ON public.treinos (treino_key)
  WHERE treino_key IS NOT NULL;
