-- 002_workout_demo.sql — compatível com started_at e weight_kg/reps/set_index

-- sessão de demo para o usuário 'seed'
WITH s AS (
  INSERT INTO workout_sessions (user_id, started_at, notes)
  VALUES ('seed', NOW() - INTERVAL '1 day', 'seed session A')
  RETURNING id
)
INSERT INTO workout_sets (
  session_id, exercicio_id, set_index, weight_kg, reps, rir, completed, rest_sec
)
SELECT id, 1, 1, 40.00, 10, 2, TRUE, 90 FROM s
UNION ALL
SELECT id, 1, 2, 42.50, 10, 2, TRUE, 90 FROM s;

-- outra sessão pra diversificar histórico
WITH s2 AS (
  INSERT INTO workout_sessions (user_id, started_at, notes)
  VALUES ('seed', NOW() - INTERVAL '2 days', 'seed session B')
  RETURNING id
)
INSERT INTO workout_sets (
  session_id, exercicio_id, set_index, weight_kg, reps, rir, completed, rest_sec
)
SELECT id, 2, 1, 25.00, 12, 2, TRUE, 60 FROM s2
UNION ALL
SELECT id, 2, 2, 27.50, 10, 1, TRUE, 60 FROM s2;
