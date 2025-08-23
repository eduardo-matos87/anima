-- 019_treino_sessions.up.sql (fix: treino_id INTEGER)
BEGIN;

-- 1) Tabela de sessões
CREATE TABLE IF NOT EXISTS public.treino_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  treino_id INTEGER NOT NULL REFERENCES public.treinos(id) ON DELETE CASCADE,
  label TEXT NOT NULL,
  division_day TEXT NOT NULL CHECK (division_day IN ('full','upper','lower','push','pull','legs')),
  day_index INT NOT NULL CHECK (day_index BETWEEN 1 AND 7),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 1.1) Constraint de unicidade (um dia por treino)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'uq_treino_sessions_treino_day'
  ) THEN
    ALTER TABLE public.treino_sessions
      ADD CONSTRAINT uq_treino_sessions_treino_day
      UNIQUE (treino_id, day_index);
  END IF;
END$$;

-- 2) Coluna session_id em treino_exercicios (se não existir)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema='public' AND table_name='treino_exercicios' AND column_name='session_id'
  ) THEN
    ALTER TABLE public.treino_exercicios
      ADD COLUMN session_id UUID NULL;
  END IF;
END$$;

-- 2.1) FK (se não existir)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_treino_exercicios_session'
  ) THEN
    ALTER TABLE public.treino_exercicios
      ADD CONSTRAINT fk_treino_exercicios_session
      FOREIGN KEY (session_id) REFERENCES public.treino_sessions(id) ON DELETE SET NULL;
  END IF;
END$$;

-- 3) Backfill: cria 1 sessão 'Full Day 1' para treinos sem sessão
INSERT INTO public.treino_sessions (treino_id, label, division_day, day_index)
SELECT t.id, 'Full Day 1', 'full', 1
FROM public.treinos t
WHERE NOT EXISTS (
  SELECT 1 FROM public.treino_sessions s WHERE s.treino_id = t.id
);

-- 4) Atribui exercícios sem session_id à sessão criada
UPDATE public.treino_exercicios te
SET session_id = s.id
FROM public.treino_sessions s
WHERE te.treino_id = s.treino_id
  AND te.session_id IS NULL;

-- 5) Índices
CREATE INDEX IF NOT EXISTS idx_treino_sessions_treino_day
  ON public.treino_sessions (treino_id, day_index);

CREATE INDEX IF NOT EXISTS idx_treino_exercicios_session
  ON public.treino_exercicios (session_id);

COMMIT;
