-- 020_auth.sql
-- Auth básica: senha + refresh tokens
BEGIN;


ALTER TABLE public.users
ADD COLUMN IF NOT EXISTS password_hash TEXT;


CREATE TABLE IF NOT EXISTS public.refresh_tokens (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
token TEXT NOT NULL,
expires_at TIMESTAMPTZ NOT NULL,
revoked BOOLEAN NOT NULL DEFAULT false,
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON public.refresh_tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_tokens_token ON public.refresh_tokens(token);


COMMIT;