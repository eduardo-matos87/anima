-- Arquivo: db/migrations/001_init_schema_postgres.sql

-- Remove se existir (em dev)
DROP TABLE IF EXISTS treino_exercicios;
DROP TABLE IF EXISTS treinos;
DROP TABLE IF EXISTS exercicios;
DROP TABLE IF EXISTS grupos_musculares;
DROP TABLE IF EXISTS objetivos;
DROP TABLE IF EXISTS usuarios;

-- Tabela de usuários (registro/login)
CREATE TABLE usuarios (
  id SERIAL PRIMARY KEY,
  nome TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  senha_hash TEXT NOT NULL
);

-- Objetivos possíveis: emagrecimento, massa magra, resistência...
CREATE TABLE objetivos (
  id SERIAL PRIMARY KEY,
  nome TEXT UNIQUE NOT NULL
);

-- Grupos musculares: peito, costas, pernas...
CREATE TABLE grupos_musculares (
  id SERIAL PRIMARY KEY,
  nome TEXT UNIQUE NOT NULL
);

-- Exercícios: nome e grupo muscular relacionado
CREATE TABLE exercicios (
  id SERIAL PRIMARY KEY,
  nome TEXT NOT NULL,
  grupo_id INTEGER REFERENCES grupos_musculares(id)
);

-- Treinos gerados pelo sistema ou criados pelo usuário
CREATE TABLE treinos (
  id SERIAL PRIMARY KEY,
  usuario_id INTEGER REFERENCES usuarios(id),
  nivel TEXT NOT NULL,
  objetivo_id INTEGER REFERENCES objetivos(id),
  dias INTEGER NOT NULL,
  divisao TEXT NOT NULL
);

-- Relacionamento entre treino e exercícios
CREATE TABLE treino_exercicios (
  treino_id INTEGER REFERENCES treinos(id) ON DELETE CASCADE,
  exercicio_id INTEGER REFERENCES exercicios(id),
  PRIMARY KEY (treino_id, exercicio_id)
);

-- Inserts padrões de objetivos
INSERT INTO objetivos (nome) VALUES
  ('Emagrecimento'),
  ('Ganho de massa magra'),
  ('Resistência física'),
  ('Hipertrofia');

-- Inserts padrões de grupos musculares
INSERT INTO grupos_musculares (nome) VALUES
  ('Peito'),
  ('Costas'),
  ('Pernas'),
  ('Ombros'),
  ('Bíceps'),
  ('Tríceps'),
  ('Abdômen'),
  ('Cardio');

-- Inserts de exercícios por grupo muscular
INSERT INTO exercicios (nome, grupo_id) VALUES
  -- Peito
  ('Supino reto com barra', 1),
  ('Supino inclinado com halteres', 1),
  ('Crossover na polia', 1),
  ('Flexão de braço', 1),

  -- Costas
  ('Remada curvada com barra', 2),
  ('Puxada frontal na polia', 2),
  ('Pulldown unilateral', 2),
  ('Remada baixa na máquina', 2),

  -- Pernas
  ('Agachamento livre', 3),
  ('Leg press 45°', 3),
  ('Cadeira extensora', 3),
  ('Mesa flexora', 3),

  -- Ombros
  ('Desenvolvimento militar com barra', 4),
  ('Elevação lateral com halteres', 4),
  ('Elevação frontal com anilhas', 4),
  ('Remada alta', 4),

  -- Bíceps
  ('Rosca direta com barra', 5),
  ('Rosca alternada com halteres', 5),
  ('Rosca concentrada', 5),

  -- Tríceps
  ('Tríceps testa com barra', 6),
  ('Tríceps pulley com corda', 6),
  ('Tríceps banco', 6),

  -- Abdômen
  ('Abdominal supra no solo', 7),
  ('Prancha abdominal', 7),
  ('Elevação de pernas', 7),

  -- Cardio
  ('Esteira (cardio)', 8),
  ('Bicicleta ergométrica', 8),
  ('Escada ergométrica', 8);

-- Treinos para hipertrofia (nível intermediário, 5 dias, divisão ABCDE)
INSERT INTO treinos (nivel, objetivo_id, dias, divisao) VALUES
  ('intermediário', 4, 5, 'A'), -- Peito + Tríceps
  ('intermediário', 4, 5, 'B'), -- Costas + Bíceps
  ('intermediário', 4, 5, 'C'), -- Pernas
  ('intermediário', 4, 5, 'D'), -- Ombros
  ('intermediário', 4, 5, 'E'); -- Abdômen + Cardio

-- Associar exercícios aos treinos de hipertrofia
INSERT INTO treino_exercicios (treino_id, exercicio_id) VALUES
  -- Treino A: Peito + Tríceps
  (1, 1), (1, 2), (1, 3), (1, 20), (1, 21),

  -- Treino B: Costas + Bíceps
  (2, 5), (2, 6), (2, 7), (2, 17), (2, 18),

  -- Treino C: Pernas
  (3, 9), (3, 10), (3, 11), (3, 12),

  -- Treino D: Ombros
  (4, 13), (4, 14), (4, 15), (4, 16),

  -- Treino E: Abdômen + Cardio
  (5, 25), (5, 26), (5, 27), (5, 28), (5, 29);
