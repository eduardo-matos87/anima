-- 023: índices e views para overload (compatível com workout_sets/workout_sessions atuais)

-- Índices (todos idempotentes; alguns já existem e serão ignorados)
CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_completed_id_desc
  ON workout_sets (exercicio_id, completed, id DESC);

CREATE INDEX IF NOT EXISTS idx_workout_sets_exercicio_id_desc
  ON workout_sets (exercicio_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_workout_sets_session_id
  ON workout_sets (session_id);

CREATE INDEX IF NOT EXISTS idx_workout_sessions_user_id
  ON workout_sessions (user_id);

-- ===== GLOBAL (por exercício) =====
CREATE OR REPLACE VIEW workout_sets_recent12 AS
WITH ranked AS (
  SELECT
    id,
    session_id,
    exercicio_id,
    set_index,
    reps,
    weight_kg,
    rir,
    completed,
    rest_sec,
    created_at,
    ROW_NUMBER() OVER (PARTITION BY exercicio_id ORDER BY created_at DESC, id DESC) AS rn
  FROM workout_sets
  WHERE completed = TRUE
)
SELECT
  id,
  session_id,
  exercicio_id,
  set_index,
  reps,
  weight_kg,
  rir,
  completed,
  rest_sec,
  created_at
FROM ranked
WHERE rn <= 12;

CREATE OR REPLACE VIEW workout_overload_stats12 AS
SELECT
  exercicio_id,
  COALESCE(AVG(weight_kg), 0)::numeric(10,2) AS avg_carga_kg,
  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
  COUNT(*)                                   AS sample_count
FROM workout_sets_recent12
GROUP BY exercicio_id;

-- ===== POR USUÁRIO (recomendado) =====
CREATE OR REPLACE VIEW workout_sets_recent12_user AS
WITH ranked AS (
  SELECT
    ws.id,
    ws.session_id,
    s.user_id,
    ws.exercicio_id,
    ws.set_index,
    ws.reps,
    ws.weight_kg,
    ws.rir,
    ws.completed,
    ws.rest_sec,
    ws.created_at,
    ROW_NUMBER() OVER (
      PARTITION BY s.user_id, ws.exercicio_id
      ORDER BY ws.created_at DESC, ws.id DESC
    ) AS rn
  FROM workout_sets ws
  JOIN workout_sessions s ON s.id = ws.session_id
  WHERE ws.completed = TRUE
)
SELECT
  id,
  session_id,
  user_id,
  exercicio_id,
  set_index,
  reps,
  weight_kg,
  rir,
  completed,
  rest_sec,
  created_at
FROM ranked
WHERE rn <= 12;

CREATE OR REPLACE VIEW workout_overload_stats12_user AS
SELECT
  user_id,
  exercicio_id,
  COALESCE(AVG(weight_kg), 0)::numeric(10,2) AS avg_carga_kg,
  COALESCE(AVG(rir), 1.5)::numeric(10,2)     AS avg_rir,
  COUNT(*)                                   AS sample_count
FROM workout_sets_recent12_user
GROUP BY user_id, exercicio_id;
