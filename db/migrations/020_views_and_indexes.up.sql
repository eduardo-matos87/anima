-- 020: Índices adicionais e visão de resumo (opcional)

-- Acelera buscas por usuário + recência
CREATE INDEX IF NOT EXISTS idx_workout_sessions_user_started_desc
  ON workout_sessions(user_id, started_at DESC);

-- View simples de resumo (sessão + contagem de séries e volume total)
CREATE OR REPLACE VIEW vw_session_summary AS
SELECT
  s.id                AS session_id,
  s.user_id,
  s.treino_id,
  s.started_at,
  s.ended_at,
  s.duration_sec,
  COUNT(ws.id)                      AS total_sets,
  COALESCE(SUM(CASE WHEN ws.weight_kg IS NOT NULL AND ws.reps IS NOT NULL
                    THEN (ws.weight_kg * ws.reps)::NUMERIC
                    ELSE 0 END), 0) AS total_volume
FROM workout_sessions s
LEFT JOIN workout_sets ws ON ws.session_id = s.id
GROUP BY s.id, s.user_id, s.treino_id, s.started_at, s.ended_at, s.duration_sec;
