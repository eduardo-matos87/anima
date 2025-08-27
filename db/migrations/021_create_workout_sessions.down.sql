DROP TRIGGER IF EXISTS tr_workout_sessions_updated_at ON workout_sessions;
-- não derruba a função se for usada por outras tabelas, mas se quiser:
-- DROP FUNCTION IF EXISTS set_updated_at;
DROP TABLE IF EXISTS workout_sessions;
