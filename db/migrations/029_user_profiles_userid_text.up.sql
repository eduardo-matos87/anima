-- Converte user_profiles.user_id para TEXT se ainda for UUID
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'user_profiles'
      AND column_name = 'user_id'
      AND data_type = 'uuid'
  ) THEN
    ALTER TABLE public.user_profiles
      ALTER COLUMN user_id TYPE TEXT USING user_id::text;
  END IF;
END $$;

-- Garante que a PK exista (só cria se não houver PK)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint c
    JOIN pg_class r ON r.oid = c.conrelid
    WHERE c.contype = 'p'
      AND r.relname = 'user_profiles'
  ) THEN
    ALTER TABLE public.user_profiles
      ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id);
  END IF;
END $$;
