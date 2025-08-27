CREATE TABLE IF NOT EXISTS public.user_metrics (
  user_id uuid NOT NULL,
  measured_at date NOT NULL,
  weight_kg  numeric(6,2),
  bodyfat_pct numeric(5,2),
  height_cm integer,
  neck_cm   numeric(6,2),
  waist_cm  numeric(6,2),
  hip_cm    numeric(6,2),
  notes text,
  PRIMARY KEY (user_id, measured_at)
);

CREATE INDEX IF NOT EXISTS idx_user_metrics_user ON public.user_metrics (user_id);
CREATE INDEX IF NOT EXISTS idx_user_metrics_date ON public.user_metrics (measured_at);
