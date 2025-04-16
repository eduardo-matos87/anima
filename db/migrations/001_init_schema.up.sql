-- Tabelas principais

CREATE TABLE grupos_musculares (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nome TEXT NOT NULL
);

CREATE TABLE exercicios (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nome TEXT NOT NULL,
  grupo_id INTEGER,
  FOREIGN KEY (grupo_id) REFERENCES grupos_musculares(id)
);

CREATE TABLE objetivos (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nome TEXT NOT NULL
);

CREATE TABLE treinos (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nivel TEXT,
  objetivo TEXT,
  dias INTEGER,
  divisao TEXT
);

CREATE TABLE treino_exercicios (
  treino_id INTEGER,
  exercicio_id INTEGER,
  FOREIGN KEY (treino_id) REFERENCES treinos(id),
  FOREIGN KEY (exercicio_id) REFERENCES exercicios(id)
);

-- Dados iniciais

INSERT INTO grupos_musculares (nome) VALUES
('Peito'), ('Costas'), ('Pernas'), ('Ombros'),
('Bíceps'), ('Tríceps'), ('Abdômen'), ('Cardio');

INSERT INTO exercicios (nome, grupo_id) VALUES
('Supino reto com barra', 1),
('Supino inclinado com halteres', 1),
('Remada curvada com barra', 2),
('Puxada frontal na polia', 2),
('Agachamento livre', 3),
('Leg press 45°', 3),
('Desenvolvimento com halteres', 4),
('Rosca direta', 5),
('Tríceps pulley', 6),
('Abdominal supra', 7),
('Esteira', 8);

INSERT INTO objetivos (nome) VALUES
('Emagrecimento'), ('Ganho de massa magra'), ('Resistência física');
