-- Treino Iniciante | Emagrecimento | 3 dias | Divisão ABC
INSERT INTO treinos (nivel, objetivo, dias, divisao) VALUES
('iniciante', 'Emagrecimento', 3, 'A'),
('iniciante', 'Emagrecimento', 3, 'B'),
('iniciante', 'Emagrecimento', 3, 'C');

-- Vincula exercícios ao treino_id (assumindo que são os IDs 1, 2, 3)
-- Treino A: Peito + Cardio
INSERT INTO treino_exercicios (treino_id, exercicio_id) VALUES
(1, 1),  -- Supino reto com barra
(1, 2),  -- Supino inclinado com halteres
(1, 11); -- Esteira

-- Treino B: Pernas
INSERT INTO treino_exercicios (treino_id, exercicio_id) VALUES
(2, 5),  -- Agachamento livre
(2, 6);  -- Leg press 45°

-- Treino C: Costas + Abdômen
INSERT INTO treino_exercicios (treino_id, exercicio_id) VALUES
(3, 3),  -- Remada curvada com barra
(3, 4),  -- Puxada frontal na polia
(3, 10); -- Abdominal supra

