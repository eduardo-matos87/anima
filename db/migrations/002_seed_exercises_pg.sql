-- Equipamentos básicos
INSERT INTO equipment(name) VALUES
  ('halter'),('barra'),('maquina'),('livre')
ON CONFLICT DO NOTHING;

-- Exercícios iniciais (cobertura corpo todo)
INSERT INTO exercises (name, primary_muscle, difficulty, equipment, is_unilateral) VALUES
('Supino reto com barra','peito','beginner','barra',false),
('Supino inclinado com halteres','peito','beginner','halter',false),
('Crucifixo na máquina','peito','beginner','maquina',false),

('Puxada frontal','costas','beginner','maquina',false),
('Remada curvada','costas','intermediate','barra',false),

('Agachamento livre','pernas','intermediate','barra',false),
('Leg press','pernas','beginner','maquina',false),

('Desenvolvimento com halteres','ombros','beginner','halter',false),
('Elevação lateral','ombros','beginner','halter',false),

('Rosca direta','biceps','beginner','barra',false),
('Tríceps na polia','triceps','beginner','maquina',false),

('Panturrilha em pé','panturrilha','beginner','maquina',false),
('Prancha','core','beginner','livre',false)
ON CONFLICT DO NOTHING;

