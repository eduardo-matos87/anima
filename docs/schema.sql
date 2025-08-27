--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Ubuntu 16.9-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.9 (Ubuntu 16.9-0ubuntu0.24.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: exercicios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exercicios (
    id integer NOT NULL,
    nome text NOT NULL,
    grupo_id integer NOT NULL
);


--
-- Name: exercicios_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exercicios_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exercicios_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exercicios_id_seq OWNED BY public.exercicios.id;


--
-- Name: exercises; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exercises (
    id integer NOT NULL,
    name text NOT NULL,
    muscle_group text NOT NULL,
    equipment text[] DEFAULT '{}'::text[] NOT NULL,
    difficulty text DEFAULT 'iniciante'::text NOT NULL,
    is_bodyweight boolean DEFAULT false NOT NULL
);


--
-- Name: exercises_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exercises_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exercises_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exercises_id_seq OWNED BY public.exercises.id;


--
-- Name: generations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.generations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    input_json jsonb NOT NULL,
    output_json jsonb,
    prompt_version text NOT NULL,
    model text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: grupos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.grupos (
    id integer NOT NULL,
    nome text NOT NULL
);


--
-- Name: grupos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.grupos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: grupos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.grupos_id_seq OWNED BY public.grupos.id;


--
-- Name: objetivos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.objetivos (
    id integer NOT NULL,
    nome text NOT NULL
);


--
-- Name: objetivos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.objetivos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: objetivos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.objetivos_id_seq OWNED BY public.objetivos.id;


--
-- Name: treino_exercicios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.treino_exercicios (
    id integer NOT NULL,
    treino_id integer NOT NULL,
    exercicio_id integer NOT NULL
);


--
-- Name: treino_exercicios_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.treino_exercicios_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: treino_exercicios_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.treino_exercicios_id_seq OWNED BY public.treino_exercicios.id;


--
-- Name: treinos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.treinos (
    id integer NOT NULL,
    nivel text NOT NULL,
    objetivo text NOT NULL,
    dias integer NOT NULL,
    divisao text NOT NULL
);


--
-- Name: treinos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.treinos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: treinos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.treinos_id_seq OWNED BY public.treinos.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text,
    email text NOT NULL,
    password_hash text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: exercicios id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exercicios ALTER COLUMN id SET DEFAULT nextval('public.exercicios_id_seq'::regclass);


--
-- Name: exercises id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exercises ALTER COLUMN id SET DEFAULT nextval('public.exercises_id_seq'::regclass);


--
-- Name: grupos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.grupos ALTER COLUMN id SET DEFAULT nextval('public.grupos_id_seq'::regclass);


--
-- Name: objetivos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objetivos ALTER COLUMN id SET DEFAULT nextval('public.objetivos_id_seq'::regclass);


--
-- Name: treino_exercicios id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treino_exercicios ALTER COLUMN id SET DEFAULT nextval('public.treino_exercicios_id_seq'::regclass);


--
-- Name: treinos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treinos ALTER COLUMN id SET DEFAULT nextval('public.treinos_id_seq'::regclass);


--
-- Name: exercicios exercicios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exercicios
    ADD CONSTRAINT exercicios_pkey PRIMARY KEY (id);


--
-- Name: exercises exercises_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exercises
    ADD CONSTRAINT exercises_pkey PRIMARY KEY (id);


--
-- Name: generations generations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_pkey PRIMARY KEY (id);


--
-- Name: grupos grupos_nome_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.grupos
    ADD CONSTRAINT grupos_nome_key UNIQUE (nome);


--
-- Name: grupos grupos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.grupos
    ADD CONSTRAINT grupos_pkey PRIMARY KEY (id);


--
-- Name: objetivos objetivos_nome_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objetivos
    ADD CONSTRAINT objetivos_nome_key UNIQUE (nome);


--
-- Name: objetivos objetivos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.objetivos
    ADD CONSTRAINT objetivos_pkey PRIMARY KEY (id);


--
-- Name: treino_exercicios treino_exercicios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treino_exercicios
    ADD CONSTRAINT treino_exercicios_pkey PRIMARY KEY (id);


--
-- Name: treinos treinos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treinos
    ADD CONSTRAINT treinos_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_exercises_muscle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exercises_muscle ON public.exercises USING btree (muscle_group);


--
-- Name: idx_generations_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_generations_created_at ON public.generations USING btree (created_at);


--
-- Name: idx_treino_exercicios_exercicio_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_treino_exercicios_exercicio_id ON public.treino_exercicios USING btree (exercicio_id);


--
-- Name: idx_treino_exercicios_treino_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_treino_exercicios_treino_id ON public.treino_exercicios USING btree (treino_id);


--
-- Name: exercicios exercicios_grupo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exercicios
    ADD CONSTRAINT exercicios_grupo_id_fkey FOREIGN KEY (grupo_id) REFERENCES public.grupos(id) ON DELETE CASCADE;


--
-- Name: treino_exercicios fk_te_exercicio; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treino_exercicios
    ADD CONSTRAINT fk_te_exercicio FOREIGN KEY (exercicio_id) REFERENCES public.exercises(id) ON DELETE CASCADE;


--
-- Name: treino_exercicios fk_te_treino; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.treino_exercicios
    ADD CONSTRAINT fk_te_treino FOREIGN KEY (treino_id) REFERENCES public.treinos(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

