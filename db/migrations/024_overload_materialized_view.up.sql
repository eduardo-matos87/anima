-- Materialized view com últimas 12 amostras concluídas por exercício (agregadas)
-- Reutiliza a view workout_sets_recent12 criada na 023.

-- cria MV se não existir
CREATE MATERIALIZED VIEW IF NOT EXISTS workout_overload_stats12_mv AS
SELECT
  exercicio_id,
  COALESCE(AVG(carga_kg), 0)::numeric(10,2) AS avg_carga_kg,
  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
  COUNT(*)                                    AS sample_count
FROM workout_sets_recent12
GROUP BY exercicio_id
WITH NO DATA;

-- índice único para permitir REFRESH CONCURRENTLY
CREATE UNIQUE INDEX IF NOT EXISTS workout_overload_stats12_mv_pk
ON workout_overload_stats12_mv (exercicio_id);
