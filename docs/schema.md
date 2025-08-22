# Anima – Snapshot de Schema

- Host: `127.0.0.1:5432`
- Database: `anima`
- Schema: `public`
- Data: `2025-08-22T15:56:47-03:00`

> Gerado por tools/export_schema.sh

## Extensões

- `pgcrypto` (`1.3`)
- `plpgsql` (`1.0`)
- `unaccent` (`1.1`)

## exercicios

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('exercicios_id_seq'::regclass)` |
| `nome` | `text` | `` | `NO` | `` |
| `grupo_id` | `integer` | `` | `NO` | `` |

### Primary Key

- id

### Foreign Keys
- **exercicios_grupo_id_fkey**: `FOREIGN KEY (grupo_id) REFERENCES grupos(id) ON DELETE CASCADE`

### Índices
- **exercicios_pkey**: `CREATE UNIQUE INDEX exercicios_pkey ON public.exercicios USING btree (id)`

## exercises

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('exercises_id_seq'::regclass)` |
| `name` | `text` | `` | `NO` | `` |
| `muscle_group` | `text` | `` | `NO` | `` |
| `equipment` | `ARRAY` | `` | `NO` | `'{}'::text[]` |
| `difficulty` | `text` | `` | `NO` | `'iniciante'::text` |
| `is_bodyweight` | `boolean` | `` | `NO` | `false` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **exercises_pkey**: `CREATE UNIQUE INDEX exercises_pkey ON public.exercises USING btree (id)`
- **idx_exercises_muscle**: `CREATE INDEX idx_exercises_muscle ON public.exercises USING btree (muscle_group)`

## generations

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `uuid` | `` | `NO` | `gen_random_uuid()` |
| `user_id` | `uuid` | `` | `YES` | `` |
| `input_json` | `jsonb` | `` | `NO` | `` |
| `output_json` | `jsonb` | `` | `YES` | `` |
| `prompt_version` | `text` | `` | `NO` | `` |
| `model` | `text` | `` | `YES` | `` |
| `created_at` | `timestamp with time zone` | `` | `YES` | `now()` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **generations_pkey**: `CREATE UNIQUE INDEX generations_pkey ON public.generations USING btree (id)`
- **idx_generations_created_at**: `CREATE INDEX idx_generations_created_at ON public.generations USING btree (created_at)`

## grupos

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('grupos_id_seq'::regclass)` |
| `nome` | `text` | `` | `NO` | `` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **grupos_nome_key**: `CREATE UNIQUE INDEX grupos_nome_key ON public.grupos USING btree (nome)`
- **grupos_pkey**: `CREATE UNIQUE INDEX grupos_pkey ON public.grupos USING btree (id)`

## objetivos

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('objetivos_id_seq'::regclass)` |
| `nome` | `text` | `` | `NO` | `` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **objetivos_nome_key**: `CREATE UNIQUE INDEX objetivos_nome_key ON public.objetivos USING btree (nome)`
- **objetivos_pkey**: `CREATE UNIQUE INDEX objetivos_pkey ON public.objetivos USING btree (id)`

## schema_migrations

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `version` | `bigint` | `` | `NO` | `` |
| `dirty` | `boolean` | `` | `NO` | `` |

### Primary Key

- version

### Foreign Keys
- (nenhuma)

### Índices
- **schema_migrations_pkey**: `CREATE UNIQUE INDEX schema_migrations_pkey ON public.schema_migrations USING btree (version)`

## treino_exercicios

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('treino_exercicios_id_seq'::regclass)` |
| `treino_id` | `integer` | `` | `NO` | `` |
| `exercicio_id` | `integer` | `` | `NO` | `` |

### Primary Key

- id

### Foreign Keys
- **fk_te_exercicio**: `FOREIGN KEY (exercicio_id) REFERENCES exercises(id) ON DELETE CASCADE`
- **fk_te_treino**: `FOREIGN KEY (treino_id) REFERENCES treinos(id) ON DELETE CASCADE`

### Índices
- **idx_treino_exercicios_exercicio_id**: `CREATE INDEX idx_treino_exercicios_exercicio_id ON public.treino_exercicios USING btree (exercicio_id)`
- **idx_treino_exercicios_treino_id**: `CREATE INDEX idx_treino_exercicios_treino_id ON public.treino_exercicios USING btree (treino_id)`
- **treino_exercicios_pkey**: `CREATE UNIQUE INDEX treino_exercicios_pkey ON public.treino_exercicios USING btree (id)`

## treinos

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `integer` | `` | `NO` | `nextval('treinos_id_seq'::regclass)` |
| `nivel` | `text` | `` | `NO` | `` |
| `objetivo` | `text` | `` | `NO` | `` |
| `dias` | `integer` | `` | `NO` | `` |
| `divisao` | `text` | `` | `NO` | `` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **treinos_pkey**: `CREATE UNIQUE INDEX treinos_pkey ON public.treinos USING btree (id)`

## users

### Colunas

| coluna | tipo | len | null | default |
|-------|------|-----|------|---------|
| `id` | `uuid` | `` | `NO` | `gen_random_uuid()` |
| `name` | `text` | `` | `YES` | `` |
| `email` | `text` | `` | `NO` | `` |
| `password_hash` | `text` | `` | `YES` | `` |
| `created_at` | `timestamp with time zone` | `` | `YES` | `now()` |

### Primary Key

- id

### Foreign Keys
- (nenhuma)

### Índices
- **users_email_key**: `CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email)`
- **users_pkey**: `CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)`

