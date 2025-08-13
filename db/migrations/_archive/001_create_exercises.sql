-- Tabelas base para geração de treinos
CREATE TABLE IF NOT EXISTS exercises (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  muscle_group TEXT NOT NULL,      -- chest, back, legs, shoulders, arms, core
  equipment TEXT NOT NULL,         -- barbell, dumbbell, machine, bodyweight, cable
  difficulty TEXT NOT NULL,        -- beginner, intermediate, advanced
  is_compound BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_exercises_group ON exercises(muscle_group);
CREATE INDEX IF NOT EXISTS idx_exercises_diff  ON exercises(difficulty);
