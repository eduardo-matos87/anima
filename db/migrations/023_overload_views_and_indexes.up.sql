-- Índices para acelerar buscas por exercício / concluído / ordem decrescente
CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_completed_id_desc
  ON workout_sets (exercicio_id, completed, id DESC);

CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_id_desc
  ON workout_sets (exercicio_id, id DESC);

-- Últimos 12 sets CONCLUÍDOS por exercício (usa janela por exercício)
CREATE OR REPLACE VIEW workout_sets_recent12 AS
WITH ranked AS (
  SELECT
    id,
    session_id,
    exercicio_id,
    series,
    repeticoes,
    carga_kg,
    rir,
    completed,
    notes,
    ROW_NUMBER() OVER (PARTITION BY exercicio_id ORDER BY id DESC) AS rn
  FROM workout_sets
  WHERE completed = TRUE
)
SELECT
  id,
  session_id,
  exercicio_id,
  series,
  repeticoes,
  carga_kg,
  rir,
  completed,
  notes
FROM ranked
WHERE rn <= 12;

-- Estatísticas (médias) dos últimos 12 sets concluídos por exercício
CREATE OR REPLACE VIEW workout_overload_stats12 AS
SELECT
  exercicio_id,
  COALESCE(AVG(carga_kg), 0)::numeric(10,2) AS avg_carga_kg,
  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
  COUNT(*)                                   AS sample_count
FROM workout_sets_recent12
GROUP BY exercicio_id;
