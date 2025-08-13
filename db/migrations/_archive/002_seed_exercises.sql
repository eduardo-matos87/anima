INSERT INTO exercises (name, muscle_group, equipment, difficulty, is_compound) VALUES
-- Peito
('Supino reto com barra', 'chest', 'barbell', 'beginner', true),
('Supino inclinado com halteres', 'chest', 'dumbbell', 'beginner', true),
('Crucifixo na máquina (peck deck)', 'chest', 'machine', 'beginner', false),
('Crucifixo com halteres', 'chest', 'dumbbell', 'beginner', false),
('Crossover na polia', 'chest', 'cable', 'beginner', false),

-- Costas
('Levantamento terra (deadlift)', 'back', 'barbell', 'intermediate', true),
('Remada curvada com barra', 'back', 'barbell', 'beginner', true),
('Puxada frente na máquina', 'back', 'machine', 'beginner', false),
('Remada baixa na polia', 'back', 'cable', 'beginner', false),

-- Pernas
('Agachamento livre', 'legs', 'barbell', 'beginner', true),
('Leg press', 'legs', 'machine', 'beginner', true),
('Cadeira extensora', 'legs', 'machine', 'beginner', false),
('Mesa flexora', 'legs', 'machine', 'beginner', false),
('Panturrilha em pé', 'legs', 'machine', 'beginner', false),

-- Ombros
('Desenvolvimento com barra', 'shoulders', 'barbell', 'beginner', true),
('Elevação lateral com halteres', 'shoulders', 'dumbbell', 'beginner', false),
('Desenvolvimento na máquina', 'shoulders', 'machine', 'beginner', true),

-- Braços
('Rosca direta', 'arms', 'barbell', 'beginner', false),
('Rosca alternada', 'arms', 'dumbbell', 'beginner', false),
('Tríceps testa', 'arms', 'barbell', 'beginner', false),
('Tríceps na polia', 'arms', 'cable', 'beginner', false),

-- Core
('Prancha', 'core', 'bodyweight', 'beginner', false),
('Abdominal crunch máquina', 'core', 'machine', 'beginner', false);
