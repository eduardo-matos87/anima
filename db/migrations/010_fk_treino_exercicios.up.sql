DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_te_treino') THEN
    ALTER TABLE treino_exercicios
      ADD CONSTRAINT fk_te_treino
      FOREIGN KEY (treino_id) REFERENCES treinos(id) ON DELETE CASCADE;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_te_exercicio') THEN
    ALTER TABLE treino_exercicios
      ADD CONSTRAINT fk_te_exercicio
      FOREIGN KEY (exercicio_id) REFERENCES exercises(id) ON DELETE CASCADE;
  END IF;
END$$;
