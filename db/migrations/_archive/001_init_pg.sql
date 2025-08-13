-- Schema base do Anima (PostgreSQL)

CREATE TABLE IF NOT EXISTS muscle_groups (
	  id BIGSERIAL PRIMARY KEY,
	  name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS equipment (
		  id BIGSERIAL PRIMARY KEY,
		  name TEXT UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS exercises (
			  id BIGSERIAL PRIMARY KEY,
			  name TEXT NOT NULL,
			  primary_muscle TEXT NOT NULL,  -- peito, costas, pernas, ombros, biceps, triceps, core, panturrilha
			  difficulty TEXT NOT NULL CHECK (difficulty IN ('beginner','intermediate','advanced')),
			  equipment TEXT NOT NULL,       -- halter, barra, maquina, livre
			  is_unilateral BOOLEAN DEFAULT FALSE
			);

			CREATE TABLE IF NOT EXISTS workouts (
				  id BIGSERIAL PRIMARY KEY,
				  goal  TEXT NOT NULL,           -- hipertrofia | forca | resistencia
				  level TEXT NOT NULL,           -- beginner | intermediate | advanced
				  days_per_week INT NOT NULL CHECK (days_per_week BETWEEN 1 AND 7),
				  notes TEXT,
				  created_at TIMESTAMPTZ DEFAULT now()
				);

				CREATE TABLE IF NOT EXISTS workout_exercises (
					  id BIGSERIAL PRIMARY KEY,
					  workout_id BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
					  day_index INT NOT NULL CHECK (day_index BETWEEN 1 AND 7),
					  exercise_id BIGINT NOT NULL REFERENCES exercises(id),
					  sets INT NOT NULL CHECK (sets BETWEEN 1 AND 10),
					  reps TEXT NOT NULL,            -- ex: '8-12'
					  rest_seconds INT NOT NULL CHECK (rest_seconds BETWEEN 15 AND 600),
					  tempo TEXT
					);

					-- Índices úteis
CREATE INDEX IF NOT EXISTS idx_exercises_primary_muscle ON exercises (primary_muscle);
CREATE INDEX IF NOT EXISTS idx_exercises_equipment      ON exercises (equipment);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_wid    ON workout_exercises (workout_id);

