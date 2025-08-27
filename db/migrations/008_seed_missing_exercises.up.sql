-- Preenche exercícios que as seeds de "treino_exercicios" referenciam
INSERT INTO exercises (id, name, muscle_group, equipment, difficulty, is_bodyweight) VALUES
(1,  'Supino reto com barra',            'peito',  '{barra,anilhas}', 'intermediario', false),
(2,  'Supino inclinado com halteres',    'peito',  '{halteres}',      'intermediario', false),
(3,  'Remada curvada com barra',         'costas', '{barra,anilhas}', 'intermediario', false),
(4,  'Puxada frontal na polia',          'costas', '{polia}',         'iniciante',     false),
(5,  'Agachamento livre',                'pernas', '{barra,anilhas}', 'intermediario', false),
(6,  'Leg press 45°',                    'pernas', '{}',               'iniciante',     false),
(10, 'Abdominal supra',                  'abdomen','{}',               'iniciante',     true),
(11, 'Esteira',                          'cardio', '{}',               'iniciante',     false)
ON CONFLICT (id) DO NOTHING;
