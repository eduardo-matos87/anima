-- Cria uma sessão e 2 sets concluídos para exercicio_id=10 (exemplo)
-- Ajuste os IDs conforme seu ambiente.

-- usa o primeiro treino existente
WITH t AS (
  SELECT id FROM treinos ORDER BY id ASC LIMIT 1
)
INSERT INTO workout_sessions (treino_id, session_at, notes)
SELECT t.id, NOW() - INTERVAL '1 day', 'seed session A'
FROM t
RETURNING id;

-- pega a última sessão criada e adiciona sets
WITH s AS (
  SELECT id FROM workout_sessions ORDER BY id DESC LIMIT 1
)
INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
SELECT s.id, 10, 3, 10, 50, 2, TRUE, 'seed set 1'
FROM s;

WITH s AS (
  SELECT id FROM workout_sessions ORDER BY id DESC LIMIT 1
)
INSERT INTO workout_sets (session_id, exercicio_id, series, repeticoes, carga_kg, rir, completed, notes)
SELECT s.id, 10, 3, 10, 50, 2, TRUE, 'seed set 2'
FROM s;
