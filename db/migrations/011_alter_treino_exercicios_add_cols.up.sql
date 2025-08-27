-- Adiciona colunas de configuração por exercício no treino
ALTER TABLE public.treino_exercicios
  ADD COLUMN IF NOT EXISTS series integer NOT NULL DEFAULT 3,
  ADD COLUMN IF NOT EXISTS repeticoes text NOT NULL DEFAULT '8-12';
