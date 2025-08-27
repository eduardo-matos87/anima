CREATE TABLE IF NOT EXISTS exercises (
  id            SERIAL PRIMARY KEY,
  name          TEXT        NOT NULL,
  muscle_group  TEXT        NOT NULL,
  equipment     TEXT[]      NOT NULL DEFAULT '{}',
  difficulty    TEXT        NOT NULL DEFAULT 'iniciante',
  is_bodyweight BOOLEAN     NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_exercises_muscle ON exercises(muscle_group);
