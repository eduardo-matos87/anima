-- baseline inicial do schema (não conflita com outras migrations)
CREATE SCHEMA IF NOT EXISTS public;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- === Core mínimo exigido pela 002_add_treinos ===
CREATE TABLE IF NOT EXISTS treinos (
  id       SERIAL PRIMARY KEY,
  nivel    TEXT    NOT NULL,
  objetivo TEXT    NOT NULL,
  dias     INT     NOT NULL,
  divisao  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS treino_exercicios (
  id           SERIAL PRIMARY KEY,
  treino_id    INT NOT NULL,
  exercicio_id INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_treino_exercicios_treino_id ON treino_exercicios(treino_id);
CREATE INDEX IF NOT EXISTS idx_treino_exercicios_exercicio_id ON treino_exercicios(exercicio_id);
