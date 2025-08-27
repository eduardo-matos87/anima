CREATE TABLE IF NOT EXISTS admin_metadata (
  key TEXT PRIMARY KEY,
  value TEXT,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- marca a hora do refresh com trigger simples
CREATE OR REPLACE FUNCTION admin_touch_refresh() RETURNS void AS $$
BEGIN
  INSERT INTO admin_metadata(key, value, updated_at)
  VALUES ('overload_last_refresh', NOW()::text, NOW())
  ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
END; $$ LANGUAGE plpgsql;
