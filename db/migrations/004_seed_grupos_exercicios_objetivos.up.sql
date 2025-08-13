-- Tabelas (PostgreSQL)

CREATE TABLE IF NOT EXISTS grupos (
  id SERIAL PRIMARY KEY,
  nome TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS objetivos (
  id SERIAL PRIMARY KEY,
  nome TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS exercicios (
  id SERIAL PRIMARY KEY,
  nome TEXT NOT NULL,
  grupo_id INT NOT NULL REFERENCES grupos(id) ON DELETE CASCADE
);

-- Seeds (idempotentes)
INSERT INTO grupos (nome) VALUES
  ('Peito'), ('Costas'), ('Pernas'), ('Ombros'),
  ('Bíceps'), ('Tríceps'), ('Abdômen'), ('Cardio')
ON CONFLICT (nome) DO NOTHING;

INSERT INTO exercicios (nome, grupo_id) VALUES
  ('Supino reto com barra', (SELECT id FROM grupos WHERE nome='Peito')),
  ('Supino inclinado com halteres', (SELECT id FROM grupos WHERE nome='Peito')),
  ('Remada curvada com barra', (SELECT id FROM grupos WHERE nome='Costas')),
  ('Puxada frontal na polia', (SELECT id FROM grupos WHERE nome='Costas')),
  ('Agachamento livre', (SELECT id FROM grupos WHERE nome='Pernas')),
  ('Leg press 45°', (SELECT id FROM grupos WHERE nome='Pernas')),
  ('Desenvolvimento com halteres', (SELECT id FROM grupos WHERE nome='Ombros')),
  ('Rosca direta', (SELECT id FROM grupos WHERE nome='Bíceps')),
  ('Tríceps pulley', (SELECT id FROM grupos WHERE nome='Tríceps')),
  ('Abdominal supra', (SELECT id FROM grupos WHERE nome='Abdômen')),
  ('Esteira', (SELECT id FROM grupos WHERE nome='Cardio'))
ON CONFLICT DO NOTHING;

INSERT INTO objetivos (nome) VALUES
  ('Emagrecimento'), ('Ganho de massa magra'), ('Resistência física')
ON CONFLICT (nome) DO NOTHING;
