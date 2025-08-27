-- 024: materialized view de estatísticas de overload por USUÁRIO + exercício
-- Compatível com:
--   workout_sets(weight_kg, reps, rir, set_index, completed, created_at, session_id)
--   workout_sessions(user_id)
--   view workout_sets_recent12_user (criada na 023)

-- Recria de forma idempotente
DROP MATERIALIZED VIEW IF EXISTS workout_overload_stats12_user_mv;

CREATE MATERIALIZED VIEW workout_overload_stats12_user_mv AS
SELECT
  user_id,
  exercicio_id,
  COALESCE(AVG(weight_kg), 0)::numeric(10,2) AS avg_carga_kg,
  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
  COUNT(*)                                   AS sample_count
FROM workout_sets_recent12_user
GROUP BY user_id, exercicio_id;

-- Índice para lookups rápidos pelo handler
CREATE UNIQUE INDEX IF NOT EXISTS idx_overload12_user_mv_pk
  ON workout_overload_stats12_user_mv (user_id, exercicio_id);

-- Primeira materialização
REFRESH MATERIALIZED VIEW workout_overload_stats12_user_mv;
