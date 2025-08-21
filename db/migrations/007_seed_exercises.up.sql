INSERT INTO exercises (name, muscle_group, equipment, difficulty, is_bodyweight) VALUES
('Supino reto com barra', 'peito',  '{barra,anilhas}', 'intermediario', false),
('Agachamento livre',     'pernas', '{barra,anilhas}', 'intermediario', false),
('Puxada frontal',        'costas', '{polia}',         'iniciante',     false),
('Flexão de braços',      'peito',  '{}',              'iniciante',     true)
ON CONFLICT DO NOTHING;
